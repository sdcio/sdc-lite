package params

import (
	"context"
	"testing"

	"github.com/beevik/etree"
	"github.com/sdcio/data-server/pkg/tree/api"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type spyConfigShowEntry struct {
	gotOnlyNewOrUpdated bool
}

func (s *spyConfigShowEntry) ToJson(_ context.Context, onlyNewOrUpdated bool) (any, error) {
	s.gotOnlyNewOrUpdated = onlyNewOrUpdated
	return map[string]any{"k": "v"}, nil
}

func (s *spyConfigShowEntry) ToJsonIETF(_ context.Context, onlyNewOrUpdated bool) (any, error) {
	s.gotOnlyNewOrUpdated = onlyNewOrUpdated
	return map[string]any{"m:k": "v"}, nil
}

func (s *spyConfigShowEntry) ToXML(_ context.Context, onlyNewOrUpdated bool, _ bool, _ bool, _ bool) (*etree.Document, error) {
	s.gotOnlyNewOrUpdated = onlyNewOrUpdated
	return etree.NewDocument(), nil
}

func (s *spyConfigShowEntry) GetHighestPrecedence(_ context.Context, onlyNewOrUpdated bool, _ bool, _ bool) api.LeafVariantSlice {
	s.gotOnlyNewOrUpdated = onlyNewOrUpdated
	return nil
}

func (s *spyConfigShowEntry) GetDeletes(_ context.Context, _ bool) (treetypes.DeleteEntriesList, error) {
	return nil, nil
}

type stubExecutor struct {
	entry ConfigShowInterface
}

func (s *stubExecutor) GetTreeJson(context.Context, *sdcpb.Path) (any, error) { return nil, nil }
func (s *stubExecutor) GetDiff(context.Context, *DiffConfig) (string, error)  { return "", nil }
func (s *stubExecutor) SchemaDownload(context.Context, *SchemaLoadConfig) (*sdcpb.Schema, error) {
	return nil, nil
}
func (s *stubExecutor) TreeLoadData(context.Context, *ConfigLoad) error { return nil }
func (s *stubExecutor) TreeShow(context.Context, *ConfigShowConfig) (ConfigShowInterface, error) {
	return s.entry, nil
}
func (s *stubExecutor) TreeValidate(context.Context) (treetypes.ValidationResults, *treetypes.ValidationStats, error) {
	return nil, nil, nil
}
func (s *stubExecutor) TreeBlame(context.Context, *ConfigBlameParams) (*sdcpb.BlameTreeElement, error) {
	return nil, nil
}

func TestConfigShowConfig_ForwardsAllFlag(t *testing.T) {
	formats := []types.ConfigFormat{
		types.ConfigFormatJson,
		types.ConfigFormatJsonIetf,
		types.ConfigFormatYaml,
		types.ConfigFormatXml,
		types.ConfigFormatXPath,
	}

	for _, format := range formats {
		for _, all := range []bool{false, true} {
			spy := &spyConfigShowEntry{}
			exec := &stubExecutor{entry: spy}

			cfg := NewConfigShowConfig().SetOutputFormat(format).SetAll(all)
			out, err := cfg.Run(context.Background(), exec)
			if err != nil {
				t.Fatalf("format=%s all=%v: Run() error: %v", format, all, err)
			}
			if _, err := out.ToString(context.Background()); err != nil {
				t.Fatalf("format=%s all=%v: ToString() error: %v", format, all, err)
			}

			wantOnlyNewOrUpdated := !all
			if spy.gotOnlyNewOrUpdated != wantOnlyNewOrUpdated {
				t.Fatalf("format=%s all=%v: got onlyNewOrUpdated=%v, want %v", format, all, spy.gotOnlyNewOrUpdated, wantOnlyNewOrUpdated)
			}
		}
	}
}
