package cmd

import (
	"fmt"
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	cdtypes "github.com/sdcio/sdc-lite/pkg/types"
	"github.com/spf13/cobra"
)

// cconfigValidateCmd represents the validate command
var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "validate config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		ctx := cmd.Context()

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
		valResult, valStats, err := cd.TreeValidate(ctx)
		if err != nil {
			return err
		}

		vs := cdtypes.NewValidationStatsOutput(targetName, valResult, valStats)

		err = WriteOutput(vs)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
	AddPipelineCommandOutputFlags(configValidateCmd)
	EnableFlagAndDisableFileCompletion(configValidateCmd)
}
