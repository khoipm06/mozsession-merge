package firefox

// FirefoxSessionData represents the raw unmarshaled structure of sessionstore.jsonlz4
type FirefoxSessionData struct {
	Version        any              `json:"version,omitempty"`
	Windows        []map[string]any `json:"windows"`
	SelectedWindow int              `json:"selectedWindow,omitempty"`
	ClosedWindows  []map[string]any `json:"_closedWindows,omitempty"`
	SavedGroups    []map[string]any `json:"savedGroups,omitempty"`
	MaxSplitViewID any              `json:"maxSplitViewId,omitempty"`
	Session        map[string]any   `json:"session,omitempty"`
	Global         map[string]any   `json:"global,omitempty"`
	Cookies        []map[string]any `json:"cookies,omitempty"`
}
