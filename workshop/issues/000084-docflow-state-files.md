---
id: 000084
status: working
deps: []
github_issue:
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.75
---

# docflow: store review state in .git/docflow/ files, not .git/config (sandbox-compatible)

## Problem

docflow stashes per-branch state in `.git/config` via `git config`
(`branch.<rb>.docflowBase`, `branch.<rb>.docflowFile`). The Claude Code Bash
sandbox **denies writes to `.git/config`** (and `.git/hooks/`) by design — both
can execute code, so they're a deliberate security boundary (confirmed against
the sandbox docs + empirically: a plain file under `.git/` writes fine, `git
config` is `Operation not permitted`). So every `docflow start`/`finish` needs a
per-command `dangerouslyDisableSandbox` override when run in the session's
working repo (surfaced dogfooding docflow in xianxu.dev — its `.git/config` is
the guarded one; ariadne's worked only because it's a *sibling*, not the cwd).

The collision is tiny and avoidable: docflow's *only* `.git/config` writes are
the two state keys. Everything else (`checkout`/`add`/`commit`/`merge`/`branch
-d`) writes refs/objects, which the sandbox allows. `round` is already fully
sandbox-clean.

## Spec

Move docflow's per-review state out of `.git/config` into plain files under
`.git/docflow/<slug>/` (writable inside the sandbox; git ignores the dir). This
is the **root-cause** fix — it keeps the `.git/config` security deny intact
rather than punching an `allowWrite` hole through it (`ARCH-Simplicity`:
sidestep the boundary, don't weaken it).

Storage layout (mirrors the two retired config keys 1:1):
- `.git/docflow/<slug>/base`  — single line, the base branch (was `docflowBase`).
- `.git/docflow/<slug>/files` — one in-scope path per line (was `docflowFile`).

`<slug>` = branch name minus `review/` (alnum+dash, already path-safe). The git
dir is resolved with `git rev-parse --git-dir` (worktree-correct). A
`docflow_meta_dir <review-branch>` helper returns the dir; `review_base` and
`inscope_files` read from it (`cat ... || default`); `start` writes `base` +
appends deduped `files`; `finish` cleans up with `rm -rf` of the meta dir.

No migration: there are no live review branches using the old config keys (the
#79/#81 reviews are merged + finished). Note it in the Log; don't add migration code.

## Done when

- docflow stores `base`/`files` under `.git/docflow/<slug>/`; **no** `git config`
  write remains in the script (grep clean).
- `review_base` / `inscope_files` read the files; `start` writes them (deduped);
  `finish` `rm -rf`s the meta dir.
- Test asserts: state lives in the files (not `git config` — assert
  `git config --get branch.<rb>.docflowBase` is empty), dedup still holds, files
  scope still works, meta dir removed after finish. All prior assertions pass.
- End-to-end sandbox proof: after propagating to xianxu.dev, `docflow start/round/
  finish` on a throwaway doc runs **without** any `dangerouslyDisableSandbox`.
- atlas `atlas/workflow/docflow.md` updated (state now in `.git/docflow/`, and the
  *why* — `.git/config` is sandbox-denied).

## Plan

Single-pass refinement (one review boundary) → plain checkboxes.

- [x] `docflow_meta_dir <rb>` helper: `$(git rev-parse --git-dir)/docflow/<slug>`.
- [x] `review_base` / `inscope_files`: read `base` / `files` from the meta dir.
- [x] `cmd_start`: `mkdir -p` meta dir, write `base`, append deduped `files`
  (replaces the two `git config` writes, both branch paths).
- [x] `cmd_finish`: `rm -rf` the meta dir (replaces the two `git config --unset`).
- [x] `docflow.test.sh`: assert file-based storage + no `.git/config` branch keys
  + meta-dir cleanup; keep dedup / scope / space-path assertions green.
- [x] atlas note: state in `.git/docflow/`, why (sandbox denies `.git/config`).
- [x] Verify end-to-end sandboxed in xianxu.dev (no override needed).

## Log

### 2026-06-04

Filed from the sandbox investigation in the xianxu.dev session. Root cause: docflow's
only sandbox collision is its `git config` state writes to the security-denied
`.git/config`. Fix = plain state files under `.git/docflow/`. Sibling of #81 (both
refine #79's docflow); related to #82 (cross-repo base-layer workflow ergonomics).

Implemented. New `docflow_meta_dir()` derives `$(git rev-parse --git-dir)/docflow/<slug>`
(ARCH-DRY: one place, reused by review_base/inscope_files/start/finish). `review_base`/
`inscope_files` now `cat` `base`/`files`; `start` does `mkdir -p` + write base + append
deduped files; `finish` `rm -rf`s the meta dir. **Zero `git config` left in the script**
(grep-clean). Folded in both plan-quality notes: (1) rewrote the 3 stale config-reading
test assertions to read the meta files (no contradictory pair) + added "no .git/config
leak" + "meta dir removed after finish"; (2) **--git-dir is worktree-local** (vs the old
repo-shared config) — a deliberate semantic shift, fine for docflow's start→round→finish-
in-one-place model (finish already guards against the base being checked out elsewhere).

Also (sandbox UX): `git branch -d` tries to prune the branch's `.git/config` section and
the sandbox denies that write — harmless (exit 0, branch deleted), but it printed a scary
`error: could not write config file`. Suppressed that stderr in finish.

Verified end-to-end **sandboxed in xianxu.dev** (the repo whose `.git/config` is guarded):
`docflow start/round/finish` on a throwaway doc (merged to a throwaway base so main stayed
clean) all ran `rc=0` with **no `dangerouslyDisableSandbox`** and clean output. Test 35/35
unsandboxed. This is the Done-when sandbox proof.
