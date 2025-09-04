package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/spf13/cobra"
)

var (
	outputAll bool // !onlyNewOrUpdates
)

// configLoadCmd represents the list command
var configShowCmd = &cobra.Command{
	Use:          "show",
	Short:        "show config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		scr := params.NewConfigShowConfigRaw().SetAll(outputAll).SetOutputFormat(outFormatStr).SetPath(path)

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if pipelineFile != "" {
			return AppendToPipelineFile(pipelineFile, scr)
		}

		ctx := cmd.Context()

		opts := config.ConfigOpts{}
		out, err := RunFromRaw(ctx, opts, optsP, false, scr)
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
	configCmd.AddCommand(configShowCmd)
	configShowCmd.Flags().StringVarP(&outFormatStr, "out-format", "o", "json", fmt.Sprintf("output formats one of %s", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	configShowCmd.Flags().BoolVarP(&outputAll, "all", "a", false, "return the whole config, not just new and updated values")
	AddPathPersistentFlag(configShowCmd)
	AddPipelineCommandOutputFlags(configShowCmd)
	EnableFlagAndDisableFileCompletion(configShowCmd)

	params.GetCommandRegistry().Register(types.CommandTypeConfigShow, func() params.RpcRawParams { return params.NewConfigShowConfigRaw() })
}
