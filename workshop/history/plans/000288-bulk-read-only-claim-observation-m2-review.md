# Boundary Review — ariadne#288 (milestone M2)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b0477c9861fb09d422858fe79a6a8fcb91024640..cf96c4556985958b3357a35507361cde94763b2e |
| command | sdlc milestone-close --issue 288 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T14:29:19-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 does what the issue asks for. `localMachine` gives `claim` and inventory one shared identity derivation. `PlaceClaims` is a pure join and is table-tested across read quality, identity availability and the kinds of claimant. The strict JSON contract rejects null lists, unknown states, an error on the wrong state, raw machine IDs and inactive claims. The recovery catalog entry and the atlas and help text are in place. Nothing blocks the boundary, but three Important gaps are cheap to fix first:

- **Wrong `dangling_claims`:** placement and dangling detection are keyed by the git checkout's identity, while claims belong to the tracker. Two separate clones of one repository therefore report each other's claims as dangling.
- **Untested branches:** `repoClaimsFrom` decides between partial, absent and unknown, and none of those branches is tested. The Done-when promises "partial … tested".
- **Undeclared network cost:** inventory now checks the remote tracker of every repository that has rows, one after another, with no stated limit on how long that can take.

### 1. Strengths
- **Pure core, thin shell (ARCH-PURE):** `PlaceClaims` (`internal/fleet/claims.go:96`) is a pure join with no IO. `collectClaims` (`inventory.go:146`) is a thin shell that loads each repository once and reads the identity once.
- **An unread value is never reported as "no claims":** rows without a claims read default to `unknown` "claims were not collected" (`types.go` `withDefaults`). A machine identity source that is nil, or that fails, makes every row `unknown`, never `present`.
- **One identity derivation (ARCH-DRY):** `localMachine` (`claimant.go:51`) feeds both `claim` and `fleetMachine`. The integration test checks that `machine.fingerprint` equals the claimant recorded on the real card.
- **`Records.Require` closes the M1 class at the rule level:** every reader that decides on a card field (close, actual, push, project status, issuefiles, the PR GitHub-link lookup, the not-done guard) now goes through one lookup. The not-done guard is newly proven by `expectDie` in `malformedcard_test.go:98`.
- **Contract test:** `TestInventoryClaimsContract` has one rejection per invariant, and each mutation is checked to have actually applied.

### 2. Critical findings
None.

### 3. Important findings
- **I-1 — Separate clones of one repository report each other's claims as dangling** (`claims.go:96-140`).
  - **Cause:** `placed` and `dangling` are keyed by `RepoIdentity`, which is the git directory. Each repository read returns every card on the tracker.
  - **Failure:** take two separate clones of one GitHub repository in the fleet, A and B. A claim made in B's tree is placed on B's row. A's pass doesn't find it in its own rows, so it also lists the claim as dangling. The help text says dangling means "a removed slot, or a checkout outside the fleet", which isn't true here. A consumer (pair#367) could then act on a live claim, for example by reclaiming it.
  - **Fix:** call a claim dangling only when its `claimant.worktree` matches no row's `tree_path` anywhere in the inventory, and report each claim at most once. Alternatively, key the join by the tracker's repository. Add a `PlaceClaims` case with two repository identities whose cards are the same.
- **I-2 — `repoClaimsFrom`'s quality branches are untested** (`issues.go:158-200`). **This is the 2nd finding in family `per-site-branch-untested`.**
  - **What's tested:** only the present and stale paths, end to end. `TestPlaceClaims` feeds hand-built `RepoClaims`, so the derivation itself is never exercised for:
    - an unreadable card leading to `partial` with its ref named;
    - a `CardClaimant` parse error leading to unreadable;
    - absent versus unknown, which is decided two ways;
    - stale and partial together;
    - skipping duplicates.
  - **Done-when:** bullet 3 promises partial is tested.
  - **Rule:** every pure function that maps an input state space to a quality value gets a table test over that state space. That is the class here; a test of the downstream consumer doesn't count.
  - **Fix:** a table test of `repoClaimsFrom`. If `tracker.Records` can't be built from fleet tests, add a narrow test constructor in `tracker`. Also extend `TestOneMalformedCardDoesNotBlockOthers` to assert `fleet inventory` reports `claims_state: partial` naming #42.
  - **Integration gap:** the Done-when's "other-machine omitted" and "unclaimed omitted" cases are tested only in the pure test, not against the real tracker fixture.
