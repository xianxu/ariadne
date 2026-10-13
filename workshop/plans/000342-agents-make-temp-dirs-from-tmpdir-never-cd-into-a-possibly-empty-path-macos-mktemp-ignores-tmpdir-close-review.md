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

---

## Re-review — 2026-10-13T16:06:27-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 342 — agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR) |
| repo | ariadne |
| issue file | workshop/issues/000342-agents-make-temp-dirs-from-tmpdir-never-cd-into-a-possibly-empty-path-macos-mktemp-ignores-tmpdir.md |
| boundary | whole-issue close |
| milestone | — |
| window | 16b7779a5eb8240e679146e61466c05a8a0b6a5c..7fa5dce54b23874e8b2588dad7a9ff2cc2c30227 |
| command | sdlc close --issue 342 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T16:06:27-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both prior findings are fixed in this interdiff, and nothing in it is a Critical or Important defect. In `pkg/shelllint/mktemp.go:105-109`, `commandPosition` now accepts `{`, `sudo`, `time`, `env`, `nohup`, `nice` and a trailing `NAME=value` word. The `cdIntoMktemp` regex at `mktemp.go:27` now matches `(cd|pushd)`. The reason text and `atlas/workflow/sandbox.md` were updated to match. There is a unit case for each newly covered form at `mktemp_test.go:21-25`, plus a `${mktemp}` look-alike guard. `go test ./pkg/shelllint/` passes at head. The scope is unchanged: no Spec, Revisions or Done-when edits, and no behaviour beyond what the findings asked for. One new Minor finding, in the same family, is described below.

1. **Strengths**
   - `mktemp_test.go:21-25`: every form the prior finding named has its own case, so taking out any part of the fix breaks a specific test.
   - `mktemp_test.go:37`: the "braced variable" case guards against the most likely false positive that accepting `{` could cause.
   - The fix covers more than the cited prefixes: `env`, `nohup` and `nice` were added too.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Third finding in the `lint-command-position-coverage` family.** Earlier rounds fixed individual prefixes. I probed the fixed scanner, and these still report 0 findings:
     - `if mktemp -d; then …` (the list has no `if`, `elif`, `while` or `until`)
     - `sudo -u x mktemp -d`, `time -p mktemp -d`, `env -i mktemp -d` (a prefix command followed by options)

     Don't add these words one by one. The rule that covers all of them: a word is in command position when the word before it is a shell reserved word that opens a command list (`if elif while until then do else ! { (`), or when it is the first non-option, non-assignment word after a prefix command. To implement that, `commandPosition` should skip `-opt` words (and, for `sudo -u`, their arguments) and `NAME=value` words back to a prefix command or reserved word, rather than checking only the last field.

     There is also a small false-positive edge: `echo "see X=1 mktemp -d"` is flagged, because the assignment check doesn't know it is inside a quoted string. That fails loudly, not silently.

5. **Test coverage notes**
   - The new cases pin each prefix and `pushd`.
   - Nothing covers reserved-word openers like `if` and `while`, or prefix commands with options. Those are the forms in the Minor finding.

6. **Architectural notes**
   - ARCH-DRY: pass. `assignment` is a single regex, and the prefix list lives in one switch.
   - ARCH-PURE: pass. The change only touches pure text functions, tested with no IO.
   - ARCH-PURPOSE: flagged, Minor. The fix went past the named instances, but the class (command-position recognition) still isn't covered by a stated rule. See the Minor finding.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      commandPosition (mktemp.go:105-109) accepts {, sudo, time, env, nohup, nice and NAME=value; tests at mktemp_test.go:21-24; go test passes.
  - id: BR-2
    disposition: addressed
    note: |
      cdIntoMktemp regex now (cd|pushd) at mktemp.go:27; reason text + atlas/workflow/sandbox.md updated; test mktemp_test.go:25.
findings:
  - id: new
    severity: Minor
    family: lint-command-position-coverage
    title: |
      commandPosition still misses reserved-word openers (if/elif/while/until) and option-taking prefixes (sudo -u x, time -p, env -i)
    detail: |
      3rd finding in this family: probed `if mktemp -d; then`, `sudo -u x mktemp -d`, `time -p mktemp -d` and `env -i mktemp -d`, and all report 0 findings. Fix the rule, not more instances: treat a word as a command word when the word before it is a list-opening reserved word, or when it is the first non-option, non-assignment word after a prefix command. That means skipping -opts and NAME=value words in commandPosition (mktemp.go:96) instead of checking only the last field. No tracked site is affected today.
