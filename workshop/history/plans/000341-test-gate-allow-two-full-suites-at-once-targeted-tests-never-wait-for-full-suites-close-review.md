# Boundary Review — ariadne#341 (whole-issue close)

| field | value |
|-------|-------|
| issue | 341 — test gate: allow two full suites at once; targeted tests never wait for full suites |
| repo | ariadne |
| issue file | workshop/issues/000341-test-gate-allow-two-full-suites-at-once-targeted-tests-never-wait-for-full-suites.md |
| boundary | whole-issue close |
| milestone | — |
| window | 668a929df6e81e84c740d52af345bf9f27c1543a..8f9acde073791c7fc2544c1c8e3aa4b09f1434f8 |
| command | sdlc close --issue 341 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T16:27:24-07:00 |
| verdict | SHIP |

## Review

I've read the diff. Now I'm running the tests to check the claims.

```verdict
verdict: SHIP
confidence: high
```

The change delivers all three Spec points and both Done-when clauses. `scripts/test-lock.py` now uses N flock slot files (`WF_TEST_SUITES`, default 2). Slot 0 keeps #319's filename, so an older copy of the script still counts against the cap. `pkg/testgate` no longer checks the full-suite lock at all, so targeted runs only wait for their own `test-gate.<k>.lock` slots. The atlas, the `Makefile.workflow` comment and `AGENTS.base.md` all state the new rule. I ran both test suites in the review checkout:

- `go test ./pkg/testgate`: passes, and `go vet` is clean.
- `scripts/test-lock.test.sh`: 22 of 23 pass. The one failure is "waiter names the holder". It hardcodes the repo name `ariadne`, but this checkout's folder is named `0e643ad2f44af2b3`. The behaviour is correct (the waiter does name the holder); the test just depends on the folder name. That makes it a Minor finding, not a code bug.

All the new two-slot cases pass. Nothing blocks shipping.

1. **Strengths**
   - `scripts/test-lock.py:52-66`: each check takes at most one free slot and releases any extra free slot it locked. This keeps the "first free slot" rule simple and leaks no locks.
   - `scripts/test-lock.py:121-124`: a nested run counts as re-entrant even when another slot is free. It releases the free slot it had taken first, so a nested `make test` never takes a second slot. The test covers this both with a slot free and with the holder on slot 1.
   - Keeping slot 0 as `test-suite.lock` keeps the cap working while some checkouts still run the old script. This is noted in the code and the atlas.
   - `pkg/testgate/testgate_test.go:158-178`: `TestEnterIgnoresFullSuites` holds both suite files exclusively and fails on any sleep. If the old suite check came back, it would wait and fail this test, so the test proves the new behaviour.
   - The leftover Go code was removed cleanly: `suiteLock`, `suiteRunning`, the suite branch in `describe`, and the `record` fields only that branch used. A search for `suiteLock` or `WF_TEST_LOCK_FILE` in Go code finds nothing left.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `scripts/test-lock.test.sh:23` hardcodes `ariadne` as the repo name, so the test fails in any checkout whose folder has another name (it fails in this review checkout). The two-slot block avoids this by matching on `make pid $p` only. I checked the whole window and this is the only assertion that depends on the repo name.
   - `atlas/workflow/sdlc-binary.md:1658`: the "Slots: it **then** takes one of…" bullet still says "then", but the step before it (waiting for a full suite) is gone.
   - `Makefile.workflow:9`: the edited comment line is much wider than the lines around it. Rewrap it.

5. **Test coverage**
   - Done-when, full-suite side: two suites acquire, a third waits naming both holders, then takes the freed slot. All covered, using `sleep` processes as holders and private lock files.
   - Done-when, targeted side: a targeted run starts at once while two suites hold their slots. Covered in Go.
   - The single-slot mode (`WF_TEST_SUITES=1`) is still exercised by the original cases, so both modes are tested.
   - Bad `WF_TEST_SUITES` values: a non-number is tested. Zero or negative values are clamped to 1 by `max(1, …)` but not tested; that's acceptable.

