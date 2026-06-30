---
type: target
slug: agent-binary-handoff-schema
status: active
created: 2026-06-30
updated: 2026-06-30
sources:
  - "ariadne#147 — the boundary-review verdict handoff (free-text stdout regex-parsed → spurious `unknown`) crystallized the principle"
  - "ariadne#139 — close-after-verdict, blocked on a robust verdict, surfaced the cost of an unstructured handoff"
---

# Target: Agent↔binary handoffs are schema'd — every stochastic→deterministic boundary crosses a CUE contract

Every place where a **stochastic** component (an LLM agent or subagent) hands a result back to a **deterministic** component (the `sdlc` binary, or any code that gates or branches on that result) must cross a **schema** — a CUE model the agent generates *against* predictably and the binary *validates and reads* robustly. We do not parse an agent's prose. The agent's freedom is bounded to **filling a known shape**; the binary's read is a **validation**, not a heuristic. When the shape is violated, that is a genuine, detectable protocol error — not a formatting near-miss the binary silently mis-reads or papers over.

The shape we defend: the **CUE model is the single source** for the handoff's vocabulary and structure, and every consumer derives from it — the prompt that instructs the agent renders its accepted values *from* the model; the parser validates *against* the model; the downstream gates read the model's semantics; a drift guard fails when any of them diverge. This is the same discipline `issue.cue` brought to the issue lifecycle ([[issue-lifecycle]]), generalized to the *boundary itself*: wherever a stochastic process speaks to a deterministic one, the words it may say and the way it must say them are modeled, not hand-synced across a prompt string, a regex, and an enum.

Why it must be a schema and not "just be careful": prose parsing is *structurally* fragile — an agent that buries the answer in a sentence ("the verdict stands: FIX-THEN-SHIP") defeats even a careful regex, and a regex loose enough to catch it self-triggers when the agent reviews the parser. And hand-enumerating the legal values in N places guarantees drift. A schema dissolves both: the agent fills a block the model defines; the binary validates that block; one edit to the model updates every consumer.

## Why now

The boundary-review **verdict** handoff (ariadne#147) was the crystallizing case. `sdlc` invoked the review agent and captured its free-text stdout, then regex-scanned for a `VERDICT:` line. Reviewers routinely emitted the verdict in prose instead ("the verdict stands: **FIX-THEN-SHIP**"), so the parser — correctly strict, to avoid self-triggering — fell back to `unknown` on perfectly good reviews. Worse, the legal verdict tokens were hand-synced across **four** places (the prompt, the output-contract constant, the parser regex/enum, and the consumers), with the enum's own comment begging maintainers to keep them in step. ariadne#139 (close should finalize only after the verdict) then surfaced the cost: a sound "halt on `unknown`" policy is unusable while `unknown` is a frequent false alarm. The fix was not a looser regex — it was to **model the verdict in CUE** (`verdict.cue`), have the agent emit a fenced block the model defines, and have the binary validate it. That move is not specific to verdicts; it is the general law for every agent↔binary boundary, and worth defending as one.

## What this is NOT

- **Not provider-specific forced output.** Forcing a tool-call / `--output-format` is a per-agent feature and would break ariadne's agent-neutrality. The *schema* is the contract; the *format* the agent emits to satisfy it (a fenced block, a frontmatter'd markdown file) stays agent-neutral — anything that can write text can comply.
- **Not "schema every agent call."** Only the **handoffs the binary gates or branches on** need this. An agent doing exploratory reading, or prose meant for a human, is not a schema'd boundary. The trigger is: *deterministic code makes a decision from this output.*
- **Not a rewrite of the prompt layer.** The prompt still instructs in natural language — it just renders its legal values *from* the model and asks for the modeled shape, rather than restating an enumeration that drifts.
- **Not fail-open by default.** A missing/invalid handoff is a protocol error to surface (halt, investigate), not a value to guess. A prose fallback may exist transitionally, but the schema'd path is authoritative.

## Open questions

- **Which boundaries adopt this next?** The pre-merge tri-state judge tokens (`CLEAN`/`INFO`/`FAILURE`), the change-code plan/estimate judges, and any future `sdlc state`-style agent digest are all candidates. Order by how much deterministic code branches on them.
- **Where do handoff schemas live?** The review verdict fits the noun layer (`construct/vocabulary/*.cue`) because it *is* a vocabulary. Some handoffs may be transport shapes, not nouns — do those want a separate `construct/handoff/*.cue` home, or does the vocabulary layer absorb them?
- **Migration posture.** How long do prose fallbacks live before a boundary goes fail-closed? Per-boundary call, or a global policy?
- **How far down does "deterministic consumer" reach?** A human reading a sidecar is not a schema'd boundary; a binary gating on it is. Where exactly is the line when a handoff feeds both?
