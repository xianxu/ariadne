---
gate: boundary-review
issue: 281
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T17:26:30-07:00"
      dispatched: "2026-10-12T17:15:46-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Displacements done before a failed Apply are never reported; the edit moves to displaced/ silently
          detail: ownership.go renames changed outputs into DisplacedRoot before Apply. On any later error, both compilePrepared callers return without printing the report. The next compile finds the destination absent and writes it with no warning. Print the report, or put the displaced paths in the error, and add a test with an Apply that fails after displacement.
          family: error-path-drops-effect-report
          round: 1
        - id: BR-2
          severity: Important
          title: Data-pass displacement warning gives an owner-relative path without naming the owner layer
          detail: printManagedReport prints filepath.Join(DisplacedRoot, path) relative to root. The data pass runs per owner layer, so a displaced base-layer mount points the operator at the leaf's displaced/ dir. Prefix the path with root when it is not the leaf.
          family: warning-names-ambiguous-location
          round: 1
        - id: BR-3
          severity: Minor
          title: Non-git roots print a "cannot tell whether regenerated outputs are tracked" warning on every compile that rewrites an output
          family: warning-noise
          round: 1
        - id: BR-4
          severity: Minor
          title: 'Displaced copies of a path and its parent (a vs a/b) collide: RemoveAll deletes the sibling copy, or MkdirAll fails mid-displacement'
          family: displaced-path-collision
          round: 1
        - id: BR-5
          severity: Minor
          title: DisplacedRoot is defined in plan rather than next to RootRel/InventoryPath in pkg/weaveownership
          family: reserved-path-single-source
          round: 1
        - id: BR-6
          severity: Minor
          title: Displaced copies of paths weave no longer generates are never swept; only the operator can remove them
          family: artifact-removal-path
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T17:57:01-07:00"
      dispatched: "2026-10-12T17:44:57-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: ApplyManaged's fail() wraps every post-displacement error with the copy paths; TestManagedFailureAfterDisplacementNamesCopy goes red when fail() is unwrapped (verified).
          round: 2
        - id: BR-2
          disposition: addressed
          note: printManagedReport takes leaf and owner and names ancestor-owner paths absolutely; TestPrintManagedReportNamesOwnerAndSkipsNonRepo goes red with leaf-relative names (verified).
          round: 2
        - id: BR-3
          disposition: addressed
          note: Silent outside a repository (mutation-verified). The guard matches git's English stderr, so a non-English locale still warns; consider exit code or LC_ALL=C.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Flat url.PathEscape names under DisplacedRel are injective; TestDisplacedPathIsFlat goes red without the escaping (verified).
          round: 2
        - id: BR-5
          disposition: addressed
          note: DisplacedRel and DisplacedPath now live in pkg/weaveownership beside RootRel and InventoryPath; plan only delegates.
          round: 2
        - id: BR-6
          disposition: addressed
          note: sweepDisplaced removes copies whose outputs left the inventory; TestManagedSweepsDisplacedCopyWithItsOutput goes red when the sweep is a no-op (verified).
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 05c14052d6ea17d6aa20274086ca67dfc0ffea40
      blocked: false
---

# Gate ledger — ariadne#281 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T17:26:30-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `error-path-drops-effect-report` Displacements done before a failed Apply are never reported; the edit moves to displaced/ silently
  ownership.go renames changed outputs into DisplacedRoot before Apply. On any later error, both compilePrepared callers return without printing the report. The next compile finds the destination absent and writes it with no warning. Print the report, or put the displaced paths in the error, and add a test with an Apply that fails after displacement.
- **BR-2** [Important] `warning-names-ambiguous-location` Data-pass displacement warning gives an owner-relative path without naming the owner layer
  printManagedReport prints filepath.Join(DisplacedRoot, path) relative to root. The data pass runs per owner layer, so a displaced base-layer mount points the operator at the leaf's displaced/ dir. Prefix the path with root when it is not the leaf.
- **BR-3** [Minor] `warning-noise` Non-git roots print a "cannot tell whether regenerated outputs are tracked" warning on every compile that rewrites an output
- **BR-4** [Minor] `displaced-path-collision` Displaced copies of a path and its parent (a vs a/b) collide: RemoveAll deletes the sibling copy, or MkdirAll fails mid-displacement
- **BR-5** [Minor] `reserved-path-single-source` DisplacedRoot is defined in plan rather than next to RootRel/InventoryPath in pkg/weaveownership
- **BR-6** [Minor] `artifact-removal-path` Displaced copies of paths weave no longer generates are never swept; only the operator can remove them

## Round 2 — 2026-10-12T17:57:01-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — ApplyManaged's fail() wraps every post-displacement error with the copy paths; TestManagedFailureAfterDisplacementNamesCopy goes red when fail() is unwrapped (verified).
- BR-2 — addressed — printManagedReport takes leaf and owner and names ancestor-owner paths absolutely; TestPrintManagedReportNamesOwnerAndSkipsNonRepo goes red with leaf-relative names (verified).
- BR-3 — addressed — Silent outside a repository (mutation-verified). The guard matches git's English stderr, so a non-English locale still warns; consider exit code or LC_ALL=C.
- BR-4 — addressed — Flat url.PathEscape names under DisplacedRel are injective; TestDisplacedPathIsFlat goes red without the escaping (verified).
- BR-5 — addressed — DisplacedRel and DisplacedPath now live in pkg/weaveownership beside RootRel and InventoryPath; plan only delegates.
- BR-6 — addressed — sweepDisplaced removes copies whose outputs left the inventory; TestManagedSweepsDisplacedCopyWithItsOutput goes red when the sweep is a no-op (verified).

## Open findings

(none — every finding has been disposed)
