---
id: 000450
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# sandbox: deny agents read access to ~/.config/rclone (Drive backup token)

## Problem

pair's history backup will upload to Google Drive through rclone (operator decision 2026-10-18). rclone keeps its OAuth token in ~/.config/rclone/rclone.conf, which the agents' Claude Code sandbox can read today. The sandbox read-deny list in .claude/settings.ariadne.json covers ~/.aws, ~/.config/gcloud and others, but not ~/.config/rclone.
