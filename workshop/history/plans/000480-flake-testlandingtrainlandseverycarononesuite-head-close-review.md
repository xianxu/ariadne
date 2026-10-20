# Boundary Review — ariadne#480 (whole-issue close)

| field | value |
|-------|-------|
| issue | 480 — flake: TestLandingTrainLandsEveryCarOnOneSuite/head |
| repo | ariadne |
| issue file | workshop/issues/000480-flake-testlandingtrainlandseverycarononesuite-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | eff65960043040be38d0efb3dfb59c95ea1b8723..740b851e5dbb5431f927352b93037c422fbd4666 |
| command | sdlc close --issue 480 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T05:41:35-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change satisfies the revised Spec and Plan: the default caps shards at eight, preserves smaller-host behavior and explicit overrides, and retains suite weight 2. Measurement artifacts support the documented rationale. No blocking defects found.

```findings
{}
```

1. **Strengths**
   - One pure helper owns the default policy (`scripts/test-shard.py:38`).
   - Parser tests exercise actual default selection and explicit override behavior (`scripts/test-shard.test.py:70`).
   - Atlas documents the measured tradeoff and its limits accurately.
   - Verified 54 artifact/payload hashes and #484’s canonical `wontfix` status.

2. **Critical findings:** None.

3. **Important findings:** None. Atlas is updated; no new command, flag, or configuration key requires README changes.

4. **Minor findings:** None.

5. **Test coverage:** Seven Python tests and the Go wrapper pass. Executing the regression against the base parser fails as expected (`32 != 8`). Thirteen pure metrics tests pass. Verified all original train variants passed in each completed measurement and checked the recorded final suite’s digest and successful eight-shard summary; the full suite was not rerun.

6. **Architecture:** ARCH-DRY passes: one policy helper. ARCH-PURE passes: policy and metrics remain independently testable without IO. ARCH-PURPOSE passes against the explicitly revised five-run acceptance. ARCH-EVOLVE passes: no production persisted-format changes. Core-concept entities exist at their documented locations.

7. **Plan revisions:** None required. The existing revisions explicitly supersede the original nine-run protocol and preserve diagnostic uncertainty.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"80682ee6-1fdd-45c0-91e9-11c4ff5fe98d","repository":"github.com/xianxu/ariadne","issue_id":"000480","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/ea794f8fa78aa09c","reviewed_head":"740b851e5dbb5431f927352b93037c422fbd4666","slot":"ariadne:1","requested_at":"2026-10-20T12:40:11.576609Z"},"phase":"finished","pid":68293,"started_at":"2026-10-20T12:40:11.670202Z","finished_at":"2026-10-20T12:41:35.078031Z","session_id":"01a15953-9395-71e1-a10d-61e90172189f","outcome":"returned"}] -->
