---
id: 000058
status: working
deps: []
github_issue:
created: 2026-06-01
updated: 2026-06-03
estimate_hours: 1.5
---

# milestone-close — base review window on prev milestone close (close inter-milestone gap)

## Problem

`sdlc milestone-close` computes its fresh-eyes review window as **"first commit
referencing `#<issue> Mx`"^..HEAD** — the parent of the first commit tagged with
that specific milestone. So a commit that references the issue (`#<issue>`) but
**not** a milestone — a side-quest, a fix, a plain `#N` commit made between two
milestone closes — can fall in the gap: it post-dates the previous milestone's
window and is *excluded* as the base of the next one. It escapes both reviews.

Surfaced in #57: `c52f482` ("#57: dev-aliases — build to owner bin/") referenced
`#57` but not `M2`. M1's window ended at its close (`76a2828`); M2's window based
at `699635d^` = `c52f482`, excluding it. The M2 judge happened to review it as
compensation (it was the base), but that's luck, not guarantee — a commit two or
more before the milestone's first `Mx` commit would simply vanish from review.

This defeats the milestone-review's purpose (every change reviewed before the
next milestone) precisely for the unplanned commits (`side-quest:`, fixes) that
most warrant a fresh eye.

## Spec

**Base each milestone's review window on the previous review boundary, not on
the first `#N Mx` commit.** Concretely:

- The window for `Mx` is `[prev-boundary .. HEAD]`, where `prev-boundary` is the
  commit at which the *previous* milestone-close's review ended — i.e. the prior
  milestone-close commit carrying this issue's `Review-Verdict` trailer (the
  `Review-Window: …..<head>` it recorded).
- For the **first** milestone (no prior boundary), base on the issue's branch
  start (the merge-base with `main`, or the first `#N` commit's parent) — same
  as today's first-commit behavior, just not keyed to the `Mx` tag.

This guarantees every commit since the last review boundary — including
inter-milestone `#N`-but-not-`Mx` commits — lands in exactly one milestone
window, so nothing slips the net.

Mechanism options for the implementer to weigh (don't over-specify):
- Read the prior milestone-close commit's `Review-Window` trailer head SHA and
  use it as the new base (authoritative — it's exactly where the last review
  stopped).
- Or `git log --grep "Review-Verdict:" -- <issue file>` for the most recent
  prior boundary on this branch.

**Non-goals:** don't change the verdict trailers, the judge prompt, or the
`#N Mx[: ]` subject matcher (that's the *detection* fix from #57's side-quest,
already landed). This is purely the *window base* computation.

## Done when

- A `Mx` milestone-review window includes inter-milestone `#N`-but-not-`Mx`
  commits made since the previous milestone close (the #57 `c52f482` shape).
- The first milestone still bases on the branch start (unchanged behavior).
- A regression test pins it: a fixture branch with `[M1 work][M1 close][#N
  side-quest, no Mx][M2 work][M2 close]` asserts the M2 window covers the
  side-quest commit.
- `helptext/milestone-close.md` "Diff window" wording updated;
  `go test ./cmd/sdlc/...` green.

## Plan

Focused single-pass change (one milestone — not split).

- [ ] In `cmd/sdlc/milestoneclose.go`, replace the window base (`first #N Mx
      commit`^) with the previous review boundary: prior milestone-close
      `Review-Window` head, else branch start for the first milestone. Keep the
      `#N Mx[: ]` matcher only for *which milestone* is being closed, not for the
      window base.
- [ ] Regression test (real-git fixture per `close_test.go`'s pattern): an
      inter-milestone `#N` commit between M1 close and M2's first commit must
      fall inside M2's window.
- [ ] Update `cmd/sdlc/helptext/milestone-close.md` ("Diff window") + verify
      `go test ./cmd/sdlc/...` green; rebuild.

## Log

### 2026-06-01
Filed from #57's M2 review, which caught `c52f482` slipping between the M1 and
M2 auto-windows. The judge reviewed it as the M2 base and found it correct, so
no harm landed — but the gap is real: inter-milestone `#N`-but-not-`Mx` commits
(side-quests, fixes) can escape review entirely. Fix is to base each window on
the previous review boundary so coverage is gap-free. Pairs with the #57
side-quest that hardened the `#N Mx close:` subject *matcher* (`b083bab`-era) —
that fixed detection of the close commit; this fixes the window it reviews.
