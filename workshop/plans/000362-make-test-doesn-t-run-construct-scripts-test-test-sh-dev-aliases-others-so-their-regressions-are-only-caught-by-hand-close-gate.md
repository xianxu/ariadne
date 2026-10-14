---
gate: boundary-review
issue: 362
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T22:26:13-07:00"
      dispatched: "2026-10-13T22:24:48-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: TestConstructScripts runs the shell runner without a deadline
          detail: scripts/test/scripts_test.go:22 uses exec.Command with no context, so a hanging construct test stalls the shard until the 30m go test timeout. Use exec.CommandContext with testgate.HangCeiling. This is the only subprocess call in the window.
          family: unbounded-subprocess-wait
          round: 1
        - id: BR-2
          severity: Minor
          title: Runner header says "macOS /bin/bash 3.2 locally", but it runs whatever bash is on PATH
          detail: construct-scripts.test.sh:6 says the runner uses /bin/bash 3.2 on macOS. That only holds when /bin/bash comes first on PATH; with Homebrew bash first it runs bash 5. Reword to "the bash on PATH". This is the only such claim in the window; the atlas wording is accurate.
          family: doc-claim-accuracy
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: e246d7e5c7b97298bf7de66455e6693b3d9899ee
      blocked: false
    - "n": 2
      timestamp: "2026-10-13T22:28:34-07:00"
      dispatched: "2026-10-13T22:27:15-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'exec.CommandContext with its own process group killed whole, plus WaitDelay. Checked by forcing a hang with a 20s deadline and a test that leaves a background child: it failed at 20.13s with "did not finish within 20s". Reverted afterwards.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: construct-scripts.test.sh:6-8 now says "the first bash on PATH (on macOS that is /bin/bash 3.2 unless a newer bash comes first)", which is accurate.
          round: 2
      findings:
        - id: BR-3
          severity: Minor
          title: TestConstructScripts adds its own 5m runnerDeadline instead of the shared testgate.HangCeiling
          detail: 'scripts/test/scripts_test.go:25. pkg/testgate/ceiling.go says HangCeiling exists so every package''s tests share one value. The runner takes about 13s, so the 2m ceiling still leaves about 9x headroom. ARCH-DRY: use testgate.HangCeiling, or explain why a whole-suite runner needs its own ceiling.'
          family: test-ceiling-single-source
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 52fd1fd07d6b9b764a1a9b4af9b36044738034cb
      blocked: false
---

# Gate ledger — ariadne#362 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T22:26:13-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `unbounded-subprocess-wait` TestConstructScripts runs the shell runner without a deadline
  scripts/test/scripts_test.go:22 uses exec.Command with no context, so a hanging construct test stalls the shard until the 30m go test timeout. Use exec.CommandContext with testgate.HangCeiling. This is the only subprocess call in the window.
- **BR-2** [Minor] `doc-claim-accuracy` Runner header says "macOS /bin/bash 3.2 locally", but it runs whatever bash is on PATH
  construct-scripts.test.sh:6 says the runner uses /bin/bash 3.2 on macOS. That only holds when /bin/bash comes first on PATH; with Homebrew bash first it runs bash 5. Reword to "the bash on PATH". This is the only such claim in the window; the atlas wording is accurate.

## Round 2 — 2026-10-13T22:28:34-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — exec.CommandContext with its own process group killed whole, plus WaitDelay. Checked by forcing a hang with a 20s deadline and a test that leaves a background child: it failed at 20.13s with "did not finish within 20s". Reverted afterwards.
- BR-2 — addressed — construct-scripts.test.sh:6-8 now says "the first bash on PATH (on macOS that is /bin/bash 3.2 unless a newer bash comes first)", which is accurate.

### Raised

- **BR-3** [Minor] `test-ceiling-single-source` TestConstructScripts adds its own 5m runnerDeadline instead of the shared testgate.HangCeiling
  scripts/test/scripts_test.go:25. pkg/testgate/ceiling.go says HangCeiling exists so every package's tests share one value. The runner takes about 13s, so the 2m ceiling still leaves about 9x headroom. ARCH-DRY: use testgate.HangCeiling, or explain why a whole-suite runner needs its own ceiling.

## Open findings

- **BR-3** [Minor] `test-ceiling-single-source` TestConstructScripts adds its own 5m runnerDeadline instead of the shared testgate.HangCeiling
