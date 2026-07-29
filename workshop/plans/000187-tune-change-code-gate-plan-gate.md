---
gate: plan-quality
issue: 187
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-07-29T12:05:30-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: Task 14's replay harness reads an error runChangeCode never returns; round 1 os.Exit(1)s the test process
          detail: runChangeCode iterates changeCodeGates and calls exitWithCode(1) on a gate error (changecode.go:130-139 → term.go:56-59 → os.Exit), so the plan's `err := runChangeCode(…)` loop (plan:2489-2492) can never observe a blocking round — and the blocking round is the point. expectDie (die_test.go:33-36) swaps the `die` var, which this path bypasses. Respecify as a subprocess run of the built binary (reading exit code + the scratch repo's -plan-gate.md) or as a direct runPlanQualityJudge drive, which returns an error and owns the full ledger path. Also drop the "seed a reconciling
          round: 1
        - id: PQ-2
          severity: Important
          title: churnForWindow specified on gitx.Capture, which cannot report the git failure Task 13 promises to warn on
          detail: gitx.Capture returns "" on any error (window.go:50-56) and its doc explicitly warns against using it where errored must be distinguished from empty (:47-49). Task 13's "degrade with a warning, never break" contract is then unimplementable — a git error prints an all-zero churn line indistinguishable from an empty window. Use gitx.RunGit (window.go:38-40), the exported error-returning variant.
          round: 1
        - id: PQ-3
          severity: Minor
          title: Task 13 Step 4 still says nine appended ledger columns where Step 3 says ten
          detail: plan:2359 reads "Ledger columns become nine, not seven" while Step 3 enumerates ten (indices 10–19, 20 total, len(c) >= 20). The enumeration is correct; the sentence is stale from before round 4's arithmetic correction.
          round: 1
        - id: PQ-4
          severity: Minor
          title: ClassifyPath has four buckets and no stated default for paths outside them
          detail: The Task 11 table (plan:2158) covers cmd/, pkg/, atlas/, workshop/, construct/ and AGENTS.base.md but not go.mod/go.sum, Makefile, .github/, docs/vision/ or construct/base.manifest. Name the fallback (an explicit Other bucket, or a stated default) so lockfile-sized churn doesn't silently inflate a real bucket.
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-07-29T12:09:24-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Harness now drives runPlanQualityJudge directly; verified signature match at changecode.go:404 and the exitWithCode path at changecode.go:130-139.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Task 12 mandates gitx.RunGit with the error-distinguishing rationale; one stale `gitx.Capture` mention remains at plan:163-164 in the authoritative Integration points table — fix in passing.
          round: 2
        - id: PQ-3
          disposition: not-addressed
          note: plan:2387 corrected to TEN, but "seven" survives at plan:2301 (Task 13 Files header), plan:2439 (Step 7 atlas), and the issue's M2 Plan row.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: code-prod named as the explicit default; go.mod/go.sum/Makefile.workflow/.github/base.manifest/docs pinned in the ClassifyPath table.
          round: 2
      blocked: false
content_hash: 97f15a6fc16c57f3a0961114202dded46dd815b17ab273d595e8cfb432c5f2af
---

# Gate ledger — ariadne#187 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-07-29T12:05:30-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] Task 14's replay harness reads an error runChangeCode never returns; round 1 os.Exit(1)s the test process
  runChangeCode iterates changeCodeGates and calls exitWithCode(1) on a gate error (changecode.go:130-139 → term.go:56-59 → os.Exit), so the plan's `err := runChangeCode(…)` loop (plan:2489-2492) can never observe a blocking round — and the blocking round is the point. expectDie (die_test.go:33-36) swaps the `die` var, which this path bypasses. Respecify as a subprocess run of the built binary (reading exit code + the scratch repo's -plan-gate.md) or as a direct runPlanQualityJudge drive, which returns an error and owns the full ledger path. Also drop the "seed a reconciling
- **PQ-2** [Important] churnForWindow specified on gitx.Capture, which cannot report the git failure Task 13 promises to warn on
  gitx.Capture returns "" on any error (window.go:50-56) and its doc explicitly warns against using it where errored must be distinguished from empty (:47-49). Task 13's "degrade with a warning, never break" contract is then unimplementable — a git error prints an all-zero churn line indistinguishable from an empty window. Use gitx.RunGit (window.go:38-40), the exported error-returning variant.
- **PQ-3** [Minor] Task 13 Step 4 still says nine appended ledger columns where Step 3 says ten
  plan:2359 reads "Ledger columns become nine, not seven" while Step 3 enumerates ten (indices 10–19, 20 total, len(c) >= 20). The enumeration is correct; the sentence is stale from before round 4's arithmetic correction.
- **PQ-4** [Minor] ClassifyPath has four buckets and no stated default for paths outside them
  The Task 11 table (plan:2158) covers cmd/, pkg/, atlas/, workshop/, construct/ and AGENTS.base.md but not go.mod/go.sum, Makefile, .github/, docs/vision/ or construct/base.manifest. Name the fallback (an explicit Other bucket, or a stated default) so lockfile-sized churn doesn't silently inflate a real bucket.

## Round 2 — 2026-07-29T12:09:24-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Harness now drives runPlanQualityJudge directly; verified signature match at changecode.go:404 and the exitWithCode path at changecode.go:130-139.
- PQ-2 — addressed — Task 12 mandates gitx.RunGit with the error-distinguishing rationale; one stale `gitx.Capture` mention remains at plan:163-164 in the authoritative Integration points table — fix in passing.
- PQ-3 — not-addressed — plan:2387 corrected to TEN, but "seven" survives at plan:2301 (Task 13 Files header), plan:2439 (Step 7 atlas), and the issue's M2 Plan row.
- PQ-4 — addressed — code-prod named as the explicit default; go.mod/go.sum/Makefile.workflow/.github/base.manifest/docs pinned in the ClassifyPath table.

## Open findings

- **PQ-3** [Minor] Task 13 Step 4 still says nine appended ledger columns where Step 3 says ten
