# Structured Verdict Handoff (verdict.cue) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the fragile prose-regex verdict handoff with a **schema-validated structured handoff**: the (read-only) review agent emits a fenced ```` ```verdict ```` block; the binary parses it deterministically and validates it against a CUE-modeled `verdict.cue` single-source; `unknown` then means a genuine protocol violation, not a formatting near-miss. Agent-neutral (B2 — no provider-specific forced output); preserves the read-only anti-collusion property.

**Architecture:** `construct/vocabulary/verdict.cue` (parallel to `issue.cue`) is the **single source** for the boundary-review verdict tokens + their semantics (finalizing / blocking / system-internal). `pkg/vocab.Verdict()` exposes them to Go (the same embedded-JSON binding pattern as `Issue()`), so the prompt's accepted set, the parser, and (later, #139) the close policy all **derive** from one model — killing the 4-place hand-sync. The review prompt instructs a fenced ```` ```verdict ```` block; a pure `ParseVerdictBlock` extracts + validates it against `vocab.Verdict()`; the legacy prose `ParseVerdict` stays only as a logged fallback. The validated verdict is carried into the #136 sidecar so the durable artifact records it. Read-only is untouched — the agent emits in stdout; the binary writes + validates (ARCH-PURE: pure parse/validate core, thin IO).

**Tech Stack:** CUE (`verdict.cue` + `cue vet`); Go (`pkg/vocab` binding, embedded JSON, the block parser); `pkg/frontmatter` for instance validation; tests mirroring the issue/pensive vocabulary patterns; atlas.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `verdict.cue` (`#Verdict`, categories) | `construct/vocabulary/verdict.cue` | new |
| `VerdictModel` + `Verdict()` | `pkg/vocab/vocab.go` | new |
| `ParseVerdictBlock` | `cmd/sdlc/internal/judge/classify.go` | new |

