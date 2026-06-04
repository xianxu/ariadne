---
id: 000082
status: working
deps: [000080]
github_issue:
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 3.5
---

# smoother cross-repo base-layer workflow

## Problem

ariadne is the **base layer**: derivative repos (nous, pair, …) symlink its
substrate (per `construct/setup.sh`), so a change to ariadne's working tree is
*live* in every derivative immediately. In this stage of development the base
layer is high-churn — it's still being extended — which the symlink model is
designed to support (iterate the base while developing on it).

The cost is **concurrency the operator can't easily reason about**. At the leaf,
conflicts are obvious ("am I editing the same thing in two pair sessions?"). At
the base, reasoning breaks down: it's easy to forget that work in nous *and* pair
will collide *if both touch ariadne*. The shared-live working tree means one
session's uncommitted/branched base state is invisibly read by every other
session. Two patterns make it worse:

1. **Base changes are driven from any derivative** (the AGENTS.md "peers" model —
   an agent in nous can just go make the ariadne change). Convenient, but the
   base work is initiated from a context that doesn't "own" ariadne, so it's easy
   to lose track of who's touching the base.
2. **Even *filing* a base issue (not working it) collides with sdlc gates.** A
   new `workshop/issues/*.md` is untracked working-tree residue; with #78/#80 it
   no longer blocks a merge by accident, but the broader friction remains — base
   issues are discovered mid/late in a *derivative* session, and capturing one
   shouldn't entangle with the base repo's working-tree state at all.

This issue makes the base-layer workflow smoother **without adding a gate to the
common path** — drafting and brainstorming stay free; the one new check fires
only where commitment happens.

## Spec

### Reframe — three layers, different physics

The pain comes from conflating three things that should behave differently:

- **Tracker state** (issues, claims, status) — append-only, instantly shared,
  committed to main *out-of-band*; should never be working-tree residue.
- **Base-layer code** (`construct/`, `cmd/`) — shared *live* via symlinks;
  high-churn; the real contention surface.
- **Leaf code** (derivative-specific) — naturally isolated per session; fine.

Leaf work isn't the problem. The fixes target (a) tracker state leaking into
working-tree guards, and (b) shared *uncommitted/branched* base being invisible.

### Unifying principle

The symlink model implies **the base should behave like a continuously-integrated
trunk**: every derivative reads ariadne's working tree live, so the longer
ariadne sits on a branch, the longer *every* derivative reads in-flux base.
High base-churn is fine — *long-lived, concurrent, invisible* base branches are
what break reasoning. So: keep base-layer changes small, fast, commit-to-main-
frequently, and make any in-flux base state *visible*. Heavy isolation (a
worktree-set; see Non-goals) is the rare escape hatch, not the default.

### The three moves (→ milestones)

