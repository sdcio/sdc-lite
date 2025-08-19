package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var targetShowCmd = &cobra.Command{
	Use:          "show",
	Short:        "show existing target",
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

		target, err := cd.TargetGet(targetName)
		if err != nil {
			return err
		}

		switch {
		case jsonOutput:
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(target.Export()); err != nil {
				return err
			}
		default:
			fmt.Print(target.StringDetail())
		}
		return nil
	},
}

func init() {
	targetCmd.AddCommand(targetShowCmd)
	err := AddTargetPersistentFlag(targetShowCmd)
	if err != nil {
		log.Error(err)
	}
	EnableFlagAndDisableFileCompletion(targetShowCmd)
}
