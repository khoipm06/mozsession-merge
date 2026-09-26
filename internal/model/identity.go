package model

import (
	"fmt"
	"hash/fnv"
)

// SynthesizeTabID generates a deterministic ID for non-Zen tabs by hashing
// full navigation history, userContextID, and pinned state.
func SynthesizeTabID(entryURLs []string, userContextID int, pinned bool) EntityID {
	h := fnv.New64a()
	for _, u := range entryURLs {
		h.Write([]byte(u))
		h.Write([]byte{0})
	}
	fmt.Fprintf(h, "%d:%t", userContextID, pinned)
	return EntityID(fmt.Sprintf("synth-%x", h.Sum64()))
}

// ResolveTabID returns the canonical EntityID for a raw tab map.
// Prefers the zenSyncId field when present; otherwise synthesizes a
// deterministic ID from the tab's navigation history via SynthesizeTabID.
func ResolveTabID(rawTab map[string]any) EntityID {
	if syncID, ok := rawTab["zenSyncId"].(string); ok && syncID != "" {
		return EntityID(syncID)
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
	return SynthesizeTabID(entryURLs, userContextID, pinned)
}
