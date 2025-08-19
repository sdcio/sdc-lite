package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

var detailed bool

// datastoreCmd represents the datastore command
var targetListCmd = &cobra.Command{
	Use:          "list",
	Short:        "list existing target",
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

		targets, err := cd.TargetList()
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			fmt.Println("no targets found")
			return nil
		}

		switch {
		case jsonOutput:
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(targets.Export()); err != nil {
				return err
			}
		case detailed:
			fmt.Print(targets.StringDetail())
		default:
			fmt.Print(targets.String())
		}
		return nil
	},
}

func init() {
	targetCmd.AddCommand(targetListCmd)
	targetListCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "list detailed target information")
	EnableFlagAndDisableFileCompletion(targetListCmd)
}
