package cmd

import (
	"context"

	configdiff "github.com/henderiw/config-diff/pkg/config-diff"
	"github.com/henderiw/config-diff/pkg/config-diff/config"
	"github.com/spf13/cobra"
)

// SchemaLoadCmd represents the list command
var configLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load config",
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
		return nil
	},
}

func init() {
	configCmd.AddCommand(configLoadCmd)
}
