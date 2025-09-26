package cmd

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var pipelineCmd = &cobra.Command{
	Use:          "pipeline",
	Short:        "pipeline based actions",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		optsP = append(optsP, config.WithSuccessfullSchemaLoad())
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	EnableFlagAndDisableFileCompletion(pipelineCmd)
}
