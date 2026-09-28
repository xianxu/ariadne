---
id: '000147'
status: done
started: 2026-06-30T15:04:53-07:00
created: 2026-06-30
updated: 2026-06-30
estimate_hours: 0.9
actual_hours: 0.81
---

# Structured review verdict handoff via a verdict schema

## Problem

The boundary-review verdict handoff from the review subagent back to `sdlc` is
**unstructured**: `judge.Run` execs `claude -p <prompt>` and captures
`cmd.CombinedOutput()` — the agent's free-text response as one blob — and
`ParseVerdict` regex-scans it for a `VERDICT: <TOKEN>` line. The parser is
deliberately strict (it refuses tokens mid-sentence, because the judges review
the parser code itself and would self-trigger), so when a reviewer buries the
verdict in prose — observed repeatedly this session: *"the verdict stands:
**FIX-THEN-SHIP**"* (#143, #137) — the regex correctly misses it and the verdict
falls back to `unknown`. The fragility is **structural**: we are parsing an
agent's prose.

The verdict states are also **hand-synced across four places** — the prompt
instruction (`code-review.md`), the machine-read contract (`contract.go`
`ContractPreamble` + `ContractTokens` + `blockingTokens`), the parser
(`classify.go` regexes + the `Verdict` enum), and the consumers (close
milestone-verdict guard, the `Review-Verdict:` trailer, the log-line mirror).
The `Verdict` enum's own comment begs maintainers to update the prompt + verifier
on every change — a textbook drift hazard, exactly what `issue.cue` exists to kill.

This blocks #139 (close should finalize after the verdict): if `unknown` is
common because of prose verdicts, a halt-on-`unknown` policy fires constantly on
false alarms.
