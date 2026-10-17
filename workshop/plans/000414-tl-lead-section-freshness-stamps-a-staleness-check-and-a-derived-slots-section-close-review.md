# Boundary Review — ariadne#414 (whole-issue close)

| field | value |
|-------|-------|
| issue | 414 — tl-lead: section freshness stamps, a staleness check, and a derived Slots section |
| repo | ariadne |
| issue file | workshop/issues/000414-tl-lead-section-freshness-stamps-a-staleness-check-and-a-derived-slots-section.md |
| boundary | whole-issue close |
| milestone | — |
| window | e1bb1ba557be3270d16129e8d9dd311cf1b22ccd..78490f1455b440e25bf2c8f6848e9eb002e00826 |
| command | sdlc close --issue 414 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T17:42:19-07:00 |
| verdict | SHIP |

## Review

Running the tl-lead test suite and vet, then checking a few edge cases in the stamp parser.

```verdict
verdict: SHIP
confidence: high
```

I recommend shipping ariadne#414. Everything the plan checks off is in the code. `writeSection` is still the only write path, and it now owns the stamp: it strips a leading stamp before `compute` sees the body and puts a fresh one on what it writes. So `section`, `thread` and `slots` all stamp, and a stamp can never be doubled. `status check` uses a small budget table. `submit`, `owed` and every write end with the same check lines. Slots gets each slot's agent and idle state through a new `CouchActors` seam, and shows `?` when couch can't be read. The atlas and `help.md` are both updated. `go test ./cmd/tl-lead/` and `go vet` pass at the head commit.

The issue's Spec asks for three things that were changed or moved elsewhere, and the TL approved all three in the Log:
- **Stale list in the ping:** the couch-side ping is a pair change, filed separately.
- **`--refresh` flag:** not added; the existing `status slots` verb does the job.
- **Needs you:** no time budget; it's flagged only when it has no stamp.

None of my findings blocks the ship. All are Minor.

**Strengths**
- `cmd/tl-lead/status.go:41-50`: putting stamp handling inside `writeSection` covers every writer from one place. The compare-and-replace hash is still taken from the raw read, so the guard isn't weakened.
- `cmd/tl-lead/stamp.go:19-31`: the stamp parser is pure and tolerant. A missing or malformed stamp counts as "unknown age" rather than an error, and `TestSplitStamp` covers the edge cases.
- `cmd/tl-lead/freshness.go:24-29`: the budget table is one source of truth. The help text, atlas and `check --help` describe it consistently.
- `cmd/tl-lead/trains_test.go:894-913`: `TestGatherSlots` now checks the whole fixture-fleet output exact, instead of `Contains`, plus a degraded run with couch unreadable. That meets the Done-when item.
- `cmd/tl-lead/freshness_test.go:126-164`: `TestGitNewestLanding` runs real git on temporary repos and checks that brain repos and non-SDLC directories are excluded.

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
1. **Stamp only has minute precision** (`freshness.go:187`, `stamp.go:37`). The comparison is `land.At.After(at)`, where `at` is cut to the minute and the landing time keeps its seconds. Example: a landing at 17:05:10 already in local refs, then a `slots` write at 17:05:40. The write is stamped 17:05, so the same write's own output reports Slots stale. Fix: compare `land.At.Truncate(time.Minute).After(at)`.
2. **Missing-section check relies on couch's error wording** (`freshness.go:173`). It matches the text "heading line, found 0". If couch rewords that message, every check prints an "unread" line for missing sections. That failure is visible, not silent, but the live conformance test (`conformance_test.go`) doesn't pin the wording. Fix: add a missing-section read to `TestSessionSyncConformance`.
3. **Leaving a field empty silently uses the real thing** (`main.go:508-510`, `freshness.go:225-230`). Most existing tests (e.g. `status_test.go:60,298`) build `deps{ss: f}` with no `landing` or `now`. Those writes then call the real `gitNewestLanding` and the wall clock. Fix: set the production values explicitly in `main()` and treat an empty field as "no landing" or a fixed clock.
4. **`couch --actors` has no live check** (ARCH-MOCK). `TestDecodeActors` uses a reply trimmed from a live run, but nothing re-checks that shape against the real couch over time. Fix: add an `Actors()` call to the existing `TL_LEAD_LIVE` conformance test.
5. **More couch reads per event than the plan says** (ARCH-CONSTRAINTS). The plan says "one couch read", but `check` does one session-sync read per budgeted section (4), plus one git call per sibling repo. It's still cheap and inventory-free; the plan wording is just off.
6. **Slots label differs from the plan.** The plan's example line shows `claude·working`; the code shows `claude` (working) and `claude·idle`. `help.md` documents what the code does; only the plan text is out of date.
7. **One slot, several actors:** `decodeActors` keeps whichever row comes last and ignores the `Known`/`Pending` fields. That's fine for now; it should be documented if couch can ever report more than one actor per slot.
8. **Two error paths have no test:** the "newest landing unknown" line, and a section that fails to read for some other reason than being missing.

