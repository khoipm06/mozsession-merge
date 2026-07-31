# Motivation & Comparison with Firefox Sync

This document explains why `mozsession-merge` was created, its core purpose, and how it differs fundamentally from Mozilla Firefox's built-in Sync service.

---

## 1. Problem Statement & Motivation

Modern web browsers like **Firefox** and **Zen Browser** store active tab sessions, workspaces, groups, folders, and split-view states in compressed session files (`sessionstore.jsonlz4` and `zen-sessions.jsonlz4`).

Users who work across multiple machines (e.g., desktop and laptop) or manage separate browser profiles often face the problem of session fragmentation:
- Profile A on Desktop has 15 open tabs organized into Zen Workspaces and tab groups.
- Profile B on Laptop has 10 open tabs with new research tabs and adjusted workspace folders.
- Overwriting Profile A's session file with Profile B's file destroys Profile A's tabs.
- Manually copying URLs across machines is tedious, loses tab history, and strips workspace/group/folder metadata.

Existing browser tools provide no mechanism to perform an **offline, safe, 3-way merge** of active browser session files. `mozsession-merge` addresses this gap.

---

## 2. Fundamental Comparison: `mozsession-merge` vs. Firefox Sync

| Feature | Firefox Sync | `mozsession-merge` |
|---|---|---|
| **Execution Model** | Cloud service / WebSockets | Pure offline local CLI tool |
| **Merged Target** | "Synced Tabs" sidebar / menu | Direct 3-way merge of local session files (`.jsonlz4`) |
| **Active Session State** | Does **NOT** merge active tab sets or restore window state | **Merges active tab sets**, window layouts, and workspace structures |
| **Zen Browser Features** | Ignored (Firefox Sync only knows standard tabs/history) | Full support for **Zen Spaces, Folders, Groups, and Split-Views** |
| **Conflict Handling** | Last-write-wins at cloud server; no conflict reports | Configurable strategies (`newer`, `keep-both`, `manual`) with structured JSON/terminal reports |
| **Baseline Snapshotting** | None | Local 3-way baseline tracking (detects additions, updates, and deletions) |
| **Safety & Atomicity** | Network sync dependent | Process lock detection, pre-write `.bak` copies, atomic file rename |
| **Privacy & Account** | Requires Firefox Account & Internet | **100% offline**, zero network requests, zero accounts |

### Key Insight: Why Synced Tabs ≠ Session Merging
Firefox Sync syncs tab URLs to Mozilla's cloud servers so you can view a list under "Synced Tabs from Other Devices". However, Firefox Sync **never merges active tab sessions**. Opening a tab from "Synced Tabs" opens a single new tab—it does not reconstruct your tab groups, split views, workspace folder trees, navigation back/forward history, or window geometries.

`mozsession-merge` operates directly on the underlying session store files to produce a unified session state where tabs, groups, folders, and spaces from both profiles coexist seamlessly.

---

## 3. Core Purpose & Use Cases

1. **Multi-Device Session Consolidation**: Merge your laptop's session into your desktop profile after traveling.
2. **Profile Migration & Hybrid Use**: Seamlessly merge a Stock Firefox profile into Zen Browser (or vice-versa).
3. **Automated Sync Workflows**: Wire `mozsession-merge` into offline file-sync utilities (like Syncthing or rsync) for automated session synchronization between devices when browsers are closed.
