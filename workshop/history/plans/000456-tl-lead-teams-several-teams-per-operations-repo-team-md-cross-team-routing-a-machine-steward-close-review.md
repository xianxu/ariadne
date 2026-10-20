# Boundary Review — ariadne#456 (whole-issue close)

| field | value |
|-------|-------|
| issue | 456 — tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward |
| repo | ariadne |
| issue file | workshop/issues/000456-tl-lead-teams-several-teams-per-operations-repo-team-md-cross-team-routing-a-machine-steward.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5058591f86fce002653fdca0bf6e7f1b5c4460e6..1e324f8a89c20d0da7237b96906a37b05b81af46 |
| command | sdlc close --issue 456 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T14:05:31-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #456 diff: I've read the plan and the new tl-lead sources, and next I'm checking them against pair's `.hub` grammar and running the tests.

Still reviewing: checking how `sdlc state` reports issue owners, since fleet resume matches them against slot addresses.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix, then ship.** The revised scope (R1–R6) is delivered and tested: `team new`, `team list [REPO]`, `team issues` and `fleet resume [--apply]`, plus help and atlas. The package tests pass (`go test ./cmd/tl-lead ./cmd/tl-worker`: ok). `ownerOf` matches pair's `couchcore.Resolve` on every rank I checked: the hub's own `:0`, exact over repo, ties, unclaimed. `fleet resume` also keeps a failed `--actors` or `sdlc state` read as unknown. Two issues need fixing first:

- **Owner reads.** `fleet resume`'s `inFlight` handles per-issue owner reads that fail or don't match differently from the rule the plan states. An `owner_error`, or an owner that only gives a worktree path, is silently treated as "no work", so the slot is classed *idle*.
- **The `.hub` grammar.** The copy is advertised as "pinned by pair's own test vectors", but one of pair's vectors (non-UTF-8) was dropped. The UTF-8 check it enforces is missing from tl-lead.

Neither causes data loss, so neither blocks shipping, but both are cheap to fix. R7 (the full suite and the live dry runs from ops) is still unticked; the log says ops needs adopting with `team new` first.

### 1. Strengths
- **One ownership resolver.** `ownerOf` (`cmd/tl-lead/team.go:113`) is used by `renderTeams`, `planResume` and `ownedRepos`. I compared it against `~/workspace/pair/cmd/internal/couchcore/hub_ownership.go` `Resolve` and the ranks agree. Ties and unclaimed slots come back as such instead of being guessed (ARCH-DRY).
- **`fleet resume` holds the uncertainty rules for the whole-read cases.**
  - A failed `--actors` read makes every slot unknown, so nothing resumes (`fleet.go:153`).
  - An uncertain resume is reported and never retried (`fleet.go:255-262`).
  - The stateful `fakeCouch` turns a resumed slot live, so `TestFleetResumeApplyThenRerun` genuinely proves a rerun skips it.
- **`writeScaffold` checks before it writes.** It checks every file before writing any and writes with `O_EXCL`; reruns and refusals are both tested (`team_cmd.go:44`).
- **Tolerant decodes.** The `sdlc state` and `--resume-candidates` decoders ignore unknown fields, and the tests use trimmed live replies (`TestDecodeRepoIssues`, `TestDecodeCandidates`).
- **Help states the contract.** The tl-lead and tl-worker help match the revised design (hub contract, steward, routing, direct filing with `--no-claim`), and `main_test.go` and `rules.tsv` pin them.

### 2. Critical findings
None.

### 3. Important findings

**1. `inFlight` treats an unreadable or unmatched owner as "no work in flight"** (`cmd/tl-lead/fleet.go:116-129`, `planResume` at `:166`). ARCH-ORDER and ARCH-PURPOSE.
- **Where it breaks:**
  - `inFlight` checks only `is.Owner == addr`.
  - An issue in a non-terminal state with `OwnerErr != ""` is skipped, so its slot can come out as **idle** and be left down with no mention.
  - The same happens when `owner.workspace` is empty and `decodeRepoIssues` falls back to the worktree path, which never equals a `repo:N` address.
