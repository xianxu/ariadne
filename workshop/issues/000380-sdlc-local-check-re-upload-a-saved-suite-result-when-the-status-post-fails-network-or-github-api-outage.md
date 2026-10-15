---
id: 000380
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '715030d3bd9efa91fa11096dac679229a7f98346' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc local-check: re-upload a saved suite result when the status POST fails (network or GitHub API outage)

## Problem

On 2026-10-14 pair:7 landed pair#559 during a GitHub API outage. Both recorded full suites passed, but the commit
status upload failed, so the merge gate saw no status. sdlc has no verb to upload a saved result, so the worker
replayed the failed status POST by hand (with the exact head, the saved log's sha256 and exit 0). It was honest and
verified, but a hand-built API call is outside the binary's control: nothing checks that the replayed status matches
a recorded result.

## Spec

- A verb (e.g. `sdlc local-check upload --name suite [--head SHA]`) that re-posts the status from the **saved**
  record of a completed run: same head, exit code, log hash and description as the original would have had.
- Refuses when there's no saved record for that head, when the record failed, or when HEAD moved since the run.
- The failing upload's error message names this verb as the next action.
- A test with a fake GitHub status endpoint: the first POST fails, the verb succeeds, and the gate passes; a tampered
  or missing record is refused.

## Done when

- A suite whose upload failed can be published with one sdlc verb, without rerunning it and without hand-built API calls.
- Tests cover success, missing record, failed record and moved head.

## Plan

- [ ]

## Log

### 2026-10-14
