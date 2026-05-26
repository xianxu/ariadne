# Setup & Replication

`construct/setup.sh` is the unified base-layer-replication mechanism. One
canonical script lives in ariadne and gets vendored down to every derivative
via `construct/base.manifest`. Same source-of-truth file at every layer.

Design rationale: `workshop/issues/000032`.

## What it does

For a target repo, walks each transitive upstream layer's
`construct/base.manifest` in topological order (ancestors first), applying
the manifest's `symlink` / `copy` / `merge` / `scaffold` / `touch` actions.
Then runs post-processing: creates `Makefile` + `Makefile.local` if absent,
applies `.gitignore` entries, syncs local-skill symlinks, records mode.

The "walk N manifests" generalizes today's depth-specific scripts:
- ariadne at depth 0: no ancestors, self-refresh just runs skill sync.
- nous at depth 1 (post-migration): walks ariadne's manifest, applies into nous.
- baby brain at depth 2: walks ariadne's, then nous's, then its own
  contributions (if any).

## Ancestor discovery

Two modes:

1. **Go-managed** — target has `go.mod`. `go list -m -f '{{.Dir}}' all`
   returns the dep graph in resolution order. Script filters to modules that
   ship `construct/base.manifest` and walks them in order. Pinning a version
   (or replace-directive-trunk-follow) for each upstream lives entirely in
   the target's `go.mod`.

2. **Fallback (no go.mod, or no Go installed)** — single ancestor = the
   script's own resolved upstream (`dirname $(readlink -f $0)/..`). This is
   today's `../ariadne/construct/setup.sh` invocation pattern, preserved
   for backward-compat with pre-Go-modules consumers.

## Three operating modes (orthogonal to depth)

Recorded in `.ariadne-mode` (legacy filename, kept for backward compat):

| Mode | Manifest entries become | Use case |
|---|---|---|
| `symlink` (default) | symlinks into the upstream | Sibling-checkout development; trunk-follow |
| `vendor` | copies in target tree | Pinned snapshot; offline / hermetic builds |

Switch with `--symlink` / `--vendor` flags. Mode change requires confirmation
(`--yes` to skip).

Note: the `symlink` mode here is orthogonal to Go's `replace` directive.
Go's replace controls where Go *imports* resolve to; `symlink` here controls
how non-code text files are vendored from the upstream's resolved location.
Both can use the same upstream path.

## Adopting a fresh derivative

```bash
cd /path/to/new-derivative
git init
../ariadne/construct/setup.sh --yes      # depth-1 adoption
```

After this:
- `AGENTS.md`, `Makefile.workflow`, scripts, etc. are linked into the derivative.
- `Makefile`, `Makefile.local`, `AGENTS.local.md` are created if absent.
- `.ariadne-mode` records `symlink` (or `vendor` if `--vendor` was passed).
- `construct/setup.sh` itself is linked, so the derivative can self-refresh.

For Go-managed adoption (recommended once the consumer is Go-aware):

```bash
go mod init github.com/<owner>/<derivative>
go get github.com/xianxu/ariadne@latest        # or: edit go.mod manually with replace ⇒ ../ariadne
make refresh
```

Subsequent updates: `make refresh` re-runs setup.sh against the upstream
location Go resolves. Bumping the pin = editing the `require` line + re-running.

## Per-binary build opt-out

`Makefile.workflow build:` scans `cmd/*/main.go` and builds each. Some binaries
shouldn't be auto-built (e.g., the nous binary is distributed signed +
notarized — overwriting `bin/nous` with an unsigned local build invalidates
macOS keychain ACL grants and notification capabilities).

Opt out per-binary by dropping a sentinel:

    cmd/<name>/.skip-make-build

Contents are free-form prose explaining the rationale (future operators read
it). Each opted-out binary owns its sentinel; the base layer doesn't carry
derivative-specific name lists.

## Generated artifacts (not vendored)

Artifacts that are *functions of code shipping through Go modules* (e.g.,
`construct/local/sdlc/SKILL.md` from `sdlc --index`) are NOT shipped via
`base.manifest`. They regenerate at the consumer via the binary's
install/refresh path. The version of the binary determines the version of
every derived artifact automatically — no text-vs-code lockstep drift
possible.

## Related

- `construct/base.manifest` — what ariadne contributes (action + path
  pairs).
- `construct/setup.sh` — this script.
- `workshop/issues/000032` — design rationale, three operating modes,
  migration plan.
- `atlas/workflow/base-layer.md` — adopting ariadne's base layer (this
  document supersedes parts of it; cross-reference as the system stabilizes).