- **What the plan says:** plan-quality round 1 states "An owner workspace that isn't in the inventory is listed as unknown, never matched". The issue's Done-when says `fleet resume` "treats a failed read as unknown". `team issues` already follows this (its `unknown` bucket); `fleet resume` doesn't.
- **Fix sketch:** have `inFlight` return `(work, unknownReasons)`. If any non-terminal issue in that repo has an `OwnerErr` or an owner that isn't a `repo:N` address, class the down slots it could belong to as `classUnknown`, at least when there's no positive match. Add a fixture row: a `working` issue with `OwnerErr` in `/w/pair`, and assert that some slot is `unknown`, not `idle`.

**2. The mirrored `.hub` grammar drops pair's UTF-8 vector and check** (`cmd/tl-lead/team.go:58-62`, `team_test.go:14-20`). ARCH-SECURE and ARCH-DRY: the guarded copy has drifted.
- Pair's `TestHubClaimsParse` rejects `string([]byte{0xff})`, and `ParseHubClaims` refuses any `.hub` that isn't valid UTF-8.
- tl-lead's copy leaves out both the vector and the `utf8.Valid` check, though its comment says the vectors "are pair's own".
- **Effect:** `team list` and `fleet resume` read claims from a `.hub` that Couch refuses to start with.
- **Fix:** add `|| !utf8.Valid(raw)` to `parseHub`'s size check and restore the `0xff` vector.

**3. Matching owners by slot address ignores the machine** (`fleet.go:125`). ARCH-ORDER (an observation applied to the wrong entity).
- `sdlc state`'s `owner` carries `machine_name`, but `inFlight` matches on `workspace` alone. A `pair:3` claim made on another machine therefore counts as work in flight for this machine's `pair:3`.
- **Effect:** if a local conversation is proven, `--apply` resumes it.
- **Fix:** compare against this machine (`machine_name`, or better a machine id if sdlc exposes one) and class foreign-machine owners as not this slot's. Or, if sdlc only reports owners for this machine, add a comment saying so and a test that pins it.

