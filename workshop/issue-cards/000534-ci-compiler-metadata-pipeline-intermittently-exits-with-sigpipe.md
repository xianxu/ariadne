---
id: 000534
status: open
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
---

# CI compiler metadata pipeline intermittently exits with SIGPIPE

## Problem

While verifying echo#1 with the unchanged upstream portable-ci.test.sh harness,
the actual workflow Build first pinned compiler step intermittently returned
141 even though its metadata had the selected revision and modified=false.
The template pipes printf into grep -Eq under set -euo pipefail. Early grep
success can close the pipe while printf is writing, producing SIGPIPE.
