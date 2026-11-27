package pipeline

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sdcio/sdc-lite/pkg/configdiff/commandregistry"
	_ "github.com/sdcio/sdc-lite/pkg/configdiff/rawparams"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
)

func TestPersistWithoutTargetFails(t *testing.T) {
	f := filepath.Join(t.TempDir(), "p.json")
	if err := os.WriteFile(f, nil, 0644); err != nil {
		t.Fatal(err)
	}
	out := make(chan *PipelineResult, 4)
	NewPipeline(f, WithPersist()).Run(context.Background(), out)

	res := <-out
	if res == nil || !res.IsError() || !res.IsFatal() || !strings.Contains(res.GetError().Error(), "requires a target") {
		t.Fatalf("expected fatal 'requires a target' error, got %+v", res)
	}
}

func TestPipeline_NotificationsAndPipelineStopResponse(t *testing.T) {
	f := filepath.Join(t.TempDir(), "p.json")
	content := strings.Join([]string{
		// Notification pipeline-stop omitted here; first test Request pipeline-stop with string id
		`{"jsonrpc":"2.0","id":"stop-1","method":"pipeline-stop"}`,
		// Step after pipeline-stop must not run
		`{"jsonrpc":"2.0","id":2,"method":"unknown-method","params":{}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	out := make(chan *PipelineResult, 8)
	NewPipeline(f).Run(context.Background(), out)

	var results []*PipelineResult
	for r := range out {
		results = append(results, r)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result before stop, got %d", len(results))
	}
	r := results[0]
	if r.IsError() || !r.HasResponse() || r.GetMethod() != types.CommandTypePipelineStop {
		t.Fatalf("unexpected pipeline-stop result: %+v", r)
	}
	jr := rpc.NewJsonRpcSuccess(r.GetId(), r.GetMethod(), r.GetFormat(), r.GetOutput())
	b, err := jr.JsonMarshall(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"jsonrpc":"2.0","id":"stop-1","result":{"method":"pipeline-stop"}}`
	if string(b) != want {
		t.Fatalf("got %s, want %s", string(b), want)
	}
}

func TestPipeline_NotificationStopProducesNoResponse(t *testing.T) {
	f := filepath.Join(t.TempDir(), "p.json")
	if err := os.WriteFile(f, []byte(`{"jsonrpc":"2.0","method":"pipeline-stop"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := make(chan *PipelineResult, 4)
	NewPipeline(f).Run(context.Background(), out)

	for r := range out {
		t.Fatalf("expected no response for notification pipeline-stop, got %+v", r)
	}
}

func TestPipeline_BatchErrorCodesAndNotificationFailure(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantCode    int
		wantHasResp bool
	}{
		{
			name:        "parse error",
			line:        `{not json`,
			wantCode:    rpc.ErrCodeParseError,
			wantHasResp: true,
		},
		{
			name:        "invalid request missing jsonrpc version",
			line:        `{"id":1,"method":"config-validate","params":{}}`,
			wantCode:    rpc.ErrCodeInvalidRequest,
			wantHasResp: true,
		},
		{
			name:        "method not found request",
			line:        `{"jsonrpc":"2.0","id":1,"method":"no-such-method","params":{}}`,
			wantCode:    rpc.ErrCodeMethodNotFound,
			wantHasResp: true,
		},
		{
			name:        "method not found notification has no stdout response but is fatal in batch",
			line:        `{"jsonrpc":"2.0","method":"no-such-method","params":{}}`,
			wantCode:    rpc.ErrCodeMethodNotFound,
			wantHasResp: false,
		},
		{
			name:        "invalid params array",
			line:        `{"jsonrpc":"2.0","id":1,"method":"config-validate","params":[1,2]}`,
			wantCode:    rpc.ErrCodeInvalidParams,
			wantHasResp: true,
		},
		{
			name:        "execution error without loaded config",
			line:        `{"jsonrpc":"2.0","id":1,"method":"config-validate","params":{}}`,
			wantCode:    rpc.ErrCodeServerError,
			wantHasResp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "p.json")
			// Add a trailing pipeline-stop with id 99 to verify fail-fast aborts before step 2.
			body := tc.line + "\n" + `{"jsonrpc":"2.0","id":99,"method":"pipeline-stop"}` + "\n"
			if err := os.WriteFile(f, []byte(body), 0644); err != nil {
				t.Fatal(err)
			}

			out := make(chan *PipelineResult, 8)
			NewPipeline(f).Run(context.Background(), out)

			var got []*PipelineResult
			for r := range out {
				got = append(got, r)
			}
			if len(got) != 1 {
				t.Fatalf("expected fail-fast with 1 result, got %d", len(got))
			}
			res := got[0]
			if !res.IsError() || !res.IsFatal() {
				t.Fatalf("expected fatal error result, got %+v", res)
			}
			if res.GetErrorCode() != tc.wantCode {
				t.Fatalf("error code = %d, want %d", res.GetErrorCode(), tc.wantCode)
			}
			if res.HasResponse() != tc.wantHasResp {
				t.Fatalf("HasResponse() = %v, want %v", res.HasResponse(), tc.wantHasResp)
			}
		})
	}
}

func TestPipeline_FIFOStaysAliveAcrossStepErrors(t *testing.T) {
	fifoPath := filepath.Join(t.TempDir(), "pipe.fifo")
	if err := syscall.Mkfifo(fifoPath, 0600); err != nil {
		t.Skipf("mkfifo not supported: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan *PipelineResult, 8)
	go NewPipeline(fifoPath).Run(ctx, out)

	writeLine := func(line string) {
		t.Helper()
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
		if err != nil {
			t.Fatalf("open fifo for write: %v", err)
		}
		if _, err := f.WriteString(line + "\n"); err != nil {
			_ = f.Close()
			t.Fatalf("write fifo: %v", err)
		}
		_ = f.Close()
	}

	// Writer 1 sends a failing step; FIFO must emit non-fatal error and stay alive.
	writeLine(`{"jsonrpc":"2.0","id":1,"method":"config-validate","params":{}}`)
	r1 := <-out
	if r1 == nil || !r1.IsError() || r1.IsFatal() {
		t.Fatalf("expected non-fatal FIFO step error, got %+v", r1)
	}

	// Writer 2 sends pipeline-stop; FIFO must still be running and respond cleanly.
	writeLine(`{"jsonrpc":"2.0","id":2,"method":"pipeline-stop"}`)
	r2 := <-out
	if r2 == nil || r2.IsError() || r2.GetMethod() != types.CommandTypePipelineStop {
		t.Fatalf("expected clean pipeline-stop after FIFO error, got %+v", r2)
	}

	for extra := range out {
		t.Fatalf("unexpected trailing FIFO output: %+v", extra)
	}
}

func TestPipelineAppendStep_Generates32BitSafeId(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "rpc-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	factory, err := commandregistry.GetCommandRegistry().GetCommandFactory(types.CommandTypeConfigValidate)
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		if err := PipelineAppendStep(f, factory()); err != nil {
			t.Fatal(err)
		}
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		var msg struct {
			Id int64 `json:"id"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("unmarshal %s: %v", line, err)
		}
		if msg.Id < 1 || msg.Id > math.MaxInt32 {
			t.Fatalf("id %d outside 1..MaxInt32", msg.Id)
		}
	}
}

func TestOpenRPCSpec_CoversAllMethods(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "openrpc.json"))
	if err != nil {
		t.Fatalf("read openrpc.json: %v", err)
	}
	var spec struct {
		OpenRPC string `json:"openrpc"`
		Methods []struct {
			Name           string `json:"name"`
			ParamStructure string `json:"paramStructure"`
		} `json:"methods"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatalf("invalid JSON in openrpc.json: %v", err)
	}
	if spec.OpenRPC != "1.2.6" {
		t.Fatalf("openrpc = %q, want 1.2.6", spec.OpenRPC)
	}

	inSpec := map[string]bool{}
	for _, m := range spec.Methods {
		if m.ParamStructure != "by-name" {
			t.Fatalf("method %s paramStructure = %q, want by-name", m.Name, m.ParamStructure)
		}
		inSpec[m.Name] = true
	}

	wantMethods := []types.CommandType{
		types.CommandTypeSchemaLoad,
		types.CommandTypeConfigLoad,
		types.CommandTypeConfigShow,
		types.CommandTypeConfigDiff,
		types.CommandTypeConfigBlame,
		types.CommandTypeConfigValidate,
		types.CommandTypePipelineStop,
	}
	for _, m := range wantMethods {
		if !inSpec[string(m)] {
			t.Errorf("method %q missing from openrpc.json", m)
		}
	}
}
