package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	"github.com/spf13/cobra"
)

var configurationFileFormatStr string
var configurationFile string
var intentName string
var priority int32

// configLoadCmd represents the list command
var configLoadCmd = &cobra.Command{
	Use:          "load",
	Short:        "load config",
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
		configFormat, err := types.ParseConfigFormat(configurationFileFormatStr)
		if err != nil {
			return err
		}
		// read config from file
		config, err := os.ReadFile(configurationFile)
		if err != nil {
			return err
		}
		intentInfo := types.NewIntent(intentName, priority, treetypes.NewUpdateInsertFlags())
		intentInfo.SetData(configFormat, config)

		err = cd.TreeLoadData(ctx, intentInfo)
		if err != nil {
			return err
		}

		err = cd.Persist(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Workspace: %s\n", workspaceName)
		fmt.Printf("File: %s - %s - successfully loaded\n", configurationFile, intentInfo)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configLoadCmd)
	configLoadCmd.Flags().StringVar(&configurationFile, "file", "", "The configuration file to load")
	configLoadCmd.Flags().StringVar(&configurationFileFormatStr, "file-format", "", fmt.Sprintf("The format of the config to be loaded [ %s ]", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	configLoadCmd.Flags().Int32Var(&priority, "priority", 500, "The defined priority of the configuration")
	configLoadCmd.Flags().StringVar(&intentName, "intent-name", "", "The name of the configuration intent")
}
