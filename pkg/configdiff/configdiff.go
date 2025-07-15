package configdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
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
	tree        *tree.RootEntry
	schema      *sdcpb.Schema
	schemaStore store.Store
}

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

func (c *ConfigDiff) HasSchema() bool {
	return c.schema != nil
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

func (c *ConfigDiff) SetSchema(s *sdcpb.Schema) {
	c.schema = s
}

func (c *ConfigDiff) SchemaDownload(ctx context.Context, schemaDefinition []byte) (*sdcpb.Schema, error) {
	schemaStore, err := c.loadSchemaStore(ctx)
	if err != nil {
		return nil, err
	}

	schemaDef := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(schemaDefinition, schemaDef); err != nil {
		return nil, err
	}

	// check if the schema already exists
	schemaExists := schemaStore.HasSchema(store.SchemaKey{Vendor: schemaDef.Spec.Provider, Version: schemaDef.Spec.Version})

	sdcpbSchema := &sdcpb.Schema{
		Name:    "",
		Vendor:  schemaDef.Spec.Provider,
		Version: schemaDef.Spec.Version,
	}

	// return in case of existence
	if schemaExists {
		log.Infof("Schema - Vendor: %s, Version: %s - Already Exists, Skip Loading", schemaDef.Spec.Provider, schemaDef.Spec.Version)
		return sdcpbSchema, nil
	}

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
		if err := schemaLoader.Load(ctx, schemaDef.Spec.GetKey()); err != nil {
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
	log.Infof("Schema - Vendor: %s, Version: %s - Loaded Successful", rsp.Schema.Vendor, rsp.Schema.Version)

	c.SetSchema(sdcpbSchema)

	// TODO: cleanup Schemas Path
	return nil, err
}

func (c *ConfigDiff) BuildRootTree(ctx context.Context) error {
	schemaStore, err := c.loadSchemaStore(ctx)
	if err != nil {
		return err
	}

	scb := schemaclient.NewMemSchemaClientBound(schemaStore, c.schema)
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

func (c *ConfigDiff) TreeLoadData(ctx context.Context, intentInfo *types.IntentInfo) error {
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

func (c *ConfigDiff) TreeValidate(ctx context.Context) (treetypes.ValidationResults, *treetypes.ValidationStatOverall, error) {
	err := c.tree.FinishInsertionPhase(ctx)
	if err != nil {
		return nil, nil, err
	}
	valResult, valStat := c.tree.Validate(ctx, c.config.Validation())
	return valResult, valStat, nil
}
