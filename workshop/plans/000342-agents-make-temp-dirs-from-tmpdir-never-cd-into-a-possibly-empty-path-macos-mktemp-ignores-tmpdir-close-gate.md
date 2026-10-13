---
gate: boundary-review
issue: 342
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T16:04:27-07:00"
      dispatched: "2026-10-13T16:03:39-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
          detail: '`TMPDIR=/x mktemp -d`, `{ mktemp -d; }`, `sudo mktemp -d` and `time mktemp -d` report no finding; commandPosition (pkg/shelllint/mktemp.go:89) should accept these prefixes.'
          family: lint-command-position-coverage
          round: 1
        - id: BR-2
          severity: Minor
          title: cd-into-substitution rule does not cover pushd "$(mktemp ...)"
          detail: pushd has the same empty-path failure mode as cd; extend the cdIntoMktemp regex at mktemp.go:26 to (cd|pushd).
          family: lint-command-position-coverage
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: e3ff505c5fb11818cf82b9d98fe20c055668b957
      blocked: false
    - "n": 2
      timestamp: "2026-10-13T16:06:28-07:00"
      dispatched: "2026-10-13T16:05:49-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: commandPosition (mktemp.go:105-109) accepts {, sudo, time, env, nohup, nice and NAME=value; tests at mktemp_test.go:21-24; go test passes.
          round: 2
        - id: BR-2
          disposition: addressed
          note: cdIntoMktemp regex now (cd|pushd) at mktemp.go:27; reason text + atlas/workflow/sandbox.md updated; test mktemp_test.go:25.
          round: 2
      findings:
        - id: BR-3
          severity: Minor
          title: commandPosition still misses reserved-word openers (if/elif/while/until) and option-taking prefixes (sudo -u x, time -p, env -i)
          detail: '3rd finding in this family: probed `if mktemp -d; then`, `sudo -u x mktemp -d`, `time -p mktemp -d` and `env -i mktemp -d`, and all report 0 findings. Fix the rule, not more instances: treat a word as a command word when the word before it is a list-opening reserved word, or when it is the first non-option, non-assignment word after a prefix command. That means skipping -opts and NAME=value words in commandPosition (mktemp.go:96) instead of checking only the last field. No tracked site is affected today.'
          family: lint-command-position-coverage
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 7fa5dce54b23874e8b2588dad7a9ff2cc2c30227
      blocked: false
---

# Gate ledger — ariadne#342 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T16:04:27-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `lint-command-position-coverage` MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
  `TMPDIR=/x mktemp -d`, `{ mktemp -d; }`, `sudo mktemp -d` and `time mktemp -d` report no finding; commandPosition (pkg/shelllint/mktemp.go:89) should accept these prefixes.
- **BR-2** [Minor] `lint-command-position-coverage` cd-into-substitution rule does not cover pushd "$(mktemp ...)"
  pushd has the same empty-path failure mode as cd; extend the cdIntoMktemp regex at mktemp.go:26 to (cd|pushd).

## Round 2 — 2026-10-13T16:06:28-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — commandPosition (mktemp.go:105-109) accepts {, sudo, time, env, nohup, nice and NAME=value; tests at mktemp_test.go:21-24; go test passes.
- BR-2 — addressed — cdIntoMktemp regex now (cd|pushd) at mktemp.go:27; reason text + atlas/workflow/sandbox.md updated; test mktemp_test.go:25.

### Raised

- **BR-3** [Minor] `lint-command-position-coverage` commandPosition still misses reserved-word openers (if/elif/while/until) and option-taking prefixes (sudo -u x, time -p, env -i)
  3rd finding in this family: probed `if mktemp -d; then`, `sudo -u x mktemp -d`, `time -p mktemp -d` and `env -i mktemp -d`, and all report 0 findings. Fix the rule, not more instances: treat a word as a command word when the word before it is a list-opening reserved word, or when it is the first non-option, non-assignment word after a prefix command. That means skipping -opts and NAME=value words in commandPosition (mktemp.go:96) instead of checking only the last field. No tracked site is affected today.

## Open findings

- **BR-3** [Minor] `lint-command-position-coverage` commandPosition still misses reserved-word openers (if/elif/while/until) and option-taking prefixes (sudo -u x, time -p, env -i)
