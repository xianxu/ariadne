# Boundary Review — ariadne#333 (whole-issue close)

| field | value |
|-------|-------|
| issue | 333 — Full-suite lock: a sandboxed make test should fail its preflight before queueing for the lock |
| repo | ariadne |
| issue file | workshop/issues/000333-full-suite-lock-a-sandboxed-make-test-should-fail-its-preflight-before-queueing-for-the-lock.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1ecde6535c8c6899120f80f504f38293fccf4180..5006a0a7da3da93864e40981c6bc2cd6e7fa2d1d |
| command | sdlc close --issue 333 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T23:14:58-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation satisfies the pre-lock ordering contract, and the targeted suite passes. One Important documentation gap remains: README does not describe the new configuration key and automatic preflight convention.

Done-when: current — rejecting an unsuitable host before queueing and retaining normal lock acquisition still cover the revised scope; omitting ariadne’s own probe does not change those outcomes.

1. **Strengths**
   - `Makefile.workflow:28–38` runs the preflight before invoking the lock helper and preserves its diagnostic output.
   - `scripts/test/test-lock.test.sh:159–171` exercises failure against a held lock, successful preflight against held/free locks, and absence of a preflight.
   - `atlas/workflow/sdlc-binary.md:1635` documents the hook and ariadne’s opt-out rationale. The existing kernel-based `testfix.AwaitExit` supports the rationale for removing the proposed `ps` probe.

2. **Critical findings:** None.

3. **Important findings**
   - **README update missing** — `Makefile.workflow:20–27` introduces `WF_TEST_PREFLIGHT` and automatic discovery of `tests/host-preflight.sh`, but README is unchanged. Add a short usage paragraph covering both surfaces, assignment before the include, and failure before lock acquisition; link the atlas explanation. These are all instances of this documentation-gap family in the window.

4. **Minor findings:** None.

5. **Test coverage notes**
   - `bash scripts/test/test-lock.test.sh` passed, including all five new assertions.
   - A temporary fixture independently confirmed default-path discovery, preservation of exit-7 diagnostics, and exclusion of non-`test` goals. The corresponding refusal disappeared using the base Makefile.
   - The committed tests exercise custom paths but do not retain that default-path-present fixture.
   - Pair’s actual sandbox execution was not independently reproduced; the base-layer convention was verified locally.

6. **Architectural notes**
   - **ARCH-DRY: pass** — one shared hook; host-specific probes remain repository-owned.
   - **ARCH-PURE: pass** — the addition is a small IO boundary without embedded business logic.
   - **ARCH-PURPOSE: pass** — failure precedes lock acquisition, and successful probes retain existing acquisition behavior.

7. **Plan revision recommendations:** None; the existing revision explains the ariadne scope adjustment.

```findings
findings:
  - id: new
    severity: Important
    family: user-facing-surface-documentation
    title: |
      README update appears missing for the host-preflight hook
    detail: |
      Makefile.workflow:20–27 introduces two user-facing surfaces: automatic tests/host-preflight.sh discovery and the WF_TEST_PREFLIGHT override. Both are documented in atlas/workflow/sdlc-binary.md:1635, but README.md is unchanged. Add a concise README usage paragraph covering the convention, setting the override before including Makefile.workflow, and failure before lock acquisition, with an atlas link. These are all instances of this family in the window.
```

---

## Re-review — 2026-10-13T23:17:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 333 — Full-suite lock: a sandboxed make test should fail its preflight before queueing for the lock |
| repo | ariadne |
| issue file | workshop/issues/000333-full-suite-lock-a-sandboxed-make-test-should-fail-its-preflight-before-queueing-for-the-lock.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1ecde6535c8c6899120f80f504f38293fccf4180..ca1831dbccacfc1c2ca73fdd30fb8cb6d14ebcf4 |
| command | sdlc close --issue 333 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T23:17:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The preflight executes before lock acquisition, preserves diagnostics, and stops failing hosts without queueing. Passing and absent preflights retain normal lock behavior. BR-1 is addressed by the README addition; no blocking findings remain.

Done-when: current — both failure-before-queueing and normal lock acquisition still cover the revised scope; omitting ariadne’s own probe does not change those outcomes.

1. **Strengths**
   - `Makefile.workflow:27–38` places the shared hook before the lock helper.
   - `scripts/test/test-lock.test.sh:159–171` checks failing, passing, and absent preflights against held/free locks.
   - `README.md:222` and `atlas/workflow/sdlc-binary.md:1635` document discovery, override placement, and failure behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - `bash scripts/test/test-lock.test.sh` passed, including all five new assertions.
   - A disposable fixture verified default-path discovery, exit-7 diagnostics, and exclusion of non-`test` goals. The same failure assertion fails against the base Makefile.
   - Pair’s actual sandbox run was not independently reproduced; the base-layer behavior was verified.
   - `git diff --check` passed.

6. **Architectural notes**
   - **ARCH-DRY: pass** — one shared hook delegates host checks to repository-owned scripts.
   - **ARCH-PURE: pass** — the change remains a small IO boundary.
   - **ARCH-PURPOSE: pass** — refusal precedes lock acquisition, while successful probes preserve acquisition behavior.

7. **Plan revision recommendations:** None. The existing revision explains ariadne’s opt-out; the kernel-based `testfix.AwaitExit` supports its stated rationale.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:222–227 now documents tests/host-preflight.sh discovery, WF_TEST_PREFLIGHT assignment before inclusion, and failure before queueing, with an atlas link. These claims match Makefile.workflow:27–38. This is a prose-only correction; no wording-presence test is required.
```
