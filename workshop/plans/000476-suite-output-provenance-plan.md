---
collaboration-mode: discussion
---

# Suite output provenance implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Attribute the pair#712/#689 failure to the outer warm-reattached leaf while keeping quoted helper output inert.

**Architecture:** Preserve column-zero control-record provenance in the existing text scanners. Admit indented test outcomes only through observed four-space ancestors and complete parent names. Keep stored formats, registry policy and the real retry seam unchanged.

**Tech Stack:** Go strings/iterators, existing local-check fixtures, Go AST lint catalog.

This records the correction approved by ops:0 in receipt
09f8538f-08a4-4e8f-aebe-0786f25ec39e. It is a durable checkpoint of that approved
design after the correction took the branch beyond the quick code-size shell,
not a new design or a retrospective claim of plan-quality review.

## Core concepts

| Name | Kind | Lives in | Status |
| --- | --- | --- | --- |
| suiteControlLine | PURE | cmd/sdlc/suitefailures.go | new |
| suiteIndentLevel | PURE | cmd/sdlc/suitefailures.go | new |
| failingNameLines | PURE | cmd/sdlc/suitefailures.go | new |
| parseSuiteLogFacts | PURE | cmd/sdlc/teststats_parse.go | modified |
| unattributedFailures | PURE | cmd/sdlc/teststats_parse.go | modified |
| suiteLineProvenance | PURE | pkg/golint/rules.go | new |

`failingTests` and `failingSubtests` share one scanner. Its transient ancestor
stack belongs to one scan and dies on return. Facts and unknown-failure reasons
use the same raw-line control predicate. No second format, persisted field or
external dependency is introduced (ARCH-DRY, ARCH-PURE, ARCH-EVOLVE).

| Name | Kind | Lives in | Status | Wraps |
| --- | --- | --- | --- | --- |
| TestKnownFlakeGateQuotedNestedOutput | INTEGRATION | cmd/sdlc/suitefailures_nested_test.go | new | existing runLocalCheck |
| TestKnownFlakeGateCannotRegisterAQuotedLeaf | INTEGRATION | cmd/sdlc/suitefailures_nested_test.go | new | existing runLocalCheck |

The integration tests reuse localRunRepo's private Git fixture, fake suite-lock
helper and stateful ciFakeGH. The production IO seam is unchanged. They assert
the recorded outcome/status and selected retry, rather than mocking the
attribution decision. No live sandbox, shared checkout or real failed suite is
needed to reproduce this bug.

## Admission and transitions

One scan owns pending names, the current summary package and an ancestor slice.
Column-zero raw test outcomes establish a root; package trailers flush pending
names and clear ancestors. An authoritative shard header flushes the previous
block, names its package/test and seeds its top-level context. Header names are
evidence even when the producer omitted raw ancestors.

For an indented FAIL record, measure spaces; do not count slashes:

```go
level := suiteIndentLevel(line)
if level <= 0 || level > len(ancestors) ||
    !strings.HasPrefix(test, ancestors[level-1]+"/") {
    // ignore: no real parent at the preceding observed level
} else {
    ancestors = append(ancestors[:level], test)
    // admit this descendant, truncating any previous sibling branch
}
```

Only complete FAIL records reach that decision. Misaligned or tab indentation,
skipped levels, quoted top-level markers, package trailers and JSON cannot
change the stack. A literal slash inside a t.Run name can still occupy one
observed level. Rejected records leave the stack unchanged.

The smallest alternative, dropping all indentation, loses real #466 leaves.
The first prefix-only exception was rejected by BR-1: a helper with an
overlapping root could invent a leaf and even explain an unrelated outer
failure. Structural admission replaces that exception at its one owner.

## Visible states and operating envelope

Actions and the three existing visible outcomes remain those in the issue:
registered retry, repair a named failure, or inspect/repair/rerun an
unattributable failure. Raw logs remain intact; this adds no authority state,
record family, artifact lifecycle or retry policy.

Input is untrusted captured output (ARCH-SECURE). The existing complete-log
and line caps still apply. Each failure-name pass visits each line once,
inspects bytes linearly, and does one indexed ancestor lookup plus slice
truncate/append per admitted record. Distinct names retain existing sorting
cost. No ancestor searches or per-record log rescans are added.
TestFailingNamesLineVisits counts 4N+1 visits for N=8 and 16, with real children
and quoted helpers, rejecting repeated traversals. All parser memory is bounded
by input bytes/names and dies with the scan (ARCH-CONSTRAINTS, ARCH-FUNERAL).

## Implementation and verification

- [x] Preserve the issue excerpt in cmd/sdlc/testdata/nested-go-test-476.txt;
  synthetic helper-child output extends it to exercise ignored descendants.
  TestSuiteQuotedNestedOutput checks plain, CRLF and sharded contexts.
- [x] Observe the original fixture and real retry fail red; correct raw control
  records in suitefailures.go and both teststats_parse.go loops.
- [x] Observe TestSuiteQuotedOverlappingNames, TestSuiteIndentedFailureAncestry
  and TestKnownFlakeGateCannotRegisterAQuotedLeaf fail before BR-1's correction.
  Apply the observed-ancestor rule across the shared scanner.
- [x] Preserve real siblings, grandchildren and literal-slash names; reject
  shared-root helper quotations, skipped depths and invalid indentation. Keep
  genuine build, timeout, package and unknown failures blocking.
- [x] Add suite-line-provenance to pkg/golint/rules.go with red/green catalog
  fixtures. Its named-scanner lexical tripwire does not claim control-flow or
  alias analysis; behavioral tests own admission correctness.
- [x] Run quoted, ancestry, real gate, parser, saved-evidence and environment
  checks: go test ./cmd/sdlc with the issue's focused regex (12.864s, pass).
  Run go test ./pkg/golint ./cmd/sdlc/internal/flakes (both pass).
- [x] Update README, atlas/workflow/sdlc-binary.md and workshop/lessons.md.

Workflow evidence stays in the issue Log: push corrected code, run its recorded
full suite, re-run close to dispose BR-1, then let the normal merge own final
integration and CI. No passing old-head suite exists: its queued attempt was
cancelled before admission. A review correction requires fresh verification.
