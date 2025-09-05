package cmd

import (
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/spf13/cobra"
)

// pipelineRunCmd
var pipelineRunCmd = &cobra.Command{
	Use:          "run",
	Short:        "run pipeline",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// if strings.TrimSpace(pipelineFile) == "-" {
		// 	inputFifo := "/tmp/mytool_in"
		// 	outputFifo := "/tmp/mytool_out"

		// 	_ = os.Remove(inputFifo)
		// 	_ = os.Remove(outputFifo)
		// 	err := syscall.Mkfifo(inputFifo, 0600)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	err = syscall.Mkfifo(outputFifo, 0600)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// spawn another instance of this tool in background
		// 	self, _ := os.Executable()
		// 	cmd := exec.Command(self, "pipeline", "run", "--pipeline-file", inputFifo)

		// 	// DETACH: new session, ignore signals from parent
		// 	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 		Setsid: true,
		// 	}

		// 	if err := cmd.Start(); err != nil {
		// 		return err
		// 	}

		// 	// print the variable for the shell
		// 	fmt.Printf("export MY_STDIN=%s\n", inputFifo)
		// 	fmt.Printf("export MY_STDOUT=%s\n", outputFifo)
		// 	return nil
		// }

		pipe := pipeline.NewPipeline(pipelineFile)
		err := pipe.Run(ctx)
		return err
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineRunCmd)
	pipelineRunCmd.Flags().StringVarP(&pipelineFile, "pipeline-file", "f", "", "pipeline file to run")
	EnableFlagAndDisableFileCompletion(configShowCmd)
}
