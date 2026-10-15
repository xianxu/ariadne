# Boundary Review — ariadne#382 (whole-issue close)

| field | value |
|-------|-------|
| issue | 382 — test-shard.test.py creates its temp dir under /tmp, which agent sandboxes deny; use TMPDIR |
| repo | ariadne |
| issue file | workshop/issues/000382-test-shard-test-py-creates-its-temp-dir-under-tmp-which-agent-sandboxes-deny-use-tmpdir.md |
| boundary | whole-issue close |
| milestone | — |
| window | ef5e0bf028c33b130d93b7014be47d310812fc15..5650a7a84b70f1c803fafa294614a273c254f516 |
| command | sdlc close --issue 382 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T23:00:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The two allocation changes satisfy the issue’s contract. Both pinned base versions fail when directory creation outside `TMPDIR` is denied; both head versions pass. No blocking findings remain.

Done-when: current — it covers restricted-root execution and permits commented exceptions, including the documented brainstorming cleanup protocol.

```findings
{}
```

1. **Strengths**
   - `scripts/test-shard.test.py:24` delegates root selection to Python’s existing tempfile policy.
   - `scripts/test/slot-dependencies.test.sh:15` applies the same correction to the sibling allocation.
   - The remaining brainstorming paths match the documented cleanup exception; environment-respecting fallbacks remain intact.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Independently confirmed base-fails/head-passes for both allocations using a Python audit hook, including a denied raw-`/tmp` control. The report passes with and without `TMPDIR`, and `TestShardedReportRetainsSubtestDiagnostics` passes. The recorded full-suite log reports all shards passing; its hash matches the issue, and only issue documentation changed afterward. Native `sandbox-exec` replay was unavailable because `sandbox_apply` returned “Operation not permitted”; the audit-hook check verified the allocation invariant.

6. **Architecture:** ARCH-DRY passes through reuse of tempfile policy. ARCH-PURE passes: the changes remain within existing IO setup. ARCH-PURPOSE passes: both test overrides are removed, and the remaining directory-allocation family is explicitly classified. No new surface requires atlas or README updates.

7. **Plan revisions:** None required.
