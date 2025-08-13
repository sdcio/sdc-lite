package cmd

import (
	"context"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// SchemaListCmd represents the list command
var SchemaRemoveCmd = &cobra.Command{
	Use:          "remove",
	Short:        "remove a certain schema",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := context.Background()

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		log.Infof("Schema Remove")
		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}

		if schemaDefinitionFile != "" {
			schema, err := utils.GetSchemaConfig(schemaDefinitionFile)
			if err != nil {
				return err
			}
			schemaVendor = schema.Spec.Provider
			schemaVersion = schema.Spec.Version
		}

		err = cd.SchemaRemove(ctx, schemaVendor, schemaVersion)
		if err != nil {
			return err
		}
		log.Infof("Schema Vendor: %s, Version: %s - Removed Successful", schemaVendor, schemaVersion)
		return nil
	},
}

func init() {
	schemaCmd.AddCommand(SchemaRemoveCmd)
	SchemaRemoveCmd.PersistentFlags().StringVarP(&schemaDefinitionFile, "schema-def", "f", "", "The KRM that defined the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&schemaVendor, "vendor", "", "", "The vendor name of the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&schemaVersion, "version", "", "", "The version of the schema")
	EnableFlagAndDisableFileCompletion(SchemaRemoveCmd)
}
