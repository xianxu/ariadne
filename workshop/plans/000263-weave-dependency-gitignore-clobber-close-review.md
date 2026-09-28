# Boundary Review — ariadne#263 (whole-issue close)

| field | value |
|-------|-------|
| issue | 263 — weave refresh clobbers a dependency's committed gitignore block |
| repo | ariadne |
| issue file | workshop/issues/000263-weave-dependency-gitignore-clobber.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7fc9a236ac087bb67870cd58cd7009ce04bd8a43..882434878417da36b8b0f8027619898dcc3a4734 |
| command | sdlc close --issue 263 |
| reviewer | claude |
| timestamp | 2026-09-28T15:44:15-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-09-28T15:46:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 263 — weave refresh clobbers a dependency's committed gitignore block |
| repo | ariadne |
| issue file | workshop/issues/000263-weave-dependency-gitignore-clobber.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7fc9a236ac087bb67870cd58cd7009ce04bd8a43..882434878417da36b8b0f8027619898dcc3a4734 |
| command | sdlc close --issue 263 |
| reviewer | claude |
| timestamp | 2026-09-28T15:46:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The code fix is correct and goes after the root cause. `managedIgnore` now keeps any existing block entry that the checkout's prior inventory did not own. It still retires entries that inventory owned, and it still moves legacy entries that sit outside the block. The fix is in `ApplyManaged`, so it covers every caller, including a dependency that has data mounts. It does not work by skipping dependency owners in `compilePrepared`. `go test ./cmd/weave/...` passes (I ran it). What stops a plain SHIP is the issue's own contract: the Spec, a Done-when clause and a ticked Plan item still say legacy entries inside the block migrate away. The Log's design correction deliberately dropped that behaviour, and the code no longer does it. The contract needs a `## Revisions` entry so it matches what shipped.

1. **Strengths**
   - `ownership.go:174`: the old inventory (all scopes) is passed into `managedIgnore`. Ownership comes from real provenance, so this is a proper fix rather than a special case for dependencies.
   - `managed_ignore.go:107-115`: the owned set is built with `escapeIgnore`, the same form used for block lines, so the comparison lines up (no mismatch between escaped and raw paths).
   - `TestManagedPreservesCommittedBlockWithoutInventory` checks that `.gitignore` stays byte-identical. That is exactly the Done-when regression, and it includes the real-world `/AGENTS.md` entry that is also on the legacy list.
   - `TestManagedRetiresOnlyOwnedBlockEntries` covers both states: an entry that belongs to someone else survives, and an owned entry is added and then retired.
   - Moving legacy entries from outside the block is still covered by the existing test at `ownership_test.go:198`.

2. **Critical:** none.

3. **Important**
   - The issue's contract contradicts the implementation on legacy entries inside the block. Every place in the family:
     - (a) Spec: "Exception: the legacy fixed list … still migrates away, as today"
     - (b) Done-when clause 3: "Legacy fixed-list entries inside the block still migrate away"
     - (c) Plan item 2, ticked: "legacy entry inside block migrates"

     The code now keeps such entries when they are unowned, and no test covers (b) or (c) as written. Fix: add a `## Revisions` entry that restates the clause as "legacy entries outside the block still migrate (ownership_test.go:198); entries inside the block follow the ownership rule", with the reason from the Log.

4. **Minor**
   - ARCH-DRY: `blockEntries` repeats `managedIgnoreText`'s BEGIN/END parsing loop. A single parse that returns both the kept text and the entries inside the block would remove the duplicate.
   - The regression test covers a *missing* inventory, not an explicit empty `{"outputs": []}` file, which is the state actually observed. It is probably the same code path, but a one-line variant would pin it.

5. **Test coverage:** Both states of the ownership rule are tested (owned entries retire, unowned entries are preserved). Outside-block legacy migration is still covered. The inside-block legacy clause is untested, and under the implemented design it no longer applies (see Important).

6. **Architecture**
   - ARCH-DRY: flagged as Minor (the duplicated block parser).
   - ARCH-PURE: pass. `blockEntries` is pure, and the IO in `managedIgnore` is still a thin read.
   - ARCH-PURPOSE: pass. The rule is fixed where it lives, and a dependency with mounts is covered.
   - Accepted consequence: stale upstream entries stay behind in checkouts that have no inventory. They are harmless and documented.

