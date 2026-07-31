package zen_test

import (
	"os"
	"path/filepath"
	"testing"

	"mozsession-merge/internal/adapter/zen"
)

func TestZenAdapterRoundTrip(t *testing.T) {
	samplePath := "../../../testdata/Profile1/zen-sessions.jsonlz4"
	absPath, err := filepath.Abs(samplePath)
	if err != nil {
		t.Fatalf("failed to resolve path %s: %v", samplePath, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("failed to read sample file %s: %v", absPath, err)
	}

	adapter := zen.NewAdapter()
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

	t.Logf("ZenAdapter successfully round-tripped %d tabs, %d spaces, %d groups, %d folders",
		len(model2.Tabs), len(model2.Spaces), len(model2.Groups), len(model2.Folders))
}

func TestPlaceholderTabFiltering(t *testing.T) {
	sessionData := zen.ZenSessionData{
		LastCollected: 1000,
		Tabs: []map[string]any{
			{"zenSyncId": "tab-1", "zenIsEmpty": false},
			{"zenSyncId": "tab-placeholder", "zenIsEmpty": true, "groupId": nil},
			{"zenSyncId": "tab-in-group", "zenIsEmpty": true, "groupId": "group-1"},
		},
	}

	m, err := zen.ToModel(&sessionData)
	if err != nil {
		t.Fatalf("ToModel failed: %v", err)
	}

	if len(m.Tabs) != 2 {
		t.Fatalf("expected 2 tabs (placeholder tab filtered out), got %d", len(m.Tabs))
	}

	tabIDs := make(map[string]bool)
	for _, tab := range m.Tabs {
		tabIDs[string(tab.ID)] = true
	}

	if tabIDs["tab-placeholder"] {
		t.Fatalf("placeholder tab was not filtered out")
	}
	if !tabIDs["tab-1"] || !tabIDs["tab-in-group"] {
		t.Fatalf("expected tabs missing from model: %v", tabIDs)
	}
}
