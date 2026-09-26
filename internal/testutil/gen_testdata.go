//go:build ignore

// gen_testdata.go generates minimal synthetic .jsonlz4 test fixtures for
// testdata/Profile1 and testdata/Profile2.
//
// Run with: go run internal/testutil/gen_testdata.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/khoipm06/mozsession-merge/internal/mozlz4"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func writeJsonlz4(path string, v any) {
	data, err := json.Marshal(v)
	must(err)
	encoded, err := mozlz4.Encode(data)
	must(err)
	must(os.MkdirAll(filepath.Dir(path), 0755))
	must(os.WriteFile(path, encoded, 0644))
	fmt.Printf("wrote %s (%d bytes)\n", path, len(encoded))
}

func zenSession(tabID1, tabID2, spaceUUID, groupID string, lastCollected int64) any {
	return map[string]any{
		"lastCollected": lastCollected,
		"tabs": []any{
			map[string]any{
				"zenSyncId":       tabID1,
				"entries":         []any{map[string]any{"url": "https://example.com/tab1", "title": "Example Tab 1"}},
				"lastAccessed":    lastCollected - 1000,
				"pinned":          false,
				"hidden":          false,
				"zenIsEmpty":      false,
				"zenEssential":    false,
				"userContextId":   0,
				"zenWorkspace":    spaceUUID,
				"groupId":         groupID,
				"_zenIsActiveTab": true,
			},
			map[string]any{
				"zenSyncId":     tabID2,
				"entries":       []any{map[string]any{"url": "https://example.org/tab2", "title": "Example Tab 2"}},
				"lastAccessed":  lastCollected - 2000,
				"pinned":        false,
				"hidden":        false,
				"zenIsEmpty":    false,
				"zenEssential":  false,
				"userContextId": 0,
				"zenWorkspace":  spaceUUID,
			},
		},
		"spaces": []any{
			map[string]any{
				"uuid": spaceUUID,
				"name": "Test Space",
				"theme": map[string]any{
					"type":           "none",
					"gradientColors": []any{},
					"opacity":        1.0,
					"texture":        0,
				},
				"containerTabId":          0,
				"hasCollapsedPinnedTabs":  false,
			},
		},
		"groups": []any{
			map[string]any{
				"id":              groupID,
				"name":            "Test Group",
				"color":           "blue",
				"collapsed":       false,
				"pinned":          false,
				"saveOnWindowClose": false,
			},
		},
		"folders":       []any{},
		"splitViewData": []any{},
	}
}

func firefoxSession(tabURL, tabTitle string, lastUpdate int64) any {
	return map[string]any{
		"version":        []any{"sessionrestore", 1},
		"selectedWindow": 1,
		"session": map[string]any{
			"lastUpdate": float64(lastUpdate),
			"startTime":  float64(lastUpdate - 3600000),
		},
		"global": map[string]any{},
		"windows": []any{
			map[string]any{
				"tabs": []any{
					map[string]any{
						"entries": []any{
							map[string]any{"url": tabURL, "title": tabTitle},
						},
						"index":         1,
						"lastAccessed":  float64(lastUpdate - 500),
						"pinned":        false,
						"hidden":        false,
						"userContextId": 0,
					},
				},
				"selected":      1,
				"groups":        []any{},
				"folders":       []any{},
				"spaces":        []any{},
				"splitViewData": []any{},
				"width":         1280,
				"height":        800,
				"screenX":       0,
				"screenY":       0,
				"sizemode":      "normal",
			},
		},
	}
}

func main() {
	// Profile1 — Zen profile
	writeJsonlz4("testdata/Profile1/zen-sessions.jsonlz4",
		zenSession(
			"synth-tab-p1-a",
			"synth-tab-p1-b",
			"{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}",
			"synth-group-p1",
			1700000001000,
		),
	)
	writeJsonlz4("testdata/Profile1/sessionstore.jsonlz4",
		firefoxSession("https://example.com/profile1", "Profile 1 Tab", 1700000001000),
	)

	// Profile2 — Zen profile with overlapping + unique tabs
	writeJsonlz4("testdata/Profile2/zen-sessions.jsonlz4",
		zenSession(
			"synth-tab-p1-a",   // shared with Profile1 (same ID, same content)
			"synth-tab-p2-only", // unique to Profile2
			"{ffffffff-0000-1111-2222-333333333333}",
			"synth-group-p2",
			1700000002000,
		),
	)
	writeJsonlz4("testdata/Profile2/sessionstore.jsonlz4",
		firefoxSession("https://example.com/profile2", "Profile 2 Tab", 1700000002000),
	)

	fmt.Println("done")
}
