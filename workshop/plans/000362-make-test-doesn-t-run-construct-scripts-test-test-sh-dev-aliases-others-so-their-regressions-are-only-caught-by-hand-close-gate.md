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

## Open findings

- **BR-1** [Minor] `unbounded-subprocess-wait` TestConstructScripts runs the shell runner without a deadline
- **BR-2** [Minor] `doc-claim-accuracy` Runner header says "macOS /bin/bash 3.2 locally", but it runs whatever bash is on PATH
