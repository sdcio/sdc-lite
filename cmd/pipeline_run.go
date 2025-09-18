package cmd

import (
	"fmt"

	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

var pipelineFile string

// pipelineRunCmd
var pipelineRunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run pipeline",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		pipe := pipeline.NewPipeline(pipelineFile)
		outputChan := make(chan *pipeline.PipelineResult)

		var outFormat rpc.OutFormat

		switch {
		case detailed:
			outFormat = rpc.OutFormatDetailed
		case jsonOutput:
			outFormat = rpc.OutFormatJson
		default:
			outFormat = rpc.OutFormatString
		}

		go func() {
			pipe.Run(ctx, outputChan)
		}()
		var jr *rpc.JsonRpcResult
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context deadline exceeded")
			case out, ok := <-outputChan:
				if !ok {
					// channel closed
					return nil
				}
				jr = rpc.NewJsonRpcResult(out.GetId(), nil, out.GetOutput())
				if out.IsError() {
					// error received
					jr = rpc.NewJsonRpcResult(out.GetId(), out.GetError(), nil)
				}
				data, err := jr.JsonMarshall(outFormat)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			}
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineRunCmd.Flags().StringVar(&pipelineFile, "pipeline-file", "", "specify the pipeline that is to be run.")
	pipelineRunCmd.RegisterFlagCompletionFunc("pipeline-file", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"*.json"}, cobra.ShellCompDirectiveFilterFileExt
	})
	AddDetailedFlag(pipelineRunCmd)
	EnableFlagAndDisableFileCompletion(configShowCmd)
}
