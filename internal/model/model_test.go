package model_test

import (
	"testing"

	"mozsession-merge/internal/model"
)

func TestSynthesizeTabID(t *testing.T) {
	urls1 := []string{"https://example.com/page1", "https://example.com/page2"}
	urls2 := []string{"https://example.com/page1", "https://example.com/page2"}
	urls3 := []string{"https://example.com/other"}

	id1 := model.SynthesizeTabID(urls1, 0, false)
	id2 := model.SynthesizeTabID(urls2, 0, false)
	id3 := model.SynthesizeTabID(urls3, 0, false)
	id1Pinned := model.SynthesizeTabID(urls1, 0, true)

	if id1 != id2 {
		t.Fatalf("expected identical IDs for identical history, got %s vs %s", id1, id2)
	}
	if id1 == id3 {
		t.Fatalf("expected different IDs for different history, got collision %s", id1)
	}
	if id1 == id1Pinned {
		t.Fatalf("expected different IDs for pinned vs unpinned tab, got collision %s", id1)
	}
}

func TestStripAndContentEqual(t *testing.T) {
	map1 := map[string]any{
		"url":             "https://example.com",
		"title":           "Example",
		"_zenIsActiveTab": true,
		"lastAccessed":    int64(123456789),
		"index":           1,
	}

	map2 := map[string]any{
		"url":             "https://example.com",
		"title":           "Example",
		"_zenIsActiveTab": false,
		"lastAccessed":    int64(987654321),
		"index":           5,
	}

	mapDifferent := map[string]any{
		"url":   "https://different.com",
		"title": "Example",
	}

	if !model.IsContentEqual(map1, map2) {
		t.Fatalf("expected map1 and map2 to be content equal after stripping ephemeral fields")
	}

	if model.IsContentEqual(map1, mapDifferent) {
		t.Fatalf("expected map1 and mapDifferent to not be equal")
	}
}
