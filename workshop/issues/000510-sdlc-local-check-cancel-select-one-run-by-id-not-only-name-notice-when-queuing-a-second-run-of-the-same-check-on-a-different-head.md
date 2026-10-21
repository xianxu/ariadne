---
id: 000510
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ad6900df69f7e31efadc016f4815603a995ed095' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T02:17:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Add runtime CLI regressions for exact cancellation of one of two real queued runs, retained sibling FIFO, and foreign/invalid selectors.
- [ ] Expose the existing lease ID; add exact-ID cancellation and same-check/different-head notices using the shared lease reader and managed interruption path.
- [ ] Verify targeted cancellation/notice controls, update help/README/atlas and run the recorded full suite before landing.

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

## Revisions

- 2026-10-21: corrected the discovery premise: lease IDs currently exist only in filenames, not in a `local-check status` command. Expose them in run/notice output so the requested exact selection is usable; no new status command or automatic replacement.
