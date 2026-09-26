package merge

import (
	"fmt"
	"github.com/khoipm06/mozsession-merge/internal/model"
)

func sweepDanglingReferences(m *model.NormalizedModel) []Change {
	var changes []Change

	groupIDs := make(map[string]bool, len(m.Groups))
	for _, g := range m.Groups {
		groupIDs[string(g.ID)] = true
	}

	spaceUUIDs := make(map[string]bool, len(m.Spaces))
	for _, s := range m.Spaces {
		spaceUUIDs[string(s.ID)] = true
	}

	folderIDs := make(map[string]bool, len(m.Folders))
	for _, f := range m.Folders {
		folderIDs[string(f.ID)] = true
	}

	tabIDs := make(map[string]bool, len(m.Tabs))
	for _, t := range m.Tabs {
		tabIDs[string(t.ID)] = true
	}

	// 1. Clean tabs
	for i := range m.Tabs {
		raw := m.Tabs[i].Raw
		if raw == nil {
			continue
		}

		if gidVal, exists := raw["groupId"]; exists && gidVal != nil {
			if gidStr, ok := gidVal.(string); ok && gidStr != "" {
				if !groupIDs[gidStr] {
					delete(raw, "groupId")
					changes = append(changes, Change{
						ID:     m.Tabs[i].ID,
						Entity: "tab",
						Kind:   "integrity-fix",
						Detail: fmt.Sprintf("removed dangling groupId reference %q", gidStr),
					})
				}
			}
		}

		if wsVal, exists := raw["zenWorkspace"]; exists && wsVal != nil {
			if wsStr, ok := wsVal.(string); ok && wsStr != "" {
				if !spaceUUIDs[wsStr] {
					delete(raw, "zenWorkspace")
					changes = append(changes, Change{
						ID:     m.Tabs[i].ID,
						Entity: "tab",
						Kind:   "integrity-fix",
						Detail: fmt.Sprintf("removed dangling zenWorkspace reference %q", wsStr),
					})
				}
			}
		}
	}

	// 2. Clean folders
	for i := range m.Folders {
		raw := m.Folders[i].Raw
		if raw == nil {
			continue
		}

		if wsVal, exists := raw["workspaceId"]; exists && wsVal != nil {
			if wsStr, ok := wsVal.(string); ok && wsStr != "" {
				if !spaceUUIDs[wsStr] {
					delete(raw, "workspaceId")
					changes = append(changes, Change{
						ID:     m.Folders[i].ID,
						Entity: "folder",
						Kind:   "integrity-fix",
						Detail: fmt.Sprintf("removed dangling workspaceId reference %q", wsStr),
					})
				}
			}
		}

		if pidVal, exists := raw["parentId"]; exists && pidVal != nil {
			if pidStr, ok := pidVal.(string); ok && pidStr != "" {
				if !folderIDs[pidStr] {
					raw["parentId"] = nil
					changes = append(changes, Change{
						ID:     m.Folders[i].ID,
						Entity: "folder",
						Kind:   "integrity-fix",
						Detail: fmt.Sprintf("cleared dangling parentId reference %q", pidStr),
					})
				}
			}
		}
	}

	// 3. Clean splitViewData
	var validSplitViewData []model.Entity
	for _, sv := range m.SplitViewData {
		raw := sv.Raw
		if raw == nil {
			validSplitViewData = append(validSplitViewData, sv)
			continue
		}
		// If splitview references tab1 / tab2 IDs, check if both exist
		dangling := false
		if t1, exists := raw["tab1"]; exists {
			if t1Str, ok := t1.(string); ok && !tabIDs[t1Str] {
				dangling = true
			}
		}
		if t2, exists := raw["tab2"]; exists {
			if t2Str, ok := t2.(string); ok && !tabIDs[t2Str] {
				dangling = true
			}
		}
		if dangling {
			changes = append(changes, Change{
				ID:     sv.ID,
				Entity: "splitview",
				Kind:   "integrity-fix",
				Detail: "dropped splitViewData entry referencing removed tabs",
			})
		} else {
			validSplitViewData = append(validSplitViewData, sv)
		}
	}
	m.SplitViewData = validSplitViewData

	return changes
}
