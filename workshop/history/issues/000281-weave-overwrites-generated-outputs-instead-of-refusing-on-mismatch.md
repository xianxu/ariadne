---
id: 000281
status: done
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-12
estimate_hours:
card_mirror: 'dccbc0d43cc6489440648e9b2d13032527953b16' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:44:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.07
---

# weave overwrites generated outputs instead of refusing on mismatch

## Problem

`weave compile` refuses with `apply: preserving authored replacement at <path>`
whenever a recorded output's bytes differ both from what weave last wrote
(`construct/generated/weave/ownership.json`) and from what it wants to write
(`plan/ownership.go`, `ApplyManaged`). The refusal has no recovery path:

- `weave compile` has no override flag, and the message names no next action.
- The only ways out are deleting the file, which loses the "protected" edit
  anyway, or recreating the exact recorded bytes by hand.

**Observed in parley.nvim slot2, 2026-10-01.** ariadne#277 added the `claimant`
card field. Compile regenerated the committed
`construct/generated/vocabulary/issue.json` and recorded the new bytes. Later
something put the stale HEAD version back. The likely cause is a git
restore/stash to get the clean tree `weave refresh` requires; it happened twice,
at 17:52 and at 18:10. From then on, `weave compile` and `weave refresh` both
refused on a file nobody had edited by hand. The way out was committing the
regenerated JSON (parley PR #221). Nothing in the error pointed there.

The check came from #239 ("one verb, two ownership classes"), where weave used
to clobber a repo's own root `Makefile`. #239 itself fixed that by no longer
generating the root Makefile. Repo-specific behaviour now goes in overlay sources
weave never writes (`AGENTS.local.md`, `Makefile.local`). So for files weave does
generate, an edit belongs in an overlay, and refusing protects nothing: it only
delays losing the edit. The refusal also leaves `.gitignore` half-migrated
(#265 item 1).

## Spec

- **Generated outputs always converge.** When a recorded output's bytes differ
  from the last write, weave overwrites it and prints a warning. The warning
  names the path and says where the edit belongs: the overlay for that output
  family, or the layer source for vocabulary and datatype exports. Weave never
  refuses on content mismatch for a path it owns in the inventory.
- **Never lose bytes silently.** Before overwriting a mismatched output, copy the
  old bytes to a backup beside the inventory (e.g. `construct/generated/weave/
  displaced/<path>`). Each successful compile keeps only the latest displaced copy
  per path, so the backup's lifecycle is named (ARCH-FUNERAL).
- **The inventory keeps one job: safe retirement.** Weave removes an output it no
  longer generates only when the file still matches the recorded identity.
  Otherwise it leaves the file, and warns once and names it.
- **Unowned occupied destinations** are a different case: a path never in the
  inventory, holding a file weave did not create. Today's refusal stays for
  these, but the message must give the next action (move the file aside, or
  declare it an overlay).
- **Vendored outputs.** Some generated outputs are also committed. parley.nvim
  commits `construct/generated/vocabulary/issue.json` for fresh clones (parley
  #208). For these, a compile that changes a tracked output reports "regenerated
  tracked output <path>: commit it" instead of leaving an unexplained dirty
  tree.

## Done when

- A compile over a recorded output whose bytes were reverted to an older weave
  write (the parley case: HEAD's stale copy) overwrites it, warns, keeps a
  displaced copy, and exits 0 (regression test).
- A compile over a recorded output with an arbitrary hand edit does the same,
  and its warning names the overlay or layer source where the edit belongs.
- Retirement still deletes only matching files; a mismatched retired output is
  left in place with a warning (test).
- An unowned occupied destination still refuses, and the message states the next
  action (test).
- A compile that changes a tracked output prints the commit hint (test).
- `atlas/workflow/weave.md` describes the convergence rule. "preserving
  authored replacement" no longer appears in the code.

### Design (2026-10-12)

`plan.ApplyManaged` returns a `ManagedReport` instead of `[]string`, and
`compilePrepared` prints it. The decision per wanted output, after the existing
cross-scope and parent checks:

| destination | today | now |
|---|---|---|
| absent, or matches the wanted identity, or matches a recorded one | write | write (unchanged) |
| recorded, mismatched, file or symlink | refuse | move it to `construct/generated/weave/displaced/<path>`, write, warn with its edit home |
| recorded, mismatched, a directory | refuse | refuse, naming the next action (move it aside) |
| unrecorded, data scope or staged output | refuse | refuse, naming the next action and the edit home |
| unrecorded, other artifacts | overwrite | overwrite (unchanged, out of scope) |

- Displacement is a rename, so bytes, mode and link targets survive exactly; a
  directory is never moved, so no data tree ends up in a backup that a later
  displacement replaces.
- Retirement: unchanged recorded outputs retire as before. A recorded output that
  is no longer produced and no longer matches is reported as `Abandoned` and
  warned once; it drops out of the inventory that same compile, so it is not
  warned again.
- `Regenerated` lists outputs that existed with different content and were
  rewritten. `compilePrepared` intersects it with `git ls-files` (whole index, no
  pathspec, so no globbing or argument limit) and prints "regenerated tracked output
  <path>: commit it". A git failure prints one warning; it does not fail compile.
- Edit home is a pure function of the path (`plan.EditHome`): entry files →
  `AGENTS.local.md`; `settings.json` → sibling `settings.local.json`;
  `construct/generated/…` → the layer source it is generated from; skill links →
  the skill's source in its owning layer; else the owning layer's manifest
  source or a local overlay.
- ARCH-FUNERAL: `displaced/` is created by a converging compile; the operator is
  the last to need a copy; the next displacement of the same path replaces it,
  so it holds at most one copy per generated path (bounded by the outputs
  themselves) and grows by zero on a compile without a mismatch. It sits under
  the reserved, ignored `construct/generated/weave/`, which prune skips.

## Plan

- [x] Tests first (`ownership_test.go`): reverted-to-older-write and hand-edit
  both converge, warn, keep the displaced copy, exit 0; a second displacement
  replaces the first; retired mismatched output left and reported once; recorded
  directory and unowned data/staged destinations still refuse with a next action;
  update the existing refusal tests (`TestManagedProtectsEditedActiveOutput`,
  staged/publication permission-edit and link cases) to the convergence rule.
- [x] `EditHome` unit test (each family).
- [x] Implement `ManagedReport`, displacement and abandoned reporting in
  `ownership.go`; remove the "preserving authored replacement" refusal.
- [x] `compilePrepared` prints displaced/abandoned warnings and the tracked-output
  commit hint; compile-level tests in `cmd/weave` (git-tracked fixture gets the
  hint; the parley shape: committed stale vocabulary file).
- [x] `atlas/workflow/weave.md`: convergence rule replaces "edited outputs remain".
- [x] Targeted `go test ./cmd/weave/... ./pkg/weaveownership/...` and the propagate-base tests.
- [x] Full `make test` before close (exit 0, 12 shards; a first run tripped only the hermeticity guard because this session committed mid-run).

## Revisions

- 2026-10-12, close review round 1 (BR-1..BR-6). Displaced copies are flat
  (`weaveownership.DisplacedPath`, one escaped entry per output), not a mirror
  tree, so copies of `a` and `a/b` cannot delete or block each other. A copy is
  swept when its output leaves the inventory, so the ARCH-FUNERAL bound is the
  live outputs, not every path ever generated. A failure after a displacement
  carries the copy's path in its error. Data-pass warnings name the owner root.
  Outside a Git repository the tracked-output check is silent.

## Log

### 2026-10-01
- Filed from a parley.nvim slot2 session after weave refused on a reverted vendored vocabulary file. Related, not blocking: #265 (refusal leaves .gitignore half-migrated), #239 (origin of the check), parley.nvim#208 (vendored vocabulary), parley PR #221 (workaround).

### 2026-10-12
- 2026-10-12: closed — make test exit 0 after review round 1 fixes (12 shards, 0 failed). Convergence: TestManagedConvergesChangedRecordedOutput (reverted + edited), TestCompileConvergesRevertedTrackedOutput (parley shape, commit hint only for tracked), TestCompileConvergesEditedGeneratorDestination, TestManagedDataScopeDisplacesChangedMount. Refusals with next action: TestManagedRefusesDirectoryAtRecordedOutput, TestStagedOutputsConvergeEditedAndRefuseUnownedDestinations, TestManagedDataRejectsAuthoredDestinations. Retirement: TestManagedReportsAbandonedOutputOnce, abandoned warning in TestCompileRetiresDisconnectedOutputsPreservesAuthoredFiles. Round-1 fixes: TestManagedFailureAfterDisplacementNamesCopy, TestManagedSweepsDisplacedCopyWithItsOutput, TestDisplacedPathIsFlat, TestPrintManagedReportNamesOwnerAndSkipsNonRepo. Every new guard mutation-checked. atlas/workflow/weave.md updated; preserving authored replacement removed from code. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-12: flow upgraded quick → full — 202 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Implemented in `cmd/weave/internal/plan/ownership.go` (`ManagedReport`, `DisplacedRoot`, `EditHome`) and `cmd/weave/main.go` (`printManagedReport`, used by both the data and artifacts passes).
- Tests converted from the refusal rule: `TestManagedConvergesChangedRecordedOutput` (reverted + edited, latest-copy replacement), `TestStagedOutputsConvergeEditedAndRefuseUnownedDestinations`, `TestStagedFileModesAreOwnedAndPreserved` (mode edit displaced with its mode), `TestCompileConvergesEditedGeneratorDestination`. New: `TestManagedRefusesDirectoryAtRecordedOutput`, `TestManagedReportsAbandonedOutputOnce`, `TestEditHome`, `TestCompileConvergesRevertedTrackedOutput` (the parley shape: committed stale AGENTS.md gets the commit hint, an untracked regenerated CLAUDE.md does not), abandoned warning asserted in `TestCompileRetiresDisconnectedOutputsPreservesAuthoredFiles`.
- Mutation checks, each caught: displacement deletes instead of renaming; tracked hint suppressed; abandoned report dropped; a directory displaced instead of refused.
- Not changed: an unrecorded, occupied, non-staged artifacts destination is still overwritten without a backup (pre-existing behavior; first-inventory migration would otherwise displace every legacy output).
- Close review round 1, FIX-THEN-SHIP, all six fixed in round: BR-1 (failure after displacement unreported → `ApplyManaged` wraps every later error with the copies, `TestManagedFailureAfterDisplacementNamesCopy`), BR-2 (data-pass path ambiguity → owner-rooted names, `TestPrintManagedReportNamesOwnerAndSkipsNonRepo`), non-git warning noise (same test), `a`/`a/b` collision (flat `DisplacedPath`, `TestDisplacedPathIsFlat`), `DisplacedRel` moved beside `RootRel`, displaced copies swept with their output (`TestManagedSweepsDisplacedCopyWithItsOutput`). Also added the data-scope displacement test the review asked for. Mutation checks caught each fix's removal.

