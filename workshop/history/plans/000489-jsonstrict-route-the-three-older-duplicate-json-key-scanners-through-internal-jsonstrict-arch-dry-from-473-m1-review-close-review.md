# Boundary Review — ariadne#489 (whole-issue close)

| field | value |
|-------|-------|
| issue | 489 — jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review) |
| repo | ariadne |
| issue file | workshop/issues/000489-jsonstrict-route-the-three-older-duplicate-json-key-scanners-through-internal-jsonstrict-arch-dry-from-473-m1-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | bf7eaa4d17eaf48b7b36b71982f65bdcf5f9d114..497ca73e8dd4c7923e0c4a071656065480f858a6 |
| command | sdlc close --issue 489 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T05:39:49-07:00 |
| verdict | SHIP |

## Review

Still reviewing: diff read, now running the affected packages' tests and the lint catalog.

Need to keep the import used; append a blank-identifier reference.

```verdict
verdict: SHIP
confidence: high
```

The change does what #489 set out to do. All six recursive duplicate-key JSON scanners are deleted and now route through `jsonstrict.Scan` or `jsonstrict.RejectDuplicateKeys`: observe, fleet, teststats_store, receipt, reviewusage and actualcorrection. The two top-level allowlist decoders (issueassets policy and estimate correction) also dropped their own `seen` maps. Every caller keeps its error prefix, and a new golint rule stops the scanner being copied again. Behavior was checked in three ways:
- **Depth bounds:** the receipt limit (12) and the reviewusage limit (16) still count depth the same way as before (`depth > max`, checked on entry, top-level value at 0).
- **Number handling:** `UseNumber` moved into `RejectDuplicateKeys`, so teststats still accepts `1e400`.
- **Accepted inputs:** the only change is case-fold strictness, which can only reject more input, never less. The Log says this openly.

All affected packages pass, including `./cmd/sdlc` and `pkg/golint` (which runs `TestTrackedGoFilesPassTheCatalog`). I also broke each caller's duplicate check on purpose to see whether a test catches it. Every caller's test went red except one, and that gap is older than this change (Minor below). Nothing blocks SHIP.

