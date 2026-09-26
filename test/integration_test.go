package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/khoipm06/mozsession-merge/internal/adapter/firefox"
	"github.com/khoipm06/mozsession-merge/internal/adapter/zen"
	"github.com/khoipm06/mozsession-merge/internal/model"
)

func TestMergeZenVsZen(t *testing.T) {
	tempDir := t.TempDir()
	profileA := filepath.Join(tempDir, "ZenA")
	profileB := filepath.Join(tempDir, "ZenB")
	stateDir := filepath.Join(tempDir, "state")
	reportFile := filepath.Join(tempDir, "report_zen_zen.json")

	os.MkdirAll(profileA, 0755)
	os.MkdirAll(profileB, 0755)

	zenAdapter := zen.NewAdapter()

	modelA := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-shared", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://shared.com"}}, "zenSyncId": "tab-shared", "lastAccessed": int64(100)}},
			{ID: "tab-zenA", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://zenA.com"}}, "zenSyncId": "tab-zenA", "lastAccessed": int64(200)}},
		},
		Spaces: []model.Entity{
			{ID: "{space-A}", Raw: map[string]any{"uuid": "{space-A}", "name": "Work"}},
		},
		Meta: model.Meta{LastCollected: 1000},
	}

	modelB := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-shared", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://shared.com"}}, "zenSyncId": "tab-shared", "lastAccessed": int64(100)}},
			{ID: "tab-zenB", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://zenB.com"}}, "zenSyncId": "tab-zenB", "lastAccessed": int64(300)}},
		},
		Spaces: []model.Entity{
			{ID: "{space-B}", Raw: map[string]any{"uuid": "{space-B}", "name": "Personal"}},
		},
		Meta: model.Meta{LastCollected: 1050},
	}

	if err := zenAdapter.Save(profileA, modelA); err != nil {
		t.Fatalf("failed saving ZenA: %v", err)
	}
	if err := zenAdapter.Save(profileB, modelB); err != nil {
		t.Fatalf("failed saving ZenB: %v", err)
	}

	cmd := exec.Command("go", "run", "../cmd/mergetabs",
		"--force",
		"--state-dir", stateDir,
		"--report-file", reportFile,
		profileA, profileB,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Zen vs Zen merge failed: %v\nOutput: %s", err, string(out))
	}

	mergedA, err := zenAdapter.Load(profileA)
	if err != nil {
		t.Fatalf("Load merged ZenA failed: %v", err)
	}
	mergedB, err := zenAdapter.Load(profileB)
	if err != nil {
		t.Fatalf("Load merged ZenB failed: %v", err)
	}

	if len(mergedA.Tabs) != 3 || len(mergedB.Tabs) != 3 {
		t.Fatalf("expected 3 merged tabs, got A: %d, B: %d", len(mergedA.Tabs), len(mergedB.Tabs))
	}
	if len(mergedA.Spaces) != 2 || len(mergedB.Spaces) != 2 {
		t.Fatalf("expected 2 merged spaces, got A: %d, B: %d", len(mergedA.Spaces), len(mergedB.Spaces))
	}
}

