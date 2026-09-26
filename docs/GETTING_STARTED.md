# Getting Started Guide

This guide walks you through compiling, configuring, and running `mozsession-merge`.

---

## 1. Installation & Build

### Requirements
- **Go**: Version 1.21 or newer.
- **OS**: Linux, macOS, or Windows.

### Option A: go install (recommended)
```bash
go install github.com/khoipm06/mozsession-merge/cmd/mergetabs@latest
```

### Option B: Build from source
```bash
git clone https://github.com/khoipm06/mozsession-merge.git
cd mozsession-merge
go build -o mergetabs ./cmd/mergetabs
```

---

## 2. Command-Line Reference

```bash
mergetabs [options] <profileA> <profileB>
```

| Flag | Shorthand | Default | Description |
|---|---|---|---|
| `--dry-run` | `-d` | `false` | Perform merge in memory and print report without modifying files |
| `--conflict-strategy` | `-c` | `"newer"` | Conflict strategy: `newer`, `keep-both` |
| `--report-file` | `-r` | `"merge-report.json"` | Path for outputting structured JSON merge report |
| `--state-dir` | `-s` | platform config dir | Directory where baseline snapshots are stored |
| `--force` | `-f` | `false` | Bypass active process scanning and profile lock checks |
| `--version` | | | Print version and exit |

---

## 3. Step-by-Step Usage Workflows

### Step 1: Close Your Browsers
Close both Firefox and Zen Browser instances before running the tool to avoid lock errors or lost state.

### Step 2: Run a Dry Run
Verify the merge plan before writing changes to disk:
```bash
mergetabs --dry-run ~/.zen/profile1 ~/.zen/profile2
```
Review the terminal output summary and detailed change log.

### Step 3: Perform Live Session Merge
Run the live merge:
```bash
mergetabs ~/.zen/profile1 ~/.zen/profile2
```
`mergetabs` will:
1. Load session files (`zen-sessions.jsonlz4` or `sessionstore.jsonlz4`).
2. Load baseline snapshot from the platform config directory.
3. Execute 3-way merge and post-merge reference integrity sweep.
4. Output terminal change summary and write `merge-report.json`.
5. Create `.bak` backup copies of original session files in both profile directories.
6. Atomically update session files in both profiles.
7. Save updated baseline snapshot.

### Step 4: Open Your Browser
Re-open your browser. All merged tabs, spaces, groups, and folders will be restored.
