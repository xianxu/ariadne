---
id: 000179
status: done
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours: 1.16
started: 2026-07-15T11:49:25-07:00
actual_hours: 0.24
---

# sdlc migrate: move a markdown artifact across peer repos with ref rewrite

## Problem

#171's direction (peer-repo `repo#id` addressing; artifact residency = soft
center-of-gravity default) makes moving a markdown artifact between repos a
NORMAL operation — a project file follows the work, an SDLC artifact leaves
brain. But a move today is a hand job with a silent correctness trap: **bare
`#NNN` refs inside the file are repo-relative.** Moved verbatim, they
re-resolve against the destination repo's issue numbering — pointing at
unrelated issues without any error. The rewrite rules are fixed patterns (the
formal ref grammar `sdlc resolve` already owns), so this belongs in the
binary, not in agent judgment.

## Spec

`sdlc migrate <path-to-file> <dest-repo-dir>` (e.g.
`sdlc migrate data/project/metis-v2.md ../kbench`) — deterministic, no LLM:

1. **Rewrite outbound refs.** Bare `#NNN` in the file body is source-relative
   → qualify to `<source-repo>#NNN` so it resolves identically from the
   destination. Refs already qualified with the DESTINATION repo
   (`<dest>#MMM`) may normalize to bare `#MMM` (they become local). Qualified
   refs to third repos pass through untouched. Skip fenced code blocks (the
   #66 meta-document lesson: docs about the ref grammar quote it literally).
2. **Verify before writing.** Every ref the rewrite touches must resolve —
   `sdlc resolve` machinery against the live fleet; a dangling ref aborts the
   migration with the offending ref named (fail-closed, nothing half-moved).
3. **Move + commit mechanics.** Write the rewritten file at the destination
   (same relative path by default, `--dest-path` to override), remove the
   source, and commit BOTH sides scoped to exactly the files touched
   (`git add -- <file>`, per the lessons.md broad-add rules) — or
   `--no-commit` to stage only. Refuse on a dirty destination (the
   propagate-base Rule-4 posture).
4. **Report inbound refs, don't rewrite them (v1).** Other files across the
   fleet pointing AT the moved artifact (by path, or `<source>#NNN` when the
   artifact is an issue) keep working only if the ref grammar is
   location-independent — issue refs are; PATH references are not. v1: scan
   the fleet (super-repo grep over the fixed patterns) and PRINT the inbound
   sites for the operator/agent to judge; no automatic fleet-wide rewrite.

Design questions (settle at plan time):
- **Issue artifacts renumber.** Issue IDs are per-repo sequences; migrating
  an issue file needs a fresh dest-side ID (allocate via the `issue new`
  counter) + a tombstone/redirect line in the source repo's Log or history.
  Possibly v2 — the #171 driver (project files, slug-named) needs no
  renumbering.
- Whether `migrate` refuses in a brain repo (it's a lifecycle-ish verb;
  probably yes via the #176 guard set — spine `WorkflowVerbs()` membership).
- Frontmatter touch-ups: bump `updated:`; anything else per datatype.

Related: supports #171 (not a blocking dep — the project-gate lift can land
first; migrate makes the residency default cheap to change).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.5   impl=0.3
item: atlas-docs             design=0.05  impl=0.08
item: milestone-review       design=0.0   impl=0.15
design-buffer: 0.15
total: 1.16
```

Σdesign 0.55 × 1.15 + Σimpl 0.53 × 1.0 = 1.16. greenfield-go-module = the new
migrate verb (pure rewriter + segmenter + IO shell + fixture e2e — single
concern, impl at 40% of the 0.75 mid-range); atlas-docs = helptext/migrate.md
+ atlas paragraph; milestone-review = the close-time boundary review. +15%
buffer: thorough plan doc (fresh-eyes reviewed, 6 findings folded in).
Design hours are NOT ×0.2-discounted despite the plan doc existing: the plan
was authored inside this issue's active window (claim-early #113 — claimed
today, plan written after), so `sdlc actual` measures its authoring time and
the estimate must carry it; the v2 Step-3 discount's rationale ("design
already happened before the window") doesn't apply. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Done when

- `sdlc migrate <file> <dest-repo>` moves the file with all outbound bare
  refs correctly qualified, dest-local refs normalized, fenced blocks
  untouched, and every touched ref verified to resolve — or refuses naming
  the dangling ref.
- Both repos end with scoped commits (or staged with `--no-commit`); dirty
  destination refuses.
- Inbound references across the fleet are reported with file:line.
- A round-trip (migrate there, migrate back) is ref-stable — no rewrite
  churn.

## Plan

Durable design: `workshop/plans/000179-sdlc-migrate-plan.md` (fresh-eyes
reviewed; 6 findings folded: span-parses-as-single-ref rule, dest-vantage
verification, canonicalization+idempotence round-trip, SplitFences as new
segmenter not a stripCodeFences refactor, same-repo/outside-repo guards,
parseRef candidate filter). Single-pass, plain checkboxes.

- [x] design at start-plan: ref-rewrite rules as a pure entity over the
      existing ref grammar (ARCH-DRY with `sdlc resolve`), thin IO for
      move/commit; fixture repos for the round-trip test
- [x] `issue.SplitFences` segmenter + tests
- [x] `rewriteRefs` pure core (scanner + parseRef filter + span rule + 3 rewrite rules) + grammar drift test
- [x] `runMigrate` IO shell: guards, dest-vantage verification, scoped two-repo commits, registration + helptext
- [x] inbound-ref report + round-trip canonicalization/idempotence e2e + mutation check
- [x] atlas + bookkeeping

## Log

### 2026-07-15
- 2026-07-15: closed — go test ./cmd/sdlc/... 11/11 green: 20-case rewriter table + grammar drift test, 10 e2e scenarios on real two-repo fixtures (guards, fail-closed verification, scoped commits, --no-commit), round-trip canonicalization+idempotence, bare-ref mutation check reddens 12 tests; live dogfood on scratch fixture passed end-to-end and caught+fixed the symlinked-cwd guard misfire; review verdict: FIX-THEN-SHIP

Filed from the #171 brainstorm (operator): moves become normal under
peer-repo addressing, rules are fixed patterns → binary-owned. Bare-ref
requalification + existence verification are the core; inbound refs are
report-only in v1.

Implementation (all TDD, red→green per task): pure core landed exactly per
plan (ARCH-PURE: rewriteRefs/SplitFences string-in string-out, zero IO in
their tests; ARCH-DRY: parseRef stays the sole grammar authority via the
candidate filter + drift test; stripCodeFences deliberately NOT refactored —
different unterminated-fence policy, cross-referenced comments). Mutation
check: neutering the bare-ref rule reddens 12 tests. Live dogfood on a
scratch two-repo fixture found a REAL bug the hermetic suite missed: under a
symlinked cwd (macOS /tmp), os.Getwd's logical-$PWD preference vs git's
resolved paths made the inside-repo guard misfire — fixed with EvalSymlinks
on both sides + a $PWD-setting regression test (the lessons.md "IO needs a
live run" pattern, again).
