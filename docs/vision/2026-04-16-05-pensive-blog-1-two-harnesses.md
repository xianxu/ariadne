# Pensive: Blog Post 1 — The Tale of Two Harnesses

**Date:** 2026-04-16
**Status:** Thinking out loud
**Series:** AI Workflow Blog Series (1 of 3)
**Related:** [Exoskeletons](2026-04-16-01-pensive-exoskeletons-for-knowledge-workers.md), [Tweaking Loop](2026-04-16-04-pensive-tweaking-loop-and-personal-processes.md)

---

## The Story Arc

This post is the proof. It earns the right to generalize in posts 2 and 3 by grounding everything in a concrete, lived experience. 🤖{This meta-commentary ("earns the right") reads as outline scaffolding, not blog voice. In the final draft, cut it or fold into a quieter aside — the reader doesn't need to be told the post is earning something.}

The arc: I set out to build features in Parley using AI. Started small — a review mode, a skill runner, some inline annotation handling. Each feature got more ambitious. Somewhere along the way, I realized I wasn't just "using AI tools." I was running two distinct harnesses on the same underlying state (a repo of markdown files), each good at fundamentally different things. That discovery — that the **state is shared but the harnesses are specialized** — is the insight. 🤖{The insight arrives too early — this bold punchline in paragraph one removes the reason to keep reading. Consider making this section a private outline note and letting the journey earn the discovery organically. Also: is "harness" the right word for a general dev audience? It's precise but technical. "Lens" lands more naturally, or commit to "harness" and define it with a one-line analogy early on.}

## Opening

Door B — personal anecdote, placed immediately.

Something like: "I've been building a Neovim plugin called Parley. It started as a way to have AI conversations inside my editor. What happened over the last few weeks surprised me."

Or Door A — question: "Have you noticed that the more you use AI coding tools, the more you want a *different* AI tool for thinking?"

## The Journey (narrative section)

The features got bigger. Walk through 3-4 concrete examples of increasingly complex things built with AI assistance:

1. **Small:** A simple inline review marker (㊷[]). Claude Code writes the Lua, tests it, done. Pure coding task. One harness.
2. **Medium:** A voice style guide generator (/xx-voice-gen). Now there's a design step — what should the output look like? I found myself brainstorming the format in Parley (branching conversation, exploring options), then handing the spec to Claude Code for implementation. Two harnesses, first time deliberately. 🤖{Weakest of the four examples. A skeptical reader thinks "couldn't you just prompt Claude Code to brainstorm formats too?" This needs a specific moment where the single-harness approach failed or felt wrong — what made you reach for Parley instead of just asking Claude Code to iterate?}
3. **Large:** The skill system itself — skills that teach AI how to run recurring workflows. This required iterating on the *concept* (what is a skill? what's the right abstraction?) in one environment, while building the *implementation* in another. The thinking and the building couldn't happen in the same tool.
4. **Meta:** Writing this blog post series. Pensives in Parley to explore ideas. Critique mode to pressure-test arguments. Voice apply to make it sound like me. The harness builds the harness.

Each step is a concrete scene. Short paragraphs. Specific details (not "I built a feature" but "I typed /xx-voice-gen xian and watched it analyze 15 of my blog posts"). 🤖{This is a writing-instruction-to-self — good for the outline, but in the final draft these directives need to become the actual scenes they describe. Flag for yourself: don't let these stay as meta-instructions.}[feedback]

## The Discovery

The moment I noticed: these aren't two tools. They're two harnesses on the same state.

- **Claude Code** is the execution harness. It reads the repo, writes code, runs tests, iterates. It's convergent — it takes a spec and drives toward a concrete implementation. Disposable conversations. The AI does the work.
- **Parley** is the thinking harness. It's where I explore, branch, change my mind, annotate, critique, and shape ideas. It's divergent — it opens up possibility space. Persistent, editable transcripts. The human does the work, AI assists.
- **The repo** is the shared state. Both harnesses read and write the same files. A pensive written in Parley becomes a spec that Claude Code implements. A skill built in Claude Code becomes a workflow that Parley invokes. The state is the integration layer. 🤖{Needs a skeptic's pushback addressed here. A reader thinks: "isn't that just... files? Isn't this what git already does?" The difference is that both harnesses carry AI context shaped by the shared state — it's files-as-protocol-between-agents-and-human, not just version control. One sentence distinguishing this would preempt the objection.}

This is not obvious. Most people think of AI tools as standalone apps. The insight is that the *state* (your files, your repo, your accumulated context) is the real workspace. The tools are just different lenses on it. 🤖{This restates the Story Arc insight for the third time (arc, here, bridge). In the final draft one of these needs to go — probably the arc becomes the lede only after the journey lands, and this becomes the single moment of crystallization.}

## Why This Matters (bridge to post 2)

End with a forward-looking statement. Something like:

"I started thinking about what this means beyond coding. If the pattern is right — shared state, specialized harnesses, human steering — then it's not a developer workflow. It's a knowledge work workflow. But that's the next post." 🤖{Bridge is too thin. "But that's the next post" is a newsletter trick. For developers, a stronger bridge would be a concrete provocation — name one non-coding knowledge task where you've already seen this pattern work (even briefly), so the reader believes the generalization before you make it.}

Don't over-explain. Don't summarize. Just plant the seed.

## Voice Notes

- Ground every claim in a specific moment from the building process
- Use "pretty" naturally ("it was pretty clear by the third feature that...")
- Parenthetical asides for the self-deprecating moments ("I'm a Neovim person, so take this with appropriate salt")
- Bold for the key insight: **the state is shared but the harnesses are specialized**
- Keep the 🤖-> critiques from pensive 01 in mind: don't say "tool for the smart" — say "tool for people who think structurally about their own process"
- The WhatsApp comparison from the letter is tempting here but save it — it belongs in post 3 where organizational stakes are discussed

## Open Questions

- How much Parley-specific detail? Too much and it reads like a product announcement. Too little and there's no proof.
- Do I name Claude Code explicitly, or keep it generic ("a coding agent")?  Probably name it — specificity is the voice.
- How technical can the examples get? The audience for post 1 is developers. Posts 2 and 3 broaden.
