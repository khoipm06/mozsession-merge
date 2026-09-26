package firefox

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/khoipm06/mozsession-merge/internal/adapter"
	"github.com/khoipm06/mozsession-merge/internal/model"
	"github.com/khoipm06/mozsession-merge/internal/mozlz4"
	"github.com/khoipm06/mozsession-merge/internal/writer"
)

type FirefoxAdapter struct{}

func NewAdapter() *FirefoxAdapter {
	return &FirefoxAdapter{}
}

func (a *FirefoxAdapter) Name() string {
	return "firefox"
}

func (a *FirefoxAdapter) Detect(profileDir string) (bool, error) {
	sessionStore := adapter.SessionStorePath(profileDir)
	zenSessions := adapter.ZenSessionsPath(profileDir)

	if adapter.FileExists(sessionStore) && !adapter.FileExists(zenSessions) {
		return true, nil
	}
	return false, nil
}

func (a *FirefoxAdapter) Load(profileDir string) (model.NormalizedModel, error) {
	sessionStorePath := adapter.SessionStorePath(profileDir)
	data, err := os.ReadFile(sessionStorePath)
	if err != nil {
		return model.NormalizedModel{}, fmt.Errorf("firefox load: failed to read %s: %w", sessionStorePath, err)
	}

	return a.LoadBytes(data)
}

func (a *FirefoxAdapter) LoadBytes(data []byte) (model.NormalizedModel, error) {
	jsonBytes, err := mozlz4.Decode(data)
	if err != nil {
		return model.NormalizedModel{}, fmt.Errorf("firefox load: mozlz4 decode failed: %w", err)
	}

	var sessionData FirefoxSessionData
	if err := json.Unmarshal(jsonBytes, &sessionData); err != nil {
		return model.NormalizedModel{}, fmt.Errorf("firefox load: json unmarshal failed: %w", err)
	}

	return ToModel(&sessionData)
}

func (a *FirefoxAdapter) Save(profileDir string, m model.NormalizedModel) error {
	encoded, err := a.SaveBytes(m)
	if err != nil {
		return err
	}
	sessionStorePath := adapter.SessionStorePath(profileDir)
	if err := writer.AtomicWrite(sessionStorePath, encoded, false); err != nil {
		return fmt.Errorf("firefox save: %w", err)
	}
	return nil
}

func (a *FirefoxAdapter) SaveBytes(m model.NormalizedModel) ([]byte, error) {
	sessionData := FromModel(m)
	jsonBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("firefox save: json marshal failed: %w", err)
	}

	encoded, err := mozlz4.Encode(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("firefox save: mozlz4 encode failed: %w", err)
	}

	return encoded, nil
}

// ToModel converts FirefoxSessionData into a NormalizedModel
func ToModel(s *FirefoxSessionData) (model.NormalizedModel, error) {
	var m model.NormalizedModel
	m.Meta.SourceFormat = "firefox"

	if s.Session != nil {
		if lu, ok := s.Session["lastUpdate"].(float64); ok {
			m.Meta.LastCollected = int64(lu)
		}
	}
	if m.Meta.LastCollected == 0 {
		m.Meta.LastCollected = time.Now().UnixMilli()
	}

	tabMap := make(map[model.EntityID]model.Entity)
	groupMap := make(map[model.EntityID]model.Entity)
	folderMap := make(map[model.EntityID]model.Entity)
	spaceMap := make(map[model.EntityID]model.Entity)
	splitMap := make(map[model.EntityID]model.Entity)

	var windowLayouts []model.WindowLayout

	for winIdx, win := range s.Windows {
		winID := fmt.Sprintf("win-%d", winIdx)
		layout := model.WindowLayout{
			WindowID: winID,
			Geometry: make(map[string]any),
		}

		// Extract geometry
		for _, geoKey := range []string{"width", "height", "screenX", "screenY", "sizemode", "sidebar", "workspaceID", "title", "chromeFlags"} {
			if val, ok := win[geoKey]; ok {
				layout.Geometry[geoKey] = val
			}
		}

		// Process tabs in window
		if rawTabs, ok := win["tabs"].([]any); ok {
			for tabIdx, rawTabAny := range rawTabs {
				rawTab, ok := rawTabAny.(map[string]any)
				if !ok {
					continue
				}

				tabID := resolveTabID(rawTab)
				layout.TabIDs = append(layout.TabIDs, tabID)

				// Selected tab check (selected is 1-based index)
				selectedIdx := 1
				if selVal, ok := win["selected"].(float64); ok {
					selectedIdx = int(selVal)
				}
				if tabIdx+1 == selectedIdx {
					layout.Selected = tabID
				}

				// Deduplication: if ID already seen, only replace if this instance is the active tab.
				if _, exists := tabMap[tabID]; exists {
					if actVal, ok := rawTab["_zenIsActiveTab"].(bool); ok && actVal {
						tabMap[tabID] = model.Entity{ID: tabID, Raw: rawTab}
					}
				} else {
					tabMap[tabID] = model.Entity{ID: tabID, Raw: rawTab}
				}
			}
		}

		windowLayouts = append(windowLayouts, layout)

		// Extract groups
		extractEntities(win, "groups", "id", groupMap)

		// Extract folders
		extractEntities(win, "folders", "id", folderMap)

		// Extract spaces
		extractEntities(win, "spaces", "uuid", spaceMap)

		// Extract splitViewData
		extractEntities(win, "splitViewData", "id", splitMap)
	}

	m.Meta.WindowLayout = windowLayouts

	// Convert maps to slices with deterministic ordering by entity ID.
	for _, e := range tabMap {
		m.Tabs = append(m.Tabs, e)
	}
	for _, e := range groupMap {
		m.Groups = append(m.Groups, e)
	}
	for _, e := range folderMap {
		m.Folders = append(m.Folders, e)
	}
	for _, e := range spaceMap {
		m.Spaces = append(m.Spaces, e)
	}
	for _, e := range splitMap {
		m.SplitViewData = append(m.SplitViewData, e)
	}
	sort.Slice(m.Tabs, func(i, j int) bool { return m.Tabs[i].ID < m.Tabs[j].ID })
	sort.Slice(m.Groups, func(i, j int) bool { return m.Groups[i].ID < m.Groups[j].ID })
	sort.Slice(m.Folders, func(i, j int) bool { return m.Folders[i].ID < m.Folders[j].ID })
	sort.Slice(m.Spaces, func(i, j int) bool { return m.Spaces[i].ID < m.Spaces[j].ID })
	sort.Slice(m.SplitViewData, func(i, j int) bool { return m.SplitViewData[i].ID < m.SplitViewData[j].ID })

	return m, nil
}

