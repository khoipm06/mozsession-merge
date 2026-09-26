package profile

import (
	"errors"

	"github.com/khoipm06/mozsession-merge/internal/adapter"
)

var ErrProfileNotFound = errors.New("profile: directory does not exist or contains no session files")

// DetectBrowser identifies if a profile is Zen ("zen") or Stock Firefox ("firefox").
func DetectBrowser(profileDir string) (string, error) {
	zenPath := adapter.ZenSessionsPath(profileDir)
	if adapter.FileExists(zenPath) {
		return "zen", nil
	}

	firefoxPath := adapter.SessionStorePath(profileDir)
	if adapter.FileExists(firefoxPath) {
		return "firefox", nil
	}

	return "", ErrProfileNotFound
}
