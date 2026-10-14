# Boundary Review — ariadne#350 (milestone M1)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 1c39182532fa3cc78ea4935abff4ab1b87a6f01d..b696b7d263d37bc9e4287a0c73ebd13f920afdf5 |
| command | sdlc milestone-close --issue 350 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T08:47:13-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing #350 M1; ariadne#354 exemptions are on main. Running targeted tests next.

Checking README coverage and tl-lead tests, then writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 is delivered. `cmd/tl-worker` embeds its prose as help in the same shape as `doc-review`. The pointer stub, the AGENTS §15 load bullet, the Makefile `tools` entry and the new atlas page are all in place. The tests check every inventoried rule from a typed table (`pkg/tlrules` reading `rules.tsv`). The blocker logged for M1, the sandbox exemption from ariadne#354, is now on main: `.claude/settings.ariadne.json:106-107` lists both binaries, and `settingsx` passes. `go build ./...` is clean. Targeted tests pass in `cmd/tl-worker`, `cmd/tl-lead`, `pkg/markdoc`, `pkg/trains`, `pkg/tlrules`, `cmd/sdlc/internal/{issue,project}` and `settingsx`.

The window also carries M2 and M3 code. I reviewed it as well; nothing in it is Critical. One Important item blocks SHIP: README.md's list of the tools ariadne builds is now stale. Everything else is Minor, plus some plan bookkeeping.

**1. Strengths**
- **One shared parser:** `cmd/sdlc/internal/issue/markdoc.go` forwards every name to `pkg/markdoc` and holds no logic. The trains parser moved as a rename. The plan-reader source guard (`cmd/sdlc/commitpathspec_guard_test.go:145`, `:193`) now scans the whole module, so it covers the moved scanners and `tl-lead`.
- **Rule table:** `TestWorkerHelpCarriesEveryRule` checks one row per rule, against both `--help` and the bare command, so no row can pass by default. The `own`/`cited` scope column is a neat way to share the table with the lead's tests.
- **Write loop:** `writeSection` (`cmd/tl-lead/status.go:31`) is bounded, re-reads and recomputes on a stale hash, refuses a held lock immediately, and only folds owed IDs that are both owed and already seen.
- **Failures stay visible:** a malformed seen-set is an error, never treated as empty (`seen.go:27`). An unexpected `schema_version` is refused, and a failed car read renders `?`.
- **Bounded fan-out:** car reads run at most 8 at once (`trains.go:16`) with a 60 s timeout each, and the fleet read overlaps them.

**2. Critical:** none.

**3. Important**
- **README is stale (docs gate):** `README.md:195` still says "Ariadne builds `sdlc`, `datatype`, `vocabulary`, and `doc-review`". `make tools` now also builds `tl-worker` and `tl-lead`. Fix: add both to that sentence, with a one-line pointer to `--help`.

**4. Minor**
- **Plan names a restatement the test allows:** the plan says `TestWorkerHelpDoesNotRestateSection15` forbids "whole Bash call". The test omits that phrase, and `cmd/tl-worker/help.md:11` contains it. Either reword the citation line or record the change in the plan's Revisions.
- **Fold path never checked live (ARCH-MOCK):** `TestSessionSyncConformance` never folds (the Log gives the reason), so the fake's fold refusals are unchecked against the real couch. The argv form `--fold ID,ID` does match `couch --session-sync --help`. Revisit at the M3 close.
- **`ErrDeliveryGap` not built:** the plan names it, but the code has no such error and no `--accept-gap` handling. A delivery-gap refusal surfaces as a generic exit 1. It is not swallowed, but the plan drifts.
- **Off-train failure aborts the whole render (ARCH-SECURE):** if the fleet inventory fails, `gatherTrains` (`trains.go:68`) renders nothing, including the `## Trains` block, which doesn't depend on the inventory. It could render the off-train block as `?` instead.
- **Plan table names a missing file:** the Core concepts table lists `cmd/tl-lead/sdlc_fake_test.go`; the file is `cmd/tl-lead/fake_test.go`.

**5. Test coverage notes**
- `TestSection15NamesTheBinaries`, `TestSkillStubsPointAtBinaries` and the rule-table tests pin M1's contract.
- The issue's M1 line says the couch skill isn't contradicted. Only "no restatement" is tested; nothing checks for contradiction. That is acceptable for prose.
- `pkg/workspace TestEnvironmentLegacyAndArchive` fails in this checkout ("nil Git reader"). This diff doesn't touch `pkg/workspace`, so it is not attributable to this issue.

