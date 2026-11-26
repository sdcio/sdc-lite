package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersistWithoutTargetFails(t *testing.T) {
	f := filepath.Join(t.TempDir(), "p.json")
	if err := os.WriteFile(f, nil, 0644); err != nil {
		t.Fatal(err)
	}
	out := make(chan *PipelineResult, 4)
	NewPipeline(f, WithPersist()).Run(context.Background(), out)

	res := <-out
	if res == nil || !res.IsError() || !strings.Contains(res.GetError().Error(), "requires a target") {
		t.Fatalf("expected 'requires a target' error, got %+v", res)
	}
}
