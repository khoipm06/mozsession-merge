package zen

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"mozsession-merge/internal/adapter"
	"mozsession-merge/internal/model"
	"mozsession-merge/internal/mozlz4"
)

type ZenAdapter struct{}

func NewAdapter() *ZenAdapter {
	return &ZenAdapter{}
}

func (a *ZenAdapter) Name() string {
	return "zen"
}

func (a *ZenAdapter) Detect(profileDir string) (bool, error) {
	zenSessions := adapter.ZenSessionsPath(profileDir)
	if adapter.FileExists(zenSessions) {
		return true, nil
	}
	return false, nil
}

func (a *ZenAdapter) Load(profileDir string) (model.NormalizedModel, error) {
	zenSessionsPath := adapter.ZenSessionsPath(profileDir)
	data, err := os.ReadFile(zenSessionsPath)
	if err != nil {
		return model.NormalizedModel{}, fmt.Errorf("zen load: failed to read %s: %w", zenSessionsPath, err)
	}

	return a.LoadBytes(data)
}

func (a *ZenAdapter) LoadBytes(data []byte) (model.NormalizedModel, error) {
	jsonBytes, err := mozlz4.Decode(data)
	if err != nil {
		return model.NormalizedModel{}, fmt.Errorf("zen load: mozlz4 decode failed: %w", err)
	}

	var sessionData ZenSessionData
	if err := json.Unmarshal(jsonBytes, &sessionData); err != nil {
		return model.NormalizedModel{}, fmt.Errorf("zen load: json unmarshal failed: %w", err)
	}

	return ToModel(&sessionData)
}

func (a *ZenAdapter) Save(profileDir string, m model.NormalizedModel) error {
	encoded, err := a.SaveBytes(m)
	if err != nil {
		return err
	}

	zenSessionsPath := adapter.ZenSessionsPath(profileDir)
	tmpPath := zenSessionsPath + ".tmp"
	if err := os.WriteFile(tmpPath, encoded, 0644); err != nil {
		return fmt.Errorf("zen save: failed writing tmp file: %w", err)
	}

	if err := os.Rename(tmpPath, zenSessionsPath); err != nil {
		return fmt.Errorf("zen save: failed renaming tmp file: %w", err)
	}

	return nil
}

func (a *ZenAdapter) SaveBytes(m model.NormalizedModel) ([]byte, error) {
	sessionData := FromModel(m)
	jsonBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("zen save: json marshal failed: %w", err)
	}

	encoded, err := mozlz4.Encode(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("zen save: mozlz4 encode failed: %w", err)
	}

	return encoded, nil
}

