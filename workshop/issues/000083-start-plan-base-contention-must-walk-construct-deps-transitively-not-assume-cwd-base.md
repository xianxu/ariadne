---
id: 000083
status: working
deps: []
github_issue:
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 2.0
---

# start-plan base-contention must walk construct/deps transitively, not assume cwd==base

## Problem

#82 M3 added a `start-plan` base-contention heads-up, gated on a predicate
`isBaseRepo(root)` = "`construct/` is a real directory, not a symlink". **That
premise is false**, so the feature is wrong in two ways:

1. **Wrong signal.** Every repo — base *and* derivative — has a *real*
   `construct/` directory carrying its own `base.manifest` / `deps` / `scripts`,
   with only specific *subpaths* symlinked (`construct/adapted -> ../../ariadne/
   construct/adapted`). Verified: `nous` and `pair` both have a real `construct/`.
   So `isBaseRepo` returns `true` in **every** derivative, and `start-plan` fires
   `base (nous): …` / `base (pair): …` in derivatives — mislabeling the derivative
   as the base and reporting *its own* git state as "base contention". The exact
   opposite of the intended "silent unless you're looking at the base". The
   shipped `TestIsBaseRepo` passed only because it encoded the same wrong model
   (real-dir vs a synthetic symlinked `construct/`).

2. **Wrong shape.** It's not "one base". The substrate graph is a **chain**
   declared in `construct/deps` (the sole substrate-graph carrier since #60):
   - `ariadne/construct/deps` — absent → ariadne is the **root** (no upstream).
   - `nous/construct/deps` → `substrate ../ariadne`.
   - `brain/construct/deps` → `substrate ../ariadne` (+ a `data` row).
   Because a repo reads **all** its transitive upstreams' working trees live (via
   the symlinks), the "moving ground" you build on is the whole **dependency
   path**, not a single base. Working in `brain`, contention in `ariadne` (and
   any intermediate substrate it depends on) is what matters — and `start-plan`
   in a derivative should surface *that*, which was the Spec's primary motivation
   ("base changes are discovered mid/late in a derivative session"). The cwd-only
   gather never reads upstream trees at all.

(Vendoring is gone — symlink-only model now — so there is no vendor-mode branch to
handle; `construct/deps` is the single source of the dependency edge.)

## Spec

Replace `isBaseRepo` + the cwd-only gather with a **transitive walk of
`construct/deps`**, reusing the established parser semantics (the shell
`construct/scripts/lib-deps.sh` `deps_substrate_targets`; no Go port exists yet,
so add a minimal one).

- **The dependency path** = the transitive closure of `substrate <path>` rows
  starting at cwd's repo root. Resolve each target relative to its declaring
  repo root, dedupe, skip absent peers (present-walker semantics — don't abort on
  a missing sibling, unlike the clone driver).
- **Report contention per repo on the path** (the upstreams you read live):
  branch / dirty *code* (tracker files still excluded — reuse `assessDirty.Blocking`)
  / other `status: working` issues — gathered against each upstream root via
  `git -C <root>` and `listIssues(<root>/workshop/issues)`.
- **Root case** (cwd has no upstream, i.e. you ARE ariadne): the path is empty →
  fall back to reporting cwd itself (its concurrent in-flight work), preserving
  the one case #82 M3 happened to get right.
- Still **non-blocking**, still emitted after the architecture block.

### Pure / IO split (ARCH-PURE)

- **Pure** (unchanged, keep): `baseContention` struct (now the per-repo unit),
  `Clean()`, `baseContentionSummary(baseContention) string`, `issueRef`.
- **Pure** (new): `parseSubstrateTargets(depsContent string) []string` — the
  parse half of `deps_substrate_targets` (word-split, `#` comments,
  `kind=="substrate"`). Table-tested.
- **IO seam** (new/changed): `substrateChain(root) []string` (transitive walk,
  resolve + dedupe + present-skip); `gatherBaseContention(root, excludeIssue)`
  retargeted to an arbitrary root via `GitInDir`.
- **Delete**: `isBaseRepo` + `TestIsBaseRepo` (wrong premise).

## Done when

- `sdlc start-plan` in a **derivative** (e.g. nous) prints contention for its
  **upstream(s)** (`base (ariadne): …`), NOT for the derivative itself; in
  **ariadne** (root) it reports ariadne's own concurrent work as today.
- Detection is driven by `construct/deps` (transitive), not by `construct/`'s
  link state. `isBaseRepo` is gone.
- A repo whose `construct/deps` names a 2-hop chain surfaces every hop on the
  path (transitive walk, deduped, absent peers skipped).
- Tests: pure `parseSubstrateTargets` (substrate rows, comments, blanks, data
  rows ignored, absolute vs relative) + a transitive-walk test over a temp
  fixture tree (A→B→C dirs with `construct/deps`); existing pure summary/issueRef
  tests still pass. `go test ./cmd/sdlc/...` green.
- atlas `sdlc-binary.md` + `base-layer.md` corrected (drop the real-dir-vs-symlink
  story; describe the deps-chain walk).

## Plan

Single-pass atomic correction (one feature, contained to startplan.go + tests +
atlas). No Mx.

- [ ] Add pure `parseSubstrateTargets(depsContent) []string` (port the parse half
      of `lib-deps.sh deps_substrate_targets`: strip `#`, word-split, keep
      `substrate` rows' target). Table-test it.
- [ ] Add `substrateChain(root) []string` — transitive walk: read
      `<root>/construct/deps`, resolve each `substrate` target relative to its
      declaring root (`filepath.Join` + `Clean`; `EvalSymlinks` best-effort),
      dedupe via a seen-set, skip non-existent dirs, recurse. Test over a temp
      A→B→C fixture.
- [ ] Retarget `gatherBaseContention(root, excludeIssue)` to read an arbitrary
      root via `mergeRunner.GitInDir(root, …)` for branch + `status --porcelain`
      (→ `assessDirty(...).Blocking` count) and `listIssues(filepath.Join(root,
      issuesDir))`. (Pure `assessDirty`/summary reused unchanged.)
- [ ] Rewrite `runStartPlan`'s tail: `chain := substrateChain(root)`; if empty →
      `[root]` (root reports self); for each repo print `cok`/`cwarn` of
      `baseContentionSummary` by `Clean()`.
- [ ] Delete `isBaseRepo` + `TestIsBaseRepo`.
- [ ] Update atlas `sdlc-binary.md` + `base-layer.md` (deps-chain walk, not
      construct-symlink).
- [ ] `go test ./cmd/sdlc/...` green; verify live: `sdlc start-plan` in ariadne
      (self) and, if a derivative is handy, in nous (reports ariadne).

## Log

### 2026-06-04
Filed as a fast-follow on #82 M3, after the operator caught two false premises in
the shipped detection: (1) `construct/` is a real dir in derivatives too (only
subpaths are symlinked), so `isBaseRepo` fires everywhere; (2) the right signal is
the **dependency chain** in `construct/deps` (`ariadne < nous < brain`-style),
walked transitively — warn on the whole upstream path, since a repo reads all its
transitive upstreams' working trees live. Vendoring has been removed (symlink-only
now), so no vendor-mode branch. Reuse target: `construct/scripts/lib-deps.sh`
`deps_substrate_targets` (parse semantics) — no Go port exists yet, add a minimal
pure one. The pure `baseContentionSummary` from #82 M3 is correct and stays; only
the detection + gather (cwd→deps-chain) change. See [[000082]] (M3 shipped the
cwd==base version).
