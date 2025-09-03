package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	outFormatStr string
	targetName   string
	path         string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "config based actions",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		optsP = append(optsP, config.WithSuccessfullSchemaLoad(), config.WithTargetName(targetName))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	EnableFlagAndDisableFileCompletion(configCmd)
	err := AddTargetPersistentFlag(configCmd)
	if err != nil {
		log.Error(err)
	}
}

func AddPathPersistentFlag(c *cobra.Command) error {
	c.PersistentFlags().StringVarP(&path, "path", "p", "", "limit the output to given branch (xpath expression e.g. \"/interface[name=ethernet-1/1]\") ")

	// Register autocompletion for the flag
	err := c.RegisterFlagCompletionFunc("path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		target, err := cmd.Flags().GetString("target")
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		}

		opts := config.ConfigOpts{}
		optsP = append(optsP, config.WithTargetName(target))
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		}

		cdp, err := configdiff.NewConfigDiffPersistence(cmd.Context(), c)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		}

		err = cdp.InitTargetFolder(cmd.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		}

		f, err := os.OpenFile("/tmp/trace", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		fmt.Fprintf(f, "toComplete: %s\n", toComplete)

		results := cdp.GetPathCompletions(cmd.Context(), toComplete)

		fmt.Fprintf(f, "result: %s\n", strings.Join(results, ", "))
		return results, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	})
	return err
}
