---
id: 000594
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'd37d13d0eefd054859396bc1163fb81b967684bd' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

2026-10-24 — Filed unclaimed at ops:0 request (delivery 4ea9e0aa-75b7-4d0d-b85a-d30c69d94c1b); queued after #589. Evidence supplied directly by pair:6 (delivery ad33b571-b208-47bf-b08d-9ca25e6bf5ed), not independently reread from its checkout:

sdlc close --issue 830 refused: "branch attribution for session 0bca9e6b-3a49-425b-a689-eff7360aa8a1 (pair-slot6): concurrent claims have unknown, detached or ambiguous active branch during 2026-10-24T02:55:02Z to 02:57:54Z".

Reflog reported by pair:6:
- 2026-10-24T02:55:02Z checkout 000849-cast-notes-dom-submission-loops-on-feedback-changed-and-never-submits -> stage/000849-000850. TL-authorized local staging-only branch, created from pair#849.
- At the same timestamp, merge origin/000850-... into the stage branch; peer-reported object abbreviation 6f0478a3a.
- 2026-10-24T02:57:54Z checkout stage/000849-000850 -> 000849-cast-notes-...; peer-reported object abbreviation 3de33c10b.
- Claims held: pair#849 and pair#830. Window: 172 seconds (2m52s). Combined staging could reasonably be excluded, but charging it to pair#849 would infer ownership from context and needs a separate policy decision.

Hash abbreviations above are quoted peer evidence, not independently verified Git output. No actuals or source telemetry were modified.
