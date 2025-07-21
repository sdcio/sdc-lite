package cmd

import (
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var optsP = config.ConfigPersistentOpts{}
var workspaceName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "config-diff",
	Short: "A CLI tool to interact with NOS configs based on YANG schemas",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		optsP = append(optsP, config.WithWorkspaceName(workspaceName))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&workspaceName, "workspace-name", "w", "default", "name of the workspace to work in")
}
