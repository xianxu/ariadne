---
id: 000040
status: open
estimate_hours: 1.5
deps: []
created: 2026-05-27
updated: 2026-05-27
related: [cmd/sdlc/internal/judge/prompts.go, cmd/sdlc/internal/judge/classify.go, cmd/sdlc/internal/judge/judge_test.go]
---

# Judge classifier: structured VERDICT line instead of free-text grep

## Problem

`sdlc judge plan` (and `specs`, `lessons`) classify the subagent's
output via `classify.go:Classify` — a regex match against a tight
allowlist of "clean" sentinels (`no DRY/PURE violations found`,
`in sync`, `no issue files changed`, etc.). Anything non-empty that
doesn't match → `Failure`.

In practice, when a Plan judge actually approves a close (e.g.
`"✅ Looks good — properly closed"`, `"No action needed"`), its prose
doesn't match any sentinel and gets scored `Failure`. `sdlc merge`
then refuses, and the operator reaches for `--no-judge` — which
`sdlc merge --help` calls "emergency only" but is in fact the
routine path. The guard ends up with negative value.

Concrete example from pair#23 close (2026-05-27):

```
==> invoking claude for Check issue plan completeness …
# TPM Review: 000023-...
## Status: ✅ Looks good — properly closed
...No action needed — issue is complete and well-documented.
  [!] Check issue plan completeness: findings reported — review above
Error: pre-merge judges failed: Check issue plan completeness: failure
```

The right shape already exists in the same package: `ParseVerdict`
(classify.go:91) keys off a structured first-line
`SHIP | FIX-THEN-SHIP | REWORK` for `MilestoneReview`. The
fix is to give Plan + Specs the same structured-verdict treatment.

## Done when

- `sdlc judge plan` and `sdlc judge specs` instruct the subagent to
  emit a `VERDICT: CLEAN | INFO | FAILURE` line as line 1 of its
  response, with optional `(confidence: ...)` parenthetical to
  match MilestoneReview's shape.
- `Classify` keys off the verdict line first; falls back to the
  existing `cleanRE`/`infoRE` grep for safety (handles outputs
  from prompts that haven't been migrated, agents that ignore the
  instruction, etc.).
- `Lessons` keeps its current `LessonsReminder` Info path — no
  agent invocation, no verdict line needed.
- `MilestoneReview` keeps its existing `SHIP / FIX-THEN-SHIP /
  REWORK` verdict (different semantics; mapped via a separate code
  path). No churn there.
- Existing tests still pass; new tests cover the verdict-line path
  for each of CLEAN / INFO / FAILURE plus a "verdict line missing,
  legacy grep wins" case.

## Spec

**New convention:** Plan + Specs prompts append a section like:

```
Produce a structured report. First line must be:

  VERDICT: CLEAN | INFO | FAILURE   (confidence: high | medium | low)

CLEAN  = no issues found, ready to ship.
INFO   = informational notes only; no action required to ship.
FAILURE = issues found that must be addressed before shipping.

Then on subsequent lines: a 1-paragraph summary + any findings.
```

**Classifier change:** `Classify` becomes:

```go
func Classify(output string) Outcome {
    s := strings.TrimSpace(output)
    if s == "" {
        return Failure
    }
    if v, ok := parseVerdictLine(s); ok {
        return v // CLEAN→Clean, INFO→Info, FAILURE→Failure
    }
    // Legacy fallback.
    if cleanRE.MatchString(s) { return Clean }
    if infoRE.MatchString(s)  { return Info  }
    return Failure
}
```

`parseVerdictLine` is a tiny helper that matches
`^\s*VERDICT:\s*(CLEAN|INFO|FAILURE)\b` on the first non-empty line.
Modeled on `ParseVerdict` at `classify.go:91`.

## Plan

- [ ] Edit `prompts.go` Plan branch: append the structured-verdict
      instructions to the existing template.
- [ ] Edit `prompts.go` Specs branch: same instruction block.
- [ ] Edit `classify.go`: add `verdictLineRE` + `parseVerdictLine`,
      gate `Classify` on it before the legacy grep.
- [ ] Edit `judge_test.go`: add cases for each verdict value, plus a
      legacy-fallback case (no verdict line → falls through to
      `cleanRE`/`infoRE`/Failure as today). Confirm existing test
      table still passes.
- [ ] Atlas: if `atlas/` documents the judge classifier or prompt
      contract, add a one-line note that Plan/Specs subagents must
      emit `VERDICT:` on line 1. Skip if no such atlas page exists.

## Log

(empty)
