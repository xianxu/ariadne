---
id: 000265
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-10-12
estimate_hours:
card_mirror: 'a5ba30ca515031175b2b9f91ba9232016de60b1d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T18:01:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# weave robustness: atomic compile, symlinked Makefile, legacy include, dependency writes

## Problem

Weave failure modes found during fleet adoption (#241 Log, 2026-09-27) and
#263, still present on main after #263 landed:

1. **A refused compile isn't atomic.** `compilePrepared` (`cmd/weave/main.go`)
   runs the data pass (`ApplyManaged(..., ScopeData)`) for every owner,
   including the root, before the artifact pass. The data pass rewrites
   `.gitignore` even with zero mounts. In a checkout with no ownership record
   yet, that migrates the legacy fixed list out of the file. When the
   artifact pass then refuses ("preserving authored replacement",
   `plan/ownership.go:170`), the old list is gone and ignored generated files
   show up as untracked. (Hand-restored in metis and parli.)
2. **Weave writes into dependency checkouts it only reads.** Data passes over
   non-root owners create `construct/generated/weave/ownership.json` and
   rewrite `.gitignore` there even with nothing to link (#263 fixed the block
   clobber, not the write).
3. **Symlinked root Makefile is followed.** `startup/tools.go:17` uses
   `fs.Stat`, so `Makefile -> ../ariadne/Makefile` passes as a regular file and
   `make tools` builds the wrong repo's tools. `Lstat` would refuse clearly.
4. **Older indirect workflow include isn't recognized.**
   `plan/apply.go:122` matches only `-include Makefile.workflow` /
   `include Makefile.workflow`; the `-include $(WF_WORKFLOW)` form gets a second
   include prepended (parli, tools, xianxu.dev, pair hit it).

Deferred from #241 as non-goals. Also out of scope there: an adopt mode and
per-repo `propagate-base` (you decided adoption was a one-off, now done).

## Spec

Direction: prepare every pass, then write once — one `.gitignore` write after
all passes succeed, and no writes for an owner/pass with nothing to do. Each
item gets a regression test reproducing the observed failure.

## Done when

- A compile that refuses in the artifact pass leaves `.gitignore` and the
  ownership inventory byte-identical (test with a pre-inventory checkout
  carrying the legacy list).
- A compile over a dependency with no data mounts writes nothing in it.
- A symlinked root Makefile is refused with a clear message.
- The indirect include form is recognized (no duplicate prepended).

## Plan

- [x] Item 1: split `ApplyManaged` into a no-write `prepareManaged` (every
  inspection and refusal) and a commit; export `CheckManaged` (prepare only).
  `compilePrepared` collects every owner's data mounts, runs generation into
  its stage, plans artifacts, then checks every pass (each owner's data,
  root artifacts, and a root data-vs-artifacts path collision) before the
  first write; only then applies data, then artifacts. ARCH-DRY: one
  inspection routine for check and apply. Generators read the layer graph,
  not data mounts (both `.dynamic-skill` markers), so generating before the
  data apply is safe.
- [x] Item 1 test: compile-level, pre-inventory leaf carrying the legacy list
  and a data mount, with a generated output path occupied by an unowned file;
  compile refuses and `.gitignore` is byte-identical, no inventory, no mount.
- [x] Item 2 test: compile-level, a dependency with no data mounts gets no
  inventory and no `.gitignore` (already fixed by #264; pin it at compile).
- [x] Item 3: `startup.Tools` uses `Lstat`; a symlinked owner Makefile is
  refused with a message naming the link and saying weave builds each owner's
  tools from its own Makefile. Test.
- [x] Item 4: `hasWorkflowInclude` recognizes `-include $(WF_WORKFLOW)` /
  `include $(WF_WORKFLOW)`. Test that no second include is prepended.

## Log

### 2026-10-12
- 2026-10-12: closed — go test ./cmd/weave/... ./pkg/weaveownership/... green; regression tests for all four items and every projection branch (records, created output at/above/below, retired path), each failing with its branch disabled; full make test: 2 failures, both pre-existing flakes (fleet TestHangingRemote passes 5/5 alone; weave TestCompileRecoversAfterProcessDeath fails ~1/20 with and without this change); review verdict: SHIP
- 2026-10-12: flow upgraded quick → full — 172 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Re-scoped after #264 and #281. #264 (`1bc6c5f4`) moved legacy-list
  retirement to the artifacts pass and made a pass with nothing to apply or
  retire write nothing, which fixes item 2 and the original trigger of item 1.
  #281 removed the recorded-mismatch refusal ("preserving authored
  replacement"). Still live: a root data pass *with* mounts commits its
  inventory and `.gitignore` block before the artifacts pass can refuse
  (unowned occupied generated output, directory at a recorded path), so the
  refusal is not atomic. Items 3 and 4 unchanged. Fleet Makefiles all use the
  direct include form now; item 4 still guards older checkouts.
- Implemented. `TestCompileRefusalLeavesCheckoutUntouched` fails on the old
  compile order (the root data pass wrote its `.gitignore` block before the
  artifacts refusal) and passes with the check-all-then-apply order. A symlinked
  Makefile now refuses in `startup.Tools`, which runs before seed-once, so
  `SeedOnceInstruction`'s non-regular advice no longer reaches a linked
  Makefile in compile; it still covers other non-regular paths.
- Each pass is now inspected twice (check, then apply re-prepares against
  the state the earlier passes left). Compile is a batch step over tens of
  outputs, so the extra hashing is negligible (ARCH-CONSTRAINTS).
- Full `make test`: two failures, both pre-existing flakes.
  `TestHangingRemoteDegradesOnlyItsRepository` (fleet, untouched) passed 5/5
  alone. `TestCompileRecoversAfterProcessDeathDuringGeneration` fails about
  1 in 20 both with and without this branch's compile reorder ("owned stage
  survived retry: .publication-weave-*"). Likely the killed helper's generator
  grandchild is still alive when the retry reclaims, so the stage is kept as
  owned. Reported to the hub, not fixed here.
- Close review round 1: REWORK. BR-1 (Critical): the up-front artifacts check
  read the pre-data inventory, so a path moving from a data mount to a
  generated output refused on every compile ("belongs to data"); the old order
  converged it. BR-2 (Important): the exact-path collision loop missed a mount
  above or below a generated output. One class (family
  cross-pass-check-ignores-prior-pass-effects): a later pass's check must see
  what the earlier pass at the same root does. The enumeration of the data
  pass's effects that an artifacts refusal can read: (1) inventory records,
  (2) outputs it creates, (3) outputs it retires; displacement lands at a
  created output and the ignore block refuses nothing new. `CheckManaged` now
  takes ordered passes and prepares each against the earlier ones (projected
  records, overlap at/above/below created outputs, retired paths treated as
  absent, keeping the retired path's own parent check). That replaces the
  ad hoc loop (ARCH-DRY). Minors: the doc comment now says check skips the
  stage reclaim; the hoist is moot. Both review findings reproduce as test
  failures with the projection disabled.
- Close review round 2: FIX-THEN-SHIP, one blocker. BR-5 (Important): no
  committed test reached the retired-path skip. Coverage of the enumeration
  above: records (BR-1 test), created output at the path (collision test),
  above it (BR-2 test), below it (plan-level `TestCheckManagedSeesEarlierPassOutputs`),
  retired path or ancestor (new `TestCompileMovesDataMountToGeneratedParent`,
  which fails with the skip disabled: "non-directory parent"). The Minor about
  the retirement rule being restated is fixed: `managedPass.retiring` is shared
  by commit and the check.

### 2026-09-28