**6. Architecture**

| Principle | Result | Note |
|---|---|---|
| ARCH-DRY | pass | `tl-worker` and `tl-lead` `main.go` share a tiny cobra shape; not worth extracting |
| ARCH-PURE | pass | `planBar`, the renderers, `foldable`/`without` and `prependThread` are pure; the IO sits behind the `SessionSync`/`SDLC` seams |
| ARCH-PURPOSE | pass for M1 | the ops pointer and deleting the Python are explicitly the TL's ops change at M3 |
| ARCH-MOCK | minor | stateful fake plus opt-in live check; fold isn't covered live |
| ARCH-CONSTRAINTS | pass | 13.7 s against a ~10 s baseline, measured and accepted in the Log |
| ARCH-SECURE | pass | argv arrays only, version checks, typed decoding; one minor item on degrading |
| ARCH-ORDER | pass | stale/lock injectable at a chosen call; bounded retries |
| ARCH-FUNERAL | pass | seen-set pruned to the owed set; temp body file removed; `.clobbered` removed |

**7. Plan revisions to add**
- "TestWorkerHelpDoesNotRestateSection15 permits the one-line citation 'whole Bash call'; it forbids 'sandbox exemption', 'substitution', '$(' and the other phrases."
- "ErrDeliveryGap / `--accept-gap` were not built; a gap refusal surfaces as a generic error (exit 1)."
- "`sdlc_fake_test.go` is `fake_test.go`."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: readme-tool-surface-sync
    title: |
      README still lists only sdlc/datatype/vocabulary/doc-review as ariadne's make tools binaries
    detail: |
      README.md:195 is stale now that the Makefile tools target builds tl-worker and tl-lead. Add both binaries and a one-line pointer to their --help.
  - id: new
    severity: Minor
    kind: defect
    family: plan-test-contract-drift
    title: |
      TestWorkerHelpDoesNotRestateSection15 omits the planned "whole Bash call" refusal; help.md:11 contains it
    detail: |
      Either reword the citation in cmd/tl-worker/help.md or add a plan Revisions entry saying the one-line citation is allowed.
  - id: new
    severity: Minor
    kind: defect
    family: conformance-covers-modeled-behavior
    title: |
      Live session-sync conformance never exercises fold, so the fake's fold refusals are unchecked against couch
    detail: |
      cmd/tl-lead/conformance_test.go never folds. The argv form matches couch help, but the fold-not-owed refusal the fake models has no live check. Revisit at the M3 close.
  - id: new
    severity: Minor
    kind: defect
    family: plan-core-concepts-drift
    title: |
      Plan names ErrDeliveryGap and sdlc_fake_test.go; the code has neither (fake is fake_test.go, no gap error)
    detail: |
      Add Revisions entries so the Core concepts table and the SessionSync description match the code.
  - id: new
    severity: Minor
    kind: defect
    family: degrade-visibly-not-abort
    title: |
      A fleet-inventory failure aborts the whole trains render, including the independent Trains block
    detail: |
      cmd/tl-lead/trains.go:68 returns fleetErr before rendering the Trains block. It could render the off-train block as ? instead.
