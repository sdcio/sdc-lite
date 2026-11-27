package rpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	sdcTypes "github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

func TestRpcId_NotificationVsNullVsValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantSet  bool
		wantJSON string
	}{
		{
			name:    "omitted id is notification",
			input:   `{"jsonrpc":"2.0","method":"schema-load","params":{}}`,
			wantSet: false,
		},
		{
			name:     "explicit null id is request",
			input:    `{"jsonrpc":"2.0","id":null,"method":"schema-load","params":{}}`,
			wantSet:  true,
			wantJSON: "null",
		},
		{
			name:     "integer id",
			input:    `{"jsonrpc":"2.0","id":42,"method":"schema-load","params":{}}`,
			wantSet:  true,
			wantJSON: "42",
		},
		{
			name:     "string id",
			input:    `{"jsonrpc":"2.0","id":"step-1","method":"schema-load","params":{}}`,
			wantSet:  true,
			wantJSON: `"step-1"`,
		},
		{
			name:    "fractional id rejected",
			input:   `{"jsonrpc":"2.0","id":1.5,"method":"schema-load","params":{}}`,
			wantErr: true,
		},
		{
			name:    "boolean id rejected",
			input:   `{"jsonrpc":"2.0","id":true,"method":"schema-load","params":{}}`,
			wantErr: true,
		},
		{
			name:    "object id rejected",
			input:   `{"jsonrpc":"2.0","id":{},"method":"schema-load","params":{}}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var env rpc.JsonRpcMessageRaw
			err := json.Unmarshal([]byte(tc.input), &env)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %s, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if env.Id.IsSet() != tc.wantSet {
				t.Fatalf("IsSet() = %v, want %v", env.Id.IsSet(), tc.wantSet)
			}
			if tc.wantSet {
				b, err := json.Marshal(env.Id)
				if err != nil {
					t.Fatalf("Marshal(Id) error: %v", err)
				}
				if string(b) != tc.wantJSON {
					t.Fatalf("Marshal(Id) = %s, want %s", string(b), tc.wantJSON)
				}
			}
		})
	}
}

type stubTree struct {
	val any
}

func (s stubTree) ToJson(context.Context, bool) (any, error) { return s.val, nil }
func (s stubTree) ToXML(context.Context, bool, bool, bool, bool) (*etree.Document, error) {
	doc := etree.NewDocument()
	el := doc.CreateElement("interface")
	el.CreateElement("name").SetText("ethernet-1/1")
	return doc, nil
}

func TestJsonRpcResult_SelfDescribingEnvelopes(t *testing.T) {
	ctx := context.Background()

	t.Run("null output omits data field", func(t *testing.T) {
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdInt32(1), sdcTypes.CommandTypeSchemaLoad, "", output.NewNullOutput())
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"jsonrpc":"2.0","id":1,"result":{"method":"schema-load"}}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", string(b), want)
		}
	})

	t.Run("config-diff preserves empty string data", func(t *testing.T) {
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdString("d1"), sdcTypes.CommandTypeConfigDiff, sdcTypes.ConfigFormatJson, output.NewConfigDiffOutput(""))
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"jsonrpc":"2.0","id":"d1","result":{"method":"config-diff","format":"json","data":""}}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", string(b), want)
		}
	})

	t.Run("config-show yaml returns yaml string in data", func(t *testing.T) {
		out := output.NewConfigShowYamlOutput(stubTree{val: map[string]any{"name": "ethernet-1/1"}})
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdInt32(3), sdcTypes.CommandTypeConfigShow, sdcTypes.ConfigFormatYaml, out)
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var parsed struct {
			Result struct {
				Method string `json:"method"`
				Format string `json:"format"`
				Data   string `json:"data"`
			} `json:"result"`
		}
		if err := json.Unmarshal(b, &parsed); err != nil {
			t.Fatalf("expected string data for yaml, got raw %s: %v", string(b), err)
		}
		if parsed.Result.Method != "config-show" || parsed.Result.Format != "yaml" || !strings.Contains(parsed.Result.Data, "name: ethernet-1/1") {
			t.Fatalf("unexpected yaml result: %+v", parsed.Result)
		}
	})

	t.Run("config-show xml returns xml string directly in data", func(t *testing.T) {
		out := output.NewConfigShowXmlOutput(stubTree{})
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdInt32(4), sdcTypes.CommandTypeConfigShow, sdcTypes.ConfigFormatXml, out)
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var parsed struct {
			Result struct {
				Method string `json:"method"`
				Format string `json:"format"`
				Data   string `json:"data"`
			} `json:"result"`
		}
		if err := json.Unmarshal(b, &parsed); err != nil {
			t.Fatalf("expected string data for xml, got raw %s: %v", string(b), err)
		}
		if !strings.Contains(parsed.Result.Data, "<config>") {
			t.Fatalf("expected <config> root in xml data, got %q", parsed.Result.Data)
		}
	})

	t.Run("config-validate uses lowercase json keys", func(t *testing.T) {
		stats := types.NewValidationStats()
		out := output.NewConfigValidateOutput(types.ValidationResults{}, stats)
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdInt32(5), sdcTypes.CommandTypeConfigValidate, "", out)
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), `"Errors"`) || !strings.Contains(string(b), `"errors":[]`) || !strings.Contains(string(b), `"warnings":[]`) || !strings.Contains(string(b), `"stats":{`) {
			t.Fatalf("unexpected validate json: %s", string(b))
		}
	})

	t.Run("config-blame uses protojson encoding", func(t *testing.T) {
		elem := sdcpb.NewBlameTreeElement("interface").SetOwner("running")
		elem.Value = &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "ethernet-1/1"}}
		out := output.NewBlameResultOutput(elem)
		jr := rpc.NewJsonRpcSuccess(rpc.NewRpcIdInt32(6), sdcTypes.CommandTypeConfigBlame, "", out)
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"stringVal":"ethernet-1/1"`) {
			t.Fatalf("expected protojson stringVal in blame output, got %s", string(b))
		}
	})

	t.Run("structured error object with method data", func(t *testing.T) {
		jr := rpc.NewJsonRpcError(rpc.NewRpcIdInt32(7), sdcTypes.CommandTypeConfigShow, rpc.ErrCodeServerError, errors.New("no configuration loaded"))
		b, err := jr.JsonMarshall(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"jsonrpc":"2.0","id":7,"error":{"code":-32000,"message":"no configuration loaded","data":{"method":"config-show"}}}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", string(b), want)
		}
	})
}
