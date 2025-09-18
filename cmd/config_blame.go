package cmd

import (
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rawparams"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

var includeDefaults bool

// cconfigValidateCmd represents the validate command
var configBlameCmd = &cobra.Command{
	Use:          "blame",
	Short:        "blame config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		rawParam := rawparams.NewConfigBlameParamsRaw()
		rawParam.SetPath(path).SetIncludeDefaults(includeDefaults)

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if rpcOutput {
			return pipeline.PipelineAppendStep(os.Stdout, rawParam)
		}

		opts := config.ConfigOpts{}
		out, err := RunFromRaw(ctx, opts, optsP, false, rawParam)
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
	configCmd.AddCommand(configBlameCmd)
	configBlameCmd.Flags().BoolVar(&includeDefaults, "include-defaults", false, "include the schema based default values in the output")
	AddPathPersistentFlag(configBlameCmd)
	AddRpcOutputFlag(configBlameCmd)
	EnableFlagAndDisableFileCompletion(configBlameCmd)
}
