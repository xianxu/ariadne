---
id: 000322
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'cdd56d5767c3cd2531e60781b3aa69206a83b2d7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T17:46:02-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3f11ad44", done: "a58b67bd"}
---

# merge: landing instance-conformance gate validates main-tip..worktree, not the branch's own changes

## Problem

`sdlc merge`'s landing path runs the instance-conformance gate (#124) as
`validateChangedInstancesFn(pr.BaseOID, "", …)` (`cmd/sdlc/landing.go`). `pr.BaseOID` is
the PR base, which is **main's current tip**, and `""` is the working tree. So the window is
main-tip..worktree, not the branch's own changes. Every file main changed since the branch
point reads as "changed" by the branch. An issue that main archived after the branch was cut
shows up as *added* (it is still in the branch's tree), and it gets validated. If that file
does not conform (for example an empty `## Plan`), the merge refuses on a file the branch
never touched.

Hit landing ariadne#189 (PR #180): it refused on `000317-*.md`, which main had archived.
The workaround was to merge main into the branch. Every slot whose branch predates an archive
of a nonconforming issue is exposed (ariadne:1's #320 and ariadne:4's #183 at filing).

The legacy merge path (`merge.go`) uses `gitx.DiffBase()`, which is already the branch
point, and `sdlc push` is on main, so only the landing path is wrong.

## Spec

- The landing instance gate validates the branch's own changes:
  `merge-base(main, head)..head`, where main is the pinned `targetMain` that landing already
  fetched and head is the pinned branch head being landed (not the working tree).
- The merge-base comes from the same probe the landing duplicate-id gate already uses, one
  helper rather than a second copy (`ARCH-DRY`). Unavailable merge-base fails closed.

## Done when

- A branch cut before main archives a nonconforming issue lands without that file being
  validated. A regression test drives the landing instance gate on a real git fixture and
  fails against the old `pr.BaseOID`..worktree window.
- A nonconforming issue the branch itself changes still refuses (the gate keeps its teeth).
- The duplicate-id gate and the instance gate share one merge-base probe.

## Plan

- [x] Failing test: real-git fixture where main archives a nonconforming issue after the
      branch point; the landing instance gate must pass
      (`TestLandingInstanceGate_IgnoresMainsChangesSinceBranchPoint`). Second case: the
      branch's own nonconforming issue refuses
      (`TestLandingInstanceGate_RefusesBranchsOwnNonconformingIssue`). No merge-base fails
      closed (`TestLandingInstanceGate_FailsClosedWithoutMergeBase`)
- [x] `landingMergeBase` helper shared with `runLandingDuplicateGate`;
      `runLandingInstanceGate(mainOID, head, …)` validates `merge-base..head`
- [x] `landing.go` calls it with `targetMain` and the pinned head. The call site is pinned
      by `TestLandingPassesBranchWindowToInstanceGate`: `runMerge` after main moved on, and
      the gate must receive branch point..PR head
- [x] Mutation-checked: reverting the call site fails the call-site test; diffing from
      main's tip inside the gate fails the archive test
- [x] Atlas (`workflow/sdlc-binary.md`, validategate entry)
- [x] Full `make test`: green except the sandbox-blocked processgroup `/bin/ps` test, then close

## Log

### 2026-10-10
- 2026-10-10: closed — Regression: TestLandingInstanceGate_IgnoresMainsChangesSinceBranchPoint (a real-git branch cut before main archives a nonconforming issue lands without that file being validated); it fails when the gate diffs from main tip. TestLandingPassesBranchWindowToInstanceGate runs runMerge after main moved on and asserts the gate gets branch point..PR head; it fails when the landing.go call site is reverted. The branch-own nonconforming issue still refuses; no merge-base fails closed. Targeted landing/merge/duplicate tests green. make test: all 12 cmd/sdlc shards green; only processgroup TestCancellationKillsDescendants fails, because the sandbox blocks /bin/ps (unrelated); review verdict: SHIP
- Filed from ariadne#189's landing, at the TL's direction (ops:0); fix before #202.
- Targeted: `TestLanding|Duplicate|TestValidateChanged|TestRunMerge` all green (423s).
- head is the pinned commit, not the working tree. `readIssueFileFn` still reads the
  checkout, which `revalidateLanding` holds at that same head.