7. **Plan revisions:** add a `## Revisions` entry rewriting the Spec exception, Done-when clause 3 and Plan item 2 as described in Important.

```findings
findings:
  - id: new
    severity: Important
    family: contract-drift-from-design-correction
    title: |
      Spec, Done-when clause 3 and Plan item 2 still claim legacy entries inside the block migrate; the code now keeps them
    detail: |
      Three places: the Spec "Exception: legacy fixed list still migrates away", Done-when "Legacy fixed-list entries inside the block still migrate away", and the ticked Plan item "legacy entry inside block migrates". The Log's design correction dropped this rule and no test covers it. Add a Revisions entry restating it as outside-block migration (covered by ownership_test.go:198).
  - id: new
    severity: Minor
    family: duplicated-block-parser
    title: |
      blockEntries repeats the BEGIN/END parse loop from managedIgnoreText (ARCH-DRY)
    detail: |
      managed_ignore.go:71 vs managedIgnoreText. A single parse returning the kept text plus the entries inside the block would remove the duplicate.
  - id: new
    severity: Minor
    family: regression-fixture-fidelity
    title: |
      Regression test covers a missing inventory, not an explicit empty outputs inventory file
    detail: |
      The observed state was an inventory file containing an empty outputs list. It is likely the same code path, but a one-line variant would pin it.
```

---

## Re-review — 2026-09-28T15:49:37-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 263 — weave refresh clobbers a dependency's committed gitignore block |
| repo | ariadne |
| issue file | workshop/issues/000263-weave-dependency-gitignore-clobber.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7fc9a236ac087bb67870cd58cd7009ce04bd8a43..7b31373ab682a5426a8d404848c2a01f5330f01e |
| command | sdlc close --issue 263 |
| reviewer | claude |
| timestamp | 2026-09-28T15:49:37-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All three open findings from earlier rounds are fixed, and I found nothing new that blocks the close. `managedIgnore` now removes a block entry only when the old inventory shows the checkout owned it (`managed_ignore.go:103-115`). The old inventory it uses is the one `ApplyManaged` already loads, covering all scopes (`ownership.go:174`). The issue's contract now matches the code. The Spec, Done-when clause 3 and Plan item 2 were rewritten, and a `## Revisions` entry records why. The duplicated BEGIN/END parsing now lives in one function, `splitIgnore`, which both callers use. The regression test now covers both a missing inventory and an inventory file with an empty `outputs` list. I ran `go test ./cmd/weave/...` at head and it passes.

1. **Strengths**
   - **Ownership check in the right place.** The rule is fixed inside `ApplyManaged`, not by skipping dependency owners in `compilePrepared`. So a dependency that does have data mounts is covered too.
   - **One parser.** `splitIgnore` (`managed_ignore.go:19-56`) is the only BEGIN/END parser. It returns both the text outside the block and the entries inside it. `managedIgnoreText` and `managedIgnore` both call it, and malformed-block errors are raised the same way for both.
   - **Paths compared in the same form.** Ownership keys are built with `escapeIgnore(id.Path)`, the same escaping used for block lines, so the comparison lines up.
   - **Regression test checks exact bytes.** `TestManagedPreservesCommittedBlockWithoutInventory` requires `.gitignore` to stay byte-identical when the inventory is missing and when it has empty `outputs`. The fixture includes `/AGENTS.md`, which is also on the legacy list — the real-world case that broke the first draft.
   - **Both sides of the rule tested.** `TestManagedRetiresOnlyOwnedBlockEntries` checks that an entry the checkout doesn't own survives, and that an owned entry is added and later retired.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - In `managed_ignore.go:111`, the loop variable in `for _, e := range block` shadows the error variable `e` from the line above. The code is correct, but the name makes it harder to read; something like `entry` would be clearer. This is a one-off, not a pattern elsewhere in the diff.

5. **Test coverage**
   - Done-when clause 1: covered for both a missing and an empty inventory.
   - Clause 2: covered by the retire test plus the existing tests.
   - Clause 3: covered by `TestManagedIgnoreMigrationAndLocalNegations` (`ownership_test.go:196`, confirmed present).
   - Clause 4: the full weave test suite passes.

