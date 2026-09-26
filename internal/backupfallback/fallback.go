package backupfallback

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/khoipm06/mozsession-merge/internal/adapter"
	"github.com/khoipm06/mozsession-merge/internal/mozlz4"
)

var ErrNoValidSessionFile = errors.New("backupfallback: no valid session or backup file found")

// DefaultFallbackCandidates returns standard relative backup paths under profileDir.
func DefaultFallbackCandidates(profileDir string, kind string) []string {
	backupsDir := filepath.Join(profileDir, "sessionstore-backups")
	if kind == "zen" {
		return []string{
			filepath.Join(backupsDir, "zen-sessions.jsonlz4"),
			filepath.Join(backupsDir, "recovery.jsonlz4"),
			filepath.Join(backupsDir, "recovery.baklz4"),
			filepath.Join(backupsDir, "previous.jsonlz4"),
		}
	}

	return []string{
		filepath.Join(backupsDir, "recovery.jsonlz4"),
		filepath.Join(backupsDir, "recovery.baklz4"),
		filepath.Join(backupsDir, "previous.jsonlz4"),
		filepath.Join(backupsDir, "clean.jsonlz4"),
	}
}

// LoadWithFallback attempts to read and decode primary path first.
// If primary is missing or fails decoding, tries candidate fallbacks in order.
// Returns (rawFileBytes, decodedBytes, usedPath, error).
func LoadWithFallback(primaryPath string, fallbacks []string) (rawBytes []byte, decodedBytes []byte, usedPath string, err error) {
	pathsToTry := append([]string{primaryPath}, fallbacks...)

	for _, p := range pathsToTry {
		if !adapter.FileExists(p) {
			continue
		}

		data, readErr := os.ReadFile(p)
		if readErr != nil {
			continue
		}

		decoded, decErr := mozlz4.Decode(data)
		if decErr != nil || len(decoded) == 0 {
			continue
		}

		return data, decoded, p, nil
	}

	return nil, nil, "", fmt.Errorf("%w: tried primary %s and %d fallbacks", ErrNoValidSessionFile, primaryPath, len(fallbacks))
}
