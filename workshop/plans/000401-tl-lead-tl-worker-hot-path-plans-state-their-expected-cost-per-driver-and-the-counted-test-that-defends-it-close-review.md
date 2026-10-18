# Boundary Review — ariadne#401 (whole-issue close)

| field | value |
|-------|-------|
| issue | 401 — tl-lead/tl-worker: hot-path plans state their expected cost per driver and the counted test that defends it |
| repo | ariadne |
| issue file | workshop/issues/000401-tl-lead-tl-worker-hot-path-plans-state-their-expected-cost-per-driver-and-the-counted-test-that-defends-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa939386228ed537c2d46fc524d9f488943b8eca..649edd3f741c7747fc505fca35d4849fc47cea40 |
| command | sdlc close --issue 401 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T20:32:47-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fulfills the issue’s Spec and Done when. Both help surfaces receive the same hot-path planning contract, and regression checks confirm the phrase guards fail without it. No blocking findings.

```findings
findings: []
```

1. **Strengths**
   - `pkg/tlrules/planning.md:20` requires expected operations for every growth driver and counted tests at N and 2N.
   - Both binaries derive their instructions through `tlrules.ComposeHelp`, preserving one source.
   - The example’s triangular counts, 36 and 136, correctly exceed the stated linear bounds.
   - `atlas/workflow/tl-skills.md:41` documents the new rule.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` passed. In a disposable copy of pinned HEAD, restoring only the base planning text made all six new guards fail for both commands and both invocation forms. The disposable copy was removed.

6. **Architectural notes:** **ARCH-DRY: pass**—both consumers derive from shared text. **ARCH-PURE: pass**—no new IO or business logic. **ARCH-PURPOSE: pass**—the rule reaches both intended roles. Atlas is updated; no new command, flag, configuration, or usage step requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f3d1d67d-b6bf-4a67-a1d1-4717559691d5","repository":"github.com/xianxu/ariadne","issue_id":"000401","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c2a11751b7e1b41f","reviewed_head":"649edd3f741c7747fc505fca35d4849fc47cea40","slot":"ariadne:1","requested_at":"2026-10-18T03:31:57.269985Z"},"phase":"finished","pid":31513,"started_at":"2026-10-18T03:31:57.527243Z","finished_at":"2026-10-18T03:32:46.403048Z","session_id":"01a14d10-ef5f-71b3-8367-bba54c1e013a","outcome":"returned"}] -->
