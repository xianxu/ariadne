---
id: 000405
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'ae7347889e1a24d8fc21f9860810cb7e98ceb91b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T20:34:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "dc91444b", done: "bf1822d9"}
---

# dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale

## Problem

Every call of a dev-aliases function rebuilds the binary in its owner's checkout, which is the :0 slot. That includes calls from agent shells in Couch slots. Agent sessions snapshot the operator's shell, so in each slot `couch`, `sdlc` and `weave` resolve to the dev-alias functions. Each agent call then writes `bin/X` into pair:0 or ariadne:0, the operator's staged checkouts that only the TL and the operator own. On 10-15 at 20:28, right after a Couch restart, ariadne:2's `couch --message-status` calls rebuilt `bin/couch` in pair:0. Agent sessions whose shell snapshot predates #355 also `rm -f` the binary before building, which is why the binaries briefly vanished during the restart. A sandboxed call can't write there, so its build fails, and the call fails with it.

Every call also pays a `go build`, even when nothing changed.

## Spec

The emitted function skips the build and runs the existing `OWNER/bin/X` when either of these holds:
- `COUCH_TREE` is set (an agent shell in a Couch slot), and `bin/X` exists; or
- `bin/X` is not stale: no regular file under the owner (outside `.git` and `bin`) is newer than it.

Otherwise it builds as it does today (dev-prepare, then a temp-directory build renamed into place, #355). A missing `bin/X` always builds. A skipped build also skips dev-prepare, which only feeds the build. The staleness scan is `find -newer -print -quit`: about 0.13s for a full scan of pair (7.6k files). A toolchain change alone doesn't count as stale.

Build-to-temp-then-rename is already on main (#355). The 20:28 `rm -f` came from a shell snapshot from before #355. The tests keep pinning it.

## Done when

- The emitted function skips the build when `COUCH_TREE` is set and `bin/X` exists, and when `bin/X` is newer than every owner file. It builds when `bin/X` is missing or a file is newer.
- `construct/scripts/test/dev-aliases.test.sh` runs the emitted function against a stub go and proves each case (stub-go call counts), plus the unchanged #355 rename and #299 dev-prepare shapes.
- The script's header comment states the skip rules. The atlas sentence about "fresh every call" is corrected wherever it appears.

## Plan

- [x] Emit the skip guard in `construct/dev-aliases.sh` and update its header comment.
- [x] Tests: behavioural cases with a stub go (COUCH_TREE with the binary present and missing, fresh, stale, dev-prepare skipped when fresh); update the exact-shape assertions.
- [x] Atlas wording: base-layer (the rule that :0's bin/ is rebuilt only by the operator, the TL, sdlc stage --build or make) and the issue-tracker rollout sentence. README needed no change.
- [x] Full `make test`: green, 12 shards, exit 0.

## Log

### 2026-10-15

- TL approved the plan and its consequence: slots run whatever :0 last built, and no slot rebuilds even when the binary is stale. The atlas states who rebuilds :0's bin/.
- #355 already builds to a temporary name and renames, so that part of the request was done. The 20:28 rm -f came from an agent shell snapshot from 10-13, before #355. Sessions keep their snapshot's functions until they restart.
- side-quest 22a5ca01: weave's .gitignore lines for the xx-universal-review lowering, missed in #399. They dirtied every slot after a refresh and blocked start-plan.
- the tests unset COUCH_TREE, because this slot's own environment would otherwise skip every build. File times are set with touch -t, so no case depends on mtime resolution. dev-aliases.test.sh: 55 passed.
