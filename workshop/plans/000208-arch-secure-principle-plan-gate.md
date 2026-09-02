---
gate: plan-quality
issue: 208
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-02T15:26:13-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Plan names one hardcoded marker list; four more sites fail or drift on a 6th entry
          detail: |-
            judge_test.go:332 asserts len(markers)==5 via t.Fatalf and hard-fails.
            judge_test.go:120 and :364 hold Contains-only marker lists (:364's literal
            passes only if ARCH-SECURE is appended AFTER ARCH-CONSTRAINTS — placement is
            unstated). Four testdata/golden/*.prompt files carry the registry verbatim
            plus "each of the 5 entries" and byte-drift, requiring a deliberate
            -update-golden past golden_test.go's stop-sign (blessed for registry edits by
            atlas/workflow/architecture-principles.md:16-20). Enumerate the class and
            sweep it this round.
          family: marker-enumeration-sites
          round: 1
        - id: PQ-2
          severity: Important
          title: Deferred guard hardcodes ARCH-AUTHORITY, falsifying the file's own activation claim
          detail: |-
            architecture-deferred.md will say "move a section into architecture.md to
            activate it — that is the whole activation step", but a guard asserting the
            literal ARCH-AUTHORITY also fails on activation. Derive the forbidden set by
            scanning architecture-deferred.md with the existing archMarkerRE
            (architecture.go:11), so the move empties the set. Also assert the deferred
            file parses at least one marker, or a rename or deletion makes the guard
            vacuously pass.
          family: derive-not-restate
          round: 1
        - id: PQ-3
          severity: Important
          title: Done-when claims the AGENTS.md drift guard sees per-marker names; it does not
          detail: |-
            judge_test.go:442 TestArchitecture_NarrativeRoutesToArchPrinciples checks only
            for the routing string and a bare ARCH- mention. Per its comment, issue 128
            deliberately removed per-marker enumeration from the constitution. "Confirm it
            sees the new marker" invites re-adding the restatement 128 deleted; reword to
            state the guard is marker-agnostic by design and needs no change.
          family: unbacked-existing-behavior
          round: 1
        - id: PQ-4
          severity: Minor
          title: No ARCH-CONSTRAINTS envelope line, not even an explicit N/A
          detail: |-
            The registry is embedded verbatim into four gate prompts plus start-plan; a
            sixth entry adds roughly 1.4 KB per dispatch and one more mandatory item in
            the "work through each of the N entries" header. One line stating the accepted
            prompt-growth and attention cost, or marking the category N/A with a basis,
            discharges the lens.
          family: operating-envelope-unstated
          round: 1
        - id: PQ-5
          severity: Minor
          title: Atlas step covers the workflow page but not the index hook that enumerates coverage
          detail: |-
            atlas/index.md:15's hook reads "includes stateful external doubles and explicit
            runtime operating envelopes" — an enumeration that goes stale when a sixth
            entry lands. Add it to the atlas row.
          family: atlas-index-hook-drift
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-02T15:29:17-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Spec states the one-tripwire rule and enumerates all five sites; verified each against the tree.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Guard now derives forbidden set via archMarkerRE plus gate text via promptFS walk; Plan row wording is shorthand for the Done-when.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Done-when now states the guard is marker-agnostic by design and needs no change, matching judge_test.go:442.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Operating-envelope section states cost, basis, and the accepted tradeoff.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: atlas/index.md hook de-enumeration is now both a Plan row and a Done-when bullet.
          round: 2
      blocked: false
content_hash: 3f243da7bd83af58619318a910e0f07428debea00d6bf1429947540f1497aa62
---

# Gate ledger — ariadne#208 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T15:26:13-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `marker-enumeration-sites` Plan names one hardcoded marker list; four more sites fail or drift on a 6th entry
  judge_test.go:332 asserts len(markers)==5 via t.Fatalf and hard-fails.
  judge_test.go:120 and :364 hold Contains-only marker lists (:364's literal
  passes only if ARCH-SECURE is appended AFTER ARCH-CONSTRAINTS — placement is
  unstated). Four testdata/golden/*.prompt files carry the registry verbatim
  plus "each of the 5 entries" and byte-drift, requiring a deliberate
  -update-golden past golden_test.go's stop-sign (blessed for registry edits by
  atlas/workflow/architecture-principles.md:16-20). Enumerate the class and
  sweep it this round.
- **PQ-2** [Important] `derive-not-restate` Deferred guard hardcodes ARCH-AUTHORITY, falsifying the file's own activation claim
  architecture-deferred.md will say "move a section into architecture.md to
  activate it — that is the whole activation step", but a guard asserting the
  literal ARCH-AUTHORITY also fails on activation. Derive the forbidden set by
  scanning architecture-deferred.md with the existing archMarkerRE
  (architecture.go:11), so the move empties the set. Also assert the deferred
  file parses at least one marker, or a rename or deletion makes the guard
  vacuously pass.
- **PQ-3** [Important] `unbacked-existing-behavior` Done-when claims the AGENTS.md drift guard sees per-marker names; it does not
  judge_test.go:442 TestArchitecture_NarrativeRoutesToArchPrinciples checks only
  for the routing string and a bare ARCH- mention. Per its comment, issue 128
  deliberately removed per-marker enumeration from the constitution. "Confirm it
  sees the new marker" invites re-adding the restatement 128 deleted; reword to
  state the guard is marker-agnostic by design and needs no change.
- **PQ-4** [Minor] `operating-envelope-unstated` No ARCH-CONSTRAINTS envelope line, not even an explicit N/A
  The registry is embedded verbatim into four gate prompts plus start-plan; a
  sixth entry adds roughly 1.4 KB per dispatch and one more mandatory item in
  the "work through each of the N entries" header. One line stating the accepted
  prompt-growth and attention cost, or marking the category N/A with a basis,
  discharges the lens.
- **PQ-5** [Minor] `atlas-index-hook-drift` Atlas step covers the workflow page but not the index hook that enumerates coverage
  atlas/index.md:15's hook reads "includes stateful external doubles and explicit
  runtime operating envelopes" — an enumeration that goes stale when a sixth
  entry lands. Add it to the atlas row.

## Round 2 — 2026-09-02T15:29:17-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Spec states the one-tripwire rule and enumerates all five sites; verified each against the tree.
- PQ-2 — addressed — Guard now derives forbidden set via archMarkerRE plus gate text via promptFS walk; Plan row wording is shorthand for the Done-when.
- PQ-3 — addressed — Done-when now states the guard is marker-agnostic by design and needs no change, matching judge_test.go:442.
- PQ-4 — addressed — Operating-envelope section states cost, basis, and the accepted tradeoff.
- PQ-5 — addressed — atlas/index.md hook de-enumeration is now both a Plan row and a Done-when bullet.

## Open findings

(none — every finding has been disposed)
