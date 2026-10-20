# Boundary Review — ariadne#472 (whole-issue close)

| field | value |
|-------|-------|
| issue | 472 — merge: issue-only main commits shouldn't cost a ready merge an integration attempt or a CI wait |
| repo | ariadne |
| issue file | workshop/issues/000472-merge-issue-only-main-commits-shouldn-t-cost-a-ready-merge-an-integration-attempt-or-a-ci-wait.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4b4e3f202ff4eefe81bc4354cbdac477929d8fac..2b74d047b6d241597e30adde4819a5a29f50f704 |
| command | sdlc close --issue 472 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T16:42:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the issue’s purpose: issue-only main movement preserves the code integration budget, permits unchanged-head landing with exact CLEAN evidence, and otherwise retains integration and exact-head CI. No blocking defects found. Documentation covers the changed behavior.

Done-when: current — the fixtures verify three issue-only publications preserve the tested head and suite, while code movement still requires charged integration.

```findings
{}
```

1. **Strengths**

   - `cmd/sdlc/mainmovement.go:35` examines commit history, merge-parent differences, and rename sources, preventing code-add/revert sequences from appearing harmless.
   - `cmd/sdlc/integrationstate.go:44` keeps accounting and recovery decisions in the existing transition function, including preservation of a pending integration’s charge.
   - `cmd/sdlc/landing.go:587` reclassifies final main movement and rechecks destination gates before direct landing.
   - Real Git fixtures test observable outcomes: preserved heads, suite reuse, pending new-head CI, conflicts, and duplicate-ID refusal.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Four focused selections passed: classifier/state/readiness (1.180s), new issue-drift landing fixtures (47.167s), existing integration/CI/suite-reuse controls (54.697s), and uncertainty/cleanup recovery (25.095s). `git diff --check` passed; the checkout remained clean. Full-suite and live GitHub conformance runs were not repeated during this review.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — one path policy and shared observer serve all modified integration boundaries. |
   | ARCH-PURE | Pass — classification and transitions have direct tests without IO; Git/API operations remain injected. |
   | ARCH-PURPOSE | Pass — covers direct landing, fallback, registration, replay/push, and recovery. |
   | ARCH-MOCK | Pass — shared GitHub seam, stateful per-head CI fake, real Git fixtures, and an opt-in live conformance check. |
   | ARCH-CONSTRAINTS | Pass — constant Git call count per observation, bounded readiness output, existing IO deadlines. |
   | ARCH-SECURE | Pass — typed identity/OID validation, fixed GitHub authority, conservative unavailable-evidence handling. |
   | ARCH-ORDER | Pass — production changes use the transition function; tests exercise interrupted and concurrent movement sequences. |
   | ARCH-FUNERAL | Pass — introduces no durable runtime artifacts or background actors. |
   | ARCH-EVOLVE | Pass — no persisted format changes; unknown readiness states disable the optimization. |

   Core-concept entities exist at their declared paths and match their classifications. README, help, and atlas updates cover the new surface.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"31b3478b-0725-41e1-980b-20df7fb3ca67","repository":"github.com/xianxu/ariadne","issue_id":"000472","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7d85b22ae7f5018d","reviewed_head":"2b74d047b6d241597e30adde4819a5a29f50f704","slot":"ariadne:3","requested_at":"2026-10-19T23:40:16.75001Z"},"phase":"finished","pid":69884,"started_at":"2026-10-19T23:40:16.839181Z","finished_at":"2026-10-19T23:42:26.312586Z","session_id":"01a15689-8b3d-73b1-857e-19054913aa5d","outcome":"returned"}] -->
