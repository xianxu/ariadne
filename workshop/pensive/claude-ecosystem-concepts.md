---
type: pensive
date: 2026-05-31
topic: Mapping the Claude ecosystem concepts, and where ariadne's bet sits
mode: thoughts
description: A layered concept map of the Claude/Anthropic ecosystem (model → grounding → context → composition → harness/deployment), captured in my framing with grounding tweaks. The real thought underneath the map — Anthropic productizes the attended/generic harness and library-izes the unattended/vertical one, which leaves the substrate/customization band uncommoditized, and that band is exactly ariadne's bet.
references: [AGENTS.md]
---

# Pensive: Mapping the Claude ecosystem, and where ariadne's bet sits

I started ariadne to understand this AI wave by building the infrastructure myself rather than drinking the kool-aid. This is the orientation pass: lay out the key concepts in the Claude ecosystem in my own framing, corrected against the live docs, and find where the commercial ground actually is.

## The thought underneath the map

The cleanest spine is **not** "oneshot vs multishot" — Managed Agents are stateful, long-running, and steerable, so autonomy isn't the dividing line. The real axis is **who owns the loop and where it runs**, and reframed onto **attended vs unattended** (is a human in the loop *during* execution) it explains the whole product lineup:

- **Attended/interactive harness** (Claude Code, Cowork): the hard part is the *steering UX* — diff review, plan mode, permission prompts, slash commands, multi-surface. That UX is **generic across domains** → winner-take-most. Building your own distinctive attended harness almost never pays; the rational move is to take the generic one and **customize it via skills/hooks/CLAUDE.md/brain**. That is ariadne — a customization layer, not a new harness.
- **Unattended/embedded agent** (Agent SDK): there's no interaction UX; the value is all **vertical** (domain tools, guardrails, eval harness, how it slots into a queue/CI). Anthropic can't pre-build every vertical, so it ships a **library**, not a product. That's why the SDK "feels oneshot."

So the asymmetry is deliberate: **Anthropic productized the attended case** (generic UX → one winner) and **library-ized the unattended case** (vertical value → hand you the parts). The market for a *distinctive* third-party harness is thin, and its remaining distinctiveness migrates off loop-logic onto the **deployment envelope** (runs-as-a-service, multi-tenant, web-reachable) — which Managed Agents is already eating. Squeezed from both ends, what's left uncommoditized for builders is the **substrate/customization band**: skills, memory, brain, domain tools. **That band is ariadne's bet, and it looks well-placed.** Caveat to my own framing: ariadne isn't purely attended — routines and autonomous loops are unattended edges riding the *same* substrate. Attended core + unattended periphery, one customization layer serving both.

## The concept map

Bottom-up by layer. Each entry: **my framing → grounding tweak.**

### L0 · Model & training
- **Pretraining** — the whole web distilled into weights. → Distillation is the right intuition; mechanism is next-token prediction.
- **Post-training** — adapt the base model to mimic a knowledge worker's workflow. → It's a *stack*: SFT → preference RL (RLHF/RLAIF) → **RLVR** (RL on verifiable rewards). RLVR is what makes reasoning actually work.
- **Thinking / reasoning** — just generating a longer sequence before answering; amazing that it works, but not shocking next to in-context learning itself. → In Claude: **extended thinking**, budget-capped, **interleaved** with tool calls. Works because trained against a grader, not merely "longer."
- **Model tiers** *(gap)* — Opus / Sonnet / Haiku as a capability↔cost dial; harnesses route per-subtask.

### L1 · Grounding & I/O
- **Tools** — bring deterministic external knowledge/judgment into the thinking process; the determinism is what introduces grounding. → The model only *emits* a `tool_use`; the **harness executes** and returns `tool_result`. A tool is a contract between the model and my deterministic shell — which is why I favor it.
- **MCP** — decentralized, 3rd-party/SaaS-provided way for the model to call external systems; tools are the internal bridge. I favor tools+skills for the tighter, faster loop. → At the model level it's *still* a `tool_use`; MCP is a **transport/packaging standard** (also carries resources + prompts). So the preference is operational, not model-level — and coherent.
- **Structured output** — constrain format (e.g. JSON); either pure statistics or grammar-constrained decoding that rejects rule-breaking tokens. → Both are real; the forced-tool pattern is one implementation.
- **Citations** *(gap)* — grounding *with provenance*: a verifiable pointer back to the exact source span. Anti-hallucination via traceability — same spirit as tools, but for retrieved text.

