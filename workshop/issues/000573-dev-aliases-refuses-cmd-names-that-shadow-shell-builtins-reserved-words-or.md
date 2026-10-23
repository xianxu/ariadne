---
id: 000573
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ea65c3d744aa47a87f4f93cea87b7776c41283e1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T08:01:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.27
---

# dev-aliases refuses cmd names that shadow shell builtins, reserved words or system commands

## Problem

## Spec

**dev-aliases shadows shell builtins and system commands (found by echo:0, verified by ops-echo:0, 10-23).**
`construct/dev-aliases.sh` emits a shell function for every owned `cmd/X` with no check on the name. echo has
`cmd/echo`, so every operator shell that sources dev-aliases (from the operator's zshrc) got an `echo()` function
overriding the builtin. The prompt calls `echo`; the function rebuilt `~/workspace/echo/bin/echo` when missing and
ran echo's room CLI, which broke the operator's shell, and deleting the binary didn't stick. echo#19 renames that CLI
to `echoctl`, but nothing stops the next repo doing the same.

Fix: dev-aliases refuses, with one warning naming the repo and command, any `cmd/X` whose name is a shell builtin or
reserved word (in the sourcing shell: `builtin`/`whence -w` in zsh, `type -t` in bash), or that already resolves to a
system command outside the workspace (at least the POSIX utilities). No function is emitted for a refused name; the
binary stays reachable by its path.

## Done when

- Sourcing dev-aliases with a repo that has `cmd/echo` (and `cmd/test`, `cmd/ls`) emits no function for them and
  prints one warning each; ordinary cmd names still get their functions (test in bash and zsh).
- The check runs at source time without measurable shell-start delay (measured).

## Plan

Quick flow: about 60 added lines in `construct/dev-aliases.sh`, no milestones.

**Where the check runs.** In the generator, not in the sourcing shell. The generator already runs at shell start, as a bash child of the operator's shell. A name refused there emits no function at all, warns once from one place, and stays out of `--list`. Checking inside the sourcing shell would need two shell dialects in the emitted text and a symlink resolution that bash can't do without forking.

**Rule.** Refuse `cmd/X` when any of these holds:
1. X is a builtin or keyword in bash (`type -t`, in the generator's own process).
2. X is a builtin or reserved word in zsh. One batched `zsh -f -c 'whence -w -- "$@"'` call covers every candidate name, about 6ms. If zsh is absent, this check is skipped.
3. X is on PATH (`type -ap`) as a file whose resolved path lies outside both the workspace and `$HOME`. That covers /bin, /usr/bin, /sbin, /opt/homebrew and similar. Symlinks are resolved first, so `~/.local/bin/sdlc` (a symlink into `workspace/ariadne/bin`) stays allowed.

Checking both shells means a name that is a builtin in either one is refused in both. That is stricter than checking only the sourcing shell, and it's deliberate: such a name is a bad CLI name anywhere.

**Why rule 3 exempts `$HOME`.** A user-installed copy such as a stale `go install` in `~/go/bin/sdlc` must not cost the fleet its dev functions. The cost of this choice: a name that matches a tool installed under `$HOME` (`~/.cargo/bin/rg`, say) is not refused.

**Warning.** One line per refused name: `dev-aliases: not defining X (cmd/X in REPO): it would shadow the zsh builtin X; run REPO/bin/X by its path or rename the command`. The reason is one of: bash builtin, bash keyword, zsh builtin, zsh reserved word, or system command PATH.

**Shape (ARCH-PURE).** The pure function `shadow_reason NAME BASHKIND ZSHKIND PATHS` returns the reason or nothing. The IO is gathered once per run: the zsh batch call, and `type -ap` plus symlink resolution per name. Symlinks are resolved only for PATH matches, so in practice there are a handful of `readlink` forks.

**ARCH notes.**
- **ARCH-CONSTRAINTS (shell start):** the baseline generator takes about 0.21s. The budget for the change is at most 10ms of median added time, measured over 20 runs, main against the branch, and recorded in the Log. Growth: one zsh fork per run (constant), one PATH scan per name, and readlink forks only on PATH symlink hops.
- **ARCH-ORDER, ARCH-FUNERAL, ARCH-EVOLVE, ARCH-MOCK, ARCH-SECURE:** N/A. Holds no state between events and creates nothing durable; the generator emits text to stdout. Its inputs are the caller's own PATH and shells.

**Tests**, in a new `construct/scripts/test/dev-aliases-shadow.test.sh`:
- [x] Use a fixture workspace with `cmd/echo`, `cmd/test`, `cmd/ls`, `cmd/print` (zsh-only builtin) and `cmd/oktool`. Source the output in bash and in zsh. Assert that echo, test and print still resolve to the builtin, `ls` to `/bin/ls`, and `oktool` to a function. Stderr must hold exactly one warning per refused name; `--list` must omit the refused names.
- [x] Symlink case: put a PATH directory under a fake HOME with a symlink into the workspace, and a direct symlink to `/bin/ls`. The first is allowed, the second refused.
- [x] Unit table for `shadow_reason`, loaded with the existing sed-eval pattern.
- [x] Counted test at N and 2N owners: zsh is invoked exactly once, and `readlink` only once per PATH symlink hop.
- [x] Wire the test into the Makefile next to `dev-aliases.test.sh`, update the script's header and `atlas/workflow/base-layer.md`, and measure the start time.

## Log

### 2026-10-23
- 2026-10-23: closed — Round 2 head 5903d2db7: BR-1 fixed. The test derives the expected system paths with the generator's own cd -P resolution, so merged-usr Linux expects /usr/bin/ls and macOS /bin/ls. The minors are fixed: shared follow_links with a cycle cap, and the atlas wording. dev-aliases-shadow, dev-aliases, -path and -slot tests and pkg/shelllint pass, with the recorded suite queued on this head. Earlier evidence still holds: suite on d8789eaa 0 failed; over the real ~/workspace only cmd/echo is refused, the other 26 functions unchanged; generator CPU main 641ms (sd 38) against branch 624ms (sd 41). The TL took the real-fleet run as the live check, with no stage. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 106 added lines in code files (limit 100)
- Plan approved by the hub (ops:0), with the $HOME exemption on rule 3.
- Built: `shadow_reason` (pure; sets `reason`) and `system_command`/`outside_user_trees` in `construct/dev-aliases.sh`. One bash subshell gathers every name's `type -t` and `type -P`, and one `zsh -f` call gathers every `whence -w`. Symlinks are resolved only when a PATH match lies outside the workspace and $HOME, so `~/.local/bin/sdlc` costs no `readlink`. Refused names warn once and are dropped from the definitions and from `--list`.
- The first cut walked PATH in shell and called readlink for every `~/.local/bin` link: about 31ms per run, over budget. Batching the lookups and resolving only outside matches cut the block to about 13ms wall under load average 20. By CPU: zsh 3.9ms, the bash batch subshell about 2.6ms, plus the loop.
- End-to-end, generator CPU over `--workspace ~/workspace` (27 names, 30 runs each, interleaved): main 641ms (sd 38), branch 624ms (sd 41). Wall medians 1108ms against 1107ms. The added work is inside the run-to-run noise.
- Real fleet: `--workspace ~/workspace` now refuses only echo's `cmd/echo` ("it would shadow the bash builtin echo"). The other 26 functions are unchanged.
- Tests: `dev-aliases-shadow.test.sh` passes. Disabling the refusal turns it red ("echo: 0 warnings"). `dev-aliases`, `-path`, `-slot` and `portable-makefile` all pass under /bin/bash 3.2, as do `go test ./scripts/test -run TestConstructScripts` and `./pkg/shelllint`. An empty workspace exits 0 under `set -u`.
- Close review round 1 returned FIX-THEN-SHIP, with one Important and minors, all fixed in this round. BR-1: the test hard-coded /bin/ls, but merged-usr Linux CI finds /usr/bin/ls first and resolves /bin to /usr/bin. The test now computes both expected paths with the same physical (cd -P) resolution the generator uses. Minor ARCH-DRY: the script-path dereference and system_command now share `follow_links`, which follows every hop and is capped at 40 so a link cycle ends. Minor: the atlas example now separates the $HOME exemption from the followed system-dir link.
- Close round 2 returned SHIP with one advisory Minor. It was the second finding in the test-hardcodes-host-filesystem-layout family, so I fixed the rule. Production: a name a shell already refuses no longer resolves its PATH match, so it spends no readlink. Test: every host-dependent expectation (the `ls` and `lsalias` paths and the readlink count) now comes from the generator's own `follow_links` plus `cd -P`, so Debian's `print` link and busybox's `/bin/ls` link count correctly. A symlinked `echo` in the fixture's system dir pins the skip: forcing the lookup gives "3 readlink calls, want 2".

## Revisions

- 2026-10-23: The plan's step to wire the test into the Makefile was not needed. `construct-scripts.test.sh` discovers `construct/scripts/test/*.test.sh` on its own, so the new test runs in the suite without a Makefile change.
- 2026-10-23: Rule 3 uses `type -P`, the first PATH match, not `type -ap` as the plan said. The first match is what the name would run.
