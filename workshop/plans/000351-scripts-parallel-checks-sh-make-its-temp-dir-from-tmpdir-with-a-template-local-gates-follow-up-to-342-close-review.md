# Boundary Review — ariadne#351 (whole-issue close)

| field | value |
|-------|-------|
| issue | 351 — scripts/parallel-checks.sh: make its temp dir from TMPDIR with a template (local-gates follow-up to #342) |
| repo | ariadne |
| issue file | workshop/issues/000351-scripts-parallel-checks-sh-make-its-temp-dir-from-tmpdir-with-a-template-local-gates-follow-up-to-342.md |
| boundary | whole-issue close |
| milestone | — |
| window | 79779c97190e505c34f9a72873573d56885e1595..a77603f1f6dfa5b2d8e2d1b5fe640416b34e7865 |
| command | sdlc close --issue 351 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T17:17:44-07:00 |
| verdict | SHIP |

## Review

Tests pass and the syntax check is clean. Next I'll confirm that putting back only the old script line makes the lint test fail, then run the poll loop under bash 3.2.

```verdict
verdict: SHIP
confidence: high
```

This change does what the Spec asks. `scripts/parallel-checks.sh:145` now makes its temp dir as `mktemp -d "${TMPDIR:-/tmp}/parallel-checks.XXXXXX"`, and if that fails it prints an error and returns 1 before the cleanup trap is set. The trap now expands `"${OUTDIR:?}"`. The exemption map in `pkg/shelllint/repo_test.go:15` is empty. I ran `go test ./pkg/shelllint/` at the head commit and it passes; `bash -n` on the script is clean. That test is the regression evidence for the mktemp change. At the base commit the test required the exempt script to still have a finding, so the old template-less line was being flagged. At the head commit there is no exemption and the file scans clean, so putting the old line back would turn the test red. I tried to revert it in the checkout to watch it fail, but Bash was denied in don't-ask mode partway through the review. The stat, name-status and full-patch inspections had already finished. That denial also stopped me running the bash 3.2 empty-array idiom directly. The folded-in fix (`${alive[@]+"${alive[@]}"}`) is the standard portable idiom, and the issue's Log records the failure on bash 3.2.57 and the fixed run. Nothing blocks SHIP.

1. **Strengths**
   - `scripts/parallel-checks.sh:145-149`: the failure path runs before `trap … EXIT` is set, so a failed `mktemp` never reaches `rm -rf` with an empty path. The `${OUTDIR:?}` check is a second guard on top of that.
   - `pkg/shelllint/repo_test.go:45-49`: the check that "an exemption must still have a finding" is what proves the fix here. Removing the exemption is enforced by the test, not just a cleanup someone might forget.
   - The folded-in line 108 fix has a short comment explaining why the idiom is needed, and it is in scope: same file, same operator verification.
   - The empty map is kept on purpose, and its comment says when it gets used, which matches the Plan.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **ARCH-PURPOSE (class sweep):** `construct/local/introspect/scripts/introspect-extract.sh:102` has the same bug pattern: `SEGMENTS=("${TRIMMED[@]}")` under `set -u` crashes on bash 3.2 when `--limit 0`. It is out of scope here and only noted.
   - `scripts/parallel-checks.sh:102-105`: if `MAX_PARALLEL_CHECKS=0`, the script loops forever, and on bash 3.2 it hits the same empty `"${pids[@]}"` crash. This was already there before this change.
   - **Test coverage:** the line 108 bash 3.2 fix has no automated regression test. There is no test harness for `parallel-checks.sh`, and the Log describes only a manual stub run.

5. **Test coverage notes:** the shelllint repo-scan test pins the mktemp change. The `set -u` empty-array fix depends on manual verification. A shelllint rule for `x=("${y[@]}")` under `set -u` would cover the whole class, including the introspect site.

6. **Architecture**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass. This is a shell glue script and the lint logic stays pure in `MktempFindings`.
   - **ARCH-PURPOSE:** pass for the issue. The sibling noted under Minor is outside its scope.
   - **ARCH-MOCK:** N/A. No new external calls.
   - **ARCH-CONSTRAINTS:** pass. Concurrency is unchanged.
   - **ARCH-SECURE:** pass. The template is under `TMPDIR`, and `mktemp` creates the dir with 0700 permissions.
   - **ARCH-ORDER:** pass. The trap is installed only after the temp dir exists, and background subshells reset the trap (line 52).
   - **ARCH-FUNERAL:** pass. The EXIT trap removes `OUTDIR`, and the Log confirms nothing was left behind.
   - **Docs:** no new user-facing surface, so no atlas or README update is needed.

7. **Plan revisions:** none. The Plan matches the code. The merge gate, the operator applying the `local-gates-verified` label, is correctly kept out of the Plan.

```findings
findings:
  - id: new
    severity: Minor
    family: bash32-empty-array-set-u
    title: |
      Sibling empty-array expansion under set -u at construct/local/introspect/scripts/introspect-extract.sh:102
    detail: |
      SEGMENTS=("${TRIMMED[@]}") dies on bash 3.2 when --limit 0 leaves TRIMMED empty; same class as the folded-in line 108 fix. A shelllint rule would cover the class.
  - id: new
    severity: Minor
    family: shell-script-regression-coverage
    title: |
      Folded-in bash 3.2 pids fix in parallel-checks.sh has no automated regression test
    detail: |
      Verified only by a manual stub run per the Log; also MAX_PARALLEL_CHECKS=0 still spins and hits empty "${pids[@]}" at line 105 (pre-existing).
```

{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_01LftBrtpYveJpoRQhVeB8eG","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"810d9d87-bfe8-439b-a90a-e33a5a635191","session_id":"fdf904fe-8653-42c3-ade5-79a944df8063"}
{"type":"system","subtype":"permission_denied","tool_name":"Bash","tool_use_id":"toolu_019kaCZFneYUiER5Em4N4axq","decision_reason_type":"mode","message":"Permission to use Bash has been denied because Claude Code is running in don't ask mode. IMPORTANT: You *may* attempt to accomplish this action using other tools that might naturally be used to accomplish this goal, e.g. using head instead of cat. But you *should not* attempt to work around this denial in malicious ways, e.g. do not use your ability to run tests to execute non-test actions. You should only try to work around this restriction in reasonable ways that do not attempt to bypass the intent behind this denial. If you believe this capability is essential to complete the user's request, STOP and explain to the user what you were trying to do and why you need this permission. Let the user decide how to proceed.","uuid":"ea49711f-ada8-49ad-8d9e-94ceaacd0fb9","session_id":"fdf904fe-8653-42c3-ade5-79a944df8063"}
