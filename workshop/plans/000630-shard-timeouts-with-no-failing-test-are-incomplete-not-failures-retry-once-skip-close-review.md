# Boundary Review — ariadne#630 (whole-issue close)

| field | value |
|-------|-------|
| issue | 630 — Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis |
| repo | ariadne |
| issue file | workshop/issues/000630-shard-timeouts-with-no-failing-test-are-incomplete-not-failures-retry-once-skip.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5e64f147f2af796b27d900400406e0eea954dfd1..4d08de9898e029c2b9bf5abef50195eab892c9c7 |
| command | sdlc close --issue 630 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T18:14:36-07:00 |
| verdict | SHIP |

## Review

Classifier guard mutations all go red so far (BR-11's Typed.Seen guard is now pinned). Re-running three mutations whose patterns didn't apply cleanly.

All 12 claimed classifier guards go red when removed. Next: checking BR-12 (resume-record log selection), BR-7, and BR-8.

Prior findings all check out. Now running the related Go and Python tests at the pinned head.

```verdict
verdict: SHIP
confidence: high
```

All four open Minor findings are fixed at head `4d08de989`, and I'm raising nothing new.

- **BR-11 (the classifier's typed-runner guard wasn't tested):** fixed. The "a typed runner's evidence" case now uses well-formed typed evidence that no earlier branch rejects. Removing the `f.Typed.Seen` guard turns it red, which I confirmed by actually reverting it.
- **The broader rule behind BR-11:** I removed each of the 12 checks in `classifyIncomplete` one at a time (the plan's Revisions entry claims each is pinned by its own test). Every removal turned the named `TestClassifyIncomplete` case red; the exit check also fails fuzz seed 3, as the plan says.
- **BR-12 (log chosen by comparing digests):** fixed. `resolvedSavedRecord` now returns the attempt whose log backs the record, and both callers take `LogPath` from it.
- **BR-7 (fields with no consumers):** fixed. `Running` and `Retried` are now both set and read.
- **BR-8 (wrong baseline message):** fixed. The message now reads "main's baseline is unknown" and is tested.

Test results at the pinned head:
- `go vet` is clean.
- The targeted `cmd/sdlc` tests pass (`Incomplete|Resume|Republish|SuiteAttempt|TestStats|Typed|HostLoad|Flake|LandingTrain`, 503s).
- `scripts/test-shard.test.py` passes.
- `GOOS=freebsd go build ./cmd/sdlc` succeeds.

**Strengths**
1. `classifyIncomplete` (`cmd/sdlc/suiteincomplete.go:45`) is a pure function. Each of its checks now has a case that only that check satisfies, and the fuzz property also covers package-level and process-level reasons (`suiteincomplete_test.go:269-282`).
2. `resolvedSavedRecord` returns the attempt whose log backs the record (`localcheck_resume_saved.go:117-158`). The failed-retry branch is now stated directly instead of being guessed from matching digests.
3. There is one definition of a clean retry (`retryPassed`), and both the run path and the proof path use it. That closes off a retry landing green while its log still names a failure.
4. `suiteIncompleteError.Error()` changes its message for a main baseline and for a second cut-short run, and the tests check both wordings (`suiteincomplete_test.go:412,424,590`).
5. `ThreadPoolExecutor(max(1, len(jobs)))` (`scripts/test-shard.py:134`) fixes a crash when the selection compiles no tests, and it has its own test.

**Critical findings:** none.

**Important findings:** none.

**Minor findings:** none new.

**Test coverage notes**
- I reverted every claimed check and each one went red, so the plan-claimed-test-missing family is now covered as a rule, not just at the sites earlier rounds named.
- The BR-12 change is a refactor with no change in behaviour. The existing assertion that a failed retry's record points at the retry's log (`suiteincomplete_test.go:405-407`) still covers it.

**Architecture**
| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass | `retryPassed` and `classifyIncomplete` are each the only definition, shared by the run and proof paths. |
| ARCH-PURE | pass | The classifier, retry outcome and retry commands are pure; host load and the main affected config come through injected `localRunDeps`. |
| ARCH-PURPOSE | pass | The retry runs only the selected packages, through `SDLC_AFFECTED_GO_PKGS` in `test-shard.py`. |
| ARCH-MOCK | pass | The live Go-alarm conformance test covers it (disposed under BR-6). |
| ARCH-CONSTRAINTS | pass | The retry gets the check's own weight and timeout, and exactly one retry is allowed by structure. |
| ARCH-SECURE | pass | The retry's selection is re-derived from the kept log, and `select_packages` refuses packages outside the module. |
| ARCH-ORDER | pass | The retry goes through the existing resume state machine (`resumeAdvance`); the second-incomplete exit is explicit. |
| ARCH-FUNERAL | pass | There is no new kind of file; v5 attempts use the existing retention. |
| ARCH-EVOLVE | pass | `load_average` is an additive field on v5 only. The older-reader test (reader version 4) and strict v5 decoding (no typed identities or known flakes) are both tested. |

**Architectural notes for upcoming work:** v5 attempts and links are now shared by the affected and incomplete-retry paths. Any future change to the version checks should repeat the `grep -n 'Version [!=]= [0-9]'` sweep recorded in the plan.

**Plan revision recommendations:** none. The close-review mutation table in the Revisions entry matches what I measured.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Running is consumed by incompleteRun.String (suiteincomplete.go:91) and pinned at suiteincomplete_test.go:78; Retried is set at localcheck_resume.go:295, consumed in Error(), asserted in TestRunLocalCheckRecordsIncompleteTimeout and the baseline test.
  - id: BR-8
    disposition: addressed
    note: |
      Baseline consequence now reads "main's baseline is unknown, so no failure is attributed to main or to the change" (suiteincomplete.go:138-140), asserted at suiteincomplete_test.go:412,424,590.
  - id: BR-11
    disposition: addressed
    note: |
      Typed fixture is now well-formed (no Unattributed); reverting f.Typed.Seen turns the typed-runner case red. All 12 Revisions-claimed guards were reverted one at a time and each went red on its named case.
  - id: BR-12
    disposition: addressed
    note: |
      resolvedSavedRecord returns the log-owning attempt (localcheck_resume_saved.go:117-158); both callers use it (localcheck_resume.go:98-102, resume_saved.go:245-249). No digest comparison remains. Refactor only; the retry-log assertion at suiteincomplete_test.go:405 still covers it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"89bfb3ee-aa6b-4388-9e17-02ad0c955e36","repository":"github.com/xianxu/ariadne","issue_id":"000630","recipe":"milestone-review","round":5,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/32b23db7796b0c40","reviewed_head":"4d08de9898e029c2b9bf5abef50195eab892c9c7","slot":"ariadne:5","requested_at":"2026-10-26T01:03:30.174425Z"},"phase":"finished","pid":76753,"started_at":"2026-10-26T01:03:30.27495Z","finished_at":"2026-10-26T01:14:35.03451Z","session_id":"d513df2c-0cdf-4010-8bfa-7d16c654cd20","outcome":"returned"}] -->
