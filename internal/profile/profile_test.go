package profile_test

import (
	"os"
	"testing"

	"github.com/khoipm06/mozsession-merge/internal/model"
	"github.com/khoipm06/mozsession-merge/internal/profile"
)

func TestDetectBrowser(t *testing.T) {
	kind1, err := profile.DetectBrowser("../../testdata/Profile1")
	if err != nil {
		t.Fatalf("DetectBrowser Profile1 failed: %v", err)
	}
	if kind1 != "zen" {
		t.Fatalf("expected Profile1 to be zen, got %s", kind1)
	}
}

func TestBaselineStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mozsession-test-baseline-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	profileID := "test-profile-123"
	model1 := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
		},
	}

	// 1. Check non-existent baseline
	_, exists, err := profile.LoadBaseline(tempDir, profileID)
	if err != nil {
		t.Fatalf("unexpected error loading missing baseline: %v", err)
	}
	if exists {
		t.Fatalf("expected baseline to not exist initially")
	}

	// 2. Save baseline
	if err := profile.SaveBaseline(tempDir, profileID, model1); err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// 3. Load baseline and compare
	model2, exists, err := profile.LoadBaseline(tempDir, profileID)
	if err != nil {
		t.Fatalf("LoadBaseline failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected baseline to exist after saving")
	}

	if len(model2.Tabs) != 1 || model2.Tabs[0].ID != "tab-1" {
		t.Fatalf("loaded baseline tab mismatch: %v", model2.Tabs)
	}
}