1. **Strengths**
   - `jsonstrict.Scan(dec, maxDepth)` (`cmd/sdlc/internal/jsonstrict/jsonstrict.go:50-56`) is a small API that will hold up. The caller owns the decoder settings and whatever follows the value, which lets the stream callers (receipt, reviewusage, actualcorrection) keep their own trailing-data checks.
   - The work covers the whole class, not just the three named sites (ARCH-PURPOSE). The sweep found three more recursive scanners and two allowlist `seen` maps, and the Revisions section records the wider scope. I checked the decoders left in place: `tracker/reader.go:222`, `ghclient.go:235`, landingline, qodertranscript and reviewprepare. None is a recursive key-set scanner.
   - The lint rule `json-recursive-scanner` (`pkg/golint/rules.go:450-504`) enforces the single owner, and its tests include negative cases: a recursive skipper with no key set, a one-level member walk, and a walk with no tokens.
   - The atlas lint catalog is updated (`atlas/workflow/sdlc-binary.md:2123`).
   - No removed error string is referenced anywhere else in `cmd`, `pkg`, `atlas` or `README.md`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/internal/reviewusage/identity.go:59`: no test covers this caller's duplicate-key check. Replacing the `Scan` call with `nil` leaves the reviewusage suite green. The gap predates this change, but the line was touched here.
   - The lint rule's limits: it skips methods (`x.Recv == nil`), only counts recursion through a bare identifier, needs the `map[string]` literal inside the function body, and can't see iterative stack walks like the old observe scanner. The Log already notes the last of these.
   - The issue's Problem section is empty.

5. **Test coverage notes**
   - Results of breaking each caller's check:

     | Caller | Test that went red |
     |---|---|
     | reviewusage `Parse` | `TestStoreStrictDocument` |
     | estimate | `TestParseCorrectionRejectsMalformed` |
     | issueassets | `TestPolicyStrictAndFailClosed` |
     | receipt | `TestReceiptRejectsUntrustedShapeIdentityAndImpossibleState` (forged `"version"` twice) |
     | observe | `TestObservationJSONIsStrict` |
     | fleet | `TestJSONContract_RejectsDuplicateKeysAtEveryDepth` |
     | reviewusage identity | none |

   - `TestScan` pins the depth boundary at 12, the one-value-and-leave-the-stream contract, and the `1e400` case.

6. **Architecture notes**
   - **Pass:** ARCH-DRY, ARCH-PURPOSE, ARCH-PURE (pure functions, tests run with no IO), ARCH-SECURE (the depth bounds on untrusted receipts and usage records are kept, and failures still return errors).
   - **N/A:** ARCH-MOCK (no external dependency), ARCH-ORDER (no state is held between events), ARCH-FUNERAL (nothing durable is created), ARCH-CONSTRAINTS (same complexity as before; the bounds are kept).
   - **ARCH-EVOLVE:** pass, with one note. Teststats checks unknown future metadata with `RejectDuplicateKeys`, so a future map-shaped value whose keys differ only by case (`"A"` and `"a"`) will now be refused. Those are distinct keys for a `map[string]X`. This is deliberate and only ever stricter. If such metadata ever becomes real, `jsonstrict` should offer an exact-match mode for opaque values.

7. **Plan revisions:** none. The Revisions entry already matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: strict-check-untested-caller
    title: |
      reviewusage identity.go duplicate-key Scan has no test that fails without it
    detail: |
      Replacing jsonstrict.Scan(..., 16) at cmd/sdlc/internal/reviewusage/identity.go:59 with nil leaves the reviewusage suite green. The gap predates this change, but a session-event line with a duplicate key should be rejected by a test.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-shape-coverage
    title: |
      json-recursive-scanner misses method receivers, selector recursion and iterative stack walks
    detail: |
      recursiveJSONScanners skips FuncDecls with a receiver, only treats a bare identifier call as recursion, and cannot see iterative walks (the Log notes the last). A future method-based scanner would get past the single-owner guard.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"107f4f76-91ba-4082-b761-bd1381005a95","repository":"github.com/xianxu/ariadne","issue_id":"000489","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/8b9699b35ff09a6e","reviewed_head":"497ca73e8dd4c7923e0c4a071656065480f858a6","slot":"ariadne:2","requested_at":"2026-10-20T12:33:25.42607Z"},"phase":"finished","pid":96433,"started_at":"2026-10-20T12:33:25.582099Z","finished_at":"2026-10-20T12:39:48.701623Z","session_id":"0cba0e04-56ac-47e9-8e74-aaa6156b5f8f","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T05:42:06-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 489 — jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review) |
| repo | ariadne |
| issue file | workshop/issues/000489-jsonstrict-route-the-three-older-duplicate-json-key-scanners-through-internal-jsonstrict-arch-dry-from-473-m1-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | eac2a35b5770f79a1671912390c3923eb12ceb2f..07e0d9724226706fbe4751c3de68aef6eedbe852 |
| command | sdlc close --issue 489 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T05:42:06-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This fix window fully fixes BR-1 and mostly fixes BR-2. It doesn't change `## Spec`, `## Revisions` or `## Done when`, and it adds nothing no finding asked for. The issue file only changes status, flow and the Log, and the two new `workshop/plans/` files are generated gate and review records. For BR-1, I disabled the `jsonstrict.Scan` call in `identity.go`, and the new case in `TestIdentityStrictProtocol` went red. BR-2 has two parts. Selector recursion is fixed and tested: deleting `rules.go:477` turns `TestRules` red. Methods are now scanned too, because the `x.Recv == nil` filter is gone (`rules.go:489`). But putting that filter back leaves `pkg/golint` green. The test harness wraps every snippet in a receiver-less `func f()` (`rules_test.go:22`), so no test runs the rule on a method. The Log says methods are covered "with a table row and a mutation check", which overstates this. Only a missing test remains, and it is Minor, so nothing blocks the boundary. Both affected packages pass at the pinned head, including the tree-wide lint catalog, so widening the rule flags nothing new in the tree.

