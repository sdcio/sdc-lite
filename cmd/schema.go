package cmd

import "github.com/spf13/cobra"

var (
	vendor  string
	version string
)

// datastoreCmd represents the datastore command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "schema based actions",
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
