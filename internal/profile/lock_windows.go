//go:build windows

package profile

import (
	"os"
	"syscall"
)

func checkOSSpecificLock(lockPath string, info os.FileInfo) (bool, string, error) {
	file, err := os.OpenFile(lockPath, os.O_RDWR, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return true, "exclusive file handle locked by Windows browser process", nil
		}
		if pathErr, ok := err.(*os.PathError); ok {
			if errno, ok := pathErr.Err.(syscall.Errno); ok {
				// ERROR_SHARING_VIOLATION (32) or ERROR_LOCK_VIOLATION (33)
				if errno == 32 || errno == 33 {
					return true, "sharing violation: lock held by active Windows browser process", nil
				}
			}
		}
		return false, "", nil
	}
	file.Close()

	return false, "", nil
}
