# Boundary Review — ariadne#284 (milestone M4)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | 6c682381ad458e8fc5075301a2d71cf72f5c6844..c766a471e99857ae553b6a44e04baf5f444a4bb4 |
| command | sdlc milestone-close --issue 284 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-10-07T21:34:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M4 delivers what it set out to: `sdlc state` now shows each issue's owner and claim age, plus views grouped by slot and by operator. Claim times come from a pure stream parser (`tracker.ClaimTimes`) that has unit tests, a fuzz target and a test that it stops reading early. A real-git integration test checks that a later `start` write does not reset the claim time. One real defect blocks a clean SHIP. When reading claim ages fails, `runState` appends an "info" drift line, but the next statement overwrites `s.Drift`, so the failure never reaches the output. A related gap: if `git log` itself fails, its exit status is thrown away. Both belong to the `silent-error-in-io-glue` family, which now has 5–6 findings, so this report states the rule rather than patching two sites.

**Inspection.** I ran the stat, name-status and full-diff commands for the pinned range. I also checked the code each claim depends on: the `operationToken` call sites, `startplan.go`'s card write, `move.go`'s relocation, and the `runState` drift sequence.

**Strengths**
- `cmd/sdlc/internal/tracker/claimtimes.go:29` — `ClaimTimes` is genuinely pure (only an `io.Reader` in). It skips malformed records, keeps the newest claim per card, and stops early, which bounds the cost by the oldest live claim. It is pinned by `TestClaimTimesStopsEarly` and `FuzzClaimTimes` (ARCH-PURE, ARCH-SECURE).
- `cmd/sdlc/stateclaims_test.go:47` — the integration test compares against the claim commit's own time after a later `start` write. That is exactly the "a later status change doesn't reset age" contract, tested against real tracker history.
- `renderProseAt` takes `now` as an input (`state.go:443`), which makes the age rendering deterministic and testable.
- No new card field: claim times come from the existing `Tracker-Operation` trailer, so older binaries' strict claimant parse stays safe. This is a good call, and the project file records it.

**Critical findings:** none.

**Important findings**
1. `cmd/sdlc/state.go:172-177` — the claim-ages info line is overwritten. `s.Drift = append(...)` is immediately followed by `s.Drift = detectDrift(...)`, so the line is discarded and no test catches it. **This is the 5th finding in family `silent-error-in-io-glue`.**
   - **Rule:** a probe in IO glue returns its failure as a value, and the caller *accumulates* findings — only one assignment to `s.Drift`, and everything after it appends. The same rule covers child-process status (item 2).
   - **Fix:** run `fillClaimTimes` after `detectDrift`, or have it return a `[]DriftFinding` that is appended.
   - **Test:** add one where `recordsRepository` or the git log fails, and assert the info line appears.
2. `cmd/sdlc/state.go:609-611` — `_ = cmd.Wait()` throws away `git log`'s exit status every time, not only after a deliberate early kill. A missing tracking ref or a git error yields empty stdout, so ages silently disappear. Same family and rule.
   - **Fix:** treat `Wait`'s error as real unless `ClaimTimes` returned early because every wanted card was resolved (`len(times) == len(want)`). Then report through the accumulating drift path from item 1.

**Minor findings**
- `claimtimes.go:19` — `claimKinds` restates, by hand, which verbs set an owner. Tokens are minted as string literals at the call sites (`claim.go:224`, `reclaim.go:130`, `claimant.go:221`), so a future verb that sets an owner silently drops out of claim ages (ARCH-DRY/ARCH-PURPOSE). Name the verbs as shared constants in `tracker` and use them both at the mint sites and in `claimKinds`.
- `state.go:600` — `exec.CommandContext("git", …, repo.TrackingRef(), …)` runs a git command against a tracker ref outside `tracker.Repository`, which owns that ref. **This is the 2nd finding in family `external-call-outside-seam`.** Rule: git reads of tracker refs live on `tracker.Repository`. A `repo.ClaimLog(ctx)` reader there would also own `ClaimLogFormat`.
- `state.go:563` — `claimViews` gets the short age `(3h)` by stripping the prefix and suffix from `claimAge`'s prose. Factor out a `shortAge(d)` and have `claimAge` wrap it.
- A card claimed before operation trailers existed (no claim-kind record in history) makes every `sdlc state` scan the whole tracker log. This is acceptable at current sizes; note it in the helptext or atlas.
- `ClaimTimes` matches a card's current path only. History before a path rename (if one is possible) won't match, so the age is then absent, not wrong.

