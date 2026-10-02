# Boundary Review — ariadne#278 (whole-issue close)

| field | value |
|-------|-------|
| issue | 278 — Add operator-directed reclaim for recovery |
| repo | ariadne |
| issue file | workshop/issues/000278-operator-reclaim.md |
| boundary | whole-issue close |
| milestone | — |
| window | d97ba867ffefc3f07e10fe145741754ef75f2a83..8bbf5d5bed0074a1a1767c6bac7b79ebcc299b5c |
| command | sdlc close --issue 278 |
| reviewer | claude |
| timestamp | 2026-10-01T21:05:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** `sdlc reclaim` delivers what the Spec asks for. Inspect only reads. Confirm is a CAS pinned by `--expect` and needs a one-line `--reason`. The tracker commit carries `Reclaim-From/To/Reason` trailers, and the same functions write and read them. Retries and lost responses are decided from the card. A real-git race test shows exactly one winner. An AST guard keeps the transfer reachable only from the command, and `sdlc move` is unchanged apart from its refusal text. Nothing here blocks the close. What remains is a few small mismatches between the plan or help and the code:
- Rerunning after a lost response skips the mirror refresh that the help promises.
- The plan says the new owner passes the gates after a reclaim, but no test checks that.
- The history limit and the help's "every past reclaim" wording disagree.

### 1. Strengths
- **Pure core.** `reclaimDecision` (`cmd/sdlc/reclaim.go:35`) is pure and checks things in the right order. It returns "already mine" before checking `--expect`, which is why an identical retry, or a rerun after a lost response, ends as a no-op. It refuses a stale revision before anything is written. It reuses `SetCardClaimant` and `MatchClaimant` (ARCH-DRY).
- **One CAS on one read.** The decision and the CAS both use the card from one snapshot (`reclaim.go:146-169`). Each error case maps to a clear next step: `ErrCardChanged` says to inspect again, and `ErrPublicationUncertain` says to rerun the same command. `TestReclaimStaleAndLostResponse` injects a lost response after the effect has landed and checks that the rerun reconciles without a new commit.
- **Trailers are validated at the tracker boundary.** `UpdateCardWithTrailers` (`repository.go:64`) refuses CR/LF and trailers without a key before publishing anything. The message of plain `UpdateCard` stays byte-identical, and a test checks that (ARCH-SECURE).
- **The no-automation rule is mechanical, not just prose.** `TestReclaimIsOnlyOperatorInvoked` limits who may reference each entry point, and it fails if any guarded symbol is never seen, so it cannot pass vacuously.
- **Docs.** The help covers out-of-band coordination, what reclaim does not do, and the guarantees. The atlas gains a section and a verb-table row, and every #277 refusal now names the command.

### 2. Critical
None.

### 3. Important
- **The plan claims the new owner passes, but no test checks it** (`reclaim_test.go:169-173`).
  - Plan step 2 says the old workspace's start-plan, change-code and close "refuse as Foreign, and the new one passes".
  - The test only runs start-plan from the old workspace.
  - Fix: run `ownershipGate` for change-code and close from the old slot, and one gate from `r.root`, asserting an empty refusal. Alternatively, revise the plan step.

### 4. Minor
- **No mirror refresh on the "already mine" path** (`reclaim.go:160-163`). After a lost response that actually landed, the rerun returns early, so the local details `card_mirror` stays stale. That contradicts the help ("the details' card mirror is refreshed"). Claim's own "already mine" path does refresh (`claim.go:168-176`). It heals at the next change-code, but it is cheap to refresh here too.
- **The history limit drifts from the plan and the help.** `reclaimHistory` reads `--max-count=50` card commits (`reclaim.go:233`). The plan says 20, and the help says "every past reclaim". The limit counts all card commits, not just reclaims, so older reclaims can silently drop out of the view.
- **Inspect silently ignores `--reason` when `--expect` is missing.** Refusing, or at least noting it, would avoid an operator thinking they had confirmed.
- **The AST guard only scans function bodies.** A reference in a package-level `var x = reclaimEffect` would escape it. This is low risk.
- **The reclaimable statuses are hard-coded** (`reclaim.go:42`): `working`, `blocked`, `codecomplete`. That matches the codebase's existing idiom, but AGENTS.md says to read the status enum from the vocabulary model.

### 5. Test coverage notes
- The decision table covers each status, mine/foreign/unattributed ownership, a stale `--expect`, and empty or multi-line reasons.
- The trailer round trip covers a reason containing `:` and unicode.
- The real-git tests cover:
  - inspect not writing;
  - the transfer itself and its trailers;
  - dirty work left untouched in both worktrees;
  - a stale confirm leaving the card unchanged;
  - an injected lost response;
  - a two-clone race;
  - history appearing in inspect.
- Gaps: the new owner passing the gates, and mirror state after a rerun that follows a lost response.

### 6. Architecture notes
- **ARCH-DRY: pass.** It reuses the claimant helpers, and one source writes and reads the trailers.
- **ARCH-PURE: pass.** The decision and trailer functions are pure; `runReclaim` is a thin IO shell.
- **ARCH-PURPOSE: pass.** Every Done-when item is addressed; the one test gap is listed under Important.
- **ARCH-MOCK: pass.** The tests run against real git, plus a built-binary race.
- **ARCH-CONSTRAINTS: pass.** The history read is bounded (see the Minor on how).
- **ARCH-SECURE: pass.** The reason is limited to one line at two layers. History parsed from the shared tracker is only displayed.
- **ARCH-ORDER: pass.** `--expect` and the CAS guard against stale reads, the race test checks interleaving across two binaries, and uncertain outcomes are reconciled from the card rather than assumed.
- **ARCH-FUNERAL: pass.** The only durable output is one tracker commit per operator action.

### 7. Plan revision recommendations
- Add a Revisions entry: the history limit is 50 card commits, not 20, unless the code changes to 20. Also align the help's "every past reclaim" wording.
- If the new-owner and change-code/close assertions are not added, revise real-git test 2 to say "old workspace's start-plan refuses".

```findings
findings:
  - id: new
    severity: Important
    family: plan-claimed-test-missing
    title: |
      The wrong-owner test checks only the old workspace's start-plan; the plan's "new one passes" and change-code/close refusals are untested
    detail: |
      reclaim_test.go:169-173. Add ownershipGate runs for change-code and close from the old slot, plus a passing gate from r.root, or revise plan step 2.
  - id: new
    severity: Minor
    family: retry-path-skips-side-effects
    title: |
      The "already mine" rerun skips the local mirror refresh, so after a lost response that landed, details stay stale
    detail: |
      reclaim.go:160-163 returns before refreshLocalMirror. Claim's own "already mine" path refreshes (claim.go:168-176), and the help promises the refresh.
  - id: new
    severity: Minor
    family: doc-code-bound-drift
    title: |
      reclaimHistory reads 50 card commits; the plan says 20 and the help says "every past reclaim"
  - id: new
    severity: Minor
    family: silently-ignored-flag
    title: |
      --reason without --expect is silently ignored (inspect mode)
  - id: new
    severity: Minor
    family: guard-scope-gap
    title: |
      TestReclaimIsOnlyOperatorInvoked scans only function bodies; package-level var initializers escape it
```
