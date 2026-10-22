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
