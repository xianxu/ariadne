---
gate: boundary-review
issue: 325
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T08:15:34-07:00"
      dispatched: "2026-10-12T08:07:10-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: repolock_test.go hard-codes time.Minute hang ceilings instead of testfix.HangCeiling
          detail: 'cmd/sdlc/internal/repolock/repolock_test.go:315,348,361,397 can import testfix (no cycle) but restate the ceiling as a literal. Same family, separate case: 9 weave sites use time.Minute because they cannot import cmd/sdlc/internal/testfix (generator_lifetime x3, release/main, staging/stage, plan/publication, weavefs/runner_owned x2, runner_setup), plus 2 ExecGit timeouts in acquire/git_refresh_test.go. A shared weave test constant would make the 1 min vs 2 min ceilings explicit.'
          family: hang-ceiling-single-source
          round: 1
        - id: BR-2
          severity: Minor
          title: review_concurrency_unix_test.go copies executeSDLCTestCommand's body to pass ctx and stderr
          detail: cmd/sdlc/review_concurrency_unix_test.go:170-177 rebuilds the root command by hand. Extending cmd/sdlc/repolock_test.go:504 with a variant that takes a context and a stderr writer would keep one helper. This is the only such copy in the window.
          family: test-command-helper-reuse
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: 9933e8dbbb8aabd0ef85c067501b3169bc73c8c7
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T09:01:13-07:00"
      dispatched: "2026-10-12T08:53:14-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: repolock_test.go:317,350,363,399 use testfix.HangCeiling; all 9 weave sites and both acquire ExecGit timeouts use testgate.HangCeiling (pkg/testgate/ceiling.go:12); no time.Minute hang ceiling remains in the window.
          round: 2
        - id: BR-2
          disposition: addressed
          note: executeSDLCTestCommandWith(ctx, w, args) at cmd/sdlc/repolock_test.go:513 is the one runner; executeSDLCTestCommand wraps it and review_concurrency_unix_test.go:171 calls it instead of rebuilding the root.
          round: 2
      findings:
        - id: BR-3
          severity: Minor
          title: testgate_test.go hand-rolls a needle-watching pipe reader that duplicates testfix.Watch (ARCH-DRY)
          detail: 'pkg/testgate/testgate_test.go:203-223 collects output and signals on "[test-gate] waiting", the same job as cmd/sdlc/internal/testfix/wait.go:19-51 Watch. The HangCeiling move already set the precedent: move Watch into pkg/testgate and alias it from testfix. Instances in window: 2 (testfix.Watch, the testgate goroutine).'
          family: event-wait-helper-single-source
          round: 2
      recipe: small-diff-review
      isolation: checkout
      reviewed: f48c5d4d54460da13f0f9e9fa95d48f5a637bde8
      blocked: false
---

# Gate ledger — ariadne#325 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T08:15:34-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `hang-ceiling-single-source` repolock_test.go hard-codes time.Minute hang ceilings instead of testfix.HangCeiling
  cmd/sdlc/internal/repolock/repolock_test.go:315,348,361,397 can import testfix (no cycle) but restate the ceiling as a literal. Same family, separate case: 9 weave sites use time.Minute because they cannot import cmd/sdlc/internal/testfix (generator_lifetime x3, release/main, staging/stage, plan/publication, weavefs/runner_owned x2, runner_setup), plus 2 ExecGit timeouts in acquire/git_refresh_test.go. A shared weave test constant would make the 1 min vs 2 min ceilings explicit.
- **BR-2** [Minor] `test-command-helper-reuse` review_concurrency_unix_test.go copies executeSDLCTestCommand's body to pass ctx and stderr
  cmd/sdlc/review_concurrency_unix_test.go:170-177 rebuilds the root command by hand. Extending cmd/sdlc/repolock_test.go:504 with a variant that takes a context and a stderr writer would keep one helper. This is the only such copy in the window.

## Round 2 — 2026-10-12T09:01:13-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — repolock_test.go:317,350,363,399 use testfix.HangCeiling; all 9 weave sites and both acquire ExecGit timeouts use testgate.HangCeiling (pkg/testgate/ceiling.go:12); no time.Minute hang ceiling remains in the window.
- BR-2 — addressed — executeSDLCTestCommandWith(ctx, w, args) at cmd/sdlc/repolock_test.go:513 is the one runner; executeSDLCTestCommand wraps it and review_concurrency_unix_test.go:171 calls it instead of rebuilding the root.

### Raised

- **BR-3** [Minor] `event-wait-helper-single-source` testgate_test.go hand-rolls a needle-watching pipe reader that duplicates testfix.Watch (ARCH-DRY)
  pkg/testgate/testgate_test.go:203-223 collects output and signals on "[test-gate] waiting", the same job as cmd/sdlc/internal/testfix/wait.go:19-51 Watch. The HangCeiling move already set the precedent: move Watch into pkg/testgate and alias it from testfix. Instances in window: 2 (testfix.Watch, the testgate goroutine).

## Open findings

- **BR-3** [Minor] `event-wait-helper-single-source` testgate_test.go hand-rolls a needle-watching pipe reader that duplicates testfix.Watch (ARCH-DRY)
