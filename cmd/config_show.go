package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/sdcio/sdc-lite/pkg/types"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		scr := params.NewConfigShowConfigRaw().SetAll(outputAll).SetOutputFormat(outFormatStr).SetPath(path)
		scconfig, err := scr.ToConfigShowConfig()
		if err != nil {
			return err
		}

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if pipelineFile != "" {
			pipel := pipeline.NewPipeline(pipelineFile)
			pipel.AppendStep(scr)
			return nil
		}

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

		data, err := cd.TreeGetString(ctx, scconfig)
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
	AddPathPersistentFlag(configShowCmd)
	AddPipelineCommandOutputFlags(configShowCmd)
	EnableFlagAndDisableFileCompletion(configShowCmd)

	params.GetCommandRegistry().Register(types.CommandTypeConfigShow, func() any { return params.NewConfigShowConfigRaw() })
}
