# Boundary Review — ariadne#440 (whole-issue close)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..3b3545a9cf592714228e6e89cf16395a756cbd04 |
| command | sdlc close --issue 440 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T06:24:38-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared syntax index, annotation convention, and catalog integration are implemented and the existing tests pass. However, additional source probes confirmed false positives for rebound targets and overwritten decoders, violating the Spec’s conservative-provenance guarantee. The new contributor-facing annotation also lacks a README update.

Done-when: current — it covers the annotations, direct-chain rule, cross-file resolution, conservative matching, and counted index workload.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: wire-target-provenance
    title: |
      Target tracking retains wire identity after reassignment and guesses through interfaces
    detail: |
      pkg/golint/wire.go:152-169 sets variables[name] from an initializer but never invalidates target bindings; branch invalidation at line 188 only handles decoders. A source probe with `var m any = &Peer{}; m = &Local{}; d.Decode(m)` incorrectly reports wire-strict-decode, as does conditional reassignment. Even `var m any = Peer{}; d.Decode(&m)` reports despite encoding/json decoding that interface into a map without rejecting unknown fields (confirmed at runtime). ARCH-PURPOSE: honor the conservative destination contract by distinguishing declared types from initializer guesses and invalidating target provenance across direct, multiple, and conditional assignments. Add regression fixtures for this enumeration.
  - id: new
    severity: Critical
    kind: defect
    family: wire-decoder-strictness-provenance
    title: |
      Overwriting a decoder leaves stale strictness in the matcher
    detail: |
      pkg/golint/wire.go:163-165 skips non-identifier assignment targets, and invalidateAssigned at lines 205-208 has the same limitation. Consequently `d := json.NewDecoder(r); d.DisallowUnknownFields(); *d = *json.NewDecoder(r); d.Decode(&Peer{})` incorrectly reports a strict decode. A source probe reproduced the finding; a runtime control confirmed the overwritten decoder accepts unknown fields. ARCH-PURPOSE: invalidate strictness when assignments overwrite a tracked decoder, including conditional writes, and add negative regression fixtures.
  - id: new
    severity: Important
    kind: defect
    family: readme-user-surface
    title: |
      README update appears missing for the wire annotation convention
    detail: |
      atlas/workflow/sdlc-binary.md:1918 introduces the contributor-facing `// wire: <reason>` syntax and line 1929 documents its rule-specific suppression, but README.md is unchanged in the pinned range. The review’s README gate covers new surface a reader types. Add a concise annotation example and link to the atlas coverage and exception guidance.
```

1. **Strengths**

   - `CheckSources` parses and indexes the source set once, while `checkParsed` preserves shared suppression handling ([check.go:67](pkg/golint/check.go:67)).
   - Tests cover cross-file/import resolution, import aliases, generated-file exclusion, suppression behavior, and several ambiguous flows.
   - All four required annotations describe independently evolving producers; their actual consumers use tolerant `json.Unmarshal`.
   - Atlas guidance clearly distinguishes wire messages from intentionally strict persisted contracts.

2. **Critical findings**

   - **Target provenance:** [wire.go:152](pkg/golint/wire.go:152). Track proven destination identity and clear it when assignments make it uncertain.
   - **Decoder strictness:** [wire.go:163](pkg/golint/wire.go:163). Account for writes through decoder pointers before reporting subsequent calls.

3. **Important findings**

   - **README coverage:** document the annotation syntax and link to [the atlas convention](atlas/workflow/sdlc-binary.md:1918).

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./pkg/golint -count=1` and range `git diff --check` passed. The repository scan checked 1,174 Go files against six rules in 271 ms. Four additional source probes reproduced the false positives described above; runtime controls confirmed that the overwritten decoder and nonpointer interface destination accept unknown fields. The full suite was not rerun. No tracked files were changed.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one shared index and suppression implementation.
   - **ARCH-PURE — pass:** index and matcher operate on ASTs; source collection remains the IO boundary.
   - **ARCH-PURPOSE — flag:** the two provenance defects contradict the explicitly conservative rule contract.
   - **ARCH-MOCK — pass:** no new external dependency; pure matcher tests supply source maps directly.
   - **ARCH-CONSTRAINTS — pass:** counted index tests and the measured repository scan support the declared workload.
   - **ARCH-SECURE — pass:** parse errors propagate; no new secret handling.
   - **ARCH-ORDER — pass:** no state persists between external events or checker invocations.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts or background work.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying the conservative-provenance regression matrix: interface destinations, direct/multiple/conditional target reassignment, and direct/conditional decoder overwrites. Preserve the existing contract and record how the expanded fixtures enforce it.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"abba2962-8c81-4e86-9de0-13a41af27890","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a2c8b5e09ac57163","reviewed_head":"3b3545a9cf592714228e6e89cf16395a756cbd04","slot":"ariadne:1","requested_at":"2026-10-18T13:21:43.829254Z"},"phase":"finished","pid":71529,"started_at":"2026-10-18T13:21:43.927145Z","finished_at":"2026-10-18T13:24:37.910164Z","session_id":"01a14f2c-e2e4-7fd3-a08e-129b561c4583","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T06:51:14-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..42793ebdf419c790073bcc343420225349588850 |
