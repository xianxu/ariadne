---
gate: boundary-review
issue: 310
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T09:52:15-07:00"
      dispatched: "2026-10-11T09:45:16-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
          detail: AllowedTools is Read,Grep,Glob,Bash under dontAsk, so any Bash command runs. For ContainNone judges (plan-quality, sdlc judge) that write reaches the operator tree. Either narrow Bash to read-only patterns or reword close.md:376-384 and sdlc-binary.md:996-1000.
          family: read-only-claim-exceeds-enforcement
          round: 1
        - id: BR-2
          severity: Minor
          title: The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
          family: warning-scope-too-broad
          round: 1
        - id: BR-3
          severity: Minor
          title: Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
          family: plan-claims-undelivered-item
          round: 1
        - id: BR-4
          severity: Minor
          title: Validate's "reviewer_pairs is read only with" wording can be misread as read-only
          family: error-message-clarity
          round: 1
        - id: BR-5
          severity: Minor
          title: 'cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310''s scope)'
          family: read-only-claim-exceeds-enforcement
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-11T09:55:58-07:00"
      dispatched: "2026-10-11T09:54:42-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: 'close.md REVIEWER and atlas sdlc-binary.md were reworded honestly, but the same range added "judges always run read-only" to the --sandbox flag usage and MarkDeprecated text (judge.go:68-69, changecode.go:104-105) and helptext judge.md:70 / change-code.md:185, which are exactly the unfenced ContainNone judges whose claude Bash can write the operator tree; milestone-close.md:44 also says "read-only". 3rd instance of read-only-claim-exceeds-enforcement. Rule: no user-facing text calls a judge read-only without the enforced scope (Edit/Write refused, no auto-approve; shell writes bounded only by checkout/fence). Sweep every such string.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: milestoneclose.go:676 gates the warning on runtime.GOOS == "darwin"; TestBoundaryReviewOtherPolicySwapsReviewer asserts warned == (codex and darwin).
          round: 2
        - id: BR-3
          disposition: addressed
          note: The plan checkbox now says openshell-sandbox describes the main agents and is unchanged; no atlas claim remains for that file.
          round: 2
        - id: BR-4
          disposition: addressed
          note: agent_resolver.go:99 now reads "reviewer_pairs applies only to boundary_reviewer \"explicit\"".
          round: 2
        - id: BR-5
          disposition: addressed
          note: doc-review/review.go:165 now uses dontAsk; TestBuildArgs_ReadOnly asserts --permission-mode dontAsk and fails without the fix.
          round: 2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 3
      timestamp: "2026-10-11T10:05:13-07:00"
      dispatched: "2026-10-11T10:00:30-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: close.md:376-384, sdlc-binary.md:996-1000, judge.md, change-code.md, milestone-close.md and the --sandbox strings now state Edit/Write refused, Bash unrestricted, fence-only containment; matches dispatch.go agentArgs.
          round: 3
      findings:
        - id: BR-6
          severity: Minor
          title: prompts.go AllowedTools comment still claims judges are read-only by construction
          detail: '3rd finding in this family. Rule: any judge read-only claim names its enforcing mechanism (tool list stops Edit/Write, agent sandbox stops codex, fence stops Bash) and claims no more. This comment (prompts.go:186-192) is the one remaining instance under cmd/, atlas/ and README.md.'
          family: read-only-claim-exceeds-enforcement
          round: 3
        - id: BR-7
          severity: Minor
          title: doc-review header still says the judge contract uses --full-auto
          detail: 'cmd/doc-review/review.go:13-17 justifies the duplicate AgentCLI by a --full-auto judge contract that #310 removed; update the rationale.'
          family: stale-rationale-after-contract-change
          round: 3
      recipe: milestone-review
      isolation: checkout
      reviewed: c707d179591146e36e6e6302c63195233d9c7478
      blocked: false
    - "n": 4
      timestamp: "2026-10-11T10:07:27-07:00"
      dispatched: "2026-10-11T10:06:44-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: 'prompts.go:186-192 now names tool list + dontAsk for Edit/Write and says Bash is restricted only by the #204 fence; matches dispatch.go:154-190; sweep of cmd/ atlas/ README.md finds no other judge read-only claim.'
          round: 4
        - id: BR-7
          disposition: addressed
          note: review.go:13-18 rationale now cites no Bash (review.go:165), codex -o capture (review.go:152-154), no containment levels; --full-auto reference removed.
          round: 4
      recipe: fix-review
      isolation: checkout
      reviewed: 96ccbb2345825ccbd5fde1782076759fdd3879e6
      blocked: false