### 4. Minor findings
- `execResume.Candidates` (`fleet.go:42`) has no timeout, unlike `Resume` and the SDLC seam's 60 s. A hung Couch hangs the dry run (ARCH-CONSTRAINTS).
- `execResume.Resume` (`fleet.go:75`) reports an error that isn't an exit error (e.g. `couch` not on PATH) as `refused: (none)`, losing the cause. It should include `err`.
- Run from a worktree slot (whose parent isn't the workspace), `team list` prints nothing at all. It should say "no teams found in DIR".
- `teamNew` doesn't check that the folder's base name is a valid Couch repo name (no whitespace, `:`, `#`), so `team.md` can name an unusable `X:0` address.
- `renderTeams`'s two-claimant warning covers only repo claims. Two hubs claiming the same exact `repo:N` aren't warned about.
- `ownedRepos` repeats `planResume`'s "sole owner is self" filter. It could share a predicate (ARCH-DRY, small).

### 5. Test coverage notes
- `TestFleetResumeDryRun`, `TestFleetResumeApplyThenRerun` and `TestFleetResumeActorsUnreadable` cover every class, refusal, uncertain outcome and the rerun skip, all against a stateful fake.
- **Missing:**
  - a per-issue `OwnerErr` or worktree-only owner in `fleet resume` (finding 1);
  - the non-UTF-8 `.hub` vector (finding 2);
  - a foreign-machine owner (finding 3).
- `Resume` has no live conformance check. That's acceptable: it's heavy and an external effect. `--resume-candidates` and `sdlc state` both have opt-in live decodes behind `TL_LEAD_LIVE`.

### 6. Architectural notes for upcoming work
- **ARCH-DRY:** pass. The grammar is a deliberate, guarded copy (fix its drift, finding 2), and there is one ownership resolver.
- **ARCH-PURE:** pass. `planResume`, `finishResume`, `ownerOf`, `parseHub`, `scaffold` and `unclaimedIssues` are pure; `fleetResume` and `teamNew` are thin glue.
- **ARCH-PURPOSE:** pass for R1–R6. R7 (the live dry runs) is still open, and the issue tracks it.
- **ARCH-MOCK:** pass. `fakeCouch` is stateful across `Actors`, `Candidates` and `Resume`. git runs for real in temp repos, as the plan says.
- **ARCH-CONSTRAINTS:** minor flag (no timeout on `Candidates`).
- **ARCH-SECURE:** flag (the UTF-8 gap, finding 2). Otherwise `.hub` and `--claim` are parsed into typed claims at the boundary.
- **ARCH-ORDER:** flag (findings 1 and 3). The rerun, uncertain and refusal handling is otherwise sound.
- **ARCH-FUNERAL:** pass. `team new` creates a repo whose retirement the help documents; nothing else durable is created.
- **ARCH-EVOLVE:** pass. `.hub` and `team.md` are never rewritten, and the decoders tolerate unknown fields.
- **Follow-up:** the planned `owner` field on `couch --actors --json` would make `ownerOf` only a fallback and settle the runtime-moves caveat.

### 7. Plan revision recommendations
- **If finding 1 is deferred:** add a "## Revisions" entry saying owners that fail to read or don't match a slot are not yet classed unknown in `fleet resume`. Otherwise the plan-quality round 1 line ("An owner workspace that isn't in the inventory is listed as unknown") claims something the code doesn't do.
- **If finding 2 is deferred:** correct the plan and atlas wording "pinned by pair's own test vectors" to note the omitted UTF-8 vector.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: failed-read-is-never-absence
    title: |
      fleet resume's inFlight treats an owner_error or worktree-only owner as no work, classing the slot idle
    detail: |
      cmd/tl-lead/fleet.go:116-129 matches only is.Owner == addr. A non-terminal issue with OwnerErr, or an owner whose workspace is empty (decodeRepoIssues falls back to the worktree path), never matches, so the slot is silently classed idle. Plan-quality round 1 says an unmatched owner is listed as unknown, never matched, and the Done-when says a failed read is unknown. Class such slots unknown and add a fixture with an OwnerErr row.
  - id: new
    severity: Important
    kind: defect
    family: mirrored-contract-drift
    title: |
      tl-lead's .hub grammar copy omits pair's UTF-8 check and its 0xff test vector
    detail: |
      pair couchcore.ParseHubClaims refuses non-UTF-8 input, and TestHubClaimsParse includes string([]byte{0xff}). cmd/tl-lead/team.go parseHub has no utf8.Valid check, and team_test.go dropped that vector while claiming the vectors are pair's own. Add the check and restore the vector.
  - id: new
    severity: Important
    kind: defect
    family: observation-bound-to-exact-entity
    title: |
      fleet resume matches issue owners to slots by workspace address only, ignoring the owner's machine
    detail: |
      sdlc state owner rows carry machine_name, but inFlight compares owner.workspace alone (fleet.go:125). A claim on another machine's pair:3 makes this machine's down pair:3 count as work in flight, and --apply resumes it when a local conversation is proven. Filter to this machine, or document and pin that sdlc reports only local owners.
  - id: new
    severity: Minor
    kind: defect
    family: bounded-external-calls
    title: |
      execResume.Candidates has no timeout, so a hung couch hangs fleet resume
    detail: |
      fleet.go:42 uses exec.Command without a context, unlike Resume (3 min) and the SDLC seam (60 s).
  - id: new
    severity: Minor
    kind: defect
    family: error-cause-preserved
    title: |
      execResume.Resume reports an exec failure that is not an exit error as "refused: (none)"
    detail: |
      fleet.go:75-77 drops err when stderr is empty (for example, couch not on PATH). Include err in the message.
  - id: new
    severity: Minor
    kind: defect
    family: empty-result-says-why
    title: |
      team list prints nothing when run from a checkout whose parent folder is not the workspace
    detail: |
      From a worktree slot, readHubs scans the slot folder, finds no .hub, and the command prints nothing. It should say no teams were found in that folder.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e73ca3aa-5f97-4d6a-b600-15c2586ad542","repository":"github.com/xianxu/ariadne","issue_id":"000456","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/d4423222b0f94ad0","reviewed_head":"1e324f8a89c20d0da7237b96906a37b05b81af46","slot":"ariadne:4","requested_at":"2026-10-19T21:02:09.491137Z"},"phase":"finished","pid":83725,"started_at":"2026-10-19T21:02:09.792951Z","finished_at":"2026-10-19T21:05:30.831598Z","session_id":"a75c1f88-ea07-4b8b-8e66-08a412a1f6c3","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T14:10:47-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 456 — tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward |
| repo | ariadne |
| issue file | workshop/issues/000456-tl-lead-teams-several-teams-per-operations-repo-team-md-cross-team-routing-a-machine-steward.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5058591f86fce002653fdca0bf6e7f1b5c4460e6..d7ec954cfbeb9bf0de22969bc899078c50dfff5b |
| command | sdlc close --issue 456 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T14:10:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Focus on the round-1 fix commit first.

I'm checking the round-1 fixes against the code. Next I'll run the tl-lead tests and revert-test the fixes.

Now revert-testing the BR-2 UTF-8 fix and the BR-1/BR-3 decode.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 1 got four of six findings fully right. BR-1 and BR-3 are fixed at the source: `fleet resume` no longer matches owners to slots itself. It now reads each slot's claims from `sdlc fleet inventory` rows, and sdlc places those claims (`PlaceClaims`) by this machine's fingerprint and the claimed worktree path. BR-2 and BR-6 are also fixed. I reverted each behavioral fix in the disposable checkout, and the matching test failed every time: the UTF-8 check fails `TestHubClaimVectors`, the `claims_state` mapping fails `TestDecodeFleetClaims`, and the `ClaimsKnown` guard fails `TestFleetResumeDryRun`. `go test ./cmd/tl-lead/ ./cmd/tl-worker/` and `go vet` pass at head.

BR-4 and BR-5 change the code correctly, but neither has a regression test, so I'm keeping them `not-addressed`. Both are Minor.

The new design brings one Important gap. The claims join reads only the slot's **host** row. sdlc's own slot model judges every member, dependency clones included, so a slot whose in-flight claim sits on a dependency clone is reported `idle` and left down.

1. **Strengths**
   - `cmd/tl-lead/sdlc.go:141-176`: the claims join is delegated to sdlc. tl-lead's owner-to-slot match (`inFlight`, `ownedRepos`) and the per-repo `sdlc state` reads in `fleet resume` are deleted (ARCH-DRY). That removes both the machine-blind match and the owner_error gap.
   - A slot with no inventory row, or a `claims_state` other than present or absent, becomes unknown and never idle (`sdlc.go:169-176`). `TestDecodeFleetClaims` covers all four cases: present, absent, unknown, and no row.
   - `execResume.Resume` now separates "did not run" (no `ExitError`) from "refused" and keeps the error text (`fleet.go:85-90`).
   - `planResume` remains a pure function that takes its reads already fetched (ARCH-PURE).

2. **Critical findings:** none.

3. **Important findings**
   - **A slot's claims are read from its host row only; dependency members are ignored** (`cmd/tl-lead/sdlc.go:169`). sdlc's `AssembleSlots` (`cmd/sdlc/internal/fleet/slots.go:213-238`) judges every member: the host plus each declared dependency clone, and each of those is also an inventory row. Consider a worker in `pair:3` that claims an ariadne issue in its slot's own ariadne clone. The claimant worktree is that clone, so the host row has no claims, and `fleet resume` prints `idle pair:3` and leaves the work down.
     - Fix: decode the slot members' paths. Set `Claims` to the union of non-terminal claims across all member rows. Set `ClaimsKnown` only if every member row is present or absent; a missing member row makes the slot unknown.
     - Add a fixture with a claim on a dependency member.
     - This is the 2nd finding in family `observation-bound-to-exact-entity`. The rule: a slot observation covers the slot's full member set, exactly as sdlc's slot verdict defines it, never just its host.

4. **Minor findings**
   - `decodeFleet` drops `claims_error`, so a stale or unknown slot prints only "claims unreadable: claims stale", without the reason sdlc gave. `decodeObservation` keeps `Card.Error`. This is the 2nd finding in family `error-cause-preserved`. The rule: when a decoded read-quality state is not an answer, carry its error text through to the printed reason. Sweep the decoders in `sdlc.go` and `team_issues.go`.
   - `decodeFleet` treats `stale` as unknown, while `decodeObservation` accepts `stale` as an answer. The fleet choice is conservative and defensible, but the two decoders in the same file disagree without a comment saying why.
   - `sdlc.go:153-156` hand-filters `done`, `wontfix` and `punt`. The inventory already guarantees that claims hold only lock-holding statuses (`validateClaims`), so this list is redundant and can drift from the vocabulary (ARCH-DRY, cosmetic).
   - The "workspace … (the parent of this checkout …)" line printed by `team list` (BR-6) has no test at the cobra layer; only the `renderTeams` empty case is pinned.

5. **Test coverage**
   - Fix-reverting checks went red for BR-1 (twice) and BR-2.
   - BR-3 is fixed by construction: by contract, the inventory contains only this machine's claims, and sdlc's `PlaceClaims` tests cover that.
   - BR-4 and BR-5 have no tests. For BR-5, a cheap test is `execResume{bin: "/nonexistent"}.Resume("x")`, asserting "did not run". BR-4 needs the timeout injected before it can be tested.

6. **Architecture**
   - **ARCH-DRY:** pass. The owner match was removed in favour of sdlc's join; the redundant terminal-status list is Minor.
   - **ARCH-PURE:** pass. `planResume`, `finishResume` and `decodeFleet` are pure.
   - **ARCH-PURPOSE:** flag. The host-only claims join is the Important finding above.
   - **ARCH-MOCK:** pass. Fake sdlc and fake couch sit behind the interfaces, with live decodes behind `TL_LEAD_LIVE`.
   - **ARCH-CONSTRAINTS:** pass. Every couch and sdlc call is now bounded, and candidate reads are sequential.
   - **ARCH-SECURE:** pass. `.hub` is size-bound and checked for UTF-8, and the inventory `schema_version` is checked.
   - **ARCH-ORDER:** pass. An uncertain resume is reported and never retried, and the rerun skips running slots.
   - **ARCH-FUNERAL:** pass. Only `team new` creates durable output (the ops repo), and that is intentional.
   - **ARCH-EVOLVE:** pass. tl-lead only reads these formats and decodes them tolerantly.

7. **Plan revisions:** add a Revisions entry recording that `fleet resume` takes claims from the inventory's per-member rows, and that it no longer uses `sdlc state` owner matching, which the plan's R5 still describes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Claims now come from inventory rows; a claims_state other than present or absent, or a missing row, is unknown. TestDecodeFleetClaims and TestFleetResumeDryRun go red when the fix is reverted.
  - id: BR-2
    disposition: addressed
    note: |
      utf8.Valid added in team.go parseHub; the 0xff vector is restored and TestHubClaimVectors goes red without the check.
  - id: BR-3
    disposition: addressed
    note: |
      tl-lead's owner match is deleted; sdlc PlaceClaims places only claims whose claimant is this machine's fingerprint and worktree (cmd/sdlc/internal/fleet/claims.go mine).
  - id: BR-4
    disposition: not-addressed
    note: |
      Candidates now uses CommandContext with a 1-minute timeout, but no test pins it and the timeout cannot be injected; make the timeout injectable and test it with a hanging fake binary.
  - id: BR-5
    disposition: not-addressed
    note: |
      The code fix is right (fleet.go:85-90, "did not run"), but there is no regression test; execResume with a nonexistent bin asserting "did not run" is cheap.
  - id: BR-6
    disposition: addressed
    note: |
      renderTeams prints "no teams" for an empty hub set (tested), and team list names the scanned workspace folder.
findings:
  - id: new
    severity: Important
    kind: defect
    family: observation-bound-to-exact-entity
    title: |
      fleet resume reads a slot's claims from its host row only, ignoring dependency members, so a slot with in-flight work on a dependency clone is classed idle
    detail: |
      This is the 2nd finding in this family. The rule: a slot observation covers the slot's full member set, as sdlc's AssembleSlots (slots.go:213-238) defines it, not just the host. decodeFleet (sdlc.go:169) joins only HostPath, but a claim made in a slot's own dependency clone is recorded on that clone's row. Fix: collect member paths, union their non-terminal claims, and set ClaimsKnown only when every member row is present or absent. Add a fixture with a claim on a dependency member.
  - id: new
    severity: Minor
    kind: defect
    family: error-cause-preserved
    title: |
      decodeFleet drops claims_error, so an unknown or stale slot prints no reason
    detail: |
      This is the 2nd finding in this family. The rule: when a decoded read-quality state is not an answer, carry its error text through to the printed reason, as decodeObservation does with Card.Error. Add claims_error to the Rows struct, include it in ClaimsWhy, and sweep the other decoders in sdlc.go and team_issues.go.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"056232a3-4732-4b9e-adad-58ce9774bffd","repository":"github.com/xianxu/ariadne","issue_id":"000456","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c7d8873c8e12dc96","reviewed_head":"d7ec954cfbeb9bf0de22969bc899078c50dfff5b","slot":"ariadne:4","requested_at":"2026-10-19T21:08:55.851753Z"},"phase":"finished","pid":60507,"started_at":"2026-10-19T21:08:56.011022Z","finished_at":"2026-10-19T21:10:46.126563Z","session_id":"943c5591-e926-424c-ab3e-ab9cc466a73e","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T14:13:22-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 456 — tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward |
| repo | ariadne |
| issue file | workshop/issues/000456-tl-lead-teams-several-teams-per-operations-repo-team-md-cross-team-routing-a-machine-steward.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5058591f86fce002653fdca0bf6e7f1b5c4460e6..5dbc3fc3c36276e7c15858a9de82eac60063cc88 |
| command | sdlc close --issue 456 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T14:13:22-07:00 |
| verdict | SHIP |

## Review

Prior findings center on fleet.go and sdlc.go. Let me read them.

Tests pass, and I've checked all four open findings against the code. Writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

All four open findings are fixed, and each fix has a regression test that pins it. `go test ./cmd/tl-lead/ ./cmd/tl-worker/` passes and `go vet` is clean at the pinned head. `decodeFleet` (`cmd/tl-lead/sdlc.go:115-197`) now builds a slot's claims from every member row, the host and its dependency clones. It marks claims as known only when every member row was read, and it carries `claims_error` through to the printed reason. Both couch calls now have a time limit, and their causes are told apart: a timeout is "uncertain", a missing binary is "did not run", and a refusal is "refused" with the error and stderr. I found one new issue, and it is Minor: a dependency that is declared but missing makes the whole slot "unknown". Nothing blocks SHIP.

**1. Strengths**
- `cmd/tl-lead/sdlc.go:173-194`: claims are joined over the slot's members in the same way sdlc's `AssembleSlots` builds them (`cmd/sdlc/internal/fleet/slots.go:206-240`). A missing row or a row that isn't `present`/`absent` sets `ClaimsKnown` to false, so a failed read is never taken as "no claims".
- `cmd/tl-lead/fleet.go:95-104`: an uncertain resume is reported and never retried. "Did not run" and "refused" stay separate errors, so the operator knows whether anything was attempted.
- `TestExecResumeBoundsAndCauses` (`fleet_test.go:257-280`) runs real scripts behind the seam (a hung couch, a missing binary, a refusing couch), with time limits a test can inject.
- `planResume` and `finishResume` are pure; all reads happen up front in `fleetResume`.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- `cmd/tl-lead/sdlc.go:179-182`: sdlc lists a declared but missing dependency clone (`MemberMissing`, `slots.go:232-234`) as a member with no tree row. `decodeFleet` reports that as "no inventory row" and sets `ClaimsKnown=false`. The slot is then classed `unknown` and `fleet resume` will never resume it, even when the host has work in flight. A checkout that doesn't exist holds no claims on this machine. The fix is to read the member's `verdict` (`missing`) from sdlc rather than infer it from a missing row join.

