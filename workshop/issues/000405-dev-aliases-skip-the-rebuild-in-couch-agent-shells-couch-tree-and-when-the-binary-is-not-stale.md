---
id: 000405
status: codecomplete
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '73ac54528d82d6db42fae1b5458531f71371be63' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T20:34:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "dc91444b", done: "bf1822d9"}
actual_hours: 0.53
---

# dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale

## Problem

Every call of a dev-aliases function rebuilds the binary in its owner's checkout, which is the :0 slot. That includes calls from agent shells in Couch slots. Agent sessions snapshot the operator's shell, so in each slot `couch`, `sdlc` and `weave` resolve to the dev-alias functions. Each agent call then writes `bin/X` into pair:0 or ariadne:0, the operator's staged checkouts that only the TL and the operator own. On 10-15 at 20:28, right after a Couch restart, ariadne:2's `couch --message-status` calls rebuilt `bin/couch` in pair:0. Agent sessions whose shell snapshot predates #355 also `rm -f` the binary before building, which is why the binaries briefly vanished during the restart. A sandboxed call can't write there, so its build fails, and the call fails with it.

Every call also pays a `go build`, even when nothing changed.

## Spec

The emitted function skips the build and runs the existing `OWNER/bin/X` when either of these holds:
- `COUCH_TREE` is set (an agent shell in a Couch slot), and `bin/X` exists; or
- `bin/X` is not stale: no file or directory outside `.git` and `bin` is newer than it, in the owner or in a local `go.mod` replace target outside it (revised in close rounds 1 and 2; see Revisions).

Otherwise it builds as it does today (dev-prepare, then a temp-directory build renamed into place, #355). A missing `bin/X` always builds. A skipped build also skips dev-prepare, which only feeds the build. The staleness scan is `find -newer -print -quit`: about 0.13s for a full scan of pair (7.6k files). A toolchain change alone doesn't count as stale.

Build-to-temp-then-rename is already on main (#355). The 20:28 `rm -f` came from a shell snapshot from before #355. The tests keep pinning it.

## Done when

- The emitted function skips the build when `COUCH_TREE` is set and `bin/X` exists. It also skips it when `bin/X` is newer than every file and directory outside `.git` and `bin`, in the owner and in each local `go.mod` replace target outside it. It builds when `bin/X` is missing, a file is newer, or a file was added, deleted or renamed. A scan error counts as stale. A built binary carries its build's start time, so an edit made mid-build makes the next call rebuild.
- `construct/scripts/test/dev-aliases.test.sh` runs the emitted function against a stub go and proves each case (stub-go call counts), plus the unchanged #355 rename and #299 dev-prepare shapes.
- The script's header comment states the skip rules. The atlas sentence about "fresh every call" is corrected wherever it appears.

## Plan

- [x] Emit the skip guard in `construct/dev-aliases.sh` and update its header comment.
- [x] Tests: behavioural cases with a stub go (COUCH_TREE with the binary present and missing, fresh, stale, dev-prepare skipped when fresh); update the exact-shape assertions.
- [x] Atlas wording: base-layer (the rule that :0's bin/ is rebuilt only by the operator, the TL, sdlc stage --build or make) and the issue-tracker rollout sentence. README needed no change.
- [x] Full `make test`: green, 12 shards, exit 0.

## Log

### 2026-10-15
- 2026-10-15: closed — Advisory Minor from the SHIP round fixed: the Spec bullet now states the scan as fresh_scan does (files and directories, owner and outside replace targets, excluding .git and bin). Text-only change to the issue file; dev-aliases.test.sh 65 passed. (fixed-then-ship); review verdict: SHIP
- 2026-10-15: closed — Round 2 fixes: BR-4, the scan drops -type f so directory times catch added, deleted and renamed files; new tests for an added file and a deleted file fail when -type f is restored (mutation-checked); the real pair tree still scans fresh. BR-5: Done when covers the replace targets, scan error and build-start stamp. Minor: the slot exception reads as an event (a slot writes to :0 only to build a missing binary, dev-prepare included) in the script header, base-layer and Revisions. dev-aliases.test.sh 65 passed. Full make test green on round 0 (12 shards); rounds 1 and 2 touch only the dev-aliases script and its test, and the construct-scripts runner passed 6 of 6.; review verdict: SHIP

- TL approved the plan and its consequence: slots run whatever :0 last built, and no slot rebuilds even when the binary is stale. The atlas states who rebuilds :0's bin/.
- #355 already builds to a temporary name and renames, so that part of the request was done. The 20:28 rm -f came from an agent shell snapshot from 10-13, before #355. Sessions keep their snapshot's functions until they restart.
- side-quest 22a5ca01: weave's .gitignore lines for the xx-universal-review lowering, missed in #399. They dirtied every slot after a refresh and blocked start-plan.
- the tests unset COUCH_TREE, because this slot's own environment would otherwise skip every build. File times are set with touch -t, so no case depends on mtime resolution. dev-aliases.test.sh: 55 passed.
- Close round 1 returned FIX-THEN-SHIP. Each finding was fixed as a class:
  - BR-1: the docs claimed a slot never writes to :0, but a missing bin/X builds from a slot too. Swept every instance: base-layer, issue-tracker, the script header and my fleet memory note. Each now states that one exception.
  - BR-2: the freshness scan missed local replace dependencies. fresh_scan now adds each local go.mod replace target outside the owner as a find root, with its .git and bin pruned. Targets inside the owner, like pair's ./third_party/vt, add none. Only the main module's replaces apply to a build.
  - Minor: a find error now counts as stale. A built binary takes the time its build started (touch -r from a .start stamp), so an edit made mid-build makes the next call rebuild. A mutation check confirmed the test fails without the stamp.
  - dev-aliases.test.sh: 63 passed.

## Revisions

- 2026-10-15, close round 1: the freshness scan now also covers local go.mod replace targets outside the owner. A find error counts as stale, and a built binary is stamped with its build's start time. The docs now state the one write a slot can make in :0: building a missing binary.
- 2026-10-15, close round 2: the scan counts directories too (no `-type f`), so an added, deleted or renamed file makes bin/X stale. Done when now covers round 1's replace targets, scan error and build-start stamp. The slot exception is now worded as an event: a slot writes to :0 only to build a missing binary, dev-prepare included. That replaces round 1's "the one write a slot can make".
