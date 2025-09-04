package cmd

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		fmt.Println("unimplemented")
		return nil

		// ctx := cmd.Context()

		// opts := config.ConfigOpts{}

		// intents := []params.RpcRawParams{}

		// for _, configFile := range configurationFiles {
		// 	var intent *params.ConfigLoadRaw

		// 	fw := utils.NewFileWrapper(configFile)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	input, err := fw.Bytes()
		// 	if err != nil {
		// 		return err
		// 	}

		// 	sdcC, err := LoadSDCConfigCR(input)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	intent, err = ConvertSDCConfigToInternalIntent(ctx, cdp.ConfigDiff, sdcC)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	intents = append(intents, intent)

		// }
		// out, err := RunFromRaw(ctx, opts, optsP, true, intents...)
		// if err != nil {
		// 	return err
		// }
		// err = WriteOutput(out)
		// if err != nil {
		// 	return err
		// }

		// return nil
	},
}

func init() {
	configLoadCmd.AddCommand(configLoadBulkCmd)
	configLoadBulkCmd.Flags().StringSliceVar(&configurationFiles, "files", nil, "The sdc configuration files to load")
	EnableFlagAndDisableFileCompletion(configLoadBulkCmd)
}