---

# Gate ledger — ariadne#310 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T09:52:15-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `read-only-claim-exceeds-enforcement` Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
  AllowedTools is Read,Grep,Glob,Bash under dontAsk, so any Bash command runs. For ContainNone judges (plan-quality, sdlc judge) that write reaches the operator tree. Either narrow Bash to read-only patterns or reword close.md:376-384 and sdlc-binary.md:996-1000.
- **BR-2** [Minor] `warning-scope-too-broad` The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
- **BR-3** [Minor] `plan-claims-undelivered-item` Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
- **BR-4** [Minor] `error-message-clarity` Validate's "reviewer_pairs is read only with" wording can be misread as read-only
- **BR-5** [Minor] `read-only-claim-exceeds-enforcement` cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310's scope)

## Round 2 — 2026-10-11T09:55:58-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — not-addressed — close.md REVIEWER and atlas sdlc-binary.md were reworded honestly, but the same range added "judges always run read-only" to the --sandbox flag usage and MarkDeprecated text (judge.go:68-69, changecode.go:104-105) and helptext judge.md:70 / change-code.md:185, which are exactly the unfenced ContainNone judges whose claude Bash can write the operator tree; milestone-close.md:44 also says "read-only". 3rd instance of read-only-claim-exceeds-enforcement. Rule: no user-facing text calls a judge read-only without the enforced scope (Edit/Write refused, no auto-approve; shell writes bounded only by checkout/fence). Sweep every such string.
- BR-2 — addressed — milestoneclose.go:676 gates the warning on runtime.GOOS == "darwin"; TestBoundaryReviewOtherPolicySwapsReviewer asserts warned == (codex and darwin).
- BR-3 — addressed — The plan checkbox now says openshell-sandbox describes the main agents and is unchanged; no atlas claim remains for that file.
- BR-4 — addressed — agent_resolver.go:99 now reads "reviewer_pairs applies only to boundary_reviewer \"explicit\"".
- BR-5 — addressed — doc-review/review.go:165 now uses dontAsk; TestBuildArgs_ReadOnly asserts --permission-mode dontAsk and fails without the fix.

## Round 3 — 2026-10-11T10:05:13-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — close.md:376-384, sdlc-binary.md:996-1000, judge.md, change-code.md, milestone-close.md and the --sandbox strings now state Edit/Write refused, Bash unrestricted, fence-only containment; matches dispatch.go agentArgs.

### Raised

- **BR-6** [Minor] `read-only-claim-exceeds-enforcement` prompts.go AllowedTools comment still claims judges are read-only by construction
  3rd finding in this family. Rule: any judge read-only claim names its enforcing mechanism (tool list stops Edit/Write, agent sandbox stops codex, fence stops Bash) and claims no more. This comment (prompts.go:186-192) is the one remaining instance under cmd/, atlas/ and README.md.
- **BR-7** [Minor] `stale-rationale-after-contract-change` doc-review header still says the judge contract uses --full-auto
  cmd/doc-review/review.go:13-17 justifies the duplicate AgentCLI by a --full-auto judge contract that #310 removed; update the rationale.

## Round 4 — 2026-10-11T10:07:27-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — prompts.go:186-192 now names tool list + dontAsk for Edit/Write and says Bash is restricted only by the #204 fence; matches dispatch.go:154-190; sweep of cmd/ atlas/ README.md finds no other judge read-only claim.
- BR-7 — addressed — review.go:13-18 rationale now cites no Bash (review.go:165), codex -o capture (review.go:152-154), no containment levels; --full-auto reference removed.

## Open findings

(none — every finding has been disposed)
