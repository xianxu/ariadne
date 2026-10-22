# Boundary Review — ariadne#507 (whole-issue close)

| field | value |
|-------|-------|
| issue | 507 — workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible |
| repo | ariadne |
| issue file | workshop/issues/000507-workingdoc-datatype-and-tl-lead-doc-settle-status-settle-side-docs-into-tracked-copies-re-settle-to-update-drift-visible.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..f8d5927d1879cdae1f12ae63c0f402a30db3a7ef |
| command | sdlc close --issue 507 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T14:02:54-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The publication guards, shared projection, and documentation are substantial and mostly coherent. Two correctness gaps block shipping: settlement can write an unreadable pointer, and first settlement can overwrite an unrelated committed destination. Recovery coverage and test isolation also need correction.

1. **Strengths**

   - Publication binds the captured payload to SHA-256 and verifies the committed blob before stamping the source.
   - Ops publication uses a private index, checks the selected tree, and preserves concurrent index changes.
   - Settlement and status share projection logic; unknown frontmatter bytes survive pointer edits.
   - README, atlas, help, and datatype documentation cover the new surface.

2. **Critical findings**

   - **Pointer serialization:** [doc_model.go:113](/Volumes/TestScratch/review-checkouts/6c9313a65937d49f/cmd/tl-lead/doc_model.go:113) inserts allowed paths directly into a plain YAML scalar. For `workshop/workingdoc/a #b.md`, YAML reads the pointer value as `workshop/workingdoc/a`, discarding the commit as a comment. Settlement can report success after corrupting its own managed metadata. Validate the generated scalar through the production parser before publication; refuse paths that cannot satisfy the plain-scalar contract. **ARCH-SECURE, ARCH-EVOLVE.**
   - **First-settlement ownership:** [doc_settle.go:349](/Volumes/TestScratch/review-checkouts/6c9313a65937d49f/cmd/tl-lead/doc_settle.go:349) accepts an occupied local destination when it equals the desired payload. With no pointer, product publication then uses ordinary trunk semantics, which allow replacing different committed content when main still matches the branch baseline. Enforce absent-or-identical remote content inside every publication CAS attempt for first settlement. **ARCH-PURPOSE.**

3. **Important findings**

   - **Recovery coverage:** [doc_fake_test.go:31](/Volumes/TestScratch/review-checkouts/6c9313a65937d49f/cmd/tl-lead/doc_fake_test.go:31) cannot simulate Apply succeeding and losing its acknowledgment. The coordinator tests also omit a lock arriving after publication and recovery on retry, despite the checked plan claiming these sequences. Add deterministic fault injection and exercise the production coordinator’s readback/retry branches. **ARCH-ORDER.**
   - **Host-binary dependency:** [doc_operating_test.go:28](/Volumes/TestScratch/review-checkouts/6c9313a65937d49f/cmd/tl-lead/doc_operating_test.go:28) leaves `sdlc agent-id` dependent on the installed host binary. The isolated test fails when that binary is absent. Inject or provide a controlled identity fixture through the production seam. **ARCH-MOCK.**

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/tl-lead ./cmd/datatype -count=1`
   - `go test ./cmd/sdlc -run 'Trunk.*Publish|WorkingdocTrunk' -count=1`
   - Pinned-range `git diff --check`

   Isolation check failed:

   ```text
   PATH=/usr/bin:/bin /opt/homebrew/bin/go test ./cmd/tl-lead \
     -run '^TestOperatingPublishExactPayloadAndRetry$' -count=1

   read commit agent identity: exec: "sdlc": executable file not found in $PATH
   ```

   Neither critical case has a regression test. The existing collision fixture covers different **local** bytes, not different committed bytes hidden by matching local content. No full-suite or live-Couch verification was performed during this review. Repository files remain unchanged.

6. **Architectural notes for upcoming work**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — shared projection and existing trunk publisher reused. |
   | ARCH-PURE | Pass — metadata, markers, projection, and classification have direct pure tests. |
   | ARCH-PURPOSE | Flag — first-settlement collision protection is incomplete. |
   | ARCH-MOCK | Flag — ops tests invoke the installed host `sdlc`. |
   | ARCH-CONSTRAINTS | Pass — bounded document reads, batched baselines, and cancellation checks exist. |
   | ARCH-SECURE | Flag — accepted paths can produce invalid managed YAML. |
   | ARCH-ORDER | Flag — claimed uncertain-Apply and late-lock recovery sequences lack coverage. |
   | ARCH-FUNERAL | Pass — temporary resources are scoped; the retained lock inode has a documented constant cost. |
   | ARCH-EVOLVE | Flag — generated known-field content can become unreadable; unknown-field preservation otherwise holds. |

7. **Plan revision recommendations**

   Append `## Revisions` entries recording:

   - Plain-scalar representability validation before publication, including adversarial destination paths.
   - First-settlement remote ownership enforcement across CAS retries.
   - The actual coordinator interruption-test inventory and isolated agent-identity seam.

   Also correct the testing table’s nonexistent `pinnedProsePublishDecision` reference to the actual reused `prosePublishDecision` and its integration caller.

