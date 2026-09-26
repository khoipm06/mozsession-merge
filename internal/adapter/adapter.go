package adapter

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/khoipm06/mozsession-merge/internal/model"
)

type Adapter interface {
	Name() string
	Detect(profileDir string) (bool, error)
	Load(profileDir string) (model.NormalizedModel, error)
	Save(profileDir string, m model.NormalizedModel) error
}

var ErrUnknownAdapter = errors.New("adapter: could not detect browser flavor for profile directory")

// DetectAdapter returns ZenAdapter if zen-sessions.jsonlz4 exists, else FirefoxAdapter if sessionstore.jsonlz4 exists.
func DetectAdapter(profileDir string, firefoxAdapter, zenAdapter Adapter) (Adapter, error) {
	if zenAdapter != nil {
		ok, err := zenAdapter.Detect(profileDir)
		if err == nil && ok {
			return zenAdapter, nil
		}
	}
	if firefoxAdapter != nil {
		ok, err := firefoxAdapter.Detect(profileDir)
		if err == nil && ok {
			return firefoxAdapter, nil
		}
	}
	return nil, ErrUnknownAdapter
}

// FileExists returns true if path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func SessionStorePath(profileDir string) string {
	return filepath.Join(profileDir, "sessionstore.jsonlz4")
}

func ZenSessionsPath(profileDir string) string {
	return filepath.Join(profileDir, "zen-sessions.jsonlz4")
}
