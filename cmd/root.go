package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	baseConfigFile = ""
	inFormat       string
	outFormat      string
	outputAll      bool // !onlyNewOrUpdates
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "config-diff",
	Short: "A CLI tool to interact with NOS configs based on YANG schemas",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&baseConfigFile, "base-config", "b", "", "base config, usually running")
	rootCmd.PersistentFlags().StringVarP(&outFormat, "in-format", "if", "json", "input format")
	rootCmd.PersistentFlags().StringVarP(&inFormat, "out-format", "of", "json", "output format")
	rootCmd.PersistentFlags().BoolVarP(&outputAll, "all", "a", false, "return the whole config, not just new and updated values")
}