- **I-3 — New network check per repository, one after another, with no limit (ARCH-CONSTRAINTS)** (`inventory.go:154-159`).
  - **Before:** tracker reads ran only for repositories with an issue-prefixed branch.
  - **Now:** every repository with rows triggers `repoRecords`, then `LoadRecords(PreferFresh)`, then `Presence()`, which checks the remote. That includes brain repositories on gcrypt remotes, as the log's "brain repos `unknown`" shows.
  - **Plan gap:** the plan budgets the number of reads ("one per repo") but not latency or which remotes get touched.
  - **Fix:** either decide `absent` locally first (no `workshop/issues/`, no tracker marker and no fetched tracking ref means no remote probe), or state the latency budget and bound each probe with a timeout through `ctx`.

### 4. Minor findings
- **`observe.go:95`** still reads `rec.Status()` through `Get`. For an unreadable card that is `""`, so `collectEvidence` takes the non-`done` evidence path. The card section already says `unknown`, so this is low impact. It is the 4th instance in family `unreadable-card-read-as-absent`, and the rule already exists (`Records.Require`); observe should reuse the card-error branch from line 49. A structural guard would close the class: have `IssueRecord.Status()` report unknown, or forbid `Get(...).Status()` in a lint test.
- **`types.go` `Inventory.MarshalJSON`:** `i.Rows[n] = row.withDefaults()` writes into the caller's slice, since only the struct is copied, not the backing array. The old code copied the row first. Marshalling should not mutate its input; copy the slice.
- **`claims.go:218` `fingerprintPattern`** duplicates `issue.fingerprintRE`. Export a validator from `issue` (ARCH-DRY).
- **`Machine.State` reuses the `Claims*` constants** (present/unknown). This couples two different vocabularies; consider dedicated constants.
- **When a read is both stale and partial, the state is `partial`.** Stale shows up only in `claims_error`, so a consumer can't key on it.
- **`ClaimAssociation.Ref` is built from `filepath.Base(repoRoot)`,** so two clones with different directory names give different refs for the same issue. This was inherited from `LookupRepoIssues`; it matters together with I-1.

### 5. Test coverage notes
- `TestPlaceClaims` covers the full cross product the plan names, and it checks the contract on every row. Good.
- The integration tests cover placed, dangling and stale on the real fixture, and the counting wrapper proves one claims lookup per repository. That wrapper counts `LookupClaims` calls, not `repoRecords` loads; the two share the cache key `repoRoot`, so this holds.
- Not covered: the `repoClaimsFrom` derivation (I-2), separate clones (I-1), and the `Inventory.MarshalJSON` mutation.

### 6. Architecture principles
- **ARCH-DRY:** pass. `localMachine` and `Require` are each one source. One Minor: the fingerprint validator is duplicated.
- **ARCH-PURE:** pass. The one blemish is the marshal mutation (Minor).
- **ARCH-PURPOSE:** pass, with one gap. Claims, `machine`, `dangling_claims`, quality and the catalog entry are all delivered. The "partial … tested" part of the Done-when is only half delivered (I-2).
- **ARCH-MOCK:** pass. Tests use the real stateful tracker fixture through the production `LookupRepoClaims` seam, and the identity source is injected.
- **ARCH-CONSTRAINTS:** flag (I-3).
- **ARCH-SECURE:** pass. The claimant record comes from the remote tracker and is validated by `parseClaimant`. The worktree path is only canonicalized (a stat) and is quoted in text output, and a parse failure shows up as unreadable/partial.
- **ARCH-ORDER:** pass. The command reads each tracker once per repository and keeps no state between events. A racing claim is simply observed as of that read.
- **ARCH-FUNERAL:** pass. Nothing durable is created beyond the fetch's remote-tracking ref, which already existed.

### 7. Plan revision recommendations
- Add a Revisions entry recording that `repoRecords` is now called for every repository with rows, not only those with issue-prefixed branches, along with the latency or probe bound chosen for I-3.
- If I-1 is fixed by tracker-keyed or global-path dangling, update the Decisions "Placement" bullet, which currently says a claim matching no row in the same `repo_identity` is dangling.

