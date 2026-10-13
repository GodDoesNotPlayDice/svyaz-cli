package context

import (
	"strings"
	"testing"
)

func TestRepoMap_Generate(t *testing.T) {
	// Point to current repo root "." (where go.mod and internal/ exist)
	rm := NewRepoMap("../../")
	res, err := rm.Generate()
	if err != nil {
		t.Fatalf("unexpected error generating repomap: %v", err)
	}

	if !strings.Contains(res, "REPOSITORY GROUNDING MAP") {
		t.Errorf("expected grounding header in repomap")
	}

	// Should contain module declaration or symbols
	if !strings.Contains(res, "svyaz-cli") {
		t.Errorf("expected svyaz-cli module mention in repomap: %s", res)
	}
}

func TestCompressOutput(t *testing.T) {
	var lines []string
	for i := 1; i <= 100; i++ {
		lines = append(lines, "log line")
	}
	raw := strings.Join(lines, "\n")

	compressed := CompressOutput(raw, 20)
	if !strings.Contains(compressed, "truncated by svyaz context compressor") {
		t.Errorf("expected truncated notice in compressed output")
	}

	resLines := strings.Split(compressed, "\n")
	if len(resLines) >= 100 {
		t.Errorf("expected fewer lines than original 100, got %d", len(resLines))
	}
}
