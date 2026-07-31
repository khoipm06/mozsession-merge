# Browser Session 3-Way Merge Tool (`mergetabs`)

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://golang.org)
[![Build & Test](https://img.shields.io/badge/tests-passing-brightgreen)](https://golang.org)

A safe, offline CLI tool written in Go to perform non-destructive **3-way merges** between Mozilla Firefox and Zen Browser session stores (`sessionstore.jsonlz4` and `zen-sessions.jsonlz4`).

---

## Why `mozsession-merge`?

Standard browser sync mechanisms (like **Firefox Sync**) sync URLs for viewing in a passive "Synced Tabs" list. They **do not merge active sessions**, nor do they preserve multi-window layouts, Zen Spaces, workspace folder trees, tab groups, split-view states, or navigation back/forward history across devices or profiles.

`mozsession-merge` operates directly on session store files (`.jsonlz4`) to perform a true **3-way merge** against local baseline snapshots, allowing active tabs and workspace structures from multiple profiles to coexist seamlessly—100% offline, without needing cloud servers or accounts.

Read more in [docs/MOTIVATION_AND_COMPARISON.md](docs/MOTIVATION_AND_COMPARISON.md).

---

## Documentation Roadmap (`docs/`)

- 📘 [Getting Started Guide](docs/GETTING_STARTED.md): Installation, compilation, command-line flags, and step-by-step usage workflows.
- 💡 [Motivation & Firefox Sync Comparison](docs/MOTIVATION_AND_COMPARISON.md): Deep-dive into session fragmentation problems and why this tool differs from cloud sync.
- 🏗️ [Technical Architecture & Data Flow](docs/ARCHITECTURE.md): System design, package layout, `NormalizedModel`, generic 3-way diff engine, and sequence diagrams.
- 🛠️ [Maintainer & Agent Contribution Guide](docs/MAINTAINER_GUIDE.md): Developer guide for adding new browser adapters, extending conflict strategies, and running test suites.

---

## Quick Start

### Build
```bash
go build -o mergetabs ./cmd/mergetabs
```

### Dry Run
```bash
./mergetabs -dry-run ~/.zen/profile1 ~/.zen/profile2
```

### Live Merge
```bash
./mergetabs ~/.zen/profile1 ~/.zen/profile2
```

---

## Key Safety Features

1. **Process Lock Detection**: Scans active process tables (`/proc`) and lock files (`parent.lock`) to prevent race conditions with open browsers.
2. **Pre-Write Backups**: Creates `.bak` copies of target session files before modifying disk state.
3. **Atomic Writes**: Writes payload to `.tmp` files before performing atomic OS renames.
4. **Integrity Sweeping**: Automatically repairs dangling group IDs, workspace UUIDs, and folder parent references post-merge.

---

## Testing

Run unit and integration test suites:
```bash
go test -v ./...
```
