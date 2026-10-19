---
id: 000468
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '99551f19ea4b464d4817174739e78322b781b89c' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-19
