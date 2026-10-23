# Boundary Review — ariadne#573 (whole-issue close)

| field | value |
|-------|-------|
| issue | 573 — dev-aliases refuses cmd names that shadow shell builtins, reserved words or system commands |
| repo | ariadne |
| issue file | workshop/issues/000573-dev-aliases-refuses-cmd-names-that-shadow-shell-builtins-reserved-words-or.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e..d8789eaa01cfde72a5863da477af87aeee73dec7 |
| command | sdlc close --issue 573 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T09:10:06-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the diff for #573; now checking workspace canonicalization, CI platform, and running the new test.

Nearly done — checking if I can confirm the Linux path issue in a container.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The generator change does what the Spec asks. `shadow_reason` is pure and tested from a table. All per-name lookups are gathered in one bash subshell plus one `zsh -f` call. A refused name produces one warning, gets no function and is left out of `--list`. The emitted text for every other name is unchanged. The new test passes at head on macOS, sourcing the output in both bash and zsh. One thing should be fixed before shipping. The new test hard-codes `/bin/ls` as the resolved system path. CI runs on `ubuntu-latest`, where `/bin` is a symlink to `/usr/bin` and the test's PATH puts `/usr/bin` first. There the generator would report `/usr/bin/ls` for both `ls` and `lsalias`, and the test would fail. I could not confirm this in a container because the Docker daemon was not running. The conclusion comes from the PATH order and how `cd -P` resolves the symlink.

**1. Strengths**
- `construct/dev-aliases.sh:401` `shadow_reason` is a pure function. The test loads it alone with the sed-eval pattern (`dev-aliases-shadow.test.sh:15`) and runs a table that covers all five reasons plus the allowed cases (ARCH-PURE pass).
- The lookups cost the same no matter how many names there are: one zsh fork and one bash subshell. `readlink` only runs for PATH matches outside the workspace and `$HOME`. The test counts these calls with stub binaries at 8 and 16 extra names (`:85-95`).
- `system_command` checks the path twice, once as found and once after resolving symlinks. So `~/.local/bin/sdlc`, and a relative PATH entry that resolves into the workspace, both stay allowed.
- The loop changed from `printf | while` to a heredoc with fds 4 and 5. The bash and zsh lists line up with `$final`, and a missing last line falls back to empty. The zsh line is also checked against `"$bin: "` before use, so a mismatch can't apply the wrong kind.
- The atlas (`atlas/workflow/base-layer.md:150`) and the script header both describe the new rule.

**2. Critical:** none.

**3. Important**
- `construct/scripts/test/dev-aliases-shadow.test.sh:24,45,64,65`: the hard-coded `/bin/ls` will not match on merged-usr Linux (the CI runner).
  - Line 64: `type -P ls` with `PATH=…:/usr/bin:/bin` returns `/usr/bin/ls`.
  - Line 65: `lsalias → /bin/ls`, then `cd -P /bin` resolves to `/usr/bin/ls`.
  - Line 24 (unit table) is pure and still passes, but it is the same assumption.
  - Line 45 (the symlink target) works on both systems.
  - Fix: compute the expected path once, e.g. `sys_ls="$(cd -P "$(dirname "$(PATH=/usr/bin:/bin type -P ls)")" && pwd)/ls"`, and use it in both greps.

**4. Minor**
- ARCH-DRY: `system_command`'s symlink loop (`dev-aliases.sh:422-425`) nearly repeats the single-hop dereference at `:151-160`. They could share one `resolve_link` helper.
- Plan says "wire the test into the Makefile". No Makefile change was needed, because `construct-scripts.test.sh` finds `*.test.sh` files on its own. The plan wording is slightly off.
- In the atlas, "a symlink into the workspace such as `~/.local/bin/sdlc` is exempt" is accurate. But the `$HOME` rule already exempts that link before anything is resolved, so the example doesn't show the symlink rule. `$fixture/sys/linked` in the test does.

**5. Test coverage**
- Both modes of every Done-when clause are covered: refused and allowed names, bash and zsh sourcing, and `--list`.
- When zsh is missing, the test skips the zsh checks and drops `print` and `repeat` from the refused set. That is correct, so the zsh-only rules go untested on CI if its image lacks zsh.
- The shell-start timing was measured by hand and recorded in the Log; nothing checks it automatically, which is fine for this budget.

**6. Architecture**
- ARCH-DRY: pass, apart from the Minor finding above.
- ARCH-PURE: pass. The IO is gathered once at the edge, and the rule itself is pure.
- ARCH-PURPOSE: pass. The fix covers every repo through the generator, not just echo's `cmd/echo`.
- ARCH-EVOLVE: not applicable. The generator stores nothing; it writes text to stdout.

