package cmd

import (
	"context"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/utils"
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

		ctx := context.Background()
		log.Infof("Schema - Loading Start")

		opts := config.ConfigOpts{
			// config.WithSchemaDefinition(schemaDefinitionFile),
			config.WithSchemaPathCleanup(schemaPathCleanup),
		}
		optsP = append(optsP, config.WithTargetName(targetName))
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cd.InitTargetFolder(ctx)
		if err != nil {
			return err
		}

		fw := utils.NewFileWrapper(schemaDefinitionFile)
		if err != nil {
			return err
		}

		schemaReader, err := fw.ReadCloser()
		if err != nil {
			return err
		}

		// download the given schema
		_, err = cd.SchemaDownload(ctx, schemaReader)
		if err != nil {
			return err
		}

		err = cd.Persist(ctx)
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
	EnableFlagAndDisableFileCompletion(SchemaLoadCmd)
	err := SchemaLoadCmd.MarkPersistentFlagRequired("schema-def")
	if err != nil {
		log.Error(err)
	}
	err = SchemaLoadCmd.MarkPersistentFlagFilename("schema-def")
	if err != nil {
		log.Error(err)
	}
	err = AddTargetPersistentFlag(SchemaLoadCmd)
	if err != nil {
		log.Error(err)
	}
}
