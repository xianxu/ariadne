---
id: 000056
status: working
deps: []
github_issue:
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 5
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

**Consolidations:**
- Extract `nextIssueID` + the template renderer out of `fetch.go` into
  `internal/issue` (shared scaffold), used by both `issue new` and the
  `--from-github` path.
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
- `sdlc issue new --from-github N` reproduces current `sdlc fetch`; `sdlc fetch`
  still works (alias).
- `sdlc issue set-status` works with guards; `sdlc set-status` still works
  (deprecated alias).
- `sdlc issue list` / `show` work.
- next-ID + render logic lives once in `internal/issue`, used by both paths.
- `xx-issues` SKILL is a pointer; the issue contract lives in `sdlc issue --help`.
- AGENTS.md / atlas / code refs to `sdlc set-status` + `sdlc fetch` updated.
- Tests for each verb; `go test ./cmd/sdlc/...` green; both binaries rebuilt.

## Plan

Three milestones, each a real fresh-eyes `sdlc milestone-close` review boundary.

### M1 — `issue` group + `new` + scaffold extraction
- [ ] Extract `nextIssueID` + template renderer from `fetch.go` into
      `internal/issue` (e.g. `scaffold.go`) with unit tests (pure: existing IDs
      → next; fields → rendered text).
- [ ] Add the `sdlc issue` parent cobra command + `helptext/issue.md` (group
      overview + issue-file contract).
- [ ] Implement `sdlc issue new "<title>" [--slug --from-github N --deps
      --target]`; print path; tests (blank issue, ID allocation, slug
      derivation).

### M2 — move `set-status` in + fold `fetch` + `list`/`show`
- [ ] Move `set-status` under `sdlc issue set-status`; keep flat `sdlc
      set-status` as a hidden deprecated alias; update internal references.
- [ ] Reimplement `sdlc fetch` as `sdlc issue new --from-github N` over the
      shared scaffold; keep `sdlc fetch` as a hidden alias; preserve the
      GH-close-on-archive behavior.
- [ ] `sdlc issue list [--status]` + `sdlc issue show <N>`; tests.

### M3 — consolidate docs/skill + reference sweep
- [ ] Shrink `xx-issues` SKILL to trigger + pointer; migrate the
      frontmatter-field / sections contract into `issue.md` helptext.
- [ ] Update AGENTS.md §2 issue flow, `atlas/workflow/{sdlc-binary,issue-lifecycle}.md`,
      and code/helptext refs to `sdlc set-status` / `sdlc fetch`.
- [ ] Verify: full `go test ./cmd/sdlc/...`, rebuild both binaries, smoke-test
      each verb.

## Log

### 2026-05-31
Created from a brainstorm. Decided the boundary: option **A** (additive `issue`
group for CRUD; checkpoint guards stay flat), with `set-status` moved into the
group since it's a field edit, not a key checkpoint. Grounding: `fetch.go`
already holds the deterministic next-ID + render core, so `issue new` is mostly
an extraction; the manual ID scan in the `xx-issues` skill is the racy step this
removes. Skill-shrink mirrors the `xx-sdlc` static-pointer move just landed.
