package config

import "os"

type ConfigPersistentOpts []ConfigPersistentOpt
type ConfigPersistentOpt func(c *ConfigPersistent) error

func WithWorkspaceBasePath(workspaceBasePath string) ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		_, err := os.Stat(workspaceBasePath)
		if err != nil {
			return err
		}
		c.workspaceBasePath = workspaceBasePath
		return nil
	}
}

func WithWorkspaceName(workspacename string) ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		c.workspaceName = workspacename
		return nil
	}
}

func WithSuccessfullSchemaLoad() ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		c.expectSchemaLoadsSuccessful = true
		return nil
	}
}

func WithSchemaDefintionFilePath(schemaDefinitionFile string) ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		c.schemaDefinitionFilePath = schemaDefinitionFile
		return nil
	}
}
