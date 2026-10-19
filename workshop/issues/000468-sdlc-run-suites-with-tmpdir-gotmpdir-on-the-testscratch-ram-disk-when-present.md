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
flow: {kind: quick, provenance: inferred, spec: "1c930b41", done: "cf887b1d"}
---

# sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present

## Problem

Test suites cause an fsync storm: a pair suite build minute showed 550,944 APFS container metadata writes in 64 s (median 8,600/s), 98% rewrites of the same ~30k blocks, i.e. thousands of forced transaction commits a second on the shared SSD (ops disk-writes analysis, 2026-10-19). Test temp files today land on the SSD: agents' TMPDIR on /Volumes/AgentScratch (an APFS volume in the same container) and suites such as ariadne's sharded run in the system temp folder (/var/folders/…). The operator approved a 16 GB RAM disk at /Volumes/TestScratch (ops teams/devinfra/tools/testscratch, a login item). 56 leftover go-build folders (~2 GB) had also accumulated in the scratch temp folder.

## Spec

When /Volumes/TestScratch is mounted and writable, `sdlc local-check run`, the suites it runs and targeted test runs it launches set TMPDIR and GOTMPDIR to /Volumes/TestScratch/tmp/<run-id>/ for the child processes (removed after the run); otherwise today's behavior. The Go build cache (GOCACHE) is unchanged. Free space on the RAM disk is checked before starting; below a floor (e.g. 2 GB) it falls back to the normal temp folder with a visible note. Also: remove stale go-build* folders older than a day from the temp folders it owns, never one in use.

## Done when

- With the RAM disk mounted, a suite's test temp files live under /Volumes/TestScratch (fixture test plus a live check); without it, nothing changes.
- A live disk-write sample (proc rusage of the suite tree) shows the SSD bytes written per suite falling.
- (Revisions) The run's own checkout and review checkouts start on the RAM disk when it has room, and fall back to their old roots when it is absent, full or refuses the write. The stale sweeps cover both roots. AGENTS.base.md §10 prefers the disk, with a fallback. A scratched suite warns about attributable SSD temp folders and counts the rest as unattributed. `tmp-literal` flags every SSD temp root.

## Revisions

### 2026-10-19: operator additions (via the hub)
Reason: the operator asked to advertise the RAM disk and to measure what still writes to the SSD.

