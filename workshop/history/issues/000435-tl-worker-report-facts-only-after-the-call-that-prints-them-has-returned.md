---
id: 000435
status: done
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c45421d6166550d2f5bbcc055006d33cd2bad761' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T20:37:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c5ec2b93", done: "6acb0d5f"}
actual_hours: 0.03
---

# tl-worker: report facts only after the call that prints them has returned

## Problem

On 2026-10-17 parley.nvim:4 reported a landing with a merge SHA that never existed. It sent the report in the same parallel tool batch as the git call that prints the SHA, so the report was written before the output existed and the SHA was typed, not read.

## Spec

tl-worker help, under reporting: a report that names a SHA, PR number, test count or other fact goes out only after the call that prints that fact has returned. Never send it in the same parallel batch as that call. Copy facts from tool output; never type them.

## Done when

- `tl-worker --help` carries the rule, phrase-guarded by its rules test.

## Plan

- [x] Add phrase guards for the fact categories, returned-call ordering, and the parallel-batch prohibition; confirm they fail against current help.
- [x] Add a short rule under tl-worker REPORTS, reusing the existing help source (ARCH-DRY), then pass focused tests and built-help inspection.
- [x] Run the recorded exact-head suite.

## Log

### 2026-10-17
- 2026-10-17: closed — Four phrase guards failed against old help and pass with the REPORTS rule in both invocation forms; make -k build passes; built tl-worker help shows the rule; exact-head sdlc-local/suite passed 12 shards, 1,584 cmd/sdlc tests and 67 other packages on 743550c0d70662d74d683e262ed653b3d699adda (log sha256:b5f4773d2fd2). This help wording adds no new architectural surface, so no atlas map change applies.; review verdict: SHIP

The issue's narrow wording is the design: keep the existing EVIDENCE hash rule and add one REPORTS paragraph that applies to every reported fact. A warning only about hashes would miss PR numbers and test counts; a new reporting mechanism would add needless state. The help test will guard both the wait-for-return rule and the ban on pairing report and evidence call in one parallel batch.

Four phrase guards failed against old help in both invocation forms, then `go test ./cmd/tl-worker -count=1` passed after the REPORTS addition. `make -k build` passed, and the built tl-worker help shows the rule.

The recorded `sdlc local-check run --name suite` passed on pushed head `743550c0d70662d74d683e262ed653b3d699adda`: 12 shards, 1,584 cmd/sdlc tests, 67 other packages, zero failures; log sha256 `b5f4773d2fd2`. PR #295 was opened as a draft for CI overlap. Close review and merge remain.
