package cmd

import (
	"context"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	schemaDefinitionFile string
	schemaPathCleanup    bool
)

// SchemaLoadCmd represents the list command
var SchemaLoadCmd = &cobra.Command{
	Use:          "load",
	Short:        "load a schema",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var c *config.Config
		var cd *configdiff.ConfigDiff

		ctx := context.Background()

		opts := config.ConfigOpts{
			// config.WithSchemaDefinition(schemaDefinitionFile),
			config.WithSchemaPathCleanup(schemaPathCleanup),
		}

		log.Infof("Schema - Loading Start")

		c, err = config.NewConfig(opts)
		if err != nil {
			return err
		}

		ws := GetWorkspace()

		cd, err = configdiff.NewConfigDiff(ctx, c, ws)
		if err != nil {
			return err
		}

		// read schema definition fron file
		schemaDefinition, err := os.ReadFile(schemaDefinitionFile)
		if err != nil {
			return err
		}

		// download the given schema
		err = cd.SchemaDownload(ctx, schemaDefinition)
		if err != nil {
			return err
		}

		err = ws.Persist()
		if err != nil {
			return err
		}
		return nil
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
