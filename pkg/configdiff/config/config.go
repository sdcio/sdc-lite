package config

import (
	"os"
	"path"

	"github.com/sdcio/data-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/store"
	"github.com/sdcio/sdc-lite/pkg/utils"
)

type Config struct {
	cachePath         string
	schemaPath        string
	schemaStorePath   string
	downloadPath      string
	loglevel          string
	schemaPathCleanup bool
	validation        *config.Validation
	schemaStore       store.Store
}

func NewConfig(opts ...ConfigOpt) (*Config, error) {
	var err error
	// create an instance of the config
	c := &Config{
		schemaPathCleanup: true,
	}

	// apply the provided options
	for _, opt := range opts {
		err = opt(c)
		if err != nil {
			return nil, err
		}
	}

	if c.validation == nil {
		c.validation = config.NewValidationConfig()
	}

	if c.cachePath == "" {
		globalCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}

		// generate the cache directory path
		c.cachePath = path.Join(globalCacheDir, "sdc-lite")
	}

	err = utils.CreateFolder(c.SchemasPath())
	if err != nil {
		return nil, err
	}
	err = utils.CreateFolder(c.DownloadPath())
	if err != nil {
		return nil, err
	}
	err = utils.CreateFolder(c.SchemaStorePath())
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) GetSchemaStore() store.Store {
	return c.schemaStore
}

func (c *Config) SchemasPath() string {
	if c.schemaPath == "" {
		c.schemaPath = path.Join(c.cachePath, "schemas")
	}
	return c.schemaPath
}

func (c *Config) SchemaStorePath() string {
	if c.schemaStorePath == "" {
		c.schemaStorePath = path.Join(c.cachePath, "schemastore")
	}
	return c.schemaStorePath
}

func (c *Config) DownloadPath() string {
	if c.downloadPath == "" {
		c.downloadPath = path.Join(c.cachePath, "downloads")
	}
	return c.downloadPath
}

func (c *Config) LogLevel() string {
	if c.loglevel == "" {
		return "info"
	}
	return c.loglevel
}

func (c *Config) SchemaPathCleanup() bool {
	return c.schemaPathCleanup
}

func (c *Config) Validation() *config.Validation {
	return c.validation
}
