package cmd

import (
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var workspaceCmd = &cobra.Command{
	Use:          "workspace",
	Short:        "workspace based actions",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
