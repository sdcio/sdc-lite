package cmd

import (
	"context"
	"os"
	"sync"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var optsP = config.ConfigPersistentOpts{}
var jsonOutput bool
var rpcOutput bool
var completionOnce sync.Once

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sdc-lite",
	Short: "A CLI tool to interact with NOS configs based on YANG schemas",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := ExecuteContext(context.Background(), nil); err != nil {
		log.Fatal(err)
	}
}

// ExecuteContext runs the CLI with explicit arguments (or os.Args[1:] when args
// is nil) and resets per-invocation flag state across commands.
func ExecuteContext(ctx context.Context, args []string) error {
	completionOnce.Do(func() {
		rootCmd.AddCommand(makeCompletionCmd(rootCmd))
	})
	optsP = nil
	resetCommandFlags(rootCmd)
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

func resetCommandFlags(c *cobra.Command) {
	resetFlagSet := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			if sliceVal, ok := f.Value.(pflag.SliceValue); ok {
				_ = sliceVal.Replace(nil)
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		})
	}
	resetFlagSet(c.Flags())
	resetFlagSet(c.PersistentFlags())
	for _, sub := range c.Commands() {
		resetCommandFlags(sub)
	}
}

func AddDetailedFlag(c *cobra.Command) error {
	c.Flags().BoolVarP(&detailed, "detailed", "d", false, "show detailed output")
	return nil
}

func AddRpcOutputFlag(c *cobra.Command) error {
	c.Flags().BoolVar(&rpcOutput, "rpc", false, "return the jsonrpc content for the command")
	return nil
}

func AddTargetPersistentFlag(c *cobra.Command) error {
	c.PersistentFlags().StringVarP(&targetName, "target", "t", "", "the target to use")
	err := c.MarkPersistentFlagRequired("target")
	if err != nil {
		return err
	}

	// Register autocompletion for the flag
	err = c.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		entries, err := os.ReadDir(c.TargetBasePath())
		if err != nil {
			log.Fatal(err)
		}

		result := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				result = append(result, entry.Name())
			}
		}

		return result, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return err
	}

	return nil
}

func EnableFlagAndDisableFileCompletion(cmd *cobra.Command) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			// Skip if already set
			if !f.Changed {
				completions = append(completions, "--"+f.Name)
			}
		})
		// skip just --help
		if len(completions) == 1 && completions[0] == "--help" {
			completions = nil
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Return output in JSON instead of text")
}