func TestMergeZenVsFirefox(t *testing.T) {
	tempDir := t.TempDir()
	profileZen := filepath.Join(tempDir, "ZenProfile")
	profileFF := filepath.Join(tempDir, "FFProfile")
	stateDir := filepath.Join(tempDir, "state")
	reportFile := filepath.Join(tempDir, "report_zen_ff.json")

	os.MkdirAll(profileZen, 0755)
	os.MkdirAll(profileFF, 0755)

	zenAdapter := zen.NewAdapter()
	ffAdapter := firefox.NewAdapter()

	modelZen := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-zen-1", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://zen-tab.com"}}, "zenSyncId": "tab-zen-1", "lastAccessed": int64(100)}},
		},
		Spaces: []model.Entity{
			{ID: "{space-zen}", Raw: map[string]any{"uuid": "{space-zen}", "name": "Zen Workspace"}},
		},
		Meta: model.Meta{LastCollected: 1000},
	}

	modelFF := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "synth-ff-1", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://ff-tab.com"}}, "lastAccessed": int64(200)}},
		},
		Meta: model.Meta{
			LastCollected: 1000,
			SourceFormat:  "firefox",
			WindowLayout: []model.WindowLayout{
				{WindowID: "win-0", TabIDs: []model.EntityID{"synth-ff-1"}, Selected: "synth-ff-1"},
			},
		},
	}

	if err := zenAdapter.Save(profileZen, modelZen); err != nil {
		t.Fatalf("failed saving profileZen: %v", err)
	}
	if err := ffAdapter.Save(profileFF, modelFF); err != nil {
		t.Fatalf("failed saving profileFF: %v", err)
	}

	cmd := exec.Command("go", "run", "../cmd/mergetabs",
		"--force",
		"--state-dir", stateDir,
		"--report-file", reportFile,
		profileZen, profileFF,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Zen vs Firefox merge failed: %v\nOutput: %s", err, string(out))
	}

	mergedZen, err := zenAdapter.Load(profileZen)
	if err != nil {
		t.Fatalf("failed loading merged Zen profile: %v", err)
	}
	mergedFF, err := ffAdapter.Load(profileFF)
	if err != nil {
		t.Fatalf("failed loading merged Firefox profile: %v", err)
	}

	if len(mergedZen.Tabs) != 2 {
		t.Fatalf("expected 2 merged tabs in Zen profile, got %d", len(mergedZen.Tabs))
	}
	if len(mergedFF.Tabs) != 2 {
		t.Fatalf("expected 2 merged tabs in Firefox profile, got %d", len(mergedFF.Tabs))
	}
}

func TestMergeFirefoxVsFirefox(t *testing.T) {
	tempDir := t.TempDir()
	profileFFA := filepath.Join(tempDir, "FFA")
	profileFFB := filepath.Join(tempDir, "FFB")
	stateDir := filepath.Join(tempDir, "state")
	reportFile := filepath.Join(tempDir, "report_ff_ff.json")

	os.MkdirAll(profileFFA, 0755)
	os.MkdirAll(profileFFB, 0755)

	ffAdapter := firefox.NewAdapter()

	modelFFA := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "synth-tab-1", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://shared.com"}}, "lastAccessed": int64(100)}},
			{ID: "synth-tab-A", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://ffA.com"}}, "lastAccessed": int64(200)}},
		},
		Meta: model.Meta{
			LastCollected: 1000,
			SourceFormat:  "firefox",
			WindowLayout: []model.WindowLayout{
				{WindowID: "win-0", TabIDs: []model.EntityID{"synth-tab-1", "synth-tab-A"}, Selected: "synth-tab-1"},
			},
		},
	}

	modelFFB := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "synth-tab-1", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://shared.com"}}, "lastAccessed": int64(100)}},
			{ID: "synth-tab-B", Raw: map[string]any{"entries": []any{map[string]any{"url": "https://ffB.com"}}, "lastAccessed": int64(300)}},
		},
		Meta: model.Meta{
			LastCollected: 1050,
			SourceFormat:  "firefox",
			WindowLayout: []model.WindowLayout{
				{WindowID: "win-0", TabIDs: []model.EntityID{"synth-tab-1"}, Selected: "synth-tab-1"},
				{WindowID: "win-1", TabIDs: []model.EntityID{"synth-tab-B"}, Selected: "synth-tab-B"},
			},
		},
	}

	if err := ffAdapter.Save(profileFFA, modelFFA); err != nil {
		t.Fatalf("failed saving FFA: %v", err)
	}
	if err := ffAdapter.Save(profileFFB, modelFFB); err != nil {
		t.Fatalf("failed saving FFB: %v", err)
	}

	cmd := exec.Command("go", "run", "../cmd/mergetabs",
		"--force",
		"--state-dir", stateDir,
		"--report-file", reportFile,
		profileFFA, profileFFB,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Firefox vs Firefox merge failed: %v\nOutput: %s", err, string(out))
	}

	mergedFFA, err := ffAdapter.Load(profileFFA)
	if err != nil {
		t.Fatalf("failed loading merged FFA: %v", err)
	}
	mergedFFB, err := ffAdapter.Load(profileFFB)
	if err != nil {
		t.Fatalf("failed loading merged FFB: %v", err)
	}

	if len(mergedFFA.Tabs) != 3 || len(mergedFFB.Tabs) != 3 {
		t.Fatalf("expected 3 merged tabs in both Firefox profiles, got A: %d, B: %d", len(mergedFFA.Tabs), len(mergedFFB.Tabs))
	}
}
