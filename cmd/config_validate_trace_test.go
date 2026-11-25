package cmd_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
	schemaSrvConf "github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/store/persiststore"
	"github.com/sdcio/sdc-lite/cmd"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"github.com/sirupsen/logrus"
)

const cliTestYang = `module sdc-lite-must {
  yang-version 1.1;
  namespace "urn:sdcio:lite:must";
  prefix "slm";

  container network {
    leaf max-mtu {
      type uint32;
    }
    list interface {
      key "name";
      leaf name {
        type string;
      }
      leaf mtu {
        type uint32;
        must ". <= ../../max-mtu" {
          error-message "mtu exceeds max-mtu";
        }
      }
      leaf description {
        type string;
        must "string-length(.) > 0" {
          error-message "description required";
        }
      }
    }
  }
}
`

func seedTargetFixture(t *testing.T, cacheHome, targetName, jsonConfig string) string {
	t.Helper()
	ctx := context.Background()

	yangDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(yangDir, "sdc-lite-must.yang"), []byte(cliTestYang), 0o600); err != nil {
		t.Fatalf("write yang: %v", err)
	}

	cacheRoot := filepath.Join(cacheHome, "sdc-lite")
	cp, err := config.NewConfigPersistent(
		config.ConfigOpts{config.WithCachePath(cacheRoot)},
		config.ConfigPersistentOpts{config.WithTargetName(targetName)},
	)
	if err != nil {
		t.Fatalf("NewConfigPersistent: %v", err)
	}

	vendor, version := "testvendor", "1.0.0"
	storePath := filepath.Join(cp.SchemaStorePath(), vendor, version)
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("mkdir storePath: %v", err)
	}
	st, err := persiststore.New(ctx, storePath, &schemaSrvConf.SchemaPersistStoreCacheConfig{WithDescription: false}, false)
	if err != nil {
		t.Fatalf("persiststore.New: %v", err)
	}
	scRef := &sdcpb.Schema{
		Name:    "",
		Vendor:  vendor,
		Version: version,
	}
	if _, err := st.CreateSchema(ctx, &sdcpb.CreateSchemaRequest{
		Schema: scRef,
		File:   []string{yangDir},
	}); err != nil {
		_ = st.Close()
		t.Fatalf("CreateSchema: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close persiststore: %v", err)
	}

	tgt := types.NewTarget(cp)
	tgt.SetSchema(scRef)
	intent := types.NewIntent("intent1", 10, dsTypes.NewUpdateInsertFlags()).
		SetData(types.ConfigFormatJson, []byte(jsonConfig))
	if err := tgt.AddIntent(intent); err != nil {
		t.Fatalf("AddIntent: %v", err)
	}
	if err := tgt.Persist(); err != nil {
		t.Fatalf("Persist target: %v", err)
	}
	return cp.TargetPath()
}

func readJSONLines(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var lines []map[string]any
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("invalid jsonl in %s: %v\nline: %s", path, err, sc.Text())
		}
		lines = append(lines, m)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return lines
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr

	outCh := make(chan string, 1)
	errCh := make(chan string, 1)
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, rOut)
		outCh <- b.String()
	}()
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, rErr)
		errCh <- b.String()
	}()

	runErr := cmd.ExecuteContext(context.Background(), args)

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = origOut
	os.Stderr = origErr
	logrus.SetOutput(origErr)

	return <-outCh, <-errCh, runErr
}

func TestConfigValidateCLI_TraceMustFilterAndExplicitDir(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	seedTargetFixture(t, cacheHome, "dev1", `{
		"network": {
			"max-mtu": 1500,
			"interface": [
				{
					"name": "eth0",
					"mtu": 9000,
					"description": "uplink"
				}
			]
		}
	}`)

	traceDir := filepath.Join(t.TempDir(), "custom-traces")
	_, stderr, err := runCLI(t,
		"config", "validate",
		"-t", "dev1",
		"--trace-must", "/network/interface/mtu",
		"--trace-dir", traceDir,
	)
	if err != nil {
		t.Fatalf("config validate returned unexpected error: %v (stderr=%s)", err, stderr)
	}
	if !strings.Contains(stderr, traceDir) {
		t.Fatalf("expected stderr to print trace directory %q, got: %s", traceDir, stderr)
	}

	indexEntries := readJSONLines(t, filepath.Join(traceDir, "index.jsonl"))
	if len(indexEntries) != 1 {
		t.Fatalf("expected 1 entry in index.jsonl, got %d: %+v", len(indexEntries), indexEntries)
	}
	idx := indexEntries[0]
	if idx["path"] != "/network/interface[name=eth0]/mtu" || idx["status"] != "failed" {
		t.Fatalf("unexpected index entry: %+v", idx)
	}
	traceFile, _ := idx["file"].(string)
	if traceFile == "" {
		t.Fatalf("missing file field in index entry: %+v", idx)
	}

	events := readJSONLines(t, filepath.Join(traceDir, traceFile))
	if len(events) < 5 {
		t.Fatalf("expected full trace JSONL events, got %d: %+v", len(events), events)
	}
	if events[0]["kind"] != "must_header" {
		t.Fatalf("expected first JSONL line to be must_header, got %+v", events[0])
	}
	if events[len(events)-1]["kind"] != "must_end" {
		t.Fatalf("expected last JSONL line to be must_end, got %+v", events[len(events)-1])
	}

	kinds := map[string]bool{}
	for _, ev := range events {
		k, _ := ev["kind"].(string)
		kinds[k] = true
	}
	for _, wantKind := range []string{"must_header", "must_program", "vm_start", "vm_instr", "adapter_navigate", "adapter_get_value", "vm_instr_done", "vm_end", "must_end"} {
		if !kinds[wantKind] {
			t.Fatalf("missing event kind %q in trace file; saw kinds=%v", wantKind, kinds)
		}
	}
}

