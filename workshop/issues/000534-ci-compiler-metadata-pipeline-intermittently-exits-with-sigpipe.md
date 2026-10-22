---
id: 000534
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4935375c2eacf639e4c33686caca8defe6352c45' # card fields mirrored from issue-cards; edit via sdlc
---

# CI compiler metadata pipeline intermittently exits with SIGPIPE

## Problem

While verifying echo#1 with the unchanged upstream portable-ci.test.sh harness,
the actual workflow Build first pinned compiler step intermittently returned
141 even though its metadata had the selected revision and modified=false.
The template pipes printf into grep -Eq under set -euo pipefail. Early grep
success can close the pipe while printf is writing, producing SIGPIPE.

## Spec

Remove early-consumer-close races from the pinned compiler workflow template's
metadata checks. Use a here-string or a grep that consumes the whole input
(without -q, redirect output to /dev/null). Sweep the analogous metadata
pipelines in the template and its test harness. Preserve exact revision and
modified=false validation and fail-closed behavior. Consumers should obtain
the change through weave; do not patch each generated workflow manually.

## Done when

- Exact revision and modified checks reliably succeed for valid metadata and
  reject mismatches under pipefail, with repeated actual-block tests.
- The metadata pipeline class is enumerated and fixed in source templates.
- portable-ci.test.sh still exercises identity failures and old PATH exclusion.

## Plan

- [ ]

## Log

2026-10-21: echo#1 observed actual workflow exit 141 locally after an earlier
pass with the same workflow. Minimal exact-pipeline reproduction on macOS Bash:
set -o pipefail; printf '%s\n' "$METADATA" | grep -Eq "build[[:space:]]+vcs.revision=$PIN$"
METADATA was three lines: binary: go1.26.3, two-space build vcs.revision equal
to 63d6bfd86509d97649e10d7a5dd1e9a06db17d21, and two-space build vcs.modified=false.
1000 fresh Bash subprocesses returned {0: 999, 141: 1}. This is a scheduling
race rather than an OS-specific contract. Echo TL authorized upstream filing
and logged reruns; no template modification was made in echo.
