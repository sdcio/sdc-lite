package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	"github.com/spf13/cobra"
)

var (
	difftypeStr  string
	difftype     types.DiffType
	contextLines int
	noColor      bool
)

// cconfigValidateCmd represents the validate command
var configDiffCmd = &cobra.Command{
	Use:          "diff",
	Short:        "diff config with running",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		difftype, err = types.ParseDiffType(difftypeStr)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := context.Background()

		fmt.Fprintf(os.Stderr, "Workspace: %s\n", workspaceName)

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cd.InitWorkspace(ctx)
		if err != nil {
			return err
		}

		result, err := cd.GetDiff(ctx, types.NewDiffConfig(difftype).SetContextLines(contextLines).SetColor(!noColor))
		if err != nil {
			return err
		}

		fmt.Println(result)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configDiffCmd)
	configDiffCmd.Flags().StringVar(&difftypeStr, "type", "side-by-side-patch", fmt.Sprintf("difftype, one of %s", strings.Join(types.DiffTypeList.StringSlice(), ", ")))
	configDiffCmd.Flags().IntVar(&contextLines, "context", 2, "number of context lines in patch based diffs")
	configDiffCmd.Flags().BoolVar(&noColor, "no-color", false, "non colorized output")
}
