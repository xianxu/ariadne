---
id: 000049
status: done
deps: []
created: 2026-05-29
updated: 2026-06-02
estimate_hours:
actual_hours: 2.5
---

# Data dependencies: a looser git-submodule for content peers

## Problem

An ariadne repo sometimes wants to **consume the content of another repo**
without that repo being part of its substrate layer graph. The motivating
case: `brain` consumes `you-decide` (candidate reads, controversies, the
voter-advisor skills) but does **not** derive its base layer from it —
`you-decide` is a *sibling* derivative of ariadne, not an ancestor of `brain`.

Today there is no first-class way to declare this. The two things you can reach
for are both wrong:

- **The substrate peer mechanism** (root/`construct` `go.mod` `replace` →
  `bootstrap-peers.sh` clone + `setup.sh discover_ancestors` apply). Declaring a
  content repo here makes the walker apply its `construct/base.manifest`,
  symlinking the peer's base-layer files into the consumer. Wrong — it couples
  *clone* with *substrate-apply* (see "the enabling seam" below).
- **A hand-made symlink + a stray go.mod in the data tree** (what the
  brain↔you-decide experiment did first). The data-tree `go.mod` is read by
  nothing — no walker recurses into `data/` — so it does not clone on
  bootstrap. It's inert decoration; only the manual symlink actually mounts.

So the operator is left hand-cloning the sibling and hand-making the symlink,
with nothing to drive it on a fresh brain bootstrap.

## Spec

*(design ticket — capture the concept + open questions, not a frozen design)*

Introduce **data dependency** as a named concept: a content repo a consumer
wants **cloned beside it and mounted via symlink**, with *no* substrate
inheritance.

It is essentially **a looser git submodule**:

| | git submodule | data dependency |
|---|---|---|
| location | nested inside consumer | sibling, beside consumer |
| pin | exact commit (gitlink) | floating — tracks remote HEAD, `git pull` to update |
| history | embedded in superproject | fully independent repo |
| mount | the nested dir itself | a committed symlink into the consumer's tree |
| substrate | n/a | explicitly **not** applied (this is the whole point) |
| encryption | inherits superproject | independent; consumer may be gcrypt, dep plaintext |

Floating-HEAD (no pin) is intentional: these are *living data* repos, not
versioned code deps — pinning would just be perpetual bump-churn. The
consumer's privacy boundary stays at the directory level (consumer private,
dep public), and a committed dangling symlink is already ariadne house style
(every Makefile/AGENTS symlink dangles on a bare clone — that's why
`bootstrap.sh` exists).

### Chosen design (implemented 2026-05-30)

A **separate manifest** — `construct/data-deps` — rather than overloading the
substrate go.mod. This was the key decision: it sidesteps the substrate walker
entirely (no clone-vs-apply coupling to untangle) and is **language-agnostic**,
which matters because brain consumes "many repos in various shapes, not always a
language dependency" (operator). Minimum mechanism: just clone + symlink.

- **Declaration.** `construct/data-deps`, two whitespace-separated columns per
  line, `#` comments / blank lines ignored:
  `<git-url>  <symlink-path-relative-to-repo-root>`.
- **Mount.** `construct/scripts/clone-data-deps.sh` (run via `make data-deps`,
  an additive prereq of `bootstrap`): clones each repo to a **sibling** named
  after the URL basename, then mounts it with a **relative** symlink at the
  declared path (relative so it survives differing absolute paths across
  machines). Idempotent — present clones skipped, symlink re-pointed each run.
  No-op when the manifest is absent. Per-dep URL override: `DATADEP_URL_<name>`.