1. **Strengths**
   - `identity_test.go:37-42` checks both an exact duplicate and a different-case duplicate (`SESSION_ID`). That covers the case-folding strictness the earlier review noted as the only behavior change.
   - The comment at `rules.go:455-456` now says plainly that iterative stack walks are out of scope for a syntactic rule.
   - The selector-recursion row (`rules_test.go:64`) is a positive case, and removing the fix turns it red.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-2 remains (not-addressed, below): no test covers the receiver half of the fix.

5. **Test coverage notes**
   - These are the results of undoing each fix and re-running the tests:

     | Fix undone | Result |
     |---|---|
     | `identity.go` Scan disabled | `TestIdentityStrictProtocol` FAIL |
     | selector-recursion line removed (`rules.go:477`) | `TestRules` FAIL |
     | `x.Recv == nil` put back (`rules.go:489`) | **pass** (no test covers it) |

   - This is the 2nd finding in family `lint-rule-shape-coverage`. The rule that covers all of them: every syntactic shape the lint rule claims to recognise needs a positive table row in that same shape. Rows are currently all bodies of a plain function, so a method needs a row built from a whole source file. The `TestRulesFollowImportNames` style at `rules_test.go:209` already does this: feed `package p … func (s *S) walk() { … s.walk() … }` to `Check` and assert it is flagged.

6. **Architectural notes**
   - **ARCH-DRY:** pass. There's no new duplication. The test reuses the existing `init` fixture.
   - **ARCH-PURE:** pass. The lint rule is a pure AST walk and its tests do no IO.
   - **ARCH-PURPOSE:** flag (Minor, under BR-2). The fix covered the class: methods, selectors, and a documented limit for iterative walks. Only the regression test for the method part is missing.
   - **ARCH-EVOLVE:** N/A. No persisted format changes in this window.

7. **Plan revision recommendations**
   - Once the method test exists, the Log line "with a table row and a mutation check" will be accurate. Until then it should say that only selector recursion has a mutation check.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Disabling jsonstrict.Scan in identity.go turns TestIdentityStrictProtocol red (identity_test.go:40); both an exact and a different-case duplicate are covered.
  - id: BR-2
    disposition: not-addressed
    note: |
      Selector recursion is fixed and tested, since removing rules.go:477 turns TestRules red. Methods are scanned now that the Recv filter is gone (rules.go:489), but putting that filter back keeps pkg/golint green. The harness wraps snippets in a receiver-less func f(), so a method needs a whole-file test case (as in TestRulesFollowImportNames).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"662b2dd4-e2dc-4d3f-9eb3-54cc7fd99bab","repository":"github.com/xianxu/ariadne","issue_id":"000489","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/99a02a4ee0898403","reviewed_head":"07e0d9724226706fbe4751c3de68aef6eedbe852","slot":"ariadne:2","requested_at":"2026-10-20T12:41:18.055295Z"},"phase":"finished","pid":98317,"started_at":"2026-10-20T12:41:18.155523Z","finished_at":"2026-10-20T12:42:05.805988Z","session_id":"86212ef0-bdb5-45dd-b6c1-84d37d023039","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T05:43:41-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 489 — jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review) |
| repo | ariadne |
| issue file | workshop/issues/000489-jsonstrict-route-the-three-older-duplicate-json-key-scanners-through-internal-jsonstrict-arch-dry-from-473-m1-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | 891e6026035610eb538e0990211ba0ec30206261..25a66f067ff83d098f417b0aa07a81b72abe0eaf |
| command | sdlc close --issue 489 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T05:43:41-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This fix window closes BR-2, the last open finding. The only code change is a new test, `TestJSONRecursiveScannerSeesMethods` (`pkg/golint/rules_test.go:230-242`). It builds a whole source file containing a method that calls itself through its receiver. It then checks that the method is flagged in package `p` and not flagged in `jsonstrict`, the one package allowed to own a recursive scanner. I checked that the test actually catches the bug, in a disposable checkout:
- **Receiver filter put back** (`x.Recv == nil &&` re-added at `rules.go:489`): the new test fails with "package p: findings [] <nil>, want 1". The previous round found that `pkg/golint` stayed green under this same change.
- **Receiver-call check removed** (`fn.Sel.Name == self` at `rules.go:477`): the new test also fails.
- **Clean head:** `go test ./pkg/golint/` passes and `go vet` is clean.