**5. Test coverage notes**
- `TestDecodeFleetClaims` covers: a present host claim, a claim only on a dependency clone (BR-7), `unknown` with its `claims_error` text (BR-8), and no row at all. Reverting either fix turns its assertion red: a host-only join returns no claims for `ariadne:2`, and dropping the error loses the `"unknown: tracker fetch failed"` text.
- No fixture covers the `stale` or `partial` claims states. Both are treated as unknown, which is the cautious choice.

**6. Architectural notes**
- ARCH-DRY: pass. The same `orDuration` default handles both call limits.
- ARCH-PURE: pass.
- ARCH-PURPOSE: pass. The BR-8 sweep held: `decodeObservation` keeps the card error and `decodeRepoIssues` keeps `owner_error`.
- ARCH-MOCK: pass. A stateful fake couch sits behind the `CouchResume` and `Actors` seams, and the exec layer is tested with real scripts.
- ARCH-CONSTRAINTS: pass. Every couch and sdlc call has a time limit, and resumes run one at a time.
- ARCH-SECURE: pass. Commands take argv arrays, and inventory JSON is decoded with a schema-version check.
- ARCH-ORDER: pass. Uncertainty is never collapsed into success or failure, and a rerun skips slots that are already running.
- ARCH-FUNERAL: pass. Nothing durable is created.
- ARCH-EVOLVE: pass. The reader is tolerant and never writes back.

