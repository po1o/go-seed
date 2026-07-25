# Agent Instructions

## General File Creation Guidelines

When creating new files:

- **Always use LF (Unix-style) line endings**, not CRLF (Windows-style).
- This repository uses `.gitattributes` to enforce LF line endings.

## Golang

When editing Go files (`*.go`):

- All sources live under `src/`. Run tasks from anywhere via go-task.
- Before committing, run the full suite:
  - `go tool task test`
- This runs `go vet`, `go tool golangci-lint run`, `go tool modernize`,
  `go tool fieldalignment`, and `go test -count=1 ./...`.
- Tool versions are pinned in `src/go.mod` via the `tool` directive — CI runs
  the same binaries. **Never** install a different version with
  `go install ...@latest`; bump via `go get -tool <pkg>@<ver>` (then
  `go mod tidy`) so CI follows.
- Full build/test reference: [`src/BUILD.md`](src/BUILD.md).

## Markdown

When editing Markdown (`*.md`):

- Use proper headings, fenced code blocks with a language, and keep lines within
  the configured limit.
- Lint from the **repo root** (not `src/`): `task markdown-lint`.
- Config: `.markdownlint-cli2.yaml` (line length 120; MD013 disabled for tables;
  MD060 disabled).

## Commit and Pull Request Guidelines

- Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary)
  for PR titles and commit messages.
- Repository rules live in `.commitlintrc.cjs`.
- **Allowed commit types** (anything else fails CI):
  `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `revert`, `style`,
  `test`. Use `ci(deps):` or `chore(deps):` for tooling/dependency bumps.
- Limit body lines to 200 characters; enforced by commitlint.
- **Do not commit initial plans or progress updates as separate commits.**
  Include planning information in the PR description instead.

Examples:

- `feat(build): embed version via -ldflags`
- `ci(deps): bump golangci-lint via go.mod tool directive`
