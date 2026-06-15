---
id: 000100
status: open
deps: []
github_issue:
target: weave-composition-algebra
created: 2026-06-14
updated: 2026-06-14
estimate_hours:
---

# weave settings round-trip: lifting output edits back into sources (inverse-merge / lens)

A research/design ticket — understand the algebra before committing to a mechanism. Relates to [[weave-composition-algebra]] (the `setting` type) and ariadne#97 (the forward DAG fold). **Under active operator reasoning — do not implement until the math is settled.**

## Problem

`settings.json` (call it **C**) is a weave *output*: `C = merge(A, B)` where **A** = the exported settings inherited from ancestor layers, **B** = the repo's own *internal* setting fragment (today `settings.local.json`; under the composition algebra it's just an `internal setting`, NOT a magic filename). But Claude Code **writes to C in-session** (the observed stray: an `enabledPlugins` reorder + a dropped `sandbox.enabled`). Those edits are **clobbered** on the next `make weave` (weave regenerates C from A+B).

To make an output-edit durable we'd have to **lift** it back into the source: given `merge(A,B)=C` and an externally-edited `C'`, find `B'` such that `merge(A,B')=C'`. That is the inverse-of-merge / **lens `put`** problem — the settings analog of the prose visibility fix. (Operator's framing: *"if A + B → C, and C changes to C', how do you change B to B' so A + B' → C'?"*)

## The math (to reason through)

A lens over the forward merge `get = merge(A, ·)`:
- **PutGet:** `merge(A, put(B, C')) = C'` — the lift reproduces the edited output.
- **GetPut:** `put(B, merge(A,B)) = B` — no spurious change when C is unedited (stability).

**Tractability by merge-operator** (from `settingsx`):
- scalar override (local-wins), dict deep-merge, array *replace* → cleanly invertible: `B' = structural-diff(C', A)` (keep in B' only what C' has that A doesn't already provide).
- `$merge_keys` array *union* → NOT uniquely invertible: given `C' = A ∪ B`, many B' union to C'; pick a canonical-minimal `B' = C' \ A`. And a *removal* (C' drops an item A supplies) can't be expressed by union — it must synthesize a `$remove` directive in B'.
- `$remove` → its own inverse case.

So the lift is well-defined for the override fragment and *ambiguous/synthesizing* for unioned arrays — the crux to design.

## Solution spectrum (to choose among)

1. **Sidestep — route the writer to the source.** The lift is only needed because Claude Code writes the *output* C. If its in-session changes land in the *source* B (`settings.local.json`) instead, weave only ever runs `get` (forward) and no inverse is needed. Hinges on Claude Code's write behavior (it already writes some local settings to `settings.local.json`).
2. **Lift — implement the inverse.** A canonical `put(B, C')` per the math above (diff for override; `$remove` synthesis for unioned-array removals). Principled but the union case is genuinely hard/ambiguous.
3. **Accept transience.** C is a pure generated artifact; output-edits are ephemeral; durable settings go in B by hand. Simplest, lossy — fine as the cutover default.

## Reframe (operator, "option c")

`settings.local.json` loses its special "local override" meaning and becomes a plain **`internal setting`** fragment (repo-only, not re-exported) supplied to the `setting` type's algebra — exactly parallel to `internal prose AGENTS.local.md`. Possible rename → `settings.repo.json`. So `setting(R) = merge(exported-settings(ancestors), internal-setting(R))`, and this ticket is about keeping `internal-setting(R)` durable when the merged output is edited externally.

## Done when

- TBD after reasoning — the round-trip is either solved (canonical lift, with the lens laws stated) or deliberately sidestepped (writer→source) or explicitly accepted-transient, with the choice and its limits written into [[weave-composition-algebra]].

## Plan

- [ ]

## Log

### 2026-06-14
- Surfaced during the ariadne #95 settings investigation: a stray `settings.json` change was Claude Code's in-session rewrite, not a weave merge bug (weave's `stripMeta(base)` preserves base keys). It exposed the contested-ownership / round-trip problem. Operator wants to reason the algebra through before any mechanism. Cutover default for now: accept transience (option 3).
