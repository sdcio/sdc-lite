package cmd

import "github.com/spf13/cobra"

var (
	schemaVendor  string
	schemaVersion string
)

// datastoreCmd represents the datastore command
var schemaCmd = &cobra.Command{
	Use:          "schema",
	Short:        "schema based actions",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
