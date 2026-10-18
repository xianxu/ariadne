---
id: 000392
status: codecomplete
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-18
estimate_hours: 4.26
card_mirror: '41ec9efadfe037f6b4f90e3ddedd93e41f7e4fed' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:39:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.14
---

# Weekly test-economy snapshot from local checks and landings

## Problem

#390's test-economy design needs a trustworthy weekly trend before changing evidence tiers or declaring package budgets. `sdlc test-stats` reports recent local attempts, but no durable weekly snapshot joins them to landed issues; older records expire after 30 days. The headline metric, full suites per landed issue, has no stored baseline.

## Spec

Produce one dated, immutable weekly per-repo summary from #376's per-attempt local-check records and authoritative landing records. Show the exact seven-day window, source/coverage counts, suite and queue p50/p90, slowest packages with sample counts, confirmed/suspected/unknown flake classes, and full-suite attempts per landed issue with numerator, denominator, failed/repeated attempts and unknown attribution. Do not infer missing attempts or landings. Store the measurement snapshot in brain under its charter, outside the SDLC issue process; link it from the ops test-economy project. The command/process that builds it belongs in ariadne and becomes a default for every SDLC repo after the ariadne/pair pilot. Brain is the output store, not the implementation repository. Follow #390's approved design and measure before changing budgets.

## Done when

- [ ] A repeatable weekly command produces a dated snapshot in brain and links it from the ops project without duplicating the numbers there.
- [x] The output separates wall and queue, gives p50/p90 and coverage for each repo, names slow packages and flake classes, and calculates full suites per landed issue from exact attempt and landing evidence.
- [x] Tests cover failed/retried attempts, missing old fields, unassignable attempts, zero landings, and stable reruns of the same weekly window; unknown remains unknown.
- [x] A first ariadne/pair snapshot is saved, and the default rollout and retention/backup responsibility are documented.

## Plan

