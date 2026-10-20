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