```findings
findings:
  - id: new
    severity: Important
    family: tracker-entity-keyed-by-checkout
    title: |
      dangling_claims keyed by git checkout identity; separate clones of one repo report each other's live claims as dangling
    detail: |
      PlaceClaims keys placed/dangling by RepoIdentity (the git dir) while each repo read returns every tracker card; a claim placed on clone B's row is also emitted as dangling under clone A. Fix: dangling only when the claimant worktree matches no row inventory-wide, or key the join by tracker repository; add a two-identity PlaceClaims case.
  - id: new
    severity: Important
    family: per-site-branch-untested
    title: |
      repoClaimsFrom's partial/absent/unknown/duplicate branches are untested; Done-when promises partial is tested
    detail: |
      2nd finding in this family. Rule: every pure function mapping an input state space to a quality value is table-tested over that state space, not only through its consumer. TestPlaceClaims feeds hand-built RepoClaims; add a repoClaimsFrom table test and assert claims_state partial naming 42 in TestOneMalformedCardDoesNotBlockOthers.
  - id: new
    severity: Important
    family: undeclared-io-fanout
    title: |
      inventory now probes every repository's tracker remote one after another with no latency bound (ARCH-CONSTRAINTS)
    detail: |
      Previously only repos with issue-prefixed branches triggered repoRecords; now every repo with rows runs LoadRecords(PreferFresh) and Presence(), which checks the remote, including gcrypt brain remotes. Decide absent locally first, or declare a budget and bound each probe via ctx.
  - id: new
    severity: Minor
    family: unreadable-card-read-as-absent
    title: |
      observe.go:95 reads rec.Status() via Get; an unreadable card's empty status selects the non-done evidence path
    detail: |
      4th instance in this family. The rule exists (Records.Require); make it structural, e.g. IssueRecord.Status reports unknown, or a lint forbidding Get(...).Status().
  - id: new
    severity: Minor
    family: marshal-mutates-input
    title: |
      Inventory.MarshalJSON writes defaults into the caller's Rows backing array
  - id: new
    severity: Minor
    family: duplicated-validator
    title: |
      fleet fingerprintPattern duplicates issue.fingerprintRE; export one validator from issue
```

---

## Re-review — 2026-10-02T14:38:08-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b0477c9861fb09d422858fe79a6a8fcb91024640..8c3f40f0889f1097735b7eef823e58b0b9ada4e1 |
| command | sdlc milestone-close --issue 288 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T14:38:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Commit `8c3f40f0` fixes BR-9, BR-10, BR-12, BR-13 and BR-14 correctly, and each fix has a test that would fail without it. The build and vet are clean, and the `fleet`, `tracker` and `issue` packages pass, as do the targeted `cmd/sdlc` tests (`TestFleetInventory*`, `TestOneMalformedCard*`, `TestRecovery*`). One finding stays open. BR-11's fix (a checkout with no tracker is now judged `absent` locally, without contacting its remote) is a behavior change, and no test covers it: you could delete the `CutOver` early return and every test would still pass. I'm not raising anything new. Add that test and this ships.

1. **Strengths**
   - `PlaceClaims` (`internal/fleet/claims.go`) now judges a claim dangling only when its worktree matches no row anywhere in the inventory, and reports each tracker issue once via `trackerIssueKey`. That key is safe because the claimant's `Repository` field is required by `parseClaimant`. `TestPlaceClaimsAcrossClones` would report three dangling claims under the old code.
   - Unexporting `Records.get` enforces the rule in the API itself: code outside the package can only look a card up through `Require`. I checked the remaining bypass, `All()`. All four of its callers check `CardErr` before reading `Status()`: `trackercompletion.go:38`, `fleet/issues.go:43,176` and `state.go:242`.
   - `TestRepoClaimsFrom` table-tests the whole state space (tracker presence × freshness × unreadable cards × claimant × duplicates) through the exported pure `ComposeRecords`, with no IO.
   - `MarshalJSON` now copies the rows before filling defaults. `TestInventoryMarshalDoesNotMutate` fails against the old code, which wrote into the caller's backing array.
   - `issue.ValidFingerprint` is now the only fingerprint validator. `machineIDFileRE` in `claimant.go:105` checks a different thing (the raw machine-ID file format), so it isn't a duplicate.

