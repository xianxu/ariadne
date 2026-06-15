---
id: 000096
status: done
deps: [ariadne#95]
github_issue:
created: 2026-06-14
updated: 2026-06-15
estimate_hours:
---

# weave: prune stale lowered symlinks (.claude/skills GC)

## Problem

weave (like `setup.sh`/`sync-local-skills.sh` before it) *creates* lowered symlinks — `.claude/skills/<prefix><name>` — but never *prunes* stale ones. When a skill is renamed (`xx-fix` → `xx-repair`) or the local prefix changes (`xx-` → `yy-`), weave writes the new lowered symlink and leaves the old one behind (dangling, or pointing at a gone target). Pre-existing gap inherited from the bash; surfaced during the weave (#95) cutover.

Note (a non-issue): the *source* dirs `construct/local` / `construct/adapted` are symlinked **as whole dirs**, so a source-side rename auto-propagates downstream via the dir symlink. The staleness is *only* in the per-name **lowered** symlinks weave emits.

## Spec

A prune/GC pass over weave's lowered locations (`.claude/skills/`, and any future lowered dir weave owns):
- Basic logic: a symlink in a lowered location whose target is **gone/dangling** → delete.
- Better: weave knows the *full intended set* of lowered symlinks for a repo, so remove any lowered symlink it did **not** (re)produce this run (orphan removal) — covers the rename/prefix-change case even when the old target still exists elsewhere.
- Scope to lowered dirs weave manages; never touch non-weave files in those dirs (track weave-owned entries).

## Done when

- Renaming a skill or changing the local prefix, then re-running weave, leaves **no** stale `.claude/skills/<old>` symlink.
- A dangling lowered symlink is removed on the next weave run.
- weave never deletes a lowered entry it doesn't own.

## Plan

Pure decision + IO seam (ARCH-PURE), post-apply pass — modeled on gitignore.go
(a pure transform + an IO seam, appended in main.planActions). Safety lives in
the pure `shouldPrune` and its tests.

- [x] Pure `shouldPrune(c, producedSet, sourceRoots)` — prune iff symlink ∧ NOT in producedSet ∧ weave-owned (lexically-resolved target under a lowering source root, dangling-safe). Real file/dir/non-weave symlink → KEEP. Plus the pure `ProducedSymlinkSet`/`ManagedLocations`/`SourceRootsFromPaths`/`PrunePlan` from `[]plan.Action`. (`cmd/weave/internal/plan/prune.go`)
- [x] IO seam `PruneOrphans(fs, repoRoot, actions, sourceRoots)`: per managed location, `ScanManagedSymlinks` (read-only), `PrunePlan`, unlink the orphans; return pruned paths. Run after Apply in `run`.
- [x] Wire dry-run: `PrunePreview` (read-only) + `formatPrunes` → `--dry-run` shows what WOULD be pruned; apply prints a `pruned N` summary.
- [x] Tests: managed-location fixture (a)produced→KEEP (b)orphan-prefix→PRUNE (c)dangling→PRUNE (d)real dir→KEEP (e)real file→KEEP (f)non-weave symlink→KEEP; asserts exactly b+c pruned. Idempotency + dry-run-is-read-only + managed-location + sibling-prefix-not-within tests. (`prune_test.go`)
- [x] Verify: ariadne self-walk dry-run = 0 prunes (clean); parley.nvim dry-run = exactly its 2 genuine orphans, 0 false positives; `go build ./... && go test ./cmd/weave/... && go vet && gofmt -l cmd/weave` clean.

## Log

### 2026-06-14

Filed alongside the #95 cutover; deferred to land on `000095-weave` (where the
dead `setup.sh`/`merge-settings.sh`/`sync-local-skills.sh` symlinks the prune
must reap originate).

### 2026-06-15 — implemented (TDD) on `000095-weave`

Implemented as a post-Apply pass, modeled on `gitignore.go`'s pure-transform +
IO-seam split (ARCH-PURE).

**Prune criteria (a symlink is removed ONLY if ALL hold; any failure / any
uncertainty ⇒ KEEP):**
1. it is a SYMLINK (the IO scan `ScanManagedSymlinks` skips real files/dirs — a
   repo's authored content never even becomes a candidate; `shouldPrune`
   re-asserts `IsSymlink` as a guard);
2. it sits in a weave-MANAGED location — a directory weave produced ≥1 symlink
   into THIS run (`ManagedLocations`, derived from the produced `Symlink`
   actions' `Dst` dirs — never hardcoded; so a self-walk owning
   `construct/scripts/` as real files never scans it);
3. it is weave-OWNED — its target, resolved LEXICALLY against the link's dir
   (so a dangling link still yields the path it WOULD point at), lies within a
   lowering SOURCE ROOT (the resolved layer roots, `SourceRootsFromPaths`);
   `targetWithinAnyRoot` uses `filepath.Rel` and rejects a `../` escape, so a
   sibling dir sharing a string prefix is NOT counted as within;
4. it is NOT in the produced-this-run set (`ProducedSymlinkSet` over the
   `Symlink.Dst`s) — the orphan condition.

**Files:** `cmd/weave/internal/plan/prune.go` (pure decision + `ScanManagedSymlinks`/`PruneOrphans`/`PrunePreview` IO seam), `prune_test.go` (the safety fixture). Wired in `main.go`: `run` calls `PruneOrphans` after `plan.Apply` (prints a `pruned N` summary), and `PrunePreview` + `formatPrunes` for the `--dry-run` preview. Golden/verify-complete untouched — the prune is not a planned Action, so the classifier is unaffected (re-ran golden on ariadne self: 0 UNEXPECTED).

**Tests:** the managed-location fixture asserts exactly (b) orphan-prefix +
(c) dangling are pruned, while (a) produced, (d) real dir, (e) real file, (f)
non-weave symlink are KEPT; plus idempotency, dry-run-is-read-only,
managed-location derivation, and sibling-string-prefix-not-within. All green;
`go build ./...`, `go test ./cmd/weave/...`, `go vet ./cmd/weave/...`, `gofmt -l
cmd/weave` clean.

**Real-repo verify (read-only / temp-copy for the destructive path):**
- ariadne self-walk `compile --target claude --dry-run` → **0 prunes** (no
  dangling links in its managed dirs; the healthy-repo over-prune guard).
- parley.nvim dry-run → **exactly 2 prunes, 0 false positives**:
  `.claude/skills/xx-data` (orphan of the `construct/local/data`→`datatype`
  rename, commit `0cc1c12`) + `scripts/sdlc-bootstrap.sh` (dead link to the
  removed ariadne script, `0038fee`). NOTE: the spec expected parley to be
  already-clean (0), but it still carries these 2 genuine orphans — precisely
  the staleness #96 targets; they self-clean when parley is woven during its
  cutover. The prune flagged ALL and ONLY the 2 dangling links in parley's
  managed dirs (every valid `construct/*` link KEPT).
- Destructive end-to-end on a `/tmp` copy of parley.nvim: real `compile` pruned
  the 2 orphans, KEPT all valid links + real files (`README.md`,
  `construct/base.manifest`), and a second `compile` pruned nothing (idempotent).

Done-when satisfied: rename / prefix-change / dangling lowered symlinks are
removed on the next weave; weave never deletes a lowered entry it doesn't own
(safety fixture + the ariadne 0-prune + the parley 0-false-positive checks).
