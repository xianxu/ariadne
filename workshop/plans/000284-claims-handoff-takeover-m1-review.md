# Boundary Review — ariadne#284 (milestone M1)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 01079ab94940e6d1df39eed60bfb48b148105aaa..55afb930b1c75dfb5d505b137a48541b1be24cb9 |
| command | sdlc milestone-close --issue 284 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-07T14:35:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M1 delivers everything its Done-when asks for. A set claim (`claim --issue 9,10,11`) is one tracker commit. A peer claiming one member fails the whole set and claims none of the others. A race on another card is decided again and lands: the tracker-level test covers a truly unrelated card, and the claim-level test covers a harmless change to a member. Refresh-on-claim is in place and tested: the resting branch fast-forwards, a dirty file in the way blocks the fast-forward with a warning, and an issue branch whose details differ from main's gets a warning. The recovery catalog, `claim.md` and the atlas verb table were updated in the same milestone. I inspected the stat, name-status and full diff and ran the new tests plus the claim, adopt, recovery and verb-contract tests: `go test ./cmd/sdlc/internal/tracker -run ChangeCards` and the targeted `./cmd/sdlc` claim tests both pass. Nothing blocks the boundary; only Minor findings remain.

**Strengths**
- `tracker.ChangeCards` (`cmd/sdlc/internal/tracker/changecards.go:23`) decides again over fresh bytes on every attempt, on top of `UpdateManyPrepared`. The old single-card path could only refuse a changed card; this one can judge it again, which is the right design for D1. It also refuses any write to a card the decision did not read.
- Single-issue and multi-issue claims share one path: `claimSetDecision` handles 1..N (ARCH-DRY), and the single-card claim now benefits from re-deciding too.
- `TestClaimSetDecision` covers every combination of the model's statuses with owner none / me / other, and checks each against the single-card `claimDecision`. That ties the set behaviour to the single-card behaviour instead of restating the implementation.
- The seams are pinned. `cardsPublish` sits next to `cardPublish`, `TestCardPublishCallers` covers both, and `loseResponses` injects a lost response into the set path as well. The rerun-after-lost-response contract stays testable.
- Refresh is warn-only after the card has landed, and it compares only the details bodies, because the mirror rewrites the frontmatter by design. Each step of the ordering can be settled by rerunning.

**Critical:** none.

**Important:** none.

**Minor**
- `cmd/sdlc/claim.go` (`refreshAfterClaim`): if `rev-parse HEAD` fails, refresh returns silently. If `merge-base --is-ancestor` fails, the warning says "has commits main lacks", which is the wrong cause. Report the actual error instead.
- ARCH-DRY: `claimRefs`, `claimArg` and `cardsMessage` each map IDs through `CLIRef` and differ only in prefix and separator. Also, `ChangeCards` re-implements `validateReplacement`'s blob and budget arithmetic in a cumulative form. A `validateReplacements` over a running wire total would keep one source.
- `ChangeCards` does not remove duplicate IDs. Its callers already do, but the tracker API should not trust them: duplicates would produce a subject like `#9,#9`.
- A started, unowned card inside a set is refused with a pointer to `--adopt`, but `--adopt` refuses sets. The message could say "claim #N alone with --adopt".
- `workshop/projects/claimant-ownership.md`: the M1 line is already ticked with `actual: 1.15h` and `closed:` before `milestone-close` has run. The M1–M4 lines are also siblings of the parent #284 line rather than nested under it. Let the close gate's measured value be the one recorded.
- `atlas/workflow/process-manual.md`: the `recovery` helptext entry is an unrelated addition (#280's surface). It's harmless, but it isn't M1 scope.

**Test coverage notes**
- Plan Task 2 calls for "a peer edits an unrelated card → the set lands" at the claim level. `TestClaimSetRace` instead changes a *member* (#9's date). That is a stronger re-decision check, but the literal unrelated-card case is proven only at the tracker level (`TestChangeCardsRedecidesAfterARace`). That's acceptable, since the claim delegates to that path.
- No test sends a claim set through a lost response (`loseResponses` plus a set rerun). The seam exists, so this would be cheap to add.

**Architecture**
- ARCH-DRY: flagged as Minor (the three ID-rendering helpers and the inlined budget check).
- ARCH-PURE: pass. `claimSetDecision` is pure, with readiness passed in; `refreshAfterClaim` is thin IO glue.
- ARCH-PURPOSE: pass. All three M1 Done-when items are delivered, and the recovery catalog, help and atlas moved in the same milestone.
- ARCH-MOCK: pass. Tests use real-git fixtures and a peer `tracker.Repository`.
- ARCH-CONSTRAINTS: pass. The batch is bounded by the snapshot wire budget. Refresh adds one more `main` snapshot read per claim, which is acceptable.
- ARCH-SECURE: pass. Card bytes are parsed, and identity is enforced per replacement. The fast-forward goes only to the pinned, fetched main and never overwrites a dirty file.
- ARCH-ORDER: pass. The order is card, then refresh. A lost response surfaces through `uncertainCardWrite` before refresh runs, and a rerun decides from the cards. The race tests inject the ordering through the `beforePush` seam, so the failing interleaving can be reproduced.
- ARCH-FUNERAL: pass. Nothing durable is created beyond tracker commits.

**Notes for M2–M4:** `ChangeCards` is the stable seam for unclaim, takeover and #286's `abandon`. Keep its decide signature as it is. The Core concepts table already places `adoptDecision`'s fold into `claimDecision` in M3.

**Plan revisions:** none needed. The code matches D1–D3 and Tasks 1–4. Task 2's "Files" list names `claimremote_test.go`/`verbcontract_test.go`, but the tests landed in `claimset_test.go`. That is cosmetic and could go in an optional Revisions line.

```findings
findings:
  - id: new
    severity: Minor
    family: silent-error-in-io-glue
    title: |
      refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
    detail: |
      cmd/sdlc/claim.go refreshAfterClaim: surface the actual git error in the warning instead of returning silently or naming the wrong cause.
  - id: new
    severity: Minor
    family: duplicated-id-rendering
    title: |
      claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
    detail: |
      ARCH-DRY: one ref-join helper parameterized by prefix/separator, and a cumulative validateReplacements in tracker/reader.go.
  - id: new
    severity: Minor
    family: api-trusts-caller-normalization
    title: |
      tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
  - id: new
    severity: Minor
    family: refusal-points-at-unusable-action
    title: |
      a started unowned card in a set is refused toward --adopt, which itself refuses sets
    detail: |
      Message should say to claim that issue alone with --adopt.
  - id: new
    severity: Minor
    family: hand-recorded-actuals
    title: |
      project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line
  - id: new
    severity: Minor
    family: test-gap-lost-response-set
    title: |
      no test drives a claim set through a lost publication response and a rerun
```
