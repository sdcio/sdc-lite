package cmd

import (
	"context"

	configdiff "github.com/sdcio/config-diff/pkg/config-diff"
	"github.com/sdcio/config-diff/pkg/config-diff/config"
	"github.com/spf13/cobra"
)

// SchemaListCmd represents the list command
var SchemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "list available schemas",
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

		err = cd.ListSchemas(ctx)
		return err
	},
}

func init() {
	schemaCmd.AddCommand(SchemaListCmd)
}
