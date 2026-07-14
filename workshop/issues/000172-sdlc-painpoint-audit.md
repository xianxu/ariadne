---
id: 000172
status: working
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-14
estimate_hours: 8.07
started: 2026-07-14T08:07:22-07:00
---

# sdlc painpoint audit

Prerequisite for #170. Where does the `sdlc` spine create friction — gates
agents route around, verbs run at the wrong time, workflow moments with no
coverage that keep going wrong? Scope is the **agent-facing** spine: `sdlc`
primarily (weave is not agent-invoked; datatype/vocabulary/doc-review are
non-interactive generators).

## Problem

The `sdlc` binary shapes every workflow transition, but we've never *measured*
where it hurts. Two obstacles:

1. **introspect (#169) will NOT surface this — by design.** Its extract/cluster
   prompts target *taste signals* ("how the **user** wants future sessions to
   behave") — i.e. user→agent friction (redirects, endorsements). Gate-bypasses
   are **agent→binary** friction the user often never sees. Worse, `extract.md`
   explicitly drops "anything specific to one project's domain — we want
   transferable taste," and clustering is activity-scoped (5 buckets, no tooling
   bucket). So substrate ergonomics get filtered out. introspect gives at most a
   faint incidental shadow.
2. **A naive grep is contaminated.** Full-transcript grep for `--force`/`--no-*`
   also catches help-text output, the constitution (which lists every flag), and
   meta-conversations like this one — inflating counts ~10-40×.

So this needs a **direct, precise instrument** — not introspect, not grep.

## Spec

**The instrument already half-exists: `sdlc process-manual --session`.** It
already parses real `sdlc <verb>` at command boundaries (deliberately dropping
prose/help/flags). Extend *that* to capture what it currently discards: bypass
flags, refusal→retry events, and firing-order/load-timing. Go-native,
uncontaminated, matched against the in-process injection catalog.

**Must cover BOTH Claude and codex sdlc usage.** `sdlc` is agent-neutral — it's
invoked from codex sessions too — but `process-manual --session` today reads only
Claude transcripts (`~/.claude/projects`). Measuring bypass/friction from Claude
alone gives a partial, biased picture (e.g. the "bypasses concentrate in peer
repos" finding could differ by agent). So T1's telemetry must ingest codex
transcripts as well. This shares the codex-transcript-reading concern with
**#173** (introspect codex ingest). The two can't share code (introspect is
Python, `process-manual --session` is Go), so the DRY point is the **format spec**:
#173 M3 documented the codex rollout format — `{timestamp,type,payload}` vocabulary,
event→field mapping table, `is_error` derivation, and the **multi-agent fork-replay
trap** (forked rollouts replay the parent transcript and carry two `session_meta`;
key off the FIRST, skip forks, or per-session counts inflate ~66%) — as the single
source of truth in **`atlas/workflow/introspect.md` → "Codex transcript format"**.
T1's Go reader MUST derive from that section (esp. the fork-skip), not re-discover
the format (ARCH-DRY).

**Baseline signal (cleaned — command-field-only grep over ~2,300 transcripts;
order-of-magnitude, not exact):** per-gate bypass frequency —

| flag | uses | read |
|---|---|---|
| `--no-judge` | **83** | most-bypassed by far (review/merge judge) |
| `--no-actual` | 48 | *measured-hours* gate skipped ~48× (tension w/ "measured, not typed") |
| `--no-project` | 47 | mostly legit (single-repo work has no project file) |
| `--no-atlas` | 40 | atlas-update gate skipped |
| `--no-verdict` | 35 | |
| `--no-reclose-guard` / `--no-plan-check` | 14 / 7 | rare |
| `--no-verified` | **0** | verification gate never bypassed — the design works |

Two trustworthy facts under the noise: **`--no-judge` dominates**, and
**`--no-verified` = 0**. Bonus finding: peers-only counts ≈ totals, so
**bypasses concentrate in derivative repos, not ariadne** — the substrate repo
follows its own gates; lighter repos route around them. Hypothesis: gates are
calibrated to ariadne's rigor and feel heavy downstream.

The per-gate `--no-<flag>` design was built precisely to make bypasses
*explicit and measurable*. This issue closes that loop.

## Done when

- **T1** — `sdlc process-manual --session` (or a sibling) emits a clean per-gate
  bypass-rate measure + refusal→retry events + load-timing anomalies, replacing
  the contaminated grep.
- **T2** — the most-bypassed gates (`--no-judge`, `--no-actual`, `--no-atlas`)
  are each triaged: **workflow gap** (gate mis-designed / too costly → fix or
  relax) vs **legit escape hatch** (working as intended), with actions recorded.
  `--no-verified` (=0) confirmed as correctly-calibrated, left alone.
- **T3** — a qualitative coverage-gap read: workflow moments with *no* gate that
  keep going wrong (needs T1 data + judgment).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module     design=0.4   impl=0.24
item: greenfield-go-module     design=0.5   impl=0.32
item: cross-cutting-refactor   design=0.3   impl=0.20
item: greenfield-go-module     design=0.4   impl=0.28
item: greenfield-go-module     design=0.25  impl=0.16
item: greenfield-go-module     design=0.4   impl=0.24
item: greenfield-go-module     design=0.3   impl=0.24
item: pensive                  design=0.5   impl=0.12
item: pensive                  design=0.4   impl=0.10
item: issue-spec               design=1.5   impl=0.12
item: milestone-review         design=0.0   impl=0.12
item: milestone-review         design=0.0   impl=0.12
item: milestone-review         design=0.0   impl=0.12
design-buffer: 0.15
total: 8.07
```

Σdesign 4.95 × 1.15 + Σimpl 2.38 × 1.0 = 8.07.
*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

Item→work: the 3 greenfield-go + 1 cross-cutting-refactor = **M1** (gatesig.go
catalog + drift guard; friction.go `classifyOutputLine` with 3 ACK grammars +
discriminators; session.go `parseEvents`/`classifyToolUse` + output linkage; corpus
walk + `detectGateEvents` + aggregate + render + dispatch). 2 greenfield-go = **M2**
(`detectRefusalRetries`; `detectFiringOrder` per-issue/iteration-aware). 1
greenfield-go = **M3** (codex.go parser, atlas-spec-derived, + cross-language golden).
2 pensive = **M4** (T2 triage findings; T3 coverage-gap read). issue-spec (1.5
design) = the durable plan v1→v3.1 across **three fresh-eyes review rounds** + the
ground-truth signature-catalog enumeration — design is deliberately weighted
mid-range, NOT ×0.2-discounted, correcting the #173 lesson (its 1.73h estimate
under-ran actual 9.04h precisely because interactive design + review-driven additions
were under-weighted). milestone-review ×3 = M1/M2/M3 boundary reviews.

## Plan

Milestones per the durable plan (`workshop/plans/000172-*-plan.md`). T1 (the
instrument) spans M1–M3; T2/T3 (analysis) are M4.

- [x] M1 — signature catalog + `classifyOutputLine` + `SdlcInvocation` + whole-corpus
  per-gate **bypass** measure (claude), anti-contamination + repo labeling. Runs over
  the real corpus; reproduces the headline (T1 partial)
- [x] M2 — refusal→retry + firing-order detectors (T1 complete for claude)
- [x] M3 — codex coverage (net-new Go parser from the atlas spec; both agents)
- [ ] M4 — T2 triage (gap vs escape-hatch) + T3 coverage-gap read

## Log

### 2026-07-14 — M3 built (codex coverage; T1 complete, both agents)
- 2026-07-14: closed M3 — go test ./cmd/sdlc/... green incl codex suites (meta-kind fork/sub-agent/root, parser end-to-end with real-shape fixtures, Failed derivation, cross-language golden vs spec-derived expected.json); both-agent real-corpus smoke: 1558 transcripts / 1833 invocations, exactly 40 fork-replays skipped (matches the atlas spec census), codex 43 / claude 37 bypasses, codex re-close pattern surfaced (no-reclose-guard 25 bypasses, 3/3 refusals resolved via bypass); review verdict: FIX-THEN-SHIP

M3 complete (2 TDD tasks): (8) `codex.go` — `codexMeta` keys off the FIRST
`session_meta` and skips fork-replays (real corpus: **exactly the spec's 40**,
validating skip-40-not-119), keeps sub-agent threads; `parseCodexInvocations`
maps `function_call` (`arguments.cmd`) + `call_id`-linked `function_call_output`
onto the SAME `SdlcInvocation`/`classifyOutputLine` — sdlc's ANSI survives the
exec_command wrapper (verified on real corpus lines). (9) walk wiring — codex
glob + `repoLabelFromPath` (cwd-based, worktree-normalized), per-agent bypass
split + `codex_forks_skipped` in the report, cross-language golden
(`testdata/codex-golden/` + spec-derived `expected.json`). M2-review deferred
items landed: `Failed` (Claude `is_error` / codex non-zero exit) guards the
ladder (REWORK rollback checked first), events classified once
(`allGateEvents`), `forEachRec` shared scan core, skill-late stated Claude-only.

**Both-agent headline (T1 instrument complete):** 1558 transcripts / 1833
invocations; **codex 43 vs claude 37 bypasses**. Codex's signature move is the
**re-close**: no-reclose-guard 25 bypasses (claude 0), and its 3 refusals
resolved 3/3 **via bypass** — the only gate agents route around after refusal.
no-actual refusals 38 corpus-wide, 35 satisfied — refusal texts keep working as
next-action specs. Firing-order unchanged by codex (16 change-code-after-close +
2 skill-late; 52 unattributed publishes). Repo split now pair 37 / brain 19 /
parley.nvim 16 / ariadne 8 — bypasses still concentrate in peers.

**Boundary review: FIX-THEN-SHIP** (sidecar
`workshop/plans/000172-sdlc-painpoint-audit-m3-review.md`). Both Importants
fixed before crossing: (1) walk-level integration test for the two-corpus seam
(`TestRunFrictionReportTwoAgentWalk` + one-sided-corpus contract); (2) the
cross-language golden now has its Python consumer
(`test_normalize.py::test_codex_golden_shared_fixture` — keep/skip decision on
the shared fixtures, downstream-safe skip). Minors folded (see plan Revisions);
post-fix smoke identical (claude 37 / codex 43, forks 40). Both test suites
green (`go test ./cmd/sdlc/...`, `python3 test_normalize.py`).

### 2026-07-14 — M2 built (refusal→retry + firing-order; T1 complete for claude)
- 2026-07-14: closed M2 — go test ./cmd/sdlc/... green incl the new M2 suites (refusal-retry pairing + dedupe, firing-order ladder/attribution/skill-late, render+JSON shape, zero-transcripts error, toolResultText array form); real-corpus smoke: M1 bypass headline reproduced post-dedupe (no-judge 17, no-verified 0, brain 19/pair 15/ariadne 3), new sections render (71 refusals / 70 retried, 18 firing-order anomalies, 37 unattributed publishes); review verdict: FIX-THEN-SHIP

M2 complete (3 TDD tasks, committed separately): (5) `detectRefusalRetries` —
pairs each refusal with the next same-verb+same-issue invocation per transcript;
`resolved` vs `via bypass` separates satisfying a gate from routing around it;
merge/push flag-omitted refusals paired by verb+context, caveat carried. The
M1-review Minor fixed at root: gate events dedupe per (kind, gate, command) per
invocation (the no-validate cwarn + die double-line), and `aggregate` reads the
deduped stream. (6) `detectFiringOrder` — per-(repo, issue) ladder across
transcripts against the AGENTS.md §2 order; only observed INVERSIONS flag
(change-code after clean close/merge); legal loops (mclose→change-code,
start-plan re-runs, REWORK reopen via `judge.ParseVerdict`) silent; merge/push
attributed from segment context or counted unattributed; `skill-late` = plan/TDD
skill load after a non-.md file edit in the same segment+issue
(Edit/Write/MultiEdit → `KindFileEdit`, filtered out of `--session` reports).
(7) report fold: refusal→retry + firing-order sections in markdown + `--json`,
zero-transcripts errors (#68 lesson), go-run/compound limits stated in the
footer, render/JSON + toolResultText-array tests, helptext section.

**Real-corpus M2 headline:** bypass table identical to M1 post-dedupe (instrument
stable). Refusal→retry: 71 refusals, 70 retried, nearly all resolved by
**satisfying** the gate — no-estimate-recon 19/19 satisfied, via-bypass rare
(no-verdict 2, no-atlas 1): the refusal texts work as next-action specs.
Firing-order: change-code-after-close 16 (12 in peers), skill-late 2 (both
brain), 37 unattributed publishes. Deviations recorded in the plan's M2
Revisions entry (inversion semantics for the ladder's first arm; .md edits
excluded from skill-late).

**Boundary review: FIX-THEN-SHIP** (sidecar
`workshop/plans/000172-sdlc-painpoint-audit-m2-review.md`). Important #1 fixed
before crossing: a gate-REFUSED close/merge no longer raises the firing-order
ladder (refused close → change-code is legal recovery); regression test added;
corpus re-run left the anomaly count at 16 (none were the false-positive class —
the fix is semantics-correcting, headline stable). Pairing caveats added to the
report footnote. Deferred to M3 with the scanner work: `tool_result.is_error`
capture, single events computation in `buildFrictionReport`, scan-and-link core
extraction before the codex sibling, codex ActivityMark coverage statement.

### 2026-07-14 — M1 built (instrument runs over the real corpus)
- 2026-07-14: closed M1 — M1 friction instrument built + runs over the real corpus (998 spine invocations / 1006 transcripts); go build + go vet clean, 7 processmanual unit suites green incl the load-bearing anti-contamination test + cross-command drift guard; reproduces the Spec headline from a CLEAN anchored measure — no-judge dominant (17), no-verified=0 (design works), bypasses concentrate in peers (brain 19/pair 15) not ariadne (3); raw grep over-counted ~4x (unlinked echoes correctly excluded); review verdict: FIX-THEN-SHIP

M1 complete (4 TDD tasks, all committed; 998 spine invocations over 1006 non-scratch
transcripts): (1) `GateCatalog` — 12 gates / 16 sigs / 3 ACK grammars + cross-command
drift guard; (2) `classifyOutputLine` — verb-anchored, reset-gated ACKs, grammar+digit
refusals, rejects warmup/source/cat-n; (3) `SdlcInvocation` — anchored `Bash(sdlc)`
calls joined to the tool_result **content-block** (verified more complete than
`toolUseResult.stdout`: no-judge ACK 105× vs 49×); (4) whole-corpus walk + aggregate +
`--friction-report` (markdown/--json), worktree-normalized repo labels.

**Real-corpus headline (clean anchored measure):** no-judge dominant (**17**),
no-actual 8, no-atlas 5; **no-verified = 0** (the design works); bypasses concentrate
in **peers (brain 19, pair 15) not ariadne (3)** — confirming the Spec's hypothesis.
The raw-grep no-judge 64 was ~70% unlinked echoes (process-manual outputs / transcript
reads) the anchoring correctly excludes — the instrument's whole purpose. Known
residual: ~2 of ~19 linked no-judge ACKs not yet classified (edge cases) — a small
accuracy tail for the M1 boundary review / M2 to tighten.

**Deviation:** Edit/Write capture (plan Task 3) deferred to M2, where the firing-order
skill-late detector actually consumes it.

### 2026-07-14 — design phase complete (plan v3.1 build-ready)

Claimed + start-plan; durable plan authored at
`workshop/plans/000172-sdlc-painpoint-audit-plan.md` and hardened through **three
fresh-eyes plan-review rounds** + a ground-truth signature-catalog enumeration.
The instrument: **anchor to `Bash(sdlc <verb>)` invocations**, classify each output
line against a source+corpus-verified **per-gate signature catalog**, whole
cross-repo corpus, both agents. Milestones M1 (catalog + bypass measure) → M2
(refusal→retry + firing-order) → M3 (codex) → M4 (T2 triage + T3 read).

**Two operator scope decisions:** (1) measure **all 12 spine gates** across
close/mclose/change-code/merge/push (not just the 8 close-gates) — the Spec's thesis
is "where does the SPINE hurt"; (2) firing-order oracle from **AGENTS.md's workflow
DAG** (issue.cue lacks the mid-verbs), iteration-aware.

**Review arc (the reviews earned their keep, as on #173):** v1's premise was
inverted — the bypass signal is in the captured **tool-result output**, not a
discarded `.stderr` field (`extractStdout` already reads it); the real problem is
**discrimination, not capture** (this repo develops sdlc, so `close.go` source +
cat-n log reads spray every gate string into tool output — command-anchoring is the
key filter). v2 affirmed the architecture but hand-picked fixtures that didn't match
reality (the "refusal" fixture was actually the `printSemanticWarmup` success line).
v3 rebuilt on the exhaustive catalog; v3.1 folded the final precision fixes.
**Corrected headline: judge bypass ≈ 65 dominant** (`--no-validate` runtime is only
~3 — the naive-grep 87 was source contamination); `--no-verified` = 0 (design works).
Verdict across all rounds: BUILDABLE, no Critical.

**Next (build phase):** `sdlc change-code` (branch + estimate now that scope is
knowable) → implement M1 via TDD (real captured fixtures, incl. the contamination
rejection cases).

### 2026-07-13

Created as a prerequisite to #170 (stack-simplification audit). Baseline
bypass measurement + the "introspect won't cover this" finding captured in Spec
above from a design session over the transcript corpus.
