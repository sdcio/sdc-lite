package output

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/sdcio/data-server/pkg/tree/api"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type stubUpdateParent struct {
	path *sdcpb.Path
}

func (s stubUpdateParent) SdcpbPath() *sdcpb.Path {
	return s.path
}

type stubTreeToXPath struct {
	gotOnlyNewOrUpdated bool
	leaves              api.LeafVariantSlice
}

func (s *stubTreeToXPath) GetHighestPrecedence(_ context.Context, onlyNewOrUpdated bool, _ bool, _ bool) api.LeafVariantSlice {
	s.gotOnlyNewOrUpdated = onlyNewOrUpdated
	return s.leaves
}

func newTestLeaf(xpath string, val *sdcpb.TypedValue) *api.LeafEntry {
	p, err := sdcpb.ParsePath(xpath)
	if err != nil {
		panic(err)
	}
	upd := treetypes.NewUpdate(stubUpdateParent{path: p}, val, 50, "intent1", 0)
	return &api.LeafEntry{Update: upd}
}

func TestConfigShowXPathOutput_ToStringEscapesNewlines(t *testing.T) {
	ctx := context.Background()
	stub := &stubTreeToXPath{
		leaves: api.LeafVariantSlice{
			newTestLeaf("/system/banner/login-banner", &sdcpb.TypedValue{
				Value: &sdcpb.TypedValue_StringVal{StringVal: "line1\nline2\r\nline3\n"},
			}),
			newTestLeaf("/interface[name=ethernet-1/1]/admin-state", &sdcpb.TypedValue{
				Value: &sdcpb.TypedValue_StringVal{StringVal: "enable"},
			}),
		},
	}

	out := NewConfigShowXPathOutput(stub).SetOnlyNewOrUpdated(true)
	got, err := out.ToString(ctx)
	if err != nil {
		t.Fatalf("ToString() unexpected error: %v", err)
	}

	if !stub.gotOnlyNewOrUpdated {
		t.Fatalf("expected onlyNewOrUpdated=true, got false")
	}

	want := "/interface[name=ethernet-1/1]/admin-state -> enable\n" +
		"/system/banner/login-banner -> line1\\nline2\\r\\nline3\\n\n"
	if got != want {
		t.Fatalf("ToString() mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestConfigShowXPathOutput_WriteToJsonPreservesRawNewlines(t *testing.T) {
	ctx := context.Background()
	stub := &stubTreeToXPath{
		leaves: api.LeafVariantSlice{
			newTestLeaf("/system/banner/login-banner", &sdcpb.TypedValue{
				Value: &sdcpb.TypedValue_StringVal{StringVal: "hello\nworld"},
			}),
		},
	}

	out := NewConfigShowXPathOutput(stub).SetOnlyNewOrUpdated(false)
	var buf bytes.Buffer
	if err := out.WriteToJson(ctx, &buf); err != nil {
		t.Fatalf("WriteToJson() unexpected error: %v", err)
	}

	if stub.gotOnlyNewOrUpdated {
		t.Fatalf("expected onlyNewOrUpdated=false, got true")
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() unexpected error: %v", err)
	}
	if got := decoded["/system/banner/login-banner"]; got != "hello\nworld" {
		t.Fatalf("decoded value = %q, want %q", got, "hello\nworld")
	}
}
