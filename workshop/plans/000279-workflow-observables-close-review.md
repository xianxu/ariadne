# Boundary Review — ariadne#279 (whole-issue close)

| field | value |
|-------|-------|
| issue | 279 — Expose authoritative workflow observations for agents |
| repo | ariadne |
| issue file | workshop/issues/000279-workflow-observables.md |
| boundary | whole-issue close |
| milestone | — |
| window | 11ce13fafae075a7878bc68aadbe2db135975813..f8780473f93ea87dd1dda0cacdd8a442c63dc97c |
| command | sdlc close --issue 279 |
| reviewer | claude |
| timestamp | 2026-10-02T01:01:41-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All six open findings are fixed in the final code, and I found no new blocking problems. The tests that cover this work pass at HEAD: `go test ./cmd/sdlc/internal/observe/ ./cmd/sdlc/internal/issue/` and `go test ./cmd/sdlc/ -run 'Observe|IssueShow|CommitPathspec'`. The last fix commit (f8780473) adds a rule to `Validate`: a review carries a verdict exactly when its artifact was read (state present) and the boundary is not `plan`. Two new rejection rows test it. That rule matches the only places `review()` in `checkpoints.go` produces values: plan reviews are present with no verdict, and every other present review has a verdict that passed `IsEmitted`. The envelope finding was fixed by narrowing the claim, in both the test comment and the atlas, so it now covers the collector only.

1. **Strengths**
   - **Validated contract.** `observe.Validate` checks every enumerated field and every "set exactly when" invariant (`relation`, `outcome`, `claimant_worktree`, `verdict`). Both `MarshalJSON` and `UnmarshalJSON` call it, and decoding also rejects duplicate and unknown keys (`json.go`). So a malformed observation can't be written or read.
   - **Pure core, thin IO.** `Assemble` and `assembleCheckpoints` only take inputs and return values. All git calls in `cmd/sdlc/observe.go` go through the counted `observeGit` seam, including `workspace.Resolve`, which uses `observeGitReader`.
   - **Shared definitions.** The milestone grammar (`issue.MilestoneTagPattern`), the flow value sets (`flow.ValidKind` / `flow.ValidProvenance`), and `TickedMilestones` are each defined once. Both the existing parsers and the new contract use them.
   - **Failed reads are never reported as absent.** If the details file is missing or can't be parsed, the section becomes `unknown` with a reason. A boundary the issue records as closed but whose artifact is missing is also `unknown`.
   - **One tracker fetch per command.** Plain `issue show` loads issue records, then the deferred observation loads them again. The second call reuses the first through the `recordsScope` cache, so the tracker is fetched only once.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `cmd/sdlc/issue.go:16`: the `observe` import sits inside the standard-library import group. Move it into the group of `github.com/xianxu/...` imports.

5. **Test coverage notes**
   - The BR-17 rows ("present review, no verdict" and "verdict on an absent review") fail if the new check is removed, because both expect the error text "verdict is set exactly when".
   - The golden v1 fixture still passes the stricter `Validate`.

6. **Architecture, principle by principle**
   - **ARCH-DRY: pass.** The grammars and value sets are shared, and artifact paths come from the existing sidecar helpers.
   - **ARCH-PURE: pass.**
   - **ARCH-PURPOSE: pass.** The JSON output, the text view and the documentation are all built from one assembled observation.
   - **ARCH-MOCK: pass.** Git runs against real temporary repositories through the counted seam.
   - **ARCH-CONSTRAINTS: pass.** The scope of the 20-command bound is now stated, and the tracker fetch is deduplicated by the shared records scope.
   - **ARCH-SECURE: pass.** Card, sidecar and ledger inputs are parsed, and a parse failure turns into an `unknown` state that gives the reason.
   - **ARCH-ORDER: pass.** The feature reads once and keeps no state between events.
   - **ARCH-FUNERAL: pass.** The query creates nothing durable; its only write is the remote-tracking ref update that every read-only view's tracker fetch already does.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      observe.go records the rev-parse failure in TrackerRefErr, and assembleTracker emits it as tracker.ref_error in both the present and stale cases.
  - id: BR-5
    disposition: addressed
    note: |
      Validate now has switches over relation and claimant_worktree, a claimant pairing invariant, and rejection rows in TestValidateRejectsEveryUnknownEnum.
  - id: BR-6
    disposition: addressed
    note: |
      The "not observed by this build" placeholder is gone from code, atlas and README; M2 sections are now really assembled.
  - id: BR-17
    disposition: addressed
    note: |
      json.go enforces that a verdict is set exactly when the read is present and the boundary is not plan; two rejection rows fail without the check.
  - id: BR-18
    disposition: addressed
    note: |
      The scope is stated in the observe_test.go comment and in atlas/workflow/issue-tracker.md: the test bounds the collector, and the tracker load is bounded by the shared records scope.
  - id: BR-19
    disposition: addressed
    note: |
      plan.go now declares MilestoneTagPattern above the milestonePlanRE comment, so each comment sits on its own declaration.
findings:
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      issue.go puts the internal observe import inside the standard-library import group
    detail: |
      Move it into the github.com/xianxu imports group to match the rest of the file.
```
