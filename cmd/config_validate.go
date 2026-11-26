package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rawparams"
	"github.com/sdcio/sdc-lite/pkg/musttrace"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

var (
	traceMustFilters []string
	traceFailingOnly bool
	traceDir         string
)

func prepareMustTraceSink(cmd *cobra.Command, opts config.ConfigOpts) (*musttrace.FileSink, error) {
	traceMustChanged := cmd.Flags().Changed("trace-must")
	traceFailingChanged := cmd.Flags().Changed("trace-failing-only") && traceFailingOnly
	traceEnabled := traceMustChanged || traceFailingChanged
	if !traceEnabled {
		if traceDir != "" {
			return nil, fmt.Errorf("--trace-dir requires --trace-must or --trace-failing-only")
		}
		return nil, nil
	}

	resolvedDir := traceDir
	if resolvedDir == "" {
		cp, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return nil, err
		}
		resolvedDir = filepath.Join(cp.TargetPath(), "traces", time.Now().UTC().Format("20060102T150405.000000000Z"))
	}
	return musttrace.NewFileSink(musttrace.Options{
		Dir:         resolvedDir,
		Filters:     traceMustFilters,
		FailingOnly: traceFailingOnly,
	})
}

var configValidateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "validate config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		errOut := cmd.ErrOrStderr()
		fmt.Fprintf(errOut, "Target: %s\n", targetName)

		ctx := cmd.Context()
		rawParam := &rawparams.ConfigValidateRaw{}

		if rpcOutput {
			return pipeline.PipelineAppendStep(os.Stdout, rawParam)
		}

		opts := config.ConfigOpts{}
		sink, err := prepareMustTraceSink(cmd, opts)
		if err != nil {
			return err
		}
		if sink != nil {
			opts = append(opts, config.WithMustTracer(sink))
		}

		out, runErr := RunFromRaw(ctx, opts, optsP, true, rawParam)
		if runErr != nil {
			return runErr
		}
		if sink != nil {
			summary, closeErr := sink.Close()
			if closeErr != nil {
				return closeErr
			}
			fmt.Fprintf(errOut, "Must traces: %d evaluation(s) written to %s\n", summary.Written, summary.Dir)
		}

		return WriteOutput(ctx, out)
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
	AddRpcOutputFlag(configValidateCmd)
	AddDetailedFlag(configValidateCmd)
	configValidateCmd.Flags().StringArrayVar(&traceMustFilters, "trace-must", nil, "trace must-statement evaluations matching path, keyless path, or expression substring (use \"*\" for all)")
	configValidateCmd.Flags().BoolVar(&traceFailingOnly, "trace-failing-only", false, "only write traces for failing or errored must-statement evaluations")
	configValidateCmd.Flags().StringVar(&traceDir, "trace-dir", "", "directory for must-statement JSONL traces (defaults to <target-cache-dir>/traces/<timestamp>)")
	EnableFlagAndDisableFileCompletion(configValidateCmd)
}
