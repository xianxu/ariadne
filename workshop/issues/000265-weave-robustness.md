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

- [ ] Item 1: split `ApplyManaged` into a no-write `prepareManaged` (every
  inspection and refusal) and a commit; export `CheckManaged` (prepare only).
  `compilePrepared` collects every owner's data mounts, runs generation into
  its stage, plans artifacts, then checks every pass (each owner's data,
  root artifacts, and a root data-vs-artifacts path collision) before the
  first write; only then applies data, then artifacts. ARCH-DRY: one
  inspection routine for check and apply. Generators read the layer graph,
  not data mounts (both `.dynamic-skill` markers), so generating before the
  data apply is safe.
- [ ] Item 1 test: compile-level, pre-inventory leaf carrying the legacy list
  and a data mount, with a generated output path occupied by an unowned file;
  compile refuses and `.gitignore` is byte-identical, no inventory, no mount.
- [ ] Item 2 test: compile-level, a dependency with no data mounts gets no
  inventory and no `.gitignore` (already fixed by #264; pin it at compile).
- [ ] Item 3: `startup.Tools` uses `Lstat`; a symlinked owner Makefile is
  refused with a message naming the link and saying weave builds each owner's
  tools from its own Makefile. Test.
- [ ] Item 4: `hasWorkflowInclude` recognizes `-include $(WF_WORKFLOW)` /
  `include $(WF_WORKFLOW)`. Test that no second include is prepended.

## Log

### 2026-10-12
- Re-scoped after #264 and #281. #264 (`1bc6c5f4`) moved legacy-list
  retirement to the artifacts pass and made a pass with nothing to apply or
  retire write nothing, which fixes item 2 and the original trigger of item 1.
  #281 removed the recorded-mismatch refusal ("preserving authored
  replacement"). Still live: a root data pass *with* mounts commits its
  inventory and `.gitignore` block before the artifacts pass can refuse
  (unowned occupied generated output, directory at a recorded path), so the
  refusal is not atomic. Items 3 and 4 unchanged. Fleet Makefiles all use the
  direct include form now; item 4 still guards older checkouts.

### 2026-09-28
