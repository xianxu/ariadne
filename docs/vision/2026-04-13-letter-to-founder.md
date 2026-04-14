# Letter to Alamin: NexHealth's AI Transformation

**Date:** 2026-04-13
**From:** Xian
**To:** Alamin Uddin, Founder & CEO

---

Alamin,

I've been spending pretty much all my free time building with AI-native workflows — not just using AI tools, but restructuring how I work around them. The leverage is real, and it's not what most people think it is. I want to make a case for why NexHealth needs to go through this transformation, and why it can't happen the normal way.

## The Thesis

Here's what changed: LLMs can now reliably execute multi-step cognitive tasks. Not just generate text — THEY RUN THE LOOPS. Read context, make decisions, use tools, iterate. This isn't the chatbot era anymore. **This is a new operating system for knowledge work, and it just booted up.**

Every company will need to be restructured around this. The question isn't whether — it's when, and whether we're ahead or behind when it happens.

For NexHealth specifically: we sit at the intersection of healthcare coordination, integrations, and workflows — all domains where AI loops can replace layers of manual process. The companies that figure out how to run on AI-native workflows will operate at 3-5x the output per person and ship higher quality product. The ones that bolt chatbots onto existing processes will wonder why they're losing. I've seen this movie before at WhatsApp — a tiny team with the right architecture outperforming orgs 10x their size. The same dynamic is about to play out again, except the leverage isn't Erlang this time. It's AI.

Now, I want to be honest: **AI-native development is not a solved problem.** The raw capability is here — agents can write code, run tools, iterate on feedback loops. But how to sustain this in a specific domain, with production stakes, with a real team? That playbook is still being written. Nobody has figured out, say, how to keep AI-generated code maintainable over a year, or how to build institutional knowledge in specs that agents can reliably use, or exactly where human checkpoints need to sit for healthcare-grade correctness. This is new territory. And that's precisely why the skunk works matters — it's not just about deploying known techniques faster. It's about **developing the discipline** for our domain before competitors do. Whoever builds that playbook first has a compounding advantage that's pretty hard to replicate.

## Why Piecemeal Won't Work

The natural instinct is to sprinkle AI across the org: add a copilot here, an AI feature there, let each team experiment. This feels safe and inclusive. It will also fail.

AI transformation isn't a feature. It's a paradigm shift. It changes how work gets structured — not just how individual tasks get executed. When AI runs the process loops and humans steer, you need different everything:

- **Hiring.** You need people who can direct and evaluate AI output, not just produce output themselves. The skillset is closer to "technical director + architect" than "individual contributor who codes fast." Pretty different from how we hire today.
- **Performance evaluation.** Output-per-person changes meaning when one person with AI leverage produces what a team of five used to. How do you evaluate someone whose primary skill is steering AI to convergence on hard problems, and their ability to set up such environment where AI can converge faster? Our current perf system doesn't even have a vocabulary for this.
- **Team structure.** Small teams with AI leverage outperform large teams without it. But you can't run a large org with one set of rules and a skunk works with another — unless you explicitly create that space.
- **Process.** Code review, sprint planning, roadmap estimation — all of these assume human-driven execution. They become bottlenecks when the execution layer moves at AI speed. Like alerts, they're easy to create and surprisingly hard to decommission[^1].

You can't change all of this piecemeal across an existing org. People have different capabilities, some will resist change, emotions and politics ensue — not out of malice, just part of how people react to fundamental changes and how they are incentivized. 

## The Proposal: A Skunk Works

I'm proposing a small, founder-sponsored team with an explicit mandate:

**Prove that a 3-person AI-native team can outperform a 15-person traditional team on a real NexHealth problem.**

Concretely:

1. **Pick one high-value domain.** Synchronizer is a good area, particularly in the Cloud EHR where we are behind. Create a new Cloud EHR integration or two, AI first.

2. **Staff it tiny.** 2-3 people max. People who are curious, adaptable, and comfortable with ambiguity. Not the most senior — but people who can steer AI, iterate fast, while having the tool building mentality. I have Erik S. in mind for now, who's the most advanced in AI coding after Jon. 

3. **Give it different rules.** Different hiring bar (optimize for AI-leverage aptitude, not years of experience in X framework). Different evaluation. Different process (AI-native workflows, not sprint ceremonies designed for human-speed work). Any tools they want to use.

4. **Protect it.** This is the critical part, and the part only you can do. This team must report directly to you — not through the existing engineering org. I'll be blunt: our current engineering leadership is not equipped to lead this. The skills required — deep intuition for AI capabilities, willingness to throw out established process, comfort with radical ambiguity — are pretty fundamentally different from what got us here. Routing this through the existing chain will produce what it's optimized to produce: incremental improvements within the current paradigm. That's the opposite of what we need. A skunk works dies when it gets dragged into consensus-driven decision-making with people who haven't seen the paradigm work and don't understand what's possible. It needs air cover from the founder, and **a direct reporting line to you**, both for air cover and for your product insights.

5. **Timebox it.** 90 days to show results. If a tiny team with AI leverage can demonstrably outperform on speed, quality, or both — that's the proof point to expand. If it can't, we've lost very little.

## Why Only the Founder Can Do This

This isn't something a VP or director can champion:

- It requires changing incentive structures (hiring, perf), which is a founder-level decision.
- It requires protecting a team from organizational and interpersonal gravity. Only the founder has the authority to say "this team operates differently, and that's by design."
- It requires making a bet on a new paradigm before the evidence is overwhelming. By the time the evidence is overwhelming, it'll be too late — the companies that moved early have compounding advantages.

The hardest part isn't the model capability — that's here. And it's not the unsolved technical questions — a small, focused team will figure those out faster than any committee. The hardest part is creating the organizational space to even run the experiment. That's a founder problem.

## What I'm Asking

Three things:

1. Your sponsorship to stand up this skunk works.
2. A real problem domain to prove it on — something where success is measurable and meaningful to the business.
3. Protection from premature scaling or dilution. Let it be small, weird, and fast until it proves itself.

I've been building and operating with these workflows daily. The leverage is real, but it requires a fundamentally different way of working — one that the current org isn't designed to support. I'd like to show you what's possible. Seeing it is more convincing than reading about it.

**WHAT WOULD YOU DO IF YOU WERE NOT AFRAID?A**

— Xian

[^1]: I wrote about this in the context of process design — every process accrues like technical debt. The AI transformation requires actually removing process, not adding more.
