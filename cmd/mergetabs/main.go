package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/khoipm06/mozsession-merge/pkg/mergetabs"
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
			return mergetabs.Run(args[0], args[1], mergetabs.Options{
				DryRun:           dryRun,
				ConflictStrategy: conflictStrategy,
				ReportFile:       reportFile,
				StateDir:         stateDir,
				Force:            force,
			})
		},
	}

	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a trial merge without modifying profile session files")
	rootCmd.Flags().StringVarP(&conflictStrategy, "conflict-strategy", "c", "newer", "Conflict resolution strategy (newer | keep-both)")
	rootCmd.Flags().StringVarP(&reportFile, "report-file", "r", "merge-report.json", "Output path for structured JSON change report")
	rootCmd.Flags().StringVarP(&stateDir, "state-dir", "s", mergetabs.DefaultStateDir(), "Directory where baseline snapshots and tool state are stored")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Bypass active profile lock checks")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