**Test coverage**

Every Done-when test exists:
- **A write stamps its section:** `TestWriteStamps`.
- **A deleted stamp comes back:** also `TestWriteStamps`, which starts from an unstamped "Needs you".
- **A stamp in a body file isn't doubled:** also `TestWriteStamps`.
- **A thread upsert keeps one stamp:** also `TestWriteStamps`.
- **The operator's section is never stamped:** also `TestWriteStamps`.
- **A stamp isn't counted as a bullet:** `TestStampIsNotABullet`. The bullet parser already ignored lines before the first `N. `, so this test pins existing behaviour rather than new code.
- **Stale by age, stale by landing, unstamped, fresh, and `owed` ending with the check lines:** `TestStatusCheck`.
- **Fixture fleet renders exactly:** `TestGatherSlots`.

Removing the stamp logic from `writeSection` would make `TestWriteStamps` fail.

**Architecture**
- **ARCH-DRY: pass.** Slots extends the existing renderer instead of adding a new one, and the stamp lives in one place.
- **ARCH-PURE: pass.** `splitStamp`, `stamped`, `agentLabel`, `renderSlots` and `hm` are pure. The clock, landing source and actors read come in through `deps`. Finding 3 is the one weak spot.
- **ARCH-PURPOSE: pass.** Moving the ping to pair is a separate change in another repo, approved by the TL. In ariadne, every event the TL handles carries the stale list.
- **ARCH-MOCK: pass with a note.** The actors fake is behind the same seam production uses, but there's no live check (finding 4).
- **ARCH-CONSTRAINTS: pass.** The per-event check doesn't touch the slow inventory, and the live Log measured 37 s only for `slots` (finding 5).
- **ARCH-SECURE: pass.** A stamp the operator edits by hand falls back to "unstamped". The couch JSON is checked for `Code == "ok"`, and failures show as `?`.
- **ARCH-ORDER: pass.** Nothing new is held between events. The retry loop recomputes the stamp on every attempt.
- **ARCH-FUNERAL: pass.** No new files, rows or caches; each section's stamp is overwritten on every write.

**Notes for upcoming work**
- The planned "landed since the last update" section should be another row in `renderedSections`, so it stamps through `writeSection` automatically.
- If more inputs need defaults, give `deps` real production values instead of more "empty field means real thing" checks.