2. **Critical:** none.

3. **Important:** BR-11 remains open (details in the findings block). Fix sketch: in `fleetclaims_test.go`, build a checkout with no cutover marker whose `origin` is an unreachable path. Assert `ClaimsAbsent` and an empty `ClaimsError`. Without the early return, that read comes back `unknown` ("could not confirm…").

4. **Minor**
   - `All()` still hands back unreadable records. Today every caller checks `CardErr`, so that half of the BR-12 rule rests on convention, not the API. This doesn't block.

5. **Test coverage:** BR-9, BR-10 and BR-13 each have a regression test that would fail without the fix. BR-10 also added a `partial`-naming-`#000042` assertion to `TestOneMalformedCardDoesNotBlockOthers`. The gap is BR-11, above.

6. **Architecture**
   - ARCH-DRY: pass. The fingerprint check is now one function.
   - ARCH-PURE: pass. `ComposeRecords`, `repoClaimsFrom` and `PlaceClaims` are pure; `LookupRepoClaims` is the thin IO layer around them.
   - ARCH-PURPOSE: pass. Every item in Done-when is delivered.
   - ARCH-MOCK: pass. Tests run against a real git tracker fixture.
   - ARCH-CONSTRAINTS: flagged under BR-11. Each repository that uses the tracker still gets one remote fetch, one after another, with no latency budget declared. That's acceptable if the plan says so; it currently doesn't.
   - ARCH-SECURE: pass. Malformed cards are quarantined and surface as `partial`/`unknown`.
   - ARCH-ORDER: pass, with a note. The command holds no state between events because it is a single read.
   - ARCH-FUNERAL: pass, with a note. It creates nothing durable because its output is printed JSON and an in-process cache.

7. **Plan revisions:** add a line to the M2 Revisions entry. It should say that inventory contacts the remote only for repositories that use the tracker, once each and in sequence, and that no per-probe timeout is set.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      PlaceClaims marks a claim dangling only when its worktree is no row anywhere and dedupes by trackerIssueKey; TestPlaceClaimsAcrossClones fails on the old per-identity key.
  - id: BR-10
    disposition: addressed
    note: |
      TestRepoClaimsFrom table covers absent/unknown x2/present/stale/partial/stale+partial plus duplicate; malformedcard_test asserts partial naming #000042.
  - id: BR-11
    disposition: not-addressed
    note: |
      The CutOver early return in LookupRepoClaims (issues.go:147) has no regression test; removing it leaves the suite green. Add an untracked checkout with an unreachable origin and assert absent with no error. Tracked repos still fetch one after another with no declared budget; record that in the plan.
  - id: BR-12
    disposition: addressed
    note: |
      Records.Get is now unexported (records.go:112), so outside callers must use Require; observe.go:97 uses Require and skips Status on cerr. All() callers each check CardErr.
  - id: BR-13
    disposition: addressed
    note: |
      MarshalJSON copies rows (types.go:420); TestInventoryMarshalDoesNotMutate fails against the old in-place write.
  - id: BR-14
    disposition: addressed
    note: |
      issue.ValidFingerprint is exported and used by fleet; the local fingerprintPattern is deleted.