```

---

## Re-review — 2026-10-13T16:09:18-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 342 — agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR) |
| repo | ariadne |
| issue file | workshop/issues/000342-agents-make-temp-dirs-from-tmpdir-never-cd-into-a-possibly-empty-path-macos-mktemp-ignores-tmpdir.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1e4f9e9c1cf90948354852b7ed7bb5275dbf628f..ad470891aa1ce396f5c1289a8877066cd9b01759 |
| command | sdlc close --issue 342 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T16:09:18-07:00 |
| verdict | SHIP |

## Review

Probes done; checking whether the backtick-in-double-quotes miss is a regression against the prior head.

```verdict
verdict: SHIP
confidence: high
```

BR-3, the only open prior finding, is fixed. The fix changes the rule itself, as the finding asked. `commandWords` (`pkg/shelllint/mktemp.go:72-95`) now flags every unquoted occurrence of `mktemp` that stands as its own word. It skips four kinds of non-call: quoted literal text, part of a larger word, a `$`/`${}` expansion, and a lookup that only names the command (`command -v`, `which`, `type`, `hash`). That covers `if`, `elif`, `while`, `until`, `sudo -u x`, `time -p` and `env -i` without listing each one.

I put the prior head's `mktemp.go` back in place and ran the tests. The new cases `if opener`, `until opener`, `sudo with an option value`, `time -p` and `which lookup` fail, so the regression tests really depend on the fix. At the pinned head, `go test ./pkg/shelllint/` passes, including the repo-wide test over tracked scripts.

The interdiff makes no Spec, Revisions or Done-when changes. The only additions beyond code and tests are the gate/review artifacts and one entry in `lessons.md`.

The new masking does add one false negative, but it doesn't block shipping. `codeOnly` blanks backquotes inside double quotes, so `` d="`mktemp -d`" `` is no longer flagged. The prior head caught it (I checked: 1 finding before, 0 now). No tracked site uses this form today, so it is Minor.

1. **Strengths**
   - The rule is inverted as asked (`mktemp.go:72-77`). Calls are flagged by default and only the provable non-calls are listed. The comment states the rule plainly.
   - `codeOnly` (`mktemp.go:118-171`) keeps `$( … )` inside double quotes as code and handles nesting. These were all flagged correctly: `"$(echo "$(mktemp -d)")"`, `$(echo ')' ; mktemp -d)` and an escaped `\"` before `$(…)`.
   - There are tests for the exceptions as well as the new catches: `command -v`, `which`, `type`, single-quoted `trap` text and `${mktemp}` (`mktemp_test.go:43-47`). Without these, the broader rule could add false positives unnoticed.
   - The `lessons.md` entry records the general lesson: invert an enumerated-context rule once a review finds a second missed context.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **This is the 4th finding in family `lint-command-position-coverage`.** `codeOnly` treats `$(` as code inside double quotes but blanks a backquote, so `` d="`mktemp -d`" `` gives 0 findings. The prior head flagged it. The general rule: anything that runs code inside double quotes must stay unmasked, and there are exactly two such forms, `$(…)` and `` `…` ``. In `codeOnly`'s `'"'` branch (`mktemp.go:141-155`), push a backquote context the same way `$(` pushes one, and pop it on the closing backquote. Add a test case for `` "`mktemp -d`" ``.
   - ARCH-DRY: the file now has three separate quote scanners: `stripComment` (`:52`), `codeOnly` (`:118`) and `arguments` (`:175`). `stripComment` could be built on `codeOnly`'s mask: a `#` at the start of a word in masked code is a comment. Merging them means a quoting fix happens in one place.

5. **Test coverage notes:** Each form BR-3 named has its own case, and those cases fail on the prior code. The only gap in this window is a backquote substitution inside double quotes. A path-qualified call like `/usr/bin/mktemp -d` is also missed, but the prior code missed it too, so the interdiff didn't introduce it.

6. **Architectural notes**
   - ARCH-DRY: flagged, Minor. Three quote scanners; see above.
   - ARCH-PURE: pass. Every change is a pure string function, tested with no IO.
   - ARCH-PURPOSE: pass. The class was fixed by changing the rule rather than adding more instances. The one remaining gap is a masking omission, not another enumerated context.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Rule inverted in commandWords/lookup/codeOnly (mktemp.go:72-171); if/until/sudo -u/time -p/which cases at mktemp_test.go:25-30,44 go red against the prior mktemp.go (verified by swap), green at head.
