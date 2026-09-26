# Security Policy

## Scope

`mozsession-merge` reads and writes live browser session files (`sessionstore.jsonlz4`, `zen-sessions.jsonlz4`). These files contain browsing history, tab URLs, workspace metadata, and navigation state.

Security-relevant concerns include:
- Bugs that could corrupt or destroy session data without a backup
- Bugs that could expose session data to unintended file system locations
- Path traversal or unsafe file operations in profile handling

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Email: `khoipm06@gmail.com`

Please include:
- A description of the vulnerability and its potential impact
- Steps to reproduce
- Any affected versions

I'll acknowledge receipt within 48 hours and aim to release a fix within 14 days for confirmed issues.
