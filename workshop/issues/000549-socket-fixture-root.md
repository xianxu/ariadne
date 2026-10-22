---
id: 000549
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
severity: medium
collaboration-mode: collaboration
card_mirror: '36cf33b834ac4f306aafc6a3db174a023b325620' # card fields mirrored from issue-cards; edit via sdlc
---

# Lint inherited temp roots used for Unix socket fixtures

## Problem

pair#792 close review BR-3 found an inverse API misuse that the shared tmp-literal rule cannot catch. A private Unix-socket RPC test used os.MkdirTemp("", "ro-") for COUCH_ISOLATED_ROOT, then derived its store with filepath.Join(root, "store") and passed that namespace to couchmessage.SocketPath. Empty-parent MkdirTemp correctly follows TMPDIR for ordinary files, but a long inherited TMPDIR makes the derived Unix socket exceed Darwin's 103-byte limit.

The failing source was cmd/internal/testcouch/agentfixture/resources_test.go in Pair. Reproduction under an existing long TMPDIR failed before starting the server with: isolated message socket exceeds 103 bytes; choose a shorter COUCH_ISOLATED_ROOT. This is a fixture-construction bug, not a production timeout or protocol failure.

Pair's BR-3 fix uses a short, canonical private socket root, prefers /Volumes/TestScratch when mounted, falls back to explicitly short /tmp, and registers cleanup immediately. Its RPC regression now also sets a long TMPDIR internally. The same external long-TMPDIR test passed with race detection in 1.229 s; the resource-observer test selection passed in 0.166 s. No production/native behavior or bound changed.

## Spec

Add a mechanical socket-fixture-root rule class to the shared Go lint machinery, or an equivalent reusable matcher that a consumer can register for its namespace APIs. Detect inherited-temp roots when the actual root/path value reaches a Unix-socket namespace or socket sink. Cover local assignments and aliases, filepath.Join-derived namespace paths, and imported API aliases; make the supported analysis boundary explicit.

Unsafe producers include os.MkdirTemp("", ...), os.TempDir(), and t.TempDir() when their resulting path is used as an unbounded socket root. Pair's concrete sink is couchmessage.SocketPath with COUCH_ISOLATED_ROOT selecting the private socket parent. Generic Unix socket APIs should use the same mistake class where their path flow can be established. Do not import consumer runtime packages into the linter just to recognize source syntax; choose shared configuration or source-level API recognition during design.

Do not ban TempDir across an entire function or file that uses sockets. The corrected Pair regression legitimately uses t.TempDir for unrelated data that constructs a long TMPDIR while its actual socket root comes from the tested short-root helper. That case must remain clean. Keep ordinary temporary files on TMPDIR/TestScratch and preserve bounded explicit-short-root exceptions with their rationale and cleanup; do not invert the existing tmp-literal policy globally.

This is an independent, unclaimed follow-up to pair#792 BR-3. Filing it does not claim that its catalog change must land before Pair's issue; that sequencing decision belongs to the hubs.

## Done when

- The original inherited-parent namespace flow is mechanically rejected, including local alias and filepath.Join variants.
- Direct inherited-temp socket paths and the supported Unix-socket/consumer namespace APIs have focused positive tests.
- The short-root fixture and unrelated TempDir usage in the same function are negative tests; no function-wide or file-wide ban produces a false positive.
- A long-TMPDIR regression verifies that the safe private socket fixture still starts the server, preserves TestScratch preference when usable, and cleans up completely without changing socket or transport bounds.
- The rule is in the shared catalog or exposed for consumer registration, with an explicit scope and guidance on the short-root exception to tmp-literal.

## Plan

- [ ]

## Log

2026-10-22: Filed from Pair's pair#792 BR-3 repair. Read-only inspection found no socket-fixture rule in Pair's pinned ariadne/pkg/golint or the current Ariadne source; tmp-literal addresses the opposite ordinary-file misuse. The Pair repair and its red/green verification remain on Pair's own issue branch. No Ariadne implementation, branch, or checkout edit is part of this filing.
