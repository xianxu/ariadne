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

---

## Re-review — 2026-10-18T07:27:55-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..ebc57fade3ae567df3aec3f1268a9782b73611c6 |
| command | sdlc close --issue 440 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T07:27:55-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared syntax index, annotations, documentation, and regression matrix largely match the plan. However, BR-2 remains reproducible: assigning a fresh decoder restores proof even when an earlier alias can still overwrite its variable. The matcher consequently reports strict decoding where runtime decoding accepts unknown fields.

Done-when: current — it still covers the revised conservative matcher, annotations, cross-file resolution, and workload checks.

```findings
dispose:
  - id: BR-2
    disposition: not-addressed
    note: |
      pkg/golint/wire.go:185-186 restores decoder proof after reassignment without remembering that the variable's address escaped. Reproduced at pinned head with: d := json.NewDecoder(r); ref := &d; d = json.NewDecoder(r); d.DisallowUnknownFields(); *ref = json.NewDecoder(r); _ = d.Decode(&Peer{}). The matcher reports wire-strict-decode, while a runtime control decodes an unknown field successfully. Existing escape regressions fail when discardEscapedDecoders is removed, but omit this escape-before-reestablishment sequence. ARCH-PURPOSE / ARCH-ORDER: retain escape information for each lexical binding, including before its first tracked constructor; distinguish aliases to an old decoder value from aliases or closures retaining the variable. Sweep address-taking and binding-capture paths before allowing proof restoration. This repeats family wire-decoder-strictness-provenance; fix the governing rule rather than adding another overwrite special case.
  - id: BR-1
    disposition: addressed
    note: |
      Retained: wire.go:159 excludes initializer inference for explicitly typed interface destinations; assignment and range invalidation clear target identity. The corresponding interface and reassignment fixtures pass.
  - id: BR-3
    disposition: addressed
    note: |
      Retained: README.md:11-17 documents the annotation syntax, producer rationale, local-store exclusion, and atlas link. These agree with markedWireTypes and the atlas convention.
```

1. **Strengths**

   - `CheckSources` parses sources once and supplies one shared annotation index to all matchers.
   - Matcher tests exercise actual source parsing, cross-file/import resolution, suppression, and ambiguous flows without IO mocks.
   - All four required types carry annotations, and their existing consumers use tolerant `json.Unmarshal`.
   - README and atlas both explain the new convention and its coverage limits.

2. **Critical findings**

   **BR-2 remains open — `pkg/golint/wire.go:185-186`.** The following valid sequence produces a false positive:

   ```go
   d := json.NewDecoder(r)
   ref := &d
   d = json.NewDecoder(r)
   d.DisallowUnknownFields()
   *ref = json.NewDecoder(r)
   _ = d.Decode(&Peer{})
   ```

   A fresh decoder replaces the object, but does not revoke aliases to its variable. Preserve binding escape state across assignments, and add negative fixtures for address-taking and closure capture before proof establishment or restoration.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - `go test ./pkg/golint -count=1` passed.
   - Removing escape invalidation in a disposable copy made six existing regression cases fail, confirming the fix is exercised.
   - The additional binding-address probe failed against pinned head; its runtime control returned `Decode error: <nil>` for an unknown field.
   - Range whitespace checks passed. Temporary review files were removed; the checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — one shared index and existing suppression machinery.
   - **ARCH-PURE: pass** — index and matcher operate on source bytes/ASTs; filesystem collection remains outside.
   - **ARCH-PURPOSE: flag** — BR-2 still violates the promised conservative provenance contract.
   - **ARCH-MOCK: pass within scope** — no new external dependency; existing Git collection remains at the boundary.
   - **ARCH-CONSTRAINTS: pass** — counted index workload and focused repository scan support the bounded approach.
   - **ARCH-SECURE: pass** — parse errors propagate; no new credential or production-state access.
   - **ARCH-ORDER: flag** — BR-2 loses binding-escape knowledge across reassignment.
   - **ARCH-FUNERAL: pass** — no new durable runtime artifacts or background work.

