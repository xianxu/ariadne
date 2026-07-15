# Boundary Review — ariadne#156 (whole-issue close)

| field | value |
|-------|-------|
| issue | 156 — change-code idempotent branch for milestone re-runs |
| repo | ariadne |
| issue file | workshop/issues/000156-change-code-idempotent-branch-for-milestone-re-runs.md |
| boundary | whole-issue close |
| milestone | — |
| window | df8f4493a40e8e056e60bc8260ee0fe9404f82c6..HEAD |
| command | sdlc close --issue 156 |
| reviewer | claude |
| timestamp | 2026-07-06T15:27:49-07:00 |
| verdict | SHIP |

## Review

All green — build, vet, gofmt, and the full `cmd/sdlc` package test suite (18.3s) pass. I've verified the code, the tests, the caller wiring, the refactor's behavior-preservation, and the atlas update. Here's my review.

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This is a clean, well-scoped bugfix that makes `change-code`'s branch creation idempotent for milestone re-runs, delivering *both* axes the issue committed to (in-place and worktree), not just the easy in-place subset that triggered the original report. The design faithfully mirrors the existing `estimateRefusal` pure/IO seam: pure deciders (`decideInPlaceBranch`, `decideWorktreeBranch`) + a pure porcelain parser, with git IO isolated to thin probe helpers. It also does the right ARCH-DRY consolidation the plan-quality review demanded — extracting `parseWorktrees` as the single porcelain grammar and refolding `listWorktrees`, `findMainWorktree`, and the new `worktreeForBranch` onto it (removing a second hand-walked parser and the now-dead `bufio`/`bytes` imports). Tests are strong and layered: pure table tests, IO-wiring tests asserting the *exact* git command per state (the original bug was a shell-wiring bug), and a real-git end-to-end reproduction of the precise `#156` failure. Nothing blocks SHIP.

**1. Strengths**
- `branchcreate.go:218-295` — the pure-decider seam is exactly the right factoring; deciders are total (explicit return in every branch, no reliance on iota zero-value), and the `wtFound`-wins-over-`branchExists` ordering is correct because git forbids the same branch in two worktrees, with the rationale in the comment (`branchcreate.go:250-252`).
- ARCH-DRY consolidation landed as planned: `state.go:147` `parseWorktrees` is now the single source; `claim.go:386-391` refolds `findMainWorktree` onto it while keeping IO (`r.Git`) at the boundary. The refactor is behavior-preserving — old `line == "branch refs/heads/main"` vs new prefix-strip-then-`== "main"` are equivalent, and existing `TestFindMainWorktree_*` / `TestMergeUsesFindMainWorktree_ViaStub` pass unchanged.
- `branchcreate_test.go:230-286` `TestCreateInPlaceBranch_RealRepo_IdempotentRerun` drives real `rev-parse`/`show-ref` output through a hermetic repo and reproduces the exact production error path — the Log even records that forcing the old unconditional `checkout -b` makes it fail with the byte-identical `exit status 128 / already exists`. That's the anti-mock discipline this checklist wants.
- Idempotent paths still emit the machine-readable location on stdout (`branchcreate.go:214`, `:187`) in *every* case including the no-op, so any downstream stdout consumer is unaffected.

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- `branchcreate.go:156` `porcelain, _ := r.Git("worktree", "list", "--porcelain")` swallows the probe error. This degrades safely (empty porcelain → `worktreeAddExisting`/`worktreeAddNew`, and a genuinely-conflicting `worktree add` still errors from git), and the pre-change code didn't call this at all, so no source contract is violated — noting only for completeness.
- `state.go:168` — `parseWorktrees` handles the `bare` label but `TestParseWorktrees` exercises only `detached`, not `bare`. Trivial gap; the `bare` path is unreachable for the `main`/target filters anyway.

**5. Test coverage notes**
- The worktree axis (`reuse` / `addExisting` / `addNew`) is covered only via `captureRunner` with synthetic porcelain strings; the real-git end-to-end test covers the in-place axis only. The parser itself is validated against realistic porcelain (`TestParseWorktrees` includes a real second-worktree block), and `listWorktrees`/`findMainWorktree` already depend on the same live format in production, so risk is low — but a real-git worktree-reuse test (create a second worktree, re-run, assert no `worktree add` + `.goto` rewrite) would be the belt-and-suspenders for the operator's own "half-idempotent worktree path is just the next bug" concern. Future hardening, non-blocking.

**6. Architectural notes for upcoming work**
- ARCH-DRY: pass — one porcelain grammar, three consumers; no third parser added.
- ARCH-PURE: pass — deciders + parser are pure and table-tested with no IO; git touches live only in `currentBranch`/`branchExists`/the two shells, injected via `gitRunner`.
- ARCH-PURPOSE: pass — shadow-sweep of the issue's purpose confirms both consumers (`createInPlaceBranch`, `createWorktreeBranch`) are made idempotent and each of the six declared state→command mappings in the Spec is delivered and pinned by a test; nothing deferred as "follow-up" that was actually the point.
- Note for `#119` (multi-agent benchmark harness, which plans to extend `createWorktreeBranch` with a `base` ref param): the new `decideWorktreeBranch`/`worktreeForBranch` seam is a stable surface to build the `bench/<task>/<agent>/<runid>` variant on — the `base` param slots into the `worktreeAddNew` arm (`worktree add -b name path <base>`) without disturbing the reuse/addExisting decisions.

**7. Plan revision recommendations** — none. The plan's Core-concepts-equivalent (the three checklist items + the restated "Done when" per-axis) matches the code exactly; all three items are genuinely delivered at the stated `branchcreate.go` locations. No `## Revisions` entry needed.
