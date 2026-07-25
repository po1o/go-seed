# go-seed

A minimal, batteries-included **scaffolding seed for new Go projects**.

`go-seed` starts you from a known-good, lint-clean baseline: sources under
`src/`, every dev tool pinned and invoked through `go tool`, and a single
[go-task](https://taskfile.dev) entrypoint that drives builds, tests, and the
full lint suite — locally and in CI, byte-for-byte identical.

## Layout

```text
.
├── Taskfile.yml            # the one entrypoint — `task <name>`
├── src/                    # all Go sources live here
│   ├── go.mod              # module + pinned `tool` directives
│   ├── main.go
│   └── build/              # build-time metadata (version)
├── .golangci.yml           # golangci-lint config (in src/)
└── .github/workflows/      # CI mirrors the local tasks
```

## Prerequisites

- **Go 1.26+** (the toolchain version is pinned in `src/go.mod`).

That's it. Every other tool — `golangci-lint`, `modernize`, `fieldalignment`,
and `go-task` itself — is pinned via the `tool` directive in `src/go.mod` and
run through `go tool`. There is nothing to install globally.

## Quick start

Everything is triggered through go-task. If you don't have the `task` binary on
your `PATH`, invoke it through Go — it is pinned like every other tool:

```bash
go tool task            # list all tasks
go tool task test       # run the full check suite
go tool task build      # build ./bin/go-seed
```

If you prefer a global `task` binary (`brew install go-task/tap/go-task`), just
drop the `go tool` prefix: `task test`.

## Tasks

| Task                  | What it does                                             |
| --------------------- | ------------------------------------------------------- |
| `build`               | Build for the host OS/arch → `./bin/go-seed`           |
| `build:all`           | Cross-compile all 5 release targets into `./bin`        |
| `build:<os>-<arch>`   | Cross-compile one target (see below)                   |
| `run`                 | `go run .`                                               |
| `clean`               | Remove ignored artifacts (bin/, caches) — keeps code    |
| `clean:dry`           | Preview what `clean` would remove                       |
| `test:unit`           | `go test -count=1 ./...`                                 |
| `test:race`           | Tests under the race detector                           |
| `fmt`                 | Format code (`golangci-lint fmt`)                       |
| `vet`                 | `go vet ./...`                                           |
| `lint`                | `golangci-lint run` (via `go tool`)                     |
| `modernize`           | `modernize` analyzer (via `go tool`)                    |
| `fieldalignment`      | `fieldalignment` analyzer (via `go tool`)               |
| `tidy`                | `go mod tidy`                                            |
| `tidy:check`          | `go mod tidy` + fail if `go.mod`/`go.sum` changed (CI)  |
| `markdown-lint`       | Lint Markdown with markdownlint-cli2                    |
| `commit-lint`         | Lint commit messages (Conventional Commits)            |
| **`test`**            | **vet + lint + modernize + fieldalignment + unit tests** |

Run `task test` before pushing — it is exactly what CI enforces.

## Cross-compiling

`task build` produces a normal host binary at `./bin/go-seed`. To cross-compile,
use `task build:all` (all targets) or a single `task build:<os>-<arch>`. Each
artifact is stamped with its OS/arch so they coexist in `./bin`:

| Target task           | Artifact                       |
| --------------------- | ------------------------------ |
| `build:linux-amd64`   | `bin/go-seed-linux-amd64`     |
| `build:linux-arm64`   | `bin/go-seed-linux-arm64`     |
| `build:darwin-arm64`  | `bin/go-seed-darwin-arm64`    |
| `build:freebsd-amd64` | `bin/go-seed-freebsd-amd64`   |
| `build:freebsd-arm64` | `bin/go-seed-freebsd-arm64`   |

Builds are pure Go (`CGO_ENABLED=0`), so cross-compilation needs no C toolchain.

## Bumping a pinned tool

Tool versions live in `src/go.mod`. Never `go install ...@latest`; bump the pin
so CI follows:

```bash
cd src
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@vX.Y.Z
go mod tidy
```

## License

See [COPYING](./COPYING).
