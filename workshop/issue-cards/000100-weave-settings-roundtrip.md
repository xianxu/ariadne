---
id: 000100
status: open
created: 2026-06-14
updated: 2026-06-14
estimate_hours:
github_issue:
---

# weave settings round-trip: lifting output edits back into sources (inverse-merge / lens)

## Problem

`settings.json` (call it **C**) is a weave *output*: `C = merge(A, B)` where **A** = the exported settings inherited from ancestor layers, **B** = the repo's own *internal* setting fragment (today `settings.local.json`; under the composition algebra it's just an `internal setting`, NOT a magic filename). But Claude Code **writes to C in-session** (the observed stray: an `enabledPlugins` reorder + a dropped `sandbox.enabled`). Those edits are **clobbered** on the next `make weave` (weave regenerates C from A+B).

To make an output-edit durable we'd have to **lift** it back into the source: given `merge(A,B)=C` and an externally-edited `C'`, find `B'` such that `merge(A,B')=C'`. That is the inverse-of-merge / **lens `put`** problem — the settings analog of the prose visibility fix. (Operator's framing: *"if A + B → C, and C changes to C', how do you change B to B' so A + B' → C'?"*)
