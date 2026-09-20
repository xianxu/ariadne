---
id: 000237
status: open
deps: [ariadne#238]
github_issue:
created: 2026-09-18
updated: 2026-09-18
estimate_hours:
---

# prd: pre-claim product spec (journey × interruptions) with a start-plan gate

## Problem

Nothing in the workflow writes down *product behavior* before engineering starts.
The operator's pre-claim framing turns straight into an engineering spec at
`start-plan`, so what the user experiences when things go wrong gets decided
mid-implementation, one interruption at a time.

Specimen, parley#266 (session 3bebf7a9): the operator supplied the product sense
before claim ("user won't remember and will get confused, when a history becomes
a tree"). The agent filed #266 with an engineering invariant as its Spec. Five
product decisions then came up during M1–M2 (edit near a streaming answer, a
failed tool, undo granularity, Stop mid-round, a crashed tool), and the last one
cost a milestone of rework. Replaying the same pre-claim framing through a
journey × interruptions method surfaced all five before claim: 0 of 30 cells
missed, against 8 of 15 missed for #266 as filed
(`brain/probes/prod-lens-replay/spec-stage/`).

Design discussion: `brain/workshop/pensive/2026-09-18-01-pensive-product-lens-user-model.md`.
Umbrella: ariadne#236. This issue is its first part, the pre-claim PRD step.
Depends on ariadne#238 (the atlas split by purpose, which creates
`atlas/journeys/`, the baseline this issue checks against). Order: #238, then
the per-repo migrations, then this issue.

## Spec

**`prd`, a new binary in `cmd/prd/`.** It sits beside `sdlc`, `weave` and
`datatype` and shares their packages (issue parsing, `pkg/vocab`, the judge).
It's a binary, not only a skill, because it owns a gate, and only a binary can
refuse. It isn't subject to the brain spine guard: product discussions often
happen in brain advisor sessions, though the PRD itself lands on the work repo's
issue.

**`prd --help` is the method.** Its first sentence, verbatim: *"This is not your
typical PRD, but a simplified version focused on the user journey and how user
and system errors interrupt it."* That sentence overrides the corporate-PRD
template models default to (personas, market sizing, "the system shall…").
The method:

1. **Main journey:** the typical user's main flow(s) the change touches, step by
   step, in the user's words.
2. **Interruptions:** cross each step with a standard list and say what the user
   sees and can do. The user edits nearby, waits, undoes, stops, regenerates,
   switches away, or quits and reopens; the system fails or never returns,
   crashes mid-effect, is slow, or is busy with another task. Skip rows that
   don't apply.
3. **Settle or escalate each row.** Settle it when the operator's framing, the
   written policies or plain product sense decide it. A real trade-off the
   operator must own (safety versus simplicity, say) becomes a **policy
   question** in user terms, with a recommendation.
4. The operator answers, and the answers are recorded.

**Output: a `## PRD` section on the issue**, containing:
- `### Journey`
- `### Interruptions`: a table of step × interruption → what the user sees
- `### Policies`: question, answer, date
- `### Changes to current behavior`, compared against the journey page in
  `atlas/journeys/`. At close, the issue updates that journey page to the new
  behavior.

**The gate: a `prd` subcommand that checks an issue (name TBD).**
- *Hard (deterministic):* the sections exist, every row states a behavior, and
  no policy question is left open.
- *Soft (judge):* consistent with the current behavior described in
  `atlas/journeys/` (ariadne#238); any contradiction must appear under
  Changes. Also consistent with the operator's written policies
  (the always-on product section, ariadne#236), and with the product principles.
  Flags a user-facing concept atlas doesn't have.
- On pass: commits the issue with a `PRD-Verdict:` trailer plus a hash of the
  `## PRD` section.

**Readiness is derived, not a status.** An issue is "prd ready" when it has a
passing `PRD-Verdict` whose hash matches the current `## PRD` section. The status
enum in `issue.cue` doesn't change: readiness is independent of the lifecycle,
because `prd` and `claim` can happen in either order. Editing the PRD after a
pass makes it not-ready again automatically. Precedent: the quick flow's quoted
contract hashes (`construct/vocabulary/issue.cue:86-94`).

**Enforcement at `sdlc start-plan`**, which becomes a gate for the first time
(today it only prints the ARCH principles). If the PRD is missing, failing or
stale, it refuses and prints the prd flow. `--no-prd "<why>"` is the escape for
work with no user surface (refactors, infra, pure bug fixes); it's announced
loudly and logged, like the other `--no-<gate>` flags. `claim` doesn't change
(#113's cheap lock).

**AGENTS.md** mentions `prd` prominently, next to `sdlc --help`: `prd` manages
the product spec (what the user experiences), and `sdlc` manages software
development (how it gets built).

**Out of scope:**
- The always-on product section (principles plus operator policies): ariadne#236.
- The boundary review checking the diff against the PRD's rows: a later
  follow-up.

## Done when

- `prd --help` prints the method, starting with the sentence above.
- The gate refuses a malformed PRD (missing section, a row without a behavior,
  an open policy question) with an actionable message. On pass it commits with
  a `PRD-Verdict:` trailer and the section hash.
- `sdlc start-plan` refuses when the PRD is missing, failing or stale, and says
  how to run the prd flow. `--no-prd "<why>"` bypasses it and logs the reason.
  A test edits `## PRD` after a pass and asserts that start-plan refuses again.
- AGENTS.md (base layer) names `prd` beside `sdlc`.
- The spec-stage replay, with `prd --help` as the instruction, does at least as
  well as the method-only arm J from run s1: 0 cells missed and at least 8 of 15
  decided correctly.
- Used before claim on one real upcoming issue, and its PRD compared in the Log
  against what the operator actually decides during that issue's implementation.
  This is the held-out test the replay can't provide.

## Plan

- [ ] Design via `start-plan` once claimed. Bootstrapping: this issue predates
  `prd`, so write its PRD by hand using the method (a CLI's interruptions: bad
  input, a stale hash, a missing atlas page, a judge that fails).

## Log

### 2026-09-18

Filed from a brain advisor session. The design was agreed in conversation:
- a binary with `--help` as the method
- a hard + soft gate
- readiness as a hash-matched verdict, not a status
- enforcement at `start-plan`, not `claim`, with a `--no-prd` escape
Evidence and discussion are in the brain pensive and `probes/prod-lens-replay/`
(r1–r4, spec-stage s1).
