package configdiff

import (
	"context"
	"fmt"
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
)

type ConfigDiffPersistence struct {
	*ConfigDiff
	config *config.ConfigPersistent
	target types.Target
}

func NewConfigDiffPersistence(ctx context.Context, c *config.ConfigPersistent) (*ConfigDiffPersistence, error) {
	cd, err := NewConfigDiff(ctx, c.Config)
	if err != nil {
		return nil, err
	}

	cdp := &ConfigDiffPersistence{
		ConfigDiff: cd,
		config:     c,
	}

	//  make sure the targets base folder exists
	err = utils.CreateFolder(c.TargetBasePath())
	if err != nil {
		return nil, err
	}

	return cdp, nil
}

func (c *ConfigDiffPersistence) InitTargetFolder(ctx context.Context) error {
	c.target = *types.NewTarget(c.config)
	c.schema = c.target.GetSchema()
	if c.schema == nil {
		// if the schema is expected to be there, but it is not, throw an error
		if c.config.ExpectSchemaLoadsSuccessful() {
			return fmt.Errorf("expected schema to be loaded from %s and to be present", c.config.SchemaDefinitionFilePath())
		}
		// if the schema does not yet exist, stop any further loading and return without an error
		return nil
	}

	err := c.buildRootTree(ctx)
	if err != nil {
		return err
	}

	err = c.intentsLoad(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigDiffPersistence) intentsLoad(ctx context.Context) (err error) {
	// retrieve intents from target
	intents := c.target.GetIntents()

	// load each intent into the tree
	for _, intent := range intents {
		err = c.TreeLoadData(ctx, intent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigDiffPersistence) TreeLoadData(ctx context.Context, intent *types.Intent) error {
	err := c.target.AddIntent(intent)
	if err != nil {
		return err
	}
	return c.ConfigDiff.TreeLoadData(ctx, intent)
}

func (c *ConfigDiffPersistence) SchemaDownload(ctx context.Context, schemaDefinition []byte) (*sdcpb.Schema, error) {
	schema, err := c.ConfigDiff.SchemaDownload(ctx, schemaDefinition)

	c.schema = schema
	c.target.SetSchema(schema)

	return c.schema, err
}

func (c *ConfigDiffPersistence) TargetGet(name string) (*types.Target, error) {
	targetConfig, err := config.NewConfigPersistent(config.ConfigOpts{}, config.ConfigPersistentOpts{config.WithTargetName(name), config.WithTargetsBasePath(c.config.TargetBasePath())})
	if err != nil {
		return nil, err
	}
	return types.NewTarget(targetConfig), nil
}

func (c *ConfigDiffPersistence) TargetList() (types.Targets, error) {
	entries, err := os.ReadDir(c.config.TargetBasePath())
	if err != nil {
		return nil, err
	}

	result := types.Targets{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		targetConfig, err := config.NewConfigPersistent(config.ConfigOpts{}, config.ConfigPersistentOpts{config.WithTargetName(e.Name()), config.WithTargetsBasePath(c.config.TargetBasePath())})
		if err != nil {
			return nil, err
		}
		wsi := types.NewTarget(targetConfig)
		result.Add(wsi)
	}
	return result, nil
}

func (c *ConfigDiffPersistence) TargetRemove() error {
	path := c.config.TargetPath()

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", path)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}
	log.Infof("target %s - successfully removed", c.config.TargetName())
	return nil
}

func (c *ConfigDiffPersistence) Persist(ctx context.Context) error {
	return c.target.Persist()
}