**7. Plan revisions**
- Optional: a "## Revisions" note saying the test is picked up by `construct-scripts.test.sh` discovery, not wired in through the Makefile.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: test-hardcodes-host-filesystem-layout
    title: |
      dev-aliases-shadow.test.sh expects /bin/ls, but merged-usr Linux (ubuntu-latest CI) resolves it to /usr/bin/ls
    detail: |
      The test's PATH lists /usr/bin before /bin. On Ubuntu 24.04, `type -P ls` returns /usr/bin/ls, and lsalias (a link to /bin/ls) resolves through `cd -P /bin` to /usr/bin/ls. The greps at lines 64 and 65 for "the system command /bin/ls;" would then fail. Every instance: lines 64 and 65 fail; line 24 (pure unit table) and line 45 (the link target) are the same assumption but harmless. Fix: compute the canonical system ls path at test start and use it in both greps. Not run on Linux: the Docker daemon was not running, so the conclusion comes from the PATH order and how cd -P resolves /bin.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-symlink-resolution
    title: |
      ARCH-DRY: the symlink loop in system_command nearly repeats the script-path dereference at dev-aliases.sh:151-160
    detail: |
      Both use readlink plus path_parts to follow links by hand, which bash 3.2 needs. Both could share one resolve_link helper.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"08e0496f-5021-4ac4-b3e5-6bd1f8fb0930","repository":"github.com/xianxu/ariadne","issue_id":"000573","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5ac2508f0f921db2","reviewed_head":"d8789eaa01cfde72a5863da477af87aeee73dec7","slot":"ariadne:2","requested_at":"2026-10-23T16:08:23.27629Z"},"phase":"finished","pid":71041,"started_at":"2026-10-23T16:08:23.371697Z","finished_at":"2026-10-23T16:10:06.387925Z","session_id":"b7949763-de2d-418b-9257-022d7495d1f9","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T09:13:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 573 — dev-aliases refuses cmd names that shadow shell builtins, reserved words or system commands |
| repo | ariadne |
| issue file | workshop/issues/000573-dev-aliases-refuses-cmd-names-that-shadow-shell-builtins-reserved-words-or.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e..5903d2db7eb2916c3f5a3ae130cd99e31d54da05 |
| command | sdlc close --issue 573 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T09:13:23-07:00 |
| verdict | SHIP |

## Review

The tests pass on macOS. Next I'm checking how the suite discovers the new test, then I'll look at the main edge cases.

```verdict
verdict: SHIP
confidence: medium
```

The change does what the issue asks. `construct/dev-aliases.sh` now refuses any `cmd/X` name that would shadow a bash or zsh builtin or reserved word, or a system command outside the workspace and `$HOME`. Each refused name gets exactly one warning, and the name is left out of both the function definitions and `--list`. All the facts are gathered in one batch: one bash subshell and one `zsh -f` call per run. I ran the tests here on macOS: `dev-aliases-shadow.test.sh` passes under both `bash` and `/bin/bash` 3.2, and so do `dev-aliases.test.sh`, `-path` and `-slot`. Both prior findings are fixed. One new Minor remains. It is the 2nd finding in the `test-hardcodes-host-filesystem-layout` family and does not block shipping.

Done-when: current. The diff builds the refusal of builtin, reserved and system names in bash and zsh with one warning each, ordinary names keep their functions, and the timing was measured and logged. That is everything both Done-when bullets require. The Revisions entry about Makefile wiring is accurate: `scripts/test/construct-scripts.test.sh:24` finds `construct/scripts/test/*.test.sh` on its own.

**1. Strengths**
- The rule is a pure function. `shadow_reason` (`construct/dev-aliases.sh:405`) is tested on its own through a unit table loaded with the sed-eval pattern, with no IO.
- The facts are gathered in one batch, so the cost stays flat as names are added. Reading them through fds 4 and 5 in `$final` order keeps them aligned with the names. The `"$bin: "` prefix check on each zsh line means a misaligned line yields an empty kind, not the wrong one.
- The test counts zsh and readlink calls at N and 2N names. That catches a fork per name coming back, not just wrong output.
- The test sources the output in real bash and zsh and checks what `type -t` and `whence -w` report, so it checks the result rather than the generator's own text.
- Each refusal warning names the repo, the command and the reason, and says how to run the binary anyway.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **This is the 2nd finding in family `test-hardcodes-host-filesystem-layout`.**
  - **The bug:** `system_command` runs for every name (`construct/dev-aliases.sh:472`), including names bash or zsh already refused. The test then asserts exactly 2 readlink calls, and it builds `lsalias_want` with `cd -P` only, without following a final symlink. Both assume that no host binary the fixture reaches is a symlink.
  - **Where it fails:** on Debian or Ubuntu with zsh and `mailcap` installed, `/usr/bin/print` is a symlink to `run-mailcap`. That makes 3 readlink calls instead of 2. On busybox systems such as Alpine, `/bin/ls` is a symlink to `busybox`, which breaks both the readlink count and `lsalias_want`.
  - **The rule:** every test expectation that depends on the host's filesystem must either be derived from the host using the generator's own resolution (`follow_links`, then `cd -P`), or the fixture must not reach host binaries at all.
  - **The fix:** in production, call `system_command` only when the bash and zsh kinds gave no reason. That also saves readlink forks. In the test, derive the expected hop count and `lsalias_want` with that same resolution, instead of the literal `2` and a `cd -P`-only lookup.