```

---

## Re-review — 2026-10-02T14:40:10-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b0477c9861fb09d422858fe79a6a8fcb91024640..d2e7f7f8226ec01ae4d21d72dc8123b132be392f |
| command | sdlc milestone-close --issue 288 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T14:40:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This boundary meets M2's Done-when. `fleet inventory` puts this machine's active claims on each worktree row, using one cached tracker read per repository. Claims whose worktree is gone appear once in `dangling_claims`, even when several clones read the same tracker. Every row says how far its claims can be trusted (`claims_state`), so an empty list never means "no claims" unless the read actually succeeded. Claim and fleet now get the machine identity from one shared function, `localMachine`. The catalog entry is registered as read-only, with proofs.

The one open finding, BR-11, is addressed. `LookupRepoClaims` now checks the local cutover marker and fetched-tracker ref before it goes near a remote (`tracker.CutOver`). So a repository that doesn't use the tracker costs no network. `TestFleetClaimsSkipUntrackedRemotes` points such a checkout at a remote that doesn't exist. Without the fix, `Presence()` would return stale with no fetch error, and the test would get `unknown` instead of `absent`, so it really does fail without the fix. Nothing left blocks; two Minors remain.

1. **Strengths**
   - `PlaceClaims` (`internal/fleet/claims.go`) is pure. Its test covers every combination of read quality, identity availability and claimant type, and checks each output row against `validateClaims`.
   - `repoClaimsFrom` is tested through the now-exported pure `tracker.ComposeRecords`. That covers BR-10's partial, stale-and-partial, unconfirmed and duplicate cases without IO.
   - Unexporting `Records.Get` so only `Require` remains turns "an unreadable card is unknown" into something the compiler enforces. That closes the `unreadable-card-read-as-absent` family as a whole, not one call site at a time (`records.go:109-129`).
   - `withDefaults` makes a row or inventory built without claims marshal as `unknown`/"not collected", never as "present, none". It also copies the rows instead of writing into the caller's (BR-13).
   - `TestFleetInventoryPlacesClaims` is end to end. It claims in a real slot, counts one read per repository across three worktrees, removes the slot, and checks the claim then shows as dangling.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The plan's network budget says each fetch is "bounded by Git's own transport timeout". That isn't backed by anything: `Repository.Presence()`/`Initialized()` take no `ctx`, and git/ssh have no overall fetch timeout by default. So one tracked repository with a hung remote stalls the whole inventory. Either thread `ctx` with a deadline into the presence fetch, or reword the budget to say it is unbounded per fetch (ARCH-CONSTRAINTS).
   - `render.go` `renderMachineClaims` compares `m.State == ClaimsPresent`, and `fleetclaims_test.go` checks `inv.Machine.State != fleet.ClaimsPresent`. Both should use `MachinePresent`. The strings match today, so it's only a type-confusion nit.

5. **Test coverage notes:** Coverage is good. Stale is tested end to end by renaming the origin; partial and unknown are pure-tested. Nothing tests a hung tracked remote (see the first Minor).

6. **Architectural notes**
   - **ARCH-DRY:** pass. `localMachine` is the one identity derivation, and `issue.ValidFingerprint` is the one validator (BR-14).
   - **ARCH-PURE:** pass. Placement and quality derivation are pure; `LookupRepoClaims` is a thin IO wrapper.
   - **ARCH-PURPOSE:** pass. Every Done-when bullet is delivered.
   - **ARCH-MOCK:** pass. Tests run against real git fixture trackers.
   - **ARCH-CONSTRAINTS:** pass, apart from the Minor above.
   - **ARCH-SECURE:** pass. Malformed claimants are quarantined as unreadable, and only the fingerprint is published, never the raw machine ID.
   - **ARCH-ORDER:** N/A. The inventory holds no state between events: it is a single-shot read, and its process-lifetime records cache is invalidated by tests.
   - **ARCH-FUNERAL:** pass. It creates nothing durable beyond the remote-tracking ref the fetch already maintained.
   - **For pair#367:** if a large fleet needs parallel fetches later, that work should also add the `ctx` deadline.

7. **Plan revisions:** the "Network budget" paragraph should either drop "bounded by Git's own transport timeout" or replace it with the actual deadline mechanism.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      LookupRepoClaims checks the local CutOver signal (marker or fetched ref) before any remote probe, so untracked repos cost no network; TestFleetClaimsSkipUntrackedRemotes would read unknown without the check. Budget declared in the plan.
findings:
  - id: new
    severity: Minor
    family: undeclared-io-fanout
    title: |
      Plan budget says tracker fetches are bounded by Git's transport timeout, but Presence takes no ctx and git fetch has no default deadline
    detail: |
      This is the 2nd finding in family undeclared-io-fanout. The rule: a network call on a read-only view path either takes the caller's ctx deadline or its budget says it is unbounded. Thread ctx into Repository.Presence/Initialized, or correct the plan's wording.
  - id: new
    severity: Minor
    family: cross-enum-constant
    title: |
      renderMachineClaims and TestFleetInventoryPlacesClaims compare Machine.State against ClaimsPresent instead of MachinePresent
```
