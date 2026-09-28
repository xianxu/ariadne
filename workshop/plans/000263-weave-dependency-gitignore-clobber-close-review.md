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
