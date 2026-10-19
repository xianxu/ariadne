---
id: 000455
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f3d5213e60e4d39799e0fc053524a8169a360675' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T21:13:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] sidedoc.RebaseLinks + AnswerName
- [ ] link accepts only side-doc answers
- [ ] SessionSync seam: ReadAll, Apply, Registrations, Rename (fake + conformance)
- [ ] planMigration (pure, golden + conflicts)
- [ ] migrate-answers executor, CLI, crash-resume test
- [ ] help, atlas

## Log

### 2026-10-18

- Claimed by ariadne:4 on the hub's dispatch. Finding: help.md has said since #412 that answers live in `tl-status-<op>.side/`, but `link` never enforced it, so the hub kept using qa/. Root cause is the unenforced rule (ARCH-PURPOSE); validation is part of the fix. Survey of ops' qa/ in the plan. Seven decisions posted to the hub.
