package merge_test

import (
	"testing"

	"mozsession-merge/internal/merge"
	"mozsession-merge/internal/model"
)

func TestNewTabEitherSide(t *testing.T) {
	base := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
		},
	}
	local := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
			{ID: "tab-local", Raw: map[string]any{"url": "https://local.com"}},
		},
	}
	remote := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
			{ID: "tab-remote", Raw: map[string]any{"url": "https://remote.com"}},
		},
	}

	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: false}
	merged, changes, conflicts, err := merge.ThreeWayMerge(base, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
	if len(merged.Tabs) != 3 {
		t.Fatalf("expected 3 tabs in merged result, got %d", len(merged.Tabs))
	}

	tabIDs := make(map[model.EntityID]bool)
	for _, tab := range merged.Tabs {
		tabIDs[tab.ID] = true
	}
	if !tabIDs["tab-1"] || !tabIDs["tab-local"] || !tabIDs["tab-remote"] {
		t.Fatalf("merged tabs missing expected IDs: %v", tabIDs)
	}

	// Verify change log recorded additions
	additions := 0
	for _, c := range changes {
		if c.Kind == "added" {
			additions++
		}
	}
	if additions != 2 {
		t.Fatalf("expected 2 addition records, got %d", additions)
	}
}

func TestDeletedTabEitherSide(t *testing.T) {
	base := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
			{ID: "tab-2", Raw: map[string]any{"url": "https://delete-local.com"}},
			{ID: "tab-3", Raw: map[string]any{"url": "https://delete-remote.com"}},
		},
	}
	local := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
			{ID: "tab-3", Raw: map[string]any{"url": "https://delete-remote.com"}},
		},
	}
	remote := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com"}},
			{ID: "tab-2", Raw: map[string]any{"url": "https://delete-local.com"}},
		},
	}

	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: false}
	merged, changes, conflicts, err := merge.ThreeWayMerge(base, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
	if len(merged.Tabs) != 1 || merged.Tabs[0].ID != "tab-1" {
		t.Fatalf("expected only tab-1 remaining, got %v", merged.Tabs)
	}

	removals := 0
	for _, c := range changes {
		if c.Kind == "removed" {
			removals++
		}
	}
	if removals != 2 {
		t.Fatalf("expected 2 removal records, got %d", removals)
	}
}

func TestConflictingRename(t *testing.T) {
	base := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com", "title": "Original Title", "lastAccessed": int64(100)}},
		},
	}
	local := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com", "title": "Local Title", "lastAccessed": int64(200)}},
		},
	}
	remote := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com", "title": "Remote Title", "lastAccessed": int64(300)}},
		},
	}

	// StrategyNewerWins should pick remote (lastAccessed 300 > 200)
	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: false}
	merged, _, conflicts, err := merge.ThreeWayMerge(base, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if len(merged.Tabs) != 1 {
		t.Fatalf("expected 1 merged tab, got %d", len(merged.Tabs))
	}
	if merged.Tabs[0].Raw["title"] != "Remote Title" {
		t.Fatalf("expected Remote Title (newer timestamp), got %v", merged.Tabs[0].Raw["title"])
	}
}

func TestMovedWorkspace(t *testing.T) {
	base := model.NormalizedModel{
		Spaces: []model.Entity{
			{ID: "space-A", Raw: map[string]any{"name": "Space A"}},
			{ID: "space-B", Raw: map[string]any{"name": "Space B"}},
		},
		Folders: []model.Entity{
			{ID: "folder-1", Raw: map[string]any{"name": "Docs", "workspaceId": "space-A"}},
		},
	}
	local := model.NormalizedModel{
		Spaces: []model.Entity{
			{ID: "space-A", Raw: map[string]any{"name": "Space A"}},
			{ID: "space-B", Raw: map[string]any{"name": "Space B"}},
		},
		Folders: []model.Entity{
			{ID: "folder-1", Raw: map[string]any{"name": "Docs", "workspaceId": "space-B"}},
		},
	}
	remote := model.NormalizedModel{
		Spaces: []model.Entity{
			{ID: "space-A", Raw: map[string]any{"name": "Space A"}},
			{ID: "space-B", Raw: map[string]any{"name": "Space B"}},
		},
		Folders: []model.Entity{
			{ID: "folder-1", Raw: map[string]any{"name": "Docs", "workspaceId": "space-A"}},
		},
	}

	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: false}
	merged, _, conflicts, err := merge.ThreeWayMerge(base, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
	if merged.Folders[0].Raw["workspaceId"] != "space-B" {
		t.Fatalf("expected folder workspaceId to be space-B, got %v", merged.Folders[0].Raw["workspaceId"])
	}
}

func TestFirstRunNoBaseline(t *testing.T) {
	emptyBase := model.NormalizedModel{}
	local := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://local-only.com"}},
		},
	}
	remote := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-2", Raw: map[string]any{"url": "https://remote-only.com"}},
		},
	}

	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: true}
	merged, changes, conflicts, err := merge.ThreeWayMerge(emptyBase, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
	if len(merged.Tabs) != 2 {
		t.Fatalf("expected 2 tabs in union, got %d", len(merged.Tabs))
	}

	hasWarning := false
	for _, c := range changes {
		if c.Kind == "warning" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Fatalf("expected first-run warning change log, found none")
	}
}

func TestIntegritySweepDanglingReferences(t *testing.T) {
	base := model.NormalizedModel{}
	local := model.NormalizedModel{
		Tabs: []model.Entity{
			{ID: "tab-1", Raw: map[string]any{"url": "https://example.com", "groupId": "non-existent-group", "zenWorkspace": "non-existent-space"}},
		},
		Folders: []model.Entity{
			{ID: "folder-1", Raw: map[string]any{"name": "Folder 1", "parentId": "non-existent-parent", "workspaceId": "non-existent-space"}},
		},
	}
	remote := model.NormalizedModel{}

	opts := merge.Options{Strategy: merge.StrategyNewerWins, FirstRun: true}
	merged, changes, _, err := merge.ThreeWayMerge(base, local, remote, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tab groupId and zenWorkspace should be removed
	if _, exists := merged.Tabs[0].Raw["groupId"]; exists {
		t.Fatalf("expected groupId to be stripped by integrity sweep")
	}
	if _, exists := merged.Tabs[0].Raw["zenWorkspace"]; exists {
		t.Fatalf("expected zenWorkspace to be stripped by integrity sweep")
	}

	// Folder workspaceId should be removed and parentId set to nil
	if _, exists := merged.Folders[0].Raw["workspaceId"]; exists {
		t.Fatalf("expected folder workspaceId to be stripped by integrity sweep")
	}
	if parentId := merged.Folders[0].Raw["parentId"]; parentId != nil {
		t.Fatalf("expected folder parentId to be nil, got %v", parentId)
	}

	integrityFixes := 0
	for _, c := range changes {
		if c.Kind == "integrity-fix" {
			integrityFixes++
		}
	}
	if integrityFixes != 4 {
		t.Fatalf("expected 4 integrity-fix change log entries, got %d", integrityFixes)
	}
}
