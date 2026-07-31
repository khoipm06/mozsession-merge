//go:build unix || linux || darwin || freebsd || openbsd || netbsd

package profile

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func checkOSSpecificLock(lockPath string, info os.FileInfo) (bool, string, error) {
	// Case 1: Symlink lock (standard Firefox/Zen lock on Unix)
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(lockPath)
		if err == nil && target != "" {
			pidStr := extractPIDFromSymlink(target)
			if pidStr != "" {
				pid, pErr := strconv.Atoi(pidStr)
				if pErr == nil && isPIDAlive(pid) {
					return true, fmt.Sprintf("locked by active PID %d (symlink -> %s)", pid, target), nil
				}
				// Stale crash lock
				return false, "", nil
			}
		}
	}

	// Case 2: Check non-blocking flock on lock file
	file, err := os.OpenFile(lockPath, os.O_RDWR, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return true, "permission denied opening lock file", nil
		}
		return false, "", nil
	}
	defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return true, "exclusive flock held by active process", nil
		}
	} else {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
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
	// Check /proc/<pid> on Linux if present
	procPath := fmt.Sprintf("/proc/%d", pid)
	if _, err := os.Stat(procPath); err == nil {
		return true
	}
	// Fallback kill 0 check on Unix (macOS, FreeBSD, Linux)
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}
