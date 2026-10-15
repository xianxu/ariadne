---
id: 000393
status: open
deps: [392, 380]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '32e68331a0a4b457c961e27cc9b784f8f69cbf1b' # card fields mirrored from issue-cards; edit via sdlc
---

# Serial per-repo landing admission with fleet-blocking priority

## Problem

Concurrent `sdlc merge` calls integrate against the same main; after one lands, the others reintegrate and rerun a roughly nine-minute suite plus queue wait. #376 required three integrations. A fleet-blocking fix (#391) also waited fourth for a suite with no priority lane. #390's operator-approved design chooses serial, per-repo landing admission before any batching.

## Spec

Keep `sdlc merge` as the agent verb. Admit one ready PR per repository at a time; the admitted candidate integrates current main, runs one full suite on its exact head, verifies review/CI/local gates and pinned PR/main OIDs, then lands. Later PRs wait; a stale head/main re-enters admission without falsely reusing suite evidence. Use an explicit crash-recoverable state model and uncertain-publication reconciliation. An operator or TL may grant a logged, expiring `fleet-blocking` priority for a named issue; it moves ahead of waiting normal candidates at the next admission, FIFO within each class, but never preempts a running suite or waives a gate. Preserve #370's short transaction-lock scope. Queue files live outside Git in Couch's storage directory when Couch is configured and in the repository Git common directory otherwise; `sdlc` reads/writes them directly with atomic writes and locking so Couch may be down. Couch only displays state. Cross-repo task dependencies remain the TL's pre-start sequencing responsibility. Follow #390's approved plan and side doc 01; rollout is an ariadne default for every SDLC repo after the ariadne/pair pilot.

## Done when

- [ ] Concurrent ready merges in one repo admit only one active integration; different repos are independent and the full-suite gate remains exact-head.
- [ ] Main/head drift, failed suite, process death in each phase, duplicate invocation and uncertain GitHub outcome recover or refuse without duplicate landing or lost evidence.
- [ ] Granted fleet-blocking work overtakes waiting normal work; same-class FIFO holds; running suites continue; expired, unauthorized or unlogged grants confer no priority. Grant and displacement are auditable.
- [ ] Landing works and recovers while Couch is down, with atomic queue files in its storage directory; a plain-shell setup uses the Git common directory.
- [ ] Stateful Git/GitHub fake and live read-only conformance cover the consumed interactions; pilot reports full suites per landed issue and p90 ready-to-land delay before/after.

## Plan

- [ ] Model legal queue/candidate states and exact-head invariants from #390; identify the existing Git/GitHub and runtime-storage seams.
- [ ] Test ordering, drift, crash and uncertainty against the production transition and fake, then implement one-repo admission and pilot.

## Log

### 2026-10-15

Filed from #390's approved design. Depends on the weekly baseline (#392) and saved local-check status recovery (#380). Batching and doc-only reuse are separate later choices.
