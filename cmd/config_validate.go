package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/data-server/pkg/tree/types"
	"github.com/spf13/cobra"
)

// cconfigValidateCmd represents the validate command
var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "validate config",
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
		valResult, valStats, err := cd.TreeValidate(ctx)
		if err != nil {
			return err
		}
		os.Stderr.WriteString(fmt.Sprintf("Workspace: %s\n", workspaceName))
		if len(valStats.GetCounter()) > 0 {
			fmt.Println("Validations performed:")
			indent := "  "
			// sort the map, by getting the keys first
			keys := make([]types.StatType, 0, len(valStats.GetCounter()))
			for typ := range valStats.GetCounter() {
				keys = append(keys, typ)
			}

			// sorting the keys
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].String() < keys[j].String()
			})
			// printing the stats in the sorted order
			for _, typ := range keys {
				fmt.Printf("%s%s: %d\n", indent, typ.String(), valStats.GetCounter()[typ])
			}
		}

		if !valResult.HasErrors() && !valResult.HasWarnings() {
			fmt.Println("Successful Validated!")
		}

		if valResult.HasErrors() {
			errStrBuilder := &strings.Builder{}
			errStrBuilder.WriteString("Errors:\n")
			for _, errStr := range valResult.ErrorsStr() {
				errStrBuilder.WriteString(errStr)
				errStrBuilder.WriteString("\n")
			}
			fmt.Println(errStrBuilder.String())
		}

		if valResult.HasWarnings() {
			warnStrBuilder := &strings.Builder{}
			warnStrBuilder.WriteString("Errors:\n")
			for _, warnStr := range valResult.ErrorsStr() {
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
