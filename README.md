# Browser Session 3-Way Merge Tool (`mergetabs`)

[![CI](https://github.com/khoipm06/mozsession-merge/actions/workflows/ci.yml/badge.svg)](https://github.com/khoipm06/mozsession-merge/actions/workflows/ci.yml)
[![Go 1.21+](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A safe, offline CLI tool written in Go to perform non-destructive **3-way merges** between Mozilla Firefox and Zen Browser session stores (`sessionstore.jsonlz4` and `zen-sessions.jsonlz4`).

---

## Why `mozsession-merge`?

Standard browser sync mechanisms (like **Firefox Sync**) sync URLs for viewing in a passive "Synced Tabs" list. They **do not merge active sessions**, nor do they preserve multi-window layouts, Zen Spaces, workspace folder trees, tab groups, split-view states, or navigation back/forward history across devices or profiles.

`mozsession-merge` operates directly on session store files (`.jsonlz4`) to perform a true **3-way merge** against local baseline snapshots, allowing active tabs and workspace structures from multiple profiles to coexist seamlessly—100% offline, without needing cloud servers or accounts.

Read more in [docs/MOTIVATION_AND_COMPARISON.md](docs/MOTIVATION_AND_COMPARISON.md).

---

## Documentation

- 📘 [Getting Started Guide](docs/GETTING_STARTED.md): Installation, command-line flags, and step-by-step usage workflows.
- 💡 [Motivation & Firefox Sync Comparison](docs/MOTIVATION_AND_COMPARISON.md): Deep-dive into session fragmentation problems and why this tool differs from cloud sync.
- 🏗️ [Technical Architecture & Data Flow](docs/ARCHITECTURE.md): System design, package layout, `NormalizedModel`, generic 3-way diff engine.
- 🛠️ [Contributor Guide](docs/MAINTAINER_GUIDE.md): Adding new browser adapters, extending conflict strategies, running tests.

---

## Quick Start

### Install
```bash
go install github.com/khoipm06/mozsession-merge/cmd/mergetabs@latest
```

Or build from source:
```bash
git clone https://github.com/khoipm06/mozsession-merge.git
cd mozsession-merge
go build -o mergetabs ./cmd/mergetabs
```

### Dry Run
```bash
mergetabs --dry-run ~/.zen/<profile-dir> ~/.zen/<profile-dir>
```

### Live Merge
```bash
mergetabs ~/.zen/<profile-dir> ~/.zen/<profile-dir>
```

### Finding Browser Profile Paths

| Browser | OS | Default Profile Location |
|---|---|---|
| **Zen Browser** | Linux | `~/.zen/<profile_dir>/`<br>*(Flatpak: `~/.var/app/io.github.zen_browser.zen/.zen/<profile_dir>/`)* |
| | macOS | `~/Library/Application Support/zen/Profiles/<profile_dir>/` |
| | Windows | `%APPDATA%\zen\Profiles\<profile_dir>\` |
| **Firefox** | Linux | `~/.mozilla/firefox/<profile_dir>/`<br>*(Flatpak: `~/.var/app/org.mozilla.firefox/.mozilla/firefox/<profile_dir>/`)* |
| | macOS | `~/Library/Application Support/Firefox/Profiles/<profile_dir>/` |
| | Windows | `%APPDATA%\Mozilla\Firefox\Profiles\<profile_dir>\` |

> [!TIP]
> You can find your exact active profile folder by navigating to `about:profiles` in Firefox or Zen and looking at the **Root Directory** field.

---

## Key Safety Features

1. **Process Lock Detection**: Scans `/proc` and lock files (`parent.lock`) to prevent race conditions with open browsers.
2. **Pre-Write Backups**: Creates `.bak` copies of target session files before modifying disk state.
3. **Atomic Writes**: Writes payload to `.tmp` files before performing atomic OS renames, ensuring no partial writes.
4. **Integrity Sweeping**: Automatically repairs dangling group IDs, workspace UUIDs, and folder parent references post-merge.

---

## Limitations

- **First-run deletions**: On the very first run (no baseline snapshot), the tool performs a union merge — tabs deleted from one profile since the last sync cannot be detected.
- **Firefox tab identity**: Firefox tabs without a `zenSyncId` are identified by a hash of their navigation history. Tabs with no history entries may hash to the same identity.
- **Closed tabs**: Remote closed-tab history is not merged; local closed tabs are preserved as-is.

---

## Testing

```bash
go test -v -race ./...
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for bug reports, pull requests, and development setup.

## License

[MIT](LICENSE) © Khoi Pham Minh
