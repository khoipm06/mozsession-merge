package merge

import (
	"github.com/khoipm06/mozsession-merge/internal/model"
)

func getTimestamp(raw map[string]any, fallback int64) int64 {
	if raw == nil {
		return fallback
	}
	if val, ok := raw["lastAccessed"]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	if val, ok := raw["closedAt"]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return fallback
}

func resolveConflict(
	local, remote model.Entity,
	localMeta, remoteMeta int64,
	entityKind string,
	s Strategy,
) (resolved model.Entity, conflict Conflict) {
	conflict = Conflict{
		ID:       local.ID,
		Entity:   entityKind,
		Local:    local.Raw,
		Remote:   remote.Raw,
		Strategy: s,
	}

	switch s {
	case StrategyNewerWins:
		localTs := getTimestamp(local.Raw, localMeta)
		remoteTs := getTimestamp(remote.Raw, remoteMeta)

		if remoteTs > localTs {
			resolved = remote
			conflict.Resolved = remote.Raw
		} else {
			resolved = local
			conflict.Resolved = local.Raw
		}

	case StrategyKeepBoth:
		// Local version remains primary under local.ID.
		resolved = local
		conflict.Resolved = local.Raw
	}

	return resolved, conflict
}
