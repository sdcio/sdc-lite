package configdiff

import (
	"context"
	"fmt"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
)

type ConfigDiffPersistence struct {
	*ConfigDiff
	config    *config.ConfigPersistent
	workspace types.Workspace
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
	return cdp, nil
}

func (c *ConfigDiffPersistence) InitWorkspace(ctx context.Context) error {

	c.workspace = *types.NewWorkspace(c.config)
	err := c.workspace.Create()
	if err != nil {
		return err
	}

	c.schema = c.workspace.GetSchema()
	if c.schema == nil {
		// if the schema is expected to be there, but it is not, throw an error
		if c.config.ExpectSchemaLoadsSuccessful() {
			return fmt.Errorf("expected schema to be loaded from %s and to be present", c.config.SchemaDefinitionFilePath())
		}
		// if the schema does not yet exist, stop any further loading and return without an error
		return nil
	}

	err = c.BuildRootTree(ctx)
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
	// retrieve intents from workspace
	intents := c.workspace.GetIntents()

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
	err := c.workspace.AddIntent(intent)
	if err != nil {
		return err
	}
	return c.ConfigDiff.TreeLoadData(ctx, intent)
}

func (c *ConfigDiffPersistence) SchemaDownload(ctx context.Context, schemaDefinition []byte) (*sdcpb.Schema, error) {
	schema, err := c.ConfigDiff.SchemaDownload(ctx, schemaDefinition)

	c.schema = schema
	c.workspace.SetSchema(schema)

	return c.schema, err
}

func (c *ConfigDiffPersistence) WorkspaceList() (types.Workspaces, error) {
	entries, err := os.ReadDir(c.config.WorkspaceBasePath())
	if err != nil {
		return nil, err
	}

	result := types.Workspaces{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		workspaceConfig, err := config.NewConfigPersistent(config.ConfigOpts{}, config.ConfigPersistentOpts{config.WithWorkspaceName(e.Name()), config.WithWorkspaceBasePath(c.config.WorkspaceBasePath())})
		if err != nil {
			return nil, err
		}
		wsi := types.NewWorkspace(workspaceConfig)
		result.Add(wsi)
	}
	return result, nil
}

func (c *ConfigDiffPersistence) WorkspaceRemove() error {
	path := c.config.WorkspacePath()

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
	log.Infof("successfully remove target %s", c.config.WorkspaceName())
	return nil
}

func (c *ConfigDiffPersistence) Persist(ctx context.Context) error {
	return c.workspace.Persist()
}
