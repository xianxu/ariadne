---
gate: plan-quality
issue: 194
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-08-20T16:55:07-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: 'Issue file has a corrupt region — a mangled heading and a second, contradictory ## Plan'
          detail: |-
            workshop/issues/000194-review-anchor-commit.md:199-216 holds `## Log` with the
            reason and E is dropped.` plus a stale duplicate ## Done when tail and a second
            ## Plan whose only open item is "Re-plan for the folded scope". Delete 199-216.
            Also: the live ## Plan header at :179 says "Four review boundaries" and the M1
            row says "regardless of M2-M4" when only M1-M3 exist.
          round: 1
        - id: PQ-2
          severity: Important
          title: One issue-wide boundary ledger blows gatestate.Decide's round cap before the close runs
          detail: |-
            Decide computes CapReached from len(l.Rounds) against DefaultRoundCap=3
            (decide.go:14,45). With M1+M2+M3 rounds accumulating in one ledger the whole-issue
            close starts past the cap and demotes Important findings on round 1; OpenFindings
            likewise spans boundaries so an M1 finding blocks M3. gatestate.Round
            (ledger.go:57-75) also has no boundary field, so "rounds tagged with their
            boundary" is a schema change the Core-concepts table omits.
          round: 1
        - id: PQ-3
          severity: Important
          title: Two prior-findings channels with two id namespaces feed one output fence (ARCH-DRY)
          detail: |-
            code-review.md:57-67, embedded into milestone-review.md via CODE_REVIEW_BODY,
            already tells the boundary reviewer to read the plan-gate ledger and re-raise its
            open PQ-* findings. Adding PRIOR_FINDINGS from the BR-* ledger gives the reviewer
            two namespaces and one fence. The plan does not say what a disposition naming an
            unknown id does, nor which carry-forward mechanism survives.
          round: 1
        - id: PQ-4
          severity: Important
          title: Free-text family slugs have no cross-round anchoring, so escalation silently never fires
          detail: |-
            FamilyCounts keys on exact strings and RenderPriorFindings (gatestate/prompt.go:17-70)
            renders no family, so a stateless reviewer writing block-opener-rule then
            block-opener leaves every count at 1 and the escalation instruction — the issue's
            stated purpose (ARCH-PURPOSE) — never triggers. Task 3.5's consistent-slug fixture
            passes by construction. Plan needs: render in-play families with a reuse
            instruction, a normalization step, and a near-miss-slug test.
          round: 1
        - id: PQ-5
          severity: Important
          title: Precedence between the boundary verdict and the new ledger refusal is unstated
          detail: |-
            M2 Task 2.2 Step 4 adds a ledger-derived refusal alongside closeVerdictOutcome
            (close.go:1064-1082). SHIP with an undisposed Important, or REWORK with everything
            disposed — which wins? Also undefined: a boundary protocol miss. Plan-quality warns
            and falls back to the verdict token (changecode.go:495); at the boundary that would
            finalize a close carrying no ledger memory.
          round: 1
        - id: PQ-6
          severity: Minor
          title: Compress M1 Task 1.1's procedural steps and line-numbered fixture inventory
          detail: |-
            Literal test bodies, patch snippets and a cited-line inventory are stale on arrival
            and one entry is already wrong: milestoneclose_test.go:120,134 are renderer tests
            that pass Head "HEAD" as fixture data (:113-115,127-129) and need no change. The
            per-function strategy lines already present are the part worth keeping.
          round: 1
        - id: PQ-7
          severity: Minor
          title: Sidecar count is wrong — 72 files exist repo-wide, not 86, and 67 carry ..HEAD
          detail: |-
            The plan and the Revisions entry both assert 86 archived sidecars all recording
            base..HEAD. Actual: 72 *-review.md repo-wide (70 archived), 67 containing ..HEAD,
            69 distinct window strings. The underlying claim is still true.
          round: 1
        - id: PQ-8
          severity: Minor
          title: finding.cue discovery.glob has no Go consumer — the string-to-list fork is empty
          detail: |-
            FindingModel (pkg/vocab/finding.go:22-28) has no Disc field, unlike IssueModel and
            ProjectModel. M2 Task 2.1 Step 1 reduces to a one-line model documentation edit;
            drop the "update every consumer" branch.
          round: 1
      blocked: true
---

# Gate ledger — ariadne#194 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-20T16:55:07-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] Issue file has a corrupt region — a mangled heading and a second, contradictory ## Plan
  workshop/issues/000194-review-anchor-commit.md:199-216 holds `## Log` with the
  reason and E is dropped.` plus a stale duplicate ## Done when tail and a second
  ## Plan whose only open item is "Re-plan for the folded scope". Delete 199-216.
  Also: the live ## Plan header at :179 says "Four review boundaries" and the M1
  row says "regardless of M2-M4" when only M1-M3 exist.
- **PQ-2** [Important] One issue-wide boundary ledger blows gatestate.Decide's round cap before the close runs
  Decide computes CapReached from len(l.Rounds) against DefaultRoundCap=3
  (decide.go:14,45). With M1+M2+M3 rounds accumulating in one ledger the whole-issue
  close starts past the cap and demotes Important findings on round 1; OpenFindings
  likewise spans boundaries so an M1 finding blocks M3. gatestate.Round
  (ledger.go:57-75) also has no boundary field, so "rounds tagged with their
  boundary" is a schema change the Core-concepts table omits.
- **PQ-3** [Important] Two prior-findings channels with two id namespaces feed one output fence (ARCH-DRY)
  code-review.md:57-67, embedded into milestone-review.md via CODE_REVIEW_BODY,
  already tells the boundary reviewer to read the plan-gate ledger and re-raise its
  open PQ-* findings. Adding PRIOR_FINDINGS from the BR-* ledger gives the reviewer
  two namespaces and one fence. The plan does not say what a disposition naming an
  unknown id does, nor which carry-forward mechanism survives.
- **PQ-4** [Important] Free-text family slugs have no cross-round anchoring, so escalation silently never fires
  FamilyCounts keys on exact strings and RenderPriorFindings (gatestate/prompt.go:17-70)
  renders no family, so a stateless reviewer writing block-opener-rule then
  block-opener leaves every count at 1 and the escalation instruction — the issue's
  stated purpose (ARCH-PURPOSE) — never triggers. Task 3.5's consistent-slug fixture
  passes by construction. Plan needs: render in-play families with a reuse
  instruction, a normalization step, and a near-miss-slug test.
- **PQ-5** [Important] Precedence between the boundary verdict and the new ledger refusal is unstated
  M2 Task 2.2 Step 4 adds a ledger-derived refusal alongside closeVerdictOutcome
  (close.go:1064-1082). SHIP with an undisposed Important, or REWORK with everything
  disposed — which wins? Also undefined: a boundary protocol miss. Plan-quality warns
  and falls back to the verdict token (changecode.go:495); at the boundary that would
  finalize a close carrying no ledger memory.
- **PQ-6** [Minor] Compress M1 Task 1.1's procedural steps and line-numbered fixture inventory
  Literal test bodies, patch snippets and a cited-line inventory are stale on arrival
  and one entry is already wrong: milestoneclose_test.go:120,134 are renderer tests
  that pass Head "HEAD" as fixture data (:113-115,127-129) and need no change. The
  per-function strategy lines already present are the part worth keeping.
- **PQ-7** [Minor] Sidecar count is wrong — 72 files exist repo-wide, not 86, and 67 carry ..HEAD
  The plan and the Revisions entry both assert 86 archived sidecars all recording
  base..HEAD. Actual: 72 *-review.md repo-wide (70 archived), 67 containing ..HEAD,
  69 distinct window strings. The underlying claim is still true.
- **PQ-8** [Minor] finding.cue discovery.glob has no Go consumer — the string-to-list fork is empty
  FindingModel (pkg/vocab/finding.go:22-28) has no Disc field, unlike IssueModel and
  ProjectModel. M2 Task 2.1 Step 1 reduces to a one-line model documentation edit;
  drop the "update every consumer" branch.

## Open findings

- **PQ-1** [Critical] Issue file has a corrupt region — a mangled heading and a second, contradictory ## Plan
- **PQ-2** [Important] One issue-wide boundary ledger blows gatestate.Decide's round cap before the close runs
- **PQ-3** [Important] Two prior-findings channels with two id namespaces feed one output fence (ARCH-DRY)
- **PQ-4** [Important] Free-text family slugs have no cross-round anchoring, so escalation silently never fires
- **PQ-5** [Important] Precedence between the boundary verdict and the new ledger refusal is unstated
- **PQ-6** [Minor] Compress M1 Task 1.1's procedural steps and line-numbered fixture inventory
- **PQ-7** [Minor] Sidecar count is wrong — 72 files exist repo-wide, not 86, and 67 carry ..HEAD
- **PQ-8** [Minor] finding.cue discovery.glob has no Go consumer — the string-to-list fork is empty
