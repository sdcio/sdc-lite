package cmd

import "github.com/spf13/cobra"

var (
	baseConfigFile = ""
	inFormat       string
	outFormat      string
	outputAll      bool // !onlyNewOrUpdates
)

// datastoreCmd represents the datastore command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config based actions",
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().StringVarP(&baseConfigFile, "base-config", "b", "", "base config, usually running")
	configCmd.PersistentFlags().StringVarP(&inFormat, "out-format", "o", "json", "output format")
	configCmd.PersistentFlags().BoolVarP(&outputAll, "all", "a", false, "return the whole config, not just new and updated values")
}