1. **Tracker ops never touch the working tree.** `sdlc issue new` auto-syncs the
   new issue file to origin/main immediately, reusing `claim`'s `syncOnMain` with
   the `--issue` *filter* (precise add — dodges the #80 broad-add). Filing a base
   issue from anywhere, on any branch, lands it on main and leaves the working
   tree clean; peers see it instantly. Enrichment that follows (brainstorm →
   edit) is flushed by `claim` (which already syncs) — and by move 2 it never
   blocks anything in between.

2. **Dirty-tree guards ignore tracker files, full stop.** An uncommitted
   `workshop/issues/*.md` is tracker state, not code — it should never block a
   dirty-tree gate (merge step-2 / step-9b, and any sibling). Same form-vs-essence
   line as #78/#80, generalized from "untracked" to "tracker-file class." Makes
   drafting/enriching frictionless without timing discipline.

3. **`start-plan` reads base contention (non-blocking).** `start-plan` is the
   "we're getting serious, here are the principles" gate (re-run per design) — the
   natural place to ask "who else is serious in the base right now?" `claim`
   already *broadcasts* the "I'm in" half to origin/main; `start-plan` reads the
   other half. Surfaces, in one line, never refuses: base repo not on clean `main`
   (branched / dirty *code* — NOT issue files), plus other `status: working` base
   issues on origin/main that aren't this one. E.g. *"base (ariadne): on branch
   `foo`, 2 other issues in-flight (#a, #b) — planning against a moving base."*

Why `start-plan` and not session-start: base changes are discovered mid/late in a
derivative session, so session-start has no signal yet; `start-plan` is the
commitment point where you'd actually want to look before you leap.

## Non-goals

- **Layout-preserving worktree-set** (the deliberate base-isolation escape hatch:
  `git worktree add` each repo into a parallel sibling tree so relative symlinks
  resolve without re-running setup.sh). Real and useful for the *rare* change that
  needs ariadne isolated while other base work continues — but a separate, larger
  concern. File separately if/when the "I need two base branches live at once"
  wall is hit. This issue is about making the *common* path smooth.
- Changing the symlink model, the peers model, or the "drive base from any
  derivative" habit — those are kept; this issue makes them *safe + visible*, not
  removed.

## Done when

- M1: `sdlc issue new` leaves the working tree clean — the new issue is on
  origin/main immediately (filtered sync, no unrelated files swept). Filing a base
  issue from a feature branch / a derivative session no longer creates residue.
- M2: an uncommitted/untracked `workshop/issues/*.md` never blocks `sdlc merge`
  (or any dirty-tree gate); a dirty *code* file still does. Regression test pins
  both directions.
- M3: `sdlc start-plan` prints a one-line base-contention summary (branch/dirty-
  code state + other in-flight base issues), non-blocking; clean base → a green
  line. Test pins the summary logic (pure, over fake git/tracker state).
- `go test ./cmd/sdlc/...` green; atlas (`sdlc-binary.md` / `base-layer.md`)
  updated for the new tracker-sync + start-plan behaviors.

## Plan

Multi-milestone (three independent review boundaries; each closes separately).
Depends on **#80** (the merge-archive broad-add fix) — land that first so M2's
guard work and the filtered-add pattern sit on clean guard hygiene.

- [x] **M1 — `issue new` filtered auto-sync.** Reuse claim's sync, don't fork it.
      - **ARCH-DRY:** extract `runClaim`'s branch-aware dispatch (the part after the
        optional `startOnClaim` flip) into `syncIssuesToMain(stdout, stderr,
        *claimFlags, r gitRunner)` — `branch=="main" ? syncOnMain : syncOnBranch`.
        Thread the `gitRunner` (don't drop it) so tests inject a capture runner;
        `runClaim` calls it with `claimRunner`. One source of truth for "broadcast
        issue files to origin/main".
      - In `runIssueNew`, after the file is written (skip on `--dry-run`), parse
        `nextID`→int and call `syncIssuesToMain` with `claimFlags{Issue: id,
        IssuesDir: f.IssuesDir, NoStart: true}`. The `--issue` filter rides #80's
        filtered add (`git add <matches>`), so the new file lands on origin/main
        and unrelated untracked files stay put.
      - On `main` (dominant path): file committed + pushed → working tree clean.
        On a feature branch: inherits claim's `syncOnBranch` (sync via the main
        worktree); a local untracked copy may remain — **M2 makes that
        non-blocking**, so M1+M2 together deliver "filing never entangles a gate."
        (Same worktree-topology constraints as `sdlc claim` today; not new here.)
      - Test: `issue new` on main → file present + committed (clean tree); an
        unrelated untracked issue file is untouched; `--dry-run` does NOT sync.
- [ ] **M2 — guards ignore tracker files.** Dedicated non-blocking bucket.
      - **ARCH-DRY + ARCH-PURE:** give `dirtyAssessment` a third `Tracker []string`
        bucket; `assessDirty(porcelain, issuesDir, historyDir)` classifies any
        `workshop/issues|history/NNNNNN-*.md` line (tracked *or* untracked) as
        Tracker, never Blocking. Reuse push.go's `parsePorcelainStatus` +
        `isIssuePath`/`isHistoryPath` to pull the path + match — no new glob logic.
        `Refuse()` stays `len(Blocking) > 0` (the elevated #78 invariant intact).
      - Update the two call sites (step-2, step-9b) to pass the dirs and surface
        `Tracker` like `Untracked` (warn, don't refuse). History included for
        symmetry (both are tracker state synced out-of-band).
      - Test (pure, both directions): dirty *issue* file → `!Refuse()`; dirty
        *code* file → `Refuse()`; mixed → refuses (code still blocks).
- [ ] **M3 — `start-plan` base-contention read.** Pure summarizer + thin IO gather.
      - **Repo-selection decision (the M3 crux, pinned):** the contention read
        inspects **cwd**, and the line is emitted **only when cwd is the base
        repo** — explicit `cwd==base` assumption ("you plan a base change while
        cd'd into ariadne"). Detect via a small pure-ish predicate `isBaseRepo` =
        `construct/` exists as a *real directory, not a symlink* (in derivatives
        `construct/` is symlinked to the base). In a derivative cwd → stay silent
        (leaf planning needs no base-contention line). Cross-repo base resolution
        (read the base's state from a derivative cwd) is a noted follow-up, scoped
        OUT here alongside the worktree-set escape hatch.
      - **ARCH-PURE:** `baseContentionSummary(in baseContention) string` — one
        non-blocking line over `{repo, branch, dirtyCode []string, others
        []issueRef}`. Clean main + no others → a green "clear to plan" line;
        branched/dirty-code or N others → the heads-up. No IO, table-tested.
      - `runStartPlan` gathers inputs (thin seam, only if `isBaseRepo`): repo name
        = basename(cwd); branch via `gitx.Capture`; dirty-code via
        `worktreeDirty`+`assessDirty` (reusing M2's tracker split so issue-file
        dirt doesn't read as base contention — only `Blocking` code counts); other
        in-flight via `listIssues` filtered `status==working && id!=this` (local
        main view; claim keeps it fresh). Prints the line after the architecture
        block, never refuses.
      - Test: pure `baseContentionSummary` fixtures — clean base, branched base,
        dirty code, N concurrent working issues (excludes self); plus a direct
        test of the `isBaseRepo` predicate (real dir vs symlink).

## Log

### 2026-06-04
- 2026-06-04: closed M1 — go test ./cmd/sdlc/... green; new TestRunIssueNew_AutoSyncsToMainCleanTree (file on main, clean tree, unrelated untracked untouched) + TestRunIssueNew_AutoSyncBestEffort (no-origin → file created, warns, no os.Exit); existing claim/fetch tests pass; go vet clean.; review verdict: SHIP
Spun out of a design conversation on cross-repo base-layer friction (the #79→#80
incident was the trigger: an untracked issue file swept onto main by a broad
`git add`). Resolved the framing: three-layer reframe (tracker / shared-base /
leaf), base-as-trunk unifying principle, and `start-plan` as the contention-check
gate (chosen over session-start because base changes surface mid/late in
derivative sessions). The layout-preserving worktree-set escape hatch was
explicitly scoped OUT (Non-goals) as a separate, larger concern. Depends on #80.
See [[000078]] (untracked-tolerant guard) and [[000080]] (broad-add fix) — M2
generalizes their form-vs-essence line to the tracker-file class.

### 2026-06-04 (M1)
Extracted `runClaim`'s branch-aware dispatch into `syncIssuesToMain` (ARCH-DRY);
`issue new` calls it post-scaffold with the new issue's `--issue` filter, routing
the sync's stdout to stderr so the path-on-stdout contract stays clean.
**Root-cause scope expansion:** the plan-quality gate plus a real test failure
(`TestRunFetch_CreatesFileWithNextID` — `fetch` delegates to `runIssueNew`, and
its repo has an unreachable github origin) exposed that `syncOnMain`/`syncOnBranch`
called `die()` (os.Exit) internally — so a failed push would kill `issue new`
itself, and my best-effort error-handling was dead code. Fixed at the root:
converted both sync helpers to **return errors** instead of die(); `claim` dies on
them (UX unchanged), `issue new` warns (best-effort — the file is already
written; offline/no-origin must not abort creation). Tests: on-main clean-tree +
filtered-add (unrelated untracked untouched), and best-effort degradation
(no-origin repo → file created, warns, returns nil). atlas/issue-sync.md updated.