- The plan says `type -ap`, but the code uses `type -P`, which returns only the first match. The atlas correctly says "first PATH match". This is a wording mismatch in the plan only.

**5. Test coverage notes**
- BR-1: the test now derives `ls_want` and `lsalias_want` from the host (test file, lines 55–58). On merged-usr Linux, the generator resolves `/usr/bin/ls` the same way: `follow_links` does nothing, then `cd -P /usr/bin`. `lsalias` goes to `/bin/ls`, and `cd -P /bin` gives `/usr/bin/ls`, which matches. I did not run this on Linux because Docker is not available here; this conclusion comes from tracing the code.
- BR-2: this was a refactor. The test that covers it is `dev-aliases-path.test.sh`, which passes under bash 3.2 and still exercises the symlinked script path.

**6. Architectural notes**
- ARCH-DRY: pass. `follow_links` is now the single symlink walk, used at `:165` and `:473`.
- ARCH-PURE: pass. The rule is a pure function, and the IO happens once in a thin gathering layer.
- ARCH-PURPOSE: pass. Builtins, reserved words and system commands are all covered, in both shells.
- ARCH-MOCK: pass. The test fakes `$HOME` and `PATH`; the zsh and readlink stubs only count calls and then run the real tools.
- ARCH-CONSTRAINTS: pass. The budget was measured and logged, and the work per name stays flat. There is a small extra saving available, described in the Minor finding.
- ARCH-SECURE: pass. Names come from directory names under `cmd/` and are passed as argv to zsh, with no string building.
- ARCH-ORDER: pass. The script holds no state between events. Moving the loop out of a pipe and into the main shell only changes the `prepared` cache's scope, and that is harmless.
- ARCH-FUNERAL: pass. The script leaves nothing on disk except the test's temp directory, which a trap removes.
- ARCH-EVOLVE: pass. No persisted format is involved.

**7. Plan revision recommendations**
- A one-line Revisions entry: "Rule 3 uses `type -P` (first PATH match), not `type -ap`."

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The test now derives ls_want and lsalias_want with cd -P, the same as system_command does; traced for merged-usr. Not run on Linux (no Docker).
  - id: BR-2
    disposition: addressed
    note: |
      follow_links (dev-aliases.sh:138) is now shared by the script-path dereference (:165) and system_command (:473); the path, slot and base tests pass under bash 3.2.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: test-hardcodes-host-filesystem-layout
    title: |
      Readlink count of exactly 2 and lsalias_want both assume no host system binary reached by the fixture is a symlink
    detail: |
      2nd finding in this family. Rule: every test expectation that depends on the host filesystem must be derived from the host with the generator's own resolution (follow_links then cd -P), or the fixture must not reach host binaries at all. system_command runs even for names already refused as builtins (dev-aliases.sh:472). So on Debian with zsh and mailcap (/usr/bin/print is a symlink to run-mailcap) the count is 3, and on busybox (where /bin/ls is a symlink to busybox) both the count and lsalias_want are wrong. Fix: call system_command only when the bash and zsh kinds gave no reason, and derive the expected hop count and lsalias_want with that same resolution.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9fc1f045-0878-4023-8d75-38beb0af90ae","repository":"github.com/xianxu/ariadne","issue_id":"000573","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/695dc5e30e4c26ff","reviewed_head":"5903d2db7eb2916c3f5a3ae130cd99e31d54da05","slot":"ariadne:2","requested_at":"2026-10-23T16:11:31.897541Z"},"phase":"finished","pid":1849,"started_at":"2026-10-23T16:11:32.000518Z","finished_at":"2026-10-23T16:13:22.543303Z","session_id":"44a4cfe1-6bf1-426b-8f0b-02cd2460bfee","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T09:16:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 573 — dev-aliases refuses cmd names that shadow shell builtins, reserved words or system commands |
