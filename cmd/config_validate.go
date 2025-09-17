package cmd

import (
	"fmt"
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/spf13/cobra"
)

// cconfigValidateCmd represents the validate command
var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "validate config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		ctx := cmd.Context()

		rawParam := &params.ConfigValidateRaw{}

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if rpcOutput {
			return pipeline.PipelineAppendStep(os.Stdout, rawParam)
		}

		opts := config.ConfigOpts{}
		out, err := RunFromRaw(ctx, opts, optsP, true, rawParam)
		if err != nil {
			return err
		}

		err = WriteOutput(out)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
	AddRpcOutputFlag(configValidateCmd)
	AddDetailedFlag(configValidateCmd)
	EnableFlagAndDisableFileCompletion(configValidateCmd)
	params.GetCommandRegistry().Register(types.CommandTypeConfigValidate, func() params.RpcRawParams { return params.NewConfigValidateRaw() })
}
