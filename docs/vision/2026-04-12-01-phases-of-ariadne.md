# Phases of Ariadne: From Solo Builder to AI-First Company

**Date:** 2026-04-12
**Context:** Brainstorming session on how to bootstrap an AI-first company, starting from a single founder using Parley, scaling to a small team, and eventually to a structure that replaces a 1000-person engineering org.

---

## The Core Thesis

Traditional software development tooling — issue trackers, notification systems, standups, PR review workflows — exists because humans were the execution loop. These are relics of a slow, human-driven process. In an AI-first company, AI runs the execution loop. Humans steer. The tools must be redesigned from first principles around this reality.

The scarce resources are **human attention** and **AI context windows**. Every process, tool, and artifact should be designed to maximize the efficiency of both. Human attention should be spent only where it has irreplaceable value: taste, divergent thinking, and high-stakes judgment. Everything else is AI-autonomous.

**Why these three, specifically:** *Taste* is the ability to recognize quality and coherence without being able to fully articulate the criteria — it is pattern recognition built from deep experience, and it cannot be prompted into existence. *Divergent thinking* is the capacity to question the frame itself: to ask whether the problem being solved is the right problem, to generate options that weren't implied by the brief. AI optimizes within a space; humans redefine the space. *High-stakes judgment* is decision-making under genuine ambiguity, where the cost of being wrong is high, the available information is incomplete, and the answer depends on values — not just facts. AI can surface options and model consequences, but the accountability for the choice must rest with a human who has skin in the game.

In all of such cases, essentially, human is acting as an evaluation function.

## Phase 0: Solo Builder (Now)

**Structure:** One founder + AI swarms.

**What exists:**
- Parley as the AI harness: road mapping, brainstorming, coding, issue tracking, spec writing all happen through AI-steered loops in a single repo
- `workshop/` as the execution space: issues, plans, and history tracked as markdown files in filesystem
- `atlas/` as the system map: high-level pointers for onboarding future humans and agents
- `AGENTS.md` as the proto-constitution: rules for how AI operates within the repo

**What's proven:**
- One person + AI can build what previously required a small team
- The repo-as-world-state model works: all artifacts in one filesystem, versionable, greppable, AI-accessible
- Multi-shot convergence is the right model: vague intent → iterative refinement → concrete outcome

**Key insight from this phase:** The leverage isn't in any single AI response. It's in the accumulation — skills that remember how you work, state that persists across sessions, the ability to re-converge when requirements change.

## Phase 1: The First Team (3 Engineers)

**Structure:** Founder (direction + fundraising) + 3 founding engineers, each a "knowledge worker" who guides AI through feature creation and system operation.

### The Hiring Thesis

Three engineers for redundancy and cross-pollination. They replace the founder in day-to-day execution so the founder can focus on funding and product direction. Each engineer should be capable of guiding AI across the full stack, not siloed into one specialty.

### What Needs to Exist on Day 1

**The Constitution (evolved from AGENTS.md):**
- Domain boundaries: which engineer is steward of which area
- Interface contracts: how domains talk to each other
- Taste document: what "good" looks like, with concrete examples from the codebase
- Architectural bets: what we chose, why, where we're headed

**Onboarding as a Convergence Loop:**
- New engineer reads atlas/ + vision/ + constitution
- They brainstorm with AI on their domain
- Founder taste-checks their first outputs
- AI learns their patterns through skills/memory
- Within a week, they're autonomous in their domain
- The first month is calibration — the engineer internalizes the founder's sense of "good"

### Domain Separation, Not Org Separation

Domains exist for **system coherence**, not turf protection. Any engineer can work in any domain. The boundary prevents accidental coupling and unreviewed contract changes, not cross-pollination.

- **Stewardship, not ownership:** Each domain has a steward — the person who holds the deepest context about why that domain is shaped the way it is. Stewardship rotates as people gain context.
- **Contract changes require steward review.** Interior changes just need tests to pass.
- **Cross-pollination is actively encouraged.** It prevents knowledge silos and makes the team resilient to losing a member.

### Monorepo Structure

Monorepo wins unambiguously at this size. AI agents work in one filesystem. Cross-component refactors are atomic. Shared tooling, one CI pipeline.

```
/
  contracts/          ← shared boundary definitions (steward-reviewed)
    service-a.proto
    service-b.proto
    shared-types.go
  services/
    service-a/        ← domain A internals (anyone + AI, free to change)
    service-b/        ← domain B internals
    service-c/        ← domain C internals
  ops/
    metrics/          ← the 100 metrics, defined as code
    deploy/
  vision/             ← shared intent (human-written, rare changes)
  atlas/              ← system map
  workshop/           ← execution space (issues, plans, history)
```

The hierarchy of human attention: **vision** (shared, human-written, rare changes) → **contracts** (steward-reviewed, infrequent changes) → **internals** (anyone + AI, fast and autonomous). 95% of work happens in internals. Human review concentrates on the top two layers.