findings:
  - id: new
    severity: Minor
    family: lint-command-position-coverage
    title: |
      codeOnly masks backquote substitution inside double quotes, so d="`mktemp -d`" is no longer flagged (regression vs prior head)
    detail: |
      4th finding in this family. Rule: every code-running construct inside double quotes ($(...) and `...`) stays unmasked. In codeOnly's '"' branch (mktemp.go:141-155), push a backquote context like the $( case and pop it on the closing backquote. Add a "`mktemp -d`" test case. No tracked site is affected.
  - id: new
    severity: Minor
    family: quote-scanner-duplication
    title: |
      Three independent quote scanners in mktemp.go (stripComment, codeOnly, arguments)
    detail: |
      ARCH-DRY: stripComment could be built on codeOnly's mask (a # at the start of a word in masked code), so quote handling is fixed in one place.
```

---

## Re-review — 2026-10-13T16:12:13-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 342 — agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR) |
| repo | ariadne |
| issue file | workshop/issues/000342-agents-make-temp-dirs-from-tmpdir-never-cd-into-a-possibly-empty-path-macos-mktemp-ignores-tmpdir.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00472964b870f0958ff0fd7328c31b1e0abcd316..7b167cedfd0e447270a97af75cf90db5f372ce71 |
| command | sdlc close --issue 342 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T16:12:13-07:00 |
| verdict | SHIP |

## Review

I've finished the edge-case probes and am writing up the verdict.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and the change introduces no new defect. The fix does more than patch the one case BR-4 named: it replaces the three separate quote scanners in `mktemp.go` (`stripComment`, `codeOnly`, `arguments`) with a single lexer. `lex` (`pkg/shelllint/mktemp.go:68-125`) tags every byte as code, quoted text or a quote character. It also cuts the trailing comment, and both kinds of command substitution count as code even inside double quotes. Masking (`masked`), argument splitting (`arguments`) and comment handling all read from it. `go test ./pkg/shelllint/` passes at the pinned head. With the prior head's `mktemp.go` swapped in, the new `backquotes in double quotes` case fails, so the regression test depends on the fix. The window changes no Spec, Revisions or Done-when text; it only adds the gate and review records.

1. **Strengths**
   - **One lexer for quoting (`mktemp.go:62-125`).** Quoting rules now live in one place. The `default` branch handles top-level code, `$(` substitutions and backquote substitutions together. The context stack treats `$(…)` and `` `…` `` the same way (`kind` is `'('` or `` '`' ``).
   - **Comments are cut in code context only (`:105`).** So `#` inside quotes and `${x#…}` are no longer mistaken for comments. The new tests (`mktemp_test.go:49-50`) cover an apostrophe inside a comment and a `#` inside quotes.
   - **Probes I ran all came out right:**
     - Flagged (no template):
       - `` "$(echo "`mktemp -d`")" ``
       - nested escaped backquotes `` "`echo \`mktemp -d\``" ``
       - `$((1+2)); mktemp -d`
       - `d=$(mktemp -d) # mktemp -d` (once only)
       - Makefile `$$(mktemp -d)`
     - Not flagged:
       - a templated `` "`mktemp -d "${TMPDIR:-/tmp}/x.XXXXXX"`" ``
       - `"$(echo a) # mktemp -d"`
   - **Signatures:** `commandWords`, `lookup` and `word` keep their logic and now take a `lexed` value instead of a raw string.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - Outside this window, not raised as a finding: `cdIntoMktemp` only matches the `$(` form, so `` cd "`mktemp -d`" `` gets the missing-template finding but not the cd finding. That regex was already there before this window.

5. **Test coverage notes:** BR-4 is covered for backquotes both inside double quotes and unquoted (`mktemp_test.go:31-32`); the quoted case fails on the prior code. BR-5 is a refactor with no behaviour change, and the existing table plus the two new comment/quote cases pin it. The repo-wide test over tracked scripts still passes.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The three quote scanners became one `lex`; consumers only read its byte classes.
   - **ARCH-PURE: pass.** Everything is a pure string function, tested with no IO.
   - **ARCH-PURPOSE: pass.** BR-4 was fixed for the whole class ("everything that runs code inside double quotes stays code"), not just the one site: both `$(` and backquotes push a code context. BR-5 was fixed by merging the scanners, not by touching just one of them.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      lex pushes a backquote context inside double quotes (mktemp.go:94-95), popped by the code branch (:110-111); the "backquotes in double quotes" test fails with the prior mktemp.go swapped in and passes at head; nested cases probed correct.
  - id: BR-5
    disposition: addressed
    note: |
      stripComment, codeOnly and the old arguments scanner are replaced by one lex (mktemp.go:68-125); masked() and lexed.arguments() read its byte classes; comment cut happens only in code context; tests at mktemp_test.go:49-50 pin it.
```
