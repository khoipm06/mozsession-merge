package backupfallback_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/khoipm06/mozsession-merge/internal/backupfallback"
	"github.com/khoipm06/mozsession-merge/internal/mozlz4"
)

func TestLoadWithFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mozsession-test-fallback-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	primary := filepath.Join(tempDir, "missing.jsonlz4")
	fallback1 := filepath.Join(tempDir, "fallback1.jsonlz4")
	fallback2 := filepath.Join(tempDir, "fallback2.jsonlz4")

	// Create valid fallback2 file
	validJSON := []byte(`{"version":["sessionrestore",1],"windows":[]}`)
	encoded, err := mozlz4.Encode(validJSON)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if err := os.WriteFile(fallback2, encoded, 0644); err != nil {
		t.Fatalf("failed writing fallback2: %v", err)
	}

	rawBytes, decodedBytes, usedPath, err := backupfallback.LoadWithFallback(primary, []string{fallback1, fallback2})
	if err != nil {
		t.Fatalf("LoadWithFallback failed: %v", err)
	}

	if usedPath != fallback2 {
		t.Fatalf("expected usedPath to be %s, got %s", fallback2, usedPath)
	}
	if len(rawBytes) == 0 || len(decodedBytes) == 0 {
		t.Fatalf("empty bytes returned")
	}
}