func FromModel(m model.NormalizedModel) *FirefoxSessionData {
	sessionData := &FirefoxSessionData{
		Version: []any{"sessionrestore", 1},
		Session: map[string]any{
			"lastUpdate": float64(time.Now().UnixMilli()),
			"startTime":  float64(time.Now().UnixMilli() - 3600000),
		},
		Windows: []map[string]any{},
	}

	tabMap := make(map[model.EntityID]model.Entity, len(m.Tabs))
	for _, t := range m.Tabs {
		tabMap[t.ID] = t
	}

	// Prepare raw slices for groups, folders, spaces, splitViewData
	var rawGroups []map[string]any
	for _, g := range m.Groups {
		if g.Raw != nil {
			rawGroups = append(rawGroups, g.Raw)
		}
	}
	var rawFolders []map[string]any
	for _, f := range m.Folders {
		if f.Raw != nil {
			rawFolders = append(rawFolders, f.Raw)
		}
	}
	var rawSpaces []map[string]any
	for _, s := range m.Spaces {
		if s.Raw != nil {
			rawSpaces = append(rawSpaces, s.Raw)
		}
	}
	var rawSplitViewData []map[string]any
	for _, sv := range m.SplitViewData {
		if sv.Raw != nil {
			rawSplitViewData = append(rawSplitViewData, sv.Raw)
		}
	}

	if len(m.Meta.WindowLayout) > 0 {
		// Multi-window reconstruction for pure Firefox
		assignedTabIDs := make(map[model.EntityID]bool)

		for winIdx, layout := range m.Meta.WindowLayout {
			winMap := make(map[string]any)
			for k, v := range layout.Geometry {
				winMap[k] = v
			}

			var winTabs []map[string]any
			selectedIdx := 1

			for _, tabID := range layout.TabIDs {
				if t, ok := tabMap[tabID]; ok {
					if t.Raw != nil {
						winTabs = append(winTabs, t.Raw)
						assignedTabIDs[tabID] = true
						if tabID == layout.Selected {
							selectedIdx = len(winTabs)
						}
					}
				}
			}

			// Assign unassigned merged tabs (new additions) to primary window (winIdx == 0)
			if winIdx == 0 {
				for _, t := range m.Tabs {
					if !assignedTabIDs[t.ID] && t.Raw != nil {
						winTabs = append(winTabs, t.Raw)
						assignedTabIDs[t.ID] = true
					}
				}
			}

			if len(winTabs) == 0 {
				winTabs = append(winTabs, map[string]any{
					"entries": []any{
						map[string]any{"url": "about:blank"},
					},
					"index": 1,
				})
			}

			winMap["tabs"] = winTabs
			winMap["selected"] = selectedIdx
			winMap["groups"] = rawGroups
			winMap["folders"] = rawFolders
			winMap["spaces"] = rawSpaces
			winMap["splitViewData"] = rawSplitViewData

			sessionData.Windows = append(sessionData.Windows, winMap)
		}
	} else {
		// Single window fallback
		winMap := map[string]any{
			"width":  1280,
			"height": 800,
		}
		var winTabs []map[string]any
		for _, t := range m.Tabs {
			if t.Raw != nil {
				winTabs = append(winTabs, t.Raw)
			}
		}
		if len(winTabs) == 0 {
			winTabs = append(winTabs, map[string]any{
				"entries": []any{
					map[string]any{"url": "about:blank"},
				},
				"index": 1,
			})
		}

		winMap["tabs"] = winTabs
		winMap["selected"] = len(winTabs)
		winMap["groups"] = rawGroups
		winMap["folders"] = rawFolders
		winMap["spaces"] = rawSpaces
		winMap["splitViewData"] = rawSplitViewData

		sessionData.Windows = append(sessionData.Windows, winMap)
	}

	sessionData.SelectedWindow = 1
	return sessionData
}

func resolveTabID(rawTab map[string]any) model.EntityID {
	return model.ResolveTabID(rawTab)
}

func extractEntities(win map[string]any, fieldKey, idKey string, targetMap map[model.EntityID]model.Entity) {
	if rawSlice, ok := win[fieldKey].([]any); ok {
		for idx, rawAny := range rawSlice {
			rawMap, ok := rawAny.(map[string]any)
			if !ok {
				continue
			}
			var idStr string
			if idVal, ok := rawMap[idKey].(string); ok && idVal != "" {
				idStr = idVal
			} else {
				idStr = fmt.Sprintf("%s-%d", fieldKey, idx)
			}
			entID := model.EntityID(idStr)
			if _, exists := targetMap[entID]; !exists {
				targetMap[entID] = model.Entity{ID: entID, Raw: rawMap}
			}
		}
	}
}
