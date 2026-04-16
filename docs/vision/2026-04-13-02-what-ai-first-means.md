# What "AI First" Actually Means

**Date:** 2026-04-13
**Companion to:** [Letter to Alamin](2026-04-13-letter-to-founder.md)

---

Over the last year, I've been building a Neovim plugin called [Parley](https://github.com/xianxu/parley.nvim) — about 29K lines of Lua, 60+ modules, 600+ tests, 700+ commits. I didn't know lua, still don't. I've never written a Neovim plugin before. The entire thing was built by AI agents, steered by me. I reviewed specs and verified behavior, make key architecture steering, just not coding. That sounds crazy until you think about it — we already do this with compilers. Nobody reviews the assembly output[^1].

That experience is why I'm writing this. When I say "AI first," I don't mean "use Copilot more." The difference is pretty hard to see until you've lived it, so let me make it concrete. I need those 700 commits to see that future clearly. 

## Three Levels

Most companies are at Level 1. We need to go to Level 3. 

**Level 1: AI-assisted.** Human drives, AI helps. You write code, Copilot autocompletes. You write a doc, ChatGPT polishes it. The workflow stays the same — you just have a faster horse. This is where most of NexHealth is today. There's also a cultural hazard at this level: AI makes it easy to produce output that *looks* plausibly thorough — docs, plans, analyses — but is hollow. It becomes hard for anyone to tell what's real thinking and what's generated fillers.

**Level 2: AI-augmented.** Human directs, AI executes chunks. You describe a feature, an agent writes the code, you review the PR. Think Claude Code or Cursor in agent mode. The workflow is still human-shaped, but bigger pieces get delegated. Quickly, human becomes the bottleneck in code reviews — and reviewing AI-generated code at volume is exhausting in a way that reviewing human code isn't. The sheer throughput drains reviewers. Sync Core is experimenting with this (Jon/Erik). 

**Level 3: AI-first.** AI runs the loops. Human steers. The workflow itself is restructured around what AI is good at (executing multi-step tasks, tireless iteration, pattern matching across large codebases) and what humans are good at (judgment, taste, knowing what "correct" means in context). **This is where the 3-5x leverage lives.** No one's here.

The gap between Level 2 and Level 3 isn't better tools. It's rethinking the entire loop — how work is specified, executed, verified, and evolved.

## Before and After: Building a Cloud EHR Integration

### Before (how we do it today)

Picture a typical new integration. A team of 2-3 engineers, one serving as PM/lead.

1. PM/lead writes requirements 
2. Engineering lead breaks it into stories, estimates in story points (a sprint planning meeting)
3. Engineers pick up stories, write code, submit PRs (2-3 sprints)
4. Code review rounds — back and forth, nit-picks, style debates (days per PR)
5. QA cycles, bug fixes, more PRs (another sprint)
6. Deploy, monitor, hotfix (ongoing)

Timeline: 2-3 months. Team: 2-3 engineers. 

On the other hand — what if the execution layer is AI?

### After (AI-first)

Same integration. Team of 1. The engineers are the PM[^2].

1. **Spec the boundaries, not the details.** Write a lightweight spec: what data flows where, what the API contract looks like, what "correct" means for each sync direction. Don't over-specify — specifying too much in imprecise human language is actually counter-productive. I learned this the hard way on Parley: when I tried to build an OAuth flow, I realized the domain was way more complicated than I initially understood. There would've been no way to specify all the nuances upfront. Treat the spec as a sketch that evolves. The model will pull in its world knowledge.

2. **AI runs the implementation loops.** Point an agent at the spec. It reads the existing codebase, flushes out details in the spec, plans the implementation, writes the code, writes the tests. When it hits ambiguity, it surfaces a checkpoint: "here's what I'm assuming about appointment status mapping — is this right?" The engineer reviews the assumption, not the code. Such discussions are synthesized and stored for future use.

3. **Ground it in reality, fast.** The agent runs tests continuously — not just the tests it wrote, but regression tests, integration tests against real EHR sandbox APIs. The feedback loop is minutes, not days. When something breaks, the agent iterates. The engineer only gets involved for judgment calls: "the EHR returns inconsistent data for this edge case — which interpretation should we use?" Right, this is fundamentally different from human code review as the verification mechanism.

4. **Verify at the decision layer.** The engineer doesn't review every line of code. They review decisions: Is the data model right? Does the sync logic handle the edge cases that actually matter? Are we building the right thing? The agent writes hundreds of tests to prove correctness. The engineer's job is to make sure "correct" is correctly defined. For reference: when I built Parley this way, I went from 0 tests to 600+ — the agent writes them compulsively, which is actually what you want.

5. **Evolve the spec as you learn.** Every integration uncovers surprises — the EHR API does something weird, a data format isn't what the docs say. In the traditional flow, these discoveries get lost in Slack threads and PR comments. In the AI-first flow, the spec file gets updated, the agent re-plans, the tests adapt. The knowledge stays in the repo, not in someone's head.

Timeline: 2-3 weeks. Team: 1 engineer. — roughly a **5-10x improvement**. And the code comes with far better test coverage and ability to keep updated while underlying EHR shift. 

## What I Actually Do All Day

This is pretty different from what most people imagine when they hear "AI coding."

