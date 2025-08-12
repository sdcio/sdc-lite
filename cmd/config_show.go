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
	outputAll bool // !onlyNewOrUpdates
)

// configLoadCmd represents the list command
var configShowCmd = &cobra.Command{
	Use:          "show",
	Short:        "show config",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		outFormat, err = parseConfigFormat()
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		ctx := context.Background()

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

		data, err := cd.TreeGetString(ctx, outFormat, !outputAll)
		if err != nil {
			return err
		}

		fmt.Println(data)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configShowCmd.Flags().StringVarP(&outFormatStr, "out-format", "o", "json", fmt.Sprintf("output formats one of %s", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	configShowCmd.Flags().BoolVarP(&outputAll, "all", "a", false, "return the whole config, not just new and updated values")
	EnableFlagAndDisableFileCompletion(configShowCmd)
}
