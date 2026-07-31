package profile

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsLocked checks if a SPECIFIC profile directory is currently locked by a running browser instance.
// Supported across Linux, macOS, FreeBSD, and Windows.
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

		// Delegate to OS-specific lock checker
		locked, reason, err := checkOSSpecificLock(lockPath, info)
		if err != nil {
			return false, "", err
		}
		if locked {
			return true, fmt.Sprintf("locked on %s (%s)", lf, reason), nil
		}
	}

	return false, "", nil
}