| repo | ariadne |
| issue file | workshop/issues/000573-dev-aliases-refuses-cmd-names-that-shadow-shell-builtins-reserved-words-or.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e..c0cb4a5090b5b149295e422a4b558d88717a2d51 |
| command | sdlc close --issue 573 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T09:16:31-07:00 |
| verdict | SHIP |

## Review

Now revert the BR-3 production fix to confirm the test goes red.

```verdict
verdict: SHIP
confidence: high
```

The only open finding, BR-3, is fixed, and I confirmed the fix by reverting it. In production, `system_command` now runs only when neither shell already refuses the name (`construct/dev-aliases.sh`, main loop: `shadow_reason … ""` first, then `if [ -z "$reason" ]`). In the test, every expectation that depends on the host comes from the generator's own `follow_links` plus `cd -P`: `ls_want`, `lsalias_want`, and the expected readlink count from `want_hops`. That count leaves out the names a shell refuses, the same way the generator does. The new `sys/echo → /bin/echo` symlink checks that the PATH lookup is skipped. Results in this checkout:
- `dev-aliases-shadow.test.sh` passes, and so does `dev-aliases.test.sh`.
- I made the generator resolve PATH even for names a shell already refuses. The test then failed with `FAIL: 8 names: 3 readlink calls, want 2`, so the regression test catches the bug.

Nothing new blocks the ship.

1. **Strengths**
   - `shadow_reason` is a pure rule. The test loads it from the production script and checks it against a table, including the case where a bash function name like `warn` is not refused.
   - All the lookups happen in batches: one bash subshell for `type -t` and `type -P`, and one `zsh -f` call. They are fed to the loop on fds 4 and 5, so the cost does not grow with a subshell per name. The counting stubs check this at N and 2N names.
   - `follow_links` is now shared by the script-path dereference and `system_command`, and it stops after 40 hops so a link cycle ends.
   - The `family` rule for this kind of bug is now written in `workshop/lessons.md`, so it covers the whole class and not just one instance.
   - The atlas entry `atlas/workflow/base-layer.md:150` describes the new refusal and its exemptions.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. One edge I noticed but am not raising: on a host with no `/bin/ls` at all (for example NixOS), the `lsalias` fixture link would dangle. The test's PATH contains no other match for it, so the generator would then allow `lsalias` instead of refusing it, and the test would fail. CI runs on ubuntu, where this can't happen, so it is below threshold.

5. **Test coverage notes:** The tests cover:
   - the unit table, sourcing the output in both bash and zsh, one warning per refused name, and `--list` leaving refused names out;
   - the `$HOME` exemption and the link into the workspace being allowed;
   - one zsh call per run and readlink calls proportional to symlink hops.

   The ordering fix has a test that fails without it.

6. **Architectural notes:**
   - **ARCH-DRY:** pass (shared `follow_links` and `path_parts`).
   - **ARCH-PURE:** pass (pure `shadow_reason`; the IO is gathered once).
   - **ARCH-PURPOSE:** pass. Every Done-when item is delivered and the start time was measured.
   - **ARCH-MOCK:** pass. The real bash and zsh are used, with counting stubs for zsh and readlink, which is right for a test that is about how the shells behave.
   - **ARCH-CONSTRAINTS:** pass. The added time is within run-to-run noise, the cost is checked structurally, and the cycle cap holds.
   - **ARCH-SECURE:** pass. Names go through argv arrays and `--`.
   - **ARCH-ORDER:** pass. The generator holds no state between events.
   - **ARCH-FUNERAL:** pass. The test's temporary directory is removed by an EXIT trap, and the generator writes only to stdout.
   - **ARCH-EVOLVE:** N/A, since no persisted format changes.

7. **Plan revision recommendations:** none. The existing Revisions entries already cover the Makefile change that wasn't needed and the switch from `type -ap` to `type -P`.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      system_command now runs only when neither shell refuses the name, and the test derives ls_want, lsalias_want and want_hops via the generator's own follow_links plus cd -P. Reverting the skip turns the test red with "8 names: 3 readlink calls, want 2".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bbf6c2e9-452a-45c6-8610-3319a00f731a","repository":"github.com/xianxu/ariadne","issue_id":"000573","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/1fcc860100633bf9","reviewed_head":"c0cb4a5090b5b149295e422a4b558d88717a2d51","slot":"ariadne:2","requested_at":"2026-10-23T16:15:30.703983Z"},"phase":"finished","pid":50972,"started_at":"2026-10-23T16:15:30.803547Z","finished_at":"2026-10-23T16:16:30.467416Z","session_id":"8ba5a2c1-718d-4bee-a808-93fde86ef083","outcome":"returned"}] -->
