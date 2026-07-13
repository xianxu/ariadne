# Boundary Review — ariadne#168

| Field | Value |
|-------|-------|
| Boundary | whole-issue close |
| Window | `9e3f269..44898bf` |
| Reviewer | Codex fresh context |
| Verdict | `REWORK` |
| Date | 2026-07-12 |

## Findings

1. **Critical — false PURE entity contract.** The plan modeled
   `EvidenceSource`, `RetroFinding`, and `FollowUpRecommendation` as PURE
   entities, but the implementation is one prose Agent Skills procedure with no
   IO-free code unit or unit-test surface.
2. **Important — evaluation retention mismatch.** A checked plan step promised
   verbatim baseline evidence, while the evaluation artifact retained only
   excerpts and ledger assertions.
3. **Important — README discovery missing.** The new user-invoked skill had an
   atlas entry but no README pointer.

The reviewer confirmed the skill implementation, existing Weave export reuse,
untrusted-evidence boundary, 24-record GREEN ledger, scoped tests, `ARCH-DRY`,
and `ARCH-PURPOSE` were otherwise sound.

## Resolution

- Revised the Core Concepts section and appended a plan revision: no executable
  pure entity exists; `session-retro` is one prompt-level integration surface
  evaluated with fixed evidence and fresh workers/scorers (`ARCH-PURE`).
- Added complete baseline and final-GREEN worker/scorer outputs to the durable
  evaluation record so every ledger row is auditable.
- Added README discovery/usage text linking the atlas map.
- Added a lesson preventing prose concepts from being mislabeled as PURE and
  requiring evaluation retention to match its promise.

Re-run `sdlc close` after committing these resolutions.
