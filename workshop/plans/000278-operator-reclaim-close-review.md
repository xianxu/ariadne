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

---

## Re-review — 2026-10-01T21:23:53-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 278 — Add operator-directed reclaim for recovery |
| repo | ariadne |
| issue file | workshop/issues/000278-operator-reclaim.md |
| boundary | whole-issue close |
| milestone | — |
| window | d97ba867ffefc3f07e10fe145741754ef75f2a83..320a1c0fe0045f1cf7b134e3c511a76ddd5c8834 |
| command | sdlc close --issue 278 |
| reviewer | claude |
| timestamp | 2026-10-01T21:23:53-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All five round-1 findings are fixed, and I checked each one against the code and tests, not just the commit messages. The targeted tests pass at HEAD: the reclaim tests, the verb-contract table, `statusDecision`, the ownership gates, the flag-help check and the tracker `TestUpdateCardTrailers`. For BR-2 I ran a mutation check in a scratch copy: with the "already mine" mirror refresh removed, `TestReclaimStaleAndLostResponse` fails at `reclaim_test.go:237` ("details do not mirror the working card"). With the refresh restored it passes. The design matches the Spec and Done-when:
- inspect, then confirm pinned by `--expect`;
- a compare-and-swap on the same card read;
- a retry decided from the card;
- trailers as the record;
- no change to any worktree;
- an AST guard that keeps the transfer operator-only.

One new finding is left, and it doesn't block. The README paragraph on #277 ownership lists the ownership verbs but doesn't mention the new `sdlc reclaim`.

1. **Strengths**
   - `reclaimDecision` (`cmd/sdlc/reclaim.go:35-70`) is pure and ordered on purpose. It checks "mine" before the stale-`--expect` check, so a retry after a lost response is a no-op even with an old revision. The table test covers this ("mine, stale expect still no-op").
   - `UpdateCardWithTrailers` (`internal/tracker/repository.go:66-71`) checks trailer shape before publishing anything. Plain `UpdateCard` messages stay byte-identical, and a test pins that.
   - Inspect reuses `reclaimDecision` (`reclaim.go:215`) instead of repeating the rules, so inspect and confirm can't drift apart (ARCH-DRY).
   - The real-git tests check behavior, not mocks: no worktree is touched (porcelain status before and after), every continuation gate refuses the old owner before any review runs, the new owner's start-plan passes, and a two-clone race has exactly one winner.
   - The lost-response test injects uncertainty *after* a real publish, which is the hard case.

2. **Critical:** none.

3. **Important**
   - **README doesn't mention `sdlc reclaim`.** `README.md:26-30` describes ownership: claim records it, `--adopt` handles cards claimed before #277, and `sdlc move` carries it with the branch. It doesn't name the new user-facing verb for moving responsibility between owners. Fix: add one sentence, e.g. "Moving responsibility to another workspace is the operator-directed `sdlc reclaim` (inspect, then `--expect REV --reason …`)." The atlas, process manual and help already cover it.

4. **Minor:** none new. In the guard (`reclaim_test.go:338`), `vs.Names[0]` labels a multi-name `var a, b = …` spec by its first name only. That affects the error label, not detection, so it isn't worth a finding.

5. **Test coverage notes:**
   - BR-1 is now covered at `reclaim_test.go:164-171` (all three gates, with `judged == false`) and `:190-198` (the new owner passes).
   - BR-2 is confirmed red without the fix and green with it (scratch mutation above).
   - BR-4 is covered at `:187-189`.
   - BR-5 now walks `GenDecl` value specs. The vacuity check (`:346-350`) keeps the guard from passing trivially.
   - BR-3: `git log --max-count` is applied after `--grep` filtering, so the limit means "the last 20 reclaims", which matches the help and plan.

6. **Architectural notes**
   - ARCH-DRY pass: decisions stamp through `SetCardClaimant`, and inspect reuses the decision.
   - ARCH-PURE pass: decision and trailer parse/format are pure and table-tested; IO stays in `runReclaim`.
   - ARCH-PURPOSE flag (the README finding only); every Done-when bullet is delivered.
   - ARCH-MOCK pass: real git, plus a built-binary race.
   - ARCH-CONSTRAINTS pass: history is bounded at 20 and read only on inspect.
   - ARCH-SECURE pass: a multi-line reason is refused, which blocks trailer spoofing, and trailers are re-checked in the tracker layer.
   - ARCH-ORDER pass: the CAS keys on the same blob read as the decision. A stale observation is refused through `--expect`. An uncertain outcome is reported as uncertain, not collapsed into success or failure, and is reconciled from the card.
   - ARCH-FUNERAL pass: the only durable output is one tracker commit per reclaim, kept as long as tracker history.

7. **Plan revision recommendations:** none. The plan's history bound ("the last 20 reclaims, selected by their trailer") now matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      reclaim_test.go:164-171 runs start-plan, change-code and close from the old slot, each refused with judged=false; :190-198 shows the new owner's start-plan passes.
  - id: BR-2
    disposition: addressed
    note: |
      reclaim.go:166-168 refreshes on the already-mine path; scratch mutation removing it turns TestReclaimStaleAndLostResponse red at reclaim_test.go:237.
  - id: BR-3
    disposition: addressed
    note: |
      reclaimHistoryLimit=20 with --grep on the trailer (max-count applies after grep); help and plan both say last 20.
  - id: BR-4
    disposition: addressed
    note: |
      reclaim.go:157-159 refuses --reason without --expect; tested at reclaim_test.go:187-189.
  - id: BR-5
    disposition: addressed
    note: |
      Guard now walks GenDecl ValueSpec initializers (reclaim_test.go:334-341); the vacuity check confirms reclaimEffect's initializer is seen.
findings:
  - id: new
    severity: Important
    family: docs-new-surface-missing
    title: |
      README's ownership paragraph lists claim, --adopt and move but not the new sdlc reclaim verb
    detail: |
      README.md:26-30 describes #277 ownership verbs; add one sentence naming `sdlc reclaim` (inspect, then --expect/--reason) as the operator-directed transfer between owners.
```
