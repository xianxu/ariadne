---
id: 000382
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '343f6315d1f97d8a8977e09912ab0b366ec0f010' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T21:42:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c44738ca", done: "e42711c6"}
---

# test-shard.test.py creates its temp dir under /tmp, which agent sandboxes deny; use TMPDIR

## Problem

`scripts/test-shard.test.py:24` (from #369) creates its temp dir with
`tempfile.TemporaryDirectory(prefix="shard-report-", dir="/tmp")`. Agent sandboxes (Claude Code's) deny writes to
`/tmp`, so `TestShardedReportRetainsSubtestDiagnostics` fails with PermissionError in every sandboxed `make test`
(found by ariadne:4 on 10-14 while running #367's suite). Unsandboxed runs pass, so the recorded suite hides it.

## Spec

Drop `dir="/tmp"` so the temp dir follows `TMPDIR` (the #342 convention). Sweep the repo's scripts and tests for other
hardcoded `/tmp` temp locations and fix them the same way.

## Done when

- The test passes with `TMPDIR` pointing at a writable directory and `/tmp` unwritable (a test or a documented check).
- No hardcoded `/tmp` temp dirs remain in scripts or tests, or each remaining one is justified in a comment.

## Plan

- [x] Remove hardcoded test temporary-directory roots found by the sweep.
- [x] Verify the report with a writable TMPDIR and denied outside allocation; run the recorded full suite.

## Log

### 2026-10-14

Claimed ariadne:5 and started from fresh main. The authorized small design is to let Python tempfile select its environment-configured root, removing explicit /tmp overrides rather than adding another fallback policy (ARCH-DRY). Initial inventory found the report test and embedded Python in scripts/test/slot-dependencies.test.sh. Wider scratch-state findings were sent to the hub for scope guidance: brainstorming start/stop is a coupled managed-skill protocol, and close-issue.py uses a warmup counter file rather than a temporary directory. A documented allocation-denial check will execute the real report test with a custom writable TMPDIR and reject temporary allocations outside it; the issue explicitly permits a documented check.

### Restricted-root verification

Before the fix, a real macOS sandbox with writable custom TMPDIR and denied writes under /tmp and /private/tmp reproduced PermissionError in both the report test and the slot-dependencies bootstrap. After removing the two dir overrides, the same check passes; the explicit raw-/tmp allocation control remains denied. Commands/results: python3 -B /tmp/ariadne382-tmpdir-check.py; RED /tmp/ariadne382-tmpdir-red.log, GREEN /tmp/ariadne382-tmpdir-green.log. The bootstrap check executes only the existing allocation prefix and verifies its evidence directory is a child of TMPDIR, without running the unrelated product acceptance suite. The existing Go wrapper TestShardedReportRetainsSubtestDiagnostics passes (0.456s; /tmp/ariadne382-shard-report.log).

Reproduce the report check from a writable TMPDIR outside /tmp:

```sh
test_root=$(mktemp -d "${TMPDIR:?}/ariadne382.XXXXXX") || exit 1
trap 'rmdir "${test_root:?}"' EXIT
TMPDIR="${test_root:?}" sandbox-exec -p '(version 1) (allow default) (deny file-write* (subpath "/tmp") (subpath "/private/tmp"))' python3 -B scripts/test-shard.test.py
```

The read-only sweep inspected allocator and /tmp matches across 693 tracked shell/Python/Go-test/testdata files in scripts, construct, cmd, pkg, .colima and .openshell. Exactly the two changed test allocations force /tmp. Intentional literals in pkg/shelllint/mktemp_test.go exercise rejection; other Go/testdata hits are sample paths or error strings. Proper ${TMPDIR:-/tmp} fallbacks remain. Wider scripts findings are pending hub scope guidance: managed brainstorming start/stop default session paths (active adaptation and frozen copies), and the legacy close-issue.py warmup counter file.

## Revisions

### 2026-10-14 — classify intentional sweep exceptions

The current Done when already permits remaining hardcoded locations when comments justify them. The read-only sweep's active brainstorming protocol is an intentional exception: construct/adapted/superpowers-brainstorming/scripts/stop-server.sh:4-7 explicitly documents deletion only under /tmp, preserving project-local .superpowers content, and line 24 enforces that allowlist. Its start-server.sh default is paired with that cleanup contract; sandboxed callers can select --project-dir. Frozen copies under construct/sources/superpowers/v5.0.2/skills/brainstorming and construct/versions/0004/superpowers-brainstorming preserve the same commented upstream protocol. They are not test allocator overrides and are not edited or auto-promoted here. The legacy scripts/close-issue.py warmup counter is a state file, not a temporary-directory allocation; shelllint's literal /tmp input intentionally tests rejection. These are explicit inventory classifications under the existing exception criterion, not an inferred approval of the unanswered hub question. The two actual test allocations are fixed and verified with raw-/tmp writes denied.

### 2026-10-14 — acceptance checkpoint

Expanded the read-only sweep to 1,139 tracked code/config/executable files, including JavaScript, TypeScript, Ruby, Makefiles and extensionless scripts. The only additional directory allocation belongs to the same brainstorming protocol: construct/adapted/superpowers-brainstorming/scripts/server.js:79 defaults to /tmp/brainstorm and creates it at line 261; frozen source/version copies match. The paired cleanup comments and --project-dir alternative above cover this protocol family. Other /tmp matches are output files, sandbox allowlists, intentional fixtures, or environment-respecting fallback templates. No additional test allocator override was found.

The recorded suite on pushed code commit c65defd6e618c09a79d59e8cfc5af22a362e2ec6 is queued via sdlc local-check run --name suite (session 94512); at 22:11 PDT it was fourth in line after about 23 minutes. Keep this existing run, rather than creating another. Only issue documentation changed after that code commit.

Current main gained #377 code while the first run was queued. Canceled that waiting run before any tests started, merged origin/main cleanly, and pushed integrated commit 517df7de41290f17677dfd2028d2e4e307374884. Replacement recorded suite session 36941 uses that integrated head; log: /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/517df7de41290f17677dfd2028d2e4e307374884/suite.log. A transient GitHub branch-push rejection was probed (remote still held c65defd6) and retried successfully before starting the replacement.

### Recorded acceptance passed

Integrated head 517df7de41290f17677dfd2028d2e4e307374884 passed sdlc local-check run --name suite, recorded as sdlc-local/suite: all 12 shards over 1,368 cmd/sdlc tests and 59 other packages exited zero, no failures, wall 564.3s. Queue wait was 2,134s. Log SHA-256 prefix f94bf6775cbf. Together with the real restricted-root RED/GREEN and focused wrapper above, this satisfies both acceptance rows. This fixes existing test temporary-root selection and adds no architectural surface.
