package writer

import (
	"fmt"
	"io"
	"os"
)

// AtomicWrite safely writes data to targetPath by:
// 1. Copying existing file to targetPath + ".bak" (skipped if dryRun is true)
// 2. Writing data to targetPath + ".tmp"
// 3. Renaming targetPath + ".tmp" to targetPath atomically
func AtomicWrite(targetPath string, data []byte, dryRun bool) error {
	if dryRun {
		return nil
	}

	bakPath := targetPath + ".bak"
	tmpPath := targetPath + ".tmp"

	// 1. Pre-write backup copy
	if _, err := os.Stat(targetPath); err == nil {
		if err := copyFile(targetPath, bakPath); err != nil {
			return fmt.Errorf("atomic write: backup creation failed: %w", err)
		}
	}

	// 2. Write temp file
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("atomic write: temp write failed: %w", err)
	}

	// 3. Atomic rename
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("atomic write: rename failed: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