The rule still can't see iterative stack walks, and its comment at `rules.go:455-456` says so. The previous round accepted that documented limit. Apart from the generated gate and review records in `workshop/plans/`, the only other change is one Log entry in the issue file. It now credits the selector case to the table row and the method case to the whole-file test, which is accurate. `## Spec`, `## Revisions` and `## Done when` are unchanged, and nothing was added that no finding asked for. There are no conflict resolutions to judge.

1. **Strengths**
   - The test follows the existing whole-file pattern in `TestRulesFollowImportNames`. It is the fix the previous round asked for.
   - In one table, it checks both that a method outside `jsonstrict` is flagged and that the same method inside `jsonstrict` is allowed.
   - The corrected Log line (issue file, Log under 2026-10-20) no longer overstates what is tested.
2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none.
5. **Test coverage notes:** both mutations above now fail. Together with the previous round's selector mutation, every shape the rule claims to recognise has a test that fails without it. That is the rule the `lint-rule-shape-coverage` family asked for.
6. **Architecture**
   - **ARCH-DRY:** pass. The test reuses `Check` and `ruleOf`.
   - **ARCH-PURE:** pass. It runs on in-memory source and does no IO.
   - **ARCH-PURPOSE:** pass. The class of shapes the finding named is now covered: methods, selector calls, and the documented limit for iterative walks.
   - **ARCH-EVOLVE:** N/A. No persisted format changes.
7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      TestJSONRecursiveScannerSeesMethods (rules_test.go:230) goes red when x.Recv == nil is restored at rules.go:489 and when the receiver-selector recursion check at rules.go:477 is removed; the iterative-walk limit is documented at rules.go:455-456.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8ec03518-ff21-46d2-b20b-37dbff662669","repository":"github.com/xianxu/ariadne","issue_id":"000489","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/209a37b03c2b1dc3","reviewed_head":"25a66f067ff83d098f417b0aa07a81b72abe0eaf","slot":"ariadne:2","requested_at":"2026-10-20T12:43:03.722098Z"},"phase":"finished","pid":24943,"started_at":"2026-10-20T12:43:03.81613Z","finished_at":"2026-10-20T12:43:41.042938Z","session_id":"7ab08645-b7d4-41e5-8df3-946a5c19008a","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T08:03:18-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 489 — jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review) |
| repo | ariadne |
| issue file | workshop/issues/000489-jsonstrict-route-the-three-older-duplicate-json-key-scanners-through-internal-jsonstrict-arch-dry-from-473-m1-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | 86b79c6d826728c4f931e373ba4e8198629bb5bd..16e4c38844f7ae13cabde1b5ed91a42ff04e4478 |
| command | sdlc close --issue 489 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T08:03:18-07:00 |
| verdict | SHIP |

## Review

Queuegate and jsonstrict tests pass. Next I'm checking which callers depend on the old error strings and what the stricter case-fold rule changes.

The one failing test turned out to be an environment issue, not the code. Writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

This window has one code change. It is the conflict resolution in `cmd/sdlc/internal/queuegate/codec.go`. While this branch was open, #481 landed `queuegate` on main with its own recursive duplicate-key scanner. This branch's `json-recursive-scanner` lint rule caught it, and the integrated suite failed on `TestTrackedGoFilesPassTheCatalog`. The fix replaces queuegate's 54-line local walker with `jsonstrict.Scan(d, 128)` and keeps queuegate's own pieces: `UseNumber`, the 1 MiB and UTF-8 checks, and the check for trailing data. The depth rule matches the old one exactly: the top-level value counts as depth 0, and anything deeper than 128 is refused. The other two files in the window are the gate and review records in `workshop/plans`. The issue diff only adds to `## Log` and a note about the refused first landing; it does not touch `## Spec`, `## Done when`, or the substance of `## Revisions`. So this is the issue's own scanner class being folded in, not a change of scope. Both prior findings were disposed in earlier rounds, and nothing open is left.

