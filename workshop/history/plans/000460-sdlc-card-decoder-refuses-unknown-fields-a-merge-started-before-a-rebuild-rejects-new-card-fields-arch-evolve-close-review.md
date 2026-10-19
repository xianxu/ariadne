# Boundary Review — ariadne#460 (whole-issue close)

| field | value |
|-------|-------|
| issue | 460 — sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE) |
| repo | ariadne |
| issue file | workshop/issues/000460-sdlc-card-decoder-refuses-unknown-fields-a-merge-started-before-a-rebuild-rejects-new-card-fields-arch-evolve.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2abdfb582380a38e3cf4add6130128aa0d5e67aa..13be7fd15e87c19edb264fd7d9384b38e2b3e30e |
| command | sdlc close --issue 460 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T03:35:18-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Unknown-field preservation is well structured and the main compatibility regressions pass. Two reproduced gaps block shipping: known YAML fields still accept incorrect scalar types, and unknown JSON metadata is constrained by float64 decoding.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: strict-known-field-types
    title: |
      Tracker validation accepts non-string completion tokens
    detail: |
      cmd/sdlc/internal/issue/card.go:95 delegates nested validation to cardEnvelope, but handoff.go:97 uses yaml.Node.Decode, which coerces numeric and boolean scalars into strings. A temporary regression replacing completion.token with 123 or true showed ParseCard accepts both, violating the Spec's strict known-field types. Validate known YAML node kinds and tags before struct decoding across all nested tracker records, and add permanent rejection tests.
  - id: new
    severity: Important
    kind: defect
    family: opaque-unknown-field-values
    title: |
      Duplicate-key validation rejects valid unknown JSON numbers
    detail: |
      cmd/sdlc/teststats_store.go:314 constructs a decoder without UseNumber, so validateJSONValue converts unknown numeric tokens to float64. Adding "future":{"number":1e400} to an otherwise valid attempt makes decodeSuiteAttempt reject it with "cannot unmarshal number 1e400 into Go value of type float64". UseNumber preserves opaque numeric tokens during duplicate checking; add root and nested regressions while retaining strict known-number validation.
```

1. **Strengths**

   - The preservation helper derives tracker ownership from existing YAML tags, avoiding a second handwritten schema.
   - Tests cover unknown claimant, completion, tracker, and release-by descendants, plus intentional removal of known records.
   - Tracker CAS and landing-log retention have concrete compatibility regressions.
   - Atlas changes describe the new compatibility contract. No new command, flag, or configuration surface requires a README update.

2. **Critical findings**

   Known tracker scalar types remain insufficiently validated. See `strict-known-field-types` above.

3. **Important findings**

   Unknown JSON numbers must remain opaque during duplicate detection. See `opaque-unknown-field-values` above.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Full issue and landingline package tests passed. Selected tracker tests and `TestDecodeSuiteAttemptRejectsUntrustedRecords` passed; `git diff --check` passed.

   Both temporary adversarial overlay tests failed on the pinned head as described above. Add them permanently. The broader selected CLI run was stopped after several minutes without a result; no full-suite result is claimed. Tracked files remain unchanged.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass: shared preservation helper uses existing schema definitions. |
   | ARCH-PURE | Pass: parsing and merging remain pure; listed entities exist at their stated paths. |
   | ARCH-PURPOSE | Flag: strict-known/tolerant-unknown delivery is incomplete in the two reproduced cases. |
   | ARCH-MOCK | Pass: existing Git seam and isolated repository fixtures remain in use. |
   | ARCH-CONSTRAINTS | Pass: existing attempt size and scan budgets remain intact. |
   | ARCH-SECURE | Flag: YAML coercion accepts malformed persisted identity fields. |
   | ARCH-ORDER | Pass: existing CAS and publication sequencing remain unchanged. |
   | ARCH-FUNERAL | Pass: no new durable artifact family or increased persistence growth. |
   | ARCH-EVOLVE | Flag: preservation works, but strict known types and opaque unknown JSON values need correction. |

   The readiness audit’s referenced commit contains the decoder, preservation writer, and named regression tests; it remains outside this pinned implementation range.

7. **Plan revision recommendations**

   Add `## Revisions` entries recording exact YAML type validation across nested tracker records and token-preserving JSON duplicate validation, with permanent regressions for both. These complete the existing contract without expanding scope.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ba87cce4-a5a9-43eb-b772-f48b40b305f4","repository":"github.com/xianxu/ariadne","issue_id":"000460","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/dcf792c35ad57902","reviewed_head":"13be7fd15e87c19edb264fd7d9384b38e2b3e30e","slot":"ariadne:1","requested_at":"2026-10-19T10:28:51.707302Z"},"phase":"finished","pid":92026,"started_at":"2026-10-19T10:28:51.925481Z","finished_at":"2026-10-19T10:35:01.633718Z","session_id":"01a153b4-fc42-7f61-a1eb-c4566455f7ba","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T04:07:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 460 — sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE) |