6. **Architecture**
   - **ARCH-DRY:** pass, now that the parser is shared.
   - **ARCH-PURE:** pass. `splitIgnore` and `managedIgnoreText` are pure functions, and `managedIgnore` only adds a thin file read around them.
   - **ARCH-PURPOSE:** pass. The ownership rule is fixed where it lives, so every caller gets it, not just the dependency path the bug report named. The known side effect is documented in the Spec: in a checkout with no inventory, an ignore line for an output removed upstream stays behind. That is harmless.

7. **Plan revisions:** none needed; the Revisions entry already covers the change.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Spec, Done-when 3 and Plan item 2 now describe outside-block migration only; a Revisions entry explains why, and the named test exists at ownership_test.go:196.
  - id: BR-2
    disposition: addressed
    note: |
      splitIgnore is the only BEGIN/END parser, used by managedIgnoreText and managedIgnore; blockEntries is gone.
  - id: BR-3
    disposition: addressed
    note: |
      The regression test loops over both a missing inventory and an explicit empty-outputs inventory file.
findings:
  - id: new
    severity: Minor
    family: identifier-shadowing
    title: |
      The loop variable e in managedIgnore shadows the error variable e from the line above
    detail: |
      managed_ignore.go:111 has `for _, e := range block`. The code is correct but the name is confusing; renaming it to entry fixes it. This is the only instance in the window.
```

---

## Re-review — 2026-09-28T15:53:50-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 263 — weave refresh clobbers a dependency's committed gitignore block |
| repo | ariadne |
| issue file | workshop/issues/000263-weave-dependency-gitignore-clobber.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7fc9a236ac087bb67870cd58cd7009ce04bd8a43..9d1c94e2d7bbe3053b908ca1a51d9691fa87e4d3 |
| command | sdlc close --issue 263 |
| reviewer | claude |
| timestamp | 2026-09-28T15:53:50-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Summary.** I checked the one open finding, BR-4, against the head commit. It is fixed: `managed_ignore.go:111` now reads `for _, entry := range block`, so the loop variable no longer reuses the name of the error variable `e`. It was a rename with no change in behaviour, so it needs no regression test. The issue's goal is to remove a block entry only when the old inventory shows this checkout owned it, and the head commit delivers that. `ApplyManaged` passes in the old inventory across all scopes (`ownership.go:174`). Owned paths and block lines are both escaped with `escapeIgnore`, so the two compare correctly. A malformed or unterminated block is still reported as an error through `splitIgnore`. I ran `go test ./cmd/weave/...` at head and it passes. I found nothing new at any severity.

1. **Strengths**
   - `splitIgnore` (`managed_ignore.go:19-56`) is now the only BEGIN/END parser, and both callers use it. The DRY fix is complete.
   - Ownership is decided from real provenance, the old inventory, not by special-casing dependency checkouts. So a dependency that has data mounts is also covered.
   - The regression test requires `.gitignore` to stay byte-identical. It covers both a missing inventory and an inventory with an empty `outputs` list, which together are Done-when clause 1 in both of its states.
   - `TestManagedRetiresOnlyOwnedBlockEntries` covers both sides of the rule: an entry this checkout doesn't own survives, and an owned entry is added and later retired.
2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new.
5. **Test coverage:** Done-when clauses 1 to 4 are each covered by a test: the two-variant regression test, the retire test, `TestManagedIgnoreMigrationAndLocalNegations`, and the passing suite. The known side effect is documented in the Spec: in a checkout with no inventory, an ignore line for an output removed upstream stays behind.
6. **Architecture:**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass. `splitIgnore` and `managedIgnoreText` are pure functions, and `managedIgnore` only adds a file read around them.
   - **ARCH-PURPOSE:** pass. The ownership rule is fixed where it lives, so every caller of `ApplyManaged` gets it.
7. **Plan revisions:** none. The existing `## Revisions` entry already matches the code.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      managed_ignore.go:111 at head reads `for _, entry := range block`; the shadowing is gone (pure rename, no behavior change).
```