| command | sdlc close --issue 440 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T06:51:14-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared syntax index, annotation convention, and documentation are delivered. BR-1 and BR-3 are addressed with supporting evidence. BR-2 remains blocking: the correction handles direct pointer writes but misses other ways to overwrite a tracked decoder.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The explicit-type guard and assignment invalidation cover the reported interface and target-reassignment cases. All five corresponding regression cases pass at head and fail with the pre-fix matcher.
  - id: BR-2
    disposition: not-addressed
    note: |
      pkg/golint/wire.go:173-174 and 208-219 invalidate only the written identifier, and the branch walker recognizes only ast.AssignStmt. Three alias-write probes and one range-reassignment probe still produce false wire-strict-decode findings. For example, after d.DisallowUnknownFields(), alias := d; *alias = *json.NewDecoder(r) leaves d marked strict. Likewise, for _, d = range []*json.Decoder{json.NewDecoder(r)} {} does not invalidate d. Runtime controls confirm the overwritten decoders accept unknown fields. ARCH-PURPOSE: the correction must cover the overwrite class. This is the 2nd finding in family `wire-decoder-strictness-provenance`. Do NOT fix only these instances: state and enforce a rule that invalidates strictness whenever an unmodeled alias can overwrite the decoder, and enumerate all assignment forms, including range bindings. Add regression fixtures for that enumeration.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:11-17 now documents the wire annotation syntax, gives a peer-message example, distinguishes version-pinned stores, and links to atlas/workflow/sdlc-binary.md, whose convention section documents coverage and reasoned exceptions.
```

Done-when: current — it still covers the annotation, four existing types, conservative decoder matching, documentation, and counted index workload.

1. **Strengths**

   - One shared index supports same-package and imported types without filesystem reads in the matcher.
   - Existing suppression handling is reused.
   - All four required input types are annotated, and both README and atlas explain the convention.
   - The seven corrective fixtures detect the original bugs: substituting the pre-fix matcher makes all seven fail.

2. **Critical findings**

   - **BR-2 remains open**, at `pkg/golint/wire.go:173-174` and `208-219`. Track alias identity or conservatively discard proof when aliases escape the supported flow; invalidate writes from every applicable assignment form.

3. **Important findings**

   None newly raised.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./pkg/golint -count=1 -v` passed, including the scan of 1,174 tracked Go files. Range diff whitespace checks passed. Temporary probes reproduced four remaining false positives; temporary artifacts were removed and the checkout remains unchanged. The full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared index and suppression implementation.
   - **ARCH-PURE: pass** — core entities operate on supplied bytes and ASTs; filesystem collection stays in the integration test.
   - **ARCH-PURPOSE: flag** — BR-2 fixes overwrite instances without completing the provenance rule.
   - **ARCH-MOCK: pass** — no new external production dependency.
   - **ARCH-CONSTRAINTS: pass** — counted index workload and repository scan support the stated scanning approach.
   - **ARCH-SECURE: pass** — parse failures propagate; no new credential handling.
   - **ARCH-ORDER: flag** — BR-2 leaves strictness state unchanged after valid overwrite transitions.
   - **ARCH-FUNERAL: pass** — no new durable runtime artifacts.

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining the complete decoder-proof invalidation rule, including aliases and range assignments, and naming the regression enumeration used to enforce it.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e8d57f8c-47c9-4b65-9b4b-56c4cc350a03","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8e9f946c7f1f5851","reviewed_head":"42793ebdf419c790073bcc343420225349588850","slot":"ariadne:1","requested_at":"2026-10-18T13:49:21.848231Z"},"phase":"finished","pid":62860,"started_at":"2026-10-18T13:49:21.942247Z","finished_at":"2026-10-18T13:51:13.914089Z","session_id":"01a14f46-2f8a-7193-bd3c-d7d7e7b8e026","outcome":"returned"}] -->
