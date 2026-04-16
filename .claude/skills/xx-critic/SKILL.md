---
name: xx-critic
description: "Use when the user wants AI to critique a document, generate ideas, and surface weaknesses. Invoked as /xx-critic <path-to-file>. Complements /xx-review which processes human feedback on AI text."
---

# Critic

Read a document and add inline critiques as `🤖->` lines. The goal is to help the human evolve their thinking, not to rewrite their text.

## Usage

```
/xx-critic <path-to-file>
```

## Process

1. **Read the file** from the supplied path
2. **Read the voice guide** from `~/.personal/xian-writing-style.md` for tone context
3. **Analyze the document** for:
   - Claims without grounding (assertions that need evidence or examples)
   - Underexplored analogies or ideas (introduced then dropped)
   - Framing problems (elitist, instrumentalizing, vague)
   - Missing "why" (reader would ask "but why?" and the text doesn't answer)
   - Contradictions or tensions the author may not have noticed
   - Structural issues (too many ideas for one piece, sections that don't connect)
4. **Insert `🤖->` critique lines** directly after the text they refer to
5. **Write the file** back. Do not wait for approval.

## Output Format

Critiques are inline, placed immediately after the paragraph or section they address:

```
Some text the author wrote about X.

🤖-> [Specific critique. Not "this is weak" but why it's weak and what would make it stronger.]
```

## Rules

- **Critique, don't rewrite.** The human evolves the text. You surface what's off.
- **Be specific.** "This section is vague" is useless. "You claim X but don't ground it. Your own experience with Y would be the evidence" is useful.
- **Be direct.** No hedging. No "you might consider." Say what's wrong.
- **Preserve existing 🤖-> lines.** Don't duplicate or remove prior critiques. Add new ones only.
- **Preserve existing ㊷[] markers.** Don't touch them.
- **Don't critique everything.** 3-7 critiques for a typical document. Focus on what matters most.
- **Suggest the fix direction, not the fix.** "The exoskeleton analogy needs the specific axes named" not "here's a rewritten paragraph."
- **Preserve all other text exactly as-is.**
