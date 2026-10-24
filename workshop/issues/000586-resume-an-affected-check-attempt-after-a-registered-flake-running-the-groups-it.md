---
id: 000586
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '035e23c253b472fcde9b528908b2873759d506ce' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T22:52:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "36d108a2", done: "410fe668"}
---

# Resume an affected-check attempt after a registered flake, running the groups it never reached

## Problem

## Spec

**An affected-check attempt can't be resumed after a registered flake; only suite attempts can (pair:9 on pair#841,
10-23).** A pair `--pre-close` affected attempt failed on a single test, which pair:9 then proved was a flake and
registered (pair#854). The supported same-head resume refused: resume only supports suite attempts, and the original
affected run had stopped at the Go failure before running the script groups. So the only path was a docs-only commit
and a fresh affected run.

Fix: support resuming an affected attempt the same way as a suite attempt: rerun the registered failing tests, then
run whatever selected groups the original attempt never reached (the script groups here), and record the combined
result for the same head. Refuse, as today, when an unregistered failure remains.

## Done when

- An affected attempt that failed only on registered flakes, and stopped before some groups, resumes: reruns those
  tests, runs the unreached groups, and records a pass for the head (test).
- An unregistered failure still refuses the resume (test).

## Plan

Durable plan: `workshop/plans/000586-resume-an-affected-check-attempt-after-a-registered-flake-running-the-groups-it-plan.md`
(full flow; design to the hub for approval before change-code).

- [x] Shared `affectedRerunBatch` (flake reruns + affected command with empty Go packages and all selected groups)
- [x] Inline known-flake step runs the groups (pre-existing gap; `TestAffectedKnownFlakeReruns` changes)
- [x] Extract `planAffectedSelection` from `runAffectedCheck`
- [x] Version 5 resume link/receipt for affected (`selection` required); proof, retention, old-reader fence
- [x] `local-check resume --name affected`: recompute selection, v5 link, batch, scope on the record; tests per Done when
- [x] Help + atlas; targeted tests; pre-close check

## Log

### 2026-10-23
- Design mapped (Explore digest + reading `localcheck_resume*.go`): resume is suite-only in ~13 places; a v3 link with
  `check: affected` would make an older sdlc mark the whole store snapshot incomplete (decodeResumeEvidence fails and
  decodeUnsupportedProof needs version > 4), so affected links are version 5 (ARCH-EVOLVE). Nothing records which
  groups ran, so the resume reruns every selected group. Same-class gap in the inline step folded in (ARCH-PURPOSE).
- Implemented per the approved plan (TL approval 10-23). Commits: ce9d89b5 (shared batch + inline fix), 2c23ef22
  (planAffectedSelection), dc0a6e93 (v5 link), f3b2ba71 (affected resume). Discovery: the environment retirement step
  compared links with `*decoded != l`, which turns false-unequal once a link holds a pointer (`Selection`); it now
  compares deeply (caught by TestResumeAffectedRunsFlakesThenGroups). Retention needed no change: it already binds
  `p.Check == l.Check` (TestResumeRetentionAffectedBundle).
- Old-reader proof is at the decode level (TestOldReaderFencesAffectedLinkPerHead): the pre-#586 fallback
  `decodeUnsupportedProof` admits v5 and rejects a v3 affected link. The frozen-binary test stays opt-in (SDLC_PRE_TYPED_BIN).
- Sandbox-only failures seen: host boot identity and signal-process tests; run unsandboxed, green.
