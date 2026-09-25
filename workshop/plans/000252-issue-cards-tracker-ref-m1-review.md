# Boundary Review — ariadne#252 (milestone M1)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e8fcf6116f82106b029519dc02009272a13adfcd..7dbc0bbfb8bb142f2f52b107f3d478d4a2f1221e |
| command | sdlc milestone-close --issue 252 --milestone M1 |
| reviewer | codex |
| timestamp | 2026-09-25T15:23:52-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The M1 implementation itself is well-tested and preserves the intended unactivated foundation, but the durable plan’s Core concepts table contradicts the pinned tree: it claims future M2–M4 entities as present/modified. The plan must be corrected before the boundary can be crossed.

1. Strengths

- Card and mirror ownership are cleanly separated and strongly validated (`card.go`, `mirror.go`).
- Tracker writes use pinned snapshots, CAS checks, exact-byte updates, and mandatory receipts (`repository.go:45-81`).
- Recovery transitions explicitly model uncertain publication and bounded retries (`receipt.go`).
- Atlas documentation and index linkage were added for the new tracker foundation.

2. Critical findings

- `workshop/plans/000252-issue-cards-tracker-ref-plan.md:19-34` — `core-concepts-inventory-drift`  
  The Core concepts table claims `cmd/sdlc/internal/tracker/migration.go` exists as a new PURE entity, but the file is absent; it also claims `cmd/sdlc/internal/activetime/commit.go` is modified, but it is unchanged in the review range. Several listed M2–M4 integration entities are likewise absent. Per the Core concepts contract, this is a Critical contradiction. Scope the table to M1 or explicitly mark future entities as not-yet-delivered, then add the required `## Revisions` entry.

3. Important findings

- `workshop/plans/000252-issue-cards-tracker-ref-plan.md:163-164` — `boundary-checklist-not-closed`  
  The final M1 checklist item remains unchecked even though the boundary submission claims M1 review. Tick it only after recording repeat-test, benchmark, atlas, and commit evidence, or clarify that the boundary is not yet claiming completion.

4. Minor findings

- None.

5. Test coverage notes

Focused verification passed:

```text
go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/tracker ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/processgroup ./pkg/vocab -count=1
git diff --check
```

The tests include real disposable Git repositories, stateful fake publication races, cancellation, mirror fuzzing, and recovery-state sequences.

6. Architectural notes

- ARCH-DRY: pass — shared `TrunkFile` CAS and vocabulary-derived ownership are reused.
- ARCH-PURE: pass — card/mirror and transition logic are isolated from Git adapters.
- ARCH-PURPOSE: pass for M1 — the foundation remains explicitly unactivated; consumer migration is deferred to later milestones.
- ARCH-MOCK: pass — stateful Git fakes and disposable real-Git tests are present.
- ARCH-CONSTRAINTS: pass — snapshot, blob, diagnostic, entry-count, and retry bounds are explicit.
- ARCH-SECURE: pass — paths, YAML, OIDs, receipts, modes, and repository identity are validated.
- ARCH-ORDER: pass — creation, transfer, and completion use explicit tagged transitions with uncertainty states.
- ARCH-FUNERAL: pass — tracker history, receipts, and recovery residue have documented lifecycle policies.

7. Plan revision recommendations

Add a revision such as:

```markdown
### 2026-09-25 — M1 boundary inventory correction

Reason: the project-wide Core concepts table listed future M2–M4 entities as if
they were present in the M1 tree.

Delta: scope the M1 review inventory to delivered entities and mark migration,
consumer-reader, completion-adapter, and cutover entities as future work.
```

```findings
findings:
  - id: new
    severity: Critical
    family: core-concepts-inventory-drift
    title: |
      Core concepts table claims future entities are delivered
    detail: |
      The plan claims tracker/migration.go is new and activetime/commit.go is modified, but the former is absent and the latter is unchanged in the pinned M1 range; several later-milestone entities are also absent. Scope or revise the inventory before crossing the boundary. ARCH-PURPOSE
  - id: new
    severity: Important
    family: boundary-checklist-not-closed
    title: |
      Final M1 plan checklist item remains unchecked
    detail: |
      The M1 row covering repeated tests, benchmarks, atlas documentation, and milestone-close evidence remains unchecked while the review is being submitted. Close the row with evidence or defer the boundary.
```
