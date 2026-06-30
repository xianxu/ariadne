---
id: 000147
status: open
deps: []
github_issue:
created: 2026-06-30
updated: 2026-06-30
estimate_hours:
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

## Spec

Make the review verdict a **structured, schema-validated handoff** — the agent
generates it predictably, code reads it robustly (the generalized principle:
*structured handoffs at every agent↔code boundary; never parse an agent's prose*).

- **Single-source the verdict states in CUE** — a `construct/vocabulary/verdict.cue`
  (parallel to `issue.cue`) modeling the verdict tokens + their lifecycle: which
  are reviewer-emitted vs system-internal (`not-run`/`unknown`), which are
  blocking, and (for #139) which finalize a close. The prompt text, the parser's
  accepted set, the close policy, and the trailer all **derive** from this model
  rather than re-enumerating it.
- **Structured emission from the agent** — instruct the review subagent to emit
  its verdict as a deterministic block (a fenced ```` ```verdict ```` block, or
  the review-sidecar (#136) frontmatter) carrying `verdict:` + `confidence:`,
  rather than relying on a prose first line. The format must be one the agent
  fills reliably and code parses without prose heuristics.
- **Robust read** — code parses the structured block / frontmatter
  deterministically (the same validator path that checks `issue.cue` instances),
  not a regex over free text. `unknown` should then mean a *genuine* protocol
  violation, not a formatting near-miss.
- **Converge with #136** — the review sidecar is already a markdown+frontmatter
  artifact; prefer carrying the verdict in its frontmatter so the durable record
  IS the handoff (one artifact, schema-validated, human-readable, archived by
  #143).
- Preserve existing behavior: the `Review-Verdict:` trailer, the log-line mirror,
  the tri-state Classify (Clean/Info/Failure), and the gate semantics keep working;
  only the handoff format + the single-source change.

## Done when

- A `verdict.cue` (or equivalent) single-sources the verdict tokens + lifecycle;
  the prompt-accepted set + parser set are derived/asserted against it (a drift
  test fails if they diverge).
- The review agent emits a structured verdict block/frontmatter; code reads it
  deterministically without a prose regex.
- A prose-buried verdict like *"the verdict stands: FIX-THEN-SHIP"* no longer
  degrades to `unknown` (the structured block is authoritative).
- Tests cover: structured parse (valid/missing/invalid token), the schema drift
  guard, and a process-level fixture for the agent→code handoff.
- Atlas documents the structured-handoff contract + the verdict vocabulary.

## Plan

- [ ] Brainstorm the emission format (fenced `verdict` block vs sidecar
      frontmatter) + how the agent is instructed/constrained to produce it.
- [ ] Model `verdict.cue`; wire it through `pkg/vocab` like `issue.cue`.
- [ ] Derive the prompt-accepted token set + the parser set from the model;
      add the drift guard.
- [ ] Read the verdict from the structured artifact; keep the legacy regex as a
      fallback only (or remove once robust).
- [ ] Tests + atlas.

## Log

### 2026-06-30
