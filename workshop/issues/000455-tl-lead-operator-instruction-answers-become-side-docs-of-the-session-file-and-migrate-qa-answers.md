---
id: 000455
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours: 3.36
collaboration-mode: collaboration
card_mirror: 'e944230e6b60b9482d4592d1c8a75b63c7680382' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T21:13:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers)

## Problem

The hub answers each Operator instructions bullet in a separate doc under `teams/<team>/qa/<operator>/`, linked from the bullet with `tl-lead status instructions link`. parley's attention finder lists an entry's children by scanning its `.side/` folder for `type: side-doc` files, so these answers never appear in the finder (operator, 2026-10-18; ops qa 2026-10-18-24-finder-answer-docs.md). The two conventions never met. The operator: "there might even be a task to migrate from qa/ to xxx.side/ folder for TL's session file?"; none existed.

## Spec

Answers to the session file become real side docs of it: `<session-file-stem>.side/NN-slug.md`, with frontmatter `type: side-doc`, `parent:` (relative path to the session file), `marker:`, `collaboration-mode`, `process-instruction`, and the backlink line under the title, the same contract parley's side docs use. `tl-lead status instructions link` accepts and validates that target; tl-lead's help and the TL skill say answers go there. A one-time migration command moves existing `qa/<operator>/` answer docs into the side folder: it rewrites their frontmatter, rewrites each `→ [answer](…)` link and the `discussions:` index in the session file, keeps side-doc subfolders (`*.side/`) attached, and moves registered files through `couch --session-sync rename`, so Couch tracking survives. It is dry-run first and refuses on any conflict. Gitignore follows wherever qa/ was ignored.

## Done when

- `link` produces side-doc answers; a new answer shows in parley's attention finder under the session file (live check).
- The migration moves the ops answers with every link and the discussions index intact, and Couch still tracks the registered ones (dry-run, then live).

## Plan

Full flow; durable plan: [000455 plan](../plans/000455-tl-lead-operator-instruction-answers-become-side-docs-of-the-session-file-and-migrate-qa-answers-plan.md). Tasks 1-6 there.

- [x] sidedoc.RebaseLinks + AnswerName
- [x] link accepts only side-doc answers
- [x] SessionSync seam: ReadAll, Apply, Registrations, Rename (fake + conformance)
- [x] planMigration (pure, golden + conflicts)
- [x] migrate-answers executor, CLI, crash-resume test
- [x] help, atlas

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.8 impl=0.08
item: smaller-go-module    design=0.1 impl=0.2
item: smaller-go-module    design=0.05 impl=0.12
item: smaller-go-module    design=0.1 impl=0.16
item: greenfield-go-module design=0.5 impl=0.32
item: smaller-go-module    design=0.2 impl=0.2
item: atlas-docs           design=0.05 impl=0.05
item: milestone-review     design=0.0 impl=0.16
design-buffer: 0.15
total: 3.36
```

Items, in order: the spec and plan with two plan-quality rounds; RebaseLinks and its fuzz test; link validation; the SessionSync seam widening with fake and conformance; the pure planner (greenfield); the executor and crash-resume test; help and atlas; the close review.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-18
- 2026-10-18: closed — Close review round 1 fixed in 7b3b597d: BR-1 resume reads registration and lock at old and new paths, crash test asserts every moved registered doc marked seen, new test refuses a lock at the destination; BR-2 gitignoreGap warns via git check-ignore, tested in a real repo; BR-3 index entries rebase from their own file; BR-4 fake MarkAllSeen per file and live conformance applies an occurrence-2 record (pass against real couch); BR-5 --from confined; BR-6 temp sweep. make test green on 7b3b597d (12 shards, 1657 cmd/sdlc tests plus 70 packages, 0 failed, wall 642s). Ops dry run on this build: refused only because the operator held the session file lock at that moment, as designed; no gitignore warning. Live check of link in the parley finder is the hub after landing.; review verdict: SHIP

- Claimed by ariadne:4 on the hub's dispatch. Finding: help.md has said since #412 that answers live in `tl-status-<op>.side/`, but `link` never enforced it, so the hub kept using qa/. Root cause is the unenforced rule (ARCH-PURPOSE); validation is part of the fix. Survey of ops' qa/ in the plan. Seven decisions posted to the hub.

- Implementation:
  - `RebaseLinks` fuzzing found an overlapping reference-definition and inline-link panic. It is fixed, and the failing input is kept as a regression seed.
  - The seam drops `ReadAll` from the plan: couch's whole-file hash is sha256[:8] hex of the bytes, so `Apply`'s hash comes from `fileHash` of a Peek. `read --all` would mark the operator's unread session-file edits seen. Live conformance pins the formula.
  - The read-only dry run against ops found the harness's empty `qa/xian-xu/.claude/.cc-writes` directory. A dot directory is now listed as left in place and never removed.
  - Outside files keep their mtimes too: a rewritten side doc elsewhere, such as `design.side/01-…`, would otherwise show in the finder as a new agent write.
  - All of this is under decision 4.
