---
id: 000403
status: codecomplete
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
card_mirror: 'f662770c11e8a42c178cd013ef1ce7b9d765ba10' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T15:20:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "97be0099", done: "11138338"}
actual_hours: 0.52
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
- Equivalent GitHub HTTPS, scp-style SSH and `ssh://` origins, including path case, suppress warnings and pass strict mode. Identical non-GitHub URLs also suppress warnings; differing transport or path case, absent origins, and unrecognized relative origins warn and fail strict mode. `--verbose` explains a suppressed collision without changing `--list` output.
- README and atlas describe the diagnostic flag and strict collision rule.
- Diagnostic fixtures compare complete `--list` stdout and exit status with and without `--verbose`, in ordinary and strict modes, while asserting stderr separately.

## Plan

- [x] Add hermetic Git fixtures for two checkouts of one origin in the slot layout, equivalent HTTPS/SSH origins, and two distinct origins. Assert the selected slot owner, stderr, and `--strict` status.
- [x] Normalize each candidate owner's origin once at the IO boundary; group candidates by binary and normalized origin for warnings only. Keep the existing deterministic last-path-wins mapping. Missing or unrecognized origins stay distinct by path so a real collision cannot be hidden. `--verbose` explains a suppressed same-origin duplicate.
- [x] Run the focused shell test, `make -k build`, and the recorded full suite; update the script's contract documentation and close on verified evidence.

## Log

### 2026-10-15

### 2026-10-17
- 2026-10-17: closed — 82/82 focused dev-aliases regressions pass, including ordinary/strict verbose stdout and status equivalence; live parley.nvim:4 strict read selects all seven substrate tools without warnings; make -k build and recorded suite on c4815824 pass (12 shards, 67 packages); PR #289 CI green; review verdict: SHIP

Claimed and started from current main. The current duplicate check counts paths, while final selection is last sorted path. The shell's sibling plus substrate scan deliberately finds both checkouts. ARCH-PURE: normalize Git origin URLs at collection time, then make the warning decision from data; leave command emission and owner selection alone. This stays inside the quick-flow shell.

The new fixture failed before implementation on the same-origin warning and strict status, then passed after the identity change (69/69). The test also exposed `/var` and `/private/var` spelling of one macOS temp directory; canonicalizing `--workspace` preserves deterministic owner selection. Read-only check from parley.nvim:4 selected its ariadne substrate and emitted no warning in `--strict --list`. `bash -n`, `git diff --check`, and `make -k build` passed.

Before the queued full suite began, I compared the shell normalization with `gitx.PublicationRepository`: GitHub owner/repo is case-insensitive, while other hosts cannot safely be folded across transports. I interrupted the queued attempt, added the mixed-case GitHub regression, and changed the identity rule accordingly. I also added a missing-origin strict collision assertion; no unmeasured success is inferred from a missing remote.

Final focused test: 70 passed, 0 failed. Recorded `sdlc-local/suite` on c64466ea43415a2281cd773e9aa50be67d1c6451: 12 shards / 1577 sdlc tests and 67 other packages, exit 0, wall 496s, log sha256:5aea23be8769. PR #289's merge-check and local-gates-guard passed. The live parley.nvim:4 read-only `--strict --list` check selected all seven tools from its slot substrate with no warnings.

Close review round 1 returned FIX-THEN-SHIP with two blocking Important findings: README omitted the new user-facing flag, and acceptance plus fixture coverage omitted parts of the revised origin policy. A Minor ARCH-DRY finding identified duplicate warning/strict classification. I enumerated the policy matrix in Done when and fixtures, documented the flag in README, and made both warning and strict status consume one classification. This requires another exact-head suite and close review.

Close review round 2 disposed of the three earlier findings but found that the verbose fixture discarded stdout, leaving the machine-readable `--list` promise untested. I added ordinary and strict stdout/status equivalence assertions for both same-origin and distinct-origin cases, alongside separate stderr assertions. The finding is in the acceptance-contract-coverage family; the added checks cover that whole diagnostic mode rather than only the named call site.

## Revisions

### 2026-10-17 — Preserve non-GitHub identity conservatively

The original plan said to normalize origin URLs generically. GitHub's known identity semantics support SSH/HTTPS and case folding; arbitrary Git hosts may distinguish those. Normalize only GitHub's recognized forms, compare other hosts by exact URL, and keep unknown origins distinct by path. The selected binary remains last sorted checkout.

### 2026-10-17 — Cover the complete origin policy

Review round 1 showed that the original Done when and test matrix described only the motivating collision. The accepted policy also covers GitHub `ssh://`, identical non-GitHub URLs, non-GitHub transport and path-case differences, and unrecognized origins. These are now explicit acceptance cases; README documentation and one shared collision classification are in scope.

### 2026-10-17 — Assert diagnostic stream separation

Review round 2 showed that checking only the presence of a verbose stderr message does not prove `--list` stdout remains machine-readable. The diagnostic-mode acceptance now requires full stdout and exit-status equivalence with and without `--verbose` for ordinary and strict invocations, with stderr checked independently.
