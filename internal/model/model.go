package model

type EntityID string

type Entity struct {
	ID  EntityID       `json:"id"`
	Raw map[string]any `json:"raw"` // original, browser-native fields — opaque to merge core
}

type WindowLayout struct {
	WindowID string         `json:"windowId"`
	TabIDs   []EntityID     `json:"tabIds"`
	Selected EntityID       `json:"selected"`
	Geometry map[string]any `json:"geometry"`
}

type Meta struct {
	LastCollected int64          `json:"lastCollected"`
	SourceFormat  string         `json:"sourceFormat"` // "zen" | "firefox"
	WindowLayout  []WindowLayout `json:"windowLayout"`
}

type NormalizedModel struct {
	Tabs          []Entity `json:"tabs"`
	Spaces        []Entity `json:"spaces"`
	Groups        []Entity `json:"groups"`
	Folders       []Entity `json:"folders"`
	SplitViewData []Entity `json:"splitViewData"`
	ClosedTabs    []Entity `json:"closedTabs"`
	Meta          Meta     `json:"meta"`
}
