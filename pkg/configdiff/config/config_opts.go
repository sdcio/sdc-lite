package config

import (
	"os"

	"github.com/sdcio/data-server/pkg/config"
	"github.com/sirupsen/logrus"
)

type ConfigOpts []ConfigOpt
type ConfigOpt func(c *Config) error

func WithLogLevel(loglevel string) ConfigOpt {
	return func(c *Config) error {
		_, err := logrus.ParseLevel(loglevel)
		if err != nil {
			return nil
		}
		c.loglevel = loglevel
		return nil
	}
}

func WithSchemaDefinition(schemaDefinitionPath string) ConfigOpt {
	return func(c *Config) error {
		_, err := os.Stat(schemaDefinitionPath)
		if err != nil {
			return err
		}
		c.SchemaDefPath = schemaDefinitionPath
		return nil
	}
}

func WithCachePath(path string) ConfigOpt {
	return func(c *Config) error {
		c.cachePath = path
		return nil
	}
}

func WithSchemaPathCleanup(cleanup bool) ConfigOpt {
	return func(c *Config) error {
		c.schemaPathCleanup = cleanup
		return nil
	}
}

func WithValidation(validation *config.Validation) ConfigOpt {
	return func(c *Config) error {
		c.validation = validation
		return nil
	}
}