**Plan revisions**
- A `## Revisions` entry noting:
  - `check` does one couch read per budgeted section, not one in total.
  - The Slots agent label is `agent` / `agent·idle`, and the free line now shows each free slot's agent.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: timestamp-precision-mismatch
    title: |
      Landing rule compares a minute-truncated stamp against a second-precision landing time
    detail: |
      freshness.go:187 land.At.After(at), where at is parsed from a minute-resolution stamp; a landing earlier in the same minute as the write flags the just-written section stale. Compare land.At.Truncate(time.Minute).After(at).
  - id: new
    severity: Minor
    kind: defect
    family: unpinned-external-contract
    title: |
      Missing-section detection matches couch's error text with no conformance check pinning it
    detail: |
      freshness.go:173 substring-matches "heading line, found 0"; add a missing-section read to TestSessionSyncConformance (TL_LEAD_LIVE) and an Actors() call for the couch --actors JSON shape.
  - id: new
    severity: Minor
    kind: defect
    family: nil-seam-defaults-to-production
    title: |
      Empty deps.landing and deps.now fall back to real git and the wall clock in tests
    detail: |
      Most existing tests build deps{ss: f}, so every write's printStale runs gitNewestLanding against the temp dir. Set the production values explicitly in main() and make an empty field a no-op.
  - id: new
    severity: Minor
    kind: defect
    family: plan-text-drift
    title: |
      Plan says check is one couch read and shows a claude·working label; code does 4 reads and renders claude / claude·idle
    detail: |
      Add a Revisions entry so the plan matches the delivered behavior.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"56c4edc5-aa25-44aa-969d-98da2848a151","repository":"github.com/xianxu/ariadne","issue_id":"000414","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/f0defb8211d44440","reviewed_head":"78490f1455b440e25bf2c8f6848e9eb002e00826","slot":"ariadne:4","requested_at":"2026-10-17T00:40:10.461601Z"},"phase":"finished","pid":56019,"started_at":"2026-10-17T00:40:10.570033Z","finished_at":"2026-10-17T00:42:18.980192Z","session_id":"b998f0b5-b665-4faf-ad0d-3f1ca29aaa6b","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T17:44:41-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 414 — tl-lead: section freshness stamps, a staleness check, and a derived Slots section |
| repo | ariadne |
| issue file | workshop/issues/000414-tl-lead-section-freshness-stamps-a-staleness-check-and-a-derived-slots-section.md |
| boundary | whole-issue close |
| milestone | — |
| window | e1bb1ba557be3270d16129e8d9dd311cf1b22ccd..2059abecb29b36c46b1e01b733adba7b26abfb87 |
| command | sdlc close --issue 414 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T17:44:41-07:00 |
| verdict | SHIP |

## Review

Prior-round fixes check out on read and the tests pass. Next I'm reverting the BR-1 fix to confirm its regression test goes red.

```verdict
verdict: SHIP
confidence: high
```

All four findings from round one are fixed, and I found no new ones. I re-checked each fix against the code. The BR-1 regression test fails when the fix is reverted. The tl-lead package tests pass at the pinned head (`go test ./cmd/tl-lead/`). The issue promised four things — section stamps through the one write path, a light `status check`, the stale lines appended to every status verb, and agent and idle labels in Slots from couch — and all four are delivered and covered by tests. The atlas and help.md cover the new `check` verb, the stamps and the couch actors input. The plan's Revisions entry now matches what the code does.

**1. Strengths**
- **One stamp owner.** `writeSection` (`cmd/tl-lead/status.go:41-50`) removes the old stamp before `compute` runs and adds a fresh one after. So `section`, `thread` and the rendered sections can't produce a doubled stamp, and Operator instructions is left out through `stampedSection`.
- **Visible failures.** `staleLines` (`cmd/tl-lead/freshness.go:46-83`) reports a failed couch read or an unknown landing as a `check:` line. It does not quietly call the section fresh. `TestStatusCheckEdges` pins both lines.
- **Couch's error text is pinned.** The missing-section text is now one constant, `missingSection` (`freshness.go:27`). The fake returns couch's real refusal text (`sessionsync_fake_test.go:43`), and the opt-in live conformance test checks it against the real couch.
- **Graceful degradation in Slots.** A failed or missing actors read renders `?` and never blocks the section (`slots.go:26-38`, `readActors`).
- **No new durable state.** The stamp lives inside the section it describes. `check` writes nothing.

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
None new.
- BR-3 is only partly covered: an empty `deps.now` still falls back to the wall clock. That is the usual Go pattern and reads no outside state, so I consider it addressed. Details are in the disposition below.

