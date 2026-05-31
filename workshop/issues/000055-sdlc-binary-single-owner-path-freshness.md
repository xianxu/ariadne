---
id: 000055
status: open
deps: []
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 3
---

# sdlc binary: single-owner build + PATH aggregation + freshness

## Problem

`sdlc` is the **only** binary in the ariadne ecosystem that gets duplicated
across repos. Every other tool is single-owner (nous owns `nous`/`gmail`/
`oneshot`; pair owns `pair`/`pair-wrap`/…; charon owns `charon`; gstack owns its
60+ scripts — one copy each, in the owning repo's `bin/`). But `sdlc` is built
*per derivative* via `construct/go.mod` (`replace => ../../ariadne`), so the same
ariadne source compiles into N independent `bin/sdlc` binaries, each with its own
staleness clock.

Snapshot taken 2026-05-31 — three different versions live on one machine:

```
ariadne      May 31 12:42   4842882   ← current (post-#51)
you-decide   May 31 14:28   4842882   ← current
nous         May 31 10:16   4826338   ← stale
pair         May 28 10:34   4826338   ← stale
parley.nvim  May 27 10:27   4773634   ← stale-r
```

This is the vestige of the abandoned "derivative is self-sustained / builds the
substrate itself" framing. We moved to peer-symlink + `./bootstrap.sh` (clone
base-layer deps as siblings); under that framing each base layer should **own**
the build of its own binaries and derivatives should not rebuild them. `sdlc`
just never got migrated to match.

