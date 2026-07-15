---
id: 000036
status: done
deps: [000031]
created: 2026-05-26
updated: 2026-06-03
estimate_hours:
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

Resolution (2026-06-03 audit): only M2 was still live; M1 and M3 were fixed
incidentally by later work. M1/M3 marked done-by-other-work (not milestone-closed
here → close uses `--no-verdict`).

- [x] **M1 — gh issue close field lookup → already fixed.** `push.go:456-459`
  reads `github_issue:` via `issue.GetField(fm, "github_issue")`, guards empty,
  and passes the *number* to `ghClient.IssueClose` — not the `created:` string.
  The wrong-field bug is gone.
- [x] **M2 — make push fall-through → fixed (this issue).** Root cause: the
  delegation ran `bin/sdlc push; exit 0` in the first recipe line, but make runs
  each recipe line in its own shell, so it proceeded to the next line (the legacy
  fallback ending in `$(MAKE) pre-merge` → interactive `pre-merge-checks.sh`).
  Fixed with a **Makefile-level conditional** (`ifneq ($(wildcard bin/sdlc),)`)
  that *excludes* the fallback when the binary exists, rather than a shell
  `exit 0` that only skips one line. Verified both branches via `make -n push`
  (binary present → only `bin/sdlc push`; absent → fallback). Note: `pull-request`
  / `merge` / `fetch` share the old pattern (same low impact — legacy make
  targets; binary is canonical) — left as a flagged follow-up, not in scope here.
- [x] **M3 — plan-judge classifier false-negative → fixed by #70.** #70 rewrote
  `classify.go` onto the `VERDICT:` contract (`ParseVerdictToken`): a positive
  review emits `VERDICT: CLEAN|INFO|SHIP` and classifies as pass. The literal-
  string scraping that mis-classified detailed-but-positive reviews as `Failure`
  is gone (`cleanRE` survives only as a thin legacy fallback).

## Log

### 2026-06-03 — closed (M2 fixed; M1/M3 resolved elsewhere)
- 2026-06-03: closed — M2 fixed: make push fall-through removed via Makefile-level conditional (ifneq wildcard bin/sdlc) excluding the legacy fallback when binary present; verified both branches with make -n push. M1 already fixed (push.go reads github_issue: correctly), M3 fixed by #70 (VERDICT-contract classifier). --no-verdict (Mx not per-milestone-reviewed; M1/M3 fixed by other work), --no-judge (6-line Makefile fix, verified via make -n), --no-atlas (no new surface).; review verdict: not-run
- Audited all three bugs against current code. **M1** already correct in
  `push.go`; **M3** fixed by #70's verdict-contract rewrite; **M2** was the only
  live one — fixed here via a Makefile-level conditional (excludes the fallback
  when `bin/sdlc` exists), dogfood-verified with `make -n push`.
- Scope flag: the same `@if [x bin/sdlc]; then …; exit 0; fi` + separate-fallback
  pattern lives in `pull-request`/`merge`/`fetch`; same fall-through, same low
  impact (binary is the canonical path post-#51). File a fresh sweep issue if it
  bites.
- `--no-verdict`: M1/M2/M3 are Mx rows but none went through a per-milestone
  review (M1/M3 fixed by other issues; M2 is a 6-line Makefile fix). `--no-judge`:
  trivial Makefile change, verified by `make -n` rather than an LLM pass.

### 2026-05-26 — issue created

All three bugs surfaced during the same `make push` invocation on ariadne main (pushing 12 local commits including #31 + #32 close + walk-through artifacts). The push ultimately succeeded with `NO_JUDGE=1` bypassing the plan-judge false-negative; the archival succeeded (11 issues moved including #31 + #32) despite the gh-issue-close errors; the post-push fall-through to pre-merge-checks.sh produced a confusing trailing failure.

None of these are critical — push completed, archive completed, work shipped. But all three undermine the operator experience of the canonical sdlc push flow. Each is small (sub-day) and bounded; they're bundled in one issue because they all surfaced in one session and conceptually live in the "polish sdlc push" bucket.

deps: [000031] — these are all followups to sdlc push code that landed in #31.
