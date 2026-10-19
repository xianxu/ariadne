---
id: 000454
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# ARCH principle: persisted formats evolve additively; readers ignore unknown fields; refusal only when ignoring is unsafe

## Problem

The ARCH registry has no principle for how persisted formats evolve. ARCH-SECURE covers reading input written by an *older* version, but nothing says what an older reader should do with data from a *newer* writer. Designs default to refusing unknown records (pair#684's first draft, precedent #198). The operator (2026-10-18): "typical robust strategy would be compatibility, that old binary will just ignore new attributes ... both simpler and also more robust ... if this is not your default, I wonder if we should add [it] to architecture".