**7. Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      fleet.go:55-71 runs Candidates under a context with a 1-minute default limit (injectable); TestExecResumeBoundsAndCauses times out a hung script.
  - id: BR-5
    disposition: addressed
    note: |
      fleet.go:99-103 reports a non-exit error as "did not run: <err>" and a refusal as "refused (<err>): <stderr>"; both paths are covered by the missing-couch and refuse-script cases.
  - id: BR-7
    disposition: addressed
    note: |
      sdlc.go:173-194 joins claims over every member row, with ClaimsKnown only when all are read; the ariadne:2 fixture holds a claim on a dependency clone only, and a host-only join would fail it.
  - id: BR-8
    disposition: addressed
    note: |
      claims_error is decoded (sdlc.go:121), added to the state, and printed through ClaimsWhy; the test asserts "unknown: tracker fetch failed". Sibling decoders already carry their errors (Card.Error, owner_error).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: consume-producer-verdict
    title: |
      A declared but missing dependency clone makes the slot's claims unknown, so fleet resume never resumes it
    detail: |
      sdlc's AssembleSlots (slots.go:232-234) lists a MemberMissing dependency with verdict "missing" and no tree row. decodeFleet (sdlc.go:179-182) reads the missing row as an unread member and sets ClaimsKnown=false. Fix: read members[].verdict and treat a "missing" member as holding no claims. Add a fixture.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2826d585-7d9a-4d9a-b623-450100fedbc5","repository":"github.com/xianxu/ariadne","issue_id":"000456","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/78fa6df45ade5e8f","reviewed_head":"5dbc3fc3c36276e7c15858a9de82eac60063cc88","slot":"ariadne:4","requested_at":"2026-10-19T21:12:14.493445Z"},"phase":"finished","pid":88166,"started_at":"2026-10-19T21:12:14.605416Z","finished_at":"2026-10-19T21:13:22.136908Z","session_id":"cdaaf6fa-a2e0-4170-8d2c-d8cb4fe632da","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T14:15:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 456 — tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward |
