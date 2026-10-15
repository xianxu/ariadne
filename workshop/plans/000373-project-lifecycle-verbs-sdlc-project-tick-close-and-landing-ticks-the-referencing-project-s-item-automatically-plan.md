# Project lifecycle verbs Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Landing an issue ticks it in every referencing project, including the TL-owned ones in ops. `sdlc project tick` does the same by hand. `project status` shows MVP and done_when progress, and `project close --verified` closes a real ops project.

**Architecture:** Ticking at landing is one pure edit, `landingTick`. It sets `[ ]`/`[.]` to `[x]` and appends `(PR #N)`. It runs on each referencing project's **origin main** through the existing compare-and-swap path, `gitx.TrunkFile` (the #308 peer publish). It never runs in someone's checkout. `sdlc merge` calls it after the archive step, and `sdlc project tick --issue N` is the same call made by hand. Status and close gain an MVP and done_when view, `--verified` and `--dry-run`, and they handle ops's symlinked records.

**Tech Stack:** Go, cobra, `gitx.TrunkFile`, the existing `internal/project` Doc and guards.

---

## What exists, and the gaps

| Exists | Gap (this issue) |
|---|---|
| `project.TickAllTaskRowsForIssue`, `projectCloseEdit` (close-time tick) | No tick at **landing**; no `project tick` verb |
| `project.DiscoverInRoots` fleet-wide find, `project find` | — (reused) |
| `publishPeerProject`: CAS publish to a peer's origin (#308) | It refuses TL-owned projects (#326), and every live ops project is TL-owned |
| `project status`: board, frontier, forecast | No MVP landed/open against `mvp_scope`; no `done_when` |
| `project close`: lifecycle, retro gate, archive move | No `--verified`; no MVP re-check; no `--dry-run`; on an ops symlink it renames the **link** and leaves `teams/…/project.md` orphaned |

## The ownership call (TL decides)

#326 made TL-owned projects off-limits to workers, so the TL hand-ticks ops after every landing. Proposal:
- A **landing tick** may write a TL-owned project. It is a fact (the issue merged, and here is the PR), not a TL decision. It changes only checkbox state and appends `(PR #N)`. It adds no detail block and no Log or Decisions text.
- It is written as one commit on ops **origin** main through TrunkFile, never in the TL's checkout (kink 38). The TL's next push needs a `git pull --rebase` when origin moved. The edit is a one-line change on a different line, so a conflict needs the TL to be editing that same row.
- Close-time behaviour stays as #326 left it (no tick or snapshot of a TL-owned project at `sdlc close`). The tick moves to landing, where the TL already ticks by hand.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `project.TickLandedIssue(text, repo, id, pr)` | `cmd/sdlc/internal/project/project.go` | new |
| `project.MVPProgress(meta, terminal)` | `cmd/sdlc/internal/project/mvp.go` | new |
| `landingTickTransform(ref, pr)` | `cmd/sdlc/projectlandtick.go` | new |
| `projectCloseEdit` | `cmd/sdlc/projecttick.go` | unchanged (close-time path) |

- **TickLandedIssue** — this reuses `TickAllTaskRowsForIssue`'s row match (ARCH-DRY: one matcher, `[ ]` and `[.]` only, so `[-]` cancelled rows stay). It also appends ` (PR #N)` to each matching row that doesn't already mention `PR #N`, ticked or not. It is idempotent: a second run returns identical bytes, so nothing publishes.
- **MVPProgress** — takes the `mvp_scope` refs and an `IssueTerminal` lookup and returns landed and open lists plus lookup errors. The `mvp-landed` guard is rewritten on top of it, so status, close and the guard share one definition of "MVP done" (ARCH-DRY).
- **landingTickTransform** — the bytes-to-bytes edit TrunkFile applies on origin's copy: parse, then `TickLandedIssue`. It does not check `tl:` (see the ownership call). A copy that fails to parse returns an error, and origin is left alone.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `tickLanded(ctx, ref, pr)` | `cmd/sdlc/projectlandtick.go` | new | `DiscoverInRoots` + `gitx.TrunkFile` per repo |
| `sdlc project tick` | `cmd/sdlc/projectlandtick.go` | new | cobra |
| merge post-archive hook | `cmd/sdlc/merge.go` | modified | calls `tickLanded` |
| `project status` MVP block | `cmd/sdlc/projectstatus.go` | modified | card lookup |
| `project close` `--verified`/`--dry-run`/symlink | `cmd/sdlc/projectclose.go` | modified | filesystem |

- **tickLanded** — finds every active project that references `repo#id`, fleet-wide, through the same overlay roots close uses. For each project, in its own repo or a peer, it resolves the real path (`peerProjectPath`, which follows the ops symlink) and runs `peerTrunk(...).Update(rel, "project: <slug>: tick <ref> landed (PR #N)", landingTickTransform)`. Its own repo goes through origin too, because at landing the branch has merged and the resting branch only fast-forwards. A brain repo is refused, as now. A row that only appears as a train car is reported and skipped.
  - **Failure policy:** any failure warns, and merge still succeeds. The issue has landed, and the message names `sdlc project tick --issue N --pr M` as the retry.
