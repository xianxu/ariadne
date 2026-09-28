---
id: '000026'
status: done
created: 2026-05-19
updated: 2026-05-19
estimate_hours: 0.5
actual_hours: 0.3
---

# tart: TART_SUFFIX knob for parallel VMs per repo

## Problem

`.tart/Makefile` currently hard-codes one VM per consuming repo:

```make
TART_VM ?= $(REPO_NAME)-test
```

Two VMs in the same repo (e.g., testing two recipient configs
simultaneously in `nous`, or "Emma's view" alongside "operator's
view" of a shared brain) require the operator to override `TART_VM`
fully — and the name is operator-invented, no convention. Easy to
forget cleanup, easy to collide with a teammate's mental model.

The 90% case is still "one VM per repo" — `make tart` should keep
working unchanged. We just want a friendly knob for the parallel
case that picks a name we can talk about.
