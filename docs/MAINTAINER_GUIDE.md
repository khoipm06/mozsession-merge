# Contributor Guide

This guide is for developers extending or maintaining `mozsession-merge`.

---

## 1. Ground Rules & Principles

1. **Minimal Dependencies**: Prefer the Go standard library wherever practical. Current external dependencies: `github.com/pierrec/lz4/v4` (mozLz4 codec) and `github.com/spf13/cobra` (CLI).
2. **Never Touch Live Browser Profiles in Tests**: Unit and integration tests must strictly use temporary directories (`t.TempDir()`) or programmatic mock session models.
3. **Never Guess Schemas or Binary Offsets**: Always verify header structure or JSON payloads empirically before defining raw structs.
4. **Conventional Commits**: All commits must follow Conventional Commits formatting (e.g., `feat(scope): ...`, `test(scope): ...`, `fix(scope): ...`).

---

## 2. Adding a New Browser Adapter

To add support for a new Firefox fork or custom browser (e.g. Floorp, LibreWolf, Waterfox):

1. Create a new package under `internal/adapter/<browsername>/`.
2. Define raw session structs in `schema.go`.
3. Implement `adapter.Adapter` interface in `<browsername>.go`:
   ```go
   type Adapter interface {
       Name() string
       Detect(profileDir string) (bool, error)
       Load(profileDir string) (model.NormalizedModel, error)
       Save(profileDir string, m model.NormalizedModel) error
   }
   ```
4. Register the new adapter in `internal/adapter/adapter.go` dispatcher `DetectAdapter`.
5. Add unit tests in `<browsername>_test.go` verifying `LoadBytes` $\rightarrow$ `SaveBytes` round-trip fidelity.

---

## 3. Running & Extending Tests

### Running Unit Tests
```bash
go test -v ./internal/...
```

### Running Integration Tests
```bash
go test -v ./test/...
```

### Writing Tests for New Features
- **Codecs & Serialization**: Must include round-trip equality assertions (`bytes.Equal(original, reDecoded)`).
- **Merge Engine Changes**: Add hand-built fixture test cases in `internal/merge/merge_test.go`.
