package firefox_test

import (
	"os"
	"path/filepath"
	"testing"

	"mozsession-merge/internal/adapter/firefox"
)

func TestFirefoxAdapterRoundTrip(t *testing.T) {
	samplePath := "../../../testdata/Profile1/sessionstore.jsonlz4"
	absPath, err := filepath.Abs(samplePath)
	if err != nil {
		t.Fatalf("failed to resolve path %s: %v", samplePath, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("failed to read sample file %s: %v", absPath, err)
	}

	adapter := firefox.NewAdapter()
	model1, err := adapter.LoadBytes(data)
	if err != nil {
		t.Fatalf("LoadBytes failed: %v", err)
	}

	if len(model1.Tabs) == 0 {
		t.Fatalf("expected non-empty tabs in loaded model")
	}

	encoded, err := adapter.SaveBytes(model1)
	if err != nil {
		t.Fatalf("SaveBytes failed: %v", err)
	}

	model2, err := adapter.LoadBytes(encoded)
	if err != nil {
		t.Fatalf("re-loading encoded bytes failed: %v", err)
	}

	if len(model1.Tabs) != len(model2.Tabs) {
		t.Fatalf("tab count mismatch after round trip: %d vs %d", len(model1.Tabs), len(model2.Tabs))
	}

	t.Logf("FirefoxAdapter successfully round-tripped %d tabs across %d windows",
		len(model2.Tabs), len(model2.Meta.WindowLayout))
}