### L2 · Context management
- **Context window** — size of transcript the model can attend to.
- **O(n²) attention** — quadratic cost, but a *feature*: every token attends to every token, no human-style recency decay or getting-lost-in-a-churning-doc. Omniscience over the window.
- **Prompt caching** *(gap, I raised it)* — KV-cache reuse of a stable prefix neutralizes the per-turn O(n²) cost. Compaction is expensive precisely *because* it busts this.
- **Compaction** — summarize context to fit the window. → Automatic near the limit; lossy *and* cache-busting.
- **Memory** — prose summary of what the user used the agent for. → Underspecified: it's a **writable external store** in three flavors — `CLAUDE.md` (human), **auto-memory** (harness-written), API **memory tool** (model-written). The property that matters is **persistence across context resets** — the antidote to compaction's amnesia. `brain/MEMORY.md` is this, by hand.

### L3 · Composition & customization
- **Skills** — progressively-disclosed prose = system prompt, optionally with deterministic scripts; far more flexible than tools or MCP, and probably the direction of travel. The extreme is the **"skill binary"**: the deterministic binary *is* the skill, the prose is just its help text. → Matches Anthropic's actual design (`SKILL.md` + progressive disclosure). Skills are the one primitive with a *continuous dial* between indeterminism and determinism — tools and MCP are fixed points on it.
- **Subagents** — a separate agent invocation to segregate context; fresh/bounded context reduces hallucination. → Also buys **parallel fan-out** and **fresh-eyes review** (no confirmation bias from work just done).
- **Hooks** *(gap)* — deterministic shell commands fired on lifecycle events (pre/post tool, on stop). The primitive that bolts determinism onto the loop *from outside* — core to the deterministic-shell thesis.
- **Plan mode / gates** *(gap)* — design-before-code. Plan mode is a **hard shell** (Edit/Write mechanically disabled until approval). ariadne's `sdlc change-code` is the **soft-shell, finer-grained** cousin: a gate I'm trusted to run, that withholds code edits while still allowing plan-markdown edits.

### L4 · Harnesses & deployment *(the loop-ownership spine)*
- **Messages API** — raw substrate. I own and run the loop on my infra.
- **Agent SDK** — the loop + ~14 built-in tools + context-mgmt as a **library** (ex-"Claude Code SDK"). For **unattended/embedded vertical** agents, on my infra. "Give the agent a computer."
- **Claude Code** — the generic **attended** harness, as a product; runs on my machine; human in the loop turn-by-turn. Same engine across terminal/IDE/desktop/web/Slack.
- **Managed Agents** — Anthropic-hosted **harness-as-a-service**. Four concepts: **Agent** (model + prompt + tools + MCP + skills) / **Environment** (cloud or self-hosted sandbox) / **Session** (a running instance) / **Events** (the message stream). Stateful, long-running, *steerable & interruptible* — not "oneshot."
- **Cowork** — Claude Code's engine **packaged for non-coders**; desktop, local files/apps, vertical **plugins**. ariadne's target market, attacked by Anthropic from the product side (Jan 2026 research preview, "Claude Code power for knowledge work").

### L5 · Autonomy over time
- **Routines / scheduled tasks / background agents** *(gap)* — cron-on-Anthropic-infra, `/loop`, background sessions. Orthogonal axis to "how autonomous within one run."

### Operational API primitives
- **Batch API** — async bulk, ~50% discount, 24h SLA. The throughput lane: evals, bulk classification, synthetic data, backfills.
- **Files API** — upload-once / reference-by-`file_id`. Dedup + a stable handle for long-running sessions; feeds the sandboxes.

## Open questions

- **Positioning vs Cowork.** Is ariadne *competing with*, *complementary to*, or *deliberately upstream of* Cowork? Cowork attacks "Claude Code for knowledge workers" from product/UX; ariadne attacks the same target from infrastructure/substrate. The answer changes what gets built next.
- **Does the substrate band stay uncommoditized?** The bet is that skills/memory/brain/domain-tools is the band Anthropic won't eat. But skills are an Anthropic primitive and Cowork ships plugins — how much of the substrate is Anthropic already reaching into? Where's the durable moat — the *brain/consistency-prosthesis* layer rather than skills themselves?
- **Hard vs soft shell for the gates.** Plan mode (hard) vs `sdlc change-code` (soft) is a deliberate tradeoff. Is the finer granularity worth giving up mechanical enforcement, or should some ariadne gates become hard shells (hooks that actually block)?
- Should this map split into two artifacts later — a clean **concept reference** (evergreen, maybe an atlas entry) and this **strategic memo** (the spine + the band argument)? Right now they're fused.

## References

- Claude Managed Agents overview — https://platform.claude.com/docs/en/managed-agents/overview
- Claude Code overview — https://code.claude.com/docs/en/overview
- Agent SDK overview — https://code.claude.com/docs/en/agent-sdk/overview
- Claude Cowork — https://claude.com/product/cowork
