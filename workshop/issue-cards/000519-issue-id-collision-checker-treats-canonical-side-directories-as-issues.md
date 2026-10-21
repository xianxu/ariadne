---
id: 000519
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# Issue ID collision checker treats canonical side directories as issues

## Problem

pair#731 PR503 CI run39542144320 failed its duplicate-issue-ID gate because a canonical sibling side-document directory was counted as a second issue: workshop/issues/000731-a-pane-that-fell-back-to-a-native-shell-still-classifies-as-a-pair-terminal.side alongside the corresponding .md. The directory contained only metadata-capture.json, not another issue. Current lint-ids does not reproduce after the capture is moved inline.
