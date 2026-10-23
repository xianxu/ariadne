---
type: side-doc
parent: ../design.md
marker: "first in the section"
---

# Notes: the update stamp goes first in the section

_From [Session file](../design.md), section "Freshness"._

First is where a reader looks before reading on, so a stale section shows itself at once. One consequence: a write
that restamps the line changes the section's hash, so a held hash for that section goes stale and is read again.