Delta to Spec:
1. **Constitution:** one line in `AGENTS.base.md` §10 (the shared constitution; AGENTS.md is generated from it), next to the temp-dir rule. It advertises `/Volumes/TestScratch` (a RAM disk) for short-lived file writes in tests and scratch work, falling back to `$TMPDIR` when it is absent.
2. **Scrutiny of paths that bypass it:**
   - **(a) After a recorded suite run on the RAM disk:** a warning, never a failure, listing the top-level entries that appeared during the run in the SSD temp folders the run doesn't own: the system temp folder (`getconf DARWIN_USER_TEMP_DIR`, `/var/folders/.../T`) and `/Volumes/AgentScratch/tmp`.
     - It reports the count and size, and each entry's name.
     - A Go `t.TempDir` folder carries its test's name (`TestX123…`), which attributes it to a test. The warning also says any process on the machine may have made an entry, so it measures, it doesn't convict.
     - It is skipped when the run fell back, since then the SSD is the run's own temp.
   - **(b) Lint:** the existing `pkg/golint` rule `tmp-literal` (ariadne#342) already flags a `/tmp` literal in `os`, `filepath.Join` and `ioutil` calls. Widen it to `/private/tmp`, `/var/folders/` and `/Volumes/AgentScratch` literals, citing #468. `t.TempDir` and `os.MkdirTemp("", …)` already follow TMPDIR. Shell scripts are covered by `pkg/shelllint`'s template-less mktemp rule.

### 2026-10-19: review checkouts too (operator, via the hub)
Reason: review checkouts run tests inside (TestPreparation... wrote under a checkout's `.sdlc-review/tmp`), and losing them costs only a rebuild.

Delta: sdlc's review checkouts go to `/Volumes/TestScratch/review-checkouts` when the disk is mounted with room; otherwise they stay at `~/Library/Caches/sdlc/review-checkouts`. The lock and identity rules are unchanged. A checkout's directory is fixed at Create, and Reset reuses it, so a checkout lives wholly on the root it started on and is never half-moved. The sweep runs over both roots. The mount probe moves to `cmd/sdlc/internal/scratchdisk`, shared by both callers (ARCH-DRY).

### 2026-10-19: the suite's own checkout too (operator, via the hub)
Reason: every recorded suite checks out the whole repo with `os.MkdirTemp("", "sdlc-local-*")` and builds in it on the SSD, which is why fseventsd kept writing.

Delta: `detachedWorktreeWithCleanup` makes the checkout under `/Volumes/TestScratch/local-checks` by the same scratchdisk rule, else in the system temp folder as before. `sweepStaleLocalWorktrees` looks in both places. A disk that is mounted but refuses the write (an agent sandbox) falls back. The review checkout also falls back now when the disk refuses the write: found when targeted tests ran in the sandbox.

### 2026-10-19: 2a's list is bounded (close round 1)
Reason: an attributed leak can number in the hundreds.

Delta: the warning lists every attributed name up to 20 per package, then "… N more". The counts and sizes always cover every entry. Sizes are measured only for attributed entries; unattributed ones are counted, not walked, since they may be another process's large tree.

Delta to Plan: steps for 1, 2a and 2b are added below. The diff will likely leave the quick-flow shell, so the close review upgrades to the full review.

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
- [x] Tests: `chooseScratch` table (unmounted, low free space, probe error, ok); the sweep (dead pid vs live pid, old vs fresh go-build, files and dirs outside the patterns untouched); a `runUnderSlot` fixture whose child records `$TMPDIR` and `$GOTMPDIR` (under `WF_TEST_SCRATCH`, the dir removed after; `off`, inherited unchanged).
- [x] Implement `testscratch.go` and the change to `runUnderSlot`.
- [x] Revision 2a: snapshot the two SSD temp folders before a scratched run and diff them after (`scratchLeaks`, pure over two listings plus sizes); print the warning. Tests: a fixture with injected folder roots.
- [x] Revision 2b: widen `tmp-literal` in `pkg/golint/rules.go`, with test cases in `rules_test.go`; fix any new hits in the tree.
- [x] Revision 1: the `AGENTS.base.md` §10 line.
- [x] Review checkouts on the RAM disk (`reviewbox.Roots`), with a test.
- [x] Docs: `sdlc local-check --help` gets a line; atlas `sdlc-binary` "Running the test suite".
- [x] Live: the RAM disk is now mounted (the operator, 10-19). Run a suite through the branch's sdlc (temp under `/Volumes/TestScratch/tmp/<run>`, removed after), and take the disk-write sample before and after.

## Log

### 2026-10-19
- Dispatched by the hub to fill the wait while #456 is held for the operator's design review. Claimed and started from ariadne:4.
- Implemented in 198978a5a. Targeted tests pass: scratchdisk, reviewbox, and the sdlc scratch, lint and help tests.
- **Disk-write sample (A/B, 10-19, unsandboxed):**
  - Setup: the same targeted workload, `go test -count=1 ./pkg/... ./cmd/sdlc/internal/...`, with TMPDIR and GOTMPDIR on the SSD (`/Volumes/AgentScratch/tmp`) or on the RAM disk. The counters are IOBlockStorageDriver write bytes for disk0 (the SSD) and disk5 (TestScratch). The disk0 counter is machine-wide, so other slots add noise.
  - SSD temp: 1,770 MB and 1,178 MB to disk0.
  - RAM temp: 1,029 MB and 1,042 MB to disk0, plus 440 MB and 471 MB to disk5.
  - Net: about 440 MB per run moved off the SSD, matching the bytes that landed on the RAM disk.
- **Live check, first recorded suite on af1e3287:**
  - The suite ran with its temp on the RAM disk. Its run folder was removed afterwards, and the leak report printed (0 attributed, 5 unattributed).
  - It failed `dev-aliases.test.sh` ("an added/deleted file makes it stale"). The RAM disk is HFS+, with one-second times, so a directory touched in the same second the fake binary was built isn't newer than it. Confirmed: the old test fails 2 checks on the RAM disk, and the fixed one passes 3 of 3 (2d081ff76).
- **Found, out of scope (follow-up proposals):**
  - (1) The suite's own temporary checkout (`sdlc-local-*`) went to the system temp folder on the SSD. Folded in by the third revision.
  - (2) The agent sandbox denies writes to `/Volumes/TestScratch`, so agents' own sandboxed test runs can't use it until the operator adds it to the sandbox allowlist (local-gates config, the operator's).
- Recorded suite green on fb3941da (sdlc-local/suite, 10m45s), with its checkout under /Volumes/TestScratch/local-checks and its temp under /Volumes/TestScratch/tmp, both removed after the run. The leak warning showed 0 attributed and 4 unattributed.
- **Close round 1 (FIX-THEN-SHIP):**
  - BR-1: the run-folder pattern now accepts hyphenated check names, with a test case.
    - Class sweep ("sweep pattern narrower than its writer"): the go-build prefix matches Go's `go-build<digits>`; the local-check glob and `MkdirTemp` share `localCheckTmpPrefix`; reviewbox's sweep is unchanged. None of these is narrower than its writer.
  - BR-2: the warning now gives size per package and in total, plus every attributed name, capped at 20 per package with "… N more".
  - Minor: the review fallback retries only on a permission or read-only error, and is worded neutrally.
