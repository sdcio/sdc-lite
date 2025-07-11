package cmd

import "github.com/spf13/cobra"

// datastoreCmd represents the datastore command
var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "config based actions",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
