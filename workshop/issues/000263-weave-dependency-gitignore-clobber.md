---
id: 000263
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '9a461fb72c19a6ca782cb3b179b0eeaec0d2a3dc' # card fields mirrored from issue-cards; edit via sdlc
---

# weave refresh clobbers a dependency's committed gitignore block

## Problem

`weave refresh` in `pair:1` rewrote the committed `.gitignore` of its ariadne
dependency checkout (`worktree/pair-slot1/ariadne`), shrinking the
`# BEGIN weave generated` block from ~60 entries to `/construct/generated/weave/`
and leaving the dependency dirty.

Mechanism: `compilePrepared` (`cmd/weave/main.go`) applies each owner layer's
data mounts with `plan.ApplyManaged(fs, owner, mounts, plan.ScopeData)` —
including dependency owners, even with zero mounts. `ApplyManaged` always
rebuilds the owner's managed ignore block from *that checkout's* ownership
inventory. The dependency was never compiled itself, so its inventory is empty
(`{"outputs": []}`) and the rebuilt block drops every entry the committed
`.gitignore` carries (those came from ariadne's own compile, in another
checkout). The same clobber hits any fresh clone whose committed block
predates a local inventory.

## Spec

Invariant: **a weave run removes an ignore entry only when it knows it owned
that path** — i.e. the path is an identity in the checkout's prior ownership
inventory (any scope). Block entries with no inventory provenance (committed
by another checkout's compile) are preserved, merged with the current set.
Exception: the legacy fixed list (`GeneratedRuntimeGitignoreEntries`) still
migrates away, as today.

Consequence: in a checkout without inventory, an entry for an output retired
upstream lingers until a checkout that owned it retires it (it is committed,
so that checkout's commit removes it). Harmless: a stale ignore line for an
absent path.

ARCH: root cause over symptom — fix the ownership rule in `ApplyManaged`, not
by skipping dependency owners in `compilePrepared` (a dependency *with* data
mounts would still clobber).

Related weave-robustness issues: #223 (lowered symlinks tracked in
derivatives), #230 (worktree dependency resolution silently skipped), #241
(refused compile not atomic — already rewrote `.gitignore`), #100
(settings.json round-trip clobbers in-session edits), #96 (peer legibility /
ariadne as shared mutable singleton).

## Done when

- `ApplyManaged` on a checkout with an empty/missing inventory and a committed
  weave block leaves `.gitignore` byte-identical (regression test).
- Entries owned per the prior inventory still retire from the block (existing
  tests pass).
- Legacy fixed-list entries inside the block still migrate away.
- `go test ./cmd/weave/...` passes.

## Plan

- [ ] Regression test: committed block + empty inventory + data apply with no mounts → `.gitignore` unchanged
- [ ] Test: unknown entry preserved alongside new outputs; known entry retires; legacy entry inside block migrates
- [ ] `ApplyManaged`/`managedIgnore`: preserve existing block entries not derived from the old inventory
- [ ] Run weave tests; restore `pair-slot1/ariadne/.gitignore` and re-run `weave refresh` there to verify no diff

## Log

### 2026-09-28
