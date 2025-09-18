package configdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/beevik/etree"
	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	loader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/data-server/pkg/tree"
	treeImporter "github.com/sdcio/data-server/pkg/tree/importer"
	treejson "github.com/sdcio/data-server/pkg/tree/importer/json"
	treexml "github.com/sdcio/data-server/pkg/tree/importer/xml"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/data-server/pkg/utils"
	schemaSrvConf "github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/store"
	"github.com/sdcio/schema-server/pkg/store/persiststore"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/diff"
	"github.com/sdcio/sdc-lite/pkg/schemaclient"
	"github.com/sdcio/sdc-lite/pkg/schemaloader"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type ConfigDiff struct {
	config      *config.Config
	tree        *tree.RootEntry
	schema      *sdcpb.Schema
	schemaStore store.Store
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
		config: c,
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
	entry, err := c.tree.NavigateSdcpbPath(ctx, path.GetElem(), true)
	if err != nil {
		return nil, err
	}
	// retrive running Tree json
	jTree, err := entry.ToJson(false)
	if err != nil {
		return nil, err
	}

	return jTree, nil
}

func (c *ConfigDiff) GetRunningJson(ctx context.Context, path *sdcpb.Path) (any, error) {
	lvs := tree.LeafVariantSlice{}
	// export running intents
	lvs = c.tree.GetByOwner("running", lvs)

	// create a new Tree for running
	runningTree, err := c.newTreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	// add running to the new running tree
	err = runningTree.AddUpdatesRecursive(ctx, lvs.ToUpdateSlice(), treetypes.NewUpdateInsertFlags())
	if err != nil {
		return nil, err
	}
	// finish InsertionPhase on running tree
	err = runningTree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}

	entry, err := runningTree.NavigateSdcpbPath(ctx, path.GetElem(), true)
	if err != nil {
		return nil, err
	}

	// retrive running Tree json
	jrunTree, err := entry.ToJson(false)
	if err != nil {
		return nil, err
	}

	return jrunTree, nil
}

func (c *ConfigDiff) GetDiff(ctx context.Context, dc *params.DiffConfig) (string, error) {

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
		err := c.closeSchemaStore()
		if err != nil {
			// TODO log error
		}
	}
	s, err := persiststore.New(ctx, filepath.Join(c.config.SchemaStorePath(), strings.ToLower(vendor), strings.ToLower(version)), &schemaSrvConf.SchemaPersistStoreCacheConfig{
		WithDescription: false,
		ReadOnly:        readOnly,
	})
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
	tc := tree.NewTreeContext(scb, "")
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
	cbv := tree.NewBlameConfigVisitor(params.GetIncludeDefaults())
	if params.GetPath() != nil {
		// process with a provided path
		start, err := c.tree.NavigateSdcpbPath(ctx, params.GetPath().GetElem(), true)
		if err != nil {
			return nil, err
		}
		err = start.Walk(ctx, cbv)
		if err != nil {
			return nil, err
		}
		return cbv.GetResult(), nil
	}
	// if no path is provided
	err := c.tree.Walk(ctx, cbv)
	if err != nil {
		return nil, err
	}
	return cbv.GetResult(), nil
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
		importer = treejson.NewJsonTreeImporter(j)

	case types.ConfigFormatXml:
		xmlDoc := etree.NewDocument()
		err := xmlDoc.ReadFromBytes(intent.GetData())
		if err != nil {
			return err
		}
		importer = treexml.NewXmlTreeImporter(&xmlDoc.Element)
	}

	// overwrite running intent with running prio
	if strings.EqualFold(intent.GetName(), tree.RunningIntentName) {
		intent.SetPrio(tree.RunningValuesPrio)
	}

	// convert base path to sdcpb.path
	path, err := utils.ParsePath(intent.GetBasePath())
	if err != nil {
		return err
	}
	// convert sdcpb.path to string slice path
	strSlicePath := utils.ToStrings(path, false, false)

	err = c.tree.ImportConfig(ctx, strSlicePath, importer, intent.GetName(), intent.GetPrio(), intent.GetFlag())
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigDiff) TreeShow(ctx context.Context, config *params.ConfigShowConfig) (params.ConfigShowInterface, error) {
	var err error
	err = c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}

	return c.tree.NavigateSdcpbPath(ctx, config.GetPath().GetElem(), true)
}

func (c *ConfigDiff) GetJson(onlyNewOrUpdated bool) (any, error) {
	return c.tree.ToJson(onlyNewOrUpdated)
}

func (c *ConfigDiff) TreeValidate(ctx context.Context) (treetypes.ValidationResults, *treetypes.ValidationStatOverall, error) {
	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, nil, err
	}
	valResult, valStat := c.tree.Validate(ctx, c.config.Validation())
	return valResult, valStat, nil
}