**Test coverage notes**
- Pure grouping and rendering are covered directly. `ClaimTimes` covers claim versus non-claim operations, multiple cards in one commit, malformed records, early stop, and fuzzing.
- Missing: the failure path of `fillClaimTimes` (both Important findings would have been caught by a single test asserting the info line).
- The `time.Sleep(1100ms)` in the integration test is tolerable but adds wall time. Passing a committer date through the env would avoid it.

**Architecture**
- **ARCH-DRY:** flag (Minor). `claimKinds` hand-restates the token set.
- **ARCH-PURE:** pass. The parser and the views are pure; IO is confined to `fillClaimTimes`.
- **ARCH-PURPOSE:** pass. Owner, age and both views are delivered as the Spec's M4 line asks.
- **ARCH-MOCK:** flag (Minor). The direct `exec` call is outside the tracker seam, though tests do run against real git fixtures, as the repo does elsewhere.
- **ARCH-CONSTRAINTS:** pass with a note. Reading is bounded by early stop; the legacy-card full scan is noted above.
- **ARCH-SECURE:** pass. Untrusted history is parsed defensively and fuzzed, and a bad record degrades to "no age" rather than a fabricated time. However, the visible-degrade half fails (Important 1/2).
- **ARCH-ORDER:** pass. `fillClaimTimes` keeps no state between events: it reads once, and the child process is killed and waited on within the function.
- **ARCH-FUNERAL:** pass. Nothing durable is created: no files, caches or records, only an in-memory map and a child process reaped by `Wait`.

**Architectural notes for upcoming work:** once there is a `repo.ClaimLog` on `tracker.Repository`, other history-derived views (release times, handoff age) can share it instead of each re-running `git log`.

**Plan revision recommendations:** none. M4's Plan row is still unticked, as expected before this close, and the code matches the Spec line.

```findings
findings:
  - id: new
    severity: Important
    family: silent-error-in-io-glue
    title: |
      runState overwrites the claim-ages drift info line with detectDrift's result
    detail: |
      state.go:172-177 appends "claim ages unavailable" to s.Drift, then s.Drift = detectDrift(...) discards it, so a failure never shows. This is the 5th finding in this family. Rule: IO-glue probes return failures as values and the caller accumulates them, with one assignment to s.Drift and appends after it. Add a test asserting the info line appears when the tracker log read fails.
  - id: new
    severity: Important
    family: silent-error-in-io-glue
    title: |
      fillClaimTimes ignores git log's exit status, so a failed read silently yields no ages
    detail: |
      state.go:609-611 always discards cmd.Wait(). Only ignore it when ClaimTimes stopped early with every wanted card resolved; otherwise report it through the accumulating drift path (same rule as the finding above).
  - id: new
    severity: Minor
    family: external-call-outside-seam
    title: |
      fillClaimTimes runs git against the tracker ref outside tracker.Repository
    detail: |
      This is the 2nd finding in this family. Rule: git reads of tracker refs belong on tracker.Repository, which owns TrackingRef. Add repo.ClaimLog(ctx) and move ClaimLogFormat behind it.
  - id: new
    severity: Minor
    family: hand-maintained-token-set
    title: |
      claimKinds restates the owner-setting verbs that call sites mint as string literals
    detail: |
      Share named verb constants between the operationToken call sites and claimKinds, so a new owner-setting verb cannot silently drop out of claim ages (ARCH-DRY).
  - id: new
    severity: Minor
    family: duplicated-id-rendering
    title: |
      claimViews derives the short age by stripping claimAge's prose
    detail: |
      This is the 2nd finding in this family. Rule: each rendered fragment is formatted once and composed. Factor a shortAge(d) that claimAge wraps.
```

---

## Re-review — 2026-10-07T21:45:08-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | 6c682381ad458e8fc5075301a2d71cf72f5c6844..97dc25b9a2f2397e8a865fa4a3ebf5242b7ca55e |
| command | sdlc milestone-close --issue 284 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-10-07T21:45:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

**Verdict: FIX-THEN-SHIP.** All five round-1 findings are fixed, and the regression tests would fail without the fixes. M4 delivers what the Spec asks for: owner and claim age per issue, and claims grouped by slot and by operator. Claim times come from tracker history through `tracker.Repository`. The targeted tests pass (`tracker -run ClaimTimes`, `cmd/sdlc -run TestState|TestClaimViews…`). Two cheap Important gaps remain before closing:

- **Released cards are not shown.** The plan's ARCH-FUNERAL line says a released issue that nobody takes "stays visible in `state` (M4)". `state` never reads the card's `release` record.
- **A bad claimant line looks like "unowned".** `owned()` drops `CardClaimant`'s parse error, so a malformed claimant becomes an unowned issue with no warning. This is the 7th finding in the silent-error family.

### Strengths
- `tracker.ClaimTimes` (`claimtimes.go:39`) is a pure parser over a stream. It stops early, is fuzzed, and is tested for stopping early by counting bytes read.
- `HistoryStream`'s `finish(stoppedEarly)` (`trunkfile.go:475`) correctly separates "stopped because the answer was known" from "the read failed". This answers BR-33 at the seam, not at one call site.
- BR-32 is fixed structurally: `s.Drift` is assigned once and the claim-age failure is appended after it (`state.go:173-177`). `TestStateReportsAFailedClaimAgeRead` fails on the old order.
- The owner-setting verbs are now named constants (`tracker.OpClaim`/`OpReclaim`/`OpRelocate`). The token minters and `claimKinds` both use them, so a new verb can't silently drop out of claim ages.
- `shortAge` is formatted once; both `claimAge` and `claimViews` build on it.

### Critical
None.

### Important
1. **`state` doesn't surface `release`** (`state.go:270-279`, `owned`). The plan's ARCH-FUNERAL entry relies on `state` to show handoffs nobody took. Without it, a released card looks exactly like a card nobody ever claimed, and the leftover `release` record can't be found.
   - Fix: read `issue.CardRelease(rec.Card.Raw)` in `owned` and add a `Released` field to `IssueState` (JSON plus prose, e.g. `[released by <slot>, 2d]`).
   - Add a test for an unclaimed card that was released.
2. **`owned` drops `CardClaimant`'s error** (`state.go:275`). This is the 7th finding in `silent-error-in-io-glue`, so the fix should apply to the whole rule, not just this site.
   - The rule: every fallible read in `state`'s gather phase returns its error to the one drift accumulator. No `err == nil && …` guard may turn a failure into an absent value.
   - Sweep: the `CardClaimant` read and the new `CardRelease` read both report through drift, or set `Unreadable`.

### Minor
- **`sdlc state` can hang on a parser error** (`claimtimes.go:103`). If `ClaimTimes` fails partway (for example `bufio.ErrTooLong` on a commit message over 16 MB in the shared history), `finish(false)` calls `cmd.Wait()` before the pipe is drained. git then blocks on a full pipe and the command hangs. Fix: when `readErr != nil`, call `finish(true)` (kill) and return `readErr`.
- **A plan test is missing.** Task 13's "real git: two workspaces' claims grouped apart" isn't delivered: the grouping is only tested on a fixed `State`, and the real-git test uses one workspace.
- **Full-history reads.** An owned card with no claim-kind commit (for example a migrated card) makes every `state` run read the whole tracker history. So "cost bounded by the oldest live claim" holds only when every owned card has a claim commit. The plan's "one extra `git log`" estimate holds, but the doc comment overstates the bound.
- **`IssueOwner` copies `issue.Claimant`'s fields by hand.** Acceptable as a JSON view; note it if `Claimant` grows.

### Test coverage notes
- **BR-33:** `TestRepositoryClaimTimesReportsAFailedRead` deletes the tracking ref. Under the old code, which ignored `Wait`'s result, the read returned empty with no error, so the test would be red.
- **BR-32:** the injected `claimTimesOf` failure is reachable through `runState`.
- **Plan row "claim then set-status keeps the claim time":** covered by claim → start-plan in `TestStateReportsOwnerAndClaimTime`.

### Architecture
| Principle | Result | Notes |
|---|---|---|
| ARCH-DRY | pass | Verb constants and `shortAge` are each defined once. |
| ARCH-PURE | pass | Parser and `claimViews` are pure; IO goes through `HistoryStream` and the `claimTimesOf` seam. |
| ARCH-PURPOSE | flag | Release visibility was promised for M4 and isn't delivered (Important 1). |
| ARCH-MOCK | pass | Real-git fixture behind `tracker.Repository`. |
| ARCH-CONSTRAINTS | pass, with note | Full-history reads in the fallback case (Minor). |
| ARCH-SECURE | flag | A malformed claimant becomes "unowned" (Important 2); an oversized history record hangs `state` (Minor). |
| ARCH-ORDER | N/A | `state` keeps nothing between runs: it's a read-only snapshot. |
| ARCH-FUNERAL | flag | The `release` record's visibility path is missing (Important 1). The new code itself creates nothing durable. |

