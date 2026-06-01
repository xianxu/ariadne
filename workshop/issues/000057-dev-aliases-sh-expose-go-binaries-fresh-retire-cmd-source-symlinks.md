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
2. **Code distributed by source-symlink (deprecated).** `nous` ships its Go
   *source* to derivatives via **3** manifest directives — `symlink lib/gmail`,
   `symlink cmd/gmail`, `symlink cmd/oneshot` — producing **9 symlinks** across
   `brain`, `brain-family`, `brain-private` (3 repos × 3). That's code flowing
   through the *substrate* (file-symlink) channel instead of the module channel
   — the same pattern already retired for sdlc (ariadne→pair/brain use
   `tool`+`replace`, no source copy). (`lib/gmail` is the shared library the
   gmail cmd imports; the derivatives only carry it because they used to compile
   gmail *locally* from the `cmd/gmail` symlink — once gmail builds in its owner
   (nous) via the dev-alias, the derivatives need neither the cmd nor the lib
   source.) Note: the *many* other brain* symlinks (Makefiles, `scripts/`,
   `.claude/skills`, `atlas/workflow`, AGENTS.md, …) are doc/config/script
   substrate with no module mechanism — they stay; only the Go source leaves.

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
X() { ( cd R && mkdir -p bin && rm -f bin/X && go build -o bin/X ./cmd/X ) || return; R/bin/X "$@"; }
```

The **build-in-owner / run-in-caller's-cwd** form is the only one that works for
both repo-bound tools (sdlc — operates on whatever repo you're standing in) and
run-anywhere tools (nous). `go run`/`go tool` can't (cwd is pinned to the module
dir). The binary lands at the **owner's official `bin/X`** (gitignored — not a
temp dir, so it's safe for a service binary like nous; same artifact `make
<name>-dev` produces). `rm -f bin/X` mirrors the owner Makefiles' code-signing
inode safety. The function only builds + runs — it does **not** manage services
(no `launchctl bootout`); use `make <name>-dev` for stop-prod-then-serve.

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

- Delete the **3** code-symlink directives from `nous/construct/base.manifest`:
  `symlink lib/gmail` (line 49), `symlink cmd/gmail` (50), `symlink cmd/oneshot`
  (54) — the generator; else `make refresh` re-creates the symlinks.
- Remove the **9** symlinks (`brain`, `brain-family`, `brain-private` ×
  {`lib/gmail`, `cmd/gmail`, `cmd/oneshot`}) — or let a post-directive
  `make refresh` clean them.
- Verify those repos obtain gmail/oneshot via the dev-alias (build-in-nous)
  and/or a PATH binary; redirect any build step that compiled the symlinked
  source. **Check whether anything in a derivative imports `nous/lib/gmail` as
  a library** (beyond the gmail cmd, which now builds in nous): if so, that's a
  `require nous` + `replace` module dependency, not a symlink — migrate it to
  the module channel rather than leaving the symlink.

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
- nous's 3 code-symlink directives (`lib/gmail`, `cmd/gmail`, `cmd/oneshot`)
  gone; the 9 brain* symlinks gone; those repos still get gmail/oneshot
  (verified), and any real `lib/gmail` library use migrated to a module dep.
- pair's stray `cmd/sdlc/` tree (`rm -rf` — it holds a `bin/` subdir, not empty)
  deleted.
- atlas updated (base-layer.md / a dev-tooling note); `.zshrc` one-liner documented.

## Plan

Two milestones, each a fresh-eyes `sdlc milestone-close` review boundary.

- [x] M1 — `construct/dev-aliases.sh` (the generator)
- [ ] M2 — retire nous source-symlinks + cleanup + docs

### M1 — generator
- [x] `construct/dev-aliases.sh`: sibling walk + `construct/` scope + active
      filter (`*legacy*`/`*.original`/`*[0-9]`); per-repo `cmd/X` filters (real
      dir, ≥1 `.go`, no `.private`); emit build-in-owner/run-in-cwd function to
      stdout. Bash-3.2-safe (macOS system bash — no `declare -A`/`mapfile`).
      `--workspace` test seam.
- [x] Dedup last-win (sorted walk) + stderr warn; `--strict` (exit≠0 on dup) +
      `--list` dry-run + `--help`.
- [x] `construct/scripts/test/dev-aliases.test.sh` — 17 hermetic assertions over
      a fixture sibling tree (tool, `.private`, non-buildable dir, re-export
      symlink, legacy copy, collision), incl. the function-body shape (fixture
      can't run `go build`). All pass; smoke-tested on the real workspace (10
      binaries, correct owners, 0 dup warnings). **No SKILL.md** — repo
      convention: `construct/scripts/*.sh` are documented by header comment +
      hermetic test; SKILL.md is reserved for agent skills (would wrongly surface
      this dev-env helper as invokable). `--help` + atlas note cover docs.

### M2 — retire symlinks + cleanup + docs
- [ ] **First, enumerate live gmail/oneshot consumers in brain* and pin the
      replacement path** (plan-quality finding). Verified so far: `nous/lib/gmail`
      is imported only by `nous/cmd/gmail` (no derivative lib import → the
      module-migration is a no-op); brain* have no `tool`+`replace` and no
      Makefile/script building or invoking gmail (only markdown references) — so
      today they compile gmail from the symlinks, and that's their only build
      path. Decide the replacement (PATH binary built in nous, and/or the
      dev-alias for interactive use — note the dev-alias is a shell function,
      not available to cron/scripts) BEFORE deleting the 9 symlinks. (M2 mutates
      peer repos incl. the base-layer `nous/construct/base.manifest` — downstream
      caveat per AGENTS.local.md.)
- [ ] Remove the 3 code-symlink directives (`lib/gmail`, `cmd/gmail`,
      `cmd/oneshot`) from `nous/construct/base.manifest`; remove the 9 symlinks
      across brain/brain-family/brain-private; verify those repos still obtain
      gmail/oneshot (dev-alias / PATH binary), redirect any build step off the
      symlinked source, and migrate any real `lib/gmail` library import to a
      module dep.
- [ ] `rm -rf` pair's stray `cmd/sdlc/` tree (holds a `bin/` subdir).
- [ ] atlas (base-layer.md / dev-tooling note) + the `.zshrc` `source <(...)`
      one-liner; verify end-to-end (`source` it, run sdlc + a nous binary from
      different cwds).

## Log


- 2026-06-01: closed M1 — 17-assertion hermetic test green; dev-aliases.sh smoke-tested on real workspace (10 binaries, correct owners, 0 dup warnings); --list/--strict/--help verified; review verdict: SHIP (high confidence, 0 critical/important; parser recorded SHIP correctly). Addressed 3 of 4 Minor findings post-review: validate `--workspace` (nonexistent/empty → exit 2, not silent-empty) + 2 new test assertions (now 19); fixed the atlas note's location wording (script at construct/dev-aliases.sh, test under construct/scripts/test/). Skipped #4 (quote the binary name in the build line) — Go cmd dir names can't contain spaces, so it's theoretical. M2 architectural note reaffirmed: enumerate consumers + verify no active derivative ends up a second owner before deleting symlinks.
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

Plan-quality gate (`sdlc change-code`) returned FAILURE — a real catch: the
inventory missed `symlink lib/gmail` (nous manifest line 49) + its 3 derivative
symlinks. True inventory is **3 directives / 9 symlinks** (not 2/6); `lib/gmail`
is in-scope Go source. Fixed Spec/Plan/Done-when counts, added the
library-import migration nuance, corrected the pair `cmd/sdlc/` wording (`rm -rf`,
not "empty"), and added the function-body-shape test assertion + active-clone
note. The other ~100 brain* symlinks are doc/config/script substrate — out of
scope (no module mechanism).

**M1 done.** `construct/dev-aliases.sh` (bash-3.2-safe — macOS system bash is
3.2.57, no `declare -A`/`mapfile`) walks active ariadne-styled siblings, emits a
build-in-owner/run-in-cwd function per owned `cmd/X`; filters skip re-export
symlinks + non-buildable dirs + `.private`; last-win dedup with stderr warn,
`--list`/`--strict`/`--help`/`--workspace`. 17-assertion hermetic test
(`construct/scripts/test/dev-aliases.test.sh`) green; real-workspace smoke =
10 binaries, correct owners (gmail/oneshot/nous→nous, sdlc→ariadne; brain
symlinks + pair's empty cmd/sdlc excluded), 0 dup warnings. No SKILL.md (matches
`construct/scripts/*.sh` convention; documented via header + test + atlas note in
base-layer.md). Plan-quality gate caught the `lib/gmail` inventory miss (3
directives/9 symlinks) before coding — folded into M2.
