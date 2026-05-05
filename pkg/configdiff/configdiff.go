package configdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/beevik/etree"
	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	loader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/data-server/pkg/pool"
	"github.com/sdcio/data-server/pkg/tree"
	"github.com/sdcio/data-server/pkg/tree/api"
	"github.com/sdcio/data-server/pkg/tree/consts"
	treeImporter "github.com/sdcio/data-server/pkg/tree/importer"
	treejson "github.com/sdcio/data-server/pkg/tree/importer/json"
	treexml "github.com/sdcio/data-server/pkg/tree/importer/xml"
	"github.com/sdcio/data-server/pkg/tree/ops"
	"github.com/sdcio/data-server/pkg/tree/ops/validation"
	"github.com/sdcio/data-server/pkg/tree/processors"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	schemaSrvConf "github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/store"
	"github.com/sdcio/schema-server/pkg/store/persiststore"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/diff"
	"github.com/sdcio/sdc-lite/pkg/schemaclient"
	"github.com/sdcio/sdc-lite/pkg/schemaloader"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type ConfigDiff struct {
	config      *config.Config
	tree        *tree.RootEntry
	schema      *sdcpb.Schema
	schemaStore store.Store
	sharedPool  *pool.SharedTaskPool
}

var _ params.Executor = (*ConfigDiff)(nil)

func NewConfigDiff(ctx context.Context, c *config.Config) (*ConfigDiff, error) {
	llevel, err := log.ParseLevel(c.LogLevel())
	if err != nil {
		return nil, err
	}
	log.SetLevel(llevel)
	log.SetOutput(os.Stderr)

	cd := &ConfigDiff{
		config:     c,
		sharedPool: pool.NewSharedTaskPool(ctx, runtime.GOMAXPROCS(0)),
	}

	return cd, nil
}

func (c *ConfigDiff) CopyEmptyConfigDiff(ctx context.Context) (*ConfigDiff, error) {
	result, err := NewConfigDiff(ctx, c.config)
	if err != nil {
		return nil, err
	}
	result.schema = c.schema
	result.schemaStore = c.schemaStore
	return result, nil
}

func (c *ConfigDiff) GetTreeJson(ctx context.Context, path *sdcpb.Path) (any, error) {
	// finish InsertionPhase on tree
	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}
	// navigate to path
	entry, err := ops.NavigateSdcpbPath(ctx, c.tree.Entry, path)
	if err != nil {
		return nil, err
	}
	// retrive running Tree json
	jTree, err := ops.ToJson(ctx, entry, false)
	if err != nil {
		return nil, err
	}

	return jTree, nil
}

func (c *ConfigDiff) GetRunningJson(ctx context.Context, path *sdcpb.Path) (any, error) {
	// export running intents
	lvs := ops.LeafsOfOwner(c.tree.Entry, consts.RunningIntentName)

	// create a new Tree for running
	runningTree, err := c.newTreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	// add running to the new running tree
	err = runningTree.AddUpdatesRecursive(ctx, lvs.ToPathAndUpdateSlice(), treetypes.NewUpdateInsertFlags())
	if err != nil {
		return nil, err
	}
	// finish InsertionPhase on running tree
	err = runningTree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}

	entry, err := ops.NavigateSdcpbPath(ctx, runningTree.Entry, path)
	if err != nil {
		return nil, err
	}

	// retrive running Tree json
	jrunTree, err := ops.ToJson(ctx, entry, false)
	if err != nil {
		return nil, err
	}

	return jrunTree, nil
}

func (c *ConfigDiff) GetDiff(ctx context.Context, dc *params.DiffConfig) (string, error) {
	//TODO: currently we only support json diff
	if dc.GetFormat() != types.ConfigFormatJson {
		return "", fmt.Errorf("diff only supported for json format currently")
	}

	runningJson, err := c.GetRunningJson(ctx, dc.GetPath())
	if err != nil {
		log.Warn(err)
	}
	treeJson, err := c.GetTreeJson(ctx, dc.GetPath())
	if err != nil {
		log.Warn(err)
	}

	differ, err := diff.NewDifferJson(runningJson, treeJson)
	if err != nil {
		return "", err
	}

	differ.SetConfig(dc)

	return differ.Diff()
}