7. **Plan revision recommendations**

   Add a `## Revisions` entry distinguishing decoder-object aliases from variable-binding aliases. State when proof may safely resume, and extend the regression matrix to escapes occurring before initial construction and before reassignment. Update the atlas’s “until a fresh direct assignment” claim to match that corrected rule.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"88d7039d-6900-402d-96a5-4583527aa1dc","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/828ff8ebde57f8f1","reviewed_head":"ebc57fade3ae567df3aec3f1268a9782b73611c6","slot":"ariadne:1","requested_at":"2026-10-18T14:26:15.537793Z"},"phase":"finished","pid":29657,"started_at":"2026-10-18T14:26:15.766551Z","finished_at":"2026-10-18T14:27:54.686121Z","session_id":"01a14f67-f86c-7b51-a0bd-1010b9ca566d","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T08:11:49-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..a60a6e6028c110b556781920e1b6cc2ed6a5b8cf |
| command | sdlc close --issue 440 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T08:11:49-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-2 is fixed, with regression tests that fail against the earlier implementation. The pure index, conservative decoder proof, annotations, and documentation are sound. One verified coverage defect remains: imported wire types declared at the module root silently bypass the rule.

Done-when: current — it still covers the revised conservative matcher, imported targets, annotations, documentation, and counted index workload.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      Whole-function eligibility rejects pointer writes, aliases, escapes, range assignments, and repeated assignments. The pinned-head regression matrix passes; substituting wire.go from 826f9776 makes the original indirect-overwrite fixtures fail, and substituting 848565fa makes the later binding-escape fixtures fail.
  - id: BR-1
    disposition: addressed
    note: |
      Target reassignment and explicit-interface negative fixtures pass at the pinned head; these fixtures fail with the original matcher from 826f9776.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:11 documents the wire annotation, independently evolving producers, and local-store exclusion, and links the supporting convention in atlas/workflow/sdlc-binary.md.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: wire-target-provenance
    title: |
      Imported marked types at the module root bypass the rule
    detail: |
      pkg/golint/wire.go:384 accepts only imports beginning with Module+"/", excluding an import equal to Module. A source-map probe with annotated Message in message.go and a direct strict decode into peer.Message from reader/read.go importing peer "example.test/module" returns no finding; moving the declaration and import to the wire subpackage produces the expected finding. This violates the Spec's module-wide imported-type coverage (ARCH-PURPOSE). This is the 2nd finding in family `wire-target-provenance`. Earlier rounds fixed instances; state and enforce the rule that every supported target identity resolves consistently. Sweep same-file, cross-file, module-root and subpackage imports, explicit/default aliases, and shadowed identities. Normalize an exact module import to the index's "." directory key and add regression coverage; that correction made the root/subpackage control probe pass.
```

1. **Strengths**

   - `singleOwnerDecoders` establishes eligibility before ordered matching, preventing stale strictness from being restored after an escape.
   - `CheckSources` shares parsing and indexing across files while keeping matcher tests free of IO.
   - All four required producer types carry annotations. README and atlas explain the convention and conservative coverage limits.

2. **Critical findings**

   - `pkg/golint/wire.go:384`: module-root imports are excluded. Reproduction and fix sketch appear above.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - `go test ./pkg/golint -count=1` passed.
   - BR-2 regression tests failed with earlier matcher versions in disposable copies.
   - The additional root-import probe failed on pinned head; the subpackage control passed. Normalizing root imports made both pass.
   - Range whitespace checks passed. Disposable copies were removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared index and suppression machinery.
   - **ARCH-PURE: pass** — declared pure entities exist and their direct tests require no IO.
   - **ARCH-PURPOSE: flag** — imported-type coverage excludes the module root.
   - **ARCH-MOCK: pass** — no new external dependency; tracked-source collection retains its existing IO boundary.
   - **ARCH-CONSTRAINTS: pass** — counted N/2N indexing tests and a passing repository scan support the intended workload.
   - **ARCH-SECURE: pass** — parse errors propagate; no new credential handling or unsafe external effects.
   - **ARCH-ORDER: pass** — decoder eligibility and ordered matching use explicit function-local state; no cross-event state is introduced.
   - **ARCH-FUNERAL: pass** — no durable runtime artifact is created.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining module-root and subpackage imports as supported target identities and recording the expanded identity-resolution regression matrix.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a07c7163-0fa0-4934-b10f-86cc1b19e050","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/f643b9c0c33e06b1","reviewed_head":"a60a6e6028c110b556781920e1b6cc2ed6a5b8cf","slot":"ariadne:1","requested_at":"2026-10-18T15:09:38.35845Z"},"phase":"finished","pid":77485,"started_at":"2026-10-18T15:09:38.473035Z","finished_at":"2026-10-18T15:11:48.628032Z","session_id":"01a14f8f-ae14-78c2-8a4c-52dddc1a0cbe","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T08:50:11-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..ae9feba152a8b629675f674bd5a754a9fef0bfac |
| command | sdlc close --issue 440 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T08:50:11-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-4 is verified: module-root imports resolve correctly, and removing the correction makes both root-import regression cases fail. The focused suite passes, and README/atlas cover the new convention. SHIP remains blocked by BR-2: the decoder ownership proof admits package-global bindings whose strictness can change outside the inspected function.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      Exact-module imports map to "." at pkg/golint/wire.go:392. TestWireTargetIdentityMatrix covers default and explicit root aliases; both cases fail when that correction is removed in a temporary package copy.
  - id: BR-1
    disposition: addressed
    note: |
      Retained disposition: target writes invalidate provenance, explicit interfaces do not inherit initializer identity, and the corresponding regression cases pass.
  - id: BR-2
    disposition: not-addressed
    note: |
      pkg/golint/wire.go:264-268 and 301-303 accept a package-global d assigned once inside the inspected function. Reproduced sequence: d = json.NewDecoder(r); d.DisallowUnknownFields(); reset(r); var m Peer; return d.Decode(&m), where reset replaces global d with json.NewDecoder(r). Check reports wire-strict-decode although the runtime decode accepts an unknown field with no error. This repeats family wire-decoder-strictness-provenance (ARCH-PURPOSE): enforce the documented function-local binding requirement before applying the single-owner proof, sweep same-file and cross-file globals alongside local declarations and shadowing, and add regression coverage.
```

