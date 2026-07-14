---
id: 000172
status: working
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-14
estimate_hours:
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

## Plan

- [ ] T1 — extend `process-manual --session` to capture bypass flags,
  refusal→retry, and firing-order/load-timing; emit a per-gate friction report
- [ ] T2 — triage top-bypassed gates (gap vs escape-hatch) using T1's clean data
- [ ] T3 — coverage-gap read for un-gated workflow moments that recur as errors

## Log

### 2026-07-13

Created as a prerequisite to #170 (stack-simplification audit). Baseline
bypass measurement + the "introspect won't cover this" finding captured in Spec
above from a design session over the transcript corpus.
