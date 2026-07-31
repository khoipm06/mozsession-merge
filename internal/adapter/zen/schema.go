package zen

// ZenSessionData represents the raw unmarshaled structure of zen-sessions.jsonlz4
type ZenSessionData struct {
	LastCollected int64            `json:"lastCollected"`
	Tabs          []map[string]any `json:"tabs"`
	Folders       []map[string]any `json:"folders"`
	Groups        []map[string]any `json:"groups"`
	Spaces        []map[string]any `json:"spaces"`
	SplitViewData []map[string]any `json:"splitViewData"`
}
