---
gate: boundary-review
issue: 289
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T16:40:22-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: JudgeCheckout reads an empty issue association on an issue-prefixed branch as a failed probe, conflating "lookup ok, no matching issue" with a lookup error
          detail: AssociateBranchIssue returns empty with nil error for len(matches)!=1, nil lookup, or a backslash branch (issues.go:165-187); slots.go:80-85 then reports unknown with fabricated text "could not be read", and the real lookup error is only in Diagnostics. issueBranchStem also duplicates AssociateBranchIssue's prefix parsing (ARCH-DRY). Record the association outcome on TreeRow (like ClaimsState/ClaimsError), have JudgeCheckout read it, delete issueBranchStem, and add tests for lookup-ok-no-match (ready) and lookup-error (unknown with the real error).
          family: verdict-infers-probe-failure-from-absence
          round: 1
        - id: BR-2
          severity: Minor
          title: cmd/weave/internal/refresh/git.go:104 keeps a third, divergent operation-marker list
          detail: Lacks REBASE_HEAD and BISECT_LOG. gitx is under cmd/sdlc/internal so weave cannot import it; moving OperationMarkers to pkg/ would make it one source repository-wide.
          family: single-source-marker-list
          round: 1
        - id: BR-3
          severity: Minor
          title: gitOperationInProgress no longer uses its root parameter (issuemovedetail.go:71)
          family: dead-parameter
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-02T16:48:15-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: TreeRow.IssuesError recorded in inventory.go:286-303 and read at slots.go:77; issueBranchStem deleted; slots_test.go cases "lookup failed" (unknown with real error) and "lookup ok, no match" (ready) go red on the prior inference.
          round: 2
        - id: BR-2
          disposition: addressed
          note: weave refresh/git.go:105-114 now uses workspace.ActiveOperation over the one pkg/workspace list; refresh tests pass.
          round: 2
        - id: BR-3
          disposition: addressed
          note: gitOperationInProgress(git) drops root (issuemovedetail.go:72); callers at :100 and move.go:249 updated.
          round: 2
      findings:
        - id: BR-4
          severity: Important
          title: Atlas (and issue Log) name gitx.OperationMarkers; the list lives at workspace.OperationMarkers and weave is an unlisted consumer
          detail: atlas/workflow/workspace-branching.md:35 was written before the move to pkg/workspace and not updated by df660572; replace gitx.OperationMarkers with workspace.OperationMarkers (pkg/workspace) everywhere it appears (atlas, issue Log) and add weave refresh to the consumer list.
          family: docs-restate-moved-identifier
          round: 2
        - id: BR-5
          severity: Minor
          title: An ambiguous issue match (len(matches)>1) collapses to no association, so an issue-prefixed branch reads ready
          detail: This is the 2nd finding in the family. The rule is that an answer which is not a clean match/no-match (an error or an ambiguous result) must be recorded on the row as an outcome rather than folded into empty. issues.go:183 should return an error or set IssuesError for ambiguity, so that JudgeCheckout reports probe:issue.
          family: verdict-infers-probe-failure-from-absence
          round: 2
        - id: BR-6
          severity: Minor
          title: gitx/operation_test.go tests only pkg/workspace functions; it lives in gitx because testfix is internal
          family: test-placement
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-02T16:54:54-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: atlas/workflow/workspace-branching.md:35-38 now names workspace.OperationMarkers (pkg/workspace/operation.go) and lists weave refresh; issue Log uses the new name.
          round: 3
        - id: BR-5
          disposition: addressed
          note: 'issues.go:190 returns ErrAmbiguousIssue; inventory.go:286 records it as IssuesError, so JudgeCheckout gives probe:issue; tests: the issues_test wantErr row and the cardinality property test.'
          round: 3
        - id: BR-6
          disposition: addressed
          note: The gitx test was deleted; the same scenario is now in pkg/workspace/operation_test.go with its own git helper, and it passes.
          round: 3
      findings:
        - id: BR-7
          severity: Minor
          title: Issue Plan M1 row still says gitx detector; issuemovedetail comment names removed root param
          detail: '2nd in family. Rule: a commit that moves or renames an identifier, or drops a parameter, greps the issue, the plan, the atlas and nearby comments for the old name in the same commit. Remaining sites: issue line 104 ("`gitx` operation detector"), issuemovedetail.go:70 ("in root''s worktree"), and plan Chunk 1 ("reports rebase-merge"; REBASE_HEAD is reported first).'
          family: docs-restate-moved-identifier
          round: 3
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 4
      timestamp: "2026-10-02T17:09:58-07:00"
      agent: claude
      findings:
        - id: BR-8
          severity: Important
          title: membership.go re-implements weave's construct/deps reader, weaker (follows symlinks, blocks on a FIFO)
          detail: 'readDeclaration (cmd/sdlc/internal/fleet/membership.go:85) duplicates acquire.ReadDeclarations (cmd/weave/internal/acquire/acquire.go:404) and its 1 MiB limit, without the regular-file check or O_NOFOLLOW/O_NONBLOCK, so a FIFO construct/deps hangs the read-only inventory. This is the 2nd finding in family single-source-marker-list (M1 unified the operation markers into pkg/workspace). Rule: any slot-state read both weave and sdlc perform lives once in pkg/; cmd/sdlc never re-implements cmd/weave/internal behavior. Fix the rule: move ReadDeclarations (+limit) to pkg/layergraph, use it from weave refresh/acquire and fleet membership, and search for other duplicates (the substrate walk: weave canonicalizes, DeclaredMembers is lexical).'
          family: single-source-marker-list
          round: 4
        - id: BR-9
          severity: Minor
          title: '"null slots" contract mutation also adds an unknown key, so the non-null check is never what fails'
          detail: slots_test.go TestSlotsContract replaces `"slots":[{` with `"slots":null,"z":[{`; strict decoding rejects "z" regardless, so validateSlots' nil check has no test that isolates it.
          family: contract-rejection-test-isolates-invariant
          round: 4
        - id: BR-10
          severity: Minor
          title: DeclaredMembers builds member paths lexically while rows are canonical
          detail: 'A symlinked member or environment does not match its row: the member reports unknown with a misleading "not a Git checkout" error, and the alias lookup misses (an extra tracker read). Conservative, never wrongly ready.'
          family: lexical-vs-canonical-path-join
          round: 4
        - id: BR-11
          severity: Minor
          title: collectDependencyRows calls readDeclaration/statPath directly but has collect and git injected
          family: io-not-injected
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-10-02T17:17:39-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: 'The reader half is fixed (layergraph.ReadDeclaration shared, weave limits_test pins it). Still open: the substrate walk is neither shared with acquire.Restore nor recorded as separate, and DeclaredMembers dedups and checks environment membership lexically (membership.go:59-75), so a member symlinked to a :0 checkout counts as in-environment and takes that checkout''s verdict, a layout weave''s canonical policy refuses.'
          round: 5
        - id: BR-9
          disposition: addressed
          note: editJSON sets slots to null alone (slots_test.go:195), so validateSlots' nil check (types.go:425) is what fails; the sweep also covered the claims contract mutations.
          round: 5
        - id: BR-10
          disposition: not-addressed
          note: The canonicalization was added (inventory.go:220-235) without a regression test; no fleet test creates a symlinked member or environment, so removing it stays green. The lexical visited set can also yield duplicate members after canonicalization.
          round: 5
        - id: BR-11
          disposition: addressed
          note: read and stat are now parameters of collectDependencyRows (inventory.go:208), passed from CollectInventory.
          round: 5
      findings:
        - id: BR-12
          severity: Minor
          title: pkg/layergraph.Walk still reads construct/deps via OSFS os.ReadFile, beside the new guarded reader
          detail: '3rd finding in family. Rule: every construct/deps read goes through layergraph.ReadDeclaration. walk.go:96 and fs.go:43 follow symlinks, block on a FIFO and are unbounded (used by weave compile and cmd/datatype). Fix: route OSFS declaration reads through ReadDeclaration and add a guard test that no non-test code opens construct/deps another way.'
          family: single-source-marker-list
          round: 5
        - id: BR-13
          severity: Minor
          title: pkg/layergraph.ReadDeclaration has no colocated test; covered only through weave acquire's Restore
          family: test-placement
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-10-02T17:25:29-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: ReadDeclaration lives in pkg/layergraph/read.go; acquire.readDeclarations delegates (acquire.go:412); fleet readDeclaration uses it; placement rule shared via workspace.ValidSlotDependency.
          round: 6
        - id: BR-10
          disposition: addressed
          note: DeclaredMembers canonicalizes via canon and judges declared+canonical paths with ValidSlotDependency; TestDeclaredMembers "symlink to a checkout elsewhere" pins it.
          round: 6
        - id: BR-12
          disposition: not-addressed
          note: walk.go:96 (OSFS os.ReadFile), startplan.go:409 and weave/link.go:68 still read construct/deps directly; no guard test.
          round: 6
        - id: BR-13
          disposition: not-addressed
          note: pkg/layergraph/read_test.go absent; ReadDeclaration still covered only via weave acquire.
          round: 6
      findings:
        - id: BR-14
          severity: Minor
          title: Dependency-clone rows are collected after the pending-diagnostic flush, so their worktree-list failures are dropped
          detail: inventory.go flushes repoStates pending diagnostics before collectDependencyRows runs collectInventoryRepo on clones; a clone whose worktree list fails records a pending diagnostic that is never appended, and its member reports a misleading "not a Git checkout" error. Move the flush after dependency collection and add a test.
          family: late-collector-skips-finalization
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#289 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T16:40:22-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `verdict-infers-probe-failure-from-absence` JudgeCheckout reads an empty issue association on an issue-prefixed branch as a failed probe, conflating "lookup ok, no matching issue" with a lookup error
  AssociateBranchIssue returns empty with nil error for len(matches)!=1, nil lookup, or a backslash branch (issues.go:165-187); slots.go:80-85 then reports unknown with fabricated text "could not be read", and the real lookup error is only in Diagnostics. issueBranchStem also duplicates AssociateBranchIssue's prefix parsing (ARCH-DRY). Record the association outcome on TreeRow (like ClaimsState/ClaimsError), have JudgeCheckout read it, delete issueBranchStem, and add tests for lookup-ok-no-match (ready) and lookup-error (unknown with the real error).
