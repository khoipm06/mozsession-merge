# Getting Started Guide

This guide walks you through compiling, configuring, and running `mozsession-merge`.

---

## 1. Installation & Build

### Requirements
- **Go**: Version 1.21 or newer.
- **OS**: Linux, macOS, or Windows.

### Compilation
Clone the repository and build the binary:
```bash
git clone https://github.com/example/mozsession-merge.git
cd mozsession-merge
go build -o mergetabs ./cmd/mergetabs
```

---

## 2. Command-Line Reference

```bash
./mergetabs [options] <profileA> <profileB>
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `-dry-run` | bool | `false` | Perform merge in memory and print report without modifying files |
| `-conflict-strategy` | string | `"newer"` | Conflict strategy: `newer`, `keep-both`, `manual` |
| `-report-file` | string | `"merge-report.json"` | Path for outputting structured JSON merge report |
| `-state-dir` | string | `~/.config/mozsession-merge` | Directory where baseline snapshots are stored |
| `-force` | bool | `false` | Bypass active process scanning and profile lock checks |

---

## 3. Step-by-Step Usage Workflows

### Step 1: Close Your Browsers
Close both Firefox and Zen Browser instances before running the tool to avoid lock errors or lost state.

### Step 2: Run a Dry Run
Verify the merge plan before writing changes to disk:
```bash
./mergetabs -dry-run ~/.zen/profile1 ~/.zen/profile2
```
Review the terminal output summary and detailed change log.

### Step 3: Perform Live Session Merge
Run the live merge:
```bash
./mergetabs ~/.zen/profile1 ~/.zen/profile2
```
`mergetabs` will:
1. Load session files (`zen-sessions.jsonlz4` or `sessionstore.jsonlz4`).
2. Load baseline snapshot from `~/.config/mozsession-merge/baselines/`.
3. Execute 3-way merge and post-merge reference integrity sweep.
4. Output terminal change summary and write `merge-report.json`.
5. Create `.bak` backup copies of original session files in both profile directories.
6. Atomically update session files in both profiles.
7. Save updated baseline snapshot.

### Step 4: Open Your Browser
Re-open your browser. All merged tabs, spaces, groups, and folders will be restored.
