package config

import "os"

type ConfigPersistentOpts []ConfigPersistentOpt
type ConfigPersistentOpt func(c *ConfigPersistent) error

func WithTargetsBasePath(targetBasePath string) ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		_, err := os.Stat(targetBasePath)
		if err != nil {
			return err
		}
		c.targetBasePath = targetBasePath
		return nil
	}
}

func WithTargetName(targetName string) ConfigPersistentOpt {
	return func(c *ConfigPersistent) error {
		c.targetName = targetName
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
