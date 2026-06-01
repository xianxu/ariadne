---
id: 000057
status: working
deps: []
github_issue:
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 4
---

# dev-aliases.sh — expose Go binaries fresh; retire cmd source symlinks

## Problem

Two related pains, surfaced shipping #56:

1. **Dev staleness.** Developing a Go tool (sdlc) meant rebuilding the binary
   after every edit — and *two* binaries (`ariadne/bin/sdlc` + the on-PATH
   `pair/bin/sdlc`). `go run`/`go tool` don't fix it uniformly: in a *derivative*
   the `tool`+`replace` live in `construct/go.mod` (a separate module from the
   app `go.mod`, per #37/#38), so `go tool X` only resolves from `construct/`
   — whose cwd is wrong for the tool to operate on the repo. Only the *owner*
   repo (tool in root go.mod) can `go run`/`go tool` directly.
2. **Code distributed by source-symlink (deprecated).** `nous` ships its
   binaries' *source* to derivatives via `symlink cmd/gmail`/`symlink cmd/oneshot`
   manifest directives → 6 symlinks across `brain`, `brain-family`,
   `brain-private`. That's code flowing through the *substrate* (file-symlink)
   channel instead of the module channel — the same pattern already retired for
   sdlc (ariadne→pair/brain use `tool`+`replace`, no source copy).

## Spec

**Principle: ownership = location.** If a binary's source is physically in a
repo, that repo owns + builds it. Derivatives never hold the source (no copies,
no symlinks); they *run* the binary or *compile-in-owner*. Two orthogonal
mechanisms, not to be conflated:

- **Expose** (this issue) = location-based. Every owned `cmd/X` is a dev-fresh
  shell function. Independent of who consumes it.
- **Consume** = a derivative's `construct/go.mod` `tool`+`replace` (build wiring).
  Unchanged by this issue.

Scope note: this is for **code** (Go). Markdown/config substrate (AGENTS.md,
skills) has no module mechanism and keeps distributing by symlink/copy — only
*code* leaves the substrate channel.

### `construct/dev-aliases.sh` (in ariadne)

Emits shell function definitions to **stdout** (for `source <(dev-aliases.sh)`),
warnings to **stderr**. For each owned binary X in owner repo R:

```
X() { ( cd R && go build -o "${TMPDIR:-/tmp}/X-dev" ./cmd/X ) || return; "${TMPDIR:-/tmp}/X-dev" "$@"; }
```

The **build-in-owner / run-in-caller's-cwd** form is the only one that works for
both repo-bound tools (sdlc — operates on whatever repo you're standing in) and
run-anywhere tools (nous). `go run`/`go tool` can't (cwd is pinned to the module
dir).

Discovery:
1. Walk same-level siblings of ariadne (incl. ariadne itself); a repo is in
   scope iff it has `construct/`. Skip inactive clones (`*legacy*`, numbered
   like `brain1`/`brain2`).
2. For each in-scope R, expose each `R/cmd/X` that is: a **real dir (not a
   symlink)**, has **≥1 `.go` file** (skips empty/stray dirs like pair's
   `cmd/sdlc/`), and has **no `cmd/X/.private`** marker (git-tracked opt-out).
3. Dedup: **last-win** with a *deterministic* (sorted) walk order, and **warn**
   (stderr) on every duplicate name. `--strict` exits non-zero on any dup (for
   CI / pre-commit). `--list` dry-run prints `binary → owner`.

`.zshrc` shrinks to one line: `source <(~/workspace/ariadne/construct/dev-aliases.sh)`.

The two filters (skip symlinks, require real `.go`) make "ownership = location"
operational and remove most dup collisions for free — re-exported symlinks
(brain's gmail) and empty dirs (pair's cmd/sdlc) drop out, leaving only genuine
same-name conflicts for `--strict`/warn.

### Retire the nous→brain source-symlinks

- Delete the 2 `symlink cmd/{gmail,oneshot}` directives from
  `nous/construct/base.manifest` (the generator — else `make refresh` re-creates
  the symlinks).
- Remove the 6 symlinks (`brain`, `brain-family`, `brain-private` × gmail,
  oneshot) — or let a post-directive `make refresh` clean them.
- Verify those repos obtain gmail/oneshot via the dev-alias (build-in-nous)
  and/or a PATH binary; redirect any build step that compiled the symlinked
  source.

### Non-goals

- Don't change consume-wiring (`construct/go.mod` `tool`+`replace` stays).
- Don't touch substrate symlinks (`symlink AGENTS.md` etc. — docs, not code).
- Legacy clones (`brain.legacy`, `brain1`, `brain2`) hold *real* duplicate
  source — deleting those is separate cruft cleanup, out of scope here.

## Done when

- `source <(construct/dev-aliases.sh)` defines a fresh-build function per owned,
  non-`.private`, buildable `cmd/X` across active ariadne-styled siblings;
  invoking one builds in the owner and runs in the caller's cwd.
- Symlinks excluded, empty/stray `cmd/X` excluded, `.private` honored.
- Dup names: last-win (sorted order) + stderr warn; `--strict` exits non-zero;
  `--list` dry-run works.
- nous's `symlink cmd/X` directives gone; the 6 brain* symlinks gone; those
  repos still get gmail/oneshot (verified).
- pair's stray empty `cmd/sdlc/` deleted.
- atlas updated (base-layer.md / a dev-tooling note); `.zshrc` one-liner documented.

## Plan

Two milestones, each a fresh-eyes `sdlc milestone-close` review boundary.

- [ ] M1 — `construct/dev-aliases.sh` (the generator)
- [ ] M2 — retire nous source-symlinks + cleanup + docs

### M1 — generator
- [ ] `construct/dev-aliases.sh`: sibling walk + `construct/` scope + active
      filter; per-repo `cmd/X` filters (real dir, ≥1 `.go`, no `.private`);
      emit the build-in-owner/run-in-cwd function to stdout.
- [ ] Dedup last-win (sorted walk) + stderr warn; `--strict` (exit≠0 on dup) +
      `--list` dry-run.
- [ ] Tests against a fixture sibling tree (a tool, a `.private` tool, an
      empty `cmd/X`, a symlinked `cmd/X`, a name collision) — assert emitted
      functions + warn/strict behavior. Ship a `SKILL.md` for the script.

### M2 — retire symlinks + cleanup + docs
- [ ] Remove `symlink cmd/{gmail,oneshot}` from `nous/construct/base.manifest`;
      remove the 6 symlinks across brain/brain-family/brain-private; verify
      those repos still obtain gmail/oneshot (dev-alias / PATH binary), redirect
      any build step off the symlinked source.
- [ ] Delete pair's stray empty `cmd/sdlc/`.
- [ ] atlas (base-layer.md / dev-tooling note) + the `.zshrc` `source <(...)`
      one-liner; verify end-to-end (`source` it, run sdlc + a nous binary from
      different cwds).

## Log

### 2026-06-01
Created from the dev-mode / ownership discussion that came out of #56's ship.
Key decisions captured in Spec: ownership=location (not a `tool` manifest
declaration); expose (location-based, all owned `cmd/X`) is orthogonal to
consume (`construct/go.mod` wiring); the build-in-owner/run-in-cwd **function**
form is the only mechanism that serves both repo-bound (sdlc) and run-anywhere
(nous) tools — `go run`/`go tool` can't, because in derivatives the tool lives
in the separate `construct/` module whose cwd is wrong. Symlink + real-`.go`
filters operationalize the rule and kill most dup collisions; last-win+warn with
`--strict` handles the rest.
