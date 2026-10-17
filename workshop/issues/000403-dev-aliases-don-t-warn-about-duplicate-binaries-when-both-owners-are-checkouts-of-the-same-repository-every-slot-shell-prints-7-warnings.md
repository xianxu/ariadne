---
id: 000403
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
card_mirror: '6992460b802bdfde97043cc911efdf892ca58c40' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T15:20:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "97be0099", done: "11138338"}
---

# dev-aliases: don't warn about duplicate binaries when both owners are checkouts of the same repository (every slot shell prints 7 warnings)

## Problem

Every shell that starts inside a slot worktree prints seven warnings from `construct/dev-aliases.sh`, e.g.
`dev-aliases: duplicate binary 'sdlc' owned by: /Users/xianxu/workspace/ariadne
/Users/xianxu/workspace/worktree/parley.nvim-slot4/ariadne (last in sorted order wins)`. Reproduced from
parley.nvim-slot3 and -slot4 (2026-10-15); the operator saw it when parley.nvim:4's pane started and asked whether a
`.zshrc` change caused it.

The two owners are the **same repository**: the slot's own ariadne clone (its substrate) and the workspace sibling
`~/workspace/ariadne`. That's the designed slot layout, not a conflict, and last-in-sorted-order picks the slot's clone,
which is the intended owner. The warning is noise that trains people to ignore real duplicates (two *different*
repos owning one binary name).

## Spec

- Treat checkouts of the same repository (same origin remote URL, normalized) as one owner: no warning; pick the
  current repo's own substrate clone deterministically (as now), and say so only under a verbose flag.
- Keep the warning, and `--strict`'s non-zero exit, for genuinely different repositories owning one binary name.

## Done when

- A fixture test: a slot layout (workspace sibling plus the slot's substrate clone of the same origin) prints nothing
  and maps binaries to the slot's clone; two different repos owning one name still warn, and fail under `--strict`.

## Plan

- [x] Add hermetic Git fixtures for two checkouts of one origin in the slot layout, equivalent HTTPS/SSH origins, and two distinct origins. Assert the selected slot owner, stderr, and `--strict` status.
- [x] Normalize each candidate owner's origin once at the IO boundary; group candidates by binary and normalized origin for warnings only. Keep the existing deterministic last-path-wins mapping. Missing or unrecognized origins stay distinct by path so a real collision cannot be hidden. `--verbose` explains a suppressed same-origin duplicate.
- [ ] Run the focused shell test, `make -k build`, and the recorded full suite; update the script's contract documentation and close on verified evidence.

## Log

### 2026-10-15

### 2026-10-17

Claimed and started from current main. The current duplicate check counts paths, while final selection is last sorted path. The shell's sibling plus substrate scan deliberately finds both checkouts. ARCH-PURE: normalize Git origin URLs at collection time, then make the warning decision from data; leave command emission and owner selection alone. This stays inside the quick-flow shell.

The new fixture failed before implementation on the same-origin warning and strict status, then passed after the identity change (69/69). The test also exposed `/var` and `/private/var` spelling of one macOS temp directory; canonicalizing `--workspace` preserves deterministic owner selection. Read-only check from parley.nvim:4 selected its ariadne substrate and emitted no warning in `--strict --list`. `bash -n`, `git diff --check`, and `make -k build` passed.

Before the queued full suite began, I compared the shell normalization with `gitx.PublicationRepository`: GitHub owner/repo is case-insensitive, while other hosts cannot safely be folded across transports. I interrupted the queued attempt, added the mixed-case GitHub regression, and changed the identity rule accordingly. I also added a missing-origin strict collision assertion; no unmeasured success is inferred from a missing remote.

## Revisions

### 2026-10-17 — Preserve non-GitHub identity conservatively

The original plan said to normalize origin URLs generically. GitHub's known identity semantics support SSH/HTTPS and case folding; arbitrary Git hosts may distinguish those. Normalize only GitHub's recognized forms, compare other hosts by exact URL, and keep unknown origins distinct by path. The selected binary remains last sorted checkout.
