# Ledger Landscape — Where State and Evidence Live

## Principle

No single ledger answers all questions. Match the ledger to the question.

State and evidence in ariadne are distributed across many surfaces, each tuned for a specific question. Conflating ledgers (or duplicating fact across them) creates drift; respecting the separation keeps the system inspectable.

## The ledgers

| Ledger | Lives in | What it answers | Audience |
|---|---|---|---|
| Issue body (`## Spec`, `## Plan`) | `workshop/issues/<N>-*.md` (git) | What are we building, and how? | humans + agents reading the issue |
| Issue Log section | same file (git) | What happened during the work, narratively? | humans skimming the issue |
| Plan file (complex case) | `workshop/plans/<N>-*-plan.md` (git) | Detailed implementation breakdown — Core concepts, file structure, bite-sized tasks | execution sessions, milestone reviewers |
| Target file | `workshop/targets/<slug>.md` (git) | What shape do we defend against drift? | humans + agents reading the system |
| Project file | per-repo `workshop/projects/<slug>.md` (git); terminal records archive to `workshop/history/projects/` | Committed baseline, scope events, and retros; `sdlc project status` derives live multi-issue progress from referenced issue records | operator + contributors coordinating the project |
| Calibration ledger | `brain/.../velocity/calibration-ledger.tsv` (git) | **One row per CLOSE, not per issue** (#192 — readers MUST dedupe). How well do estimates match measured actuals, and what did the work cost? (per-issue estimate↔actual + per-model drift, #117/#127; churn / rework / plan-gate round-trips, #187) | velocity calibration; the estimate shell |
| Atlas entries | `atlas/**.md` (git) | How is the system built — architectural map | first-level onboarding |
| Git commits (messages + trailers) | git history (immutable) | What changed, why, and what checkpoint state was crossed? | tooling, history readers, audit |
| Claude transcripts | `~/.claude/projects/<repo-id>/*.jsonl` (local) | What was the AI actually saying that day? | audit, active-time-v3, memory writers |
| Memory files | `~/.claude/projects/.../memory/*.md` (local) | What facts about user / repo / collaboration persist across sessions? | future Claude sessions (auto-loaded via MEMORY.md) |
| Lessons | `workshop/lessons.md` (git) | What patterns went wrong; rules to prevent recurrence | review agents, future sessions |
| Pensive / Parley | `docs/vision/*-pensive-*.md`, `workshop/parley/<chat>.md` (git) | What was the brainstorm before this work crystallized? | humans tracing intent |
| History archive | `workshop/history/` (git) | What we did, archived for the rare lookback | low-signal; avoid unless asked |

## Design principles

1. **Append-only.** Logs accumulate; don't mutate old entries. Git enforces this for commits and committed files; convention enforces it for Log sections. When intent shifts mid-stream, add a `## Revisions` entry (per AGENTS.md §1) — don't rewrite history.
2. **One authoritative source per fact; simple mirrors where helpful.** Drift comes from duplicated mutable state. If a fact lives in two places, designate one as authoritative.
3. **Tooling reads the authoritative; humans read the mirror.** Two surfaces is OK when the mirror is derivable.
4. **Cross-machine durability matters.** Transcripts and memory files live on individual disks — they don't ship in git. For team-shared or operator-portable state, only git-tracked surfaces are reliable.
5. **The right ledger matches the question.** "Was the checkpoint crossed?" wants a structured marker. "What did the reviewer conclude?" wants the semantic final response. Process diagnostics belong to terminal/logging surfaces. Different questions, different ledgers.

## Choosing a ledger — worked examples

**"Was the post-milestone code review conducted, and what was the verdict?"**
- *Authoritative:* git commit trailer on the milestone-close commit (`Review-Verdict: SHIP`). Parseable, immutable, ships in git.
- *Human mirror:* Log line in the issue file (`review verdict: SHIP`).
- *Durable detail (#136/#201):* boundary metadata plus the reviewer's semantic final response is persisted to a git-tracked sidecar in `workshop/plans/` (`NNNNNN-slug-close-review.md` / `-m<x>-review.md`; re-runs append a `## Re-review` section). Its `window` row names the reviewed commit (#194). Harness diagnostics/progress stay on terminal stderr; the separate `*-gate.md` is the machine finding/disposition ledger. Per principle #4 the review sidecar is the reliable human-detail surface; the local agent transcript is the fallback when no sidecar was written (`--no-judge`/dry-run/not-run).
- *NOT* in the project file — that tracks portfolio status, not per-milestone evidence.

**"How many hours did this issue actually take?"**
- *Authoritative:* `actual_hours:` in the issue frontmatter, derived from the in-binary active-time-v3 engine (`sdlc actual` / `sdlc active-time`, `cmd/sdlc/internal/activetime`) over the commit window.
- No mirror needed — frontmatter is already terse.
- *Unit (#118/#92):* the engine measures **ship wall-clock**, not operator-attention — idle gaps still truncate at 15 min, but a subagent-execution span (an `Agent` `tool_use` dispatch → its `tool_result` return, both in the operator's transcript) counts **in full** even when it exceeds the cap. Overlaps collapse only within one transcript source; overlapping sessions remain separate claimable issue work. Activity runs are claimed by nearby issue-referenced commit boundaries, with no-ref commits acting as neutral cut points. This matches the current estimate model's unit (`estimate-logic-v3.1` estimates ship wall-clock directly), so the calibration ledger compares like-for-like.
- *One row per CLOSE, not per issue (#192):* re-closing a done issue is legal, and each
  re-close measures a LONGER cumulative window of the same work — so repeat rows are **partial
  sums**, not repeated observations (`ariadne#167`: 7 rows summing 14.57h for work that measured
  2.71h). **Every reader must dedupe per issue**, keeping the newest row —
  `estimate.NewestPerIssue` is the single implementation. Not doing so inflated the blessed
  throughput baseline 1.41× and every forecast derived from it; the write path is correct and
  the ledger stays append-only, so the dedupe belongs at the read.
- *What counts as a ref (#190):* a **bare** `#N` is local; `<thisrepo>#N` is local; and
  `<otherrepo>#N` is **foreign** — attributable to no local issue, and reported as
  `pair#127 foreign ref ignored` rather than dropped silently. The qualifier names the repo the
  **commits** come from (`--git-repo`), not the process cwd, so measuring a peer reads that
  peer's refs as local. `-` and `.` sit outside the boundary class, so `#174-#176` is two refs,
  not one qualified one. The grammar is single-sourced in `cmd/sdlc/internal/issueref`
  (`parseRef` in `helptext/resolve.md` remains the canonical *validator*). Before this,
  `pair#127` matched as local 127 and charged 46 minutes of #187's work to an unrelated
  archived issue.
- *Known limit (#118):* span matching is **per-transcript-file** — a subagent run whose dispatch and return straddle a session-compaction boundary (dispatch in file A, return in file B) is not paired, so that gap truncates at 15 min. Forward-looking only (all historical spans were within-file and sub-cap); when long delegated runs routinely cross files, aggregate the pending-dispatch map across files in `loadEvents`.

**"Did the plan-quality gate earn its cost on this issue?"** (#187)
- *Authoritative:* the calibration ledger's appended columns — `gate_rounds`, `gate_forced`,
  `gate_addressed`, `gate_withdrawn`, `gate_open`, beside `churn_prod` / `churn_test` /
  `churn_atlas` / `churn_workshop` / `rework`. Twenty columns total; the #187 block is
  indices 10–19.
- *Schema contract:* columns are **APPENDED, never reordered or inserted.** `ParseRows`
  indexes positionally and live ledgers are full of rows written by older binaries, so an
  insertion would not fail — it would silently re-interpret every historical row. Rows
  shorter than the full width parse as legacy: they keep their estimate↔actual data point
  and carry no metrics.
- *Human mirror:* the two lines `sdlc close` prints (see
  [gate-state.md § What close reports](gate-state.md)). Unconditional, unlike the row.
- *Not* a substitute for the gate ledgers themselves (`workshop/plans/NNNNNN-*-plan-gate.md`
  for plan-quality, `NNNNNN-*-close-gate.md` for the boundary review since #194),
  which holds the findings and dispositions these counts summarize.

**"What's the current convention for human-machine markdown markers?"**
- *Authoritative:* `review-convention.md` bundled in the `xx-fix` skill dir (`construct/local/fix/review-convention.md`), reachable in **every** repo via the neutral Agent Skills path `.agents/skills/xx-fix/review-convention.md` (#158). It rides the skill symlink into derivatives, unlike a `workshop/targets/` file which is ariadne-only.
- *Reference:* `AGENTS.base.md §1` points at that path; atlas may too.

**"What was the operator thinking when they proposed this feature?"**
- *Primary:* the pensive or parley file that crystallized into the issue.
- *Secondary:* `## Spec` in the issue (distilled).
- *Audit:* transcript of the brainstorming session.

**"What does this codebase look like architecturally?"**
- *Authoritative:* atlas/. Updated at milestone close per AGENTS.md §8.
- *NOT:* commit messages or Log entries — atlas is the durable map.

**"Why does this convention exist and what trade-offs were considered?"**
- *Primary:* the issue file's `## Spec` section, or the pensive/parley that fed it.
- *Audit:* commit messages along the work (per AGENTS.md §12, commit body explains why).

## Commit trailers — the structured checkpoint ledger

Conventional git trailers (`Key: Value` at the end of a commit message, preceded by a blank line) extend the commit-as-ledger pattern with machine-parseable fields. Already in use: `Co-Authored-By:`. Per-checkpoint additions over time:

- `Review-Verdict:` — milestone-review verdict (SHIP | FIX-THEN-SHIP | REWORK | not-run)
- `Review-Window:` — `<base>..<head>` SHAs the review covered. Both ends are concrete
  commits since #194; the head was previously the literal `HEAD`, so pre-#194 sidecars
  and trailers record `<base>..HEAD` and cannot say which commit was read.
- `Review-Reason:` — when `not-run`, why (e.g., `--no-judge` + reason)

Future trailers may emerge as more checkpoints land. Tooling reads trailers via `git log --grep "Key:"`. Operators rarely need to read trailers directly; they read the Log mirror.

## Related

- [`atlas/workflow/artifact-hierarchy.md`](artifact-hierarchy.md) — narrower view focused on `workshop/` paths and their lifecycle.
- [`AGENTS.md`](../../AGENTS.md) §1 — artifact hierarchy and revisions discipline.
- [`AGENTS.md`](../../AGENTS.md) §5 — verification-before-done; testing threads through stages.
- [`AGENTS.md`](../../AGENTS.md) §8 — atlas / project file maintenance discipline.
- [`AGENTS.md`](../../AGENTS.md) §12 — commit conventions (subject shape, body for why).
- `sdlc --help` — canonical SDLC stage narrative; checkpoint guards.
