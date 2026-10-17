---
id: 000419
status: codecomplete
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours:
card_mirror: 'c7026676c7beebf25af73ebfd0a66824254219cf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T21:39:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.60
---

# Makefile.workflow: the generic Go build publishes bin/<name> in place; readers see missing or partial binaries

## Problem

## Spec

From pair#624, 2026-10-16. pair:0's `bin/couch` went missing for seconds at 15:40, 16:06 and 16:08 while Couch and the slots were calling it, so sends failed. pair:7 found that `Makefile.workflow`'s generic Go build writes each binary straight to `bin/<name>`, because Go's final step copies onto the target path. A private probe that replaced a live binary saw 1 missing-file and 418 partial-file observations. pair#624 fixes pair's own build; this is the substrate half, shared by every ariadne-style repo.

The generic build should compile to a temp file next to the target and `rename` it over `bin/<name>`, so a reader sees either the old binary or the new one, never a missing or partial file.

## Done when

- The generic build publishes each binary by building to a temp file and renaming it.
- A test with a paused or failing compiler shows `bin/<name>` stays the old, complete binary until the rename.
- A failed build leaves the old binary and no temp files.

## Plan

One publisher, `WF_GO_PUBLISH` in `Makefile.workflow`, in the same shape as pair#624's `scripts/build-go-atomic.sh`
(pair:7) and `construct/dev-aliases.sh`'s rebuild. It makes a staging dir `mktemp -d "<target dir>/.go-build.XXXXXX"`
(same filesystem, so the rename is atomic; one per invocation, so concurrent builds don't collide), runs
`go build -o <staging>/binary ARGS`, then `mv -f` over the target. A subshell traps EXIT to remove the staging dir, and
INT/TERM to exit through it. Callers:
- `make build`'s generic loop (the issue).
- ariadne's own `Makefile` `tools`/`weave-build` targets: the same in-place `go build -o bin/X` into :0's `bin/`, which
  every agent alias runs. It's the same defect and two lines (ARCH-DRY: one publisher, not a second copy).

- [x] `WF_GO_PUBLISH` + both callers.
- [x] Test (`cmd/sdlc/makebuild_publish_test.go`, a fake `go` on PATH that writes a partial `-o` file): paused build →
      `bin/x` still the old complete bytes, then the new ones after release; failing build → old bytes, no
      `.go-build.*` left; a temp dir is never left after success.

## Log



- 2026-10-17: closed — Round-1 SHIP minors fixed: stale .go-build sweep (TestMakeBuildSweepsStaleStaging) and atlas note; go test -run TestMakeBuild ./cmd/sdlc, pkg/shelllint, scripts/test trimpath and portable-makefile.test.sh pass. Full suite green earlier at 2420f13a. (fixed-then-ship); review verdict: SHIP
- 2026-10-17: closed — make test green at 2420f13a (12 shards, 1525 cmd/sdlc tests + 66 packages, 0 failed). New TestMakeBuildPublishesByRename / FailureKeepsOldBinary / InterruptKeepsOldBinary (fake go writing a partial -o file) fail on the old in-place build and pass now; real make tl-lead-build leaves no .go-build dir; pkg/shelllint and portable-makefile.test.sh pass.; review verdict: SHIP
### 2026-10-16
- The tests fail against the old in-place build (`bin/x` reads "partial" mid-build and after a failure), and pass with
  the publisher.
- Added an interrupt test (SIGTERM to the build's process group). make exits on the signal without waiting for its
  recipe shell, so the staging dir is removed by that shell's EXIT trap just after make returns (~20 ms). The test waits
  for it, bounded at 5 s.
- `pkg/shelllint` (#342) refused `mktemp -d "$(dirname …)/…XXXXXX"`. The staging dir must sit beside the target for the
  rename to be atomic, so the directory is assigned to a variable first (`p=$(dirname "$out")`, then
  `mktemp -d "$p/.go-build.XXXXXX"`), which is the rooted-template form the lint accepts.
- First full suite: `scripts/test` failed in 2 fixtures that assumed an in-place build. `TestOwnerMakeCommandsUseTrimpath`
  copied only `Makefile`, so `WF_GO_PUBLISH` was empty, and its fake `go` wrote no output. `portable-makefile.test.sh`'s
  fake `go` only `touch`ed its output; in place, that kept the old 0755 mode, but a renamed new file is 0644, and
  `sdlc-install.sh` checks `-x`. Real `go build` output is executable, so the fake now `chmod +x`s. The trimpath test
  now copies `Makefile.workflow`, writes the output, and matches the staging path.
- Close review (round 1): SHIP, 2 Minors, both fixed in-round. `WF_GO_PUBLISH` sweeps `.go-build.*` dirs over an
  hour old (a SIGKILL'd build's; no build runs that long), tested by `TestMakeBuildSweepsStaleStaging`, which also
  checks a recent dir is kept. The publish-by-rename guarantee is now in `atlas/workflow/setup-and-replication.md`.
