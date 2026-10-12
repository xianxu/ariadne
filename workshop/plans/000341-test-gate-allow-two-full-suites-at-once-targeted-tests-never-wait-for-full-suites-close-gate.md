---
gate: boundary-review
issue: 341
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T16:27:24-07:00"
      dispatched: "2026-10-12T16:26:04-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: test-lock.test.sh "waiter names the holder" hardcodes the repo name ariadne
          detail: scripts/test-lock.test.sh:23 greps for "  - ariadne (", but the repo field is the checkout folder's name, so the case fails in any checkout not named ariadne (it fails in this review checkout, 0e643ad2f44af2b3). Match on "make pid $a" only, as the new two-slot block does, or use "$(basename "$(git rev-parse --show-toplevel)")". This is the only such assertion in the window.
          family: test-env-coupling
          round: 1
        - id: BR-2
          severity: Minor
          title: Atlas Slots bullet still says "then", but the step before it was removed
          detail: atlas/workflow/sdlc-binary.md:1658 "it then takes one of WF_TEST_GATE_SLOTS slots" followed the removed "wait for a full suite" step; drop "then". Makefile.workflow:9 also has an unwrapped, overlong comment line.
          family: doc-claim-drift
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: 8f9acde073791c7fc2544c1c8e3aa4b09f1434f8
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T16:30:00-07:00"
      dispatched: "2026-10-12T16:29:16-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: test-lock.test.sh:23 now takes the repo name from git rev-parse --show-toplevel, as test-lock.py does; all 23 cases pass in checkout 087fa04e4cd2d4ce, and no other ariadne hardcode is left in the file.
          round: 2
        - id: BR-2
          disposition: addressed
          note: atlas sdlc-binary.md:1658 drops "then"; Makefile.workflow:8-15 comment rewrapped to the width of the lines around it.
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 822279b2bf80deff628defb8165529060ad13de1
      blocked: false
---

# Gate ledger — ariadne#341 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T16:27:24-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `test-env-coupling` test-lock.test.sh "waiter names the holder" hardcodes the repo name ariadne
  scripts/test-lock.test.sh:23 greps for "  - ariadne (", but the repo field is the checkout folder's name, so the case fails in any checkout not named ariadne (it fails in this review checkout, 0e643ad2f44af2b3). Match on "make pid $a" only, as the new two-slot block does, or use "$(basename "$(git rev-parse --show-toplevel)")". This is the only such assertion in the window.
- **BR-2** [Minor] `doc-claim-drift` Atlas Slots bullet still says "then", but the step before it was removed
  atlas/workflow/sdlc-binary.md:1658 "it then takes one of WF_TEST_GATE_SLOTS slots" followed the removed "wait for a full suite" step; drop "then". Makefile.workflow:9 also has an unwrapped, overlong comment line.

## Round 2 — 2026-10-12T16:30:00-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — test-lock.test.sh:23 now takes the repo name from git rev-parse --show-toplevel, as test-lock.py does; all 23 cases pass in checkout 087fa04e4cd2d4ce, and no other ariadne hardcode is left in the file.
- BR-2 — addressed — atlas sdlc-binary.md:1658 drops "then"; Makefile.workflow:8-15 comment rewrapped to the width of the lines around it.

## Open findings

(none — every finding has been disposed)
