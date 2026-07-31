package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"mozsession-merge/internal/merge"
)

type Report struct {
	Timestamp time.Time          `json:"timestamp"`
	DryRun    bool               `json:"dryRun"`
	Sources   map[string]string  `json:"sources"` // profile -> file used
	Changes   []merge.Change     `json:"changes"`
	Conflicts []merge.Conflict   `json:"conflicts"`
}

func NewReport(dryRun bool, sources map[string]string, changes []merge.Change, conflicts []merge.Conflict) Report {
	return Report{
		Timestamp: time.Now(),
		DryRun:    dryRun,
		Sources:   sources,
		Changes:   changes,
		Conflicts: conflicts,
	}
}

func (r Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("report write json: marshal failed: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (r Report) WriteHuman(w io.Writer) error {
	modeStr := "LIVE"
	if r.DryRun {
		modeStr = "DRY RUN"
	}

	fmt.Fprintf(w, "=== Session 3-Way Merge Report [%s] ===\n", modeStr)
	fmt.Fprintf(w, "Timestamp: %s\n\n", r.Timestamp.Format(time.RFC3339))

	if len(r.Sources) > 0 {
		fmt.Fprintln(w, "--- Sources ---")
		for src, file := range r.Sources {
			fmt.Fprintf(w, "  %s: %s\n", src, file)
		}
		fmt.Fprintln(w)
	}

	// Summary stats per entity kind
	stats := make(map[string]map[string]int) // entity -> kind -> count
	for _, c := range r.Changes {
		if stats[c.Entity] == nil {
			stats[c.Entity] = make(map[string]int)
		}
		stats[c.Entity][c.Kind]++
	}

	fmt.Fprintln(w, "--- Summary of Changes ---")
	if len(stats) == 0 {
		fmt.Fprintln(w, "  No changes detected.")
	} else {
		for entity, kinds := range stats {
			fmt.Fprintf(w, "  %s: ", entity)
			first := true
			for k, v := range kinds {
				if !first {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprintf(w, "%d %s", v, k)
				first = false
			}
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintln(w)

	if len(r.Conflicts) > 0 {
		fmt.Fprintf(w, "--- Conflicts (%d resolved) ---\n", len(r.Conflicts))
		for _, conf := range r.Conflicts {
			fmt.Fprintf(w, "  [%s] ID: %s (Strategy: %s)\n", conf.Entity, conf.ID, conf.Strategy)
		}
		fmt.Fprintln(w)
	}

	if len(r.Changes) > 0 {
		fmt.Fprintln(w, "--- Detailed Change Log ---")
		for _, c := range r.Changes {
			fmt.Fprintf(w, "  [%s] %s %s: %s\n", c.Entity, c.Kind, c.ID, c.Detail)
		}
	}

	return nil
}
