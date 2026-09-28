---
id: 000263
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '3ec13c5a2fe0cca34aeb35544c17ee775f1c14bf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T15:38:02-07:00
flow: {kind: quick, provenance: inferred, spec: "9a47d4ff", done: "6e35a782"}
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

- [x] Regression test: committed block + empty inventory + data apply with no mounts → `.gitignore` unchanged
- [x] Test: unknown entry preserved alongside new outputs; known entry retires; legacy entry inside block migrates
- [x] `ApplyManaged`/`managedIgnore`: preserve existing block entries not derived from the old inventory
- [x] Run weave tests; restore `pair-slot1/ariadne/.gitignore` and re-run `weave refresh` there to verify no diff

## Log

### 2026-09-28
- Reproduced with a unit test: committed block + empty inventory + `ApplyManaged(nil, ScopeData)` shrank the block to `/construct/generated/weave/`.
- Design correction: first draft also excluded the legacy fixed list (`GeneratedRuntimeGitignoreEntries`) from preservation; that still clobbered, because the real committed block contains `/AGENTS.md`, `/CLAUDE.md`, `/.claude/settings.json` (exact outputs that are also in the legacy list). Git history (`8d2d08bb`, #239) shows the legacy list only ever lived *outside* the block, where `managedIgnoreText` still migrates it — so inside-block entries need no legacy rule. Dropped it.
- Live verification: restored `worktree/pair-slot1/ariadne/.gitignore`, ran the patched `weave compile` in `pair-slot1/pair` (same `compilePrepared` path as `refresh`, minus the git pulls): 107 actions applied, dependency and pair both `git status` clean.
