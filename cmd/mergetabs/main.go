package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"mozsession-merge/internal/adapter"
	"mozsession-merge/internal/adapter/firefox"
	"mozsession-merge/internal/adapter/zen"
	"mozsession-merge/internal/backupfallback"
	"mozsession-merge/internal/merge"
	"mozsession-merge/internal/model"
	"mozsession-merge/internal/profile"
	"mozsession-merge/internal/report"
)

func defaultStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".mozsession-merge"
	}
	return filepath.Join(home, ".config", "mozsession-merge")
}

func parseStrategy(s string) (merge.Strategy, error) {
	switch s {
	case "newer", "newer-wins":
		return merge.StrategyNewerWins, nil
	case "keep-both":
		return merge.StrategyKeepBoth, nil
	case "manual":
		return merge.StrategyManual, nil
	default:
		return 0, fmt.Errorf("invalid conflict strategy %q (allowed: newer, keep-both, manual)", s)
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

func main() {
	var (
		dryRun           bool
		conflictStrategy string
		reportFile       string
		stateDir         string
		force            bool
	)

	flag.BoolVar(&dryRun, "dry-run", false, "Perform a trial merge without modifying profile session files")
	flag.StringVar(&conflictStrategy, "conflict-strategy", "newer", "Conflict resolution strategy (newer | keep-both | manual)")
	flag.StringVar(&reportFile, "report-file", "merge-report.json", "Output path for structured JSON change report")
	flag.StringVar(&stateDir, "state-dir", defaultStateDir(), "Directory where baseline snapshots and tool state are stored")
	flag.BoolVar(&force, "force", false, "Bypass active profile lock checks")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <profileA> <profileB>\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	if err := runMergeTabs(args[0], args[1], dryRun, conflictStrategy, reportFile, stateDir, force); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMergeTabs(profileA, profileB string, dryRun bool, conflictStrategy, reportFile, stateDir string, force bool) error {
	strat, err := parseStrategy(conflictStrategy)
	if err != nil {
		return err
	}

	if !force {
		lockedA, reasonA, errA := profile.IsLocked(profileA)
		if errA == nil && lockedA {
			return fmt.Errorf("profile A (%s) is locked: %s. Close browser or use -force", profileA, reasonA)
		}

		lockedB, reasonB, errB := profile.IsLocked(profileB)
		if errB == nil && lockedB {
			return fmt.Errorf("profile B (%s) is locked: %s. Close browser or use -force", profileB, reasonB)
		}
	}

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

	primaryA := adapter.SessionStorePath(profileA)
	if kindA == "zen" {
		primaryA = adapter.ZenSessionsPath(profileA)
	}
	fallbacksA := backupfallback.DefaultFallbackCandidates(profileA, kindA)
	rawA, _, usedA, err := backupfallback.LoadWithFallback(primaryA, fallbacksA)
	if err != nil {
		return fmt.Errorf("failed loading session for profile A (%s): %w", profileA, err)
	}

	primaryB := adapter.SessionStorePath(profileB)
	if kindB == "zen" {
		primaryB = adapter.ZenSessionsPath(profileB)
	}
	fallbacksB := backupfallback.DefaultFallbackCandidates(profileB, kindB)
	rawB, _, usedB, err := backupfallback.LoadWithFallback(primaryB, fallbacksB)
	if err != nil {
		return fmt.Errorf("failed loading session for profile B (%s): %w", profileB, err)
	}

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

	idA := profileID(profileA)
	baseModel, hasBase, err := profile.LoadBaseline(stateDir, idA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed loading baseline for profile A: %v\n", err)
	}

	opts := merge.Options{
		Strategy: strat,
		DryRun:   dryRun,
		FirstRun: !hasBase,
	}

	mergedModel, changes, conflicts, err := merge.ThreeWayMerge(baseModel, modelA, modelB, opts)
	if err != nil {
		return fmt.Errorf("3-way merge failed: %w", err)
	}

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

	if !dryRun {
		if err := adapterA.Save(profileA, mergedModel); err != nil {
			return fmt.Errorf("save merged session to profile A (%s) failed: %w", profileA, err)
		}
		if err := adapterB.Save(profileB, mergedModel); err != nil {
			return fmt.Errorf("save merged session to profile B (%s) failed: %w", profileB, err)
		}

		if err := profile.SaveBaseline(stateDir, idA, mergedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed updating baseline snapshot for profile A: %v\n", err)
		}
		idB := profileID(profileB)
		if err := profile.SaveBaseline(stateDir, idB, mergedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed updating baseline snapshot for profile B: %v\n", err)
		}
		fmt.Println("Successfully saved merged session store to both profiles and updated baselines.")
	} else {
		fmt.Println("Dry run completed. No files were modified.")
	}

	return nil
}
