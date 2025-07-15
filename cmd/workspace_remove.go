package cmd

import (
	"context"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var workspaceRemoveCmd = &cobra.Command{
	Use:          "remove",
	Short:        "remove existing workspaces",
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

		err = cd.WorkspaceRemove()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceRemoveCmd)
}
