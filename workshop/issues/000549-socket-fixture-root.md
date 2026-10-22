---
id: 000549
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
severity: medium
collaboration-mode: collaboration
card_mirror: '8133e9838daa23e9df627027d162bd2b5eeec7cb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T13:51:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9802344c", done: "a64e72e1"}
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

- [ ] Red-first recorded source replay and source-to-socket flow controls.
- [ ] Implement the shared catalog rule and counted local provenance analysis.
- [ ] Verify short-root Unix sockets under long TMPDIR, document the boundary, run recorded suite and close review.

Detailed design: [implementation plan](../plans/000549-socket-fixture-root-plan.md). Awaiting TL approval before code.

## Log

2026-10-22: Filed from Pair's pair#792 BR-3 repair. Read-only inspection found no socket-fixture rule in Pair's pinned ariadne/pkg/golint or the current Ariadne source; tmp-literal addresses the opposite ordinary-file misuse. The Pair repair and its red/green verification remain on Pair's own issue branch. No Ariadne implementation, branch, or checkout edit is part of this filing.

2026-10-22: Claimed by ariadne:5 and entered planning after #552 landed. Read the shared AST-only catalog, existing temp-sweep provenance matcher, Pair's corrected resource fixture and actual SocketPath implementation. Important distinction: SocketPath hashes its namespace, while COUCH_ISOLATED_ROOT selects its socket parent; the rule must recognize the isolation-root setting as well as namespace flow. Pair's existing corrected fixture remains clean despite unrelated t.TempDir used to construct long TMPDIR. Source-level API recognition avoids importing Pair runtime packages. No code changed; plan submitted for TL approval. Initial Couch socket was briefly unavailable; subsequent canonical receipt lookup verified the dispatch.

## Revisions

2026-10-22: Added the concrete implementation plan and analysis boundary without changing the requested acceptance criteria. Use the shared catalog with source-level API recognition; do not add a consumer runtime dependency or a general interprocedural analyzer.