| repo | ariadne |
| issue file | workshop/issues/000460-sdlc-card-decoder-refuses-unknown-fields-a-merge-started-before-a-rebuild-rejects-new-card-fields-arch-evolve.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2abdfb582380a38e3cf4add6130128aa0d5e67aa..f6124b5ef7985ae6e3f91eb378c396fee0652228 |
| command | sdlc close --issue 460 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T04:07:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range satisfies the additive compatibility contract. Unknown card members survive the affected rewrite paths, known tracker types remain strict, and test-stats accepts opaque future values without weakening known numeric validation. Both prior findings are addressed, with regression tests confirmed to fail when their respective fixes are disabled.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      handoff.go validates known YAML tags recursively before struct decoding. Disabling that boundary check makes TestParseCardRejectsMalformedKnownTrackerRecord fail for numeric and boolean completion tokens; both pass with the fix restored.
  - id: BR-2
    disposition: addressed
    note: |
      validateJSONNoDuplicates uses UseNumber. Removing it makes the permanent opaque-number regressions fail at root, unit, and known-flake depths; restored code passes those tests and rejects overflow in known seconds fields.
```

1. **Strengths**

   - Shared preservation logic in `cardunknown.go` retains nested unknown members while allowing deliberate removal of known records.
   - `cardEnvelope` validates types and versions at the shared boundary used by tracker accessors and setters.
   - Tests exercise actual serialized cards, Git-backed CAS writes, and landing-event retention.
   - Atlas updates accurately describe compatibility behavior. No new command, flag, or configuration key requires a README update.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed the complete `issue`, `tracker`, and `landingline` package tests, plus targeted suite-attempt decoder, storage, reader, and pruning tests. Both prior fixes passed mutation verification. `git diff --check` passed, and temporary mutations were restored with a clean checkout.

   Inspected the referenced #449 source and regression tests directly: readiness unknown members are captured as raw JSON and carried through vouch rewrites. This review did not rerun the repository-wide recorded suite.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — shared nested preservation and fact-object validation helpers. |
   | ARCH-PURE | Pass — parsing, validation, and merging remain directly testable without IO. |
   | ARCH-PURPOSE | Pass — scalar/title, claimant, envelope, and CAS paths checked; persisted-record audit recorded. |
   | ARCH-MOCK | Pass — no new external dependency; existing integration seams remain unchanged. |
   | ARCH-CONSTRAINTS | Pass — attempt size and scan limits remain enforced; no new concurrency. |
   | ARCH-SECURE | Pass — recursive YAML checks and JSON duplicate/type validation guard untrusted records. |
   | ARCH-ORDER | Pass — no new state between events; existing CAS ordering remains intact. |
   | ARCH-FUNERAL | Pass — no new durable artifact family or increased per-event storage. |
   | ARCH-EVOLVE | Pass — nested unknown members survive rewrites; known types and versions remain strict. |

7. **Plan revision recommendations:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"33f63a37-57e3-477b-8ae2-5dea4978e3aa","repository":"github.com/xianxu/ariadne","issue_id":"000460","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/434da79369551076","reviewed_head":"f6124b5ef7985ae6e3f91eb378c396fee0652228","slot":"ariadne:1","requested_at":"2026-10-19T11:05:12.637843Z"},"phase":"finished","pid":41277,"started_at":"2026-10-19T11:05:12.730906Z","finished_at":"2026-10-19T11:07:22.990857Z","session_id":"01a153d6-4210-7850-91a3-c228aa57a0cf","outcome":"returned"}] -->
