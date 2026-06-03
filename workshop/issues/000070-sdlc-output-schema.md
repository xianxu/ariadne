---
id: 000070
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours:
---

# schematize sdlc judge/verb output: a shared schema file (first-line verdict) both prompts and the parser reference

## Problem

The judge↔sdlc result protocol is **prose-shaped**, and it drifts between the two sides:

- **Producer side (prose):** judge prompts tell the agent to emit a verdict
  (`SHIP | FIX-THEN-SHIP | REWORK`, or `CLEAN`) plus free-text findings.
- **Consumer side (code):** `sdlc merge`/`push`/`milestone-close` parse that text to decide
  pass/fail and to lift the `Review-Verdict:` trailer.

Because the contract lives only as prose on each side, the consumer mis-reads it. Observed
2026-06-02 (`nous#41` merge): the judge returned **`VERDICT: CLEAN` — "no action required
to ship"**, but the parser saw the judge's *explicitly-non-blocking note* and classified the
whole run as `failure`, forcing `--no-judge` (which then disables **all** judges). The
verdict token said pass; the prose presence said fail; the parser believed the prose.

There's no single definition either side can point at, so prompt authors and the parser can
silently disagree about what a "passing" result even looks like.

## Spec (operator framing)

> Define a **schema file** so that **prose and code both reference it.** Generally:
> schematize sdlc output — maybe just the **first line** of output.

Concretely:

- A single **machine-readable output contract** lives in one place (e.g.
  `construct/` as a schema/contract doc + optionally a JSON Schema). It defines:
  - the **first line** of any judge/verb result: `VERDICT: <token>` where `<token>` is an
    enum (`SHIP | FIX-THEN-SHIP | REWORK | CLEAN | BLOCK`);
  - optionally, severity-tagged findings (`[CRITICAL|IMPORTANT|MINOR|NIT|NOTE]`) +
    a blocking count, so the consumer can gate on severity, not on prose presence.
- **Both sides reference the one schema:** judge prompts are authored/generated to emit it;
  the sdlc parser reads **only** the structured first line (+ blocking count), never the
  presence of words like "findings"/"note." A `CLEAN`/`SHIP` with non-blocking prose passes.
- This kills the false-positive *and* removes prompt↔parser drift — change the contract in
  one file, both sides follow.

**Orthogonal, related:** `#64` (judge reads pre-merge base, not HEAD — *which diff* it sees,
not how the result is parsed) and `#67` (per-gate `--no-<gate>` bypass for close's evidence
gates; the same per-judge granularity should extend to push/merge's plan/specs/lessons
judges). This issue is the *result-protocol* half.

## Done when

- A single output-contract schema file exists under `construct/`, referenced by both the
  judge prompt(s) and the sdlc parser.
- `sdlc merge`/`push`/`milestone-close` gate on the structured verdict token (+ blocking
  severity), so a `CLEAN`/`SHIP`-with-non-blocking-notes result passes without `--no-judge`.
- Regression: a judge result that is `VERDICT: CLEAN` + a `[NOTE]` passes the merge gate.

## Plan

- [ ] Define the output contract (first-line `VERDICT:` enum + optional severity tags +
      blocking count) as a `construct/` schema doc (+ JSON Schema if useful).
- [ ] Point the judge prompt(s) at it (emit exactly this shape).
- [ ] Rewrite the sdlc result parser to read the structured first line only; gate on
      token + blocking severity.
- [ ] Regression test: CLEAN-with-NOTE passes; REWORK/BLOCK fails.

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F1). Operator:
"define some schema file so proses and code can both reference to … maybe just first line
of output."