Done-when: current — it still covers the annotation, four existing types, documented limits, provenance cases, and counted index workload.

1. **Strengths**

   - One shared syntax-only index serves all source files.
   - BR-4’s identity matrix covers root/subpackage imports, aliases, lookalike prefixes, and shadowing.
   - All four required tolerant inputs carry annotations.
   - README and atlas explain annotation usage and deliberately strict local contracts.

2. **Critical findings**

   - **BR-2 remains open — `pkg/golint/wire.go:264` and `:301`.** Assignment counts within one function do not establish exclusive ownership of a package-global decoder. Require a binding declared within that function before tracking strictness. Add same-file and cross-file global-reset regressions while preserving positive local-binding cases.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - `go test ./pkg/golint -count=1` passed.
   - BR-4 mutation check failed as expected for both module-root cases.
   - Independent lint/runtime probes confirmed the BR-2 false positive.
   - `git diff --check` passed. Full repository suite was not rerun.
   - Temporary probe copies were removed; reviewed files were unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared index and suppression handling.
   - **ARCH-PURE — pass:** index/matcher tests use source bytes without IO.
   - **ARCH-PURPOSE — flag:** BR-2 violates the promised conservative ownership rule.
   - **ARCH-MOCK — pass:** no new external dependency; existing collector seam remains.
   - **ARCH-CONSTRAINTS — pass:** counted N/2N index tests and focused repository scan pass.
   - **ARCH-SECURE — pass:** malformed source returns parse errors; no new credential handling.
   - **ARCH-ORDER — pass:** eligibility precedes ordered matching; the ownership defect is identified above.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining “local decoder” as a binding declared within the inspected function. Record the global-binding exclusion and regression matrix; the existing single-owner claim is currently stronger than its enforcement.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d98742e5-d605-4956-97b9-4d3776435e53","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/07b900c5f9e838f4","reviewed_head":"ae9feba152a8b629675f674bd5a754a9fef0bfac","slot":"ariadne:1","requested_at":"2026-10-18T15:48:25.369234Z"},"phase":"finished","pid":43031,"started_at":"2026-10-18T15:48:25.456178Z","finished_at":"2026-10-18T15:50:11.056506Z","session_id":"01a14fb3-2fd3-7522-a7e3-fe2005df4d10","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T09:29:49-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..2e2af4e0e8d9645aaa8009e78608abdd8ff5aded |