### Plan revision recommendations
- If two-workspace grouping stays tested only on a fixed `State`, add a `## Revisions` entry saying so.
- If release visibility is deliberately deferred, the plan's ARCH-FUNERAL line must say where released cards are surfaced instead.

```findings
dispose:
  - id: BR-32
    disposition: addressed
    note: |
      state.go:173-177 appends after the single detectDrift assignment; TestStateReportsAFailedClaimAgeRead injects the failure through claimTimesOf and asserts the drift line.
  - id: BR-33
    disposition: addressed
    note: |
      HistoryStream finish(stoppedEarly) judges the exit status on a full read; TestRepositoryClaimTimesReportsAFailedRead deletes the tracking ref and would pass silently under the old discard.
  - id: BR-34
    disposition: addressed
    note: |
      Repository.ClaimTimes over TrunkFile.HistoryStream (trackingRef owned by TrunkFile); state.go no longer runs git itself.
  - id: BR-35
    disposition: addressed
    note: |
      tracker.OpClaim/OpReclaim/OpRelocate are used at claim.go:224, reclaim.go:130, claimant.go:221 and in claimKinds.
  - id: BR-36
    disposition: addressed
    note: |
      shortAge (state.go) is used by both claimAge and claimViews.
findings:
  - id: new
    severity: Important
    family: artifact-without-removal
    title: |
      state never surfaces a card's release record, though the plan's ARCH-FUNERAL entry relies on it (M4)
    detail: |
      2nd in family. Rule: a persisted record whose removal waits on a human action must be visible on a read surface that names it. owned() in state.go should read issue.CardRelease and expose it in JSON and prose (released by slot, age), with a test for a released, untaken card.
  - id: new
    severity: Important
    family: silent-error-in-io-glue
    title: |
      owned() drops CardClaimant's parse error, so a malformed claimant renders as an unowned issue
    detail: |
      7th in family. Rule: every fallible read in state's gather phase returns its error to the one drift accumulator; no err==nil guard turns a failure into an absent value. Sweep the CardClaimant read and the new CardRelease read together.
  - id: new
    severity: Minor
    family: pipe-wait-before-drain
    title: |
      Repository.ClaimTimes calls finish(false), and so cmd.Wait, on a parser error before the pipe is drained, which can hang state
    detail: |
      A scanner error (e.g. ErrTooLong on a record over 16 MB in the shared tracker history) leaves git blocked on a full pipe. When readErr is non-nil, kill via finish(true) and return readErr.
  - id: new
    severity: Minor
    family: plan-test-not-delivered
    title: |
      Task 13's real-git two-workspace grouping test is missing; grouping is only tested on a fixed State
```

---

## Re-review — 2026-10-07T21:51:17-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | 6c682381ad458e8fc5075301a2d71cf72f5c6844..1a2bf5d2f0aefba72c49f09de70d72d618852214 |
| command | sdlc milestone-close --issue 284 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-10-07T21:51:17-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The two Important findings from earlier rounds are fixed, and each has a test that fails without its fix. For BR-37, `withOwnership` (`cmd/sdlc/state.go:86`) now reads `issue.CardRelease` for a card nobody owns. It shows the release in the JSON as `released` and in prose as `[released by …]`, plus a "Released, awaiting a claim:" section. A real-git handoff test that is never taken over (`stateclaims_test.go:119`) pins this. For BR-38, the read errors from both `CardClaimant` and `CardRelease` now go to `owner_error` and become a `warn` drift line. A malformed claimant shows as `[owner unreadable]`, not as unowned (`stateclaims_test.go:103`). `cardEnvelope` validates the release when it is read (`issue/handoff.go:108`), so the untrusted card data is checked when it is parsed. The BR-39 code fix is correct but has no test. BR-40 is delivered. All the targeted tests pass at HEAD: `go test ./internal/tracker -run ClaimTimes` and the five state tests in `cmd/sdlc`. Nothing blocks shipping.

