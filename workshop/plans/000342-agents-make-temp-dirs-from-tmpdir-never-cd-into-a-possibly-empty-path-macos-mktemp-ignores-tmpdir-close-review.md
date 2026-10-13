# Boundary Review — ariadne#342 (whole-issue close)

| field | value |
|-------|-------|
| issue | 342 — agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR) |
| repo | ariadne |
| issue file | workshop/issues/000342-agents-make-temp-dirs-from-tmpdir-never-cd-into-a-possibly-empty-path-macos-mktemp-ignores-tmpdir.md |
| boundary | whole-issue close |
| milestone | — |
| window | ba954921dca066bc9b3f5006fb287421d1b19614..e3ff505c5fb11818cf82b9d98fe20c055668b957 |
| command | sdlc close --issue 342 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T16:04:27-07:00 |
| verdict | SHIP |

## Review

I've finished checking the diff and probing the scanner for edge cases. Writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

The change does what the issue asked for. The temp-dir rule is now in `AGENTS.base.md` §10 and reaches peers through weave; `AGENTS.md` and `CLAUDE.md` are generated and untracked, so the base file is the only source. All eight surveyed sites were fixed. A new pure scanner, `pkg/shelllint.MktempFindings`, is unit-tested on every failing form named in the plan. The repo test applies it to every tracked shell script and Makefile, and fails if an exemption no longer has a finding. I checked that the test catches regressions: restoring the old `WORK=$(mktemp -d)` in `.colima/test/colima.test.sh` turns it red with the expected message, and `go test ./pkg/shelllint/` passes at head. What remains are minor gaps where the scanner misses some forms, none of which affect the swept tree.

1. **Strengths**
   - `pkg/shelllint/mktemp.go` contains only text checks with no IO. The IO (`git ls-files`, reading files) lives only in `repo_test.go`, so the logic is cleanly separated from the IO (ARCH-PURE).
   - `repo_test.go:47-51`: an exemption that no longer has a finding fails the test. This gives the `#351` exemption a built-in end (ARCH-FUNERAL), so it can't outlive its issue.
   - `repo_test.go:57`: the `scanned == 0` check stops a broken file filter from passing silently.
   - The table in `mktemp_test.go` tests look-alikes that must not be flagged (comments, words inside strings, `"$mktemp"`, identifiers), not just the failing forms.
   - `clone-data-deps.test.sh:13-14` now separates the assignment from the `cd` and checks it with `${ROOT:?}`, which is the rule's own pattern.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The scanner misses `mktemp` after a prefix word or env assignment. `{ mktemp -d; }`, `sudo mktemp -d`, `time mktemp -d` and `TMPDIR=/x mktemp -d` all report 0 findings (probed). The last one still ignores `TMPDIR` on macOS. Fix: in `commandPosition` (`mktemp.go:89`), accept `{`, `sudo`, `time` and a trailing `NAME=value` word.
   - The `cd` rule doesn't cover `pushd "$(mktemp …)"`, which has the same empty-path failure mode. Fix: extend the regex at `mktemp.go:26` to `(cd|pushd)`.
   - `cd $(mktemp -d)` reports two findings (template and cd). That's harmless, and arguably the right result.

5. **Test coverage notes**
   - The unit table covers each form the plan lists.
   - The repo test enforces both the sweep and the exemption's expiry. Reverting one swept site turns it red (checked).
   - Lines continued with a backslash, and heredoc bodies, aren't modelled. The likely error is a visible false positive, not a silent miss.

6. **Architectural notes**
   - ARCH-DRY: pass. The new package doesn't duplicate an existing shell lint.
   - ARCH-PURE: pass.
   - ARCH-PURPOSE: pass. Every surveyed consumer was swept, the one remaining site is exempted by name and tracked as #351, and the check enforces the rest.
   - ARCH-MOCK: pass. The only external call is `git ls-files`, a read-only call in a test against the repo.
   - ARCH-CONSTRAINTS: pass. The test scans the tracked tree in about 0.2s.
   - ARCH-SECURE: pass. Nothing new touches untrusted input or secrets. The sweep reduces writes to `$HOME`.
   - ARCH-ORDER: pass. The change holds no state between events; the scanner works line by line.
   - ARCH-FUNERAL: pass. Swept scripts keep their `trap rm -rf`, and the exemption expires on its own.
   - If more shell checks land in `pkg/shelllint`, consider a shared tokenizer, since `stripComment` and `arguments` would otherwise grow per rule.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: lint-command-position-coverage
    title: |
      MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
    detail: |
      `TMPDIR=/x mktemp -d`, `{ mktemp -d; }`, `sudo mktemp -d` and `time mktemp -d` report no finding; commandPosition (pkg/shelllint/mktemp.go:89) should accept these prefixes.
  - id: new
    severity: Minor
    family: lint-command-position-coverage
    title: |
      cd-into-substitution rule does not cover pushd "$(mktemp ...)"
    detail: |
      pushd has the same empty-path failure mode as cd; extend the cdIntoMktemp regex at mktemp.go:26 to (cd|pushd).
```