| command | sdlc close --issue 440 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T09:29:49-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-2 is addressed with meaningful regression coverage. The shared source index, annotations, and documentation match the intended design. One blocking gap remains: direct strict decodes in `var` initializers bypass the rule, despite having the same proven decoder and target as supported short declarations.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      Pointer-overwrite and global-reset regressions pass at HEAD and fail when the corresponding protections are removed in disposable copies.
  - id: BR-1
    disposition: addressed
    note: |
      Target-rebinding and interface fixtures pass; removing target invalidation makes direct, multiple, conditional, and range-rebinding regressions fail.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:11 now documents the wire annotation syntax, independently evolving producers, local-store exclusion, and atlas link, consistent with the matcher and atlas.
  - id: BR-4
    disposition: addressed
    note: |
      Module-root default-name and alias fixtures pass and both fail when exact-module import resolution is removed.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: wire-direct-call-coverage
    title: |
      Direct Decode calls in variable initializers bypass the rule
    detail: |
      pkg/golint/wire.go:143-161 processes declaration bindings without checking initializer calls. With an annotated Peer, `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; var err = d.Decode(&m); return err` produces no finding, while changing only `var err =` to `err :=` produces one. Typed and grouped var declarations also bypass detection. ARCH-PURPOSE: enumerate supported unconditional expression positions, check direct Decode calls consistently before applying binding changes, and add regression fixtures for all declaration forms.
```

Done-when: current — its direct strict-decoding requirement covers this gap; the ownership revisions refine conservative provenance without changing the completion criteria.

1. **Strengths**
   - `CheckSources` parses each source once and shares one index across matchers (`pkg/golint/check.go:67`).
   - Whole-function ownership checks conservatively reject decoder aliases, escapes, overwrites, and global bindings.
   - README and atlas explain the annotation, coverage limits, and intentional strict local-store decoding.
   - All four required production types carry annotations, enforced by a repository test.

2. **Critical findings**
   - `pkg/golint/wire.go:143`: declaration initializers skip `checkCall`. The disposable probe reproduced missing findings for inferred, explicitly typed, and grouped `var` declarations. Apply call checking consistently across unconditional expression positions.

3. **Important findings**
   - None.

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - `go test ./pkg/golint -count=1` passed, including the tracked-source catalog scan.
   - `git diff --check` passed.
   - Mutation checks established regression sensitivity for BR-1, BR-2, and BR-4.
   - The new declaration probe failed at pinned HEAD as described above.
   - Disposable copies were removed; the review checkout remains unchanged. The full repository suite was not rerun.

6. **Architectural notes**
   - **ARCH-DRY: pass** — shared index and suppression path.
   - **ARCH-PURE: pass** — core entities operate on supplied source bytes and ASTs; repository collection owns IO.
   - **ARCH-PURPOSE: flag** — direct-call coverage depends on declaration syntax.
   - **ARCH-MOCK: pass** — no new external dependency or production external-call seam.
   - **ARCH-CONSTRAINTS: pass** — single parse loop and counted N/2N index tests support the declared workload.
   - **ARCH-SECURE: pass** — malformed source returns parser errors; no credentials introduced.
   - **ARCH-ORDER: pass** — no state persists between invocations; conservative scans govern statement-local proof.
   - **ARCH-FUNERAL: pass** — no new runtime durable artifacts.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry enumerating supported unconditional call positions, including variable initializers, and require equivalent declaration forms to receive equivalent lint results.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"746c61e2-145a-4c3c-9cb0-96808e3bb524","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":6,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8d48ef835946e31e","reviewed_head":"2e2af4e0e8d9645aaa8009e78608abdd8ff5aded","slot":"ariadne:1","requested_at":"2026-10-18T16:27:49.371898Z"},"phase":"finished","pid":29786,"started_at":"2026-10-18T16:27:49.461672Z","finished_at":"2026-10-18T16:29:49.172195Z","session_id":"01a14fd7-4248-7123-bbfe-7f8db28dd49b","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T09:49:59-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..dd1e08a0ef40df3b15ee27d5ba716ba6ed04485b |
| command | sdlc close --issue 440 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T09:49:59-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-5’s declaration-initializer correction is verified: its regression test passes at the pinned head and fails in four cases with the pre-fix matcher. One related coverage gap remains: direct expression calls in `if` initializers bypass the rule despite being included in the documented contract.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Retain the accepted disposition; target reassignment, interface, and shadowing fixtures pass.
  - id: BR-2
    disposition: addressed
    note: |
      Retain the accepted disposition; whole-function ownership checks and alias, escape, reassignment, and global-binding fixtures pass.
  - id: BR-3
    disposition: addressed
    note: |
      README.md documents the annotation syntax and links the atlas convention, consistent with annotation parsing in wire.go.
  - id: BR-4
    disposition: addressed
    note: |
      Exact-module imports resolve to "."; the root/subpackage identity matrix passes.
  - id: BR-5
    disposition: addressed
    note: |
      TestWireStrictDecodeDirectCallPositions passes at head. Substituting the pre-fix matcher in a temporary copy makes inferred, typed, grouped var, and multiple-RHS cases fail; the copy was removed.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: wire-direct-call-coverage
    title: |
      Direct expression calls in if initializers bypass the rule
    detail: |
      pkg/golint/wire.go:200 only checks an if initializer when it is an AssignStmt. With an eligible strict decoder and marked destination, `if d.Decode(&m); true {}` produces no finding. Likewise, `if d.DisallowUnknownFields(); true {}; return d.Decode(&m)` loses the unconditional strictness call. Both were reproduced against the pinned head without parse errors; the assignment-initializer control reports correctly. ARCH-PURPOSE: the plan and atlas explicitly include direct calls in if initializers. This is the 2nd finding in family `wire-direct-call-coverage`. Do NOT fix only these instances: state the rule for every supported direct-call position, route supported initializer statements through shared handling, and test the expression/assignment forms with strictness-order controls.
```

