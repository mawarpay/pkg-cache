# Changelog

All notable changes to this repository will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows semantic versioning.

## [Unreleased]

### Added
- `.golangci.yml` — golangci-lint v2 config (staticcheck disabled for Go 1.26 compatibility).
- `Makefile` — local targets for fmt, vet, test (with race + coverage), bench, CI-style `check`, and `lint` / `docker-lint` via `golangci/golangci-lint` Docker image.
- `Dockerfile` — Go 1.26 test runner image.
- `docker-compose.yml` — Redis service plus `test` service that runs `go test ./... -race`.
- `docker-compose.dev.yml` — optional host port publishing for Redis (`:6379`).
- `.dockerignore` and `.gitignore` — keep Docker build context and local artifacts out of version control.

### Changed
- `README.md` — document `make test`, `make check`, and `make docker-*` workflows.
- `trace.go` — align OpenTelemetry `tracerName` with module path `github.com/mawarpay/pkg-cache`.

### Notes
- No public API or business-logic changes.
- Unit tests remain Redis-free; Docker Compose starts Redis for a consistent CI-like test environment.

## [0.2.1] - 2026-08-09

### Added
- `doc.go` — package-level documentation to improve pkg.go.dev presentation.
- `examples_test.go` — ExampleKey demonstrating `Key(...)` usage (adds a small runnable example).
- `README.md` — expanded overview, goals, tech stack, features, quick-start, and run/test instructions.

### Changed
- `config.go` — added GoDoc comments for TTL constants and `TTLConfig.For` method.
- `session.go` — added GoDoc comments for exported session helpers and errors (e.g. `HasSession`, `HasUserSession`, `HasAdminSession`, `ErrRedisRequired`).

### Pending / Prepared (not applied automatically)
- `store.go` — suggested GoDoc comments for the `Store` type and its public methods (`NewStore`, `WithEntity`, `LoadJSON`, `SetJSON`, `GetJSON`, `Del`, `InvalidatePrefix`, `MGetJSON`, `MSetJSON`). A patch/diff was prepared and is available for manual review and application.
- `middleware.go` — suggested GoDoc comment for `RegisterMetrics`. A patch/diff was prepared and is available for manual review and application.
- `.github/workflows/go-ci.yml` — GitHub Actions workflow prepared to run gofmt check, `go vet`, and `go test`. The workflow file content is ready; creation in the repo requires write permission if you want it added via the automation.

### Notes
- No business logic or public API signatures were changed.
- All added comments follow GoDoc conventions: comments start with the identifier name and are concise.
- Examples added are intentionally local and do not depend on Redis or external services.

## How to apply remaining patches locally
If you want to apply the prepared patches (store/middleware docs and CI workflow) locally, use the patches provided earlier or copy the file contents into these paths and commit.

Example commands to add the CI workflow locally:

```bash
mkdir -p .github/workflows
cat > .github/workflows/go-ci.yml <<'YAML'
name: Go CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Set up Go
        uses: actions/setup-go@v7
        with:
          go-version: '1.26'

      - name: Display go version
        run: go version

      - name: Download modules
        run: go mod download

      - name: Check formatting (gofmt)
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "The following files need gofmt -w:"; gofmt -l .; exit 1
          fi

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test -v ./...
YAML

git add .github/workflows/go-ci.yml
git commit -m "ci: add Go CI workflow (gofmt, vet, test)"
git push
```

## Tests & verification
After applying any remaining patches, run the repository checks locally:

```bash
gofmt -w .
go vet ./...
go test ./...
```

## Tagging guidance
- Update the changelog by moving the Unreleased changes under a versioned heading (e.g., `## [0.1.0] - 2026-08-09`) before creating a git tag.
- Create a signed tag and push it:

```bash
git tag -a v0.1.0 -m "v0.1.0: initial docs and CI"
git push origin v0.1.0
```

If you want, I can:
- apply the prepared store/middleware doc edits and create a branch + PR (requires repo push permission), or
- generate a single combined patch file containing all remaining changes (store docs, middleware docs, workflow) for you to apply locally. 
