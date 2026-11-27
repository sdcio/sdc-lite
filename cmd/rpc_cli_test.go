package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetCmdGlobals() {
	rpcOutput = false
	jsonOutput = false
	targetName = ""
	outFormatStr = "json"
	path = ""
	outputAll = false
	includeDefaults = false
	noColor = false
	contextLines = 2
	difftypeStr = "side-by-side-patch"
	pipelineFile = ""
	pipelineTarget = ""
	pipelinePersist = false
	optsP = nil
}

func TestCLI_RpcFlagWorksWithoutTarget(t *testing.T) {
	cases := [][]string{
		{"schema", "load", "-f", "schema.yaml", "--rpc"},
		{"config", "load", "--file", "running.json", "--file-format", "json", "--intent-name", "running", "--rpc"},
		{"config", "show", "-a", "-o", "yaml", "--path", "/interface", "--rpc"},
		{"config", "diff", "--type", "patch", "--no-color", "--context", "4", "--rpc"},
		{"config", "blame", "--include-defaults", "--path", "/interface", "--rpc"},
		{"config", "validate", "--rpc"},
	}

	for _, args := range cases {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			resetCmdGlobals()
			rootCmd.SetArgs(args)
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("Execute(%v) unexpected error: %v", args, err)
			}
		})
	}
}

func TestCLI_NonRpcStillRequiresTarget(t *testing.T) {
	for _, args := range [][]string{
		{"schema", "load", "-f", "schema.yaml"},
		{"config", "validate"},
		{"config", "load", "--file", "cr.yaml", "--file-format", "sdc", "--rpc"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			resetCmdGlobals()
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			if err == nil || !strings.Contains(err.Error(), "target") {
				t.Fatalf("Execute(%v) expected target error, got: %v", args, err)
			}
		})
	}
}

func TestCLI_PipelineRunReturnsErrorOnBatchStepFailure(t *testing.T) {
	resetCmdGlobals()
	f := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(f, []byte(`{"jsonrpc":"2.0","id":1,"method":"config-validate","params":{}}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"pipeline", "run", "--pipeline-file", f})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("expected non-nil error on failing batch step")
	}

	resetCmdGlobals()
	okFile := filepath.Join(t.TempDir(), "ok.json")
	if err := os.WriteFile(okFile, []byte(`{"jsonrpc":"2.0","id":1,"method":"pipeline-stop"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"pipeline", "run", "--pipeline-file", okFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected nil error on clean pipeline-stop, got %v", err)
	}
}
