---
id: 000056
status: working
deps: []
github_issue:
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 7
---

# Lift the issue subsystem into `sdlc issue`

## Problem

Issue creation/editing is agent-driven prose. The `xx-issues` skill tells the
agent to *"find the highest existing ID and increment by 1"* and write the
template by hand — a deterministic step done manually, which is racy under
parallel workstreams and off-by-one prone. The deterministic core already
exists inside `fetch.go` (`nextIssueID` + `renderFetchedIssue`) but isn't
surfaced as a reusable verb. The subsystem is mature (100s of issues) and
deserves a first-class binary surface, the way SDLC checkpoints got one.

## Spec

Add a noun-grouped `sdlc issue` subcommand for issue-record CRUD, **complementing
(not replacing)** the flat checkpoint guards.

**Boundary (decided — option A + set-status moved in).** `sdlc issue *` owns
authoring/CRUD of the issue *record*. The flat verbs (`close`, `claim`,
`change-code`, `pr`, `merge`, `milestone-close`, `judge`, `push`, `state`) stay
flat — they guard workflow *transitions*, not records. `set-status` moves into
the group (it's a field edit, not a key checkpoint).

**Verbs:**
- `sdlc issue new "<title>"` — allocate next ID (scan `issues/` + `history/`),
  write the standard template, print the path. `--from-github N` folds in
  today's `sdlc fetch`. Optional: `--slug`, `--deps`, `--target`.
- `sdlc issue set-status <N> <status>` — moved from flat `sdlc set-status`, same
  transition guards.
- `sdlc issue list [--status S]` — id / status / title (richer than `sdlc
  state`'s working-only view).
- `sdlc issue show <N>` — frontmatter summary + section headers (structured peek
  without loading the whole file). Lowest priority.

**Canonical template (resolves the 3-shapes ambiguity).** Today three issue
templates disagree (`renderFetchedIssue`, the `xx-issues` skill, this file). The
canonical one — what `issue new` writes and `issue --help` documents — is pinned
as (described in prose here to avoid colliding with the structural-check parser;
the literal bytes land in `helptext/issue.md` in M1):

- Frontmatter, in order: `id`, `status: open`, `deps: []`, `github_issue:`,
  `created`, `updated`, `estimate_hours:` (optional at create, required at
  →working). NOT `actual_hours` (added at close).
- Body sections, in order: a `# <Title>` heading, then `Problem`, `Spec`,
  `Done when` (seeded with one empty `-` placeholder for the author to fill;
  `change-code` later requires it non-empty), `Plan` (seeded with one `- [ ]`
  item), `Log` (seeded with a dated `###` entry). NOT a `Side quests` section
  (added when needed).
- **One parameterized renderer**, not two: `--from-github N` writes the same
  skeleton but sets `github_issue: N` and fills the `Problem` section with the
  fetched GH body (replacing fetch's loose post-title paragraph). Blank `new`
  leaves `Problem`/`Spec` empty.

**Consolidations:**
- Extract `nextIssueID`, `slugify`, and the parameterized renderer out of
  `fetch.go` into `internal/issue` (shared scaffold), used by both `issue new`
  and the `--from-github` path.
- `sdlc fetch` → `sdlc issue new --from-github N`; keep `sdlc fetch` as a hidden
  deprecated alias.
- `sdlc issue --help` becomes the single source of truth for the issue-file
  contract (frontmatter fields, required sections, statuses) — migrated out of
  the `xx-issues` skill.
- Shrink `xx-issues` SKILL to a trigger + pointer to `sdlc issue --help`, keeping
  only agent-judgment prose (what makes a good Spec, when to split milestones).
  Same move as `xx-sdlc`.

**Back-compat:** flat `sdlc set-status` and `sdlc fetch` stay as hidden
deprecated aliases for one cycle so existing references keep working while
they're updated.

**Non-goals:** no change to `close`/`claim`/`milestone-close` semantics; no new
content validation beyond existing `CheckStructural`; project/roadmap datatypes
untouched.

## Done when

- `sdlc issue new "x"` creates `workshop/issues/NNNNNN-x.md` with the correct
  next ID + template, prints the path.
- `sdlc issue new --from-github N` reproduces current `sdlc fetch`'s *behavior*
  over the canonical template; `sdlc fetch` still works (alias).
- `sdlc issue set-status` works with guards; `sdlc set-status` still works
  (deprecated alias).
- `sdlc issue list` / `show` work.
- next-ID + render logic lives once in `internal/issue`, used by both paths.
- `xx-issues` SKILL is a pointer; the issue contract lives in `sdlc issue --help`.
- AGENTS.md / atlas / code refs to `sdlc set-status` + `sdlc fetch` updated.
- Tests for each verb; `go test ./cmd/sdlc/...` green; both binaries rebuilt.

## Plan

Three milestones, each a real fresh-eyes `sdlc milestone-close` review boundary.

- [x] M1 — `issue` group + `new` + scaffold extraction
- [x] M2 — move `set-status` in + fold `fetch` + `list`/`show`
- [ ] M3 — consolidate docs/skill + reference sweep

### M1 — `issue` group + `new` + scaffold extraction
- [x] Extract `NextID`, `Slugify`, and the parameterized `Render` (per the
      canonical template above) from `fetch.go` into `internal/issue/scaffold.go`
      with unit tests (existing IDs → next; title → slug; blank vs
      `--from-github` body → rendered text).
- [x] Add the `sdlc issue` parent cobra command + `helptext/issue.md` (group
      overview + issue-file contract).
- [x] Implement `sdlc issue new "<title>" [--slug --from-github N --deps
      --target]`; print path; tests (blank issue, ID allocation, slug
      derivation, `--from-github` fills Problem, blank `new` leaves
      Problem/Spec present-but-empty).

### M2 — move `set-status` in + fold `fetch` + `list`/`show`
- [x] Move `set-status` under `sdlc issue set-status`; hidden deprecated flat
      `sdlc set-status` alias kept. (M1 review #2 was already satisfied — the
      guards live in `applyStatus`/`checkTransitionGuards` as returned errors,
      unit-tested; only the cobra wiring relocated. `setstatus_test.go` targets
      those functions, not the command tree, so no re-pointing was needed — the
      flat→group wiring is covered by the new alias test instead.)
- [x] Reimplement `sdlc fetch` as a thin call to `runIssueNew --from-github`
      over the shared scaffold; deleted `renderFetchedIssue`; `sdlc fetch`
      (keeping `--github-issue`) hidden + deprecated. GH-close-on-archive
      preserved (it reads `github_issue:` frontmatter, still set).
- [x] Updated `fetch_test.go`: deleted `TestRenderFetchedIssue_Shape` (function
      gone; canonical render covered by `scaffold_test`); the integration tests
      pass through `runIssueNew` unchanged (assertions are shape-agnostic).
- [x] `sdlc issue list [--status]` (reuses `listIssues`, sorted by ID) +
      `sdlc issue show <N>` (frontmatter + headers, no bodies); tests for both.
- [x] Alias tests: execution-based `set-status` flat-vs-grouped both mutate
      identically (via `buildRoot`, extracted from `main` to make the tree
      testable); structural check that flat `set-status`/`fetch` are
      hidden+deprecated and the grouped verbs resolve.

### M3 — consolidate docs/skill + reference sweep
- [ ] Shrink `xx-issues` SKILL to trigger + pointer; migrate the
      frontmatter-field / sections contract into `issue.md` helptext.
- [ ] Update AGENTS.md §2 issue flow (**base-layer file — write §2 for the
      general downstream audience, not ariadne-local specifics**),
      `atlas/workflow/{sdlc-binary,issue-lifecycle}.md`, and code/helptext refs
      to `sdlc set-status` / `sdlc fetch`.
- [ ] Verify: full `go test ./cmd/sdlc/...`, rebuild both binaries, smoke-test
      each verb.

## Log


- 2026-05-31: closed M1 — go test ./cmd/sdlc/... green; sdlc issue new smoke-tested (dry-run renders canonical template); GetField empty-field regression added; review verdict: SHIP (high confidence, 0 critical). milestone-close auto-recorded "unknown" because the judge put SHIP under `## 1. Verdict` not line 1 — parser fragility, verdict corrected here by hand. Important findings addressed pre-M2: atlas issue-lifecycle.md updated for blank `issue new` + the one-step claim; `--deps` CLI test added. Deferred-with-note: die()→returned-errors for testable guards (apply to set-status in M2).
### 2026-05-31
Created from a brainstorm. Decided the boundary: option **A** (additive `issue`
group for CRUD; checkpoint guards stay flat), with `set-status` moved into the
group since it's a field edit, not a key checkpoint. Grounding: `fetch.go`
already holds the deterministic next-ID + render core, so `issue new` is mostly
an extraction; the manual ID scan in the `xx-issues` skill is the racy step this
removes. Skill-shrink mirrors the `xx-sdlc` static-pointer move just landed.

Plan-quality gate (`sdlc change-code`) flagged that "the standard template" was
underspecified — three template shapes exist today. Resolved by pinning a
canonical template in the Spec (one parameterized renderer; `--from-github`
fills `## Problem`). Also folded in the judge's smaller notes: `slugify` joins
the M1 extraction, `setstatus_test.go` re-points in M2, the `--from-github` vs
`--github-issue` flag split is called out, `list`/`show` test cases are named,
and the M3 AGENTS.md edit is marked base-layer (downstream audience).

Second plan-quality pass (re-run) again returned INFO / "safe to start." Folded
in its material findings: added the `fetch_test.go` update to M2 (the canonical
template breaks `TestRenderFetchedIssue_Shape`), fixed the backwards Done-when
wording (structural requires a *non-empty* bullet), reworded "reproduces" →
"reproduces behavior," and bumped the estimate 5h→7h (dual aliases + doc sweep +
test re-pointing realistically cost more). Proceeded past the advisory gate with
`--force` since the judge cleared it as safe twice.

**M1 done.** Extracted `NextID`/`Slugify`/`Render` into
`internal/issue/scaffold.go`; `fetch.go` now delegates `NextID`/`Slugify` (output
unchanged — `renderFetchedIssue` stays until M2). Added `sdlc issue` group +
`issue new` (`issue.go`) + `helptext/issue.md` (the canonical-template contract).
Side-quest: found + fixed a latent bug in `issue.GetField` — its `\s*` gap
matched newlines, so an empty frontmatter field followed by another line (e.g.
`github_issue:` then `created:`) captured the *next* line's value. Changed to
`[ \t]*` + regression test (the old test only covered an empty field at EOF,
which hid it). All `go test ./cmd/sdlc/...` green; both binaries rebuilt;
`sdlc issue new` smoke-tested (dry-run renders canonical template).

**M2 done.** `set-status` relocated under `sdlc issue set-status`; `fetch`
folded into a thin `runIssueNew --from-github` call (deleted
`renderFetchedIssue`); both `set-status` + `fetch` kept as hidden+deprecated
flat aliases (cobra `Deprecated` prints a migration notice). Added `issue list`
(reuses `listIssues`) + `issue show`. Extracted `buildRoot()` from `main` so the
command tree is testable; alias test executes both flat + grouped `set-status`
and asserts identical mutation. Discovered the M1-review "re-point setstatus_test"
item was moot — those tests target `runSetStatus`/guards (unchanged), so the
flat→group risk is covered by the new alias test. Smoke-tested list/show/the
deprecation notice. `go test ./cmd/sdlc/...` green; vet clean; binaries rebuilt.
