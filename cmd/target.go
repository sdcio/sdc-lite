package cmd

import (
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var targetCmd = &cobra.Command{
	Use:          "target",
	Short:        "target based actions",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(targetCmd)
	EnableFlagAndDisableFileCompletion(targetCmd)
}
