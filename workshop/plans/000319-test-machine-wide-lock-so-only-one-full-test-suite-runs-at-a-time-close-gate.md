---
gate: boundary-review
issue: 319
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T11:36:40-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: A waiter crashes (KeyError or AttributeError) on a valid-JSON lock record with an unexpected shape
          detail: 'holder() handles only ValueError; describe() indexes rec[''repo''] and the rest directly. Reproduced: a held lock with record {"pid":1} makes the waiter traceback, and make test fails with "helper failed" instead of waiting. Validate in holder() (a dict, else None), use rec.get with defaults in describe(), and add a test case.'
          family: untrusted-persisted-record-shape
          round: 1
        - id: BR-2
          severity: Minor
          title: The wildcard guard on scripts/test-lock.py skips the lock without a word in repos not yet refreshed
          detail: $(WF_WORKFLOW_SOURCE_DIR)scripts/test-lock.py always resolves to ariadne's helper and removes the dependency on weave refresh; otherwise warn when the helper is missing.
          family: silent-guard-bypass
          round: 1
        - id: BR-3
          severity: Minor
          title: The watcher keeps default signal handling in make's process group, so a group signal releases the lock mid-suite
          detail: Ignore SIGINT and SIGHUP in the watcher and rely on watching make's exit; separately, the Linux polling fallback should treat PermissionError as the pid still being alive.
          family: watcher-lifetime-coupling
          round: 1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-10T11:38:42-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: holder() requires a dict with all four keys; the {"pid":1} repro now reads as unidentified and returns timeout. The [1, 2] regression case would raise AttributeError without the fix.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The helper resolves via $(WF_WORKFLOW_SOURCE_DIR), and a missing helper emits $(warning) that the run is unlocked (Makefile.workflow:15-18).
          round: 2
        - id: BR-3
          disposition: addressed
          note: The watcher ignores SIGINT and SIGHUP (tested by "watcher survives SIGINT and SIGHUP"), and the poll fallback treats PermissionError as alive.
          round: 2
      findings:
        - id: BR-4
          severity: Minor
          title: describe(None) says "record not yet written" for a foreign or malformed record too
          detail: This is message wording only; the rule (validate the shape before use) is already applied in holder(). "unidentified holder" alone would be accurate.
          family: untrusted-persisted-record-shape
          round: 2
      recipe: milestone-review
      reviewed: 67d51fbebd8b7fc37eb7592647326fc68f6c8391
      blocked: false
---

# Gate ledger — ariadne#319 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T11:36:40-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `untrusted-persisted-record-shape` A waiter crashes (KeyError or AttributeError) on a valid-JSON lock record with an unexpected shape
  holder() handles only ValueError; describe() indexes rec['repo'] and the rest directly. Reproduced: a held lock with record {"pid":1} makes the waiter traceback, and make test fails with "helper failed" instead of waiting. Validate in holder() (a dict, else None), use rec.get with defaults in describe(), and add a test case.
- **BR-2** [Minor] `silent-guard-bypass` The wildcard guard on scripts/test-lock.py skips the lock without a word in repos not yet refreshed
  $(WF_WORKFLOW_SOURCE_DIR)scripts/test-lock.py always resolves to ariadne's helper and removes the dependency on weave refresh; otherwise warn when the helper is missing.
- **BR-3** [Minor] `watcher-lifetime-coupling` The watcher keeps default signal handling in make's process group, so a group signal releases the lock mid-suite
  Ignore SIGINT and SIGHUP in the watcher and rely on watching make's exit; separately, the Linux polling fallback should treat PermissionError as the pid still being alive.

## Round 2 — 2026-10-10T11:38:42-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — holder() requires a dict with all four keys; the {"pid":1} repro now reads as unidentified and returns timeout. The [1, 2] regression case would raise AttributeError without the fix.
- BR-2 — addressed — The helper resolves via $(WF_WORKFLOW_SOURCE_DIR), and a missing helper emits $(warning) that the run is unlocked (Makefile.workflow:15-18).
- BR-3 — addressed — The watcher ignores SIGINT and SIGHUP (tested by "watcher survives SIGINT and SIGHUP"), and the poll fallback treats PermissionError as alive.

### Raised

- **BR-4** [Minor] `untrusted-persisted-record-shape` describe(None) says "record not yet written" for a foreign or malformed record too
  This is message wording only; the rule (validate the shape before use) is already applied in holder(). "unidentified holder" alone would be accurate.

## Open findings

- **BR-4** [Minor] `untrusted-persisted-record-shape` describe(None) says "record not yet written" for a foreign or malformed record too
