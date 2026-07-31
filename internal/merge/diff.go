package merge

import (
	"fmt"

	"mozsession-merge/internal/model"
)

func diffEntities(
	baseEntities, localEntities, remoteEntities []model.Entity,
	entityKind string,
	opts Options,
	localMeta, remoteMeta int64,
) (merged []model.Entity, changes []Change, conflicts []Conflict) {

	baseMap := make(map[model.EntityID]model.Entity, len(baseEntities))
	for _, e := range baseEntities {
		baseMap[e.ID] = e
	}

	localMap := make(map[model.EntityID]model.Entity, len(localEntities))
	for _, e := range localEntities {
		localMap[e.ID] = e
	}

	remoteMap := make(map[model.EntityID]model.Entity, len(remoteEntities))
	for _, e := range remoteEntities {
		remoteMap[e.ID] = e
	}

	// Order of processing: local items first, then remote-only items
	var orderedIDs []model.EntityID
	seenIDs := make(map[model.EntityID]bool)

	for _, e := range localEntities {
		if !seenIDs[e.ID] {
			seenIDs[e.ID] = true
			orderedIDs = append(orderedIDs, e.ID)
		}
	}
	for _, e := range remoteEntities {
		if !seenIDs[e.ID] {
			seenIDs[e.ID] = true
			orderedIDs = append(orderedIDs, e.ID)
		}
	}
	for _, e := range baseEntities {
		if !seenIDs[e.ID] {
			seenIDs[e.ID] = true
			orderedIDs = append(orderedIDs, e.ID)
		}
	}

	if opts.FirstRun {
		for _, id := range orderedIDs {
			l, inLocal := localMap[id]
			r, inRemote := remoteMap[id]

			if inLocal && inRemote {
				if model.IsContentEqual(l.Raw, r.Raw) {
					merged = append(merged, l)
				} else {
					res, conf := resolveConflict(l, r, localMeta, remoteMeta, entityKind, opts.Strategy)
					merged = append(merged, res)
					conflicts = append(conflicts, conf)
					changes = append(changes, Change{
						ID:     id,
						Entity: entityKind,
						Kind:   "conflict-resolved",
						Detail: fmt.Sprintf("first-run content conflict resolved using %s strategy", opts.Strategy),
					})
					if opts.Strategy == StrategyKeepBoth {
						dupID := model.EntityID(string(id) + "-remote")
						dupRaw := make(map[string]any, len(r.Raw))
						for k, v := range r.Raw {
							dupRaw[k] = v
						}
						dupEntity := model.Entity{ID: dupID, Raw: dupRaw}
						merged = append(merged, dupEntity)
						changes = append(changes, Change{
							ID:     dupID,
							Entity: entityKind,
							Kind:   "added",
							Detail: "kept remote duplicate under StrategyKeepBoth",
						})
					}
				}
			} else if inLocal {
				merged = append(merged, l)
			} else if inRemote {
				merged = append(merged, r)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "added",
					Detail: "first-run union addition from remote",
				})
			}
		}
		return merged, changes, conflicts
	}

	// Standard 3-way diff (with baseline)
	for _, id := range orderedIDs {
		b, inBase := baseMap[id]
		l, inLocal := localMap[id]
		r, inRemote := remoteMap[id]

		if inBase && inLocal && inRemote {
			localUnchanged := model.IsContentEqual(b.Raw, l.Raw)
			remoteUnchanged := model.IsContentEqual(b.Raw, r.Raw)

			if localUnchanged && remoteUnchanged {
				merged = append(merged, l)
			} else if !localUnchanged && remoteUnchanged {
				merged = append(merged, l)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "updated",
					Detail: "updated on local side",
				})
			} else if localUnchanged && !remoteUnchanged {
				merged = append(merged, r)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "updated",
					Detail: "updated on remote side",
				})
			} else {
				// Changed on both sides
				if model.IsContentEqual(l.Raw, r.Raw) {
					merged = append(merged, l)
				} else {
					res, conf := resolveConflict(l, r, localMeta, remoteMeta, entityKind, opts.Strategy)
					merged = append(merged, res)
					conflicts = append(conflicts, conf)
					changes = append(changes, Change{
						ID:     id,
						Entity: entityKind,
						Kind:   "conflict-resolved",
						Detail: fmt.Sprintf("conflicting modification resolved using %s strategy", opts.Strategy),
					})
					if opts.Strategy == StrategyKeepBoth {
						dupID := model.EntityID(string(id) + "-remote")
						dupRaw := make(map[string]any, len(r.Raw))
						for k, v := range r.Raw {
							dupRaw[k] = v
						}
						dupEntity := model.Entity{ID: dupID, Raw: dupRaw}
						merged = append(merged, dupEntity)
						changes = append(changes, Change{
							ID:     dupID,
							Entity: entityKind,
							Kind:   "added",
							Detail: "kept remote duplicate under StrategyKeepBoth",
						})
					}
				}
			}
		} else if inBase && !inLocal && !inRemote {
			// Deleted by both
			changes = append(changes, Change{
				ID:     id,
				Entity: entityKind,
				Kind:   "removed",
				Detail: "deleted on both local and remote sides",
			})
		} else if inBase && !inLocal && inRemote {
			remoteUnchanged := model.IsContentEqual(b.Raw, r.Raw)
			if remoteUnchanged {
				// Local deleted it, remote didn't change it -> drop
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "removed",
					Detail: "deleted on local side",
				})
			} else {
				// Deleted on local vs updated on remote -> Conflict
				res, conf := resolveConflict(model.Entity{ID: id, Raw: nil}, r, localMeta, remoteMeta, entityKind, opts.Strategy)
				if res.Raw != nil {
					merged = append(merged, res)
				}
				conflicts = append(conflicts, conf)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "conflict-resolved",
					Detail: "delete vs update conflict resolved",
				})
			}
		} else if inBase && inLocal && !inRemote {
			localUnchanged := model.IsContentEqual(b.Raw, l.Raw)
			if localUnchanged {
				// Remote deleted it, local didn't change it -> drop
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "removed",
					Detail: "deleted on remote side",
				})
			} else {
				// Updated on local vs deleted on remote -> Conflict
				res, conf := resolveConflict(l, model.Entity{ID: id, Raw: nil}, localMeta, remoteMeta, entityKind, opts.Strategy)
				if res.Raw != nil {
					merged = append(merged, res)
				}
				conflicts = append(conflicts, conf)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "conflict-resolved",
					Detail: "update vs delete conflict resolved",
				})
			}
		} else if !inBase && inLocal && !inRemote {
			merged = append(merged, l)
			changes = append(changes, Change{
				ID:     id,
				Entity: entityKind,
				Kind:   "added",
				Detail: "added on local side",
			})
		} else if !inBase && !inLocal && inRemote {
			merged = append(merged, r)
			changes = append(changes, Change{
				ID:     id,
				Entity: entityKind,
				Kind:   "added",
				Detail: "added on remote side",
			})
		} else if !inBase && inLocal && inRemote {
			// Added independently on both sides
			if model.IsContentEqual(l.Raw, r.Raw) {
				merged = append(merged, l)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "added",
					Detail: "added on both sides identically",
				})
			} else {
				res, conf := resolveConflict(l, r, localMeta, remoteMeta, entityKind, opts.Strategy)
				merged = append(merged, res)
				conflicts = append(conflicts, conf)
				changes = append(changes, Change{
					ID:     id,
					Entity: entityKind,
					Kind:   "conflict-resolved",
					Detail: "independent addition conflict resolved",
				})
				if opts.Strategy == StrategyKeepBoth {
					dupID := model.EntityID(string(id) + "-remote")
					dupRaw := make(map[string]any, len(r.Raw))
					for k, v := range r.Raw {
						dupRaw[k] = v
					}
					dupEntity := model.Entity{ID: dupID, Raw: dupRaw}
					merged = append(merged, dupEntity)
					changes = append(changes, Change{
						ID:     dupID,
						Entity: entityKind,
						Kind:   "added",
						Detail: "kept remote duplicate under StrategyKeepBoth",
					})
				}
			}
		}
	}

	return merged, changes, conflicts
}