- **`sdlc project tick --issue <ref> [--pr N] [--dry-run]`** — the same call. `--pr` defaults to the landed PR that merge passes in. By hand it is optional and omits the suffix. `--dry-run` prints each target and the row diff.
- **merge hook** — runs after the archive step, with the PR number merge already has.
- **status** — adds a `mvp: L/T landed · open: refs` line (plus any lookup errors) and a `done_when:` line with the frontmatter text.
- **close:**
  - `--verified '<evidence>'` is required for `close` (not `--drop`) and recorded in the close Log entry.
  - The `mvp-landed` check reruns at close, because scope can grow while a project sits in `closing`.
  - `--dry-run` evaluates every guard and prints the transition and the archive destination without writing anything.
  - For a symlinked live record, close writes the updated text to the **real** file and moves the **link** into `workshop/history/projects/`, retargeted. The team folder (handoffs, reviews) stays where it is.

### Envelope and principles
- **ARCH-CONSTRAINTS:** landing tick runs once per merge (an interactive CLI, after the PR merged). Discovery takes milliseconds over the sibling repos. Each matching repo costs one fetch and one push, usually 1–2 repos, bounded by TrunkFile's attempts. Memory, CPU, concurrency: N/A.
- **ARCH-SECURE:** origin's project bytes are untrusted. They are parsed into a Doc, and a parse failure refuses and leaves origin alone. The PR number is an int from merge's own state. No credentials beyond git's own push.
- **ARCH-ORDER:**
  - TrunkFile's compare-and-swap re-reads a moved origin and re-applies the pure edit.
  - A merge killed mid-tick leaves either no commit or one complete commit, and a rerun of `project tick` is idempotent.
  - A TL editing the same row concurrently is the one interleaving that needs a human: the TL's rebase conflicts on one line.
- **ARCH-FUNERAL:** creates nothing durable but one commit per landing per referencing project (a few hundred bytes), which lives as ops history.
- **ARCH-MOCK:** git through real temp repos with bare origins (testfix), as for #308's peer-publish tests. GitHub is not touched; the PR number is an input.
- **ARCH-PURE:** the row edit and the MVP math are pure, and the IO is discovery plus TrunkFile.
- **ARCH-PURPOSE:** the Done-when's cross-repo test is Task 3.

## Tasks

- [ ] **Task 1: pure core.** TDD `TickLandedIssue` with these cases:
  - open row
  - blocked `[.]`
  - cancelled `[-]` untouched
  - already ticked without PR (suffix added)
  - already has `PR #N` (no change)
  - milestone rows
  - a second run is byte-identical

  Then `MVPProgress`; rewrite the `mvp-landed` guard on it, and its existing tests stay green.
- [ ] **Task 2: `tickLanded` and the `project tick` verb.** Temp fleet: the issue's repo, plus a peer with a bare origin and a project whose path in workshop/projects is a symlink into `teams/x/projects/p/project.md` with `tl:` set. Assert:
  - origin main gains one commit to the real file;
  - the peer checkout's HEAD, index and worktree are unchanged;
  - a rerun publishes nothing;
  - `--dry-run` writes nothing.
- [ ] **Task 3: merge hook.** Extend the merge integration test fixture (fake gh PR state) with a referencing project in a second repo, and assert the tick lands after archive. Also test that a tick failure (unreachable origin) warns and the merge still succeeds.
- [ ] **Task 4: status.** MVP and done_when lines; table test via `buildRoot()` with a fake card lookup.
- [ ] **Task 5: close.** Cover these in tests:
  - `--verified` required and logged
  - the MVP re-check refuses with the open refs
  - `--dry-run` writes nothing
  - a symlinked record: the real file gets the update and the link moves to history, retargeted
- [ ] **Task 6: docs.** Help text for `project tick`/`status`/`close` (`cmd/sdlc/helptext`); `construct/datatype/project.md` § Ownership, so a landing tick writes TL-owned projects on origin; `atlas/workflow/sdlc-binary.md` row.
- [ ] **Task 7: verify on the real file.**
  - `sdlc project status --slug ariadne-robustness-1 --projects-dir ~/workspace/ops/workshop/projects`
  - `project close … --verified x --dry-run`, read-only
  - `project tick --issue ariadne#240 --dry-run`, an already-ticked row, so it should report "no change"
  - Then the full suite under the gate, and close.

## Out of scope
- Rewriting the TL's existing hand ticks, or backfilling `(PR #N)` on old rows.
- Ticking the trains table (cars have no checkbox; the forecast reads cards).
- Archiving ops team folders (the link moves; the folder stays).

## Revisions

### 2026-10-15 — TL review of the plan
- Points 2–4 approved as written. Point 1 changes #326's ownership rule, which is the operator's call: a TL-owned project is ticked at landing only with `landing_tick: true` in its frontmatter (schema field added to `project.cue`); unset, it follows `project.TLLandingTickDefault` (false), and merge prints the exact `sdlc project tick` command instead. Projects without `tl:` tick by default. A hand-run `project tick` writes regardless (the explicit choice).
- The landed issue is read from the branch's id prefix (#272), not from tracker completions, so a resumed landing re-ticks idempotently.

