---
id: '000172'
status: done
started: 2026-07-14T08:07:22-07:00
created: 2026-07-13
updated: 2026-07-14
estimate_hours: 8.07
actual_hours: 7.42
---

# sdlc painpoint audit

## Problem

The `sdlc` binary shapes every workflow transition, but we've never *measured*
where it hurts. Two obstacles:

1. **introspect (#169) will NOT surface this — by design.** Its extract/cluster
   prompts target *taste signals* ("how the **user** wants future sessions to
   behave") — i.e. user→agent friction (redirects, endorsements). Gate-bypasses
   are **agent→binary** friction the user often never sees. Worse, `extract.md`
   explicitly drops "anything specific to one project's domain — we want
   transferable taste," and clustering is activity-scoped (5 buckets, no tooling
   bucket). So substrate ergonomics get filtered out. introspect gives at most a
   faint incidental shadow.
2. **A naive grep is contaminated.** Full-transcript grep for `--force`/`--no-*`
   also catches help-text output, the constitution (which lists every flag), and
   meta-conversations like this one — inflating counts ~10-40×.

So this needs a **direct, precise instrument** — not introspect, not grep.