- **BR-2** [Minor] `single-source-marker-list` cmd/weave/internal/refresh/git.go:104 keeps a third, divergent operation-marker list
  Lacks REBASE_HEAD and BISECT_LOG. gitx is under cmd/sdlc/internal so weave cannot import it; moving OperationMarkers to pkg/ would make it one source repository-wide.
- **BR-3** [Minor] `dead-parameter` gitOperationInProgress no longer uses its root parameter (issuemovedetail.go:71)

## Round 2 — 2026-10-02T16:48:15-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — TreeRow.IssuesError recorded in inventory.go:286-303 and read at slots.go:77; issueBranchStem deleted; slots_test.go cases "lookup failed" (unknown with real error) and "lookup ok, no match" (ready) go red on the prior inference.
- BR-2 — addressed — weave refresh/git.go:105-114 now uses workspace.ActiveOperation over the one pkg/workspace list; refresh tests pass.
- BR-3 — addressed — gitOperationInProgress(git) drops root (issuemovedetail.go:72); callers at :100 and move.go:249 updated.

### Raised

- **BR-4** [Important] `docs-restate-moved-identifier` Atlas (and issue Log) name gitx.OperationMarkers; the list lives at workspace.OperationMarkers and weave is an unlisted consumer
  atlas/workflow/workspace-branching.md:35 was written before the move to pkg/workspace and not updated by df660572; replace gitx.OperationMarkers with workspace.OperationMarkers (pkg/workspace) everywhere it appears (atlas, issue Log) and add weave refresh to the consumer list.
