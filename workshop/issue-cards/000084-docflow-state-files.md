---
id: '000084'
status: done
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.75
actual_hours: 0.75
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
