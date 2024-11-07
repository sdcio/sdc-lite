package cmd

import (
	"context"

	configdiff "github.com/henderiw/config-diff/pkg/config-diff"
	"github.com/henderiw/config-diff/pkg/config-diff/config"
	"github.com/henderiw/config-diff/pkg/utils"
	"github.com/spf13/cobra"
)

// SchemaListCmd represents the list command
var SchemaRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a certain schema",
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

		cd, err = configdiff.NewConfigDiff(c)
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

		err = cd.RemoveSchema(ctx, vendor, version)
		return err
	},
}

func init() {
	schemaCmd.AddCommand(SchemaRemoveCmd)
	SchemaRemoveCmd.PersistentFlags().StringVarP(&schemaDefinitionFile, "schema-def", "f", "", "The KRM that defined the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&vendor, "vendor", "", "", "The vendor name of the schema")
	SchemaRemoveCmd.PersistentFlags().StringVarP(&version, "version", "", "", "The version of the schema")
}
