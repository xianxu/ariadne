---
id: 000465
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2d510510f3c8feac5b372509f8375268a2d9f3d4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T07:56:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/scripts/test.TestConstructScripts]
flow: {kind: quick, provenance: inferred, spec: "1393e822", done: "76dfe1e4"}
---

# dev-aliases script exceeds its per-script budget under suite load

## Problem

The recorded ariadne suite on head da8c62d2e111ef7fe86cc0c351e00d6512a5a87a failed only `scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`: the per-script 2m deadline fired after about 40 passing assertions. The failed attempt is 4c96305468ac871ff567625384059cc8, log SHA-256 prefix 7714ac8e8e11, saved at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/4c96305468ac871ff567625384059cc8.log`. All 12 cmd/sdlc shards and 1,688 tests passed. An unchanged exact-head targeted run of that subtest passed 82/82 assertions in 33.6 seconds. #461 expanded the script with real Git layouts and sentinel binary cases; that cost crosses #448's fixed two-minute per-script ceiling under fleet suite load. The #458 branch changes no construct script or runner code.

## Spec

Keep the per-script ceiling. Reduce the cost of `construct/scripts/test/dev-aliases.test.sh` itself through shared fixtures, fewer repeated Git initializations or independent parallel subtests, while preserving its assertions. Register the exact test as a known flake during the repair so a same-head targeted rerun can provide landing evidence.

## Done when

- All 82 original behavioral assertions remain covered, including partial/concurrent builds, failed-output preservation and start < mid-build edit < completed-output timestamps. The timestamp regression fails if build-start timestamp restoration is removed. Keep the two-minute per-script ceiling unchanged.
- Pure path splitting matches the real dirname/basename utilities for root, empty/single components, trailing/repeated slashes, spaces and newlines. Counted N/2N owner and CLI tests require zero path-utility subprocesses from the generator.
- Counted profiling demonstrates the operation reduction; repeated targeted checks and a loaded recorded full suite pass. On the noisy shared machine, acceptance uses these counts and the loaded-suite pass rather than a standalone wall-time threshold, as approved by the hub.

## Plan

- [x] Measure repeated generator scans and waits; isolate preparation and duplicate-origin cases to their relevant real fixture owners (ARCH-DRY).
- [x] Replace fixed concurrency polling with bounded readiness/finish handshakes and use explicit timestamps for edits during builds, retaining all 82 assertions and the existing deadline.
- [x] Preserve all 82 assertions and pass repeated standalone/runner-targeted checks, real layouts, path conformance and counted N/2N CLI checks.


## Log

2026-10-19: Classified while closing ariadne#458. Targeted `go test ./scripts/test -run '^TestConstructScripts/plain/dev-aliases.test.sh$' -count=1 -v` passed in 33.779 seconds after the recorded full-suite timeout. The terminal failure was `scripts_test.go:70: .../construct/scripts/test/dev-aliases.test.sh did not finish within 2m0s`; the partial output had passed the prior build and failure-preservation assertions.

2026-10-19: Claimed and measured the unchanged script: 82/82 assertions in 17.36s. A disposable fixture-cwd-only experiment took 16.74s, disproving live substrate traversal as the dominant cost here (this checkout has no construct/deps). Repeated generator calls scan unrelated owners, and fixed waits add several seconds. Quick test-only plan posted to hub; no production behavior or deadline changes.

## Revisions

2026-10-19 — Hub approved a production cost reduction after test-only timings remained 9.1–11.3s: replace primary_owner's repeated external dirname/basename processes with pure Bash path splitting. Match real utilities for trailing slashes, root, single components and repeated slashes using a small conformance table, and retain real-layout tests. This is the same owner-location rule and adds no state. The hot path visits each candidate once; path splitting should spawn zero external path utilities per candidate, with counted N/2N tests.

2026-10-19 — Acceptance clarification from hub: no standalone benchmark window or further timing chase while the shared machine is noisy. Accept the counted operation reduction, all 82 assertions passing, and a loaded recorded-suite pass. Successful build outputs share one immutable sentinel inode per run; partial, failed and concurrent output fixtures remain independent. The edit-during-build stub uses an environment flag instead of rewriting itself. The original two-minute ceiling is unchanged. Path splitting now serves all equivalent startup/collection sites, guarded by real-tool conformance and N/2N CLI counts.

2026-10-19 — Final forwarded-command profiling, original and final real generators with frozen test sources: both 82/82 passed. Generator launches 38→27; Bash launches 59→33; Git calls 165→53 (config queries 148→49; four real inits retained); dirname 410→6 and basename 783→6. Remaining 12 path utilities belong to unchanged substrate parsing in explicit non-peer/slot cases, not the generator. Final counter artifacts: `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/aliases465-final-source-count.tr4rvrvw`. Path table and N/2N CLI guard pass. A temporary production mutation removing the installed nous binary before rebuilding makes the new FIFO-selected missing/partial-binary assertion fail (77 pass, five failures including the independent shape/failure-preservation guards).

2026-10-19 — Final targeted verification on implementation head: `go test ./pkg/shelllint ./scripts/test -run 'TestTracked|TestConstructScripts/(plain|no_color_ci|forced_color_terminal)/dev-aliases(-path|-slot)?.test.sh$' -count=2 -v` passed. This runs each of the three alias scripts twice in all three color environments; all six original-script runs retain 82/82 assertions. The last original-script run completed in 4.56s under the shared runner. PR #323 is open; recorded suite attempt f538d680216897b9889f96936a7dba3c on 72b6482b43cb022b2a124ef43a078d53057850d3 is queued, next in normal FIFO. No production or test code changes after that head.

2026-10-19 — CI found a deterministic Linux fixture defect: GNU mv refuses a rename when source and installed target are hardlinks to the same inode. Local BSD mv had accepted that fixture. Changed successful build stubs to copy the single immutable sentinel template, preserving distinct output inodes and metadata. Cancelled queued attempt f538d680216897b9889f96936a7dba3c through `sdlc local-check cancel --name suite` before it acquired the machine slot; this is a branch fix, not a flake or retry. Fresh targeted/CI/recorded checks follow on the fixed head.

2026-10-19 — Review overlap checkpoint: split the last Plan row at the implementation boundary. Repeated targeted checks are complete; the recorded-suite pass remains mandatory in Done-when and at landing. This preserves scope while allowing the worker protocol's authorized close-review/queued-suite overlap. On the copied-sentinel head 6e5120fd877baaac47dd1c4c7096a8bc819d0ce4, standalone 82/82 and all six three-environment runner repetitions passed (139.566s total package time). Fresh recorded attempt 6bca6bbc7eba03b5e334ea56b3ee49ea is queued; landing waits for its pass and green CI. #465's flake registration is retired only after landing/completion, not manually during this queue.

2026-10-19 — Close round 1 BR-1/BR-2 correction: replaced the sole mid-build timestamp fixture with controlled start=2000, edit=2001, output=2002 times, so a missing restoration is observable without sleeping. Other future-date fixtures deliberately test staleness, not start-time restoration. Synchronized the current Done-when contract with the approved path/conformance/count requirements and noisy-machine acceptance; preserved all 82 assertions and the loaded recorded-suite requirement. Cancelled still-queued attempt 6bca6bbc7eba03b5e334ea56b3ee49ea before this test-code fix; a fresh final-head suite is required.

2026-10-19 — BR-1 proof: normal ordered-time fixture passes 82/82. A temporary real-generator mutation replacing `touch -r` with a no-op makes the behavioral assertion “an edit made during the build makes the next call rebuild” fail with an empty build trace (79 pass, three failures including two independent function-shape guards). Mutation artifacts: `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/aliases465-timestamp.Ib4FQ0`. No sleep or ceiling change was introduced.
