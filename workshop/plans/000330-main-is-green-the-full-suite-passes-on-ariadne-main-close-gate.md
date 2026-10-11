---
gate: boundary-review
issue: 330
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T00:47:49-07:00"
      dispatched: "2026-10-11T00:43:12-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: 'Quarantine end-condition lives only in #330''s Log; owning issue #325 does not mention deleting the two Quarantine calls'
          detail: 'workshop/issues/000325 has no quarantine/test-name reference, so #325''s fix can land with the skips still in place. Add a Done-when line to #325 (delete both testfix.Quarantine(t, "#325", ...) calls) via sdlc issue publish, or note why not in --verified.'
          family: removal-path-unrecorded-at-owner
          round: 1
        - id: BR-2
          severity: Minor
          title: TestReviewConcurrencySchedules quarantined whole (14 subtests) though only milestone-close/branch failed, once in two runs
          detail: 'Quarantining inside the failing subtest keeps cancel/relock/etc. coverage while #325 is open.'
          family: quarantine-scope-wider-than-failure
          round: 1
        - id: BR-3
          severity: Minor
          title: 'kill-0 liveness rule implemented twice with different readings (repolock: non-EPERM error = dead; weave staging: non-ESRCH = alive)'
          detail: repolock.go comment says "only ESRCH is gone" but code returns dead for any non-EPERM error; align the code or comment, or share one helper.
          family: single-liveness-predicate
          round: 1
        - id: BR-4
          severity: Minor
          title: atlas sdlc-binary.md keeps "About 4 minutes" beside the measured 7.5-8 minutes; CLAUDE.md overlay repeats the stale figure
          family: stale-doc-figure
          round: 1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-11T00:51:07-07:00"
      dispatched: "2026-10-11T00:50:33-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 000325 Done when gains a line requiring both testfix.Quarantine calls deleted, with a grep check; committed in 86401ddf.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Quarantine moved inside the 20s start-timeout branch (review_concurrency_unix_test.go:140) after cancel/release; other subtests unskipped.
          round: 2
        - id: BR-3
          disposition: addressed
          note: repolock.go:176 now returns !errors.Is(kill0, ESRCH), matching comment and weave staging; pid-1 EPERM test pins it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: atlas sdlc-binary.md and AGENTS.local.md say about 8 minutes; CLAUDE.md is a gitignored generated overlay of AGENTS.local.md.
          round: 2
      recipe: milestone-review
      reviewed: 86401ddf688ac0b18a2437972d81e411ec6658be
      blocked: false
---

# Gate ledger — ariadne#330 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T00:47:49-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `removal-path-unrecorded-at-owner` Quarantine end-condition lives only in #330's Log; owning issue #325 does not mention deleting the two Quarantine calls
  workshop/issues/000325 has no quarantine/test-name reference, so #325's fix can land with the skips still in place. Add a Done-when line to #325 (delete both testfix.Quarantine(t, "#325", ...) calls) via sdlc issue publish, or note why not in --verified.
- **BR-2** [Minor] `quarantine-scope-wider-than-failure` TestReviewConcurrencySchedules quarantined whole (14 subtests) though only milestone-close/branch failed, once in two runs
  Quarantining inside the failing subtest keeps cancel/relock/etc. coverage while #325 is open.
- **BR-3** [Minor] `single-liveness-predicate` kill-0 liveness rule implemented twice with different readings (repolock: non-EPERM error = dead; weave staging: non-ESRCH = alive)
  repolock.go comment says "only ESRCH is gone" but code returns dead for any non-EPERM error; align the code or comment, or share one helper.
- **BR-4** [Minor] `stale-doc-figure` atlas sdlc-binary.md keeps "About 4 minutes" beside the measured 7.5-8 minutes; CLAUDE.md overlay repeats the stale figure

## Round 2 — 2026-10-11T00:51:07-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — 000325 Done when gains a line requiring both testfix.Quarantine calls deleted, with a grep check; committed in 86401ddf.
- BR-2 — addressed — Quarantine moved inside the 20s start-timeout branch (review_concurrency_unix_test.go:140) after cancel/release; other subtests unskipped.
- BR-3 — addressed — repolock.go:176 now returns !errors.Is(kill0, ESRCH), matching comment and weave staging; pid-1 EPERM test pins it.
- BR-4 — addressed — atlas sdlc-binary.md and AGENTS.local.md say about 8 minutes; CLAUDE.md is a gitignored generated overlay of AGENTS.local.md.

## Open findings

(none — every finding has been disposed)