| repo | ariadne |
| issue file | workshop/issues/000456-tl-lead-teams-several-teams-per-operations-repo-team-md-cross-team-routing-a-machine-steward.md |
| boundary | whole-issue close |
| milestone | — |
| window | a95dd7e64f14479d666e0ec43f3a07cac3c50bbf..38dfba58988272731fcd1df39593a6cfda0b4311 |
| command | sdlc close --issue 456 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T14:15:25-07:00 |
| verdict | SHIP |

## Review

Checking the producer's JSON field name for the member verdict, then I'll write up the verdict.

```verdict
verdict: SHIP
confidence: high
```

The interdiff fixes the one open finding, BR-9, and adds nothing else. It changes two code files (`cmd/tl-lead/sdlc.go` and `cmd/tl-lead/fleet_test.go`) and adds two gate-record plans. `decodeFleet` now reads each member's `verdict` field from sdlc. A dependency that sdlc reports as `missing` no longer makes the slot's claims unknown. Any other member with no inventory row still does. I checked the producer side: `MemberVerdict` is embedded in `SlotMember` with the JSON tag `verdict` (`cmd/sdlc/internal/fleet/slots.go:52`), and `VerdictMissing = "missing"` (`slots.go:23`). sdlc sets that value only for `MemberMissing` dependencies (`slots.go:232-234`). A member sdlc can't place (`MemberOutside`) gets `unknown` instead, so it still falls into the "no inventory row → unknown" branch, which is the safe outcome. `go test ./cmd/tl-lead/` passes. The issue's Spec, Revisions and Done when are unchanged, and there is no conflict resolution in the window.

