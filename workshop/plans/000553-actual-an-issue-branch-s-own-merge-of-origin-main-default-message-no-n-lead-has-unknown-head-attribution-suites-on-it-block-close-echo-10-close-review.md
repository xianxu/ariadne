# Boundary Review — ariadne-553#553 (whole-issue close)

| field | value |
|-------|-------|
| issue | 553 — actual: an issue branch's own merge of origin/main (default message, no #N lead) has unknown head attribution; suites on it block close (echo#10) |
| repo | ariadne-553 |
| issue file | workshop/issues/000553-actual-an-issue-branch-s-own-merge-of-origin-main-default-message-no-n-lead-has-unknown-head-attribution-suites-on-it-block-close-echo-10.md |
| boundary | whole-issue close |
| milestone | — |
| window | e4bd907ae47d7a94b31b6067df0976f96eee3ade..db9a3d2a43dfdc6f0743022b47dbaae326b13d07 |
| command | sdlc close --issue 553 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T07:05:27-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The default-message merge fix works, with meaningful real-Git and recorded-fixture coverage. One classifier branch still bypasses structural attribution for two message values, contradicting the planned arbitrary-message behavior and leaving valid suites unable to support close.

```findings
findings:
  - id: new
    severity: Critical
    family: message-independent-structural-attribution
    title: |
      Bookkeeping subjects bypass proven structural integration attribution
    detail: |
      cmd/sdlc/actualqueue_head.go:192–196 returns unknown before the new structural fallback for both instances in this family: "issue-sync: update issues" and "archive completed issues to history". A merge with an issue-attributable first parent and all side parents on pinned main therefore remains unknown solely because of its message (ARCH-PURPOSE). Added real-Git cases fail for both subjects; allowing the unsuccessful bookkeeping match to fall through makes the complete structural-control matrix pass. Preserve single-parent bookkeeping recognition and train precedence, allow structurally valid merges through, and retain regression cases for both subjects.
```

1. **Strengths**

   - Structural proof uses pinned main ancestry and inherits only first-parent membership.
   - Real-Git controls cover foreign parents, octopus merges, repeated integrations, ref movement, replacements, grafts, and shallow boundaries.
   - Removing the structural fallback makes both the recorded echo fixture and default-message merge regression fail.
   - The atlas documents the new attribution path; no new commands, flags, or configuration require README changes.

2. **Critical findings**

   - The finding above blocks approval. Fix the early return at [actualqueue_head.go:192](/Volumes/TestScratch/review-checkouts/da68810f4a8591fd/cmd/sdlc/actualqueue_head.go:192) and cover both subject values.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Targeted queue-head, owned-actual, and Git-object tests passed. Temporary Go overlays demonstrated both failures and confirmed the proposed correction passes the structural-control matrix. The checkout remains unchanged. The owner-run live measurement and queued full-suite result were not independently verified here.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses the immutable reader and classifier/reducer.
   - **ARCH-PURE — pass:** Git ancestry evidence stays in the IO shell.
   - **ARCH-PURPOSE — flag:** two message-dependent exclusions remain.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording the bookkeeping-subject fall-through correction and regressions for both subjects, preserving malformed-train refusal and foreign-parent controls.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d44de121-8df4-4df6-a18c-62e8623b9c94","repository":"github.com/xianxu/ariadne","issue_id":"000553","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/da68810f4a8591fd","reviewed_head":"db9a3d2a43dfdc6f0743022b47dbaae326b13d07","slot":"ariadne:5","requested_at":"2026-10-22T14:03:05.841506Z"},"phase":"finished","pid":82990,"started_at":"2026-10-22T14:03:06.011797Z","finished_at":"2026-10-22T14:05:26.673335Z","session_id":"01a163ec-3379-7142-b581-650feb0e101e","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T07:16:32-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 553 — actual: an issue branch's own merge of origin/main (default message, no #N lead) has unknown head attribution; suites on it block close (echo#10) |
