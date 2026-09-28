---
id: '000070'
status: done
created: 2026-06-02
updated: 2026-06-03
estimate_hours: 2.5
actual_hours: 2.5
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
