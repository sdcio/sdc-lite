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

var (
	ErrTargetUndefined = fmt.Errorf("target not defined")
)

type ConfigPersistent struct {
	*Config
	targetBasePath              string
	targetName                  string
	expectSchemaLoadsSuccessful bool
	// schemaDefinitionFilePath if set, overwrites the default target layout
	schemaDefinitionFilePath string
}

func NewConfigPersistent(opts ConfigOpts, optsP ConfigPersistentOpts) (*ConfigPersistent, error) {
	c, err := NewConfig(opts...)
	if err != nil {
		return nil, err
	}

	cp := &ConfigPersistent{
		Config: c,
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

func (c *ConfigPersistent) TargetBasePath() string {
	if c.targetBasePath == "" {
		c.targetBasePath = path.Join(c.cachePath, "targets")
	}
	return c.targetBasePath
}

func (c *ConfigPersistent) TargetPath() string {
	return path.Join(c.TargetBasePath(), c.targetName)
}

func (c *ConfigPersistent) SchemaDefinitionFilePath() string {
	if c.schemaDefinitionFilePath == "" {
		c.schemaDefinitionFilePath = path.Join(c.TargetPath(), SchemaFileName)
	}
	return c.schemaDefinitionFilePath
}

func (c *ConfigPersistent) ConfigFileGlob() string {
	return path.Join(c.TargetPath(), IntentFileGlob)
}

func (c *ConfigPersistent) ConfigFileName(intentName string) string {
	return path.Join(c.TargetPath(), fmt.Sprintf(IntentFileName, intentName))
}

func (c *ConfigPersistent) TargetName() string {
	return c.targetName
}
