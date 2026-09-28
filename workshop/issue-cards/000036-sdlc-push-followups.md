---
id: '000036'
status: done
created: 2026-05-26
updated: 2026-06-03
actual_hours: 0.4
---

# sdlc push / judge followups — bugs surfaced during first dogfood push

## Problem

The first end-to-end dogfood of `sdlc push` (ariadne main, 2026-05-26, shipping the #31 + #32 + walk-through work) succeeded but surfaced three bugs that didn't block the push but should be fixed before this becomes the canonical operator path:

### Bug 1 — `gh issue close` pulls wrong frontmatter field

When archiving `status: done` issues, sdlc push attempts `gh issue close` for issues that declare a `github_issue:` field in their frontmatter. Two issues in this push (#024 and possibly others) triggered:

```
==> Closing GitHub issue #created: 2026-05-04...
  [!] gh issue close created: 2026-05-04 failed: gh issue close created: 2026-05-04: exit status 1
invalid issue format: "created: 2026-05-04"
```

sdlc push is reading `created:` (the issue-creation-date frontmatter field) instead of `github_issue:` (the linked GitHub issue number). Symptom: pass a string like `"created: 2026-05-04"` to `gh issue close` as the issue number; gh rejects it.

Likely a wrong-grep-pattern or wrong-field-lookup in the archival code path. Probably in `cmd/sdlc/push.go` or a helper it calls.

Doesn't block the push — falls through with `(continuing)` — but logs noisy errors and leaks the wrong contract to operators.

### Bug 2 — `make push` falls through to legacy `pre-merge-checks.sh` after `sdlc push` succeeds

After `sdlc push` completes successfully (push + archive done), the Makefile's `push:` target continues executing and invokes the legacy `pre-merge-checks.sh` (the inline shell logic that predates sdlc). That script is interactive (asks for check selection via `/dev/tty`) and fails in non-tty contexts:

```
[ok] Done.

Pre-merge checks:
  1. [dry    ] Check DRY principle
  ...
Select checks [yyyyy] ... /dev/tty: Device not configured
make[1]: *** [pre-merge] Error 1
make: *** [push] Error 2
```

Should `exit 0` after `sdlc push` succeeds. The fall-through to the legacy path is a remnant of how the make target was set up — `bin/sdlc` invocation runs, but the rest of the target body keeps going.

Fix: in `Makefile.workflow`, the `push:` target's sdlc branch should `exit 0` after the binary returns 0, instead of relying on early `&& exit 0` that may not be wired right.

### Bug 3 — `sdlc judge plan` classifier marks positive-but-detailed responses as failure

The plan judge produced a clearly-positive review on the second retry:

> No corrections needed. The #31 close is exemplary: explicit deferral note for M8, per-milestone Log entries, actuals filled, atlas update referenced in the commit history.

But sdlc classified the response as `failure` and aborted the push, because the `Classify()` function in `cmd/sdlc/internal/judge/classify.go` only recognizes specific literal "clean signal" strings (per the prompt-defined ones: "No DRY violations found", "No PURE violations found", "Everything is in sync", "No issue files changed"). Anything else falls through to `Failure`.

This is fine for the deterministic "no issues at all → emit clean signal" path. But the plan judge can produce positive feedback in scenarios where issue files DID change (status flipped, plan ticks updated, log entries added) — those are valid changes worth a positive review, not "no issue files changed."

The judge prompt should be tightened OR the classifier should add a path for "explicitly approves the diff" patterns. Either way, false-negative `Failure` classification on clean reviews undermines the discipline — operators learn to `--no-judge` past it, defeating the point.
