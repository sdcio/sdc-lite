package configdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/configdiff/workspace"
	"github.com/sdcio/config-diff/pkg/schemaclient"
	"github.com/sdcio/config-diff/pkg/schemaloader"
	"github.com/sdcio/config-diff/pkg/types"
	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	loader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/data-server/pkg/tree"
	treeImporter "github.com/sdcio/data-server/pkg/tree/importer"
	treejson "github.com/sdcio/data-server/pkg/tree/importer/json"
	treexml "github.com/sdcio/data-server/pkg/tree/importer/xml"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	schemaSrvConf "github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/store"
	"github.com/sdcio/schema-server/pkg/store/persiststore"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ConfigDiff struct {
	config      *config.Config
	workspace   workspace.Workspace
	tree        *tree.RootEntry
	schemaStore store.Store
}

func NewConfigDiff(ctx context.Context, c *config.Config, wInit workspace.WorkspaceInit) (*ConfigDiff, error) {
	llevel, err := log.ParseLevel(c.LogLevel())
	if err != nil {
		return nil, err
	}
	log.SetLevel(llevel)
	log.SetOutput(os.Stderr)

	w, err := wInit.Init(&workspace.WorkspaceInitParams{
		WorkspaceBasePath: c.WorkspacePath(),
	})
	if err != nil {
		return nil, err
	}
	cd := &ConfigDiff{
		config:    c,
		workspace: w,
	}
	w.HooksEndpointSet(ctx, cd)

	return cd, nil
}

func (c *ConfigDiff) loadSchemaStore(ctx context.Context) (store.Store, error) {
	if c.schemaStore != nil {
		return c.schemaStore, nil
	}
	s, err := persiststore.New(ctx, c.config.SchemaStorePath(), &schemaSrvConf.SchemaPersistStoreCacheConfig{
		WithDescription: false,
	})
	if err != nil {
		return nil, err
	}
	c.schemaStore = s
	return s, nil
}

func (c *ConfigDiff) SchemaRemove(ctx context.Context, vendor string, version string) error {
	if vendor == "" {
		return fmt.Errorf("vendor must not be empty")
	}
	if version == "" {
		return fmt.Errorf("version must not be empty")
	}
	store, err := c.loadSchemaStore(ctx)
	if err != nil {
		return err
	}
	_, err = store.DeleteSchema(ctx, &sdcpb.DeleteSchemaRequest{
		Schema: &sdcpb.Schema{
			Vendor:  vendor,
			Version: version,
		},
	})
	return err
}

func (c *ConfigDiff) SchemasList(ctx context.Context) ([]*sdcpb.Schema, error) {
	store, err := c.loadSchemaStore(ctx)
	if err != nil {
		return nil, err
	}
	lsr, err := store.ListSchema(ctx, &sdcpb.ListSchemaRequest{})
	if err != nil {
		return nil, err
	}

	return lsr.Schema, nil
}

func (c *ConfigDiff) SchemaDownload(ctx context.Context, schemaDefinition []byte) error {
	schemaStore, err := c.loadSchemaStore(ctx)
	if err != nil {
		return err
	}

	schemaDef := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(schemaDefinition, schemaDef); err != nil {
		return err
	}

	sdcpbSchema := &sdcpb.Schema{
		Name:    "",
		Vendor:  schemaDef.Spec.Provider,
		Version: schemaDef.Spec.Version,
	}
	// check if the schema already exists
	schemaExists := schemaStore.HasSchema(store.SchemaKey{Vendor: schemaDef.Spec.Provider, Version: schemaDef.Spec.Version})

	// return in case of existence
	if schemaExists {
		log.Infof("Schema - Vendor: %s, Version: %s - Already Exists, Skip Loading", schemaDef.Spec.Provider, schemaDef.Spec.Version)
		err = c.workspace.SchemaStore(ctx, sdcpbSchema)
		if err != nil {
			return err
		}

		return nil
	}

	schemaLoader, err := loader.NewLoader(
		c.config.DownloadPath(),
		c.config.SchemasPath(),
		schemaloader.NewNopResolver(),
	)
	if err != nil {
		return err
	}

	schemaLoader.AddRef(ctx, schemaDef)
	_, dirExists, err := schemaLoader.GetRef(ctx, schemaDef.Spec.GetKey())
	if err != nil {
		return err
	}
	if !dirExists {
		log.Info("loading...")
		if err := schemaLoader.Load(ctx, schemaDef.Spec.GetKey()); err != nil {
			return err
		}
	}
	rsp, err := schemaStore.CreateSchema(ctx, &sdcpb.CreateSchemaRequest{
		Schema:    sdcpbSchema,
		File:      schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Models,
		Directory: schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Includes,
		Exclude:   schemaDef.Spec.GetNewSchemaBase(c.config.SchemasPath()).Excludes,
	})
	if err != nil {
		return err
	}
	log.Infof("Schema - Vendor: %s, Version: %s - Loaded Successful", rsp.Schema.Vendor, rsp.Schema.Version)

	err = c.workspace.SchemaStore(ctx, rsp.GetSchema())
	if err != nil {
		return err
	}

	// TODO: cleanup Schemas Path
	return err
}

