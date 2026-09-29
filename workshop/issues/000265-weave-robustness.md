---
id: 000265
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: 'c9b50129788b025d0b8a200f59bf2edbd8fbd915' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-09-28
