package rawparams_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rawparams"
	"github.com/sdcio/sdc-lite/pkg/types"
)

func TestRawParams_KebabCaseAndUnRaw(t *testing.T) {
	t.Run("diff params serialize and deserialize kebab-case keys", func(t *testing.T) {
		raw := rawparams.NewDiffConfigRaw().
			SetDiffType("side-by-side-patch").
			SetConfig("json").
			SetContextLines(4).
			SetWidth(120).
			SetNoColor(true).
			SetShowHeader(false).
			SetPath("/interface")

		b, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		for _, wantKey := range []string{`"diff-type"`, `"context-lines"`, `"no-color"`, `"no-show-header"`} {
			if !strings.Contains(s, wantKey) {
				t.Fatalf("expected key %s in %s", wantKey, s)
			}
		}

		var decoded rawparams.DiffConfigRaw
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		cmd, err := decoded.UnRaw()
		if err != nil {
			t.Fatal(err)
		}
		dc := cmd.(*params.DiffConfig)
		if dc.GetColor() {
			t.Fatalf("expected GetColor() == false when no-color is true")
		}
		if dc.GetContextLines() != 4 {
			t.Fatalf("expected context lines 4, got %d", dc.GetContextLines())
		}
	})

	t.Run("config-load wires base-path and handles omitted flags", func(t *testing.T) {
		cfgFile := filepath.Join(t.TempDir(), "c.json")
		if err := os.WriteFile(cfgFile, []byte(`{"admin-state":"enable"}`), 0644); err != nil {
			t.Fatal(err)
		}

		input := `{"name":"overlay","prio":50,"format":"json","base-path":"/interface[name=ethernet-1/1]","file":` + mustJSON(cfgFile) + `}`
		var raw rawparams.ConfigLoadRaw
		if err := json.Unmarshal([]byte(input), &raw); err != nil {
			t.Fatal(err)
		}
		cmd, err := raw.UnRaw()
		if err != nil {
			t.Fatalf("UnRaw() failed when flags omitted: %v", err)
		}
		cl := cmd.(*params.ConfigLoad)
		if cl.GetIntent().GetBasePath() != "/interface[name=ethernet-1/1]" {
			t.Fatalf("base-path = %q, want /interface[name=ethernet-1/1]", cl.GetIntent().GetBasePath())
		}
	})

	t.Run("update-insert-flags uses only-intended", func(t *testing.T) {
		f := &rawparams.UpdateInsertFlagsRaw{Delete: true, OnlyIntended: true}
		b, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"only-intended":true`) {
			t.Fatalf("expected only-intended tag, got %s", string(b))
		}
	})

	t.Run("config-show command string matches wire method", func(t *testing.T) {
		raw := rawparams.NewConfigShowConfigRaw().SetOutputFormat("json")
		cmd, err := raw.UnRaw()
		if err != nil {
			t.Fatal(err)
		}
		if cmd.String() != types.CommandTypeConfigShow {
			t.Fatalf("String() = %q, want %q", cmd.String(), types.CommandTypeConfigShow)
		}
	})
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
