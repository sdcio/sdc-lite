package config

import (
	"fmt"
	"path"
)

const (
	SchemaFileName = "schema.json"
	IntentFileName = "intent-%s.json"
	IntentFileGlob = "intent-*.json"
)

type ConfigPersistent struct {
	*Config
	workspaceBasePath           string
	workspaceName               string
	expectSchemaLoadsSuccessful bool
	// schemaDefinitionFilePath if set, overwrites the default workspace layout
	schemaDefinitionFilePath string
}

func NewConfigPersistent(opts ConfigOpts, optsP ConfigPersistentOpts) (*ConfigPersistent, error) {
	c, err := NewConfig(opts)
	if err != nil {
		return nil, err
	}

	cp := &ConfigPersistent{
		Config:        c,
		workspaceName: "default",
	}

	// apply the provided options
	for _, opt := range optsP {
		err := opt(cp)
		if err != nil {
			return nil, err
		}
	}
	return cp, nil
}

func (c *ConfigPersistent) ExpectSchemaLoadsSuccessful() bool {
	return c.expectSchemaLoadsSuccessful
}

func (c *ConfigPersistent) WorkspaceBasePath() string {
	if c.workspaceBasePath == "" {
		c.workspaceBasePath = path.Join(c.cachePath, "workspace")
	}
	return c.workspaceBasePath
}

func (c *ConfigPersistent) WorkspacePath() string {
	return path.Join(c.WorkspaceBasePath(), c.workspaceName)
}

func (c *ConfigPersistent) SchemaDefinitionFilePath() string {
	if c.schemaDefinitionFilePath != "" {
		return c.schemaDefinitionFilePath
	}
	return path.Join(c.WorkspacePath(), SchemaFileName)
}

func (c *ConfigPersistent) ConfigFileGlob() string {
	return path.Join(c.WorkspacePath(), IntentFileGlob)
}

func (c *ConfigPersistent) ConfigFileName(intentName string) string {
	return path.Join(c.WorkspacePath(), fmt.Sprintf(IntentFileName, intentName))
}

func (c *ConfigPersistent) WorkspaceName() string {
	return c.workspaceName
}
