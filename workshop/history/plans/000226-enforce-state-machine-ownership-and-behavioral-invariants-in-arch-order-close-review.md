# Boundary Review — ariadne#226 (whole-issue close)

| field | value |
|-------|-------|
| issue | 226 — Enforce state-machine ownership and behavioral invariants in ARCH-ORDER |
| repo | ariadne |
| issue file | workshop/issues/000226-enforce-state-machine-ownership-and-behavioral-invariants-in-arch-order.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b6ebcaf60f4fe616a5ed0ba43170f1ec2f130fc..c1d295412365ff02b28c8e8e8dabae06634c92f2 |
| command | sdlc close --issue 226 |
| reviewer | codex |
| timestamp | 2026-09-14T15:40:40-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The policy matches the Spec and reaches existing consumers through the shared registry. Delivery tests pass, but the prompt golden test fails because four snapshots retain the old policy. The atlas also contradicts the revised review ordering.

## 1. Strengths

- Structural ownership and independent sequence invariants are explicit in [architecture.md](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/architecture.md:162).
- Existing cancellation, interleaving, and task-lifetime guidance is preserved.
- Prompt construction, `start-plan`, and `arch-principles` continue deriving from the canonical registry.

## 2. Critical findings

None.

## 3. Important findings

- **Prompt snapshots are stale.** [architecture.md:162](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/architecture.md:162) changes rendered prompts without updating `dry.prompt`, `pure.prompt`, `plan-quality.prompt`, and `milestone-review.prompt`. `TestBuildPrompt_Golden` fails. Regenerate and inspect all four snapshots for this intentional policy change, then rerun the judge tests. **ARCH-PURPOSE**
- **Atlas contradicts the revised policy.** [architecture-principles.md:80](/Users/xianxu/workspace/ariadne/atlas/workflow/architecture-principles.md:80) calls leading with the oracle clause a deliberate design choice. The revised review lens now leads with production routing and bypass prevention. Update this map-level explanation without duplicating the policy. **ARCH-DRY**

## 4. Minor findings

None.

## 5. Test coverage notes

Executed:

- Architecture/deferred tests: **PASS**
- CLI architecture/start-plan tests: **PASS**
- `TestBuildPrompt_Golden`: **FAIL**
- Pinned-range `git diff --check`: **PASS**

Read-only comparison confirmed all four affected snapshots contain the complete base registry and omit the head registry. The existing `bin/sdlc arch-principles` delivers the head registry.

## 6. Architectural notes

| Principle | Assessment |
|---|---|
| ARCH-DRY | **Flag:** stale atlas explanation; production consumers remain single-sourced. |
| ARCH-PURE | **Pass:** shared pure rendering remains intact. |
| ARCH-PURPOSE | **Flag:** policy delivered, but executable snapshots remain inconsistent. |
| ARCH-MOCK | **Pass:** no new external dependency. |
| ARCH-CONSTRAINTS | **Pass:** bounded static prompt expansion. |
| ARCH-SECURE | **Pass:** no new input or credential boundary. |
| ARCH-ORDER | **Pass:** ownership, bypass prevention, and independent invariants meet the Spec. |
| ARCH-FUNERAL | **Pass:** artifacts use existing issue/ledger archival conventions. |

README changes are unnecessary: no command, flag, configuration, or usage surface changes.

## 7. Plan revision recommendations

Append a dated `## Revisions` entry adding reviewed regeneration of affected prompt snapshots, `TestBuildPrompt_Golden` verification, and correction of the atlas’s review-order explanation.

```findings
findings:
  - id: new
    severity: Important
    family: executable-prompt-snapshot-consistency
    title: |
      Update all four prompt snapshots for the intentional policy change.
    detail: |
      cmd/sdlc/internal/judge/architecture.md:162 changes rendered prompts, but dry.prompt, pure.prompt, plan-quality.prompt, and milestone-review.prompt still contain the base registry. TestBuildPrompt_Golden fails; regenerate and inspect these snapshots, then rerun the judge tests. ARCH-PURPOSE.
  - id: new
    severity: Important
    family: architecture-map-consistency
    title: |
      Correct the atlas explanation of ARCH-ORDER review ordering.
    detail: |
      atlas/workflow/architecture-principles.md:80-84 describes oracle-first ordering as deliberate, while architecture.md:192 now leads with production routing and bypass prevention. Update the map-level explanation to match the approved policy without copying its clauses. ARCH-DRY.
```

---

## Re-review — 2026-09-14T15:43:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 226 — Enforce state-machine ownership and behavioral invariants in ARCH-ORDER |
| repo | ariadne |
| issue file | workshop/issues/000226-enforce-state-machine-ownership-and-behavioral-invariants-in-arch-order.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b6ebcaf60f4fe616a5ed0ba43170f1ec2f130fc..127b04f3bbf60367cc18795f1eaae47a37b82135 |
| command | sdlc close --issue 226 |
| reviewer | codex |
| timestamp | 2026-09-14T15:43:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fulfills the issue’s policy scope. All three prior findings are addressed: verification functions are named, all four prompt snapshots match the revised registry, and the atlas accurately describes review ordering. No new findings.

## 1. Strengths

- [architecture.md:162](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/architecture.md:162) explicitly requires authoritative state ownership and independently stated sequence invariants.
- Existing interruption, ordering, cancellation, and task-lifetime guidance remains intact.
- All four prompt snapshots change solely through replacement of the canonical registry; production consumers retain shared rendering.

## 2. Critical findings

None.

## 3. Important findings

None.

## 4. Minor findings

None.

## 5. Test coverage notes

Executed successfully:

- Complete judge package: `go test ./cmd/sdlc/internal/judge -count=1`
- Architecture/start-plan CLI tests.
- Pinned-range `git diff --check`.
- Exact comparison of all four snapshots against base/head registries.
- Existing `bin/sdlc` output contains the complete pinned registry.

`TestBuildPrompt_Golden` exercises production rendering. In-memory reversal of the registry replacement breaks snapshot equality for every affected prompt. No scratch mutation or fresh binary build was performed.

## 6. Architectural notes

| Principle | Assessment |
|---|---|
| ARCH-DRY | Pass — shared registry and rendering; atlas maps rather than duplicates policy. |
| ARCH-PURE | Pass — rendering remains pure; no new IO logic. |
| ARCH-PURPOSE | Pass — approved enforcement policy reaches all existing consumers. |
| ARCH-MOCK | Pass — no new external dependency. |
| ARCH-CONSTRAINTS | Pass — bounded static prompt expansion. |
| ARCH-SECURE | Pass — no new input or credential boundary. |
| ARCH-ORDER | Pass — production ownership, bypass prevention, and independent invariants explicitly required. |
| ARCH-FUNERAL | Pass — review artifacts follow existing archival conventions. |

Atlas updated. README changes are unnecessary because no command, flag, configuration, or usage surface changes.

## 7. Plan revision recommendations

None. Existing revisions cover verification clarification and expanded consistency work. No Core concepts table requires cross-checking.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The issue's verification revision names BuildPrompt, runArchPrinciples, and runStartPlan and describes full-registry assertions plus CLI comparison; inspected tests support this description.
  - id: BR-2
    disposition: addressed
    note: |
      All four snapshots exactly replace the base registry with the pinned head registry. The complete judge suite, including TestBuildPrompt_Golden, passes; reversing the replacement breaks each snapshot equality.
  - id: BR-3
    disposition: addressed
    note: |
      atlas/workflow/architecture-principles.md:80 now describes executable-model checks preceding the retained oracle critique, matching cmd/sdlc/internal/judge/architecture.md:192.
```
