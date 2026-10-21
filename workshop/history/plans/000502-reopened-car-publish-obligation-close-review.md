# Boundary Review — ariadne#502 (whole-issue close)

| field | value |
|-------|-------|
| issue | 502 — Reopened issue can silently remove a waiting car publish obligation |
| repo | ariadne |
| issue file | workshop/issues/000502-reopened-car-publish-obligation.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7be44ad5cd268115aff52056f06e869c353732cb..082414664f1091e790a874ffdac3d2f39dd84e6c |
| command | sdlc close --issue 502 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T07:15:19-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills the documented publish-only policy. Both publication paths refuse carried reopened completions before review-anchor selection, while preserving ownership rules, issue-free publication, and settlement/archive eligibility. No findings.

Done-when: current — it covers both publish paths, named recovery, ownership controls, multiple carried issues, and settlement behavior.

```findings
{}
```

1. **Strengths**

   - `cmd/sdlc/trackercompletion.go:261` extracts the existing ownership predicate without changing its repository, base-exclusion, or current-token semantics.
   - `cmd/sdlc/landingcar_records_test.go:250` exercises production readiness through ready → external reopen → refusal → re-close → ready, including rebased evidence and independently published details.
   - `cmd/sdlc/trackerpublish_eligibility_test.go:97` protects legitimate empty selections and prevents another eligible issue from hiding reopened work.
   - `atlas/workflow/sdlc-binary.md:465` documents binding authority, refusal, and recovery consistently with the implementation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Both targeted runs passed: new eligibility/readiness regressions in **50.989s**, and compatibility checks in **71.494s**. These covered generation sharing, supersession, rebase/squash behavior, cancellation, completion, and archive behavior. `git diff --check` passed. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY: pass** — ownership has one shared implementation.
   - **ARCH-PURE: pass** — status eligibility is directly tested without IO; tracker/Git operations remain in the integration shell.
   - **ARCH-PURPOSE: pass** — both publication consumers enforce the guard before anchor selection; settlement/archive retain their selectors.
   - **ARCH-EVOLVE: pass** — no persisted format or rewrite behavior changes.

   The Core concepts entries match the implementation. Atlas coverage is present; no new command, flag, or configuration syntax requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d92ebaa3-68f5-4d9f-91ab-7a8275cbfa91","repository":"github.com/xianxu/ariadne","issue_id":"000502","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/d1f398bb5bfacbaf","reviewed_head":"082414664f1091e790a874ffdac3d2f39dd84e6c","slot":"ariadne:2","requested_at":"2026-10-21T14:12:23.040356Z"},"phase":"finished","pid":8623,"started_at":"2026-10-21T14:12:23.128912Z","finished_at":"2026-10-21T14:15:18.935521Z","session_id":"01a15ece-56e1-7793-9e9b-d1fe6800f0dc","outcome":"returned"}] -->
