# chat-retry

submitted: status.md
outcome: ok

Row 18 for chat. A previous round wrote the answer doc, but its write to the parent was refused (exit 3, the lock),
so the link never landed and the item is unlinked. This round relinks the answer doc by its `parent:` and
`marker:`, and creates no second one.

## Reply

- Item 1: linked status.side/2026-10-16-01-s4-cost.md, which an earlier round wrote before its write was refused.
- collaboration-mode map is a legacy key; its chat sections are read as @@chat@@
