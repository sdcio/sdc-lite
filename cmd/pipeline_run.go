package cmd

import (
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

// pipelineRunCmd
var pipelineRunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run pipeline",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		pipe := pipeline.NewPipeline(pipelineFile)
		err := pipe.Run(ctx)
		return err
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineRunCmd.Flags().StringVarP(&pipelineFile, "pipeline-file", "f", "", "pipeline file to run")
	EnableFlagAndDisableFileCompletion(configShowCmd)
}
