---
type: continuation
slug: sdlc-painpoint-audit-build
agent: claude
created: 2026-07-14T11:29:16
branch: main
worktree: /Users/xianxu/workspace/ariadne
issues: [000172, 000170]
---

# Continuation: sdlc-painpoint-audit build phase

## NEXT ACTION

Build **#172** (sdlc painpoint audit). **M1 COMPLETE** (all 4 tasks committed, on
branch `000172-sdlc-painpoint-audit`; instrument runs over the real corpus and
reproduces the headline). **Next: `sdlc milestone-close --issue 172 --milestone M1`**
(dispatches the fresh-eyes boundary review; fix Critical/Important, then cross). Then
M2 → M3 → M4.

- M1 commits: `a6c3559` (catalog+drift guard) · `113ee31` (classifyOutputLine) ·
  `77875c5` (SdlcInvocation) · `97c7b2c` (corpus walk + `--friction-report`).
- Run it: `go run ./cmd/sdlc process-manual --friction-report`. Headline: no-judge 17
  dominant, no-verified 0, peers (brain 19/pair 15) ≫ ariadne 3.
- M1 residual for the boundary review to note: ~2 of ~19 linked no-judge ACKs
  unclassified (classifier edge cases); repo-label slug→repo is lossy (best-effort).

**M2 (next milestone):** `detectRefusalRetries` (pair a Refusal invocation with the
next same-verb+issue invocation; per-gate refusal sigs already in the catalog) +
`detectFiringOrder` (per-issue, iteration-aware: mclose→change-code / start-plan
re-runs / rework legal; merge/push have no --issue → attribute from segment context).
**Edit/Write→KindFileEdit capture (deferred from M1 Task 3) lands here** for the
skill-late arm. Then M3 (codex parser from the atlas spec) → M4 (T2/T3 analysis).

---
[superseded — M1 Tasks 1-2 detail retained below for reference]

