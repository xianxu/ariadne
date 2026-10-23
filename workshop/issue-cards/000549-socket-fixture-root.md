---
id: 000549
status: codecomplete
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 0.882
github_issue:
severity: medium
started: 2026-10-22T13:51:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 4.49
tracker:
    version: 1
    completion:
        token: close-500302bd2761
        repository: github.com/xianxu/ariadne
        reviewed_head: d32c7b0811908755c8cdbb0ce2d7a3b8b17d2cda
        evidence_commit: 3e6762c94afc23dc3305c2906449d3236d674bc4
---

# Lint inherited temp roots used for Unix socket fixtures

## Problem

pair#792 close review BR-3 found an inverse API misuse that the shared tmp-literal rule cannot catch. A private Unix-socket RPC test used os.MkdirTemp("", "ro-") for COUCH_ISOLATED_ROOT, then derived its store with filepath.Join(root, "store") and passed that namespace to couchmessage.SocketPath. Empty-parent MkdirTemp correctly follows TMPDIR for ordinary files, but a long inherited TMPDIR makes the derived Unix socket exceed Darwin's 103-byte limit.

The failing source was cmd/internal/testcouch/agentfixture/resources_test.go in Pair. Reproduction under an existing long TMPDIR failed before starting the server with: isolated message socket exceeds 103 bytes; choose a shorter COUCH_ISOLATED_ROOT. This is a fixture-construction bug, not a production timeout or protocol failure.

Pair's BR-3 fix uses a short, canonical private socket root, prefers /Volumes/TestScratch when mounted, falls back to explicitly short /tmp, and registers cleanup immediately. Its RPC regression now also sets a long TMPDIR internally. The same external long-TMPDIR test passed with race detection in 1.229 s; the resource-observer test selection passed in 0.166 s. No production/native behavior or bound changed.
