---
id: 000510
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2e34a0c614f9fa8ab124ee4498c9914a535829a2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T02:17:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "51125e6d", done: "a2b9b094"}
actual_hours: 0.52
---

# sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head

## Problem

`sdlc local-check cancel` selects runs only by `--name`. On 10-20, pair:6 had two queued `suite` runs from one
checkout (an original on an older head, and an affected-check fallback to the full suite on the current head, the
pair#738 case) and could not cancel just one: `--name suite` cancels both and drops the original's FIFO place, so it
had to cancel both and requeue.

## Spec

- `sdlc local-check cancel --run ID` (the leased run id that `local-check status` already shows), alongside `--name`;
  refuses an ID owned by another checkout.
- When `local-check run` would queue a second run of the same check from the same checkout on a **different** head,
  say so and offer to replace the stale one, instead of silently queueing both.

## Done when

- Tests pin: cancel by run ID leaves the sibling run queued in place; a foreign ID refuses; the duplicate-run notice.

## Plan

- [x] Add runtime CLI regressions for exact cancellation of one of two real queued runs, retained sibling FIFO, and foreign/invalid selectors.
- [x] Expose the existing lease ID; add exact-ID cancellation and same-check/different-head notices using the shared lease reader and managed interruption path.
- [x] Verify targeted cancellation/notice controls and update help/README/atlas.

Actions: start a check; cancel one ID; cancel by name/all; retry a completed or refused selection.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Two owned runs, same check, different heads | Both keep their FIFO places; notice names IDs/heads and exact cancel command | Cancel one ID | Selected run's queue place only |
| Owned selected live lease, matching PID generation | Exact selected run is cancelled; sibling stays queued/running | Start a check again if needed | Selected run's unrecorded progress and FIFO place |
| Foreign, invalid or unknown ID; unreadable identity | Refusal, no process signalled | Inspect the printed run ID and retry selection | None |
| Process generation no longer matches, or unwind times out | Existing refusal/unconfirmed diagnostic | Inspect and retry | Cancellation may have interrupted selected progress; sibling is untouched |

## Log

### 2026-10-20

- 2026-10-21: Dispatched by ops:0 after #512 landed. Claimed and entered planning from fresh main. #520 is queued after this item, plan first; it is not claimed or stacked here.
- Root cause: cancellation scans every held lease for this checkout/check, with no exact selector. Existing run lifetime flocks and PID-start checks already provide authority; reuse them (ARCH-DRY/ARCH-SECURE) rather than inspect global processes. The lease filename is the unique ID, held for the run lifetime; no new persisted identity/state is needed.
- Source correction: there is no `local-check status` command, and lease IDs are not printed. Queue `op-…` IDs are separate. Print the lease ID when publishing a run and in duplicate notices. The offer to replace is an exact cancel command; neither a prompt nor automatic cancellation changes the old run. Existing name/all cancellation remains explicit and unchanged.
- Alternatives: select by SHA would remain ambiguous for multiple attempts; reuse queue operation IDs would cross two lifetime/ownership models. Select the existing lease ID instead. Decode and flock checks stay in the IO boundary; selection is based on the canonical checkout, check, ID and SHA (ARCH-PURE/ARCH-ORDER). Missing proof never widens cancellation authority.
- The duplicate notice is advisory, so observing one lease census is sufficient: it does not claim or enforce absence of concurrent duplicates. One scan is O(N) in current lease files, no subprocess per record; preserve the existing cancellation wait budget. Tests use private suite locks and only fixture-owned child processes, never live fleet runs. The mistake is semantic selector authority, defended by runtime/FIFO regressions rather than a lexical lint rule.
- Quick plan posted to the hub, proceeding unless objected. Existing issue requirements settle the behavior; clarified the absent status premise without adding a status subsystem.


### 2026-10-21 — implementation verification
- 2026-10-21: closed — Resolved README/help integration conflicts by retaining main scoped-check docs and exact-ID cancellation docs; executable files merged automatically. Integrated targeted real FIFO/cancellation/notice/signal tests passed (8.507s), and recorded full suite passed 1924 SDLC tests plus 75 packages with zero failures. Initial review SHIP/high, no findings; hub accepted actual built CLI proof. (fixed-then-ship; full suite: Recorded sdlc-local/suite passed on dcc50b776875f0dd40a0d2639a5fed3a79063d5b: 1924 cmd/sdlc tests across 8 shards and 75 other packages, zero failures, 12m13s, log sha256:532513943510.) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-21: closed — Targeted cancellation/lifetime/ownership matrix passed (248.227s); notice controls and linear N=8/16 scans passed (2.845s); signal cleanup passed (6.603s); Go catalog passed (0.795s). Own clean built CLI 18e7cda53cfbfcee5bfe99fd75005b94873d113d passed private real two-run FIFO, exact-ID and refusal proof (5.960s), output p510-live-proof.txt outside checkout. Recorded suite queued on pushed PR364 head; final-head pass required before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Real two-run regression failed before implementation: IDs and duplicate notice were absent (1.802s). After exposing the lease IDs/notice, it independently failed at the unknown `--run` flag (1.884s). With exact selection implemented it passed (2.177s): the newer run cancels via managed SIGTERM, while the older lease and FIFO ticket retain their inode, bytes and place; releasing capacity admits the survivor, which still cancels by name.
- Foreign, unknown, pathlike, wildcard, empty and mixed ID/name selectors refuse without signalling legitimate children. Existing PID-generation, stale-lease, checkout ownership and lease lifetime controls passed with the new tests (248.227s package run). Duplicate notice controls and one-census-read-per-file at N=8/16 passed (2.845s package run). Real signal/helper cleanup controls passed (6.603s). Tracked-file Go catalog passed (0.795s).
- Shared `openRunLease` owns decode/flock observation; selection retains existing PID-generation checks and managed cancellation. IDs remain existing filename stems. Notices are advisory and preserve both runs. Help, README and atlas now state exact ID, by-name/all and duplicate behavior. Executable/help diff remains within the quick shell (96 added lines).
- Added opt-in `SDLC_RUNID_CLI` to reuse the private real-queue fixture against the built executable, with local Git/GitHub transport stubs; this does not touch actual fleet checks or :0 tools. Committed-build proof follows, then the recorded suite and close review.
- Hub changed the queue: #520 is assigned to ariadne:3. After #510 lands, report to ops:0 and wait for the next item; do not claim or plan #520 here.

- Committed-build proof: own bin/sdlc reports vcs.revision=18e7cda53cfbfcee5bfe99fd75005b94873d113d, vcs.modified=false. SDLC_RUNID_CLI against this executable passed the complete two-real-run FIFO/cancellation/refusal regression (5.960s package run), with private capacity and local transport fixtures. Output: /Users/xianxu/workspace/worktree/ariadne-slot2/p510-live-proof.txt. No fleet jobs or :0 tools were touched. Implementation acceptance is complete; recorded full suite and close review remain required before landing.

- Hub accepted the isolated real-CLI proof as sufficient (authenticated delivery e0f42988-6984-4ebc-99f5-d43218e1cbbd); no separate live check or :0 staging. First close returned SHIP/high, no findings. Reviewer independently passed the new controls and observed a 120s timeout in unchanged TestLocalCheckCancelOnlyOwnRun cleanup (already-consumed result channel); the implementation matrix passed with a larger budget, and the full suite will retain the normal budget.
- Landing integrated current main but found README/help conflicts with #509 scoped checks. Preserved main's scoped-check material and this issue's exact-ID cancellation documentation; executable files merged automatically. Cancelled only this checkout's single obsolete queued suite (still waiting, no acceptance result), then queued integrated-head verification before the required narrow re-close.

## Revisions

- 2026-10-21: corrected the discovery premise: lease IDs currently exist only in filenames, not in a `local-check status` command. Expose them in run/notice output so the requested exact selection is usable; no new status command or automatic replacement.

- 2026-10-21: kept the checklist on implementation acceptance; the required recorded suite and close/landing gates are separate workflow steps, avoiding a gate becoming its own unchecked prerequisite. A final-head recorded suite pass remains mandatory before landing.