- **BR-5** [Minor] `verdict-infers-probe-failure-from-absence` An ambiguous issue match (len(matches)>1) collapses to no association, so an issue-prefixed branch reads ready
  This is the 2nd finding in the family. The rule is that an answer which is not a clean match/no-match (an error or an ambiguous result) must be recorded on the row as an outcome rather than folded into empty. issues.go:183 should return an error or set IssuesError for ambiguity, so that JudgeCheckout reports probe:issue.
- **BR-6** [Minor] `test-placement` gitx/operation_test.go tests only pkg/workspace functions; it lives in gitx because testfix is internal

## Round 3 — 2026-10-02T16:54:54-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — atlas/workflow/workspace-branching.md:35-38 now names workspace.OperationMarkers (pkg/workspace/operation.go) and lists weave refresh; issue Log uses the new name.
- BR-5 — addressed — issues.go:190 returns ErrAmbiguousIssue; inventory.go:286 records it as IssuesError, so JudgeCheckout gives probe:issue; tests: the issues_test wantErr row and the cardinality property test.
- BR-6 — addressed — The gitx test was deleted; the same scenario is now in pkg/workspace/operation_test.go with its own git helper, and it passes.

### Raised

- **BR-7** [Minor] `docs-restate-moved-identifier` Issue Plan M1 row still says gitx detector; issuemovedetail comment names removed root param
  2nd in family. Rule: a commit that moves or renames an identifier, or drops a parameter, greps the issue, the plan, the atlas and nearby comments for the old name in the same commit. Remaining sites: issue line 104 ("`gitx` operation detector"), issuemovedetail.go:70 ("in root's worktree"), and plan Chunk 1 ("reports rebase-merge"; REBASE_HEAD is reported first).

## Round 4 — 2026-10-02T17:09:58-07:00 (claude) — BLOCKED

### Raised

