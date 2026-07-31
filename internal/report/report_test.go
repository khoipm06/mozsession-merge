package report_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mozsession-merge/internal/merge"
	"mozsession-merge/internal/report"
)

func TestReportWriteHumanAndJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mozsession-test-report-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sources := map[string]string{
		"profileA": "/path/to/profileA/zen-sessions.jsonlz4",
		"profileB": "/path/to/profileB/zen-sessions.jsonlz4",
	}

	changes := []merge.Change{
		{ID: "tab-1", Entity: "tab", Kind: "added", Detail: "added on local side"},
		{ID: "tab-2", Entity: "tab", Kind: "removed", Detail: "deleted on remote side"},
	}

	conflicts := []merge.Conflict{
		{ID: "tab-3", Entity: "tab", Strategy: merge.StrategyNewerWins},
	}

	rep := report.NewReport(true, sources, changes, conflicts)

	// Test human format output
	var buf bytes.Buffer
	if err := rep.WriteHuman(&buf); err != nil {
		t.Fatalf("WriteHuman failed: %v", err)
	}
	humanOutput := buf.String()
	if !strings.Contains(humanOutput, "DRY RUN") {
		t.Fatalf("human report missing DRY RUN banner")
	}
	if !strings.Contains(humanOutput, "added on local side") {
		t.Fatalf("human report missing detailed change text")
	}

	// Test JSON format output
	jsonPath := filepath.Join(tempDir, "report.json")
	if err := rep.WriteJSON(jsonPath); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read written JSON report: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("JSON report file is empty")
	}
}