```

---

## Re-review — 2026-10-14T08:51:30-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 1c39182532fa3cc78ea4935abff4ab1b87a6f01d..392c1079c8f81086a51ede0b695c4cc6e8bebdcc |
| command | sdlc milestone-close --issue 350 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T08:51:30-07:00 |
| verdict | SHIP |

## Review

Verifying the prior fixes now: running the tests, then reverting each fix to confirm its test fails without it.

pkg/workspace's failure is pre-existing (it fails at base too, and the window doesn't touch it). Next, reverting each fix to confirm its regression test fails.

Both regression tests go red when their fix is reverted. Last check: whether couch is available to run the live conformance test.

```verdict
verdict: SHIP
confidence: high
```

All five prior findings are fixed, and each fix that changes behavior has a regression test I confirmed fails without it. I reverted `trains.go` and the `delivery-gap` branch of `decodeReply` in the disposable checkout. `TestGatherTrainsFleetFailureDegrades` and `TestDecodeReplyDeliveryGap` both failed. Reverting `cmd/tl-worker/help.md` made `TestWorkerHelpDoesNotRestateSection15` fail. The live check also passed against the real couch on this host: `TL_LEAD_LIVE=1 go test ./cmd/tl-lead -run TestSessionSyncConformance` passes with the new fold step. `go build ./...` is clean, and the `cmd/tl-lead`, `cmd/tl-worker`, `pkg/markdoc`, `pkg/trains`, `pkg/tlrules`, `cmd/sdlc/internal/{issue,project}` tests pass. One failure, `pkg/workspace` `TestEnvironmentLegacyAndArchive` ("nil Git reader"), is not from this branch: nothing in the window touches that package, and it fails the same way with the base version restored. One new Minor finding (a stale comment); nothing blocks.

1. **Strengths**
   - `cmd/tl-lead/trains.go:63-70`: when the fleet read fails, only the off-train block degrades and it says why; the Trains block still renders. The fleet read runs alongside the car reads, so it adds no time.
   - `cmd/tl-lead/sessionsync.go:94-96`: couch's `delivery-gap` reply becomes a typed `ErrDeliveryGap`, and the error text names the next step. `TestDeliveryGapRefusesFoldOnly` covers both sides: a fold is refused, a plain write still lands.
   - The BR-1 fix went past the one site named. README and all three atlas tool lists now name `tl-worker` and `tl-lead`; a grep for these tool lists in Markdown outside `workshop/` finds no stale copy.
   - The live conformance test checks that a refused fold leaves the section unchanged, not just that it errors (`conformance_test.go:66-68`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `cmd/tl-lead/conformance_test.go:16-17`: the doc comment still ends "and the test never folds", but the test now folds (lines 53-68). Change it to say the fold names an ID that isn't owed, so no real event is touched.

5. **Test coverage notes**
   - `TestDeliveryGapRefusesFoldOnly` uses a fake that returns `ErrDeliveryGap` directly. The real couch-to-error mapping is covered separately by `TestDecodeReplyDeliveryGap`, and together they cover the path.
   - The degraded Slots text puts `fleetErr.Error()` inside a code fence unescaped. Stderr containing a code fence would break it. This is speculative, so I'm not raising it.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The scanners and trains parser moved into `pkg/markdoc` and `pkg/trains` instead of being copied.
   - **ARCH-PURE:** pass. The rendering functions (`renderTrains`, `offTrainSlots`, `foldable`) are pure, and IO goes through the `SDLC` and `SessionSync` interfaces.
   - **ARCH-PURPOSE:** pass. The BR-1 class (every list of the tools ariadne builds) was swept, not just the README instance.
   - **ARCH-MOCK:** pass. `fakeSessionSync` keeps state across calls, and the live conformance check now covers the fold refusals it models.
   - **ARCH-CONSTRAINTS:** pass. The fleet read is bounded and overlapped; a timeout now degrades the output instead of aborting.
   - **ARCH-SECURE:** pass. couch is called with an argv array, and an unreadable reply is a visible error.
   - **ARCH-ORDER:** pass. The write is guarded by the read hash (`ErrStale`), and a delivery gap is treated as "unknown", not "nothing owed".
   - **ARCH-FUNERAL:** pass. The seen-set is pruned of folded IDs (`without(seen, fold)`), and the conformance test works in a temporary directory.
   - **Upcoming:** a later round could have the live check also cover the `delivery-gap` reply format.

7. **Plan revision recommendations:** none needed. The 2026-10-14 Revisions entry already overrides the Core concepts row `cmd/tl-lead/sdlc_fake_test.go` (plan line 48). Optionally, change that row to point straight at `cmd/tl-lead/trains_test.go`.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:195 and atlas base-layer.md:75, setup-and-replication.md:54, sandbox.md:13 all list tl-worker and tl-lead; a grep sweep finds no stale list.
  - id: BR-2
    disposition: addressed
    note: |
      help.md:11 reworded; the test forbids "whole Bash call"; restoring the old help.md fails TestWorkerHelpDoesNotRestateSection15.
  - id: BR-3
    disposition: addressed
    note: |
      conformance_test.go:53-68 folds an ID not owed against real couch; it passed live on this host with TL_LEAD_LIVE=1.
  - id: BR-4
    disposition: addressed
    note: |
      ErrDeliveryGap exists and is mapped in decodeReply; reverting the mapping fails TestDecodeReplyDeliveryGap; a plan Revisions entry fixes the fake file names.
  - id: BR-5
    disposition: addressed
    note: |
      trains.go:63-70 renders Trains and marks Slots unread; reverting fails TestGatherTrainsFleetFailureDegrades.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: comment-tracks-code
    title: |
      TestSessionSyncConformance doc comment still says the test never folds
    detail: |
      cmd/tl-lead/conformance_test.go:16-17 says "the test never folds", but lines 53-68 now fold a not-owed ID; reword it to say the fold names an ID that isn't owed.
```
