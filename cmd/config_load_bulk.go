package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	"github.com/spf13/cobra"
)

var (
	configurationFiles = []string{}
)

// configLoadCmd represents the list command
var configLoadBulkCmd = &cobra.Command{
	Use:          "bulk",
	Short:        "load config bulk",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		ctx := context.Background()

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cdp, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cdp.InitTargetFolder(ctx)
		if err != nil {
			return err
		}

		for _, configFile := range configurationFiles {
			var intent *types.Intent

			configByte, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}

			sdcC, err := LoadSDCConfigCR(configByte)
			if err != nil {
				return err
			}
			intent, err = ConvertSDCConfigToInternalIntent(ctx, cdp.ConfigDiff, sdcC)
			if err != nil {
				return err
			}

			err = cdp.TreeLoadData(ctx, intent)
			if err != nil {
				return err
			}
			fmt.Printf("File: %s - %s - successfully loaded\n", configFile, intent)
		}

		err = cdp.Persist(ctx)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	configLoadCmd.AddCommand(configLoadBulkCmd)
	configLoadBulkCmd.Flags().StringSliceVar(&configurationFiles, "files", nil, "The sdc configuration files to load")
	EnableFlagAndDisableFileCompletion(configLoadBulkCmd)
}
