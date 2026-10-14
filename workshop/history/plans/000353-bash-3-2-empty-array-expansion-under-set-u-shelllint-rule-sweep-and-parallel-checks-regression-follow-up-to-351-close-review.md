# Boundary Review — ariadne#353 (whole-issue close)

| field | value |
|-------|-------|
| issue | 353 — bash 3.2 empty-array expansion under set -u: shelllint rule, sweep, and parallel-checks regression (follow-up to #351) |
| repo | ariadne |
| issue file | workshop/issues/000353-bash-3-2-empty-array-expansion-under-set-u-shelllint-rule-sweep-and-parallel-checks-regression-follow-up-to-351.md |
| boundary | whole-issue close |
| milestone | — |
| window | bad07d64b64dd8967bdb8496733932c0cb4b1d8e..23f0b486641a3857cf9aa375de6034e906a7276c |
| command | sdlc close --issue 353 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T19:21:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Lint goes red when the fix is reverted. Next I'm checking a few edge cases against /bin/bash 3.2.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**FIX-THEN-SHIP: the lint is correct and the sweep is complete; two small findings remain, and neither blocks.**

The issue delivers its purpose. `pkg/shelllint.ArrayFindings` catches the empty-array class under `set -u`, the repo-wide test applies it to every tracked bash script, and all 21 findings it reported are guarded. `MAX_PARALLEL_CHECKS` now refuses anything that isn't a positive integer, so the `<=0` case has a defined behavior.

I checked rather than trusted the commit messages:
- **Tests:** `go test ./pkg/shelllint/` passes.
- **Revert check:** restoring the base `scripts/parallel-checks.sh` turns the repo test red at lines 80, 100, 105 and 116. So the lint really does pin the #351 fix and this sweep.
- **Bash 3.2 behavior:** on `/bin/bash` 3.2.57 with an empty array under `set -u`, `"${a[@]:1}"`, `"${a[@]#x}"`, `"${!a[@]}"` and the guarded form all succeed. That matches the narrowed rule and the atlas wording.

The two findings:
- **Important:** the new `MAX_PARALLEL_CHECKS` refusal has no automated test. It was only checked by hand with a stub.
- **Plan drift:** the plan's checklist still describes a default-deny rule ("every expansion is a finding unless…"). The code flags only the measured-unsafe operators. The Log explains why, but the plan text was never revised.

1. **Strengths**
   - `pkg/shelllint/array.go:12-41`: the rule only runs when the script enables nounset, and the regex accepts every spelling: `-u`, `-eu`, `-euo`, `-o nounset`, and a trailing comment. Commented-out or quoted `set -u` doesn't count (pinned by `TestArrayFindingsNeedsNounset`).
   - The rule was narrowed by measuring each operator on bash 3.2, not by guessing. The atlas (`atlas/workflow/sandbox.md:40-43`) and `workshop/lessons.md` record the result.
   - `lexed.expanding()` together with the new `squoted` class (`mktemp.go:53-56,143-153`) separates single-quoted literals from double-quoted text, which the shell does expand. It keeps byte offsets intact, and `arguments()` still treats `squoted` as literal text, so the mktemp rule is unchanged.
   - The `MAX_PARALLEL_CHECKS` check (`scripts/parallel-checks.sh:99-103`) fails loudly instead of quietly clamping. Because `main` runs under `set -e`, the `return 1` ends the run as intended.
   - The table tests (`array_test.go`) cover what matters: guards inside quotes, guarded substrings, two hits on one line, hits inside `$(...)`, comments and single-quoted text.

2. **Critical findings:** none.

3. **Important findings**
   - `scripts/parallel-checks.sh:99-103`: the refusal of `MAX_PARALLEL_CHECKS` values `0` and `abc` has no regression test. If someone removes the check, nothing goes red and the spin-forever bug from the Problem section comes back silently. Fix sketch: add a `scripts/*.test.sh` case (alongside `scripts/test/*.test.sh`) that runs the script with `MAX_PARALLEL_CHECKS=0` against stubbed checks and expects exit 1 plus the error message. This file is a local-gates path, so the operator's label is still needed.

4. **Minor findings**
   - `repo_test.go:103-118`: `isShell` reimplements the shebang-field loop that the new `shebangNames` already provides. It could be `shebangNames(src, "sh","bash","zsh","dash","ksh")` (ARCH-DRY).
   - The plan checklist item 1 still says "default-deny… every expansion is a finding unless…". The code instead flags only the bare form and the `^`/`,` operators. Any operator it doesn't recognise counts as safe, which is fine on bash 3.2, where bash-4.4 operators like `@Q` are a syntax error anyway.
   - The rule doesn't see `-u` turned on from the shebang (e.g. `#!/bin/bash -u`) or from `bash -u script` on the command line. It misses findings in that case rather than over-flagging. No tracked script does this today.
   - `array.go:35`: `for range unguardedArrays(...)` throws the offsets away, so it could return a count. Alternatively, keep the offsets and report a column.
   - Heredoc bodies and multi-line quoted strings are lexed one line at a time. That can over-flag, which is the safe direction, and the exemption map covers it.

5. **Test coverage notes**
   - The lint is a pure function tested with no IO.
   - The repo-wide test is the regression guard for the sweep, and I confirmed it fails when the fix is reverted.
   - The one behavior change with no test is the `MAX_PARALLEL_CHECKS` refusal (above).
   - The changed `pre-merge-checks.sh` line was only run by hand (`nosuchcheck`). That's acceptable for a `:-` default.

6. **Architectural notes**
   - **ARCH-DRY:** flagged as Minor (the `isShell`/`shebangNames` duplication).
   - **ARCH-PURE:** pass. `ArrayFindings` is pure; the file walking lives in the test.
   - **ARCH-PURPOSE:** pass. The whole class was enumerated and swept, the symlinked `.tart` copy is handled, and nothing purposeful was deferred.
   - **ARCH-MOCK:** N/A. No new external dependency (the existing `git ls-files` in the repo test is unchanged).
   - **ARCH-CONSTRAINTS:** pass. The concurrency limit is now validated, which closes off the unbounded spin.
   - **ARCH-SECURE:** N/A. The only input is the env var, which is now parsed with a regex before use.
   - **ARCH-ORDER:** holds no new state between events. The `pids` polling loop's semantics are unchanged; only its empty case is guarded.
   - **ARCH-FUNERAL:** pass. Nothing durable is created, and the exemption map is self-pruning (a stale exemption fails the test).
   - **Future work:** if per-element index access `${a[0]}` (which also fails when empty, in every bash version) or shebang `-u` matter later, extend the nounset detection in the same pure function.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry. Suggested text: "ArrayFindings flags only the bare `${a[@]}`/`${a[*]}` and the bash-4 case operators (`^` `,`); substring, pattern-removal, replacement and default/alternate operators were measured safe on bash 3.2.57 and are not flagged. This supersedes the default-deny wording in checklist item 1."

```findings
findings:
  - id: new
    severity: Important
    family: behavior-change-without-regression-test
    title: |
      MAX_PARALLEL_CHECKS positive-integer refusal has no automated test
    detail: |
      scripts/parallel-checks.sh:99-103 refuses 0/non-numeric values (a Done-when item) but was verified only by a manual stub run; add a scripts test asserting exit 1 for MAX_PARALLEL_CHECKS=0 and =abc (local-gates label applies).
  - id: new
    severity: Minor
    family: plan-text-drifts-from-code
    title: |
      Plan checklist still describes default-deny; code flags only measured-unsafe operators
    detail: |
      Add a Revisions entry recording the narrowed operator set (bare, ^, ,) per the 3.2.57 measurement noted in the Log.
  - id: new
    severity: Minor
    family: duplicated-helper-logic
    title: |
      isShell reimplements the shebang-field loop now provided by shebangNames
    detail: |
      pkg/shelllint/repo_test.go:103-118 could call shebangNames(src, "sh","bash","zsh","dash","ksh") (ARCH-DRY).
  - id: new
    severity: Minor
    family: nounset-detection-coverage
    title: |
      Nounset detection ignores -u on the shebang line or bash -u invocation
    detail: |
      This misses findings rather than over-flagging; no tracked script currently enables -u that way.
```

---

## Re-review — 2026-10-13T19:26:01-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 353 — bash 3.2 empty-array expansion under set -u: shelllint rule, sweep, and parallel-checks regression (follow-up to #351) |
| repo | ariadne |
| issue file | workshop/issues/000353-bash-3-2-empty-array-expansion-under-set-u-shelllint-rule-sweep-and-parallel-checks-regression-follow-up-to-351.md |
| boundary | whole-issue close |
| milestone | — |
| window | bad07d64b64dd8967bdb8496733932c0cb4b1d8e..ae1b29a6c0fbc177443e36687e706f04f1a2d8a0 |
| command | sdlc close --issue 353 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T19:26:01-07:00 |
| verdict | SHIP |

## Review

Tests pass; now checking how the CI merge check invokes the new test, then a quick mutation check.

```verdict
verdict: SHIP
confidence: high
```

All four findings from the last round are fixed, and I checked each fix myself. `go test ./pkg/shelllint/` passes at head ae1b29a6, including the repository test that now runs `ArrayFindings` over every tracked bash script. The new `scripts/test/parallel-checks.test.sh` passes all 6 checks. I then deleted the positive-integer check from `scripts/parallel-checks.sh` and ran the test again: 3 of 6 failed. I restored the file afterwards and the tree is clean. CI runs the test through `scripts/merge-checks.d/20-ci-workflow-tests.sh`, which calls `bash` on every `scripts/test/*.test.sh`, so the file's 100644 mode doesn't matter. I found nothing new that blocks shipping.

Done-when: current. The rule fails on an unguarded `"${a[@]}"` under `set -u` and passes the guarded form (array_test.go). The repository test runs with an empty exemption map and passes. `MAX_PARALLEL_CHECKS<=0` is defined: the script refuses it and exits 1. The Revisions entry only narrowed which operators get flagged, and none of the three Done-when items depends on that.

1. **Strengths**
   - The rule flags only what was actually measured to fail. The Log records the bash 3.2.57 measurements, and the cases measured as safe are tested as safe (`pkg/shelllint/array_test.go:29-32`).
   - The lexer change is small: a new `squoted` class and an `expanding()` view (`pkg/shelllint/mktemp.go:53-56,143-153`). Double-quoted expansions are scanned and single-quoted literals are not, and the mktemp rule keeps working because `arguments` treats both quote classes the same.
   - The parallel-checks test runs the real script against a stub, with a perl-alarm time limit. A regression that spins forever (as `MAX_PARALLEL_CHECKS=0` did) fails the test instead of hanging the suite. The `MAX_PARALLEL_CHECKS=1` case covers the #351 pids rebuild.
   - The sweep is complete: the repository test shows zero findings with no exemptions.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:**
   - The header comment in `scripts/parallel-checks.sh` still says only "(default: 3)". It could also say the value must be a positive integer. Not raised as a finding.
   - Lines are lexed one at a time, so a heredoc body with a quoted delimiter would be flagged by mistake. That errs toward too many findings, and no tracked script triggers it today. Not raised.

5. **Test coverage:**
   - The table tests cover each operator class and every nounset spelling, including shebang flags (`#!/bin/bash -eu`, `env -S bash -u`).
   - The repository test enforces the sweep.
   - The script test failed when I removed the fix, so it catches that regression.

6. **Architecture:**
   - ARCH-DRY passes: `shebangNames` is shared by `isBash` and `isShell`.
   - ARCH-PURE passes: `ArrayFindings` is a pure function of the source string, and only the repository test touches git and the filesystem.
   - ARCH-PURPOSE passes: the whole class was swept, not just the instances #351 named, and the lint enforces it.
   - ARCH-MOCK passes: the script test uses a local scratch git repo and a stub checks runner, with no external services.
   - ARCH-CONSTRAINTS is not applicable: this is a test-time lint, and the parallelism limit is now validated.
   - ARCH-SECURE passes: the only input is a bounded environment variable, rejected by a strict regex.
   - ARCH-ORDER is not applicable: the code holds no state between events beyond the existing pids loop, which is unchanged apart from the guard.
   - ARCH-FUNERAL passes: the test's temp directory is removed by its `trap` on exit.

7. **Plan revisions:** none. The Revisions entry already records the narrowed operator set.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      scripts/test/parallel-checks.test.sh asserts exit 1 for 0/abc/-2; removing the validation turns 3/6 cases red (verified), and 20-ci-workflow-tests runs it.
  - id: BR-2
    disposition: addressed
    note: |
      The Revisions entry dated 2026-10-13 records the narrowed operator set (bare, ^, ,), matching array.go unguardedArrays.
  - id: BR-3
    disposition: addressed
    note: |
      isShell now calls shebangNames(src, "sh","bash","zsh","dash","ksh") at repo_test.go:95-105.
  - id: BR-4
    disposition: addressed
    note: |
      shebangNounset regex in array.go plus TestArrayFindingsNeedsNounset cases for "#!/bin/bash -eu" and "env -S bash -u"; a caller's bash -u is noted as statically invisible.
```
