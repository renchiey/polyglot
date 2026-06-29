package generate

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// TestPipelineOllamaIntegration runs the full pipeline against a real local
// Ollama model and the real HSK auditor. It is skipped unless OLLAMA_MODEL is
// set, so it never runs in CI. Run it with:
//
//	OLLAMA_MODEL=gemma4:26b go test ./internal/generate -run Ollama -v -timeout 10m
//
// Optionally set OLLAMA_HOST to override the daemon address.
func TestPipelineOllamaIntegration(t *testing.T) {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		t.Skip("set OLLAMA_MODEL to run the Ollama integration test")
	}

	client := llm.NewOllamaClient(os.Getenv("OLLAMA_HOST"), model)

	auditor, err := lexaudit.NewRegistry().Get("zh")
	if err != nil {
		t.Fatalf("load auditor: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()

	const target = 2
	res, err := NewPipeline(client, DefaultMaxRounds).Run(ctx, auditor, GenRequest{
		Language:    "zh",
		TargetLevel: target,
		Topic:       "ordering coffee at a cafe",
		Kind:        "short story",
		KnownWords:  []string{"我", "你", "喜欢", "咖啡", "喝", "谢谢", "请", "服务员"},
	})
	if err != nil {
		t.Fatalf("pipeline run: %v", err)
	}

	t.Logf("passed=%v rounds=%d sentence_level=%d", res.Passed, res.Rounds, res.Report.SentenceLevel)
	t.Logf("out_of_bounds=%v unknown=%v", res.Report.OutOfBounds, res.Report.Unknown)
	t.Logf("text:\n%s", res.Text)

	if res.Text == "" {
		t.Error("expected non-empty generated text")
	}
	if res.Report.TargetLevel != target {
		t.Errorf("report target level = %d, want %d", res.Report.TargetLevel, target)
	}
}