func TestConfigValidateCLI_TraceFailingOnlyDefaultTargetCacheDir(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	targetPath := seedTargetFixture(t, cacheHome, "dev2", `{
		"network": {
			"max-mtu": 1500,
			"interface": [
				{
					"name": "eth0",
					"mtu": 9000,
					"description": "passing-desc"
				}
			]
		}
	}`)

	_, stderr, err := runCLI(t,
		"config", "validate",
		"-t", "dev2",
		"--trace-failing-only",
	)
	if err != nil {
		t.Fatalf("config validate returned unexpected error: %v (stderr=%s)", err, stderr)
	}

	tracesBase := filepath.Join(targetPath, "traces")
	runDirs, err := os.ReadDir(tracesBase)
	if err != nil || len(runDirs) != 1 {
		t.Fatalf("expected 1 timestamped trace dir under %s, got entries=%v err=%v", tracesBase, runDirs, err)
	}
	defaultTraceDir := filepath.Join(tracesBase, runDirs[0].Name())
	if !strings.Contains(stderr, defaultTraceDir) {
		t.Fatalf("expected stderr to mention default trace dir %q, got: %s", defaultTraceDir, stderr)
	}

	indexEntries := readJSONLines(t, filepath.Join(defaultTraceDir, "index.jsonl"))
	if len(indexEntries) != 1 {
		t.Fatalf("expected failing-only to record 1 entry (excluding passing description must), got %d: %+v", len(indexEntries), indexEntries)
	}
	if indexEntries[0]["path"] != "/network/interface[name=eth0]/mtu" || indexEntries[0]["status"] != "failed" {
		t.Fatalf("unexpected failing-only index entry: %+v", indexEntries[0])
	}
}

func TestConfigValidateCLI_TraceDirWithoutFilterErrors(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	seedTargetFixture(t, cacheHome, "dev3", `{"network":{"max-mtu":1500}}`)

	_, _, err := runCLI(t,
		"config", "validate",
		"-t", "dev3",
		"--trace-dir", filepath.Join(t.TempDir(), "unused"),
	)
	if err == nil || !strings.Contains(err.Error(), "--trace-must") {
		t.Fatalf("expected error explaining --trace-dir requires --trace-must or --trace-failing-only, got %v", err)
	}
}

func TestConfigValidateCLI_RepeatableExpressionFilterAndWildcard(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	seedTargetFixture(t, cacheHome, "dev4", `{
		"network": {
			"max-mtu": 1500,
			"interface": [
				{
					"name": "eth0",
					"mtu": 1400,
					"description": "uplink"
				}
			]
		}
	}`)

	exprDir := filepath.Join(t.TempDir(), "expr-traces")
	_, _, err := runCLI(t,
		"config", "validate",
		"-t", "dev4",
		"--trace-must", "string-length",
		"--trace-dir", exprDir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	exprIndex := readJSONLines(t, filepath.Join(exprDir, "index.jsonl"))
	if len(exprIndex) != 1 || exprIndex[0]["path"] != "/network/interface[name=eth0]/description" || exprIndex[0]["status"] != "passed" {
		t.Fatalf("expected 1 passed trace matching expression substring, got %+v", exprIndex)
	}

	allDir := filepath.Join(t.TempDir(), "all-traces")
	_, _, err = runCLI(t,
		"config", "validate",
		"-t", "dev4",
		"--trace-must", "*",
		"--trace-dir", allDir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	allIndex := readJSONLines(t, filepath.Join(allDir, "index.jsonl"))
	if len(allIndex) != 2 {
		t.Fatalf("expected wildcard --trace-must '*' to capture both must evaluations, got %d: %+v", len(allIndex), allIndex)
	}

	missingTargetDir := filepath.Join(cacheHome, "sdc-lite", "targets", "nonexistent")
	_, _, err = runCLI(t,
		"config", "validate",
		"-t", "nonexistent",
		"--trace-failing-only",
	)
	if err == nil {
		t.Fatalf("expected error for missing target")
	}
	if _, statErr := os.Stat(missingTargetDir); !os.IsNotExist(statErr) {
		t.Fatalf("expected missing target directory %s not to be created on failed validation, got statErr=%v", missingTargetDir, statErr)
	}
}

