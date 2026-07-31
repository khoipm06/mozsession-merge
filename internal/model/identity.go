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
