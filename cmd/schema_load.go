package cmd

import (
	"context"

	configdiff "github.com/sdcio/config-diff/pkg/config-diff"
	"github.com/sdcio/config-diff/pkg/config-diff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	schemaDefinitionFile string
	schemaPathCleanup    bool
)

// SchemaLoadCmd represents the list command
var SchemaLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load a schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var c *config.Config
		var cd *configdiff.ConfigDiff

		ctx := context.Background()

		opts := config.ConfigOpts{
			config.WithSchemaDefinition(schemaDefinitionFile),
			config.WithSchemaPathCleanup(schemaPathCleanup),
		}

		c, err = config.NewConfig(opts)
		if err != nil {
			return err
		}

		cd, err = configdiff.NewConfigDiff(c)
		if err != nil {
			return err
		}
		err = cd.DownloadSchema(ctx, schemaDefinitionFile)
		return err
	},
}

func init() {
	schemaCmd.AddCommand(SchemaLoadCmd)
	SchemaLoadCmd.PersistentFlags().StringVarP(&schemaDefinitionFile, "schema-def", "f", "", "The krm that defined the schema")
	SchemaLoadCmd.PersistentFlags().BoolVarP(&schemaPathCleanup, "cleanup", "c", true, "Cleanup the Schemas directory after loading the schema")
	err := SchemaLoadCmd.MarkPersistentFlagRequired("schema-def")
	if err != nil {
		log.Error(err)
	}
	err = SchemaLoadCmd.MarkPersistentFlagFilename("schema-def")
	if err != nil {
		log.Error(err)
	}
}
