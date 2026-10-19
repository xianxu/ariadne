# Boundary Review — ariadne#471 (whole-issue close)

| field | value |
|-------|-------|
| issue | 471 — abandon: --issue N acted on the current (other issue's) branch and deleted it on origin |
| repo | ariadne |
| issue file | workshop/issues/000471-abandon-issue-n-acted-on-the-current-other-issue-s-branch-and-deleted-it-on-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | c1fbcfe858e5d72512ee2c63555c7c5d15b7fb9b..c2d924406b8704c12eb36f01b723865f5406acab |
| command | sdlc close --issue 471 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T16:36:21-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The branch guards address the reported wrong-checkout incident, and the focused tests pass. However, local branch deletion still violates the Spec’s compare-and-swap requirement: a deterministic regression reproduced deletion of an unarchived tip while abandon returned success. README also needs the changed usage requirements.

```findings
findings:
  - id: new
    severity: Critical
    family: destructive-ref-cas
    title: |
      Local branch deletion can remove a tip that was never archived
    detail: |
      cmd/sdlc/abandon.go:422 checks the local tip before remote operations and checkout, but line 475 later deletes it using unconditional git branch -D. A temporary regression moved that ref to an unarchived commit in the post-checkout hook; abandon deleted it and returned nil. Use an atomic deletion conditioned on rec.Head, and commit a deterministic regression covering movement between observation and deletion. This violates the Spec and ARCH-ORDER.
  - id: new
    severity: Important
    family: user-surface-docs
    title: |
      README update appears missing for changed branch requirements
    detail: |
      README.md:84 and :110 describe attachment and abandonment without the new checkout requirements, and README has no change in the pinned range. Document canonical-branch admission for started work, resting-branch attachment for open issues, and terminal abandon retry requirements. Atlas and abandon help updates do not satisfy the separate README gate.
```

1. **Strengths**

   - `requireIssueBranch` provides a shared, pure admission check with direct table tests (`cmd/sdlc/trackerbranch.go:11`).
   - Wrong-branch abandon tests assert preservation of local refs, remote refs, archive, cards, and main (`cmd/sdlc/abandon_test.go:99`).
   - Remote retirement verifies the archive and uses explicit leases for advancement and deletion (`cmd/sdlc/abandon.go:415–461`).
   - Renamed-close recovery checks exact tip equality and includes a different-head negative case.

2. **Critical findings**

   `cmd/sdlc/abandon.go:475`: replace unconditional local deletion with an expected-old-OID operation, such as `git update-ref -d <ref> <rec.Head>`. Preserve a moved branch and return a visible error. The earlier tip check cannot authorize a later unconditional deletion.

3. **Important findings**

   `README.md:84,110`: update the existing attachment and abandon instructions to explain the new branch requirements.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused abandon, admission, close, milestone-close, attach, and renamed-close tests passed: `317.931s`.
   - Temporary overlay test `TestReview471LocalDeleteCAS` failed as expected: the unarchived local ref disappeared and abandon returned `nil`.
   - `git diff --check` passed. Tracked files remained unchanged.
   - The full suite was not independently rerun. There are no prior findings to dispose.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared admission and existing remote helpers are reused.
   - **ARCH-PURE — pass:** strict branch admission is pure and directly tested; close’s ref resolution is integration work.
   - **ARCH-PURPOSE — flag:** wrong-checkout protection is delivered, but the explicit deletion invariant remains incomplete.
   - **ARCH-MOCK — pass:** disposable repositories and bare remotes exercise real stateful Git behavior.
   - **ARCH-CONSTRAINTS — pass:** new operations are sequential and use the existing bounded Git execution layer.
   - **ARCH-SECURE — pass:** unsafe persisted branch names and failed probes produce errors.
   - **ARCH-ORDER — flag:** local deletion relies on stale evidence; the reproduced ordering deletes unarchived work.
   - **ARCH-FUNERAL — pass:** no new durable artifact family is introduced.
   - **ARCH-EVOLVE — pass:** persisted formats are unchanged.

7. **Plan revision recommendations**

   Append a `## Revisions` entry requiring atomic expected-tip deletion locally, with a deterministic concurrent-movement regression. Add README completion to the documentation task. The existing Core concepts entries otherwise map to implemented entities.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"74c0d0d1-139e-4362-a673-7b0f40943a21","repository":"github.com/xianxu/ariadne","issue_id":"000471","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/762119f53ffa786e","reviewed_head":"c2d924406b8704c12eb36f01b723865f5406acab","slot":"ariadne:1","requested_at":"2026-10-19T23:30:21.016538Z"},"phase":"finished","pid":9072,"started_at":"2026-10-19T23:30:21.254625Z","finished_at":"2026-10-19T23:36:20.378293Z","session_id":"01a15680-758c-7cd0-b1bf-137650e44e1a","outcome":"returned"}] -->