- **BR-8** [Important] `single-source-marker-list` membership.go re-implements weave's construct/deps reader, weaker (follows symlinks, blocks on a FIFO)
  readDeclaration (cmd/sdlc/internal/fleet/membership.go:85) duplicates acquire.ReadDeclarations (cmd/weave/internal/acquire/acquire.go:404) and its 1 MiB limit, without the regular-file check or O_NOFOLLOW/O_NONBLOCK, so a FIFO construct/deps hangs the read-only inventory. This is the 2nd finding in family single-source-marker-list (M1 unified the operation markers into pkg/workspace). Rule: any slot-state read both weave and sdlc perform lives once in pkg/; cmd/sdlc never re-implements cmd/weave/internal behavior. Fix the rule: move ReadDeclarations (+limit) to pkg/layergraph, use it from weave refresh/acquire and fleet membership, and search for other duplicates (the substrate walk: weave canonicalizes, DeclaredMembers is lexical).
- **BR-9** [Minor] `contract-rejection-test-isolates-invariant` "null slots" contract mutation also adds an unknown key, so the non-null check is never what fails
  slots_test.go TestSlotsContract replaces `"slots":[{` with `"slots":null,"z":[{`; strict decoding rejects "z" regardless, so validateSlots' nil check has no test that isolates it.
- **BR-10** [Minor] `lexical-vs-canonical-path-join` DeclaredMembers builds member paths lexically while rows are canonical
  A symlinked member or environment does not match its row: the member reports unknown with a misleading "not a Git checkout" error, and the alias lookup misses (an extra tracker read). Conservative, never wrongly ready.
- **BR-11** [Minor] `io-not-injected` collectDependencyRows calls readDeclaration/statPath directly but has collect and git injected

## Round 5 — 2026-10-02T17:17:39-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — The reader half is fixed (layergraph.ReadDeclaration shared, weave limits_test pins it). Still open: the substrate walk is neither shared with acquire.Restore nor recorded as separate, and DeclaredMembers dedups and checks environment membership lexically (membership.go:59-75), so a member symlinked to a :0 checkout counts as in-environment and takes that checkout's verdict, a layout weave's canonical policy refuses.
- BR-9 — addressed — editJSON sets slots to null alone (slots_test.go:195), so validateSlots' nil check (types.go:425) is what fails; the sweep also covered the claims contract mutations.
- BR-10 — not-addressed — The canonicalization was added (inventory.go:220-235) without a regression test; no fleet test creates a symlinked member or environment, so removing it stays green. The lexical visited set can also yield duplicate members after canonicalization.
- BR-11 — addressed — read and stat are now parameters of collectDependencyRows (inventory.go:208), passed from CollectInventory.

### Raised

- **BR-12** [Minor] `single-source-marker-list` pkg/layergraph.Walk still reads construct/deps via OSFS os.ReadFile, beside the new guarded reader
  3rd finding in family. Rule: every construct/deps read goes through layergraph.ReadDeclaration. walk.go:96 and fs.go:43 follow symlinks, block on a FIFO and are unbounded (used by weave compile and cmd/datatype). Fix: route OSFS declaration reads through ReadDeclaration and add a guard test that no non-test code opens construct/deps another way.
- **BR-13** [Minor] `test-placement` pkg/layergraph.ReadDeclaration has no colocated test; covered only through weave acquire's Restore

## Round 6 — 2026-10-02T17:25:29-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — ReadDeclaration lives in pkg/layergraph/read.go; acquire.readDeclarations delegates (acquire.go:412); fleet readDeclaration uses it; placement rule shared via workspace.ValidSlotDependency.
- BR-10 — addressed — DeclaredMembers canonicalizes via canon and judges declared+canonical paths with ValidSlotDependency; TestDeclaredMembers "symlink to a checkout elsewhere" pins it.
- BR-12 — not-addressed — walk.go:96 (OSFS os.ReadFile), startplan.go:409 and weave/link.go:68 still read construct/deps directly; no guard test.
- BR-13 — not-addressed — pkg/layergraph/read_test.go absent; ReadDeclaration still covered only via weave acquire.

### Raised

- **BR-14** [Minor] `late-collector-skips-finalization` Dependency-clone rows are collected after the pending-diagnostic flush, so their worktree-list failures are dropped
  inventory.go flushes repoStates pending diagnostics before collectDependencyRows runs collectInventoryRepo on clones; a clone whose worktree list fails records a pending diagnostic that is never appended, and its member reports a misleading "not a Git checkout" error. Move the flush after dependency collection and add a test.

## Open findings

- **BR-7** [Minor] `docs-restate-moved-identifier` Issue Plan M1 row still says gitx detector; issuemovedetail comment names removed root param
- **BR-12** [Minor] `single-source-marker-list` pkg/layergraph.Walk still reads construct/deps via OSFS os.ReadFile, beside the new guarded reader
- **BR-13** [Minor] `test-placement` pkg/layergraph.ReadDeclaration has no colocated test; covered only through weave acquire's Restore
- **BR-14** [Minor] `late-collector-skips-finalization` Dependency-clone rows are collected after the pending-diagnostic flush, so their worktree-list failures are dropped
