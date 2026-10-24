---
id: 000598
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
---

# Affected planning: declared non-Go path ownership and audited reader inputs

## Problem

Pair#853 cannot safely narrow reader-certificate changes with the current affected declaration. Pair has ~560 certificates under tests/affected-readers/<Go test path>.json. Its broad tests/** full rule forces a suite for each update. Removing the rule leaves certificates unclassified. Group.go only triggers non-Go groups; it never adds Go packages. Certificates currently contain version, disposition, sha256, with no declared read paths.
