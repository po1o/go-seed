# Toolchain

How go-seed is wired: layout, pinned tools, [go-task](https://taskfile.dev) as the one entrypoint, and CI that
runs the exact same tasks.

## Design goals

1. **One entrypoint.** `task <name>` for everything — build, test, lint, format.
2. **Zero global installs.** Only prerequisite is Go. Every tool — linters *and* go-task — is pinned in
   `src/go.mod` and run via `go tool`.
3. **CI parity.** CI calls the same tasks you do. `task test` green locally ⇒ CI green.

## Repository layout

```text
.
├── Taskfile.yml                 # single entrypoint — every task
├── README.md                    # quick start
├── TOOLCHAIN.md                 # this file
├── AGENTS.md                    # notes for AI agents / contributors
├── COPYING                      # license
│
├── src/                         # ALL Go sources
│   ├── go.mod                   # module + Go version + pinned `tool` directives
│   ├── go.sum                   # checksums (deps + tools)
│   ├── main.go / main_test.go   # entrypoint
│   ├── build/                   # build-time metadata (Version, set via -ldflags)
│   ├── BUILD.md                 # build & test reference
│   └── .golangci.yml            # golangci-lint config
│
├── .github/
│   ├── dependabot.yml           # daily updates: gomod (/src) + github-actions
│   └── workflows/
│       ├── ci.yml               # test (Linux/macOS) + build → one `code-checks` gate
│       ├── gomod.yml            # `task tidy:check`
│       ├── markdown.yml         # markdownlint-cli2
│       ├── commits.yml          # commitlint
│       ├── dependabot_tidy.yml  # auto `go mod tidy` on Dependabot Go PRs
│       ├── dependabot_automerge.yml  # auto-merge Dependabot PRs once green
│       └── composite/bootstrap-go/   # reusable "install pinned Go" step
│
├── .editorconfig / .gitattributes    # LF endings, tabs for Go
├── .markdownlint-cli2.yaml           # markdown rules
└── .commitlintrc.cjs                 # commit types + line length + Dependabot skip
```

**Why `src/`?** It separates the Go module (code) from the repo (docs, CI, license). Module root is `src/`; module
path is `github.com/po1o/go-seed/src`.

## Pinned tools: the `tool` directive

Go 1.24+ declares CLI dependencies in `go.mod`:

```go
tool (
 github.com/go-task/task/v3/cmd/task
 github.com/golangci/golangci-lint/v2/cmd/golangci-lint
 golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
 golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize
)
```

Versions lock in `go.sum` like any dependency. So everyone runs the identical binary and nothing is installed
globally — `go tool` builds and caches each on first use.

**Run one:**

```bash
go tool golangci-lint run
go tool task test
```

`go vet` ships with Go, so it's `go vet`, not `go tool vet`.

**Bump one** — never `go install @latest`; bump the pin so CI follows:

```bash
cd src
go get -tool <pkg>@vX.Y.Z && go mod tidy   # then commit go.mod / go.sum
```

## go-task

[`Taskfile.yml`](./Taskfile.yml) (repo root) is the single source of truth. go-task searches upward for it, so
tasks run from anywhere.

**Directory resolution.** Go tools must run in `src/`, but the Taskfile is at root. Each Go task sets `dir: src`,
resolved relative to the *Taskfile*, not your cwd — so `task test:unit` always hits `src/`.

**`test` — the pre-push / CI gate:**

```text
test ─► vet ─► lint ─► modernize ─► fieldalignment ─► test:unit
```

Each sub-task also runs alone (`task lint`, `task test:unit`).

**All tasks:**

| Task             | Command it runs                                 | Notes                                   |
| ---------------- | ----------------------------------------------- | --------------------------------------- |
| `default`        | `go tool task --list`                           | Runs when you type `task` alone         |
| `build`          | `go build -o ../bin/go-seed .`                 | Host OS/arch; `OUTPUT=…` overrides      |
| `build:all`      | cross-compiles all 5 targets                    | Artifacts: `bin/go-seed-<os>-<arch>`   |
| `build:<os>-<arch>` | one cross target                             | linux/freebsd amd64+arm64, darwin arm64 |
| `run`            | `go run .`                                       |                                         |
| `clean`          | `git clean -dfX` + `go clean -testcache`        | Removes ignored files; keeps code       |
| `clean:dry`      | `git clean -dnX`                                | Preview only                            |
| `test:unit`      | `go test -count=1 ./...`                        | Unit tests only                         |
| `test:race`      | `go test -race -count=3 ./...`                  | Race detector                           |
| `fmt`            | `go tool golangci-lint fmt`                     | gofmt + goimports                       |
| `vet`            | `go vet ./...`                                  | Built into Go                           |
| `lint`           | `go tool golangci-lint run`                     | Config: `src/.golangci.yml`             |
| `modernize`      | `go list ./... \| xargs go tool modernize`      | Flags outdated Go idioms                |
| `fieldalignment` | `go list ./... \| xargs go tool fieldalignment` | Struct field ordering                   |
| `tidy`           | `go mod tidy`                                   |                                         |
| `tidy:check`     | `go mod tidy` + git-diff gate                   | Fails if `go.mod`/`go.sum` not tidy     |
| `markdown-lint`  | `npx -y markdownlint-cli2@0.23.1 '**/*.md'`     | Node tool (not `go tool`)               |
| `commit-lint`    | pure-npx commitlint + `NODE_PATH`               | Node tool; `SOURCE=…` overrides         |
| `test`           | vet → lint → modernize → fieldalignment → test:unit | The pre-push / CI gate              |
| `setup`          | one-time project bootstrap                      | Runs the two `setup:*` tasks            |
| `setup:branch-protection` | `gh api …/rulesets` (create/update)    | "protected" ruleset; repo auto-detected |
| `setup:merge-settings` | `gh api PATCH … allow_auto_merge/allow_merge_commit` | Auto-merge on, merge commits off  |

### Node-backed tasks: `markdown-lint` and `commit-lint`

The only non-`go tool` checks — no mature Go equivalent, so they use `npx`.

`markdown-lint` is a one-liner (`markdownlint-cli2` is self-contained). `commit-lint` is trickier: its config
`extends: '@commitlint/config-conventional'`, and commitlint resolves that preset **from the current directory**,
not from its binary — so plain npx fails with `Cannot find module`.

The fix stays **pure npx** (nothing in the repo). npx puts its cache `node_modules/.bin` on `PATH`; inside `sh -c`
we recover that `node_modules` and pass it via `NODE_PATH`:

```sh
npx -y --package @commitlint/cli@19 --package @commitlint/config-conventional@19 -- \
  sh -c 'NODE_PATH="$(dirname "$(dirname "$(command -v commitlint)")")" \
         commitlint --config .commitlintrc.cjs --last --verbose'
```

`command -v commitlint` → `<cache>/node_modules/.bin/commitlint`; two `dirname`s → `<cache>/node_modules`. The
working tree never gets a `node_modules`. Override the source:

```bash
task commit-lint SOURCE='--from origin/main --to HEAD'   # a PR range
```

CI uses `wagoid/commitlint-github-action` instead — same `.commitlintrc.cjs`.

**Why `.cjs`, not YAML?** The config sets `ignores: [(message) => …]` to skip Dependabot's own bump commits —
their changelog bodies routinely blow past the 200-char cap and we don't author them. `ignores` must be a
**function**, which YAML cannot express, so the config is CommonJS. The auto-tidy commit this repo pushes is a
normal Conventional Commit and is still linted.

**Bootstrapping go-task:** it's a pinned tool, so run `go tool task <name>`. Want a short `task`?
`brew install go-task/tap/go-task` and drop the prefix — same Taskfile.

## Linters

| Tool             | Catches                                                                     |
| ---------------- | --------------------------------------------------------------------------- |
| `go vet`         | Suspicious constructs the compiler allows (bad printf verbs, lock copies)   |
| `golangci-lint`  | ~25 linters — enabled set in `src/.golangci.yml`                            |
| `modernize`      | Code predating newer Go idioms (`for range n`, `min`/`max`, `slices`)       |
| `fieldalignment` | Struct layouts that waste memory via field ordering                         |

Enabled set, line-length (180), and exclusions live in [`src/.golangci.yml`](./src/.golangci.yml). Formatting
(`gofmt`/`goimports`) → `task fmt`.

## CI

Every workflow installs pinned Go via `bootstrap-go`, then calls a task. No check logic is duplicated.

| Workflow              | Runs                                    | Platforms    |
| --------------------- | --------------------------------------- | ------------ |
| `ci.yml`              | `task test` (matrix) + `task build` | Linux, macOS |
| `gomod.yml`           | `go tool task tidy:check`               | Linux        |
| `markdown.yml`        | markdownlint-cli2 action                | Linux        |
| `commits.yml`         | commitlint action                       | Linux        |
| `dependabot_tidy.yml` | `go mod tidy` on Dependabot PRs         | Linux        |
| `dependabot_automerge.yml` | enable auto-merge on Dependabot PRs | Linux        |

**One job, every OS.** go-task runs `cmds` through an embedded POSIX shell (`mvdan/sh`), so pipes and `xargs` work
identically everywhere — one `task test` covers the matrix. The matrix sets the shell per OS (`bash` on Linux,
`zsh` on macOS); adding an OS is one line.

### `ci.yml`: one gate, no path-filter deadlock

`ci.yml` holds four jobs and is the whole code pipeline:

```text
changes ─► test (matrix: Linux, macOS; runs `task test`) ─┐
        └► build (Linux, uploads artifact) ───────────────────┴─► code-checks
```

- **`changes`** diffs the PR against its base and outputs `code=true/false`. Docs-only changes (any `*.md`
  anywhere, `COPYING`, `.github/FUNDING.yml`) → `false`. If the base can't be resolved it fails **safe** to `true`.
- **`test`** (runs `task test` — lint + static analysis + tests) and **`build`** carry
  `if: needs.changes.outputs.code == 'true'`, so a docs-only PR **skips** them — no wasted compile.
- **`code-checks`** is the **single required status check**. It `needs` the others, runs with `if: always()`,
  and passes when each dependency **succeeded or was skipped**; it fails if the detector errored or any job failed.

**Why not a workflow-level `paths-ignore`?** Because these jobs are *required*. A workflow skipped by `paths-ignore`
never posts its check context, so a required check stays **Pending forever** and the PR is permanently blocked. A
job skipped by an `if:` **does** post — as success. Job-level gating is the only way to be both *required* and
*skippable*. `code-checks` also collapses the `test` matrix legs + `build` into **one** stable context, so
adding an OS never changes what branch protection must require.

**Tidy gate.** `gomod.yml` runs `go mod tidy`, then fails on any diff — no merging an untidy module:

```bash
go mod tidy
git diff --exit-code -- go.mod go.sum
```

### Branch protection & first-time setup

The required checks (`code-checks`, `commit check`, `go-mod`, `markdown check`) only *gate* merges once the branch
is protected.
`task setup` does that for a freshly created project:

```bash
task setup                 # → task setup:branch-protection
```

`setup:branch-protection` reads the repo slug from the checkout's `origin` remote (via `gh repo view`) and targets
`~DEFAULT_BRANCH`, so it is fully generic — every project cloned from this template runs the identical command with
nothing hardcoded. It creates (or updates in place — re-runnable, no duplicates) a **repository ruleset** named
`protected`, GitHub's modern replacement for branch protection.

The policy:

- **A PR is required** — no direct pushes to the default branch.
- **0 required approvals.** A solo repo has no second reviewer, so forcing an approval would just mean
  admin-overriding every merge. Instead you merge your own green PRs, and Dependabot can auto-merge. You still
  review by eye; the CI checks are the real gate.
- **The four checks must pass**, plus **no force-push** and **no branch deletion**.
- **Squash or rebase merges only** — no merge commits, so history stays linear (`task setup:merge-settings` also
  turns off the repo's "Allow merge commits" so the button never appears).
- **Repository admins bypass everything** (`bypass_actors`) — your force-merge escape hatch, cleaner than legacy
  protection's all-or-nothing `enforce_admins`.

Run it once (not-yet-seen check contexts are accepted as "Expected"). Needs the GitHub CLI authenticated with admin
on the repo.

**Why a ruleset, not legacy branch protection?** Rulesets are GitHub's current direction: org-level definitions
across many repos, granular per-role/team/app **bypass actors** (vs. legacy's admin-only toggle), an `evaluate`
dry-run mode, and readability without admin. If you later want a hard review requirement *and* Dependabot
auto-merge, a ruleset is what lets you keep `required_approving_review_count` at 1 while granting a bypass — legacy
protection can't express that.

### Dependabot

Pins are reproducible but rot — stale action pins land you on a dead runner (Node 20 → 24).
[`.github/dependabot.yml`](./.github/dependabot.yml) opens a daily grouped PR per ecosystem (`github-actions`,
`gomod`).

> **First push:** the config does nothing until you enable Dependabot in **Settings → Code security**. No setting →
> zero PRs, no warning. Runs and errors show under **Insights → Dependency graph → Dependabot**. ("Security
> updates" is a *separate* toggle from these version updates.)

**Watch for couplings Dependabot can't see** — e.g. `markdownlint-cli2-action` and `Taskfile.yml`'s
`MDLINT_VERSION` must move together.

### Auto-tidy for Dependabot Go PRs

Dependabot bumps versions but never runs `go mod tidy`, so a bump that shuffles a dep between the direct/indirect
blocks fails the tidy gate. [`dependabot_tidy.yml`](./.github/workflows/dependabot_tidy.yml) runs `go mod tidy` on
the PR branch and pushes the fix. Scoped to `dependabot/go_modules/*`; no-op when clean, so it never loops. The job
gates on `github.event.pull_request.user.login` (the PR author), not `github.actor` — the latter flips to a human
on any re-run/rebase and would wrongly skip the tidy.

**Why a GitHub App.** To push, the job needs an identity. `GITHUB_TOKEN` can't: on Dependabot PRs it's read-only,
and its pushes don't re-trigger checks (loop prevention) — the PR hangs on "waiting for status". A **GitHub App** is
a scoped bot that mints a **~1 h token** per run whose pushes *do* re-trigger checks. Nothing long-lived at rest —
the stored private key only *mints* tokens, it can't push. (A fine-grained PAT works too, but it's standing.)

**Setup — once per repo** (until done, the app-token step fails loudly — that's the reminder):

1. New **GitHub App** (Settings → Developer settings → GitHub Apps): **Contents: Read and write** only, no webhook.
   Note the **Client ID** (format `Iv23li…`), generate a **private key** (`.pem`).
2. **Install** it on the repo.
3. Two **Dependabot** secrets (*not* Actions — only that store reaches Dependabot runs): `TIDY_APP_CLIENT_ID`,
   `TIDY_APP_PRIVATE_KEY`.

**More repos: reuse the one App.** Add the repo under the App's *Install → Configure → select repositories*, then
copy the **same** two secrets. No new App, no new key.

**Security — one rule: never add `test`/`build`/`run`/`generate` here.** The token is safe only because nothing
untrusted runs beside it: the steps are pinned actions + `go mod tidy`, and `go mod tidy` runs no dependency code
(no compile, no `init()`, no install scripts). Worst case if it leaked: push access to this one repo for ~1 h — not
your other secrets (those stay in the withheld Actions store). Tests and builds belong in `ci.yml`, which is
secret-less.

### Auto-merge for Dependabot PRs

With the `protected` ruleset requiring **0 reviews**, a Dependabot PR needs nothing but green checks to merge —
there is no approval to wait on. [`dependabot_automerge.yml`](./.github/workflows/dependabot_automerge.yml) turns on
GitHub's **native auto-merge** for each Dependabot PR (`gh pr merge --auto --squash`), so it merges itself the
moment the checks pass. It uses the plain `GITHUB_TOKEN` (elevated to `pull-requests: write`) — **no App needed**,
because enabling auto-merge isn't a push. It runs only for `dependabot[bot]`'s own PRs, and runs no dependency code.

Two prerequisites, both handled by `task setup`:

1. **"Allow auto-merge" must be on** for the repo — `task setup:merge-settings` sets it
   (`gh api PATCH … allow_auto_merge=true`). It lets a PR *queue* a merge that fires once the required checks
   pass; without it, `gh pr merge --auto` can't arm that queue on a PR whose checks are still pending. (With no
   required checks a PR is mergeable immediately, so it merges on the spot regardless — which is how a Dependabot
   PR slipped through *before* the ruleset existed.) That same task also turns *off* "Allow merge commits" — an
   unrelated setting, see the ruleset section above.
2. **No required review** — the `protected` ruleset already sets `required_approving_review_count: 0`.

The pieces compose: Dependabot opens a PR → `dependabot_tidy.yml` fixes `go.mod` if needed (its App-token push
re-triggers checks) → `dependabot_automerge.yml` has auto-merge armed → checks go green → GitHub merges. Hands-off.
To auto-merge only non-major bumps, add a `dependabot/fetch-metadata` gate (noted in the workflow).

## Typical change

```bash
task test                 # full gate (edit code under src/ first)
cd src && go mod tidy         # only if deps changed
git commit -m "feat(build): embed version via -ldflags"   # Conventional Commit
# open PR — CI runs the same task test
```

Same tools, same tasks, locally and in CI: green `test` ⇒ green pipeline.
