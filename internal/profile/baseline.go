package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"mozsession-merge/internal/model"
)

// LoadBaseline reads the tool's saved baseline snapshot for a profile.
// Returns (model, exists, error).
func LoadBaseline(toolStateDir, profileID string) (model.NormalizedModel, bool, error) {
	if toolStateDir == "" || profileID == "" {
		return model.NormalizedModel{}, false, nil
	}

	baselinePath := filepath.Join(toolStateDir, "baselines", profileID+".json")
	if info, err := os.Stat(baselinePath); err != nil || info.IsDir() {
		return model.NormalizedModel{}, false, nil
	}

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return model.NormalizedModel{}, false, fmt.Errorf("baseline load: failed to read %s: %w", baselinePath, err)
	}

	var m model.NormalizedModel
	if err := json.Unmarshal(data, &m); err != nil {
		return model.NormalizedModel{}, false, fmt.Errorf("baseline load: failed to unmarshal JSON: %w", err)
	}

	return m, true, nil
}

// SaveBaseline writes the tool's baseline snapshot for a profile.
func SaveBaseline(toolStateDir, profileID string, m model.NormalizedModel) error {
	if toolStateDir == "" || profileID == "" {
		return errors.New("baseline save: toolStateDir and profileID must not be empty")
	}

	dir := filepath.Join(toolStateDir, "baselines")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("baseline save: failed creating directory %s: %w", dir, err)
	}

	baselinePath := filepath.Join(dir, profileID+".json")
	tmpPath := baselinePath + ".tmp"

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline save: json marshal failed: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("baseline save: failed writing tmp file: %w", err)
	}

	if err := os.Rename(tmpPath, baselinePath); err != nil {
		return fmt.Errorf("baseline save: failed renaming tmp file: %w", err)
	}

	return nil
}
