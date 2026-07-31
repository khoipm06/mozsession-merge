package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsLocked checks for lock files AND scans running processes for browser binaries.
func IsLocked(profileDir string) (bool, string, error) {
	// 1. Check lock files in profile dir
	lockFiles := []string{"parent.lock", "lock", ".parentlock"}
	for _, lf := range lockFiles {
		lockPath := filepath.Join(profileDir, lf)
		info, err := os.Lstat(lockPath)
		if err == nil {
			// Lock file exists or symlink exists
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(lockPath)
				if err == nil && target != "" {
					return true, fmt.Sprintf("lock file %s present (symlink to %s)", lf, target), nil
				}
				return true, fmt.Sprintf("lock symlink %s present", lf), nil
			}
			// Try non-blocking open / stat
			return true, fmt.Sprintf("lock file %s present", lf), nil
		}
	}

	// 2. Scan running processes on Linux / Unix via /proc
	runningBrowser, err := checkRunningProcesses()
	if err == nil && runningBrowser != "" {
		return true, fmt.Sprintf("running browser process detected: %s", runningBrowser), nil
	}

	return false, "", nil
}

func checkRunningProcesses() (string, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return "", err
	}

	targetNames := map[string]bool{
		"firefox":      true,
		"firefox-bin":  true,
		"zen":          true,
		"zen-bin":      true,
		"firefox.exe":  true,
		"zen.exe":      true,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if len(pid) == 0 || pid[0] < '0' || pid[0] > '9' {
			continue
		}

		commPath := filepath.Join("/proc", pid, "comm")
		commBytes, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}
		comm := strings.TrimSpace(string(commBytes))

		if targetNames[strings.ToLower(comm)] {
			return comm, nil
		}
	}

	return "", nil
}
