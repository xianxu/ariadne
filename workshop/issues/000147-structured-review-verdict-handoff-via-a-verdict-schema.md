---
id: 000147
status: working
deps: []
target: agent-binary-handoff-schema
github_issue:
created: 2026-06-30
updated: 2026-06-30
estimate_hours: 0.9
started: 2026-06-30T15:04:53-07:00
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
- The review agent emits a structured verdict block; code reads it from the block
  deterministically (the block is authoritative; the prose regex remains a logged
  fallback only — its removal is a follow-up).
- A prose-buried verdict like *"the verdict stands: FIX-THEN-SHIP"* no longer
  degrades to `unknown` (the structured block is authoritative).
- Tests cover: structured parse (valid/missing/invalid token), the schema drift
  guard, and a process-level fixture for the agent→code handoff.
- Atlas documents the structured-handoff contract + the verdict vocabulary.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.2 impl=0.3
item: smaller-go-module      design=0.15 impl=0.2
design-buffer: 0.15
total: 0.9
```

Two pieces: (1) greenfield — a new CUE noun (`verdict.cue`) + its `pkg/vocab.Verdict()`
Go binding (embedded JSON, conformance test), the vocabulary single-source; (2) an
extend — `ParseVerdictBlock` + block-first `ParseVerdict`, the prompt rendering the
model's token set + drift guard, the sidecar verdict frontmatter. Design pre-resolved
by the durable plan → reduced design + +15% buffer; impl at the v3.1 40%-scaled tops.

Detailed design + TDD breakdown: `workshop/plans/000147-verdict-schema-handoff-plan.md`.

## Plan

Detailed TDD breakdown in `workshop/plans/000147-verdict-schema-handoff-plan.md`.

- [x] Brainstorm the emission format → **B2** (agent emits a fenced ```verdict
      block in stdout; binary validates + writes), chosen for agent-neutrality +
      the read-only anti-collusion property.
- [x] Model `verdict.cue`; wire it through `pkg/vocab` like `issue.cue`.
- [x] Derive the prompt-accepted token set + the parser set from the model;
      add the drift guard.
- [x] Read the verdict from the structured block (authoritative); legacy regex
      kept as a logged fallback.
- [x] Tests + atlas.

## Log

### 2026-06-30

Designed with the operator (B1/B2 fork → **B2**, read-only preserved; forced
structured output rejected for agent-neutrality). Crystallized the principle into
the [[agent-binary-handoff-schema]] target. Implemented per
`workshop/plans/000147-verdict-schema-handoff-plan.md`:
- **`construct/vocabulary/verdict.cue`** — the noun: `categories` (finalizing/
  blocking/internal) partition the tokens; `#Emitted`/`#Token` derived; closed
  `#Verdict` shape. The generic validator handles `--type verdict` with no binary
  change (an invalid token → `want: FIX-THEN-SHIP|REWORK|SHIP`).
- **`pkg/vocab.Verdict()`** — the Go binding (embedded JSON, `IsEmitted/IsFinalizing/
  IsBlocking/Emitted()`, `RenderBlockInstruction()`); `TestVerdictConformance`.
- **`ParseVerdictBlock`** (classify.go) — extracts the last fenced ```verdict block,
  validates the token against the model; `ParseVerdict` is now block-first →
  prose-fallback → unknown, so a prose-narrated verdict resolves from the block.
  `verdictFor` derives from the model (no switch).
- **Prompt** — `code-review.md`'s `{{VERDICT_BLOCK}}` renders from
  `RenderBlockInstruction()` (tokens from the model, not hardcoded).
- **Drift guard** (`TestVerdictDriftGuard`) pins all SHIP-family consumers to the
  model per the plan's disposition table: enum (equality), `verdictFor` (derive),
  prose regex + `ContractTokens` + `blockingTokens` (subset), prompt (equality).

Verification: `go test ./cmd/sdlc/... ./pkg/vocab/` all pass. New tests:
`TestVerdictConformance`, `TestParseVerdictBlock`, `TestParseVerdict_BlockBeatsProse`
(the session's `"the verdict stands: FIX-THEN-SHIP"` + a block → FIX-THEN-SHIP,
not unknown), `TestVerdictDriftGuard`, `TestDispatch_ResolvesVerdictBlock` (the
process-level handoff). `cue vet verdict.cue` + `validate-instance --type verdict`
green.

**Deferred (separable, not a Done-when):** carrying the verdict into the #136
sidecar *frontmatter* (so a sidecar is itself a validatable `#Verdict` instance) —
the verdict already appears in the sidecar's metadata table; promoting it to
frontmatter is a convergence nicety for a follow-up. The close finalize-policy
consumer (`closeVerdictOutcome` reading `vocab.Verdict()`) lands in **#139**, now
unblocked. Dropping the prose fallback once block adoption is confirmed — follow-up.