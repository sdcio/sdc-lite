package configdiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	"github.com/sdcio/config-diff/pkg/utils"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

type ConfigDiffPersistence struct {
	*ConfigDiff
	config  *config.ConfigPersistent
	intents types.IntentInfos
}

func NewConfigDiffPersistence(ctx context.Context, c *config.ConfigPersistent) (*ConfigDiffPersistence, error) {
	cd, err := NewConfigDiff(ctx, c.Config)
	if err != nil {
		return nil, err
	}

	cdp := &ConfigDiffPersistence{
		ConfigDiff: cd,
		config:     c,
		intents:    types.IntentInfos{},
	}
	return cdp, nil
}

func (c *ConfigDiffPersistence) InitWorkspace(ctx context.Context) error {
	err := utils.CreateFolder(c.config.WorkspacePath())
	if err != nil {
		return err
	}

	err = c.schemaLoad()
	// if the schema loading did not work, it must probably still be created,
	// so return the cdp instance anyways.
	if err != nil {
		return nil
	}

	if c.config.ExpectSchemaLoadsSuccessful() {
		if !c.HasSchema() {
			return fmt.Errorf("expected schema to be loaded from %s and to be present", c.config.SchemaDefinitionFilePath())
		}
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

func (c *ConfigDiffPersistence) intentsLoad(ctx context.Context) error {
	matches, err := filepath.Glob(c.config.ConfigFileGlob())
	if err != nil {
		return err
	}
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		ii := &types.IntentInfo{}
		err = json.Unmarshal(data, ii)
		if err != nil {
			return err
		}
		err = c.TreeLoadData(ctx, ii)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigDiffPersistence) schemaLoad() error {
	var err error
	c.schema, err = utils.SchemaLoadSdcpbSchemaFile(c.config.SchemaDefinitionFilePath())
	return err
}

func (c *ConfigDiffPersistence) schemaPersist() error {
	schemaByte, err := protojson.Marshal(c.schema)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.config.SchemaDefinitionFilePath(), schemaByte, 0644)
	return err
}

func (c *ConfigDiffPersistence) Persist(ctx context.Context) error {
	err := c.schemaPersist()
	if err != nil {
		return err
	}
	err = c.intentPersist()
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigDiffPersistence) intentPersist() error {
	for name, content := range c.intents {
		data, err := json.Marshal(content)
		if err != nil {
			return err
		}
		err = os.WriteFile(c.config.ConfigFileName(name), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigDiffPersistence) TreeLoadData(ctx context.Context, intentInfo *types.IntentInfo) error {
	c.intents.AddIntentInfo(intentInfo)
	return c.ConfigDiff.TreeLoadData(ctx, intentInfo)
}

func (c *ConfigDiffPersistence) SchemaDownload(ctx context.Context, schemaDefinition []byte) (*sdcpb.Schema, error) {
	var err error
	c.schema, err = c.ConfigDiff.SchemaDownload(ctx, schemaDefinition)
	return c.schema, err
}

func (c *ConfigDiffPersistence) WorkspacesList() (types.WorkspaceInfos, error) {
	entries, err := os.ReadDir(c.config.WorkspaceBasePath())
	if err != nil {
		return nil, err
	}

	result := types.WorkspaceInfos{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		wsi := types.NewWorkspaceInfo(e.Name(), path.Join(c.config.WorkspaceBasePath(), e.Name()))
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
	log.Infof("successfully remove workspace %s", c.config.WorkspaceName())
	return nil
}
