# Boundary Review — ariadne#358 (whole-issue close)

| field | value |
|-------|-------|
| issue | 358 — weave: startup make tools inherits MAKEFLAGS=k from an outer make -k, so a failing tools recipe no longer stops later ones |
| repo | ariadne |
| issue file | workshop/issues/000358-weave-startup-make-tools-inherits-makeflags-k-from-an-outer-make-k-so-a-failing-tools-recipe-no-longer-stops-later-ones.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5b513d43608477ba059418ed9fd84e6068fc705..169d4e6a1dd2416bc25f40fb8364e3c68e51b639 |
| command | sdlc close --issue 358 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T19:08:24-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The production fix is correct and the regression detects its removal. Startup tests pass with both empty and inherited keep-going settings. The blocker is a tracker inconsistency: the revised acceptance contract replaces the full-suite requirement, but `## Done when` still states the old requirement.

Done-when: stale — it still requires `make -k test`, while Revisions substitutes all weave packages, both startup environments, and green PR CI.

1. **Strengths**
   - `tools.go:36` cancels keep-going at the existing shared process boundary without stripping other environment settings.
   - `tools_test.go:148–165` exercises real Make and checks both error propagation and absence of the later recipe’s output.
   - Removing only `-S` in a temporary copy made the inherited-keep-going case fail; the clean case still passed. The temporary copy was removed.

2. **Critical findings:** None.

3. **Important findings**
   - [Issue #358:48](/Users/xianxu/Library/Caches/sdlc/review-checkouts/19236f6b9a40e7c4/workshop/issues/000358-weave-startup-make-tools-inherits-makeflags-k-from-an-outer-make-k-so-a-failing-tools-recipe-no-longer-stops-later-ones.md:48): synchronize the acceptance criteria. The complete family in this window comprises the old full-suite criterion at line 48, its replacement at line 69, and line 70’s “Revised acceptance passed” statement despite PR CI remaining subsequent. Make `Done when` reflect the revised checks and distinguish completed local verification from pending CI.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed: startup package with `MAKEFLAGS=k` and with `MAKEFLAGS=`, both uncached.
   - Mutation confirmed: without `-S`, the regression fails with “later recipe ran.”
   - The broader weave run produced no result before I interrupted it; neither that run nor full-suite/PR CI success is independently established here.

6. **Architectural notes**
   - **ARCH-DRY: pass.** One production invocation serves all owner layers; the existing regression fixture is reused.
   - **ARCH-PURE: pass.** The change stays within the thin IO boundary and retains injected filesystem/process dependencies.
   - **ARCH-PURPOSE: pass for implementation.** The actual inherited-keep-going failure is fixed and independently reproduced without the fix.
   - No Core concepts table or new architectural/user-facing surface was introduced; README and atlas updates are unnecessary for this correction.

7. **Plan revision recommendation**
   - Append a dated `## Revisions` entry recording reconciliation of `Done when` with the revised acceptance contract, and accurately identify which checks remain pending.

```findings
findings:
  - id: new
    severity: Important
    family: acceptance-contract-consistency
    title: |
      Synchronize Done when with the revised acceptance contract
    detail: |
      workshop/issues/000358-weave-startup-make-tools-inherits-makeflags-k-from-an-outer-make-k-so-a-failing-tools-recipe-no-longer-stops-later-ones.md:48 still requires make -k test, while line 69 replaces it with all weave packages, both startup environments, and green PR CI. Line 70 also calls revised acceptance passed while leaving CI subsequent. These are all instances in this window: reconcile the acceptance sections and distinguish completed local evidence from pending CI.
```

---

## Re-review — 2026-10-13T19:13:33-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 358 — weave: startup make tools inherits MAKEFLAGS=k from an outer make -k, so a failing tools recipe no longer stops later ones |
| repo | ariadne |
| issue file | workshop/issues/000358-weave-startup-make-tools-inherits-makeflags-k-from-an-outer-make-k-so-a-failing-tools-recipe-no-longer-stops-later-ones.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5b513d43608477ba059418ed9fd84e6068fc705..2bc43b0cacc6bb967a0dd121341d8843ebb81d9d |
| command | sdlc close --issue 358 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T19:13:33-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The fix restores fail-stop behavior under inherited `MAKEFLAGS=k`, and the regression detects its removal. BR-1 is addressed: the current acceptance criteria distinguish completed local verification from pending PR CI. No new findings.

1. **Strengths**
   - `cmd/weave/internal/startup/tools.go:36` applies `-S` at the shared process boundary while preserving other environment settings.
   - `tools_test.go:148` exercises real Make in both environments and checks error propagation plus suppression of later recipes.
   - Issue #358 lines 48–51 and 72–75 consistently describe the revised acceptance contract.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Startup package passed uncached with both `MAKEFLAGS=k` and empty `MAKEFLAGS`.
   - Removing only `-S` in a temporary copy caused the inherited-keep-going regression to fail; the clean case passed. The copy was removed.
   - The broader weave run produced no results before interruption; its recorded success was not independently confirmed. PR CI remains required before landing.

6. **Architecture**
   - **ARCH-DRY: pass.** One shared invocation covers owner layers.
   - **ARCH-PURE: pass.** The change remains within the existing injected IO boundary.
   - **ARCH-PURPOSE: pass.** The regression demonstrates the actual failure and its correction.
   - No new architectural or user-facing surface requires atlas/README changes. No Core concepts table applies.

7. **Plan revisions:** None needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Issue #358:48–51 replaces the old full-suite criterion and explicitly marks PR CI pending at local close; lines 72–75 consistently distinguish local verification from subsequent CI. The pinned before/after diff confirms this prose-only correction.
```