6. **Architecture**
   - **ARCH-DRY: pass.** One flock mechanism is reused for N files, and the duplicate suite check in Go was deleted rather than kept alongside.
   - **ARCH-PURE: pass.** This is lock-file code, so it's IO by nature. The Go `enter` takes an injected `sleep` and config, and the tests use temporary directories with real flocks as fakes.
   - **ARCH-PURPOSE: pass.** I checked every reader of the suite lock. `test-lock.py` now uses N slots, `testgate` no longer reads it, and the atlas, Makefile and AGENTS docs are updated. No other doc restates the old one-suite rule; I searched README, AGENTS, atlas and scripts.
   - For future work: re-entrancy relies on the nested run seeing the same `WF_TEST_SUITES` as its parent. The setting is inherited through the environment, so the risk is negligible.

7. **Plan revisions:** none. The plan's checklist matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: test-env-coupling
    title: |
      test-lock.test.sh "waiter names the holder" hardcodes the repo name ariadne
    detail: |
      scripts/test-lock.test.sh:23 greps for "  - ariadne (", but the repo field is the checkout folder's name, so the case fails in any checkout not named ariadne (it fails in this review checkout, 0e643ad2f44af2b3). Match on "make pid $a" only, as the new two-slot block does, or use "$(basename "$(git rev-parse --show-toplevel)")". This is the only such assertion in the window.
  - id: new
    severity: Minor
    family: doc-claim-drift
    title: |
      Atlas Slots bullet still says "then", but the step before it was removed
    detail: |
      atlas/workflow/sdlc-binary.md:1658 "it then takes one of WF_TEST_GATE_SLOTS slots" followed the removed "wait for a full suite" step; drop "then". Makefile.workflow:9 also has an unwrapped, overlong comment line.
```

---

## Re-review — 2026-10-12T16:30:00-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 341 — test gate: allow two full suites at once; targeted tests never wait for full suites |
| repo | ariadne |
| issue file | workshop/issues/000341-test-gate-allow-two-full-suites-at-once-targeted-tests-never-wait-for-full-suites.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2267de44b35cc3304f17975e2d6f3a72e0228b18..822279b2bf80deff628defb8165529060ad13de1 |
| command | sdlc close --issue 341 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T16:30:00-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both earlier findings are fixed, and the fix diff adds no new defects. It is small: one test line, one atlas word, and a rewrapped comment, plus the gate ledger and review record files. `## Spec`, `## Revisions` and `## Done when` are unchanged, so this is not a scope change, and the window has no conflict resolutions. I ran `bash scripts/test-lock.test.sh` in this review checkout. Its folder is named `087fa04e4cd2d4ce`, not `ariadne`. All 23 cases passed, including "waiter names the holder". The earlier round saw that case fail in a checkout with a non-`ariadne` name, so the old line is the failing case and the fix is what makes it pass.

1. **Strengths**
   - `scripts/test-lock.test.sh:23` builds the expected repo name the same way the script does. The script uses `os.path.basename` of `git rev-parse --show-toplevel` (`scripts/test-lock.py:155-156`), so the test still checks the repo field instead of dropping it.
   - `atlas/workflow/sdlc-binary.md:1658` now reads correctly, with the removed step's "then" gone.
   - `Makefile.workflow:8-15`: the comment is rewrapped to match the lines around it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. One cosmetic point I'm not raising as a finding: the rewrap leaves a short line at `Makefile.workflow:12`, "releases the lock. The helper resolves beside".

5. **Test coverage**
   - The lock test suite passes in a checkout not named `ariadne`, which is exactly the case BR-1 described.
   - I searched both test files for other hardcoded `ariadne` names. `scripts/test-lock.test.sh` has none left. The only other hit is `pkg/testgate/testgate_test.go:263` (`/state/ariadne`), which is an input fixture the test sets up itself, not the checkout's name, so it isn't the same problem.

6. **Architecture (interdiff only)**
   - **ARCH-DRY: pass.** The test takes the name from the same source as the code instead of keeping its own copy.
   - **ARCH-PURE: pass.** Nothing here adds logic.
   - **ARCH-PURPOSE: pass.** The fixes cover the whole problem, not just the named line: no other test asserts a hardcoded repo name, and no other "then" was left behind by the removed step.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      test-lock.test.sh:23 now takes the repo name from git rev-parse --show-toplevel, as test-lock.py does; all 23 cases pass in checkout 087fa04e4cd2d4ce, and no other ariadne hardcode is left in the file.
  - id: BR-2
    disposition: addressed
    note: |
      atlas sdlc-binary.md:1658 drops "then"; Makefile.workflow:8-15 comment rewrapped to the width of the lines around it.
```
