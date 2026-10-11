---
id: 000322
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '75acd5e091b68c0c2072b82d29c2ac56afe1d2a1' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Failing test: real-git fixture where main archives a nonconforming issue after the
      branch point; the landing instance gate must pass. Second case: the branch's own
      nonconforming issue refuses
- [ ] `landingMergeBase` helper shared with `runLandingDuplicateGate`;
      `runLandingInstanceGate(mainOID, head, …)` validates `merge-base..head`
- [ ] `landing.go` calls it with `targetMain` and the pinned head
- [ ] Targeted tests, full `make test`, close

## Log

### 2026-10-10
- Filed from ariadne#189's landing, at the TL's direction (ops:0); fix before #202.