| repo | ariadne-553 |
| issue file | workshop/issues/000553-actual-an-issue-branch-s-own-merge-of-origin-main-default-message-no-n-lead-has-unknown-head-attribution-suites-on-it-block-close-echo-10.md |
| boundary | whole-issue close |
| milestone | — |
| window | e4bd907ae47d7a94b31b6067df0976f96eee3ade..954dea6bc67eecfd312bd92ccefa1affeb0ac827 |
| command | sdlc close --issue 553 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T07:16:32-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The default-message integration works, and BR-1’s bookkeeping cases are corrected with meaningful regression coverage. A remaining mismatch between the reader’s and classifier’s repository context still prevents structural attribution for qualified invalid issue leads, violating the explicit Done-when contract.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Both bookkeeping merge regressions pass on pinned HEAD and fail with the prior parser restored through a temporary Go overlay. Single-parent bookkeeping, foreign-parent, and train controls also pass.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: message-independent-structural-attribution
    title: |
      Qualified invalid leads still bypass structural ancestry proof
    detail: |
      cmd/sdlc/actualqueue_head_git.go:40 calls parseQueueHead(fact, ""), while membership classification uses the actual repository name. For ariadne#0, the reader sees definite foreign authority and skips ancestry proof; the classifier sees an invalid local lead and returns unknown with MainParents unset. Real-Git overlay cases fail for all three tested forms: ariadne#0 alone, ariadne#0 followed by #10, and #10 followed by ariadne#0. The affected class includes zero-padded zero IDs and qualified invalid refs anywhere in a recognized lead list, for any self-repository name. This is the 2nd finding in family message-independent-structural-attribution. Earlier rounds fixed instances. Do NOT fix only these spellings: enforce the rule that proof selection and membership classification use identical authority context, preserving definite authority and malformed-train precedence (ARCH-PURPOSE). Retain regression controls for this enumeration.
```

1. **Strengths**

   - Structural attribution inherits only first-parent membership; main’s issue membership remains excluded.
   - Real-Git controls cover repeated integrations, octopus merges, foreign parents, ref movement, history overrides, cancellation, and cache limits.
   - BR-1’s correction is reachable and regression-sensitive.
   - The atlas documents the new flow. No new command, flag, or configuration requires a README update.

2. **Critical findings**

   - The finding above blocks shipping. At [actualqueue_head_git.go:40](/Volumes/TestScratch/review-checkouts/02338d9c9a8acc8f/cmd/sdlc/actualqueue_head_git.go:40), supply the same repository context used by the membership classifier when deciding whether structural proof is needed.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Targeted queue-head, owned-actual, Git-object, and pure parser/reducer tests passed. Restoring the prior parser made both BR-1 bookkeeping regressions fail. Three added qualified-lead cases failed on pinned HEAD; supplying consistent repository context made the expanded matrix pass.

   Temporary overlays were removed; the checkout remains clean. The owner-run live measurement is recorded in the tracker but was not independently reproduced here.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses the existing authority parser and reducer.
   - **ARCH-PURE — pass:** ancestry IO remains in the injected reader.
   - **ARCH-PURPOSE — flag:** qualified invalid leads still violate message-independent fallback.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

7. **Plan revision recommendations**

   Add a `## Revisions` entry requiring identical authority context across proof selection and classification, with qualified-zero lead controls alone and in both list positions. Preserve foreign-parent and train-precedence controls.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5af370ed-a007-4239-8c71-db0b25d9145e","repository":"github.com/xianxu/ariadne","issue_id":"000553","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/02338d9c9a8acc8f","reviewed_head":"954dea6bc67eecfd312bd92ccefa1affeb0ac827","slot":"ariadne:5","requested_at":"2026-10-22T14:14:25.581358Z"},"phase":"finished","pid":45058,"started_at":"2026-10-22T14:14:25.682028Z","finished_at":"2026-10-22T14:16:31.417191Z","session_id":"01a163f6-91ab-7410-81aa-a8a74d4b6613","outcome":"returned"}] -->