**5. Test coverage notes**
- **BR-1:** I reverted the `Truncate(time.Minute)` fix and `TestStatusCheckEdges` failed with `a landing in the stamp's minute: ["stale: ## Slots (... pair landed 10-16 17:05)"]`. I then restored the file.
- **BR-2:** both conformance checks only run with `TL_LEAD_LIVE=1`, as round one asked. I did not run them live here; the issue log says both passed against the real couch.
- **Stamp tests:** they test pure functions directly (`TestSplitStamp`, `TestStampIsNotABullet`). The status-check tests inject the clock and the landing reader, so they need no real IO.

**6. Architectural notes for upcoming work**
- **ARCH-DRY: pass.** The existing Slots renderer was extended rather than duplicated, and there is one stamp helper and one write path.
- **ARCH-PURE: pass.** Stamp parsing and the budget rules are pure functions. git and couch sit behind injected seams.
- **ARCH-PURPOSE: pass.** All four spec items are delivered. The couch-side ping was deferred to pair with the TL's approval, and that ping isn't the main point of this issue.
- **ARCH-MOCK: pass.**
  - Couch actors has a seam, a fake and a live conformance check.
  - The session-sync fake now returns couch's real error text.
  - `gitNewestLanding` is tested against real temporary git repos, which is acceptable for local reads.
- **ARCH-CONSTRAINTS: pass.** `check` is limited to 4 couch reads plus one local git call per sibling repo, with no inventory read, and the help text says so.
- **ARCH-SECURE: pass.**
  - A stamp the operator hand-edits or deletes counts as unknown age, not as a parse failure.
  - git runs with an argument array, not a shell string.
  - Couch's JSON reply is decoded into a typed value and checked against `Code`.
- **ARCH-ORDER: pass.** The command keeps no state between runs except the stamp itself. The existing compare-and-replace retry still controls concurrent writes, and a fresh stamp is computed on each retry.
- **ARCH-FUNERAL: pass.** Nothing durable is added: each write replaces the section's single stamp line.

**7. Plan revision recommendations**
None. The 2026-10-16 Revisions entry already matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      freshness.go:70 now compares land.At.Truncate(time.Minute) with the stamp time. Reverting it turns TestStatusCheckEdges red (verified in this checkout).
  - id: BR-2
    disposition: addressed
    note: |
      missingSection is a named constant. TestSessionSyncConformance reads a missing section against it, and TestCouchActorsConformance pins the --actors shape. Both run only with TL_LEAD_LIVE=1, as round one asked.
  - id: BR-3
    disposition: addressed
    note: |
      main.go now wires landing to gitNewestLanding, and an empty landing seam returns no landing, so tests no longer run git. An empty clock still falls back to time.Now; that is the usual Go pattern and reads no outside state.
  - id: BR-4
    disposition: addressed
    note: |
      The plan's Revisions entry (2026-10-16 close review) records the 4 session-sync reads, the agent and agent-idle labels, none for an empty slot, and that only SDLC siblings count as landing sources. This matches freshness.go and slots.go.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6c28ce3b-abe5-42c6-a3e7-593a675eb99b","repository":"github.com/xianxu/ariadne","issue_id":"000414","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/cc9d3235aabb704b","reviewed_head":"2059abecb29b36c46b1e01b733adba7b26abfb87","slot":"ariadne:4","requested_at":"2026-10-17T00:43:41.072807Z"},"phase":"finished","pid":77017,"started_at":"2026-10-17T00:43:41.404855Z","finished_at":"2026-10-17T00:44:40.920161Z","session_id":"c9ea44d6-3002-4ccd-a435-c383448f6ddc","outcome":"returned"}] -->
