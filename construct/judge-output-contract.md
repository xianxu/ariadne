# Judge / verb output contract (#70)

The machine-readable contract for what an `sdlc` judge (or any verb that gates on
an LLM review) emits, and how the parser reads it. **Both sides reference this
one file**: the prompts embed `judge.ContractPreamble`, and the classifier
(`cmd/sdlc/internal/judge/classify.go`) reads only the verdict token. A drift
test (`judge_test.go`) keeps this doc's token list in sync with the Go source of
truth (`cmd/sdlc/internal/judge/contract.go`).

## The contract

A judge's response's **first line MUST be exactly**:

```
VERDICT: <TOKEN> (confidence: high | medium | low)
```

The parser reads **only** `<TOKEN>`. Findings, notes, and severity tags below it
are **advisory** — a non-blocking verdict *with* notes still PASSES the gate.
Do not put a title, heading, or preamble above the VERDICT line; it must lead.
(The parser tolerates a stray preamble defensively, but the contract is "lead
with the verdict".)

Every contract also carries the **unattended rule** (`judge.UnattendedRule`,
#300): the reviewer runs non-interactively and receives no notifications, so it
runs every command in the foreground and ends with its verdict. The claude
reviewer's whole run is read as a `stream-json` event stream: the review text is
every assistant message in order, so the latest verdict and findings blocks win
and a later message without them erases nothing. A run with no verdict anywhere
is dispatched once more. A run that never reviewed (its stream reports an error,
or a text-mode agent exits non-zero with a failure signature) is "review did not
run", never a verdict, unless it already gave one; an unreachable API names the
agent's host to allow.

## Tokens

| TOKEN           | gate     | meaning                                             |
|-----------------|----------|-----------------------------------------------------|
| `CLEAN`         | pass     | no issues; ready to ship                            |
| `INFO`          | pass     | informational / non-blocking notes only             |
| `SHIP`          | pass     | milestone review: ready, ship it                    |
| `FIX-THEN-SHIP` | pass     | milestone: ship after addressing findings           |
| `FAILURE`       | **fail** | issues that must be fixed before shipping           |
| `REWORK`        | **fail** | milestone: blocking, needs rework                   |
| `BLOCK`         | **fail** | generic hard block                                  |

The gate is binary — **blocking** (`FAILURE`, `REWORK`, `BLOCK`) vs **non-blocking**
(everything else). The token carries the human nuance; the parser keys off
blocking-ness only, never the presence of prose like "findings" or "note".

Which tokens a given check uses:

- **plan / plan-quality / specs**: `CLEAN | INFO | FAILURE`
- **dry / pure**: `CLEAN | FAILURE`
- **milestone-review**: `SHIP | FIX-THEN-SHIP | REWORK`

## Exception

**lessons** runs no agent — it emits a fixed `REMINDER:` line (classified as
`INFO`), not a VERDICT.

## Why

The contract used to live only as prose on each side, so the consumer mis-read
it: a `VERDICT: CLEAN` behind a title/preamble fell through to a sentinel-grep
that defaulted to `failure` and blocked merges; and a milestone review with no
leading-token-format recorded `unknown`. One referenced contract + a parser that
gates on the token (not prose) removes both the false-positive and the drift.
