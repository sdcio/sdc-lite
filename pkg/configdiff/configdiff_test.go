package configdiff

import (
	"context"
	"errors"
	"testing"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
)

func TestEmptyConfigDiffReturnsErrNoConfigLoaded(t *testing.T) {
	ctx := context.Background()
	c, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	cd, err := NewConfigDiff(ctx, c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cd.TreeShow(ctx, &params.ConfigShowConfig{})
	if !errors.Is(err, ErrNoConfigLoaded) {
		t.Fatalf("TreeShow: expected ErrNoConfigLoaded, got %v", err)
	}
	_, _, err = cd.TreeValidate(ctx)
	if !errors.Is(err, ErrNoConfigLoaded) {
		t.Fatalf("TreeValidate: expected ErrNoConfigLoaded, got %v", err)
	}
	if _, err = cd.GetJson(ctx, false); !errors.Is(err, ErrNoConfigLoaded) {
		t.Fatalf("GetJson: expected ErrNoConfigLoaded, got %v", err)
	}
	if got := cd.GetPathCompletions(ctx, "/"); got != nil {
		t.Fatalf("expected nil completions, got %v", got)
	}
}
