---
id: 000036
status: open
deps: [000031]
created: 2026-05-26
updated: 2026-05-26
estimate_hours:
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

## Spec

Three small fixes, sequenced:

### Fix 1 — `gh issue close` field lookup

- In the archival code path of `sdlc push`, locate where `github_issue:` is read from the frontmatter.
- Verify with a unit test that issues with no `github_issue:` skip the gh call entirely (don't fall back to `created:` or any other field).
- Verify with a unit test that issues with `github_issue: 42` pass `"42"` (not `"github_issue: 42"`) to `gh issue close`.
- Tests should use a fake runner (the existing `gitx.Runner` pattern) to avoid real gh invocation.

### Fix 2 — make push fall-through

- In `Makefile.workflow`, the `push:` target. Current shape (lines visible during push):

```make
push:
    @if [ -x bin/sdlc ]; then \
        bin/sdlc push $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge); \
        exit 0; \
    fi
    @branch=$$(git branch --show-current); \
    ...
```

The `exit 0` is inside the conditional block but the rest of the target body still runs because make treats each `@` line as separate. Move the sdlc branch into a single shell invocation that's the *entire* target, or use a Makefile-level conditional rather than shell-level.

Fix shape (proposed):

```make
push:
    @if [ -x bin/sdlc ]; then \
        exec bin/sdlc push $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge); \
    fi
    @# below only runs if bin/sdlc is absent (fallback)
    @branch=$$(...); ...
```

`exec` replaces the shell process, ensuring nothing else in the target runs.

### Fix 3 — plan judge classifier false-negative

Two paths to consider:

(a) **Tighten the prompt** — the plan judge prompt should instruct the agent to emit one of the literal clean-signal strings (`No issue files changed` or a new explicit one like `Plan review: clean`) when there are no actionable findings, even if it has positive prose to add. The classifier then reliably picks up the signal.

(b) **Broaden the classifier** — add a pattern like `^.*?\b(no corrections needed|looks clean|review.{0,20}clean|approved)\b.*$` as additional clean-signals.

I lean (a) — explicit instruction in the prompt is more robust than scraping arbitrary positive language. The classifier should stay deterministic; the prompt should make the judge produce signal-shaped output.

Update the plan judge prompt in `cmd/sdlc/internal/judge/prompts.go` to instruct the agent: *"If you have no actionable corrections, emit the literal string `Plan review: clean` on its own line at the start of your response, then optionally add positive feedback below."* Update the classifier to recognize that signal. Add a unit test.

## Plan

To be detailed when starting. Rough shape:

- [ ] **M1 — Fix gh issue close field lookup.** Locate the bug in push.go / archival path; add tests with fake runner; fix.
- [ ] **M2 — Fix make push fall-through.** Edit Makefile.workflow; verify locally that `make push` no longer triggers `pre-merge-checks.sh` after sdlc push completes.
- [ ] **M3 — Fix plan judge classifier false-negative.** Tighten prompt to emit explicit clean-signal; add classifier recognition; add unit tests.

## Log

### 2026-05-26 — issue created

All three bugs surfaced during the same `make push` invocation on ariadne main (pushing 12 local commits including #31 + #32 close + walk-through artifacts). The push ultimately succeeded with `NO_JUDGE=1` bypassing the plan-judge false-negative; the archival succeeded (11 issues moved including #31 + #32) despite the gh-issue-close errors; the post-push fall-through to pre-merge-checks.sh produced a confusing trailing failure.

None of these are critical — push completed, archive completed, work shipped. But all three undermine the operator experience of the canonical sdlc push flow. Each is small (sub-day) and bounded; they're bundled in one issue because they all surfaced in one session and conceptually live in the "polish sdlc push" bucket.

deps: [000031] — these are all followups to sdlc push code that landed in #31.
