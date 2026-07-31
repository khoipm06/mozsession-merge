package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"mozsession-merge/internal/adapter/zen"
)

func TestIntegrationMerge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mozsession-integration-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	profileA := filepath.Join(tempDir, "Profile1")
	profileB := filepath.Join(tempDir, "Profile2")
	stateDir := filepath.Join(tempDir, "state")
	reportFile := filepath.Join(tempDir, "report.json")

	// Copy testdata profiles to temp dirs
	if err := copyDir("../testdata/Profile1", profileA); err != nil {
		t.Fatalf("copy Profile1 failed: %v", err)
	}
	if err := copyDir("../testdata/Profile2", profileB); err != nil {
		t.Fatalf("copy Profile2 failed: %v", err)
	}

	// 1. First Run: Perform live merge without baseline
	cmd := exec.Command("go", "run", "../cmd/mergetabs",
		"--force",
		"--state-dir", stateDir,
		"--report-file", reportFile,
		profileA, profileB,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("first run mergetabs failed: %v\nOutput: %s", err, string(output))
	}

	// Verify report file was generated
	if _, err := os.Stat(reportFile); err != nil {
		t.Fatalf("report file missing after first run: %v", err)
	}

	// Verify both profiles were updated and can be loaded
	zenAdapter := zen.NewAdapter()
	modelA, err := zenAdapter.Load(profileA)
	if err != nil {
		t.Fatalf("failed to load merged profile A: %v", err)
	}
	modelB, err := zenAdapter.Load(profileB)
	if err != nil {
		t.Fatalf("failed to load merged profile B: %v", err)
	}

	if len(modelA.Tabs) != len(modelB.Tabs) {
		t.Fatalf("expected merged profile A and B to have identical tab count, got %d vs %d", len(modelA.Tabs), len(modelB.Tabs))
	}

	// 2. Second Run: Perform merge with established baseline (should report 0 changes)
	cmd2 := exec.Command("go", "run", "../cmd/mergetabs",
		"--force",
		"--state-dir", stateDir,
		"--report-file", reportFile,
		profileA, profileB,
	)
	output2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("second run mergetabs failed: %v\nOutput: %s", err, string(output2))
	}
	t.Logf("Second run output:\n%s", string(output2))
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
