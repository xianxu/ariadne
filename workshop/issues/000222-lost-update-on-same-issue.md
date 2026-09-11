---
id: 000222
status: open
deps: []
github_issue:
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
---

# Same-issue concurrent trunk edits are last-writer-wins

## Problem

Publishing issue files moved to the trunk in ariadne#207. The route it replaced
computed a merge base and REFUSED when the same issue file had changed on both
sides; the object-database route has no such check, so two agents editing one
issue file is now **last-writer-wins, silently**.

The concrete case is a stale double-claim: agent A claims issue N and publishes
`status: working`; agent B, on a branch cut before that, claims the same issue
and publishes — B's body overwrites A's, and neither is told. Issue files are
the fleet's coordination surface precisely so this cannot happen.

**Why #207 did not simply restore the old check.** A naive three-way against the
merge base would refuse ordinary republishing. Trunk commits are built
out-of-tree and never land on the branch, so after any publish the branch's
merge base with the trunk predates its own last publication: the trunk looks
"changed since we forked" on every subsequent sync. The old arm had the same
false-positive pressure and papered over it by dropping byte-identical files
before conflict detection, which only covers the no-op case.

Recorded rather than hidden: `sdlc claim --help` and
`atlas/workflow/issue-sync.md` both now state the last-writer-wins behaviour.

## Spec

Not yet designed. Sketches worth weighing, in rough order of appeal:

- **Publish-provenance.** Record the blob each publish put on the trunk (a local
  ref such as `refs/sdlc/published/<path>`). On the next publish, compare: if
  the trunk's blob is the one WE last published, nobody else touched it — safe.
  If it is anything else, someone did — refuse. This distinguishes "our own
  earlier publication" from "a peer's edit", which is exactly what the
  merge-base check cannot do. Cost: new local state, and a clone that has never
  published has no record to compare against.
- **Blob-in-our-history.** Accept the trunk's blob if it appears anywhere in our
  branch's history for that path. Needs no new state, but misses content that
  was published from the working tree before ever being committed locally —
  which is exactly what `issue new` does.
- **Refuse only on a REAL divergence.** Compare three contents rather than
  commits: trunk, ours, and the last-known-common version. Same information
  problem as above; the question is only where the common version comes from.

Whichever is chosen, the guard must not fire on the ordinary
publish-then-edit-then-publish loop, or it will be disabled within a day.

## Done when

- Two agents publishing edits to the SAME issue file from different bases: the
  second is refused, naming both versions, rather than silently overwriting.
- The ordinary loop — publish, edit locally, publish again from the same branch
  — is never refused, proven by a test that repeats it several times.
- `sdlc claim --help` and `atlas/workflow/issue-sync.md` drop the
  last-writer-wins caveat that ariadne#207 added.

## Plan

- [ ]

## Log

### 2026-09-11

### 2026-09-11

Filed from ariadne#207's boundary review (BR-5, `deleted-guard-unreplaced`). The
reviewer's own remedy was "document the loss and file a lost-update follow-up"
rather than rebuild the guard inside #207, because the replacement needs a
design and #207 had already grown past its estimate.
