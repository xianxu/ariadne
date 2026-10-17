# discussion-retry

submitted: design.md
outcome: ok

Row 18. A previous round wrote design.side/01-why-per-file.md, but its write to the parent was refused (exit 6,
stale), so the side doc exists unlinked. This round finds it, unlinked, by its `parent:`, `marker:` and section and links it, taking no
new number. The document has no `collaboration-mode:` key, so it is discussion.

## Reply

- Locks: linked design.side/01-why-per-file.md, which an earlier round wrote before its write was refused.
