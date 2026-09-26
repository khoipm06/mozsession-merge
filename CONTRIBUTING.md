# Contributing to mozsession-merge

Thank you for your interest in contributing.

## Reporting Bugs

Open an issue on [GitHub](https://github.com/khoipm06/mozsession-merge/issues) with:
- Your OS and Go version (`go version`)
- The browser type and approximate number of tabs/spaces
- The exact command you ran (redact profile paths if needed)
- The full terminal output or error message

For security-sensitive issues (e.g., the tool reads/writes live browser session data), see [SECURITY.md](SECURITY.md) instead.

## Development Setup

```bash
git clone https://github.com/khoipm06/mozsession-merge.git
cd mozsession-merge
go mod download
go test ./...
```

No external tools or code generators are required beyond a standard Go toolchain (≥ 1.21).

## Running Tests

```bash
# All tests with race detector
go test -v -race ./...

# A specific package
go test -v ./internal/merge/...
```

The test suite uses only synthetic fixtures in `testdata/` — no real browser profiles are needed.

## Submitting Pull Requests

1. Fork the repository and create a feature branch from `main`.
2. Make your changes. Add or update tests to cover new behaviour.
3. Ensure `go test -race ./...` and `go vet ./...` both pass.
4. Follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:
   - `feat(scope): add support for X`
   - `fix(scope): correct Y`
   - `test(scope): add tests for Z`
   - `docs: update getting started guide`
5. Open a pull request against `main`.

## Adding a New Browser Adapter

See [docs/MAINTAINER_GUIDE.md](docs/MAINTAINER_GUIDE.md) for step-by-step instructions on implementing the `adapter.Adapter` interface for a new browser.

## Code Style

- Standard `gofmt` formatting (enforced by CI).
- Error messages are lowercase, wrapped with `%w`.
- No global state outside `cmd/`.
- Prefer the Go standard library; new external dependencies require discussion in an issue first.