- **Floating-HEAD, no pin** (vs. submodule's gitlink). `git pull` in the sibling
  to update; `make data-deps` does not auto-pull (living repos, not pinned).
- **Reverse wiring is out of scope.** How the *dep* finds the consumer's
  private data (you-decide's `$YOU_DECIDE_PRIVATE_DIR` → brain's
  `who-to-vote-for/`) is the dep's own concern, not this primitive's.

### Why not decouple the substrate walker instead?

The alternative was teaching the go.mod substrate walker to treat some `replace`
directives as clone-only (no `base.manifest` apply). Rejected: more invasive
(touches `discover_ancestors`), stays Go-coupled, and a content dep often isn't
a Go module at all. The separate manifest is strictly simpler.

## Plan

- [x] Resolve the declaration surface → new `construct/data-deps` manifest
      (`<git-url>  <symlink-path>`), not go.mod — language-agnostic
- [x] Define the symlink-mount convention → relative symlink computed from
      symlink-dir → sibling clone
- [x] Wire the clone path → `construct/scripts/clone-data-deps.sh` + `make
      data-deps` target, additive prereq of `bootstrap`
- [x] base.manifest entry so derivatives inherit the script via refresh
- [x] Decide update story → floating-HEAD, manual `git pull`
- [x] Apply to brain↔you-decide as first consumer (manifest + go.mod deletion)
- [x] Document in atlas → `atlas/workflow/setup-and-replication.md` "Data
      dependencies (content peers)" section + workflow index entry
- [x] Verify the script symlink propagates into brain on refresh — DONE
      (2026-06-02): `brain/construct/scripts/clone-data-deps.sh` is a live
      symlink → `../../../ariadne/construct/scripts/clone-data-deps.sh`. The
      stale `UPSTREAM_REFRESH → ../nous/nous/setup.sh` that blocked it was
      removed post-ariadne#32 (brain/Makefile.local:27); brain now refreshes
      from its `substrate ../ariadne` (construct/deps).

## Log

### 2026-05-29
Filed out of the brain cleanup brainstorm (brain#11 externalized you-decide as a
standalone repo). Confirmed with operator: you-decide is a *content dependency*,
not a substrate peer. The data-tree `go.mod` from the first experiment is inert
(no walker reads `data/`) and should be deleted. Initial framing assumed we'd
need to decouple the substrate walker's clone-intent from its apply-intent; the
implementation (next entry) chose a separate manifest instead, which avoids
touching the walker at all.

Correction: an earlier draft referenced a non-existent issue "#46". I
fabricated that reference — there is no issue 46. The rationale it pointed at is
captured inline above ("Why not decouple the substrate walker instead?").

### 2026-05-30 — session summary
Implemented the primitive. Decision: a **separate `construct/data-deps`
manifest** (not the substrate go.mod), because the operator frames brain as
consuming many repos in various shapes — not always a language dependency — so
the mechanism must be language-agnostic and must not drag in substrate-apply.

Landed in **ariadne** (verified on disk via Read):
- `construct/scripts/clone-data-deps.sh` — clone sibling + relative-symlink
  mount, idempotent, python3 for relpath. `bash -n` clean; ran successfully
  against brain (detected you-decide present, mounted symlink, exit 0).
- `Makefile.workflow` — `data-deps:` target + added to `bootstrap:` prereqs +
  help block (verified lines ~127-171).
- `construct/base.manifest` — `symlink construct/scripts/clone-data-deps.sh`
  (verified line 79).

Landed in **brain** (first consumer):
- `construct/data-deps` declares you-decide → `data/life/politics/you-decide`.
- deleted inert `data/life/politics/go.mod` (`git rm`, rc=0).
- the `you-decide` symlink already existed and resolves (README reachable).

Atlas documented in `atlas/workflow/setup-and-replication.md` (new "Data
dependencies" section) + the workflow index entry. (Note: there is no
`atlas/workflow/bootstrap.md` — an earlier draft of this log claimed edits to
that file; it never existed. The real home is setup-and-replication.md.)

**Open (not a blocker for the mechanism itself):**
1. **brain refresh path looks stale.** `brain/Makefile.local` sets
   `UPSTREAM_REFRESH := ../nous/nous/setup.sh`, which does not exist (`nous/nous/`
   is empty). brain's `construct/` itself is healthy (proper nous symlinks +
   the new `data-deps`), but the script symlink won't auto-propagate into
   `brain/construct/scripts/` until brain's refresh is fixed. Workaround proven:
   the script runs fine invoked directly with `TARGET_DIR=<brain> bash
   ../ariadne/construct/scripts/clone-data-deps.sh`. Pre-existing brain issue,
   separate from this feature.

### 2026-06-02 — closeout
- 2026-06-02: closed — data-dep primitive shipped + documented; verified live on brain (clone-data-deps.sh symlink propagated, you-decide mounts + resolves); carrier since folded into construct/deps by #60. actual=judgment estimate (v3 N/A, prior session — see #68)

Revisited to close. Two updates since the 2026-05-30 entry:
- **The "open" brain-refresh blocker is resolved.** `UPSTREAM_REFRESH` was
  removed from `brain/Makefile.local` post-ariadne#32 (line 27 documents the
  removal); brain refreshes from its `substrate ../ariadne`. Verified the
  propagation landed: `brain/construct/scripts/clone-data-deps.sh` is a live
  symlink into ariadne's copy, and `brain/data/life/politics/you-decide`
  resolves. The mechanism works end-to-end on a real consumer.
- **Carrier superseded by #60.** The `construct/data-deps` manifest this issue
  introduced has been folded into the unified `construct/deps` (a `data` row);
  the legacy reader was retired in #60 M5. brain now declares you-decide as
  `data git@github.com:xianxu/you-decide.git data/life/politics/you-decide` in
  `construct/deps`. The *concept* #49 designed is live; only the file it lived
  in changed. `clone-data-deps.sh` still reads both forms.
Closing: the primitive shipped, is documented, and is verified working on brain.
