---
id: 000468
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '193ef1fb233d745154c2d504c6ea91ad38232ec8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T09:32:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present

## Problem

Test suites cause an fsync storm: a pair suite build minute showed 550,944 APFS container metadata writes in 64 s (median 8,600/s), 98% rewrites of the same ~30k blocks, i.e. thousands of forced transaction commits a second on the shared SSD (ops disk-writes analysis, 2026-10-19). Test temp files today land on the SSD: agents' TMPDIR on /Volumes/AgentScratch (an APFS volume in the same container) and suites such as ariadne's sharded run in the system temp folder (/var/folders/…). The operator approved a 16 GB RAM disk at /Volumes/TestScratch (ops teams/devinfra/tools/testscratch, a login item). 56 leftover go-build folders (~2 GB) had also accumulated in the scratch temp folder.

## Spec

When /Volumes/TestScratch is mounted and writable, `sdlc local-check run`, the suites it runs and targeted test runs it launches set TMPDIR and GOTMPDIR to /Volumes/TestScratch/tmp/<run-id>/ for the child processes (removed after the run); otherwise today's behavior. The Go build cache (GOCACHE) is unchanged. Free space on the RAM disk is checked before starting; below a floor (e.g. 2 GB) it falls back to the normal temp folder with a visible note. Also: remove stale go-build* folders older than a day from the temp folders it owns, never one in use.

## Done when

- With the RAM disk mounted, a suite's test temp files live under /Volumes/TestScratch (fixture test plus a live check); without it, nothing changes.
- A live disk-write sample (proc rusage of the suite tree) shows the SSD bytes written per suite falling.

## Plan

There is one seam: `runUnderSlot` (`cmd/sdlc/localcheck_run.go`) starts every suite and every known-flake rerun (`local-check run`, close's recorded suite and merge's integration suite), so the scratch choice lives there. The design goes in a new `cmd/sdlc/testscratch.go`:

- **Probe:** `scratchProbe{Mounted, FreeBytes, Err}`.
  - `Mounted` is true when `/Volumes/TestScratch` exists on a different device from `/Volumes`. An unmounted path is then never written to the SSD by mistake.
  - `FreeBytes` comes from statfs.
  - `WF_TEST_SCRATCH=<dir>` replaces the root and skips the mount check (fixtures, and an operator override); `WF_TEST_SCRATCH=off` disables the feature.
- **Pure decision:** `chooseScratch(probe, floor)` returns use or fall back, with a note. The floor is 2 GiB. Unmounted means a silent fallback, today's behavior. Below the floor, or a probe error, means fallback with a visible `[!]` note.
- **Per run:** `MkdirTemp(<root>/tmp, "<check>-<pid>-")`. The child gets that dir as both `TMPDIR` and `GOTMPDIR`, appended last so it wins. The dir is removed after the process group is stopped. If the mkdir fails (for example under a sandbox), fall back with a note. `GOCACHE` is unchanged.
- **Stale sweep, before each run:**
  - in `<root>/tmp`, remove a run dir whose pid is dead, or that is older than 24 h;
  - remove `go-build*` older than 24 h in `<root>/tmp` and in the fallback temp folder (`os.TempDir()`, where unscratched suites build).

  Age keeps a live dir safe: a running go build keeps writing into its folder, and suites time out well under a day.
- **ARCH notes:**
  - FUNERAL: a run dir ends with its run; a crashed one ends at the next run's sweep.
  - CONSTRAINTS: the RAM disk is 16 GB, the floor is 2 GiB, and at most 2 suites run machine-wide.
  - SECURE: no credentials are involved, and an env override names only a directory.

Steps:
- [ ] Tests: `chooseScratch` table (unmounted, low free space, probe error, ok); the sweep (dead pid vs live pid, old vs fresh go-build, files and dirs outside the patterns untouched); a `runUnderSlot` fixture whose child records `$TMPDIR` and `$GOTMPDIR` (under `WF_TEST_SCRATCH`, the dir removed after; `off`, inherited unchanged).
- [ ] Implement `testscratch.go` and the change to `runUnderSlot`.
- [ ] Docs: `sdlc local-check --help` gets a line; atlas `sdlc-binary` "Running the test suite".
- [ ] Live: this machine has no RAM disk mounted yet (`df` finds none), so the live check and the disk-write sample are the hub's or operator's once it is mounted. I'll send the exact commands.

## Log

### 2026-10-19
- Dispatched by the hub to fill the wait while #456 is held for the operator's design review. Claimed and started from ariadne:4.
