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

## Open findings

- **BR-1** [Important] `untrusted-persisted-record-shape` A waiter crashes (KeyError or AttributeError) on a valid-JSON lock record with an unexpected shape
- **BR-2** [Minor] `silent-guard-bypass` The wildcard guard on scripts/test-lock.py skips the lock without a word in repos not yet refreshed
- **BR-3** [Minor] `watcher-lifetime-coupling` The watcher keeps default signal handling in make's process group, so a group signal releases the lock mid-suite