func (c *ConfigDiff) SetSchemaStore(s store.Store) {
	c.schemaStore = s
}

func (c *ConfigDiff) closeSchemaStore() error {
	err := c.schemaStore.Close()
	c.schemaStore = nil
	return err
}

func (c *ConfigDiff) loadSchemaStore(ctx context.Context, readOnly bool, vendor string, version string) (store.Store, error) {
	if c.schemaStore != nil && c.schemaStore.IsReadOnly() == readOnly {
		return c.schemaStore, nil
	}
	if c.schemaStore != nil {
		_ = c.closeSchemaStore()
	}

	storePath := filepath.Join(c.config.SchemaStorePath(), strings.ToLower(vendor), strings.ToLower(version))

	if !utils.FolderExists(storePath) {
		utils.CreateFolder(storePath)
		readOnly = false
	}

	s, err := persiststore.New(ctx, storePath, &schemaSrvConf.SchemaPersistStoreCacheConfig{WithDescription: false}, readOnly)
	if err != nil {
		return nil, err
	}
	c.schemaStore = s
	return s, nil
}

func (c *ConfigDiff) GetSchemaStore() store.Store {
	return c.schemaStore
}

func (c *ConfigDiff) SchemaRemove(ctx context.Context, vendor string, version string) error {
	if vendor == "" {
		return fmt.Errorf("vendor must not be empty")
	}
	if version == "" {
		return fmt.Errorf("version must not be empty")
	}
	err := os.RemoveAll(filepath.Join(c.config.SchemaStorePath(), vendor, version))
	return err
}

func (c *ConfigDiff) HasSchema() bool {
	return c.schema != nil
}

func (c *ConfigDiff) SchemasList(ctx context.Context) ([]*sdcpb.Schema, error) {
	vendors, err := os.ReadDir(c.config.SchemaStorePath())
	if err != nil {
		return nil, err
	}

	schemas := []*sdcpb.Schema{}

	for _, vendor := range vendors {
		if !vendor.IsDir() {
			continue
		}

		vendorPath := filepath.Join(c.config.SchemaStorePath(), vendor.Name())

		versions, err := os.ReadDir(vendorPath)
		if err != nil {
			panic(err)
		}

		for _, version := range versions {
			if !version.IsDir() {
				continue
			}
			schemas = append(schemas, &sdcpb.Schema{Vendor: vendor.Name(), Version: version.Name()})
		}
	}

	return schemas, nil
}

func (c *ConfigDiff) SetSchema(s *sdcpb.Schema) {
	c.schema = s
}

