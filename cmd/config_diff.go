package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/spf13/cobra"
)

// cconfigValidateCmd represents the validate command
var configDiffCmd = &cobra.Command{
	Use:          "diff",
	Short:        "diff config with running",
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

		err = cd.DiffWithRunning(ctx)
		if err != nil {
			return err
		}

		os.Stderr.WriteString(fmt.Sprintf("Workspace: %s\n", workspaceName))

		return nil
	},
}

func init() {
	configCmd.AddCommand(configDiffCmd)
}
