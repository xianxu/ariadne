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
    - "n": 3
      timestamp: "2026-10-13T16:09:19-07:00"
      dispatched: "2026-10-13T16:08:23-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: Rule inverted in commandWords/lookup/codeOnly (mktemp.go:72-171); if/until/sudo -u/time -p/which cases at mktemp_test.go:25-30,44 go red against the prior mktemp.go (verified by swap), green at head.
          round: 3
      findings:
        - id: BR-4
          severity: Minor
          title: codeOnly masks backquote substitution inside double quotes, so d="`mktemp -d`" is no longer flagged (regression vs prior head)
          detail: '4th finding in this family. Rule: every code-running construct inside double quotes ($(...) and `...`) stays unmasked. In codeOnly''s ''"'' branch (mktemp.go:141-155), push a backquote context like the $( case and pop it on the closing backquote. Add a "`mktemp -d`" test case. No tracked site is affected.'
          family: lint-command-position-coverage
          round: 3
        - id: BR-5
          severity: Minor
          title: Three independent quote scanners in mktemp.go (stripComment, codeOnly, arguments)
          detail: 'ARCH-DRY: stripComment could be built on codeOnly''s mask (a # at the start of a word in masked code), so quote handling is fixed in one place.'
          family: quote-scanner-duplication
          round: 3
      recipe: fix-review
      isolation: checkout
      reviewed: ad470891aa1ce396f5c1289a8877066cd9b01759
      blocked: false
    - "n": 4
      timestamp: "2026-10-13T16:12:13-07:00"
      dispatched: "2026-10-13T16:11:15-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: lex pushes a backquote context inside double quotes (mktemp.go:94-95), popped by the code branch (:110-111); the "backquotes in double quotes" test fails with the prior mktemp.go swapped in and passes at head; nested cases probed correct.
          round: 4
        - id: BR-5
          disposition: addressed
          note: stripComment, codeOnly and the old arguments scanner are replaced by one lex (mktemp.go:68-125); masked() and lexed.arguments() read its byte classes; comment cut happens only in code context; tests at mktemp_test.go:49-50 pin it.
          round: 4
      recipe: fix-review
      isolation: checkout
      reviewed: 7b167cedfd0e447270a97af75cf90db5f372ce71
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

## Round 3 — 2026-10-13T16:09:19-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — Rule inverted in commandWords/lookup/codeOnly (mktemp.go:72-171); if/until/sudo -u/time -p/which cases at mktemp_test.go:25-30,44 go red against the prior mktemp.go (verified by swap), green at head.

### Raised

- **BR-4** [Minor] `lint-command-position-coverage` codeOnly masks backquote substitution inside double quotes, so d="`mktemp -d`" is no longer flagged (regression vs prior head)
  4th finding in this family. Rule: every code-running construct inside double quotes ($(...) and `...`) stays unmasked. In codeOnly's '"' branch (mktemp.go:141-155), push a backquote context like the $( case and pop it on the closing backquote. Add a "`mktemp -d`" test case. No tracked site is affected.
- **BR-5** [Minor] `quote-scanner-duplication` Three independent quote scanners in mktemp.go (stripComment, codeOnly, arguments)
  ARCH-DRY: stripComment could be built on codeOnly's mask (a # at the start of a word in masked code), so quote handling is fixed in one place.

## Round 4 — 2026-10-13T16:12:13-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — lex pushes a backquote context inside double quotes (mktemp.go:94-95), popped by the code branch (:110-111); the "backquotes in double quotes" test fails with the prior mktemp.go swapped in and passes at head; nested cases probed correct.
- BR-5 — addressed — stripComment, codeOnly and the old arguments scanner are replaced by one lex (mktemp.go:68-125); masked() and lexed.arguments() read its byte classes; comment cut happens only in code context; tests at mktemp_test.go:49-50 pin it.

## Open findings

(none — every finding has been disposed)