```findings
findings:
  - id: new
    severity: Critical
    family: managed-metadata-roundtrip
    title: |
      Allowed destination paths can produce an unreadable settlement pointer
    detail: |
      cmd/tl-lead/doc_model.go:113 writes paths directly into a plain YAML scalar, but workingdocPath accepts "workshop/workingdoc/a #b.md". YAML treats "#b.md @ COMMIT" as a comment, so settlement can publish, stamp, and report success while subsequent status and settle reject the source. Validate generated metadata with the production parser before effects and refuse unrepresentable paths; add round-trip regressions. ARCH-SECURE, ARCH-EVOLVE.
  - id: new
    severity: Critical
    family: first-settlement-destination-ownership
    title: |
      First product settlement can overwrite an unrelated committed destination
    detail: |
      cmd/tl-lead/doc_settle.go:349 accepts an occupied local target matching the payload and publishes with no expected commit. If committed main and the branch baseline contain different content, cmd/sdlc/trunkpublish.go:283 permits replacing it through ordinary edit semantics. A locally deleted tracked target also bypasses the collision check. Enforce absent-or-identical remote content for first settlement inside each CAS attempt, preserving ordinary trunk behavior; add product regressions. ARCH-PURPOSE.
  - id: new
    severity: Important
    family: settlement-interruption-coverage
    title: |
      Claimed coordinator recovery sequences are missing from the tests
    detail: |
      workshop/plans/000507-workingdoc-settle-plan.md:120 marks uncertain Apply and late-lock sequences delivered, but doc_fake_test.go:31 cannot return a lost acknowledgment after applying, and doc_settle_test.go only injects a source edit after publication. Add controlled Apply outcomes and late-lock/retry sequences through settleWorkingdoc, asserting confirmed pointer evidence and no duplicate publication. ARCH-ORDER.
  - id: new
    severity: Important
    family: integration-test-dependency-isolation
    title: |
      Ops publication tests require an installed host sdlc binary
    detail: |
      doc_operating_test.go:28 does not provide a controlled sdlc agent-id dependency for doc_operating.go's operatingAgentTrailer. Running PATH=/usr/bin:/bin /opt/homebrew/bin/go test ./cmd/tl-lead -run '^TestOperatingPublishExactPayloadAndRetry$' -count=1 fails with executable "sdlc" not found. Inject or fixture the identity dependency through the production seam so portable tests do not depend on an unrelated installed revision. ARCH-MOCK.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ab9c7877-c39d-4050-9048-73e8c522bff3","repository":"github.com/xianxu/ariadne","issue_id":"000507","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/6c9313a65937d49f","reviewed_head":"f8d5927d1879cdae1f12ae63c0f402a30db3a7ef","slot":"ariadne:4","requested_at":"2026-10-21T20:59:23.610973Z"},"phase":"finished","pid":70818,"started_at":"2026-10-21T20:59:23.727838Z","finished_at":"2026-10-21T21:02:53.636432Z","session_id":"01a16042-f81e-75d2-819e-85b7984b7adc","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T19:04:41-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 507 — workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible |
| repo | ariadne |
| issue file | workshop/issues/000507-workingdoc-datatype-and-tl-lead-doc-settle-status-settle-side-docs-into-tracked-copies-re-settle-to-update-drift-visible.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..925290911ba2a36408fdbf416ee06922ad5eee0a |
| command | sdlc close --issue 507 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T19:04:41-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior findings are addressed with meaningful regression evidence. Existing tests pass, but a new coordinator probe demonstrates that unresolved human commentary can be published and stamped successfully. The settlement state model also does not enforce the effect ordering promised by the plan.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      doc_model.go:120 reparses generated metadata before effects. Adversarial pointer and coordinator refusal tests pass; removing the roundtrip guard in the disposable checkout makes them fail.
  - id: BR-2
    disposition: addressed
    note: |
      doc_repository.go supplies --expect-absent for first settlement, and trunkpublish.go checks ownership inside every CAS callback. Real-remote ownership tests pass; disabling that predicate reproduces successful replacement of unrelated committed content.
  - id: BR-3
    disposition: addressed
    note: |
      doc_fake_test.go now supports pre-write failure and post-write lost acknowledgment. Coordinator tests exercise late locks, unavailable readback, intervening edits and duplicate-free retries. Disabling uncertain-Apply reconciliation makes the lost-ack regression fail.
  - id: BR-4
    disposition: addressed
    note: |
      doc_operating_test.go:28 supplies a strict fixture sdlc agent-id executable. All operating tests pass with PATH=/usr/bin:/bin; removing the fixture PATH injection reproduces the missing-sdlc failure.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: marker-literal-boundaries
    title: |
      Backticks across Markdown blocks hide unresolved review items
    detail: |
      cmd/tl-lead/doc_markers.go:171 searches for matching backticks across every non-fenced region, and line 206 marks the intervening blocks literal. Unmatched backticks in separate paragraphs, headings or list items therefore hide an intervening human question. Three coordinator probes confirmed zero actual CodeSpan nodes using the repository's Goldmark dependency, yet settleWorkingdoc returned success with Published=true and Stamped=true. ARCH-PURPOSE, ARCH-SECURE: restrict literal masking to actual Markdown code contexts and add scanner plus no-effects coordinator regressions across block boundaries, preserving legitimate multiline spans.
  - id: new
    severity: Important
    kind: defect
    family: state-transition-authority
    title: |
      Settlement transitions observe effects instead of authorizing them
    detail: |
      cmd/tl-lead/doc_settle.go:359 publishes before consulting transitionSettlement, and line 395 applies the source stamp before its transition. doc_model.go:153 returns no effects, while failure and unconfirmed events have no production consumers. This contradicts the plan's line 49 claim that only the transition model grants the next effect. ARCH-ORDER: make the pure model authorize effects and consume their outcomes, route failures through it, and test forbidden effect sequences through the production coordinator.
```