Done-when: current — its annotation, rule, provenance, import, and workload requirements still cover the revised implementation.

### 1. Strengths

- One shared source index serves all matchers; the repository collector remains the IO boundary.
- Whole-function decoder eligibility makes the conservative ownership rule explicit.
- Tests cover import identity, provenance invalidation, suppression, generated sources, and annotation boundaries.
- README and atlas both document the new convention.

### 2. Critical findings

- **Direct `if` initializer expressions are skipped**, [wire.go:200](/Users/xianxu/Library/Caches/sdlc/review-checkouts/c794b93980d18ebe/pkg/golint/wire.go:200). Handle supported initializer statements consistently through the shared call checker; add regression cases for both examples above.

### 3. Important findings

None.

### 4. Minor findings

None.

### 5. Test coverage notes

`go test ./pkg/golint -count=1` and range whitespace checks pass. BR-5’s regression was independently confirmed red without its fix. Existing tests cover assignment-form `if` initializers but omit expression-form initializers. The full repository suite was not rerun.

### 6. Architectural notes

- **ARCH-DRY: pass** — shared indexing, call checking, and suppression logic.
- **ARCH-PURE: pass** — index and matcher tests operate on source bytes without IO.
- **ARCH-PURPOSE: flag** — documented direct-call coverage remains incomplete.
- **ARCH-MOCK: pass** — no new external dependency; existing tracked-source collection remains at the boundary.
- **ARCH-CONSTRAINTS: pass** — counted N/2N index coverage and successful repository scan.
- **ARCH-SECURE: pass** — malformed source returns parsing errors; no new credential handling.
- **ARCH-ORDER: pass** — strictness is tracked in statement order with explicit eligibility and invalidation.
- **ARCH-FUNERAL: pass** — no new durable runtime artifacts.

### 7. Plan revision recommendations

Add a `## Revisions` entry enumerating supported direct-call positions and initializer statement forms. Extend the regression matrix to expression-form `if` initializers and their strictness ordering before changing the matcher.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3174a9bd-91c4-4520-94b8-6f321625d2c4","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":7,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c794b93980d18ebe","reviewed_head":"dd1e08a0ef40df3b15ee27d5ba716ba6ed04485b","slot":"ariadne:1","requested_at":"2026-10-18T16:48:18.615887Z"},"phase":"finished","pid":17954,"started_at":"2026-10-18T16:48:18.738582Z","finished_at":"2026-10-18T16:49:59.396248Z","session_id":"01a14fea-043b-7001-9335-fa8de676f891","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T09:56:17-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 440 — lint: a wire-message annotation and a rule against strict decoding into wire types |
| repo | ariadne |
| issue file | workshop/issues/000440-lint-a-wire-message-annotation-and-a-rule-against-strict-decoding-into-wire-types.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3ba12c88814b92d93a6e30300fc8061940382d14..31846daad25883cbd11e26ba451171ccb8acaee8 |
| command | sdlc close --issue 440 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T09:56:17-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The annotation index, repository integration, and regression coverage are sound for the tested cases. BR-6 is withdrawn against the explicitly revised coverage contract; its disputed measurements reproduce. One new correctness issue blocks shipping: a jump can bypass strictness while the matcher still reports a tolerant decoder as strict.

