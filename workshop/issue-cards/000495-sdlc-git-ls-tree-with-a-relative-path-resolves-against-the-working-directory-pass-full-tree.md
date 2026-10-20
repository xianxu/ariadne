---
id: 000495
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
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
