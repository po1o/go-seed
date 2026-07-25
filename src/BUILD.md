---
name: Build & test guide
description: How to build, test, and lint go-seed locally. The same commands CI runs.
---

# Build & test

Everything is driven by [go-task](https://taskfile.dev). Tool versions are
pinned via the `tool` directive in `src/go.mod` and invoked through `go tool`,
so local and CI runs are byte-identical. Go's version is pinned in `go.mod`.

Invoke tasks with a global `task` binary, or through Go (no install needed):

```bash
go tool task <name>     # or: task <name>
```

## Common tasks

```bash
go tool task build       # build ./bin/go-seed
go tool task test:unit   # go test -count=1 ./...
go tool task test:race   # go test -race -count=3 ./...
go tool task test        # the full CI check suite (see below)
```

## Full pre-push check

```bash
go tool task test
```

`test` runs, in order:

- `go vet ./...`
- `go tool golangci-lint run`
- `go list ./... | xargs go tool modernize`
- `go list ./... | xargs go tool fieldalignment`
- `go test -count=1 ./...`

If that passes locally, CI passes — `.github/workflows/ci.yml` runs the same
`go tool task test` on Linux and macOS.

## Bumping a pinned tool

```bash
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@vX.Y.Z
go mod tidy
```

## Markdown lint

From the repo root (not `src/`):

```bash
task markdown-lint       # npx -y markdownlint-cli2@0.20.0 '**/*.md'
```

Rules live in `.markdownlint-cli2.yaml` at the repo root. The cli2 version is
pinned both there (in a comment) and in `.github/workflows/markdown.yml` — bump
them together.
