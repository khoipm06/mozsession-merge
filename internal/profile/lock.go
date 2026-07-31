package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// IsLocked checks if a SPECIFIC profile directory is currently locked by a running browser instance.
// It avoids false-positive global process checks, checking only lock files and PIDs associated with profileDir.
func IsLocked(profileDir string) (bool, string, error) {
	lockFiles := []string{"parent.lock", "lock", ".parentlock"}

	for _, lf := range lockFiles {
		lockPath := filepath.Join(profileDir, lf)

		info, err := os.Lstat(lockPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, "", err
		}

		// Case 1: Lock file is a symlink (standard Firefox/Zen lock on Linux/macOS)
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(lockPath)
			if err == nil && target != "" {
				// Symlink format is usually "127.0.0.1:+12345" or "127.0.0.1:12345"
				pidStr := extractPIDFromSymlink(target)
				if pidStr != "" {
					pid, pErr := strconv.Atoi(pidStr)
					if pErr == nil && isPIDAlive(pid) {
						return true, fmt.Sprintf("locked by active browser PID %d (symlink %s -> %s)", pid, lf, target), nil
					}
					// PID is dead (stale crash lock)
					continue
				}
			}
			// Symlink exists but target couldn't be parsed; check flock
		}

		// Case 2: Check non-blocking flock on lock file
		locked, pidMsg := checkFileFlock(lockPath)
		if locked {
			return true, fmt.Sprintf("locked by active process on %s (%s)", lf, pidMsg), nil
		}
	}

	return false, "", nil
}

func extractPIDFromSymlink(target string) string {
	parts := strings.Split(target, ":")
	if len(parts) >= 2 {
		pidPart := strings.TrimPrefix(parts[len(parts)-1], "+")
		return strings.TrimSpace(pidPart)
	}
	return ""
}

func isPIDAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	// Check /proc/<pid> on Linux
	procPath := fmt.Sprintf("/proc/%d", pid)
	if _, err := os.Stat(procPath); err == nil {
		return true
	}
	// Fallback signal 0 check
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func checkFileFlock(path string) (bool, string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		// If permission denied, it's likely locked by another process
		if os.IsPermission(err) {
			return true, "permission denied opening lock file"
		}
		return false, ""
	}
	defer file.Close()

	// Try non-blocking exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return true, "exclusive file lock held"
		}
	} else {
		// Successfully acquired lock; unlock immediately
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	}

	return false, ""
}