- **verdict.cue** — `package verdict`; `categories: { finalizing: ["SHIP","FIX-THEN-SHIP"], blocking: ["REWORK"], internal: ["not-run","unknown"] }`; derived `#Emitted` (finalizing+blocking — the set the reviewer emits) and `#Token` (all); `#Verdict: { verdict: #Emitted, confidence?: ("high"|"medium"|"low") }` (closed/fail-closed, like `pensive.cue`). The one place the boundary-review verdict vocabulary lives.
  - **DRY rationale:** the prompt list, the parser's accepted set, the close finalize-policy, and the trailer all derive from `categories`; today they're hand-synced across `code-review.md`, `contract.go`, `classify.go`, and `close.go` (the `Verdict` enum's own comment begs maintainers to sync them).
  - **Future extensions:** a 5th token or a re-categorization is a one-line `categories` edit; consumers track it.
- **VerdictModel + Verdict()** — the Go binding mirroring `Issue()`: `Verdict().IsEmitted(t) / IsFinalizing(t) / IsBlocking(t) / Emitted() []string`. Backed by embedded `pkg/vocab/verdict.json` (`//go:generate vocabulary export --noun verdict`). The single Go read of the model.
- **ParseVerdictBlock** — `ParseVerdictBlock(output string) (token, confidence string, ok bool)`: extracts the LAST fenced ```` ```verdict … ``` ```` block from the agent output and parses its `verdict:`/`confidence:` keys (flat YAML, same shape as `issue.GetField`). `ok=false` when no well-formed block is present. Pure; the caller validates `token` against `vocab.Verdict().IsEmitted`.
  - **DRY rationale:** the structured counterpart to the prose `ParseVerdict`; both feed the same `Verdict` enum, but the block path is authoritative.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| review prompt (verdict block) | `cmd/sdlc/internal/judge/code-review.md` + `contract.go` | modified | agent instruction |
| `ParseVerdict` (block-first) | `cmd/sdlc/internal/judge/classify.go` | modified | verdict resolution |
| sidecar verdict frontmatter | `cmd/sdlc/reviewsidecar.go` | modified | durable record |
| drift guard | `pkg/vocab/conformance_test.go` + a judge test | new | model↔consumers |

- **review prompt** *(modified)* — `ContractPreamble` / `code-review.md` instruct the agent to emit, as a self-contained fenced block, `\n```verdict\nverdict: <TOKEN>\nconfidence: <high|medium|low>\n````\n`, with `<TOKEN>` from `vocab.Verdict().Emitted()` (rendered into the prompt, not hardcoded). Keep the existing prose `VERDICT:` line as a documented fallback during transition.
- **ParseVerdict** *(modified)* — resolve order: `ParseVerdictBlock` (validated against `vocab.Verdict()`) first; if absent, the legacy prose scan (logged as a fallback); if neither, `VerdictUnknown`. So a prose-buried verdict still has the block as the authoritative path.
- **sidecar verdict frontmatter** *(modified)* — `writeReviewSidecar` carries the validated `verdict:` (+ `confidence:`) into the sidecar so the durable artifact is itself a `#Verdict` instance (validatable via `vocabulary validate-instance --type verdict`). Convergence with #136/#143.
- **drift guard** *(new)* — a `pkg/vocab` conformance test (every token in exactly one category; `Emitted()` matches the prompt-rendered set) + a judge test asserting the parser/prompt accept exactly `vocab.Verdict().Emitted()`. Fails if the model and consumers diverge.

**Test surface.** CUE: `cue vet verdict.cue` + export (mirror `vet_test.sh`). `pkg/vocab`: `TestVerdictConformance` (mirror `TestIssueConformance`). `cmd/vocabulary`: `TestValidateInstance_VerdictGeneralizes` (a real `workshop/…-close-review.md`-style frontmatter validates against `#Verdict`; an invalid token fails) — proves the generic validator works for the new noun. `judge`: `ParseVerdictBlock` table (valid / missing / invalid token / multiple blocks → last wins) + the resolve-order test (block beats prose; prose fallback still works; the session's `"the verdict stands: FIX-THEN-SHIP"` + a `verdict` block → parses FIX-THEN-SHIP, not unknown). Process-level: stub `judge.Run` to return a body with a `verdict` block; assert dispatch resolves it.

---

## Design decisions

- **D1 — B2, agent-neutral.** The agent emits a fenced block in stdout (every agent can); the binary parses + validates. No provider-specific forced output (would break ariadne's agent-neutral contract). Read-only preserved — the reviewer never writes a file.
- **D2 — `verdict.cue` is THE source; consumers derive.** `categories` drives `#Emitted`/`#Token`; `vocab.Verdict()` is the Go read; the prompt renders `Emitted()`; the parser validates against it; a drift test fails on divergence. Kills the 4-place hand-sync.
- **D3 — block-first, prose-fallback, then unknown.** During transition the prose `VERDICT:` line still parses (no regression), but the structured block is authoritative — so the session's prose verdicts resolve correctly. Once adoption is confirmed, the fallback can be dropped (a later issue).
- **D4 — scope to boundary-review verdicts.** `verdict.cue` models SHIP/FIX-THEN-SHIP/REWORK (+ internal not-run/unknown). The pre-merge tri-state tokens (CLEAN/INFO/FAILURE/BLOCK in `contract.go`) are a *separate* judge surface — out of scope (note it; a future `judge-outcome.cue` could model them too).
- **D5 — sidecar carries the verdict (converge with #136).** The durable record becomes a `#Verdict` instance — schema-validatable, human-readable, archived by #143.
- **D6 — unblocks #139.** With robust verdicts, #139's halt-on-`unknown` fires only on genuine violations. The `vocab.Verdict().IsFinalizing/IsBlocking` predicates are exactly what #139's `closeVerdictOutcome` will read.

---

## Chunk 1: The model + Go binding (single source)

### Task 1: `verdict.cue` + `cue vet`

**Files:** Create `construct/vocabulary/verdict.cue`; mirror `construct/vocabulary/vet_test.sh`

- [ ] Model `package verdict`: `categories` (finalizing/blocking/internal), derived `#Emitted`/`#Token`, `#Verdict` (closed). `cue vet construct/vocabulary/verdict.cue` + `cue export` clean. Confirm `vocabulary validate-instance --type verdict <fixture.md>` resolves the new noun (generic engine — no binary change). **Commit** — `#147: verdict.cue — single-source the boundary-review verdict vocabulary`

### Task 2: `pkg/vocab.Verdict()` binding

**Files:** `pkg/vocab/vocab.go`, `pkg/vocab/verdict.json` (generated), `pkg/vocab/conformance_test.go`

- [ ] Add `VerdictModel` + `Verdict()` + `//go:generate vocabulary export --noun verdict`; `make vocab-embed`; commit `verdict.json`. Add predicates `IsEmitted/IsFinalizing/IsBlocking/Emitted()`. `TestVerdictConformance` mirrors the issue conformance test. **Commit** — `#147: pkg/vocab — Verdict() Go binding (derived from verdict.cue)`

## Chunk 2: Structured emission + robust parse

### Task 3: `ParseVerdictBlock` + block-first `ParseVerdict`

**Files:** `cmd/sdlc/internal/judge/classify.go`; `cmd/sdlc/internal/judge/judge_test.go`

- [ ] **TDD:** table test for `ParseVerdictBlock` (valid/missing/invalid/last-wins) + the resolve-order test (block beats prose; prose `"the verdict stands: FIX-THEN-SHIP"` alone still parses via fallback; block + prose → block wins). Implement `ParseVerdictBlock` (extract fenced ```` ```verdict ```` block, parse flat keys, validate token via `vocab.Verdict().IsEmitted`); rewire `ParseVerdict` to block-first → prose-fallback → unknown. Full judge suite green. **Commit** — `#147: parse the structured verdict block (authoritative), prose as fallback`

### Task 4: prompt emits the block (derived from the model)

**Files:** `cmd/sdlc/internal/judge/contract.go` (`ContractPreamble`), `code-review.md`; the drift test

- [ ] Render the ```` ```verdict ```` block instruction with the token set from `vocab.Verdict().Emitted()` (not hardcoded). Add the drift guard (prompt-accepted set == `Emitted()`). Update the `CodeReviewBody`/contract tests. **Commit** — `#147: instruct the reviewer to emit a fenced verdict block; add the drift guard`

## Chunk 3: Sidecar + docs

### Task 5: sidecar carries the validated verdict

**Files:** `cmd/sdlc/reviewsidecar.go`; `reviewsidecar_test.go`

- [ ] Add `verdict:`/`confidence:` to the sidecar frontmatter so the durable record is a `#Verdict` instance; assert it validates via `vocabulary validate-instance --type verdict`. **Commit** — `#147: sidecar records the validated verdict in frontmatter`

### Task 6: atlas + vocabulary doc

**Files:** `atlas/workflow/vocabulary.md`, `atlas/workflow/sdlc-binary.md`

- [ ] Document the verdict noun + the structured-handoff contract (agent emits a fenced block; binary validates against `verdict.cue`; the generalized "schema'd handoffs, never parse prose" principle). **Commit** — `#147: atlas — verdict vocabulary + structured handoff contract`

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| `verdict.cue` single-sources tokens + lifecycle; prompt/parser derived w/ drift guard | Tasks 1, 2, 4 |
| agent emits a structured block; code reads it deterministically (no prose regex) | Tasks 3, 4 |
| prose-buried verdict no longer degrades to `unknown` | Task 3 (block authoritative) |
| tests: structured parse + drift guard + process fixture | Tasks 2–5 |
| atlas documents the contract + vocabulary | Task 6 |

## Non-goals / follow-up

- The close finalize-policy consumer (`closeVerdictOutcome` reading `vocab.Verdict()`) lands in **#139** (unblocked by this).
- The pre-merge tri-state tokens (CLEAN/INFO/FAILURE) are a separate surface — a future `judge-outcome.cue`.
- Dropping the prose fallback entirely — a follow-up once block adoption is confirmed in the wild.
- The push/merge instance-conformance gate (`validategate.go`) hardcodes `issue`; parameterizing it for verdict sidecars is optional future work.
