# chat

submitted: status.md
outcome: ok

Rows 12 to 16, and the chat edges:
- the heading `### Operator instructions (open)` matches the name `"operator  instructions"` (case, whitespace and
  the trailing annotation are ignored), and its `#### Notes` subsection inherits chat;
- the fenced `## Operator instructions` is text, not a heading;
- the items sit above the topmost marker, and the review mode is copy-editing (frontier on), yet they are answered;
- the text above the first heading and `## Needs you` take the default, discussion, so the marker there opens a side
  doc, numbered, beside the dated answer docs;
- item 4 links an old `qa/` answer file, and its follow-up extends that file;
- the answer doc for "what is the quota burn?" lost its item before this round, so it is closed and left untouched.

## Changes since last read

```diff
 3. how big is the backlog? → [answer](status.side/2026-10-15-01-backlog.md)
+   - and by repo?
-6. what is the quota burn? → [answer](status.side/2026-10-15-02-quota.md)
-4. why did CI fail? → [answer](qa/op/2026-10-14-01-ci.md)
+4. why did CI fail, and is it fixed? → [answer](qa/op/2026-10-14-01-ci.md)
+1. what does S4 cost?
+2. move #353 to :2
```

## Reply

- `collaboration-mode.chat` names 'Operatr notes', which matches no heading.
- Item 1: answered in status.side/2026-10-16-01-s4-cost.md.
- Item 2: done (moved #353 to :2); recorded in status.side/2026-10-16-02-move-353.md.
- Item 3: extended status.side/2026-10-15-01-backlog.md with your follow-up.
- Item 4: extended qa/op/2026-10-14-01-ci.md with your follow-up.
- Needs you: your question opened status.side/01-why-now.md.
