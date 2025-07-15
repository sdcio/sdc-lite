package cmd

import (
	"context"
	"fmt"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// datastoreCmd represents the datastore command
var workspaceListCmd = &cobra.Command{
	Use:          "list",
	Short:        "list existing workspaces",
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

		targets, err := cd.WorkspaceList()
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			fmt.Println("no targets found")
			return nil
		}

		fmt.Println(targets.String())
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
}
