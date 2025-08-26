package cmd

import (
	"context"
	"fmt"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var targetRemoveCmd = &cobra.Command{
	Use:          "remove",
	Short:        "remove existing target",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := context.Background()

		opts := config.ConfigOpts{}
		optsP = append(optsP, config.WithTargetName(targetName))
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}

		fmt.Printf("Target: %s\n", c.TargetName())
		err = cd.TargetRemove()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	targetCmd.AddCommand(targetRemoveCmd)
	EnableFlagAndDisableFileCompletion(targetRemoveCmd)
	err := AddTargetPersistentFlag(targetRemoveCmd)
	if err != nil {
		log.Error(err)
	}
}
