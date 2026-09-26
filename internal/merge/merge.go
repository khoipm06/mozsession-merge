package merge

import (
	"github.com/khoipm06/mozsession-merge/internal/model"
)

type Change struct {
	ID     model.EntityID `json:"id"`
	Entity string         `json:"entity"` // "tab" | "space" | "group" | "folder" | "splitview"
	Kind   string         `json:"kind"`   // "added" | "removed" | "updated" | "conflict-resolved" | "warning" | "integrity-fix"
	Detail string         `json:"detail"`
}

type Conflict struct {
	ID       model.EntityID `json:"id"`
	Entity   string         `json:"entity"`
	Local    map[string]any `json:"local"`
	Remote   map[string]any `json:"remote"`
	Resolved map[string]any `json:"resolved"`
	Strategy Strategy       `json:"strategy"`
}

type Strategy int

const (
	StrategyNewerWins Strategy = iota
	StrategyKeepBoth
)

func (s Strategy) String() string {
	switch s {
	case StrategyNewerWins:
		return "newer-wins"
	case StrategyKeepBoth:
		return "keep-both"
	default:
		return "unknown"
	}
}

type Options struct {
	Strategy Strategy
	FirstRun bool // no baseline snapshot exists yet
}

// ThreeWayMerge merges base, local, and remote NormalizedModels.
func ThreeWayMerge(base, local, remote model.NormalizedModel, opts Options) (
	merged model.NormalizedModel, changes []Change, conflicts []Conflict, err error,
) {
	merged = model.NormalizedModel{
		Meta: local.Meta,
	}
	// Prefer larger LastCollected timestamp in Meta
	if remote.Meta.LastCollected > merged.Meta.LastCollected {
		merged.Meta.LastCollected = remote.Meta.LastCollected
	}

	if opts.FirstRun {
		changes = append(changes, Change{
			Entity: "meta",
			Kind:   "warning",
			Detail: "First-run merge: no baseline snapshot available. Performing union; prior deletions cannot be detected.",
		})
	}

	// 1. Merge Tabs
	var tabChanges []Change
	var tabConflicts []Conflict
	merged.Tabs, tabChanges, tabConflicts = diffEntities(base.Tabs, local.Tabs, remote.Tabs, "tab", opts, local.Meta.LastCollected, remote.Meta.LastCollected)
	changes = append(changes, tabChanges...)
	conflicts = append(conflicts, tabConflicts...)

	// 2. Merge Spaces
	var spaceChanges []Change
	var spaceConflicts []Conflict
	merged.Spaces, spaceChanges, spaceConflicts = diffEntities(base.Spaces, local.Spaces, remote.Spaces, "space", opts, local.Meta.LastCollected, remote.Meta.LastCollected)
	changes = append(changes, spaceChanges...)
	conflicts = append(conflicts, spaceConflicts...)

	// 3. Merge Groups
	var groupChanges []Change
	var groupConflicts []Conflict
	merged.Groups, groupChanges, groupConflicts = diffEntities(base.Groups, local.Groups, remote.Groups, "group", opts, local.Meta.LastCollected, remote.Meta.LastCollected)
	changes = append(changes, groupChanges...)
	conflicts = append(conflicts, groupConflicts...)

	// 4. Merge Folders
	var folderChanges []Change
	var folderConflicts []Conflict
	merged.Folders, folderChanges, folderConflicts = diffEntities(base.Folders, local.Folders, remote.Folders, "folder", opts, local.Meta.LastCollected, remote.Meta.LastCollected)
	changes = append(changes, folderChanges...)
	conflicts = append(conflicts, folderConflicts...)

	// 5. Merge SplitViewData
	var splitChanges []Change
	var splitConflicts []Conflict
	merged.SplitViewData, splitChanges, splitConflicts = diffEntities(base.SplitViewData, local.SplitViewData, remote.SplitViewData, "splitview", opts, local.Meta.LastCollected, remote.Meta.LastCollected)
	changes = append(changes, splitChanges...)
	conflicts = append(conflicts, splitConflicts...)

	// Closed tabs: passed through from the local side unchanged. Remote closed-tab history is not merged.
	merged.ClosedTabs = local.ClosedTabs

	// 6. Post-merge integrity sweep
	integrityChanges := sweepDanglingReferences(&merged)
	changes = append(changes, integrityChanges...)

	return merged, changes, conflicts, nil
}
