package writer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"mozsession-merge/internal/writer"
)

func TestAtomicWrite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mozsession-test-writer-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "session.jsonlz4")
	originalData := []byte("original session data")
	newData := []byte("new merged session data")

	// 1. Initial write
	if err := os.WriteFile(targetFile, originalData, 0644); err != nil {
		t.Fatalf("failed writing initial file: %v", err)
	}

	// 2. Dry run check
	if err := writer.AtomicWrite(targetFile, newData, true); err != nil {
		t.Fatalf("dry run AtomicWrite failed: %v", err)
	}
	content, _ := os.ReadFile(targetFile)
	if !bytes.Equal(content, originalData) {
		t.Fatalf("dry run mutated file contents")
	}

	// 3. Live atomic write
	if err := writer.AtomicWrite(targetFile, newData, false); err != nil {
		t.Fatalf("live AtomicWrite failed: %v", err)
	}

	// Check updated contents
	content, _ = os.ReadFile(targetFile)
	if !bytes.Equal(content, newData) {
		t.Fatalf("live AtomicWrite contents mismatch")
	}

	// Check backup file exists and matches original
	bakContent, err := os.ReadFile(targetFile + ".bak")
	if err != nil {
		t.Fatalf("backup file missing: %v", err)
	}
	if !bytes.Equal(bakContent, originalData) {
		t.Fatalf("backup content mismatch")
	}
}
