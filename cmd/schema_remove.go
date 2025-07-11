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
		var c *config.Config
		var cd *configdiff.ConfigDiff

		ctx := context.Background()

		opts := config.ConfigOpts{}
		c, err = config.NewConfig(opts)
		if err != nil {
			return err
		}

		log.Infof("Schema Remove")
		cd, err = configdiff.NewConfigDiff(ctx, c, GetWorkspace())
		if err != nil {
			return err
		}

		if schemaDefinitionFile != "" {
			schema, err := utils.GetConfig(schemaDefinitionFile)
			if err != nil {
				return err
			}
			vendor = schema.Spec.Provider
			version = schema.Spec.Version
		}

		err = cd.SchemaRemove(ctx, vendor, version)
		if err != nil {
			return err
		}
		log.Infof("Schema Vendor: %s, Version: %s - Removed Successful", vendor, version)
		return nil
	},
}

func init() {
	schemaCmd.AddCommand(SchemaRemoveCmd)
	SchemaRemoveCmd.PersistentFlags().StringVarP(&schemaDefinitionFile, "schema-def", "f", "", "The KRM that defined the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&vendor, "vendor", "", "", "The vendor name of the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&version, "version", "", "", "The version of the schema")
}