func (c *ConfigDiff) SchemaDownload(ctx context.Context, slc *params.SchemaLoadConfig) (*sdcpb.Schema, error) {

	schemaDef := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(slc.GetSchema(), &schemaDef); err != nil {
		return nil, err
	}

	sdcpbSchema := &sdcpb.Schema{
		Name:    "",
		Vendor:  schemaDef.Spec.Provider,
		Version: schemaDef.Spec.Version,
	}

	// open schema store in readonly first
	schemaStore, err := c.loadSchemaStore(ctx, true, schemaDef.Spec.Provider, schemaDef.Spec.Version)
	if err != nil {
		return nil, err
	}

	// check if the schema already exists
	schemaExists := schemaStore.HasSchema(store.SchemaKey{Vendor: schemaDef.Spec.Provider, Version: schemaDef.Spec.Version})

	// return in case of existence
	if schemaExists {
		log.Infof("Schema - Vendor: %s, Version: %s - Already Exists, Skip Loading", schemaDef.Spec.Provider, schemaDef.Spec.Version)
		c.SetSchema(sdcpbSchema)
		return sdcpbSchema, nil
	}

	// if schema does not exist, open in read/write and continue
	schemaStore, err = c.loadSchemaStore(ctx, false, schemaDef.Spec.Provider, schemaDef.Spec.Version)
	if err != nil {
		return nil, err
	}

	defer schemaStore.Close()

	schemaLoader, err := loader.NewLoader(
		c.config.DownloadPath(),
		c.config.SchemasPath(),
		schemaloader.NewNopResolver(),
	)
	if err != nil {
		return nil, err
	}

	schemaLoader.AddRef(ctx, schemaDef)
	_, dirExists, err := schemaLoader.GetRef(ctx, schemaDef.Spec.GetKey())
	if err != nil {
		return nil, err
	}
	if !dirExists {
		log.Info("loading...")
		if _, err := schemaLoader.Load(ctx, schemaDef.Spec.GetKey()); err != nil {
			return nil, err
		}
	}
	rsp, err := schemaStore.CreateSchema(ctx, &sdcpb.CreateSchemaRequest{
		Schema:    sdcpbSchema,
		File:      schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Models,
		Directory: schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Includes,
		Exclude:   schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Excludes,
	})
	if err != nil {
		return nil, err
	}

	c.SetSchema(sdcpbSchema)

	log.Infof("Schema - Vendor: %s, Version: %s - Loaded Successful", rsp.Schema.Vendor, rsp.Schema.Version)
	return sdcpbSchema, nil
}

func (c *ConfigDiff) newTreeRoot(ctx context.Context) (*tree.RootEntry, error) {
	schemaStore, err := c.loadSchemaStore(ctx, true, c.schema.GetVendor(), c.schema.GetVersion())
	if err != nil {
		return nil, err
	}

	scb := schemaclient.NewMemSchemaClientBound(schemaStore, c.schema)
	tc := tree.NewTreeContext(scb, c.sharedPool)
	t, err := tree.NewTreeRoot(ctx, tc)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (c *ConfigDiff) buildRootTree(ctx context.Context) (err error) {
	c.tree, err = c.newTreeRoot(ctx)
	return err
}

func (c *ConfigDiff) TreeBlame(ctx context.Context, params *params.ConfigBlameParams) (*sdcpb.BlameTreeElement, error) {
	var elem = c.tree.Entry
	var err error
	// if a path is provided, navigate to the path first
	if params.GetPath() != nil {
		elem, err = ops.NavigateSdcpbPath(ctx, c.tree.Entry, params.GetPath())
		if err != nil {
			return nil, err
		}
	}
	// create config for blame processor
	blameProcessor := processors.NewBlameConfigProcessor(&processors.BlameConfigProcessorParams{IncludeDefaults: params.GetIncludeDefaults()})
	blametree, err := blameProcessor.Run(ctx, elem, c.sharedPool)
	if err != nil {
		return nil, err
	}

	return blametree, nil
}

func (c *ConfigDiff) TreeLoadData(ctx context.Context, cl *params.ConfigLoad) error {
	var err error
	var importer treeImporter.ImportConfigAdapter

	if c.tree == nil {
		c.tree, err = c.newTreeRoot(ctx)
		if err != nil {
			return err
		}
	}

	intent := cl.GetIntent()

	switch intent.GetFormat() {
	case types.ConfigFormatJson, types.ConfigFormatJsonIetf:
		var j any
		err = json.Unmarshal(intent.GetData(), &j)
		if err != nil {
			return err
		}
		importer = treejson.NewJsonTreeImporter(j, intent.GetName(), intent.GetPrio(), false)
	case types.ConfigFormatYaml:
		var y any
		err = yaml.Unmarshal(intent.GetData(), &y)
		if err != nil {
			return err
		}
		// we use json importer since we're based basically on map[string]any
		importer = treejson.NewJsonTreeImporter(y, intent.GetName(), intent.GetPrio(), false)
	case types.ConfigFormatXml:
		xmlDoc := etree.NewDocument()
		err := xmlDoc.ReadFromBytes(intent.GetData())
		if err != nil {
			return err
		}
		importer = treexml.NewXmlTreeImporter(&xmlDoc.Element, intent.GetName(), intent.GetPrio(), false)
	default:
		return fmt.Errorf("import of format %s not supported yet", intent.GetFormat().String())
	}

	// overwrite running intent with running prio
	if strings.EqualFold(intent.GetName(), consts.RunningIntentName) {
		intent.SetPrio(consts.RunningValuesPrio)
	}

	// convert base path to sdcpb.path
	path, err := sdcpb.ParsePath(intent.GetBasePath())
	if err != nil {
		return err
	}

	_, err = c.tree.ImportConfig(ctx, path, importer, cl.GetIntent().Flag, c.sharedPool)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigDiff) TreeShow(ctx context.Context, config *params.ConfigShowConfig) (params.ConfigShowInterface, error) {
	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}

	entry, err := ops.NavigateSdcpbPath(ctx, c.tree.Entry, config.GetPath())
	if err != nil {
		return nil, err
	}

	return newConfigShowEntryAdapter(entry), nil
}

