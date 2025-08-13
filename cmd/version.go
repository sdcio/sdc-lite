package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	// import path to your main package
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nCommit: %s\n", main.version, main.commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
