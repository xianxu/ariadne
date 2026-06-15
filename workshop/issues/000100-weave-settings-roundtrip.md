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

## Resolution (reasoned 2026-06-15)

Reasoned through with the operator. The algebra is settled even though we defer building.

**The lift, done right, is a 3-way merge — not a "minimal `B'`".** Minimal-`B'` (the smallest source that re-produces `C'`) is computable but loses intent: it strips anything `B` shares with `A`, so it can't preserve an explicit "pin this value even if `A` also provides it," and the lens law GetPut (`put(B, get(B)) = B`) only holds for already-minimal `B` (otherwise it silently re-normalizes). The better algebra is **minimal *change* to `B`** — a 3-way merge:
- **base** = the last-generated `C`; **ours** = `B` (the sources); **theirs** = the edited `C'`;
- apply `delta(C → C')` to `B`, resolving removals against `A` (drop from `B` if the item is B's; synthesize `$remove` if it's A's).

This preserves B's existing entries (pins survive — they aren't in the delta) and is **unconditionally stable** (empty delta ⇒ `B' = B`, so GetPut holds for *any* `B`). It is exactly what git's 3-way merge does.

**Disambiguation ("who changed") is content-based, not mtime.** mtime is unreliable here — APFS clones (tart) and `git checkout` between branches both reset it. Instead: the sources `A`/`B` are version-controlled (git status/history = definitive "did a source change"); snapshot the **last-generated `C`** as a content baseline. Then current-`C` ≠ baseline ⇒ external `C`-edit; `A`/`B` dirty ⇒ a source changed; both ⇒ **sources win, the `C`-edit is lost** (operator precedence). The baseline-`C` snapshot serves double duty — it is both the disambiguator *and* the 3-way-merge base. One mechanism covers both.

**The sidestep avoids the inverse entirely, and is preferred when the writer is routable.** The lift only exists because the writer (Claude Code) edits the *output* `C`. If the writer edits the *source* `B` (`settings.local.json`), weave only ever runs forward — no inverse, no baseline, no reconcile. Mechanism: **target-scoped (agent-specific) prose** — a `claude`-targeted fragment telling Claude *"edit `settings.local.json`, not `settings.json`, then re-compile."* This is *guidance* (soft); the *deterministic* enforcement is weave regenerating `C` from sources (source-is-truth, output edits transient), so nothing relies on prose-obedience for correctness.
- This surfaced a **third axis** for [[weave-composition-algebra]]: **target/audience** (which backend an artifact is delivered to), alongside **visibility** (export/internal) and **type**. Prose is universal today; target-scoped prose (`AGENTS.claude.md`, or an `internal claude prose …` row) is the same `--target` axis we already use for skill lowering, applied to prose. Deserves its own ticket when the sidestep is built.

## Done when

- **(Done — research)** The round-trip algebra is decided and recorded: lift = 3-way merge (last-gen baseline, sources-win); sidestep = target-scoped prose + source-is-truth regen; both deferred behind accept-transience for #95.
- **(Deferred — build)** When a real multi-layer permission need appears: prefer the **sidestep** (needs the target/audience prose axis); fall back to the **3-way lift** (baseline-`C`, sources-win) only if a writer must edit `C` durably. Reflect the chosen mechanism's limits in [[weave-composition-algebra]].

## Plan

- [ ]

## Log

### 2026-06-14
- Surfaced during the ariadne #95 settings investigation: a stray `settings.json` change was Claude Code's in-session rewrite, not a weave merge bug (weave's `stripMeta(base)` preserves base keys). It exposed the contested-ownership / round-trip problem. Operator wants to reason the algebra through before any mechanism. Cutover default for now: accept transience (option 3).

### 2026-06-15
- Reasoned the algebra through with the operator (see `## Resolution`): lens analysis (union is non-injective ⇒ minimal-`B'` non-unique + loses pins), upgraded to the **3-way-merge lift** (minimal *change* to `B`, last-gen baseline, sources-win — intent-preserving + unconditionally stable); content-based disambiguation (git + baseline, not mtime); the **sidestep** (target-scoped prose + source-is-truth regen) as the preferred path, which surfaced the **target/audience** third axis for the composition algebra.
- **Decision for #95: accept transience.** Nothing contends on permissions across layers today, so build neither lift nor sidestep now. `settings.json` is a generated output; **Claude Code's in-session edits to it are lost on the next `weave compile`** (re-merged from `settings.ariadne.json` + `settings.local.json`). Durable settings are authored in the sources. Cutover proceeds on this basis.