func (c *ConfigDiff) GetJson(ctx context.Context, onlyNewOrUpdated bool) (any, error) {
	return ops.ToJson(ctx, c.tree.Entry, onlyNewOrUpdated)
}

func (c *ConfigDiff) TreeValidate(ctx context.Context) (treetypes.ValidationResults, *treetypes.ValidationStats, error) {
	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, nil, err
	}
	// run validation processor on the tree
	valResult, valStat := validation.Validate(ctx, c.tree.Entry, c.config.Validation(), c.sharedPool)
	return valResult, valStat, nil
}

func (c *ConfigDiff) GetPathCompletions(ctx context.Context, toComplete string) []string {
	return ops.GetPathCompletions(ctx, c.tree.Entry, toComplete)
}

// configShowEntryAdapter is an adapter that allows to use an api.Entry as a ConfigShowInterface
// wiring the api.Entry to the reqwuired ops for the ConfigShowInterface methods
type configShowEntryAdapter struct {
	entry api.Entry
}

func newConfigShowEntryAdapter(entry api.Entry) *configShowEntryAdapter {
	return &configShowEntryAdapter{
		entry: entry,
	}
}

func (a *configShowEntryAdapter) ToJson(ctx context.Context, onlyNewOrUpdated bool) (any, error) {
	return ops.ToJson(ctx, a.entry, onlyNewOrUpdated)
}

func (a *configShowEntryAdapter) ToJsonIETF(ctx context.Context, onlyNewOrUpdated bool) (any, error) {
	return ops.ToJsonIETF(ctx, a.entry, onlyNewOrUpdated)
}

func (a *configShowEntryAdapter) ToXML(ctx context.Context, onlyNewOrUpdated bool, honorNamespace bool, operationWithNamespace bool, useOperationRemove bool) (*etree.Document, error) {
	return ops.ToXML(ctx, a.entry, onlyNewOrUpdated, honorNamespace, operationWithNamespace, useOperationRemove)
}

func (a *configShowEntryAdapter) GetHighestPrecedence(_ context.Context, onlyNewOrUpdated bool, includeDefaults bool, includeExplicitDelete bool) api.LeafVariantSlice {
	return ops.GetHighestPrecedence(a.entry, onlyNewOrUpdated, includeDefaults, includeExplicitDelete)
}

func (a *configShowEntryAdapter) GetDeletes(_ context.Context, aggregatePaths bool) (treetypes.DeleteEntriesList, error) {
	return ops.GetDeletes(a.entry, aggregatePaths)
}