### Language Choice Through the Contract Lens

Strongly typed languages (Go, Rust, TypeScript) are preferred — not because the engineers need guardrails, but because **the AI swarms do.** A type system is the cheapest boundary enforcement available. The compiler rejects invalid cross-boundary changes automatically. Humans only review intentional contract changes.

Go is a strong candidate: simple, typed, fast compilation, exported/unexported distinction makes boundaries visible at a glance.

### Coordination Model

The 3 engineers don't need an issue tracker, standups, or notification systems. They need:

- **A shared vision document** (`vision/`) that defines direction, priorities, and constraints. Updated after brainstorming sessions, not daily.
- **Weekly sync** on boundary/contract changes — "constitutional conventions." Everything else is async.
- **AI-to-AI coordination through shared filesystem artifacts.** Each engineer's AI swarm writes state summaries. Other swarms read them before cross-domain changes. Humans only intervene at genuine conflicts.

### Operating Systems at AI Speed (The WhatsApp Model)

Each knowledge worker doesn't just build their domain — they operate it. AI swarms monitor, diagnose, and fix. The infrastructure philosophy:

**Logs as filesystem (S3):**
```
s3://ops/service-a/2026/04/12/hour-14/requests.jsonl
```
- Directory structure enables progressive discovery (browse broad, drill down)
- JSONL format — AI parses it with the same tools it uses for code
- No Elasticsearch, no query language, no dashboards. Just `ls` and `grep`.

**Simple telemetry:**
- No more than 100 metrics per product
- Each metric has a plain-text description of what it means and what "bad" looks like
- AI monitors all 100 simultaneously — it doesn't need alerting rules, it reads the full picture and notices anomalies
- Investigation follows the same loop: metrics → logs → code → diagnosis → fix

No PagerDuty. No runbooks. The AI is the runbook. The human gets a summary after the fact.

## Phase 2: The 10-Person Structure (1000-Dev Equivalent)

**Structure:** 1 CTO → 10 knowledge workers → AI swarms. Each knowledge worker replaces a current 10-person team and moves 10x faster. Net: 100x gain over traditional org of equivalent scope.

### Hierarchical Constitutions

The CTO's constitution governs the 10 knowledge workers. Each worker's constitution governs their AI swarms. Constitutions are the API between organizational layers.

- **CTO level:** company-wide architectural bets, cross-domain contracts, priority stack
- **Worker level:** domain-specific conventions, internal structure, operational parameters

### The Coherence Problem at AI Speed

The real scaling bottleneck isn't throughput — it's coherence. 10 knowledge workers each moving at AI speed means divergence happens fast. Worker A's AI swarm refactors the auth system while Worker B's swarm builds a feature depending on the old auth API.

**Solution: contracts as the coordination primitive.** Contract changes propagate through the constitution hierarchy. AI agents read contracts before cross-boundary work. When a contract change is proposed, AI traces impact across all domains and surfaces conflicts before humans discuss.

### The Sync Cadence

- **CTO ↔ workers:** infrequent, high-altitude. Direction, not details.
- **Worker ↔ worker:** on-demand, triggered by contract changes. AI flags when sync is needed.
- **Worker ↔ AI swarm:** continuous but mostly autonomous. Human intervenes at genuine ambiguity.

## Scaling Beyond (Open Question)

The 10-worker structure covers the equivalent of a 1000-developer org. Scaling beyond — to the equivalent of 10,000 developers — is an open question. Likely requires:

- Hierarchical knowledge workers (leads of leads)
- More sophisticated contract management (versioning, compatibility guarantees)
- AI-to-AI negotiation on boundary changes without human involvement

This is future territory. The immediate path is Phase 0 → Phase 1 → Phase 2, each validated by lived experience before the next.

## Key Principles Across All Phases

1. **Repo as world state.** All artifacts — code, specs, contracts, operational config, metrics definitions — live in the filesystem. AI reads and writes files. No separate systems.

2. **Human attention on boundaries, AI autonomy in interiors.** The membrane model: free activity inside components, controlled exchange at the surface.

3. **Contracts are the coordination primitive.** Not tickets, not meetings, not notifications. Explicit interface definitions that both humans and AI can read and enforce.

4. **Taste can't be automated.** The knowledge worker's irreplaceable value is judgment under ambiguity — knowing which direction to go when the spec is half-formed. Systems should surface decision points, not require humans to discover them.

5. **Simplicity is load-bearing.** AI makes it easy to build sprawling systems fast. The constitution must enforce simplicity: if you can't explain why a component exists in one sentence, delete it.

6. **Build the company with the tool.** Ariadne's Phase 1 product is proven by building Ariadne itself with it. If the harness can't coordinate 3 engineers building the harness, it can't coordinate anyone.
