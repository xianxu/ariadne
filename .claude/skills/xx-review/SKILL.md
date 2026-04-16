---
name: xx-review
description: "Use when a file contains ㊷[] inline feedback markers from parley.nvim review mode. Processes user comments, answers questions, and rewrites marked sections. Invoked as /xx-review <path-to-file>."
---

# Inline Review

Process ㊷[] inline feedback markers in a file, following parley.nvim's review protocol.

## Usage

```
/xx-review <path-to-file>
```

## Marker Format

The ㊷ marker supports strictly alternating user and agent turns:

```
㊷[user comment]{agent response}[user reply]{agent response}...
```

- `[]` brackets = user turns (comments, corrections, requests)
- `{}` braces = agent turns (questions, acknowledgments, explanations)
- Odd section count (1, 3, 5) = marker is "ready" for agent to process
- Even section count (2, 4) = marker awaits user response (skip it)
- Markers inside fenced code blocks are ignored

## Process

1. **Read the file** from the supplied path
2. **Parse all ㊷ markers**, identifying:
   - Ready markers (odd section count): process these
   - Pending markers (even section count): leave these untouched
3. **For each ready marker**, read the user's comment and full conversation history in the marker, then:
   - If the comment is a correction or rewrite request: apply the change to the surrounding text and remove the marker
   - If the comment needs clarification: add an agent question `{your question here}` to the marker (making it even/pending)
   - If the comment is acknowledged and done: remove the marker entirely
4. **Write the modified file** back to the same path
5. **Report** what was changed and what markers remain pending

## Rules

- Preserve all text outside of markers exactly as-is
- A marker refers to the text **before** it, up to the previous natural boundary (paragraph, bullet, section). In a typical reading flow, you read and comment at the end of what you just read
- Only modify text that a marker refers to (the preceding paragraph, bullet, or section)
- **Exception**: if the marker explicitly mentions a bigger or different scope (e.g. "this whole article", "the next section about XXX", "all the bullet points above"), follow what the marker says instead of defaulting to the preceding block
- When removing a marker, leave the corrected text in place with no trace of the marker
- When adding an agent question, append `{question}` inside the existing marker
- Respect the user's voice and style in the surrounding document
- Do not rewrite sections that have no markers