func (c *ConfigDiff) HookPostSchemaSet(ctx context.Context) error {
	err := c.BuildRootTree(ctx)
	if err != nil {
		return err
	}

	err = c.loadWorkspaceIntents(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigDiff) loadWorkspaceIntents(ctx context.Context) error {
	intents, err := c.workspace.IntentsGet()
	if err != nil {
		return err
	}
	for _, intent := range intents {
		err = c.TreeLoadData(ctx, intent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigDiff) BuildRootTree(ctx context.Context) error {
	schemaStore, err := c.loadSchemaStore(ctx)
	if err != nil {
		return err
	}

	schemaId, err := c.workspace.SchemaGet(ctx)
	if err != nil {
		return err
	}

	scb := schemaclient.NewMemSchemaClientBound(schemaStore, schemaId)
	tc := tree.NewTreeContext(scb, "")
	t, err := tree.NewTreeRoot(ctx, tc)
	if err != nil {
		return err
	}
	c.tree = t

	return nil
}

func (c *ConfigDiff) TreeBlame(ctx context.Context, includeDefaults bool) (*sdcpb.BlameTreeElement, error) {
	return c.tree.BlameConfig(includeDefaults)
}

func (c *ConfigDiff) TreeLoadData(ctx context.Context, intentInfo *workspace.IntentInfo) error {
	var err error
	var importer treeImporter.ImportConfigAdapter

	switch intentInfo.GetFormat() {
	case types.ConfigFormatJson, types.ConfigFormatJsonIetf:
		var j any
		err = json.Unmarshal(intentInfo.GetData(), &j)
		if err != nil {
			return err
		}
		importer = treejson.NewJsonTreeImporter(j)

	case types.ConfigFormatXml:
		xmlDoc := etree.NewDocument()
		err := xmlDoc.ReadFromBytes(intentInfo.GetData())
		if err != nil {
			return err
		}
		importer = treexml.NewXmlTreeImporter(&xmlDoc.Element)
	}

	// overwrite running intent with running prio
	if strings.EqualFold(intentInfo.Name, tree.RunningIntentName) {
		intentInfo.Prio = tree.RunningValuesPrio
	}

	err = c.tree.ImportConfig(ctx, nil, importer, intentInfo.GetName(), intentInfo.GetPrio(), intentInfo.GetFlag())
	if err != nil {
		return err
	}

	c.workspace.IntentStore(intentInfo)

	return nil
}

func (c *ConfigDiff) TreeGetOutput(ctx context.Context, format types.ConfigFormat, onlyNewOrUpdated bool) (string, error) {
	var err error
	err = c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return "", err
	}

	result := ""
	switch format {
	case types.ConfigFormatXml:
		x, err := c.tree.ToXML(onlyNewOrUpdated, true, true, false)
		if err != nil {
			return "", err
		}
		x.Indent(2)
		s, err := x.WriteToString()
		if err != nil {
			return "", err
		}
		result = s
		return result, nil
	case types.ConfigFormatJson, types.ConfigFormatJsonIetf:
		var j any
		if format == types.ConfigFormatJson {
			j, err = c.tree.ToJson(onlyNewOrUpdated)
		} else {
			j, err = c.tree.ToJsonIETF(onlyNewOrUpdated)
		}
		if err != nil {
			return "", err
		}

		byteDoc, err := json.MarshalIndent(j, "", " ")
		if err != nil {
			return "", err
		}
		result = string(byteDoc)
		return result, nil
	}
	return "", nil
}

func (c *ConfigDiff) TreeValidate(ctx context.Context) (treetypes.ValidationResults, error) {

	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, err
	}

	return c.tree.Validate(ctx, c.config.Validation()), nil
}
