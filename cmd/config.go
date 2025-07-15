package cmd

import (
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "config based actions",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		optsP = append(optsP, config.WithSuccessfullSchemaLoad())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
