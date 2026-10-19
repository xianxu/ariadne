---
id: 000450
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'dc99fc81f1a1eeb249915fb511f9a77354ea05bd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T15:38:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "1f6d4f79", done: "c54c9586"}
---

# sandbox: deny agents read access to ~/.config/rclone (Drive backup token)

## Problem

pair's history backup will upload to Google Drive through rclone (operator decision 2026-10-18). rclone keeps its OAuth token in ~/.config/rclone/rclone.conf, which the agents' Claude Code sandbox can read today. The sandbox read-deny list in .claude/settings.ariadne.json covers ~/.aws, ~/.config/gcloud and others, but not ~/.config/rclone.

## Spec

Add "~/.config/rclone" to sandbox.filesystem.denyRead in .claude/settings.ariadne.json, right after "~/.config/gcloud". One line; nothing else changes. This is a local-gates path: the operator verifies the diff and applies the local-gates-verified label; agents never apply it.

## Done when

- The settings diff is exactly that one line (plus SDLC issue bookkeeping in the PR), and the operator has labeled it local-gates-verified.
- After weave refresh, the checkout-local .claude/settings.json denyRead includes ~/.config/rclone.

## Plan

- [x] Verify the one-line settings change, operator label, checkout-local regeneration and recorded suite.

## Log

### 2026-10-18

### 2026-10-18 — targeted checks and operator verification

The requested ordering check failed before the change and passed afterwards. Parsed JSON verifies one unique rclone entry immediately after gcloud and byte-for-byte equality with main except that added line. Existing settingsx and plan tests passed (0.185s and 1.755s); git diff --check passed. PR #311 is open; its local-gates-verified label was applied by the operator and independently observed after TL receipt d52e519e-cc57-4ffb-9af9-432856f8a3c6. The worker did not apply the label. The recorded suite is running on the pushed settings head.

Before regeneration the checkout-local denyRead lacked rclone. weave refresh --substrate completed successfully on this issue branch, preserved its HEAD and clean tracked tree, and regenerated .claude/settings.json with rclone immediately after gcloud. Verification read only settings lists; no credential files were read.

### 2026-10-18 — recorded suite green; hold lifted

The recorded suite passed on settings head 93dd985838f2078e31e7f2e9da41d30404d5a72f: all 1,651 SDLC tests and 70 other packages, 12 shards, exit 0 in 571.2s after 2,145s queue wait. Log: .git/sdlc/local-checks/93dd985838f2078e31e7f2e9da41d30404d5a72f/suite.log, sha256 prefix 2f833a24df8b. Subsequent branch changes are issue text only. The operator's disk-trace hold arrived after the suite had finished; no heavy work was running, and none started during the hold. The TL lifted it in receipt 39691a6e-1d71-4108-bad9-f3bd5a59153a and directed close/landing. The operator-applied label remains present.

## Revisions

### 2026-10-18 — clarify generated settings target and workflow bookkeeping

The original criterion named ~/.claude/settings.json. The base manifest emits the checkout-local .claude/settings.json; the home-level file is outside weave's output. The TL approved verifying regeneration in this slot and correcting the criterion (receipt 762cc22c-c775-45bc-a245-72c70309561a); home-level activation remains the TL/operator's separate post-landing rollout. The TL also explicitly accepted the single settings line plus SDLC-generated issue metadata (receipt dd2b879e-989e-4e5e-95ed-73419526485a). The current criteria now state both facts; the functional change remains one line.
