package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	difftypeStr  string
	contextLines int
	noColor      bool
)

// cconfigValidateCmd represents the validate command
var configDiffCmd = &cobra.Command{
	Use:          "diff",
	Short:        "diff config with running",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := cmd.Context()

		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		rawParam := params.NewDiffConfigRaw().SetContextLines(contextLines).SetNoColor(!noColor).SetConfig(outFormatStr).SetPath(path)

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
	configCmd.AddCommand(configDiffCmd)
	configDiffCmd.Flags().StringVar(&difftypeStr, "type", "side-by-side-patch", fmt.Sprintf("difftype, one of %s", strings.Join(params.DiffTypeList.StringSlice(), ", ")))
	configDiffCmd.Flags().IntVar(&contextLines, "context", 2, "number of context lines in patch based diffs")
	configDiffCmd.Flags().BoolVar(&noColor, "no-color", false, "non colorized output")
	configDiffCmd.Flags().StringVarP(&outFormatStr, "out-format", "o", "json", fmt.Sprintf("output formats one of %s", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	AddPathPersistentFlag(configDiffCmd)
	AddRpcOutputFlag(configDiffCmd)
	EnableFlagAndDisableFileCompletion(configDiffCmd)

	// Register autocompletion for the diff type flag
	err := configDiffCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return params.DiffTypeList.StringSlice(), cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		logrus.Error(err)
	}
	// Register autocompletion for the out format flag
	err = configDiffCmd.RegisterFlagCompletionFunc("out-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return types.ConfigFormatsList.StringSlice(), cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		logrus.Error(err)
	}
}