**Strengths**
- `queuegate/codec.go:25`: the fold keeps both sides' intent. The 128 depth limit, `UseNumber` opacity (so `1e99999` stays opaque) and the trailing-data rejection all still hold, and the queuegate tests confirm it: `TestStrictObjectSharedBoundary…`, `TestControlCodecPreservesOpaqueAdditions` and the 130-level nesting case.
- The new lint guard caught a scanner that landed on main in the meantime. That is direct evidence that the class enforcement works (ARCH-PURPOSE).

**Critical:** none.

**Important:** none.

**Minor**
- Queuegate's duplicate check is now stricter. It refuses keys that match case-insensitively (e.g. `{"Id":1,"id":2}`) at every depth, including opaque extra values that the codec otherwise passes through untouched. Before, it refused only exact repeats. Nothing in queuegate tests this case, and the comment on `inspectJSON` doesn't mention it.

**Test coverage**
- These pass at the pinned head: `go test ./cmd/sdlc/internal/queuegate/ ./cmd/sdlc/internal/jsonstrict/` with `-count=1`, and the `pkg/golint` tests `TestTrackedGoFilesPassTheCatalog`, `TestRules` and `TestJSONRecursive*`.
- `TestQueueRealHelperPause` in `cmd/sdlc` fails here, and the cause is the review sandbox, not this change. The Python helper's `queue_boot_id` raises `PermissionError: [Errno 1] boot identity unavailable`, and that test never reaches the Go codec. It should be confirmed on the host suite.

**Architecture (interdiff only)**
- **ARCH-DRY: pass.** A duplicate scanner is removed in favour of the shared helper.
- **ARCH-PURE: pass.** It is still a pure function over bytes.
- **ARCH-PURPOSE: pass.** The new instance of the class was folded in during the same round.
- **ARCH-EVOLVE: pass, with the note above.** Unknown fields still survive the round trip and duplicate keys stay strict. The stricter case-insensitive rule now also applies to opaque nested values, which is a deliberate jsonstrict decision but undocumented in queuegate.

**Plan revisions:** none needed. The log already records the queuegate fold.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: strict-check-untested-caller
    title: |
      queuegate now refuses case-insensitive duplicate keys at every depth, but no queuegate test covers it
    detail: |
      This is the 2nd finding in family strict-check-untested-caller. Rule covering all instances: every caller of jsonstrict.Scan or RejectDuplicateKeys needs one test, through that caller's public decode, that feeds a case-variant duplicate key, e.g. DecodeControl with future set to {"Id":1,"id":2}. The interdiff silently made queuegate stricter: before, only exact repeats were refused, now case-insensitive repeats are too, including inside opaque extra values. The inspectJSON comment should say so as well. Prevalence: 2 callers found so far (reviewusage identity.go, queuegate codec.go).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a01e1cb3-8b93-4e4c-9bbd-ed08b9403002","repository":"github.com/xianxu/ariadne","issue_id":"000489","recipe":"fix-review","round":4,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/8b435d3511113395","reviewed_head":"16e4c38844f7ae13cabde1b5ed91a42ff04e4478","slot":"ariadne:2","requested_at":"2026-10-20T14:56:23.263933Z"},"phase":"finished","pid":58103,"started_at":"2026-10-20T14:56:23.44692Z","finished_at":"2026-10-20T14:57:12.7922Z","session_id":"47bf4d65-772c-4d25-ac7e-a97125da47d9","outcome":"returned"},{"version":1,"intent":{"disposable_root":true,"id":"68444b10-d7b7-4366-9d68-82855b702f27","repository":"github.com/xianxu/ariadne","issue_id":"000489","recipe":"fix-review","round":4,"attempt":2,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/8b435d3511113395","reviewed_head":"16e4c38844f7ae13cabde1b5ed91a42ff04e4478","slot":"ariadne:2","requested_at":"2026-10-20T14:57:14.362629Z"},"phase":"finished","pid":1573,"started_at":"2026-10-20T14:57:14.538516Z","finished_at":"2026-10-20T15:03:17.619016Z","session_id":"b53b494d-ca61-4a26-8d7f-88521b524dda","outcome":"returned"}] -->

## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing valid findings block"
output: "Checking queuegate tests and running a mutation to see whether they catch removal of the scan.\n\n```verdict\nverdict: SHIP\nconfidence: high\n```\n\n**Verdict: SHIP.** No prior findings are left open: BR-1 and BR-2 were closed in earlier rounds, and I did not raise them again. The only code change in this round is in `cmd/sdlc/internal/queuegate/codec.go`. Before this round, `queuegate.inspectJSON` had its own recursive duplicate-key scanner, which #481 landed on main while this issue was in progress. It now calls the shared `jsonstrict.Scan(d, 128)`. That is the purpose of this issue (fold every generic recursive scanner into `jsonstrict`), applied to one scanner the earlier review never saw. The rest of the window is generated gate and review records under `workshop/plans/`. `## Spec`, `## Revisions` and `## Done when` are unchanged, so this is not a scope change. The prompt lists no conflict resolutions to judge.\n\n1. **Strengths**\n   - **Same behavior as before, with one intended difference.**\n     - Duplicate keys are still rejected at every depth.\n     - Nesting is still capped at 128: both the old check and `Scan` reject only depth greater than 128.\n     - `UseNumber` and the trailing-data check stay in the caller (`codec.go:24`, `codec.go:28`).\n     - The old code also checked closing brackets by hand. `Scan` instead passes on whatever error `dec.Token()` returns, and the decoder already enforces the structure.\n     - The one difference: keys that differ only by case now count as duplicates. The issue's Log records this for every folded caller.\n   - **The existing tests catch removal of the scan.** I replaced `jsonstrict.Scan` with a plain `d.Decode`, ran the tests, then restored the file (`git status` clean).\n     - `go test ./cmd/sdlc/internal/queuegate/` failed in three places: the ambiguous object at `codec_test.go:17`, the duplicate `schema_version` at `codec_test.go:48`, and the nesting limit at `codec_test.go:108`.\n     - So the `strict-check-untested-caller` family does not repeat here.\n   - **The lint rule from this issue did its job.** The json-recursive-scanner rule flagged the scanner #481 brought in. `pkg/golint`, `jsonstrict` and `queuegate` all pass at head.\n\n2. **Critical:** none.\n3. **Important:** none.\n4. **Minor:** none.\n5. **Test coverage notes:** `queuegate` has no local test for the new case-variant duplicate rule, for example `\"future\":{\"a\":1,\"A\":2}`. That rule is tested in `jsonstrict` itself. The removal mutation above shows `queuegate` relies on `Scan`, so it gets the same rule. I'm not raising a finding for this.\n6. **Architecture (interdiff only)**\n   - **ARCH-DRY:** pass. The duplicated 50-line scanner is gone and replaced by the shared helper.\n   - **ARCH-PURE:** pass. `inspectJSON` is still pure work on in-memory bytes.\n   - **ARCH-PURPOSE:** pass. The scanner that arrived on main during the issue is folded in this same window, so it isn't left as a follow-up.\n   - **ARCH-EVOLVE:** pass.\n     - Unknown members are still preserved; the raw-map path in `StrictObject` is unchanged.\n     - Duplicate-key handling is stricter, not looser: keys differing only by case are now rejected, which matches how `encoding/json` matches field names.\n     - Known-field types and required fields are unchanged.\n7. **Plan revision recommendations:** none. The issue's Log already records the `queuegate` fold.\n\n```findings\n```"

truncated: false
retained-bytes: 0000003657