1. **Strengths**
   - `withOwnership` is one helper used by both the card-only branch and the full-record branch of `listIssueStates` (state.go:332, 348). There is no copy of the owner logic.
   - The `tracker.OpClaim/OpReclaim/OpRelocate` constants are now the only source for both the tokens that call sites create and the `claimKinds` set, so the parser cannot lose a verb that a writer uses (BR-35 holds).
   - `HistoryStream` separates stopping early (kill the process, don't judge its exit) from reading to the end (git's exit status is checked). Combined with the new `finish(true)` on a read error, every exit path drains the pipe or kills git.
   - The pure parser `ClaimTimes` has a fixture test, a test that it stops reading early, and a fuzz target. The IO wrapper's failure path has a real-git test (`TestRepositoryClaimTimesReportsAFailedRead`).
   - `renderProseAt` takes `now` as a parameter, so the age rendering is deterministic under test.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - BR-39 has no regression test. A reader that returns an error mid-stream, or a `ClaimTimes` stub behind a seam, would pin that `finish(true)` runs on a read error. It is kept open as Minor below.
   - Plan Revision 1, item 4 says takeover mints `operationToken("takeover")` and that the parser matches `takeover`. In the code, takeover goes through claim's `cardsPublish` with `OpClaim`, and `claimKinds` has no `takeover` entry. The code is consistent with itself; the plan text is stale (see section 7).
   - If an owned card has no claim-kind commit in history (for example, an owner set before tracker operation trailers existed), the early stop never fires. `ClaimTimes` then reads the whole card history on every `state` call. The cost is small at today's history size, but the "bounded by the oldest live claim" claim is not strictly true (ARCH-CONSTRAINTS).

5. **Test coverage notes**
   - Covered:
     - JSON `owner`/`claimed_at` are present for a claimed card and absent for an unowned one.
     - The claim time is the claim commit's time, not a later `start-plan` write's.
     - The slot and operator views on a fixed `State`.
     - Two real-git workspaces group apart.
     - A handoff release and an open release are both visible.
     - A failed history read reaches drift.
   - The `CardRelease` error branch is covered only by sharing a code path with the claimant case. Adding a malformed `tracker.release` fixture would cost about one line.

6. **Architecture notes**
   - ARCH-DRY: pass. Owner rendering goes through `issueOwnerOf` and `IssueOwner.slot()`, and the age goes through `shortAge` and `claimAge`.
   - ARCH-PURE: pass. `ClaimTimes`, `claimViews`, `releaseTip` and `shortAge` are pure. The IO is in `Repository.ClaimTimes` and `fillClaimTimes`.
   - ARCH-PURPOSE: pass. Owner, age, both views and released-but-not-taken issues are all shown, which is the M4 purpose.
   - ARCH-MOCK: pass. git is reached only through the `gitx.TrunkFile` seam, and the tests run against real git fixtures. `claimTimesOf` is only a failure-injection seam.
   - ARCH-CONSTRAINTS: pass, with the Minor above.
   - ARCH-SECURE: pass. Card data is validated when parsed, and the git log record format is NUL/SOH-delimited and parsed defensively.
   - ARCH-ORDER: pass. `state` is read-only and holds no state between events, because it is a single snapshot read.
   - ARCH-FUNERAL: pass. Nothing durable is created. The one persisted record that waits on a human (`release`) is now visible on a read surface.
   - Docs: the `state.md` help and the atlas `sdlc-binary.md` state row are updated. No README surface was added.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry: "M4 — takeover publishes under `operationToken(tracker.OpClaim)` (it runs through claim), so `claimKinds` = {claim, reclaim, relocate, adopt}; Revision 1, item 4's separate `takeover` token was not adopted."

```findings
dispose:
  - id: BR-37
    disposition: addressed
    note: |
      withOwnership reads CardRelease into IssueState.Released (JSON + prose + "Released, awaiting a claim"); real-git test TestStateShowsReleasesAndGroupsTwoWorkspaces pins an untaken handoff.
  - id: BR-38
    disposition: addressed
    note: |
      CardClaimant and CardRelease errors both set OwnerError and a warn drift line; TestWithOwnershipReportsAnUnreadableOwner fails without it.
  - id: BR-39
    disposition: not-addressed
    note: |
      The code fix (finish(true) on readErr) is correct, but no test fails without it; a reader that errors mid-stream would pin it.
  - id: BR-40
    disposition: addressed
    note: |
      TestStateShowsReleasesAndGroupsTwoWorkspaces claims under two workspaces in real git and asserts two slot groups plus the "ariadne:7  #11" prose line.
findings:
  - id: new
    severity: Minor
    family: plan-test-not-delivered
    title: |
      Plan Revision 1 item 4 names a takeover operation token that the code does not mint
    detail: |
      Takeover publishes under OpClaim via claim.go, and claimKinds omits takeover. The code is consistent; the plan needs a Revisions entry so it stops claiming a separate takeover token.
  - id: new
    severity: Minor
    family: unbounded-history-scan
    title: |
      An owned card with no claim-kind commit makes every state call scan the whole card history
    detail: |
      The early stop needs every wanted path found. A pre-trailer owner never matches, so the bound "oldest live claim" does not hold. Small cost today.
```
