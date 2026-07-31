package model

import (
	"bytes"
	"encoding/json"
	"reflect"
)

var ephemeralFields = map[string]bool{
	"_zenIsActiveTab": true,
	"lastAccessed":    true,
	"index":           true,
	"requestedIndex":  true,
	"userTypedValue":  true,
	"userTypedClear":  true,
}

// StripEphemeralFields produces a shallow/deep copy of raw map with ephemeral
// session-instance fields removed.
func StripEphemeralFields(raw map[string]any) map[string]any {
	if raw == nil {
		return nil
	}
	stripped := make(map[string]any, len(raw))
	for k, v := range raw {
		if ephemeralFields[k] {
			continue
		}
		stripped[k] = v
	}
	return stripped
}

// IsContentEqual compares two raw entity maps after stripping ephemeral fields.
func IsContentEqual(a, b map[string]any) bool {
	strippedA := StripEphemeralFields(a)
	strippedB := StripEphemeralFields(b)

	if reflect.DeepEqual(strippedA, strippedB) {
		return true
	}

	// Fallback to canonical JSON comparison if map types differ (e.g. float64 vs int)
	bytesA, errA := json.Marshal(strippedA)
	bytesB, errB := json.Marshal(strippedB)
	if errA == nil && errB == nil {
		return bytes.Equal(bytesA, bytesB)
	}

	return false
}
