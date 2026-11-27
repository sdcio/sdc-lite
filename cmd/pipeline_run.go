package cmd

import (
	"fmt"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

var (
	pipelineFile    string
	pipelineTarget  string
	pipelinePersist bool
)

var pipelineRunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run pipeline",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if strings.TrimSpace(pipelineFile) == "" {
			return fmt.Errorf("--pipeline-file is required")
		}

		if pipelinePersist && pipelineTarget == "" {
			return fmt.Errorf("--persist requires a target (-t)")
		}

		var pipeOpts []pipeline.PipelineOption
		if pipelineTarget != "" {
			pipeOpts = append(pipeOpts, pipeline.WithTarget(append(optsP, config.WithTargetName(pipelineTarget))))
		}
		if pipelinePersist {
			pipeOpts = append(pipeOpts, pipeline.WithPersist())
		}
		pipe := pipeline.NewPipeline(pipelineFile, pipeOpts...)
		outputChan := make(chan *pipeline.PipelineResult)

		go func() {
			pipe.Run(ctx, outputChan)
		}()

		var fatalErr error
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context deadline exceeded")
			case out, ok := <-outputChan:
				if !ok {
					return fatalErr
				}
				if out.IsError() && out.IsFatal() && fatalErr == nil {
					fatalErr = out.GetError()
				}
				if !out.HasResponse() {
					continue
				}
				var jr *rpc.JsonRpcResult
				if out.IsError() {
					jr = rpc.NewJsonRpcError(out.GetId(), out.GetMethod(), out.GetErrorCode(), out.GetError())
				} else {
					jr = rpc.NewJsonRpcSuccess(out.GetId(), out.GetMethod(), out.GetFormat(), out.GetOutput())
				}
				data, err := jr.JsonMarshall(ctx)
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
	pipelineRunCmd.Flags().StringVarP(&pipelineTarget, "target", "t", "", "start from this existing target's schema and intents (default: empty, in-memory)")
	pipelineRunCmd.Flags().BoolVar(&pipelinePersist, "persist", false, "write the resulting state back to the target after a successful run (requires -t)")
	EnableFlagAndDisableFileCompletion(pipelineRunCmd)
}