**M1 Tasks 1-2 committed** (the classification core):
- Task 1 (`a6c3559`): `GateCatalog` (16 sigs / 12 gates, 3 ACK grammars) +
  cross-command drift guard (`cmd/sdlc/gates_test.go` introspects each spine
  command's registered `--no-*` flags vs the catalog).
- Task 2 (`113ee31`): `classifyOutputLine` — anchors on verb, requires the runtime
  `\x1b[0m` reset for a bypass ACK, grammar+digit-anchored refusals keyed on the exact
  per-gate tail; rejects warmup/source/cat-n contamination. `GateEvent` +
  observability (force-only / flag-omitted). Tests use REAL captured fixtures.

**Remaining M1** (in `cmd/sdlc/internal/processmanual/`, TDD, real fixtures):
3. **`SdlcInvocation`** — build from anchored `Bash(sdlc <verb>)` calls joined to
   their `tool_use_id`-linked result output; add `Edit`/`Write`/`MultiEdit` →
   `KindFileEdit` in `classifyToolUse` (`session.go:393`); **extend `parseEvents` to
   attach + RETAIN the raw output** on every `KindSDLCPrompt` event (today it links
   stdout only for close/mclose verdict recovery at `session.go:216` and DISCARDS the
   raw text — the plan-quality/estimate-quality judges both confirmed this). Parse
   `issueID` from `--issue N`/`#N` in args.
4. **Whole-corpus walk** — `enumerateAllTranscripts()` (globs all
   `~/.claude/projects/*/*.jsonl`; **take injectable roots** per the plan-quality
   advisory so the temp-dir test works) + `detectGateEvents` + `aggregate` +
   `renderFrictionReport` (markdown + `--json`) + `--friction-report` dispatch in
   `processmanual.go` (mutually exclusive with `--session`, not repo-bound). Repo
   label: normalize `-worktree-ariadne-` → `ariadne`, exclude `-private-tmp-`/
   `-private-var-folders-`. **Anti-contamination test is load-bearing.**
Then `sdlc milestone-close --issue 172 --milestone M1`, then M2 (refusal→retry +
firing-order) / M3 (codex) / M4 (T2+T3).

Two judge advisories to honor (both non-blocking, noted at change-code): drift guard
keyed on distinct **(command, flag)** pairs (done); `enumerateAllTranscripts` takes
**injectable roots** (Task 4).

## State of play

Run `sdlc state`. In flight:
- **#172** (this) — status `working`, claimed 2026-07-14. Plan v3.1 build-ready. No
  code yet, no branch yet. Prereq of #170.
- **#170** — open. The audit umbrella (*simplify the ariadne stack*); `deps:[#169,
  #172]`. #172 is its last open dep. After #172 → #170.
- **#173** — **done + merged** (PR #91). introspect ingests codex end-to-end; the M3
  finding (codex does NOT reopen the taste well — same diminishing returns as #169)
  is in `atlas/workflow/introspect.md`. The **codex-format spec** it wrote (same
  atlas file → "Codex transcript format") is #172 M3's shared source for the Go codex
  parser (Python & Go can't share code, only the spec).
- **#169** — done (introspect run-3, diminishing returns).

## The #172 instrument in one breath

`sdlc process-manual --friction-report` — whole-corpus, cross-repo, both-agent. The
bypass/refusal signal is **already captured** (sdlc's stderr folds into the Bash
tool's stdout = the `tool_result` block, `tool_use_id`-linked; `extractStdout` reads
it). The problem is **discrimination, not capture**: this repo develops sdlc, so
`close.go` source + cat-n log reads spray every `--no-X` string into tool output. So:
**(1) anchor to `Bash(sdlc <verb>)` invocations** (drops the source/edit/log noise),
**(2) classify each output line against a per-gate signature catalog** (the master
table in the plan — 3 ACK grammars, per-gate refusal shapes, warmup traps), **(3)
state observability limits honestly**. Detectors: bypass tally, refusal→retry
(per-issue), firing-order (per-issue, iteration-aware, AGENTS.md workflow DAG).

## Thread arc & user model

This session: resumed the #173 codex-ingest continuation → finished #173 M3 (the
codex dogfood + the "does codex reopen the well?" finding = **no**, ~95% of the
apparent surplus was fork-replay duplication + benign-exit friction; fixed 3 codex
adapter gaps; 2 boundary reviews, all Important fixed; merged). User then steered:
*"the next step in this journey is #172."* Pivoted to #172: claimed → start-plan →
designed through **three plan-review rounds** (v1 inverted premise → v2 architecture
affirmed but wrong signatures → v3 on ground-truth catalog → v3.1 precision fixes).

User model (per AGENTS.md *Model User Intention*): systems-thinker, terse decisive
steers, engages deeply with structured tradeoff tables (answered every
AskUserQuestion crisply). Prizes **agent-neutrality** (a core value), **precision
over recall**, **honest calibration** (accepted #173's 5.2× estimate overrun as
data), and **review rigor** — explicitly chose to re-verify the plan twice rather
than rush to code, and the reviews repeatedly earned their keep (each round caught
real, empirically-grounded errors). Drives **depth-first** to completion. When a
design premise proved wrong, wanted it fixed at root (rebuild on ground truth), not
patched.

## Open questions

Both #172 scope forks are **decided** (recorded in the plan):
- Gate scope → **all 12 spine gates** (close/mclose/change-code/merge/push).
- Firing-order oracle → **AGENTS.md workflow DAG**, iteration-aware.
Corpus scope → **whole cross-repo** (`enumerateAllTranscripts`).
Run order after #172 → **#170** (the audit umbrella; #172 is its last dep).

## Artifact map

Read-first (AGENTS.md is auto-loaded via CLAUDE.md; issues/plans are NOT):
- **`workshop/plans/000172-sdlc-painpoint-audit-plan.md`** — the build-ready plan.
  The **"signal catalog" master table** is the load-bearing artifact (the classifier's
  data). Read the `## Revisions` v1→v3.1 entries for *why* each shape is what it is —
  they encode the three review rounds' findings so you don't repeat them.
- **`workshop/issues/000172-sdlc-painpoint-audit.md`** — Spec (the T1-T3 Done-when +
  the contaminated-grep baseline the clean measure replaces) + the design-phase Log.
- **`atlas/workflow/introspect.md` → "Codex transcript format"** — the shared spec the
  M3 Go codex parser derives from (fork-skip-40-not-119; is_error derivation).
- **Code the plan touches:** `cmd/sdlc/internal/processmanual/session.go`
  (`classifyToolUse` ~:393, `parseEvents` tool_use_id linkage ~:196/216 — note it
  parses the verdict then DISCARDS the raw output; M1 Task 3 must retain it),
  `processmanual.go` (dispatch ~:81), `close.go`/`milestoneclose.go`/`changecode.go`/
  `merge.go`/`push.go` (the gate flag registrations + cwarn/cinfo/explainer emit
  sites — all cited with file:line in the plan's master table), `term.go` (ANSI
  prefixes ~:34-40), `internal/transcripts/codex.go` (`codexCWDFromBytes` ~:69 —
  cwd-only + unexported, so M3 extracts fork fields net-new).

## Decisions & dead ends

- **The signal is in tool-result OUTPUT, not stderr** (v1 dead end). `extractStdout`
  already reads it; the work is classification.
- **Anchor to `Bash(sdlc <verb>)`** is the highest-leverage contamination filter —
  drops the source/edit/log-read noise that dwarfs real ACKs in this repo.
- **The warmup trap** (`printSemanticWarmup`, close.go:219, `…only if there's
  genuinely nothing`) fires on SUCCESS and is byte-adjacent to the real no-actual
  refusal (`…only when measurement is not applicable`) — match the exact per-gate
  TAIL, never the shared prefix.
- **Runtime discriminator is ACK-only**: cwarn/cinfo ACKs carry `\x1b[0m ` before the
  message; refusals (plain `\n`-joined / `die()`) do NOT — refusals discriminate by
  grammar-anchored `refusalRE` (`\d+` not `N`/`…` placeholders).
- **Silent bypasses**: change-code's 4 gates used alone emit no ACK (observable only
  via `--force`) — report labels them `force-only`, doesn't fake a complete count.
- **merge/push carry no `--issue`** — stage-5 firing-order attributed from segment
  context or `unattributed`.
- **Corrected headline**: judge ≈ 65 dominant; `--no-validate` runtime ~3 (the grep-87
  was source contamination); `--no-verified` = 0.

## Lessons learned

- **Fresh-eyes plan review, verified against real data, is worth 3 rounds on a
  measurement instrument** — where wrong signatures = wrong findings. Round 1 caught an
  inverted premise (stderr vs stdout) that would've built a dead instrument; rounds
  2-3 caught fabricated fixtures + coverage gaps. Each reviewer VERIFIED against the
  corpus rather than trusting prose. Build M1 from REAL captured fixtures.
- **Don't reason about intricate wire-format details — enumerate them.** The first
  exploration inferred "stderr carries the signal" from the code (cwarn writes to
  stderr) but didn't check what lands in the transcript (Bash folds stderr→stdout).
  The ground-truth catalog subagent (source + corpus counts) is what made v3 solid.
- **A grep baseline is an upper bound, not a measurement** — the Spec's own `--no-validate`
  87 was ~95% source contamination (runtime 3). The whole point of #172 is replacing
  that with a command-anchored measure.
