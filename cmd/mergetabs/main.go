package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/khoipm06/mozsession-merge/internal/adapter"
	"github.com/khoipm06/mozsession-merge/internal/adapter/firefox"
	"github.com/khoipm06/mozsession-merge/internal/adapter/zen"
	"github.com/khoipm06/mozsession-merge/internal/backupfallback"
	"github.com/khoipm06/mozsession-merge/internal/merge"
	"github.com/khoipm06/mozsession-merge/internal/model"
	"github.com/khoipm06/mozsession-merge/internal/profile"
	"github.com/khoipm06/mozsession-merge/internal/report"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

var (
	dryRun           bool
	conflictStrategy string
	reportFile       string
	stateDir         string
	force            bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "mergetabs <profileA> <profileB>",
		Short:   "Browser Session 3-Way Merge Tool (Firefox & Zen)",
		Long:    "Safe, offline, baseline-snapshot-based 3-way merge tool for Firefox and Zen browser session stores.",
		Args:    cobra.ExactArgs(2),
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMergeTabs(args[0], args[1])
		},
	}

	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a trial merge without modifying profile session files")
	rootCmd.Flags().StringVarP(&conflictStrategy, "conflict-strategy", "c", "newer", "Conflict resolution strategy (newer | keep-both)")
	rootCmd.Flags().StringVarP(&reportFile, "report-file", "r", "merge-report.json", "Output path for structured JSON change report")
	rootCmd.Flags().StringVarP(&stateDir, "state-dir", "s", defaultStateDir(), "Directory where baseline snapshots and tool state are stored")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Bypass active profile lock checks")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func defaultStateDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".mozsession-merge"
	}
	return filepath.Join(dir, "mozsession-merge")
}

func parseStrategy(s string) (merge.Strategy, error) {
	switch s {
	case "newer", "newer-wins":
		return merge.StrategyNewerWins, nil
	case "keep-both":
		return merge.StrategyKeepBoth, nil
	default:
		return 0, fmt.Errorf("invalid conflict strategy %q (allowed: newer, keep-both)", s)
	}
}

func profileID(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	h := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%x", h[:8])
}

func runMergeTabs(profileA, profileB string) error {
	strat, err := parseStrategy(conflictStrategy)
	if err != nil {
		return err
	}

	// 1. Profile-specific lock checks
	if !force {
		lockedA, reasonA, errA := profile.IsLocked(profileA)
		if errA == nil && lockedA {
			return fmt.Errorf("profile A (%s) is locked: %s. Close browser or use --force", profileA, reasonA)
		}

		lockedB, reasonB, errB := profile.IsLocked(profileB)
		if errB == nil && lockedB {
			return fmt.Errorf("profile B (%s) is locked: %s. Close browser or use --force", profileB, reasonB)
		}
	}

	// 2. Browser detection
	kindA, err := profile.DetectBrowser(profileA)
	if err != nil {
		return fmt.Errorf("detect profile A (%s) failed: %w", profileA, err)
	}

	kindB, err := profile.DetectBrowser(profileB)
	if err != nil {
		return fmt.Errorf("detect profile B (%s) failed: %w", profileB, err)
	}

	ffAdapter := firefox.NewAdapter()
	zenAdapter := zen.NewAdapter()

	var adapterA, adapterB adapter.Adapter
	if kindA == "zen" {
		adapterA = zenAdapter
	} else {
		adapterA = ffAdapter
	}
	if kindB == "zen" {
		adapterB = zenAdapter
	} else {
		adapterB = ffAdapter
	}

	// 3. Load profile A session data (with fallback chain)
	primaryA := adapter.SessionStorePath(profileA)
	if kindA == "zen" {
		primaryA = adapter.ZenSessionsPath(profileA)
	}
	fallbacksA := backupfallback.DefaultFallbackCandidates(profileA, kindA)
	rawA, _, usedA, err := backupfallback.LoadWithFallback(primaryA, fallbacksA)
	if err != nil {
		return fmt.Errorf("failed loading session for profile A (%s): %w", profileA, err)
	}

	// 4. Load profile B session data (with fallback chain)
	primaryB := adapter.SessionStorePath(profileB)
	if kindB == "zen" {
		primaryB = adapter.ZenSessionsPath(profileB)
	}
	fallbacksB := backupfallback.DefaultFallbackCandidates(profileB, kindB)
	rawB, _, usedB, err := backupfallback.LoadWithFallback(primaryB, fallbacksB)
	if err != nil {
		return fmt.Errorf("failed loading session for profile B (%s): %w", profileB, err)
	}

	// Convert raw bytes to NormalizedModel
	var modelA, modelB model.NormalizedModel
	if kindA == "zen" {
		modelA, err = zenAdapter.LoadBytes(rawA)
	} else {
		modelA, err = ffAdapter.LoadBytes(rawA)
	}
	if err != nil {
		return fmt.Errorf("convert profile A to model failed: %w", err)
	}

	if kindB == "zen" {
		modelB, err = zenAdapter.LoadBytes(rawB)
	} else {
		modelB, err = ffAdapter.LoadBytes(rawB)
	}
	if err != nil {
		return fmt.Errorf("convert profile B to model failed: %w", err)
	}

	// 5. Load baseline snapshot (check A first, fallback to B)
	idA := profileID(profileA)
	idB := profileID(profileB)

	baseModelA, hasBaseA, errBaseA := profile.LoadBaseline(stateDir, idA)
	if errBaseA != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed loading baseline for profile A: %v\n", errBaseA)
	}
	baseModelB, hasBaseB, errBaseB := profile.LoadBaseline(stateDir, idB)
	if errBaseB != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed loading baseline for profile B: %v\n", errBaseB)
	}

	var baseModel model.NormalizedModel
	hasBase := false

	if hasBaseA {
		baseModel = baseModelA
		hasBase = true
	} else if hasBaseB {
		baseModel = baseModelB
		hasBase = true
	}

	opts := merge.Options{
		Strategy: strat,
		FirstRun: !hasBase,
	}

	// 6. Perform 3-way merge
	mergedModel, changes, conflicts, err := merge.ThreeWayMerge(baseModel, modelA, modelB, opts)
	if err != nil {
		return fmt.Errorf("3-way merge failed: %w", err)
	}

	// 7. Write report
	sources := map[string]string{
		"profileA": usedA,
		"profileB": usedB,
	}
	rep := report.NewReport(dryRun, sources, changes, conflicts)
	if err := rep.WriteHuman(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed writing human report: %v\n", err)
	}

	if reportFile != "" {
		if err := rep.WriteJSON(reportFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed writing JSON report: %v\n", err)
		} else {
			fmt.Printf("Wrote structured JSON report to %s\n", reportFile)
		}
	}

	// 8. Atomic save to both profiles (unless dry-run)
	if !dryRun {
		if err := adapterA.Save(profileA, mergedModel); err != nil {
			return fmt.Errorf("save merged session to profile A (%s) failed: %w", profileA, err)
		}
		if err := adapterB.Save(profileB, mergedModel); err != nil {
			return fmt.Errorf("save merged session to profile B (%s) failed: %w", profileB, err)
		}

		// Save updated baseline snapshot for both profiles
		if err := profile.SaveBaseline(stateDir, idA, mergedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed updating baseline for profile A: %v\n", err)
		}
		if err := profile.SaveBaseline(stateDir, idB, mergedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed updating baseline for profile B: %v\n", err)
		}
		fmt.Println("Successfully saved merged session store to both profiles and updated baselines.")
	} else {
		fmt.Println("Dry run completed. No files were modified.")
	}

	return nil
}
