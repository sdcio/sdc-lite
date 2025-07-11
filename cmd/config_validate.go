package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// cconfigValidateCmd represents the validate command
var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "validate config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var c *config.Config
		var cd *configdiff.ConfigDiff

		ctx := context.Background()

		opts := config.ConfigOpts{}
		c, err = config.NewConfig(opts)
		if err != nil {
			return err
		}

		ws := GetWorkspace()
		cd, err = configdiff.NewConfigDiff(ctx, c, ws)
		if err != nil {
			return err
		}

		result, err := cd.TreeValidate(ctx)
		if err != nil {
			return err
		}

		if !result.HasErrors() && !result.HasWarnings() {
			fmt.Println("Successful Validated!")
		}

		if result.HasErrors() {
			errStrBuilder := &strings.Builder{}
			errStrBuilder.WriteString("Errors:\n")
			for _, errStr := range result.ErrorsStr() {
				errStrBuilder.WriteString(errStr)
				errStrBuilder.WriteString("\n")
			}
			fmt.Println(errStrBuilder.String())
		}

		if result.HasWarnings() {
			warnStrBuilder := &strings.Builder{}
			warnStrBuilder.WriteString("Errors:\n")
			for _, warnStr := range result.ErrorsStr() {
				warnStrBuilder.WriteString(warnStr)
				warnStrBuilder.WriteString("\n")
			}
			fmt.Println(warnStrBuilder.String())
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}