Surfaced concretely by the #51 dogfood (#53 Phase B): you-decide's `bin/sdlc`
was a month stale (pre-#51) and died on the in-place merge with `find main
worktree: could not find a worktree on branch 'main'`. `pair` and `parley.nvim`
are pre-#51 *right now* and would hit the same failure.

The PATH-aggregation direction is already declared — `Makefile.workflow`'s
`sdlc-install` comment states the `~/bin` symlink approach was retired "so all
repo `bin/` dirs compose uniformly on PATH." This issue completes that decision
for `sdlc` and adds the freshness guarantee single-ownership still needs.

## Spec

**Model (locked):** each ariadne-styled repo's `bin/` holds only the binaries it
*owns* (builds from its own `cmd/`). ariadne owns `sdlc`. Derivatives do not
build or carry a copy. `sdlc` is resolved from PATH (ariadne's `bin/`), made
available everywhere by a single aggregator script. No machine-level
`make install` / GOBIN — each repo's `bin/` is the de-facto serving location for
the binaries it owns.

**Decisions (locked with operator 2026-05-31):**

1. **No symlinks.** Rely on PATH only. The `Makefile.workflow` wrappers that gate
   on `[ -x bin/sdlc ]` switch to `command -v sdlc` (PATH lookup). Once
   `ariadne/bin` is on PATH, `sdlc` resolves from any cwd including inside a
   derivative.
2. **Freshness = owner-side rebuild hook + embedded-SHA staleness warning.**
   - `post-merge` + `post-checkout` (branch-switch only) git hooks **in the
     owner repo (ariadne)**, guarded to rebuild only when Go sources changed in
     the diff (doc-only pulls = no-op). Not `post-commit`. These cover the
     *consumer* case (pulled new source, forgot to rebuild) — the exact case
     that bit us. Uncommitted-source staleness during active sdlc dev stays a
     dev concern (you rebuild to test).
   - Build embeds the git SHA via `-ldflags "-X …gitSHA=…"`; `sdlc --version`
     prints it; startup warns when built SHA ≠ repo HEAD. This makes staleness
     *visible* — the root reason the dogfood lost time was that staleness was
     only inferrable from file mtime.
3. **Scope: just `sdlc`.** The scan confirms it's the only duplicated binary;
   no broader ecosystem audit in this issue.

**Non-goals:** self-rebuilding wrapper (`go build` on every invocation) —
rejected for per-call latency in judge loops; hook + SHA-warn instead. No change
to how nous/pair/etc. own their binaries (they already conform).

## Plan

### M1 — single-owner + PATH aggregation
- [ ] Delete duplicate `bin/sdlc` from derivatives (nous, pair, parley.nvim,
  you-decide) and remove `sdlc-build` from their `tools` composition so it isn't
  rebuilt downstream. (`bin/sdlc` is gitignored in derivatives, so this is local
  cleanup + Makefile/`tools`-target edits, not tracked-file deletes — confirm per
  repo.)
- [ ] `Makefile.workflow`: change the ~5 wrapper guards from `[ -x bin/sdlc ]` to
  `command -v sdlc >/dev/null` (close, judge, change-code, fetch, push, pr,
  merge, claim — audit all `bin/sdlc` references). This is a base-layer file →
  propagates to derivatives via `make refresh`.
- [ ] Create `ariadne/scripts/bin-path.sh`: scans the parent dir of ariadne for
  ariadne-styled repos (filter: has `AGENTS.md` AND non-empty `bin/` — naturally
  excludes empty-bin brains and `*.legacy`/test dirs), dedups by basename, prints
  `export PATH="<repo1>/bin:<repo2>/bin:…:$PATH"`. Lives in ariadne only (machine
  aggregator, not portable).
- [ ] Wire `bin-path.sh` into the bootstrap/install story: replace the per-repo
  "append my bin to rc" behavior of `sdlc-install.sh` with one sourced line
  emitting the deduped union. (Today N repos each append their own bin → N lines,
  order decides which `sdlc` wins. After: one source line, one owner.)
- [ ] Update `atlas/workflow/sdlc-binary.md`: replace the "downstream staleness
  gotcha" note (added in #54) with the single-owner model — sdlc is owned by
  ariadne, served from its bin/, PATH-aggregated; derivatives don't build it.
- [ ] Update `construct/base.manifest` if the Makefile.workflow change or
  bin-path.sh affects the portable path set.

### M2 — freshness safety net
- [ ] `sdlc-build` embeds git SHA via `-ldflags "-X main.gitSHA=$(git rev-parse --short HEAD)"`.
- [ ] `sdlc --version` (or `sdlc version`) prints the SHA + build info; startup
  emits a one-line warn to stderr when built SHA ≠ `git rev-parse HEAD` of the
  ariadne checkout (cheap; skip if not in a git repo / SHA unset).
- [ ] Owner-side git hooks: `post-merge` + `post-checkout` (guard: branch-switch
  flag, and only rebuild when `*.go`/`go.mod`/`go.sum` changed in the range) that
  run `make sdlc-build`. Install via `bootstrap.sh` (idempotent; ariadne only —
  derivatives have no sdlc source to rebuild).
- [ ] Update `atlas/workflow/sdlc-binary.md` with the freshness mechanism.

## Done when

- [ ] Exactly one `sdlc` binary exists per machine (ariadne's); `which -a sdlc`
  shows a single entry after sourcing `bin-path.sh`.
- [ ] `sdlc <verb>` resolves and runs from inside a derivative repo (e.g.
  you-decide) with no local `bin/sdlc`; the Makefile wrappers still work via
  `command -v sdlc`.
- [ ] `bin-path.sh` emits a deduped PATH line covering all ariadne-styled repos
  with non-empty bins, excluding legacy/test/empty-bin dirs.
- [ ] `sdlc --version` prints the git SHA; running a stale binary (HEAD moved
  past the built SHA) prints a visible staleness warning.
- [ ] After `git pull` that advances ariadne Go source, `bin/sdlc` is rebuilt
  automatically (post-merge hook); a doc-only pull does not rebuild.
- [ ] `go test ./cmd/sdlc/...` green; `make refresh` in a derivative picks up the
  `command -v sdlc` guard change.

## Log

- 2026-05-31 — Filed as the follow-on to #51/#53; the #51 in-place-merge dogfood
  (#54) surfaced the stale-duplicate-binary failure that motivates this. Decisions
  1–3 locked with operator in the same session.