Done-when: current — the criteria still cover the annotation, bounded decoder rule, four existing types, documentation, and test matrix after the initializer scope revision.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Target reassignment and interface regressions pass at head and fail with wire.go from 1e51be06^ substituted in a disposable copy.
  - id: BR-2
    disposition: addressed
    note: |
      The existing decoder-overwrite defect remains corrected. Local-ownership regressions pass at head and fail for same-file and cross-file global replacement with wire.go from 4fdbe641^.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:11-18 documents the wire annotation syntax and links the atlas; atlas/workflow/sdlc-binary.md:1918 onward explains the matching boundary and strict local-store exception.
  - id: BR-4
    disposition: addressed
    note: |
      The identity matrix passes at head. Substituting wire.go from 364505a5^ fails module-root default/alias imports and the shadowed-subpackage control.
  - id: BR-5
    disposition: addressed
    note: |
      Direct-call position tests pass at head; substituting wire.go from 01d74b72^ fails inferred, typed, and grouped var initializers and multiple assignment RHS.
  - id: BR-6
    disposition: withdrawn
    note: |
      Reproduced one finding for assignment-form if initialization and zero for both expression-form cases. Spec lines 46-49, plan Revisions lines 226-233, and atlas explicitly exclude the expression form under the documented scope revision. This is a prose scope correction, not an executable fix.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: wire-decoder-strictness-provenance
    title: |
      A goto that bypasses strictness produces a false wire-decode finding
    detail: |
      pkg/golint/wire.go:110-141 assumes the direct statement list executes in lexical order; lines 205-207 only invalidate assignments. With an eligible decoder and marked target, `goto decode; d.DisallowUnknownFields(); decode: ; return d.Decode(&m)` reports one finding although strictness never executes. A compiled runtime probe accepts an unknown field with no error; a conditional-jump variant also reports despite ambiguous strictness. This violates Spec lines 39-42 (ARCH-PURPOSE). This is the 2nd finding in family `wire-decoder-strictness-provenance`: enumerate the proof-invalidating control transfers and enforce that rule, rather than patching only this example. Conservatively rejecting functions containing goto is sufficient without adding control-flow analysis; add unconditional/conditional jump regressions and a straight-line positive control.
```

1. **Strengths**

   - `CheckSources` parses and indexes the source set once; existing rules retain shared suppression handling.
   - Pure source-map tests cover imported identities, shadowing, provenance invalidation, and direct-call positions.
   - All four requested tolerant input types carry annotations. README and atlas explain the convention and its limits.

2. **Critical findings**

   - The jump-related false positive above originates at `pkg/golint/wire.go:110`. Skip unsupported jump-bearing flows or prove strictness across their paths before reporting.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed `go test ./pkg/golint -count=1`, including the tracked-source scan.
   - Passed tests for `cmd/sdlc/internal/usage` and `cmd/sdlc/internal/judge`.
   - Verified prior behavioral regressions fail without their corrections in disposable copies, subsequently removed.
   - The new jump probes reproduce the defect; committed coverage is missing.
   - `git diff --check` passed. Full suite not rerun; changes after its recorded code head are documentation only. Checkout remains unchanged.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY: pass** — one shared index and direct-call checker.
   - **ARCH-PURE: pass** — listed pure entities exist and their source-map tests require no IO; collection remains separate.
   - **ARCH-PURPOSE: flag** — the conservative strictness guarantee fails for jumps.
   - **ARCH-MOCK: pass** — no new production binary or service dependency.
   - **ARCH-CONSTRAINTS: pass** — counted N/2N index coverage and repository scan exercise the declared workload.
   - **ARCH-SECURE: pass** — parse failures propagate visibly; no new credential handling.
   - **ARCH-ORDER: pass** — analysis state is scoped to one invocation; no state persists between external events.
   - **ARCH-FUNERAL: pass** — no new durable runtime artifacts.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining which control transfers invalidate lexical strictness proof, the conservative exclusion policy, and the regression matrix for jumps plus supported straight-line controls.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5e7575af-c635-4e55-886a-6f6651d60817","repository":"github.com/xianxu/ariadne","issue_id":"000440","recipe":"milestone-review","round":8,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0b8bfdbf6eec1396","reviewed_head":"31846daad25883cbd11e26ba451171ccb8acaee8","slot":"ariadne:1","requested_at":"2026-10-18T16:53:13.155843Z"},"phase":"finished","pid":88307,"started_at":"2026-10-18T16:53:13.285483Z","finished_at":"2026-10-18T16:56:17.019925Z","session_id":"01a14fee-82ff-7eb1-9a31-5a8d2838a6fa","outcome":"returned"}] -->