1. **Strengths**

   - Pointer generation validates against the production reader before publication and preserves unknown metadata bytes.
   - First-settlement ownership is enforced within CAS retries, including concurrent destination creation.
   - Recovery fixtures now model uncertain outcomes and verify exact pointer evidence without duplicate publication.
   - README, atlas and command help cover the new user-facing surface.

2. **Critical findings**

   The marker defect at `cmd/tl-lead/doc_markers.go:171` permits settlement of this unresolved document body:

   ```text
   An unmatched ` here.

   🤖[Do not settle before answering this.]

   A separate paragraph with ` here.
   ```

   The same failure reproduces across headings and separate list items. Use Markdown block-aware literal ranges; Goldmark is already a repository dependency.

3. **Important findings**

   At `cmd/tl-lead/doc_settle.go:359`, effect authority remains in the coordinator’s imperative sequence. The pure state machine records successful effects afterward. Make publication, verification and stamping effects originate from transitions, including failure and uncertainty handling.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `cmd/tl-lead`: 72.540s.
   - Focused workingdoc/datatype tests.
   - Existing/new trunk publication tests: 31.135s.
   - All operating tests under restricted PATH: 17.044s.

   Mutation checks established regression sensitivity for the four prior findings. Three new block-boundary probes failed against pinned HEAD. The disposable checkout was removed; the reviewed checkout remains unchanged. I did not rerun the full repository suite.

6. **Architectural notes for upcoming work**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass: shared projection, metadata handling and existing publisher reused. |
   | ARCH-PURE | Pass: core parsing/projection tests require no IO. |
   | ARCH-PURPOSE | Flag: unresolved commentary can cross the settlement boundary. |
   | ARCH-MOCK | Pass: stateful recovery and isolated Git fixtures; live conformance remains planned after SHIP. |
   | ARCH-CONSTRAINTS | Pass: bounded documents, batched baselines and cancellation coverage. |
   | ARCH-SECURE | Flag: incorrect literal classification accepts unresolved input. |
   | ARCH-ORDER | Flag: transition model lacks effect authority. |
   | ARCH-FUNERAL | Pass: temporary resources are scoped; durable copies and constant-size lock have documented lifecycles. |
   | ARCH-EVOLVE | Pass: metadata edits preserve unknown nested fields and reject ambiguous managed values. |

   Core-concept locations and PURE/INTEGRATION classifications otherwise match the implementation.

7. **Plan revision recommendations**

   Add specific `## Revisions` entries covering:

   - Markdown block-aware literal detection, with paragraph, heading and list-boundary regressions.
   - An authoritative settlement transition/effect contract replacing the currently unenforced claim at plan line 49.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"af5eed74-e535-412b-9833-04db67d9f721","repository":"github.com/xianxu/ariadne","issue_id":"000507","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5bfd30daaf9753de","reviewed_head":"925290911ba2a36408fdbf416ee06922ad5eee0a","slot":"ariadne:4","requested_at":"2026-10-22T01:59:43.319496Z"},"phase":"finished","pid":16075,"started_at":"2026-10-22T01:59:43.479128Z","finished_at":"2026-10-22T02:04:40.239309Z","session_id":"01a16155-ee04-7892-ba23-4bea4502c786","outcome":"returned"}] -->