1. **Strengths**
   - `cmd/tl-lead/sdlc.go:182-184`: the fix takes sdlc's own verdict for the member instead of guessing from the missing row. It is also narrow: only `missing` is exempt, and every other row-less member still marks the slot unknown.
   - `cmd/tl-lead/fleet_test.go:231,253-257`: the new fixture `ariadne:5` has an in-flight host plus a missing dependency, and the test asserts `ClaimsKnown` and the host's claim. Without the new `case`, that member falls to `!ok`, sets `ClaimsKnown=false`, and the assertion fails. So the test catches the regression.
2. **Critical findings:** none.
3. **Important findings:** none.
4. **Minor findings:** fixture `ariadne:5` reuses `/w/ariadne-slot1` as its host path to borrow that row's claims. It works, but a dedicated row would be clearer.
5. **Test coverage notes:** the regression is pinned by the fixture above. The unknown paths (`ariadne:3`, `ariadne:4`) are still covered.
6. **Architectural notes**
   - **ARCH-DRY:** pass. One new switch case, nothing duplicated.
   - **ARCH-PURE:** pass. `decodeFleet` stays pure and is tested on raw JSON with no IO.
   - **ARCH-PURPOSE:** pass. The class this fix belongs to (`consume-producer-verdict`: take the producer's verdict instead of inferring it) is covered. The other row-less verdict, `unknown`/outside, stays conservatively unknown.
   - **ARCH-EVOLVE:** pass. Adding a field to a read-only decoder is additive, and the schema_version check is unchanged.
7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      sdlc.go:182 exempts members whose producer verdict is "missing" (slots.go:23,52,232-234); fixture ariadne:5 in fleet_test.go fails without it; go test passes.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a44f6401-e953-4c9f-a029-f69b61d4a91b","repository":"github.com/xianxu/ariadne","issue_id":"000456","recipe":"fix-review","round":4,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0368f9f381e8ec30","reviewed_head":"38dfba58988272731fcd1df39593a6cfda0b4311","slot":"ariadne:4","requested_at":"2026-10-19T21:14:29.4341Z"},"phase":"finished","pid":3419,"started_at":"2026-10-19T21:14:29.593548Z","finished_at":"2026-10-19T21:15:23.377892Z","session_id":"3dc0ec2d-37c9-4d73-af4f-657cc1a6f54e","outcome":"returned"}] -->
