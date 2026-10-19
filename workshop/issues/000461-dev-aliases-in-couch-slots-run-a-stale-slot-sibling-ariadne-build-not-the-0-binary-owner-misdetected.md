---
id: 000461
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '623d9cfb6f89f655b75526df0fa39ea2cd419a41' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T02:31:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9f33d2b4", done: "ad0e717f"}
---

# dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected)

## Problem

In a Couch slot, the `sdlc` command (and its siblings) is a shell function from `construct/dev-aliases.sh` (ariadne#57), captured in the agent's shell snapshot. It shadows `/Users/xianxu/workspace/ariadne/bin` on PATH. Its OWNER detection resolves to the slot's sibling checkout (e.g. `~/workspace/worktree/pair-slot2/ariadne`) instead of the operator's `:0` checkout (`~/workspace/ariadne`) that the dev-aliases header says it should use. Because `COUCH_TREE` is set, the #405 rule skips the rebuild, so the slot runs a binary built whenever that sibling was last built, forever. Found by pair:2 on 2026-10-19: its `sdlc` was built Oct 18 18:19 from 3a0012b4, before #394, #452, #454 and #453, and refused #394's `flake_tests` card field. Every slot is likely affected, which explains the night's old-runner card refusals, and means slots have not been running the gates and rules the TL rebuilt into ariadne:0.

## Spec

Fix OWNER resolution in dev-aliases so that in a Couch slot (any worktree under `worktree/<repo>-slotN/`) the owner is the operator's `:0` ariadne checkout, which the TL keeps built; never a slot-sibling checkout. Keep the no-rebuild-in-Couch rule (#405) but make it apply to the right binary. Add a guard: `sdlc --version` or a startup line names the binary's path and revision, and a slot-side check warns loudly when the binary in use is not the `:0` build. A regression covers OWNER detection from a slot worktree layout, a primary checkout and a plain clone.

## Done when

- start-plan does not request a rerun merely because a slot-private dependency
  moved; it refuses when refresh advances the running executable's own source
  beyond its embedded revision, including partial-refresh failure.

- In every slot, a bare `sdlc` runs `/Users/xianxu/workspace/ariadne/bin/sdlc` (verified across live slots after a weave refresh or a relaunch).
- The regression covers the three layouts; the version line shows path and revision.

## Plan

- [x] Normalize numbered-environment alias owners to verified fleet :0 peers;
  retain primary/plain clone behavior, #405's skip rule and missing-bin exception.
- [x] Add executable path/revision version output and a Couch wrong-owner warning.
- [x] Make catch-up refuse only if the running binary's own source moved and
  its embedded revision differs; a private sibling move alone permits planning.
- [ ] Run real-layout/sentinel and catch-up regressions red/green, focused Go and
  script checks, recorded full suite, close review and hub-coordinated live checks.

## Log

### 2026-10-19

## Revisions

### 2026-10-19 — history and approved ownership model

The hub approved central :0 shell aliases after reviewing #57/#405 and catch-up.
Private substrate ownership is separate: weave still refreshes/composes private
siblings. #405 explicitly says slots run whatever :0 last built; the alternative
of rebuilding private siblings on invocation would revise that policy and add
writes. Missing or mismatched expected primaries refuse rather than selecting
stale siblings. Match origin using the existing conservative identity function
(ARCH-DRY), with three-layout real-Git regressions and sentinel binaries proving
which executable runs without building. Preserve --list's machine output and
all unrelated duplicate-origin behavior.

The hub added precise catch-up refusal to scope. Existing catchUpBeforeWork
compares only before/after checkout HEADs, not executable identity, so its old
built-before claim is unjustified for central aliases. Observe actual executable
path and embedded VCS metadata once, reuse buildRevision, and compare against
its source checkout after refresh. Keep refresh failures/freshness checks strict;
only a successful refresh with no obsolete running source may proceed directly.
The startup warning/version diagnostic is read-only; unknown metadata is named,
never invented. Existing workspace identity supplies fleet location for warning.
No new persisted format, cache, credential or service is introduced (ARCH-EVOLVE).

Actions are invoke, regenerate aliases/relaunch, and TL rebuild :0. Correct
primary + existing binary runs without build; absent/mismatched primary refuses
and names repair; wrong running path warns with actual/expected paths; advanced
own source with mismatched revision refuses and requests a fresh build/rerun.
No lifetime lock is needed for read-only diagnostic observations; existing
atomic binary rename owns build publication. Alias generation scans candidate
repos once and dedupes canonical paths; invocation retains #405's constant skip
in Couch, with no new filesystem-wide scan. Tests count canonical-owner scans
at N/2N if that normalization introduces repeated traversal (ARCH-CONSTRAINTS).

### Log evidence

- Claim/start-plan completed through the explicitly requested :0 binary. #459
  stays parked on its clean committed branch/PR319 awaiting operator verification.
- Hub approved the plan before code (receipt e15f2eb0-3492-40d9-9a91-99981138410b).
  History review confirms #405's explicit TL decision and catchup_test's unchanged
  same-process rerun. Current own-slot --list resolves sdlc to this worktree,
  reproducing the ownership mismatch without rebuilding anything.

- Initial real-Git layout regression failed because slot aliases selected the
  stale sentinel. After owner normalization, primary/slot/plain clones select
  the expected sentinels without any Go build; absent/mismatched primaries
  refuse with no emitted aliases. Existing dev-aliases tests pass 82/82.
- Go targeted version/catch-up/start-plan regressions passed before a final
  process-local metadata cache and outside-chain source test. Combined clear
  implementation exceeds the 100-line quick-review shell, so close will use
  the binary's automatic full-review upgrade rather than compressing the code.

- Final targeted Go tests pass (52.117s broad StartPlan/catch-up/version run;
  8.824s final narrow run including outside-chain source movement). Metadata is
  observed once per process, with an injectable test seam and no persisted cache.
- Shell scan-count regression with N=2 and 2N=4 candidate copies proves the
  canonical owner's cmd tree is scanned once. New layout test passes alongside
  all 82 existing dev-aliases assertions; shell lint passes.
- Built only this slot's ignored bin/sdlc for a read-only --version smoke. It
  reports its actual slot path and embedded revision/modified bit, and warns
  with expected /Users/xianxu/workspace/ariadne/bin/sdlc. The live generator's
  --list selects that primary for every Ariadne tool. Workflow verbs continue
  using the explicitly requested :0 binary until landing.
