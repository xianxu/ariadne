---
id: '000201'
status: done
started: 2026-08-25T11:51:47-07:00
created: 2026-08-21
updated: 2026-08-25
estimate_hours: 1.09
actual_hours: 1.40
---

# Boundary-review artifacts persist the harness transcript, not just the review

## Problem

Every `sdlc close` / `milestone-close` review document in `tools#4` opens with a
line that is not review content:

```
Ignoring 6 permissions.allow entries from .claude/settings.json: this workspace
has not been trusted. Run Claude Code interactively here once and accept the
trust dialog, or set projects["<repo>"].hasTrustDialogAccepted: true in
~/.claude.json.
```

The reviewer's stderr is being concatenated into the persisted artifact, so the
preamble lands once per round. In `tools#4` it now appears at
`close-review.md:18, 251, 485, 674, 878, 1043` — six occurrences, one per round,
and it was flagged by the reviewer itself at rounds 3, 4, 6, 7 and 8 as a
finding it could not converge because the cause is outside that repo.

Two distinct defects sit behind it:

1. **The artifact captures stderr it should not.** A review document should hold
   the review. Harness diagnostics belong on the operator's terminal, not
   committed into `workshop/plans/`. Anything the reviewer writes to stderr
   before the verdict block should be dropped or routed.

2. **A recurring finding with no home converges nowhere.** The reviewer
   correctly re-reports it every round because nothing in `tools` can fix it, so
   it costs a finding slot per round forever. Worth considering whether the
   review protocol should let a finding be dispositioned "external, tracked at
   `<ref>`" once, rather than re-litigated each boundary.