- **Architect the structure.** I define how workflow files connect: where issues live, where specs live, how the agent discovers what to work on, what constitutes a "checkpoint" that needs my eyes. This is the meta-work — building the harness that makes AI execution sustainable. I've iterated on this over 400+ commits in Parley, evolving from raw prompts to a structured system of issues, specs, lessons, and living workflows.
- **Steer at checkpoints.** The agent surfaces decisions. I confirm, correct, or redirect. On Parley, that looked like: "No, the exchange model should be the single source of truth — don't duplicate state in the UI layer." At NexHealth, it would be: "No, appointment status should map this way because of how downstream consumers use it." This is where domain knowledge matters and where human judgment is irreplaceable.
- **Ground the agent.** Write test harnesses. Set up linting rules. Create guardrails. Pretty much like writing unit tests, except the "unit" is the agent's behavior, not just the code. I even built `make test-changed` that runs tests based on which spec files changed — so the agent can verify faster.
- **Evolve the spec.** As the agent implements and I review, the spec gets sharper. Edge cases get documented. The spec becomes the living source of truth — not code comments, not Confluence, not someone's memory.

The key insight (and this one I borrowed from the AlphaZero analogy): **the human is the evaluation function.** AlphaZero needs a board evaluation function to distinguish winning from losing positions. AI agents need a human to distinguish correct from merely plausible. My job is to make that evaluation sustainable, not to do the execution myself.

## Why This Isn't "Just Use Copilot More"

| | AI-Assisted (Level 1-2) | AI-Native (Level 3) |
|---|---|---|
| **Who drives?** | Human writes, AI helps | AI executes, human steers |
| **Unit of work** | A function, a PR | A feature, a system |
| **Verification** | Human code review | Automated tests + human judgment at decision points |
| **Knowledge lives in** | People's heads, Slack, Jira | Repo: specs, tests, living workflow files |
| **Speed bottleneck** | Human typing speed, review cycles | Human judgment speed (much faster) |
| **Team size for a feature** | 5-8 | 2-3 |

Sprint planning assumes human-speed execution — pretty pointless when the agent finishes a "sprint's worth" of code in a day. And here's the big one: **humans stop reviewing code.** This is the hardest shift, and it's the one that makes everything else work. Code review as gatekeeping becomes a bottleneck when the agent generates more code in a day than a reviewer can meaningfully read. More importantly, it's the wrong layer to verify at. Humans review *decisions* — is the spec right, is the data model right, does the behavior match what we want.

But — and this is critical — you can't just stop reviewing code and call it a day. You need to **replace code review with something better.** The infrastructure to make this safe is real work:

- **Test coverage that's actually comprehensive.** Not aspirational. The agent writes tests compulsively, but someone needs to make sure the test harness itself is sound — that you're testing the right things, not just a lot of things. I built `make test-changed` on Parley to run tests based on which spec files changed, so the agent can verify fast and targeted.
- **Linting and static analysis as guardrails.** Rules that catch classes of mistakes deterministically. Over time, lessons the agent learns (soft knowledge in `lessons.md`) should migrate into hard linting rules. Code that you once caught in review should become a rule the agent can't violate.
- **Spec-driven verification.** The spec is the contract. If the code matches the spec and the tests pass, you don't need to read the diff. But that means the specs have to be good — maintained, precise where it matters, and actually connected to the test suite.
- **Staged autonomy.** You don't go from "review every line" to "review nothing" overnight. You start with more human checkpoints, and as the test harness and guardrails prove themselves, you dial down. On Parley, I started reviewing a lot. Now, 700+ commits in, my review surface is specs and test results, not code.

I built 25K lines of Lua this way without reading most of the code. The tests and specs were my review surface, not the diffs. Once you build the infrastructure to support this, the entire workflow unlocks. If you don't — if you just remove code review without replacing it — you're vibe coding at scale. And if you keep code review, you're stuck at Level 2. The ceiling on how fast you move is how fast humans can read diffs.

## What It Takes

To actually operate at Level 3:

1. **People who can steer.** Not everyone is comfortable working this way. You need to know what "right" looks like without personally writing every line. The skill is closer to tech lead or architect than senior IC.

2. **A test-first culture, for real.** Not "we aspire to write tests." Actually test-first, because the tests are how you ground the AI. Without tests, you're just vibe coding[^3] — the AI produces plausible output and you have no way to verify at scale. This is a feature, not a bug — it forces you to finally do what we've always said we should do.

3. **Specs as living documents.** Not the kind that get written once and immediately rot. Specs that evolve as the agent implements, checked into the repo, serving as source of truth for both humans and AI. I keep these in a `specs/` folder with an index file — the agent discovers them progressively, like how you'd onboard a new engineer.

4. **Tolerance for a different rhythm.** No sprint planning. No standup (who would you report status to — the agent knows its own status). Spec review, checkpoint review, and continuous verification. It looks weird from outside.

5. **Willingness to measure differently.** Not story points. Not PR count. Did the integration ship? Does it work? How fast? How robust? Outcomes, not activity.

That's a pretty different operating model. And that's exactly why it needs to be a protected skunk works, not a mandate from eng leadership to "adopt AI tools."

---

Happy to do a live walkthrough — actually building something with the AI-native workflow so you can see it in action. It's a lot more visceral than reading about it.

[^1]: Of course, with great power comes great responsibilities. The grounding mechanism (tests, specs, verification) is what makes this work. Without it, you're just vibe coding at scale. Pretty terrifying for production systems.

[^2]: Actually, this is one of the underappreciated implications. When AI executes at speed, the bottleneck shifts to "does the team know what to build?" The PM → engineer handoff becomes overhead. The people closest to the problem should be steering the AI.

[^3]: Andrej Karpathy coined this term. Fun for side projects. Terrifying for production systems.
