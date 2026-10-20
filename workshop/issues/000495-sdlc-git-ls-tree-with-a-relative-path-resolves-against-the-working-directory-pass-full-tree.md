---
id: 000495
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1c9d5edfe3a7b7dbe46d9581e816c2981680ca99' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc: git ls-tree with a relative path resolves against the working directory (pass --full-tree)

## Problem

`git ls-tree <rev> <path>` resolves `<path>` relative to the working directory unless `--full-tree` is passed. Found in #475: `sdlc landing status` from a subdirectory read "trains: off" because `readLandingSchedule` listed `.sdlc/landing.json` relative to the subdirectory. #475 fixed `readLandingSchedule` and `readLocalChecks`. Other sdlc call sites list a root-relative path without `--full-tree`, for example:
- `cigate.go:359` (localGatesWorkflow) and `cigate.go:372` (`.github/workflows`);
- `observe.go:117` and `observe.go:259`;
- `abandon.go:384` and `abandon.go:683`;
- `actualinputs.go:707`;
- `issuemigrate.go:538` and `issuemigrate.go:662`;
- `publishgate.go:471`.

Each is correct only while git runs at the repository root.

## Spec

Every `ls-tree` that names a root-relative path passes `--full-tree`, or runs under `git -C <top level>`. A Go lint rule in `pkg/golint` flags a new `"ls-tree"` call that lists a path without `--full-tree` (the mistake-class rule).

## Done when

- Every call site above is root-relative from any working directory, each covered by a test or the lint.
- The lint rule exists, citing this issue and #475.

## Plan

- [ ]

## Log

### 2026-10-20
