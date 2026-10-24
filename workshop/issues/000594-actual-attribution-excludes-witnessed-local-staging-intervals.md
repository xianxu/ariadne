---
id: 000594
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '840f06c0ce5d37d2ee74a9d229205ea4e01f0226' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T08:46:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Actual attribution excludes witnessed local staging intervals

## Problem

After #591, pair:6 cannot close pair#830 with numeric actuals because an explicitly witnessed local staging branch is treated as unknown/detached/ambiguous during concurrent claims.

## Spec

Investigate sanctioned stage/* branches as a witnessed exclusion, like resting branches: excluded and unassigned, without guessing which issue owns combined staging work. Preserve strict refusal for genuinely unknown, detached or ambiguous evidence and preserve adopted snapshots. This is a #591/#580 follow-up; no implementation or policy change is authorized by filing it.

## Done when

- A witnessed stage/000849-000850 interval under concurrent claims is covered by regression evidence and an approved attribution policy.
- Any excluded staging minutes are visible and unassigned; genuinely unknown evidence remains strict.
- Adopted snapshots remain immutable.

## Plan

- [ ] Approve the policy and implementation plan in workshop/plans/000594-stage-interval-attribution-plan.md with the hub before code.
- [ ] Reproduce the 172-second staging refusal, implement witnessed staging exclusion through selection and retained-proof validation, and test adversarial controls.
- [ ] Update help and atlas; run targeted checks and the recorded pre-close suite, close with the mandatory review, then land normally.

## Revisions

2026-10-24 — ops:0 approved the complete plan (submitted receipt a7713eb0-c36e-4728-9500-4a29a1cca2e6): attached witnessed stage/* exclusions under concurrent reservations, separate audit kind, shared predicate, visible minutes, preserved single-reservation ownership and adopted proofs, with the 172-second replay and malformed/detached/tamper controls. Proceeding through change-code.

2026-10-24 — ops:0 dispatched implementation planning (verified submitted receipt 570945d9-c5e2-4ea0-96ac-87545598f26b), superseding the filing-only restriction in Spec. Proposed policy: witnessed stage/* intervals under concurrent reservations are excluded and unassigned, with visible minutes; sole-reservation ownership-only behavior and adopted snapshots stay unchanged. Awaiting hub approval before code.

## Log

2026-10-24 — Filed unclaimed at ops:0 request (delivery 4ea9e0aa-75b7-4d0d-b85a-d30c69d94c1b); queued after #589. Evidence supplied directly by pair:6 (delivery ad33b571-b208-47bf-b08d-9ca25e6bf5ed), not independently reread from its checkout:

sdlc close --issue 830 refused: "branch attribution for session 0bca9e6b-3a49-425b-a689-eff7360aa8a1 (pair-slot6): concurrent claims have unknown, detached or ambiguous active branch during 2026-10-24T02:55:02Z to 02:57:54Z".

Reflog reported by pair:6:
- 2026-10-24T02:55:02Z checkout 000849-cast-notes-dom-submission-loops-on-feedback-changed-and-never-submits -> stage/000849-000850. TL-authorized local staging-only branch, created from pair#849.
- At the same timestamp, merge origin/000850-... into the stage branch; peer-reported object abbreviation 6f0478a3a.
- 2026-10-24T02:57:54Z checkout stage/000849-000850 -> 000849-cast-notes-...; peer-reported object abbreviation 3de33c10b.
- Claims held: pair#849 and pair#830. Window: 172 seconds (2m52s). Combined staging could reasonably be excluded, but charging it to pair#849 would infer ownership from context and needs a separate policy decision.

Hash abbreviations above are quoted peer evidence, not independently verified Git output. No actuals or source telemetry were modified.

2026-10-24 — Claimed and started from current main. Recovery used issue and Git evidence; prior TTY capture remains unavailable. The existing pure selector and saved-proof validator both enumerate resting-excluded, so the extension must cover both rather than changing only live selection. Current actual/active-time help still says resting evidence refuses; correct those surfaces with the new policy.
