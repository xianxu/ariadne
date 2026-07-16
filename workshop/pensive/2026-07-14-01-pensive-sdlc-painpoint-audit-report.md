---
type: pensive
date: 2026-07-14
topic: sdlc painpoint audit report (#172) — where the workflow spine actually hurts
mode: thoughts
description: Full report of the #172 friction audit — the measurement instrument, the both-agent per-gate numbers, refusal→retry resolution, firing-order anomalies, T2 per-gate verdicts, T3 coverage gaps, three post-audit follow-on studies, and the five follow-up issues filed (#174–#178).
references: [workshop/history/issues/000172-sdlc-painpoint-audit.md, workshop/history/plans/000172-sdlc-painpoint-audit-plan.md, atlas/workflow/sdlc-binary.md]
---

# Pensive: sdlc painpoint audit report (#172)

**Ticket: [#172 — sdlc painpoint audit](../history/000172-sdlc-painpoint-audit.md)**
(closed 2026-07-14, verdict SHIP, actual 7.42h vs estimate 8.07h). Corpus:
1,566 transcripts / 1,838 sdlc invocations, both agents (claude + codex).
Regenerate the live numbers anytime: `sdlc process-manual --friction-report [--json]`.

## TL;DR

- **The refusal system works.** 158 gate refusals → 158 retried → 149 resolved,
  and only **8** ever resolved by routing *around* the gate. Refusal texts
  function as next-action specs.
- **Bypasses are pre-emptive escape hatches, not dodges** — the flag is passed
  up front with a rationale, which is what the per-gate `--no-<gate>` design
  intended. `--no-verified` was never bypassed once, by either agent.
- **The real problems are three protocol gaps, not gate strictness:** the
  un-specified post-FIX-THEN-SHIP flow (the re-close loop), the no-verdict gate
  punishing recovered single-pass plans, and un-gated off-workflow use (brain
  running the spine against its own charter).
- **The naive grep lied by ~4×**; the command-anchored instrument replaces it.
  Follow-on studies overturned two more intuitions: the "tiny closes" behind
  `--no-atlas` aren't tiny, and reading `--help` doesn't prevent estimate
  refusals.
- **Five follow-up issues filed: #174–#178.**

## 1. Why an instrument, not a grep

The spine's gates were designed to make bypasses *explicit and measurable* —
but nothing ever measured them. Two obstacles: **introspect can't see this by
design** (it extracts user→agent taste; gate bypasses are agent→binary friction
the user never sees), and **a naive grep is contaminated ~4–40×** — this repo
develops sdlc, so `close.go` source reads, `cat -n` log output, help text, and
the constitution itself spray every gate string into transcripts.

The instrument (`sdlc process-manual --friction-report`) fixes discrimination,
not capture: it anchors to real `Bash(sdlc <verb>)` tool calls joined to their
`tool_use_id`-linked output, classifies each output line against a
source-verified **signature catalog** (12 gates / 16 signatures / 3 ACK
grammars, per-gate refusal shapes, warmup traps), and states its
**observability limits** in the output instead of implying complete counts. A
drift guard diffs the catalog against each command's live-registered flags, so
a new gate can't land unmeasured.

Codex coverage derives from the shared format spec
(`atlas/workflow/introspect.md` → "Codex transcript format"); fork-replay
rollouts replay their parent's transcript and are skipped — the live corpus
skipped **exactly the spec's census of 40**. Two corrections were made
mid-audit and are part of the result: the anchored no-judge count is 17 vs a
grep-suggested ~64 (the rest were echoes), and M4 removed a bare-`#N`
attribution fallback after a `git stash -m "… #145"` message forged a false
anomaly (firing-order 16 → 11).

## 2. The numbers

### Per-gate bypasses (command-anchored, contamination-filtered)

| command | gate | bypasses | refusals | observability |
|---|---|---:|---:|---|
| close | no-reclose-guard | **25** | 3 | full |
| close | no-atlas | 12 | 1 | full |
| close | no-judge | 10 | 0 | full |
| milestone-close | no-judge | 10 | 0 | full |
| milestone-close | no-actual | 6 | 10 | full |
| close | no-verdict | 6 | 8 | full |
| milestone-close | no-atlas | 4 | 3 | full |
| close | no-actual | 3 | **38** | full |
| milestone-close | no-project | 3 | 3 | full |
| merge | no-validate | 1 | 4 | full |
| change-code | no-estimate-recon | 0 | **51** | force-only ⚠️ |
| change-code | no-judge (pq/eq) | 0 | 19 | force-only ⚠️ |
| merge | no-judge (publish) | 0 | 6 | flag-omitted ⚠️ |
| close | no-plan-check | 0 | 6 | full |
| change-code | no-structural | 0 | 4 | force-only ⚠️ |
| change-code | no-estimate | 0 | 2 | force-only ⚠️ |

*force-only*: change-code gates skip silently without `--force`, so their
bypass counts are lower bounds. *flag-omitted*: the merge/push publish-gate
refusal never names its flag (best-effort pairing). `no-verified` has no row:
zero bypasses, zero drama — the one gate never routed around.

### Refusal → retry (does a refusal convert?)

Totals: **158 refusals / 158 retried / 149 resolved / 8 via bypass.**

| command | gate | refusals | resolved | via bypass |
|---|---|---:|---:|---:|
| change-code | no-estimate-recon | 51 | 49 | 0 |
| close | no-actual | 38 | 35 | 0 |
| change-code | no-judge (pq/eq) | 19 | 16 | 0 |
| milestone-close | no-actual | 10 | 10 | 0 |
| close | no-verdict | 8 | 7 | **4** |
| merge | no-judge (publish) | 6 | 6 | 0 |
| close | no-plan-check | 6 | 6 | 0 |
| change-code | no-structural | 4 | 4 | 0 |
| merge | no-validate | 4 | 4 | 0 |
| milestone-close | no-atlas | 3 | 3 | 1 |
| milestone-close | no-project | 3 | 3 | 0 |
| close | no-reclose-guard | 3 | 3 | **3** |
| change-code | no-estimate | 2 | 2 | 0 |
| close | no-atlas | 1 | 1 | 0 |

The two bold via-bypass cells are the story: `no-reclose-guard` is the only
gate agents *always* route around after refusal (3/3), and `no-verdict` is the
claude-side equivalent (4/8). Everything else converts by satisfying the gate.
(Pairing is within-transcript and milestone-blind; chains count per-record —
both mildly understate *resolved*.)

### Where and who

- **By repo:** pair 37 · brain 19 · parley.nvim 16 · ariadne 8.
- **By agent:** codex 43 · claude 37 (codex fork-replays skipped: 40).
- **Firing-order:** change-code-after-close 11 (8 in brain), skill-late 2;
  60 merge/pushes had no attributable issue context.

Bypasses concentrate in peers, and codex out-bypasses claude with a signature
move: the **re-close** (25 of its 43 bypasses are `no-reclose-guard`; claude
~0). The Spec's hypothesis — "gates calibrated to ariadne's rigor feel heavy
downstream" — holds only partially: the top drivers are protocol gaps and
off-workflow use, not gate strictness.

## 3. T2 — per-gate verdicts

| gate | verdict | read |
|---|---|---|
| no-judge (close/mclose) | mixed | Legit hatch, but its dominant driver is the re-close loop ("review ALREADY RAN, SHIP, twice") — fix the loop (#174), not the gate. |
| change-code estimate gates | working | #1 friction by volume (76 refusals) yet 0 bypasses — the gate teaches, agents comply. Monitor; don't relax. |
| no-actual | working | Refusals are mostly the designed compute-then-ask two-step — which itself should go (#178). |
| no-atlas | legit hatch | Rationales are the intended ones ("no new surface"). One correctness fix filed (#177). |
| no-verdict | **workflow gap** | Punishes recovered single-pass Mx plans — highest claude route-around rate (#175). |
| no-reclose-guard | **workflow gap** | Guard correct; the post-FIX-THEN-SHIP protocol it collides with is unspecified (#174). |
| no-project / no-validate / no-plan-check | working | Low volume, refusals convert, hatch used as designed. |
| merge/push no-judge (publish gate) | **workflow gap** | All 6 refusals are post-close bookkeeping commits tripping reviewed-HEAD-unchanged (#174). |
| no-verified | calibrated | 0 bypasses in 1,566 transcripts, both agents. Leave alone. |

## 4. T3 — coverage gaps (un-gated moments that recur)

1. **Post-FIX-THEN-SHIP is unspecified → the re-close loop.** The verdict
   finalizes the close, fixes land after, and nothing says what "recording the
   fixes" looks like — so agents re-run close and route around the guard
   (3/3 via-bypass; 25 codex re-closes; live trace on ariadne#145).
2. **Post-close bookkeeping breaks the publish gate.** Log/lessons/plan-tick
   commits move HEAD past the reviewed anchor; merge/push refuse; agents pass
   `--no-judge`. Same family as gap 1.
3. **Working on a `done` issue is un-gated until close.** change-code runs
   silently against done issues; only re-close is guarded. 11 inversions
   observed.
4. **brain runs the spine against its own charter.** A capture repo, not an
   SDLC repo — yet 19 bypasses and 8/11 inversions concentrate there; every
   gate becomes noise, and it pollutes the measurement itself.

**Confirmed non-gaps:** verification (`no-verified` = 0), refusal-text quality
(149/158 convert), plan-first discipline (skill-late ≈ 0 real signal).

## 5. Follow-on studies (post-M4 operator questions)

**Are the `--no-atlas` closes "tiny bookkeeping" windows? No — that framing was
wrong.** Diffstatting the real `Review-Window` of every trailer-bearing close
(71 windows, four repos): 10 of 11 `--no-atlas`'d closes changed **>50 code
lines** (median ~150, up to ~2,000). Their commonality is being **review-fix
re-close windows** with no *new* surface — the re-close loop again, not size.
Consequence: "no code changed → skip atlas" is a correctness fix (#177), not a
volume fix; and "workshop-only → skip judge" was declined — the observed
no-judge deltas aren't workshop-only, and analysis milestones that live
entirely in workshop are exactly what a review should see.

**Would "read `--help` first" prevent the estimate refusals? No.** Of 53
estimate/recon refusal moments: codex had already read `change-code --help`
before **27 of its 33** (and `estimate-source` before 25) — and refused anyway.
Claude read help first only 5/20 yet refuses less. 78 of 98 change-code
transcripts consulted help organically. The refusals are an
**iterate-against-the-validator loop** (write block → exact mismatch → fix
numbers), not a discovery failure. The right levers: fix the start-plan nudge
(it asks for strictly less than the gate demands — "set `estimate_hours:`"
without mentioning the block/grammar/source), and optionally make the block
checkable via `sdlc issue validate` without burning a change-code attempt.

**Why does close refuse actuals it can compute itself? It shouldn't.** The
omit-path already measures and prints "`→ close with: --actual 6.57`" — a
compute-then-ask round-trip that produced ~48 refusals whose suggested value
gets copied verbatim ~45 times. The gate's purpose was preventing *guessed*
hours; a measured value can't be a guess. Filed as #178.

## 6. Actions filed

- **[#174](../issues/000174-post-fixthenship-protocol.md)** — close: specify
  the post-FIX-THEN-SHIP protocol (kill the re-close loop + the bookkeeping
  publish-gate trip). The single highest-leverage fix — it also drains most
  no-judge and no-atlas bypass volume.
- **[#175](../issues/000175-no-verdict-single-pass-recovery.md)** — no-verdict
  gate: accept the issue-close review for trailing unclosed milestones.
  (#172's own close tripped this exact case — live evidence recorded.)
- **[#176](../issues/000176-spine-off-workflow-guards.md)** — spine guards for
  off-workflow invocations: change-code on done issues + non-SDLC repos (brain).
- **[#177](../issues/000177-atlas-gate-no-code-autoskip.md)** — atlas gate:
  auto-satisfy on no-code windows (correctness fix; volume expectation ~1/11).
- **[#178](../issues/000178-close-adopts-measured-actual.md)** — close adopts
  its measured actual (deletes the second-largest refusal volume at zero
  calibration risk).

Recommended but not filed: the start-plan estimate-nudge wording fix;
estimate-block validation in `sdlc issue validate`; re-measure after #174 lands
and revisit any judge relaxation only with that evidence.

## Open questions

- Does the estimate-recon refusal volume (51) persist after the start-plan
  nudge is fixed, or was much of it the nudge under-asking? Re-measure.
- pair's 37 bypasses are only partially explained (re-close loop + legit
  hatches); is there a pair-specific pattern worth its own read?
- After #174 lands, do the no-judge/no-atlas bypass counts drop as predicted?
  The friction report is the regression test for the workflow itself.

## References

- Ticket: `workshop/history/issues/000172-sdlc-painpoint-audit.md` (Findings = T2/T3
  core; Log = per-milestone evidence; archived at ship).
- Durable plan + 4 review sidecars: `workshop/history/plans/000172-*`.
- Atlas map: `atlas/workflow/sdlc-binary.md` → "Friction audit".
- Instrument source: `cmd/sdlc/internal/processmanual/` (gatesig.go catalog,
  friction.go detectors, codex.go parser, testdata/codex-golden/).