func (c *ConfigDiff) GetPathCompletions(ctx context.Context, toComplete string) []string {

	cleanToComplete := toComplete
	keyPart := ""
	doKeyAction := false
	if strings.LastIndex(toComplete, "[") > strings.LastIndex(toComplete, "]") {
		cleanToComplete = toComplete[:strings.LastIndex(toComplete, "[")]
		keyPart = toComplete[strings.LastIndex(toComplete, "[")+1:]
		doKeyAction = true
	}

	toCompletePath, err := sdcpb.ParsePath(cleanToComplete)
	if err != nil {
		return nil
	}
	if doKeyAction {
		if strings.Contains(keyPart, "=") {
			return c.completeKey(ctx, toCompletePath, keyPart)
		}
		return c.completeKeyName(ctx, toCompletePath, keyPart)
	}

	return c.completePathName(ctx, toCompletePath)

}
func (c *ConfigDiff) completeKey(ctx context.Context, toCompletePath *sdcpb.Path, leftover string) []string {
	attrName, attrVal, _ := strings.Cut(leftover, "=")

	entry, err := c.tree.NavigateSdcpbPath(ctx, toCompletePath.GetElem(), true)
	if err != nil {
		return nil
	}

	lastLevelKeys := toCompletePath.Elem[len(toCompletePath.Elem)-1].Key

	childs, err := entry.FilterChilds(lastLevelKeys)
	if err != nil {
		return nil
	}
	result := []string{}
	for _, e := range childs {
		em := e.GetChilds(tree.DescendMethodActiveChilds)
		lvs := tree.LeafVariantSlice{}
		lvs = em[attrName].GetHighestPrecedence(lvs, false, true, true)
		elemVal := lvs[0].Update.Value().ToString()
		if !strings.HasPrefix(elemVal, attrVal) {
			continue
		}
		newPath := toCompletePath.DeepCopy()
		if newPath.Elem[len(newPath.Elem)-1].Key == nil {
			newPath.Elem[len(newPath.Elem)-1].Key = map[string]string{}
		}
		newPath.Elem[len(newPath.Elem)-1].Key[attrName] = elemVal
		pstring := newPath.ToXPath(false)
		result = append(result, pstring)
	}
	return result
}
func (c *ConfigDiff) completeKeyName(ctx context.Context, toCompletePath *sdcpb.Path, leftOver string) []string {
	toCompletePathCopy := toCompletePath.DeepCopy()
	existingKeys := map[string]struct{}{}
	for k := range toCompletePathCopy.Elem[len(toCompletePathCopy.Elem)-1].Key {
		existingKeys[k] = struct{}{}
	}

	toCompletePathCopy.Elem[len(toCompletePathCopy.Elem)-1].Key = nil
	entry, err := c.tree.NavigateSdcpbPath(ctx, toCompletePath.GetElem(), true)
	if err != nil {
		return nil
	}
	result := []string{}
	for _, k := range entry.GetSchemaKeys() {
		_, keyexists := existingKeys[k]
		if strings.HasPrefix(k, leftOver) && !keyexists {
			result = append(result, fmt.Sprintf("%s[%s=", toCompletePath.ToXPath(false), k))
		}
	}
	return result
}
func (c *ConfigDiff) completePathName(ctx context.Context, toCompletePath *sdcpb.Path) []string {
	var err error
	var entry tree.Entry
	var incompleteLastElem *sdcpb.PathElem
	if len(toCompletePath.Elem) > 0 {
		// check if the provied path points to something that exists
		_, err := c.tree.NavigateSdcpbPath(ctx, toCompletePath.GetElem(), true)
		if err != nil {
			// path does not exist, so lets strip last elem
			if len(toCompletePath.Elem[len(toCompletePath.Elem)-1].Key) > 0 {
				// processing keys
			} else {
				// processing normal path elements
				// remove the last element since it is probably just partial
				incompleteLastElem = toCompletePath.Elem[len(toCompletePath.Elem)-1]
				toCompletePath.Elem = toCompletePath.Elem[:len(toCompletePath.Elem)-1]
			}
		}
	}

	entry, err = c.tree.NavigateSdcpbPath(ctx, toCompletePath.GetElem(), true)
	if err != nil {
		return nil
	}
	childs := entry.GetChilds(tree.DescendMethodActiveChilds)

	var resultEntries []tree.Entry

	doAdd := true
	for k, v := range childs {
		if incompleteLastElem != nil {
			doAdd = strings.HasPrefix(k, incompleteLastElem.Name)
		}
		if doAdd {
			resultEntries = append(resultEntries, v)
		}
	}

	results := make([]string, 0, len(resultEntries))
	//convert to xpath
	for _, e := range resultEntries {
		sdcpbPath, err := e.SdcpbPath()
		if err != nil {
			continue
		}
		if len(e.GetSchemaKeys()) > 0 {
			results = append(results, fmt.Sprintf("/%s[%s=", sdcpbPath.ToXPath(false), e.GetSchemaKeys()[0]))
		}
		results = append(results, fmt.Sprintf("/%s", sdcpbPath.ToXPath(false)))
	}
	sort.Strings(results)
	return results
}
