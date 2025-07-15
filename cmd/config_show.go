package cmd

import (
	"context"
	"fmt"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	"github.com/spf13/cobra"
)

var (
	outFormatStr string
	outputAll    bool // !onlyNewOrUpdates
)

// configLoadCmd represents the list command
var configShowCmd = &cobra.Command{
	Use:          "show",
	Short:        "show config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

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
		err = cd.InitWorkspace(ctx)
		if err != nil {
			return err
		}
		outFormat, err := types.ParseConfigFormat(outFormatStr)
		if err != nil {
			return err
		}

		data, err := cd.TreeGetOutput(ctx, outFormat, !outputAll)
		if err != nil {
			return err
		}
		fmt.Printf("Workspace: %s\n", workspaceName)
		fmt.Println(data)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)

	configCmd.PersistentFlags().StringVarP(&outFormatStr, "out-format", "o", "json", "output format")
	configCmd.PersistentFlags().BoolVarP(&outputAll, "all", "a", false, "return the whole config, not just new and updated values")
}
