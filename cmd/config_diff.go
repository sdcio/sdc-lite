package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	difftypeStr  string
	difftype     params.DiffType
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

		dconf := params.NewDiffConfigRaw().SetContextLines(contextLines).SetNoColor(!noColor).SetConfig(outFormatStr).SetPath(path)
		// turn raw config into actual config
		dc, err := dconf.UnRaw()
		if err != nil {
			return err
		}

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cd.InitTargetFolder(ctx)
		if err != nil {
			return err
		}

		// execute the diff
		result, err := cd.GetDiff(ctx, dc)
		if err != nil {
			return err
		}

		fmt.Println(result)

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
	AddPipelineCommandOutputFlags(configDiffCmd)
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
