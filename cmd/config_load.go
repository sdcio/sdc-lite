package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	sdcConfig "github.com/sdcio/config-server/apis/config/v1alpha1"
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

		cdp, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cdp.InitWorkspace(ctx)
		if err != nil {
			return err
		}
		configFormat, err := types.ParseConfigFormat(configurationFileFormatStr)
		if err != nil {
			return err
		}

		var configByte []byte
		// process the input, read from file or stdin
		if configurationFile == "-" {
			// read from stdin
			configByte, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		} else {
			// read config from file
			configByte, err = os.ReadFile(configurationFile)
			if err != nil {
				return err
			}
		}

		var intent *types.Intent

		switch configFormat {
		case types.ConfigFormatJson, types.ConfigFormatJsonIetf, types.ConfigFormatXml:
			intent = types.NewIntent(intentName, priority, treetypes.NewUpdateInsertFlags())
			intent.SetData(configFormat, configByte)
		case types.ConfigFormatSdc:
			sdcC, err := sdcConfig.GetConfigFromFile(configurationFile)
			if err != nil {
				return err
			}

			// create a new config diff instance that we can use to aggregate the multiple path / values from the cr spec
			cd, err := cdp.CopyEmptyConfigDiff(ctx)
			if err != nil {
				return err
			}

			// create the
			intent = types.NewIntent(sdcC.Name, int32(sdcC.Spec.Priority), treetypes.NewUpdateInsertFlags())
			for _, c := range sdcC.Spec.Config {
				jsonConfigVal := c.Value.Raw
				if err != nil {
					return err
				}
				intent.SetData(types.ConfigFormatJson, jsonConfigVal).SetBasePath(c.Path)

				err = cd.TreeLoadData(ctx, intent)
				if err != nil {
					return err
				}
			}

			jsonConf, err := cd.GetJson(false)
			if err != nil {
				return err
			}

			jsonConfByte, err := json.Marshal(jsonConf)
			if err != nil {
				return err
			}

			intent = types.NewIntent(sdcC.Name, int32(sdcC.Spec.Priority), treetypes.NewUpdateInsertFlags())
			intent.SetData(types.ConfigFormatJson, jsonConfByte)
		}

		err = cdp.TreeLoadData(ctx, intent)
		if err != nil {
			return err
		}

		err = cdp.Persist(ctx)
		if err != nil {
			return err
		}
		os.Stderr.WriteString(fmt.Sprintf("Workspace: %s\n", workspaceName))
		fmt.Printf("File: %s - %s - successfully loaded\n", configurationFile, intent)

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