- [x] Design the evidence join and snapshot schema in `workshop/plans/000392-weekly-test-economy-snapshot-plan.md` under #390's approved measurement revision; obtain hub approval before code.
- [x] Build the pure weekly model and exact attempt-to-landing join with fixture-backed red/green tests.
- [x] Add the brain writer and CLI, test immutable reruns and partial evidence, then save the first ariadne/pair snapshot and give its link to the TL for the ops project.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. The two greenfield modules are the weekly metric and exact attribution; the API integration is the bounded Git/tracker/landing evidence reader; the remaining module is the atomic brain writer and CLI. Existing Go libraries and local primitives cover the stack, so no new-library design allowance is needed. The source reports stale calibration, so these hours are provisional.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.5 impl=0.32
item: greenfield-go-module design=0.5 impl=0.32
item: api-integration design=0.6 impl=0.6
item: greenfield-go-module design=0.3 impl=0.32
item: atlas-docs design=0.1 impl=0.08
item: milestone-review design=0.1 impl=0.2
design-buffer: 0.15
total: 4.26
```

## Log


- 2026-10-18: closed — Recorded suite passed on pushed head 3a70481a5df29d2e61a0740945abfb0999a2f81f (12 shards, 1617 sdlc tests, 67 other packages, log sha256:817cd609951a); focused attribution and weekly regressions passed. The final-code 16-repo temporary pilot has SHA-256 be787f22c0d64e866fb9536eb7d09c6ca108d69da5b2ddb16af0fce1a01e11df, and the pair-to-ariadne identical rerun is recorded with hash 450102a407e97db4d9ed2c239990ac8b123b5490a3eb89c128c398b5ee1d2ad3. One intervening pair attempt correctly caused immutable conflict; details and source coverage are in the Log. Per the approved BR-5 dispute, the real brain artifact is published on SHIP and before merge, then its path goes to the TL-owned ops project. These pending publication steps explain the three unticked plan items.; review verdict: SHIP
### 2026-10-15

Filed from #390's operator-approved test-economy design. This precedes #328 and any package-budget declaration.

2026-10-17: The hub directed this slot to build #392 before dependent #402. Existing `sdlc test-stats` and per-attempt records supply suite, queue, package and failure inputs; tracker completion bindings and the landing-line event log supply a confirmed landed issue and admission window. The initial snapshot will preserve coverage unknowns for pre-#376 legacy logs and any missing admission or Git ancestry instead of assigning them by date alone (ARCH-PURPOSE, ARCH-ORDER).

The hub approved the fixed UTC window, strict joined-evidence attribution, separate unassignable count, nullable ratio and versioned JSON schema; it will link the first live brain snapshot from the path reported at landing.

Plan-quality review found that the seven-day physical attempt population and landed-issue cohort were conflated, landing events needed semantic validation, and the test strategy was too procedural. The durable plan now separates the populations, validates the enqueue/landing/card/Git binding before attribution, and names adversarial strategies by function. Minor operating-envelope and lifecycle findings are addressed in the same revision (ARCH-PURPOSE, ARCH-SECURE, ARCH-CONSTRAINTS).

Change-code passed the full-flow plan gate in three rounds; its estimate-quality gate returned INFO about the tight writer allowance and operational verification time. The accepted provisional 4.26 h estimate follows the stale v3.1 calibration. TDD slices for the pure statistics, evidence reader, attribution and brain writer are implemented and focused tests pass. A read-only ariadne pilot for the week ending 2026-10-18 found 83 physical suites, 22 confirmed landings, 12 post-enqueue attributed attempts, 47 pre-enqueue attempts and 24 in-week attempts outside the landing cohort; the latter buckets explain why the two populations differ. The pilot exposed a missing package-only slow list, now added with coverage, and prompted a main-merge inventory check for missing landing events. This pilot wrote only to a temporary brain-shaped directory, not the durable brain repo.

The first full-fleet pilot found pre-landing-line merges in the seven-day UTC week, so ariadne and pair correctly have a null headline ratio with explicit missing-event coverage. It also exposed generic fleet discovery including unrelated sibling repositories; the command now selects only repositories with the vocabulary-declared issue home. The merge's GitHub PR and branch are checked against the landing event before attribution (ARCH-SECURE). The first saved snapshot will be an intentionally partial baseline; a later complete week can establish the numeric trend.

A second pilot revealed brain repos with a retained issue-home directory. The discovery predicate now honors brain's capture-only marker first. The recorded suite on the older pushed commit had not acquired the machine lock; it was cancelled without a test result so the corrected commit can receive the single authoritative run.

Adversarial attribution tests then exposed two partial-evidence paths: an earlier enqueue without its refusal and a legacy attempt on the landed branch. Both now null the ratio. The suite queued on the preceding head was likewise cancelled before it acquired the machine lock (ARCH-SECURE, ARCH-PURPOSE).

The snapshot CLI no longer accepts a repo filter: a filtered pilot could have occupied the immutable fleet-wide date path. The read-only test-stats command still offers per-repo inspection. This keeps one path equal to one denominator (ARCH-ORDER).

The hub approved the evidence-boundary revisions, then required the snapshot headline to show coverage alongside the ratio. A TDD check exposed that the pilot JSON had only a nullable ratio and dispersed warning strings. The repository headline now states covered/observed landed issues, unknown count and top reasons; incomplete source evidence remains a null ratio. The prior queued suite was cancelled before acquisition and will be re-run on the corrected head.

The recorded suite on the headline head completed 12 shards and 67 other packages in 9m19s. Only `TestVerbContextsReachTrackerReads` failed: `weeklyGit`'s nil-context fallback minted a new root context. All real callers pass the command context; the fallback now returns an explicit error before Git starts. The targeted guard, nil-context test and weekly tests pass. The failed attempt and its log remain recorded for the retry.

The corrected head's recorded suite passed all 12 shards, 1610 sdlc tests and 67 other packages in 8m25s. Whole-issue review round 1 returned REWORK with BR-1..BR-6. The hub approved a consolidated evidence-history, prior-admission, legacy-reconciliation and bounded-reader re-plan in the durable plan. The real immutable brain file remains deferred until a SHIP review; a hashed temporary pilot will be the review evidence. No code fix for these findings has begun yet.

The review rework now has permanent red/green regressions: a source-history floor rejects expired or unrecorded early history (including a late-starting structured recorder), every prior admission is checked against the original successful enqueue in either event order, legacy physical attempts appear in reconciliation, and a real Git fixture bounds merge/graph inventory to the requested week. The README and embedded help describe the CLI and nullable coverage. The floor uses the later of retention, first retained landing event, and first retained structured attempt, per the hub's refinement (ARCH-SECURE, ARCH-ORDER, ARCH-CONSTRAINTS).

Review pilot on the corrected code head: `bin/sdlc test-stats snapshot --end 2026-10-18T00:00:00Z --brain-root /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne392-review.i4C0Om` wrote `memory/work/devinfra/test-economy/weekly/2026-10-18.json` in that temporary brain-shaped root. SHA-256: `a61b66845c242a8ef3bb8333faecdd289b45f1d953dedf727746cc5048b16634`. The same rebuilt binary invoked from the pair checkout returned the existing path (stable identical rerun). In the 2026-10-11..18 UTC window, ariadne had 104 retained attempt records, 118 landing events, 83 physical suites and 0 covered of 110 observed landings; pair had 260 records, 218 events, 244 suites and 0 covered of 180 observed. Both ratios are null: the top unknowns are pre-landing-log main merges and issue lifetimes older than retained source history, not a fabricated zero. The real immutable brain publication is deliberately after a SHIP review; the reviewed command on the reviewed head will write it once, and the hub will link that path in the TL-owned project.

The replacement exact-head suite passed 12 shards (1,614 `cmd/sdlc` tests) and 67 other packages in 10m6s; log SHA-256 begins `6c323e2c6e8e`. Close review round 2 disposed BR-1/2/3/4/6 and found BR-7: two different tracker cards could claim one landing identity. The hub approved a cohort-wide uniqueness invariant across issue, ticket, branch, PR and merge, with all claimants unknown on collision. A table-driven regression for each alias in both event orders failed on the old implementation and passed after the shared pre-index check (ARCH-SECURE, ARCH-PURPOSE). BR-5's real brain capture remains ordered after SHIP to avoid freezing an unreviewed measurement; its pilot counter-evidence will be formally disputed.

Round-3 corrected-code pilot: the rebuilt `bin/sdlc` at `84236a9773a8ae221c8e959db71d83fe6e5e3811` wrote `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne392-round3.t6HqPV/memory/work/devinfra/test-economy/weekly/2026-10-18.json` for the same fixed window; SHA-256 `31d3ede0caaa6fada0e0c03cd1e553e0bac05a544fb1e58ed0c6fe055f0f9d8c`. The identical binary run from `/Users/xianxu/workspace/pair` returned the existing path, proving stable bytes across both pilot consumers. The fleet now has 16 discovered SDLC repos; ariadne has 106 retained attempt records, 121 landing events, 83 physical suites, 0 covered of 110 observed landings; pair has 261 attempts, 223 events, 244 physical suites, 0 covered of 180 observed. BR-5 is disputed with the earlier measured pilot and the hub-approved post-SHIP immutable-publication order; this fresh pilot supplies the revised-code evidence.

Before close round 3, a new adversarial test found that an attempt completed before the physical week could match two cohort landings without making the ratio unknown. The former classifier attached the ambiguity warning only inside the week. Per the hub's approved metric re-plan, numerator ownership now classifies every retained attempt independent of the physical window: zero matches stays outside, one counts once, multiple null coverage. The week flag only selects physical reconciliation rows. A property-style matrix over one to four matching landings, both landing orders, and in/out-week attempt times failed on the old code and passes on the corrected classifier. The queued suite for the preceding head was cancelled before acquiring the machine lock; it produced no check result (ARCH-PURPOSE, ARCH-ORDER).

Final-code pilot on rebuilt `bin/sdlc` at `5b0848068d2bff6178e69450c90a41fb9711e452`: `test-stats snapshot --end 2026-10-18T00:00:00Z` wrote `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne392-finalpilot.QGIZQ0/memory/work/devinfra/test-economy/weekly/2026-10-18.json`, SHA-256 `be787f22c0d64e866fb9536eb7d09c6ca108d69da5b2ddb16af0fce1a01e11df`. It covers 16 repos; ariadne has 106 attempt records, 121 events, 83 physical suites and 0/110 observed landings covered; pair had 262 attempts, 223 events, 244 suites and 0/180 covered. A pair-checkout rerun against that path correctly refused changed bytes: a new pair attempt completed between reads (`attempt_records` 262→263), while weekly figures were unchanged. Pair then wrote an independent temp snapshot at `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne392-pairpilot.S7pQT7/memory/work/devinfra/test-economy/weekly/2026-10-18.json`, SHA-256 `450102a407e97db4d9ed2c239990ac8b123b5490a3eb89c128c398b5ee1d2ad3`; ariadne returned that existing path unchanged. Thus the same binary works from both checkouts, unchanged inputs rerun stably, and changed live evidence triggers the intended immutable conflict. The final brain capture remains after SHIP.

Close review round 3 returned SHIP with BR-1/2/3/4/6/7 addressed and BR-5 withdrawn after independent pilot verification. The final pushed-head suite passed 12 shards (1,617 `cmd/sdlc` tests) and 67 other packages in 9m2s, log SHA-256 beginning `817cd609951a`; close measured 2.14 focused hours, excluding 63.18 minutes of recorded queue wait. Staging `49ca659163539266dfba939b46e28ad69d6e8be0` on ariadne:0 built and verified all seven tools; its sdlc wrote a temporary 16-repo snapshot from the pair checkout. Clearing restored clean ariadne:0 main `e5f48d8326eab4ce46e7c8fae2d89ae96498f431` and all seven binaries at that revision. After SHIP, the same command wrote the immutable brain file at `/Users/xianxu/workspace/brain/memory/work/devinfra/test-economy/weekly/2026-10-18.json`, SHA-256 `abda56e1731483bd432aebebb46806544e11222d2680b1b3d72f871f22eae580`. Brain commit `c77fe59b8e900b343a41c3e2df3e3f0b108a4e80` was pushed encrypted to origin/main. The path and commit were sent to the TL for its project link; the file records 16 repos, with ariadne 0/110 and pair 0/180 observed landings fully covered and null headline ratios.