func ToModel(s *ZenSessionData) (model.NormalizedModel, error) {
	var m model.NormalizedModel
	m.Meta.SourceFormat = "zen"
	m.Meta.LastCollected = s.LastCollected
	if m.Meta.LastCollected == 0 {
		m.Meta.LastCollected = time.Now().UnixMilli()
	}

	tabMap := make(map[model.EntityID]model.Entity)

	for _, rawTab := range s.Tabs {
		if rawTab == nil {
			continue
		}

		// Filter placeholder tabs: zenIsEmpty == true && groupId == nil
		isEmpty := false
		if val, ok := rawTab["zenIsEmpty"].(bool); ok && val {
			isEmpty = true
		}

		gid, hasGID := rawTab["groupId"]
		if isEmpty && (!hasGID || gid == nil || gid == "") {
			// Drop placeholder tab
			continue
		}

		tabID := resolveZenTabID(rawTab)

		// Dedup on zenSyncId: prefer _zenIsActiveTab == true
		if existing, exists := tabMap[tabID]; exists {
			isActive := false
			if actVal, ok := rawTab["_zenIsActiveTab"].(bool); ok && actVal {
				isActive = true
			}
			if isActive {
				tabMap[tabID] = model.Entity{ID: tabID, Raw: rawTab}
			} else {
				_ = existing
			}
		} else {
			tabMap[tabID] = model.Entity{ID: tabID, Raw: rawTab}
		}
	}

	for _, e := range tabMap {
		m.Tabs = append(m.Tabs, e)
	}

	// Spaces
	for _, rawSpace := range s.Spaces {
		if rawSpace == nil {
			continue
		}
		var spaceID string
		if uuidVal, ok := rawSpace["uuid"].(string); ok && uuidVal != "" {
			spaceID = uuidVal
		} else {
			spaceID = fmt.Sprintf("space-%d", len(m.Spaces))
		}
		m.Spaces = append(m.Spaces, model.Entity{ID: model.EntityID(spaceID), Raw: rawSpace})
	}

	// Groups
	for _, rawGroup := range s.Groups {
		if rawGroup == nil {
			continue
		}
		var groupID string
		if idVal, ok := rawGroup["id"].(string); ok && idVal != "" {
			groupID = idVal
		} else {
			groupID = fmt.Sprintf("group-%d", len(m.Groups))
		}
		m.Groups = append(m.Groups, model.Entity{ID: model.EntityID(groupID), Raw: rawGroup})
	}

	// Folders
	for _, rawFolder := range s.Folders {
		if rawFolder == nil {
			continue
		}
		var folderID string
		if idVal, ok := rawFolder["id"].(string); ok && idVal != "" {
			folderID = idVal
		} else {
			folderID = fmt.Sprintf("folder-%d", len(m.Folders))
		}
		m.Folders = append(m.Folders, model.Entity{ID: model.EntityID(folderID), Raw: rawFolder})
	}

	// SplitViewData
	for idx, rawSV := range s.SplitViewData {
		if rawSV == nil {
			continue
		}
		var svID string
		if idVal, ok := rawSV["id"].(string); ok && idVal != "" {
			svID = idVal
		} else {
			svID = fmt.Sprintf("splitview-%d", idx)
		}
		m.SplitViewData = append(m.SplitViewData, model.Entity{ID: model.EntityID(svID), Raw: rawSV})
	}

	return m, nil
}

func FromModel(m model.NormalizedModel) *ZenSessionData {
	sessionData := &ZenSessionData{
		LastCollected: time.Now().UnixMilli(),
		Tabs:          []map[string]any{},
		Folders:       []map[string]any{},
		Groups:        []map[string]any{},
		Spaces:        []map[string]any{},
		SplitViewData: []map[string]any{},
	}

	if m.Meta.LastCollected > 0 {
		sessionData.LastCollected = m.Meta.LastCollected
	}

	for _, t := range m.Tabs {
		if t.Raw != nil {
			sessionData.Tabs = append(sessionData.Tabs, t.Raw)
		}
	}
	for _, f := range m.Folders {
		if f.Raw != nil {
			sessionData.Folders = append(sessionData.Folders, f.Raw)
		}
	}
	for _, g := range m.Groups {
		if g.Raw != nil {
			sessionData.Groups = append(sessionData.Groups, g.Raw)
		}
	}
	for _, s := range m.Spaces {
		if s.Raw != nil {
			sessionData.Spaces = append(sessionData.Spaces, s.Raw)
		}
	}
	for _, sv := range m.SplitViewData {
		if sv.Raw != nil {
			sessionData.SplitViewData = append(sessionData.SplitViewData, sv.Raw)
		}
	}

	return sessionData
}

func resolveZenTabID(rawTab map[string]any) model.EntityID {
	if syncID, ok := rawTab["zenSyncId"].(string); ok && syncID != "" {
		return model.EntityID(syncID)
	}

	var entryURLs []string
	if entries, ok := rawTab["entries"].([]any); ok {
		for _, eAny := range entries {
			if eMap, ok := eAny.(map[string]any); ok {
				if u, ok := eMap["url"].(string); ok {
					entryURLs = append(entryURLs, u)
				}
			}
		}
	}

	userContextID := 0
	if ucVal, ok := rawTab["userContextId"].(float64); ok {
		userContextID = int(ucVal)
	}

	pinned := false
	if pinVal, ok := rawTab["pinned"].(bool); ok {
		pinned = pinVal
	}

	return model.SynthesizeTabID(entryURLs, userContextID, pinned)
}
