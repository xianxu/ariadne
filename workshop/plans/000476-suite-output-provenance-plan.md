---
collaboration-mode: discussion
---

# Suite output provenance implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Attribute the pair#712/#689 failure to its proven outer test while keeping quoted helper output inert. Exact raw subtest precision is superseded by the current approved goal below.

**Architecture:** Admit column-zero records, retain exact JSON/shard leaves, and coarsen ambiguous raw indented output to a proven ancestor. Keep the complete-prefix classifier, stored formats, registry policy and retry seam unchanged. The original design and revisions below retain the decision history.

**Tech Stack:** Go strings/iterators, existing local-check fixtures, Go AST lint catalog.

This records the correction approved by ops:0 in receipt
09f8538f-08a4-4e8f-aebe-0786f25ec39e. It is a durable checkpoint of that approved
design after the correction took the branch beyond the quick code-size shell,
not a new design or a retrospective claim of plan-quality review.

## Original approved design (superseded)

The following original concepts, transitions and checked implementation record
are historical. The final Revision, **Current approved goal: proven granularity**,
is the current contract; it explicitly removes the ancestry mechanism.

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

## Revisions

### 2026-10-19 — Third-delta re-plan, awaiting hub approval

Round 2 keeps BR-1 open: measuring initial ASCII spaces and then using Fields
still admits a tab or Unicode whitespace after those spaces. Stop before a
third production change; the second queued suite was cancelled before admission.

The root is two different definitions of the same prefix: the indent helper
stops at the first non-ASCII-space byte, while Fields consumes Unicode
whitespace. Replace that split definition with one complete-prefix classifier,
owned by suiteIndentLevel. This supersedes the initial-space measurement above.

Revised rule:

1. Scan leading runes with the same unicode.IsSpace category as Fields.
2. Any whitespace rune other than ASCII space makes the prefix invalid.
3. At the first non-whitespace rune, accept only a multiple of four spaces;
   return the observed level. Empty/all-whitespace lines yield no record.
4. suiteControlLine derives column-zero admission from that same result.
   failingNameLines obtains the level once, rejects invalid prefixes before
   Fields, then uses the already approved ancestor transition unchanged.

```go
spaces := 0
for _, r := range line {
    if !unicode.IsSpace(r) {
        if spaces%4 == 0 { return spaces / 4 }
        return -1
    }
    if r != ' ' { return -1 }
    spaces++
}
return -1
```

This is one lexical definition, not a new parser, state machine, log format,
fallback or post-parse sanitization layer. The stack, explicit shard names,
facts consumers, visible outcomes and recoveries remain unchanged. One rune
inspection per consumed prefix rune is O(prefix bytes); existing bounded line
scans and indexed ancestor operations retain the N/2N visit guard.

The simplest alternative, dropping every indented record, loses the actual
warm-reattached leaf and #466 behavior. Adding tab/NBSP exceptions treats only
two members of a Unicode whitespace class and repeats the same design mistake.
Changing producers cannot repair historical logs. The single classifier covers
the complete class at its existing owner.

- [x] Reproduce mixed-prefix leaf admission red in plain, CRLF and sharded
  contexts, with and without a real child.
- [x] Enumerate every rune for which unicode.IsSpace is true: only ASCII space
  may appear in the prefix. Exercise non-space runes before and after four
  ASCII spaces, including mixed sequences and invalid widths.
- [x] Add the real gate refusal when only a mixed-indentation quoted leaf is
  registered; preserve the positive warm-reattached retry and the original
  pair#712/#689 excerpt fixture.
- [x] Implement the single prefix definition after approval. Extend the
  lexical lint tripwire to recognize the shared level classifier before Fields.
- [x] Run existing and expanded parser/gate/lint controls.

The new pushed-head recorded suite, full close review and landing are pending
workflow gates, tracked in the issue Log rather than pre-ticked implementation
steps. If yet another mechanism is needed, stop and return to the hub before
implementation.

### 2026-10-19 — Approved re-plan verification

Hub approval: 6844e546-4bc3-4e66-b635-e03181de0edb, canonical receipt verified.
The Unicode table and real gate tests failed red before the complete-prefix
correction; expanded attribution/gate/environment/saved-evidence checks passed
after it (40.830s), as did golint (1.309s) and flakes (0.070s). LF is the log's
record delimiter, so its invalid in-line prefix is tested directly against the
classifier; every other non-space whitespace rune is rejected through plain,
CRLF and sharded parsing. The original excerpt verifies normal LF framing and
the actual warm-reattached leaf across those modes. No full-suite pass claimed.

### 2026-10-19 — Re-plan context lifetime after round 3, pending approval

Round 3 reproduced real Go output in which a parent logs after a failed child.
The parent's log returns to level one, but the scanner retains the child at
level one; its quoted continuation at level two then becomes a fake grandchild.
The prefix grammar is correct; the missing model is ancestry lifetime across
ordinary output. Stop before another production change. The third queued suite
was cancelled before admission.

Replace the earlier statement that rejected records leave ancestry unchanged
with a complete line-event model. The immutable input has one owner (this scan),
but an ancestor's authority lasts only until the stream returns to its parent's
output. A lock cannot recover that historical ownership: there is no concurrent
live resource to hold; the recorded ordering is the evidence (ARCH-ORDER).

One pure line classifier owns framing and prefix provenance. Strip only the
record's trailing LF/CRLF framing, then classify the complete prefix as:
invalid; blank at a validated ASCII-space level; or content at that level.
Use a tagged result, not independent validity/blank flags. All existing
column-zero checks derive from this same result before tokenization. Preserve
the approved whole-Unicode-prefix rule and ancestor-name checks.

| Line event | Ancestry effect | Name evidence |
| --- | --- | --- |
| Authoritative top-level FAIL | Reset to this root | Keep pending top name |
| Authoritative shard header | Flush previous block; seed declared context | Keep explicit header name |
| Admissible indented FAIL at L | Truncate to parent at L-1, then append child | Keep this real descendant |
| Ordinary output at L>0 | Retire nodes at depth L and deeper: retain at most L ancestor entries | None added |
| Rejected indented FAIL candidate | Treat as output at its proven level; never attach a child | None added |
| Blank native output at a proven level | Same context retirement as output | None added |
| Unrelated column-zero output | Retire transient ancestry | Preserve already collected names |
| Invalid prefix / unknown owner | Retire transient ancestry conservatively | Preserve already collected names; add no leaf |
| Package trailer / end of input | Flush pending names; clear context | Existing attribution/unknown rules |

The line classifier is PURE in suitefailures.go. Its tagged prefix result is
transient and consumed by the existing scanner and suite-facts readers; this
does not add a log format, persisted field or runtime artifact. The pending
name/package mechanism, registry and retry policy stay unchanged. The three
visible outcomes remain exact registered retry, named failure to repair, or
unattributable failure to inspect/repair/rerun. Conservative retirement can
withhold a leaf on malformed input, never manufacture one.

The simplest alternative is to drop all raw subtest records, which loses the
actual warm-reattached leaf. Keeping context through ordinary output is the
reproduced defect. Requiring JSON producers would avoid text ambiguity but
cannot repair retained plain historical logs and requires peer changes.

Before implementation, reproduce and retain real Go output plus its small
source: parent logs before/after a failed child, child returns to grandparent,
real grandchild and sibling transitions, native blank log continuations, and
both verbose and buffered output. The committed fixtures must exercise plain,
CRLF, sharded headers and historical-log consumers. Assert that a registration
covering only the invented helper cannot explain the actual child; the real
child registration and original pair#712/#689 warm leaf still work. Preserve
the complete Unicode table and N/2N line-visit counts. Per-line transitions use
indexed lookup and slice truncation, with no look-back or ancestor rescan.

The four proposed context-model tasks are superseded by the hub's revised goal
below; none of this pending tagged-classifier model is required or implemented.

### 2026-10-19 — Current approved goal: proven granularity

Approval: c4637f8b-c20e-4bb2-99b6-9077742773b6 (canonical verified). This section
supersedes the earlier requirement for exact raw subtests and both ancestry
proposals. The incident is false unattributability; exact precision is optional.

Keep suiteIndentLevel and its Unicode/ASCII-space grammar. Remove ancestry state
from failingNameLines; only column-zero records can add names or flush packages.
Go JSON failure events and explicit shard headers remain exact evidence. Raw
indented subtests under failing ancestors cannot prove a leaf, so retain the
nearest proven ancestor. No new state machine or typed line protocol is needed.

Add one bounded coarsening note through SuiteLogFacts.Warnings, already printed
by runLocalCheck and legacy import: ambiguous indented subtest output is ignored,
matching uses proven records, and subtest-only registrations may miss a coarser
ancestor. Keep flakes.Covers, Leaves, registry and persisted formats unchanged.
Real build/timeout/missing-identity guards still block their own failures.

Visible states: exact structured leaf; coarse proven ancestor with a precision
note; genuinely unidentified failure. Existing recoveries remain. Coarsening
can withhold an automatic subtest-only retry, which the hub explicitly accepts;
it cannot invent a leaf to obtain one. Raw evidence remains available.

- [x] Capture and commit real buffered/verbose Go output plus its small source
  for parent logging before/after failed children, nested returns and siblings.
  Keep the original pair#712/#689 excerpt too.
- [x] Change current fixtures to the revised contract and observe them red:
  raw cases coarsen, structured cases remain exact, actual gate parent-only
  registration reruns and leaf-only registration stays unknown with a note.
- [x] Remove the scanner's ancestor mechanism; keep the shared prefix guard.
  Add the coarsening note in the existing facts scan without adding a log pass.
- [x] Verify saved-log success guard and historical replay: coarse raw evidence
  cannot justify leaf-only registrations; structured leaf evidence still can.
  Use Go JSON in the existing real-Go #466 integration fixture so it retains
  exact provenance, and add raw-mode coarse behavior separately.
- [x] Verify full Unicode, N/2N visits, parser, actual gate and lint controls.

The recorded suite, full close (disposing BR-1 under revised acceptance) and
normal merge remain workflow gates after these implementation steps. Existing
cost bounds hold: fixed linear scans, one prefix classification per record,
map/name collection and sorting. No live sandbox, peer workflow edit or schema
change is involved.


### 2026-10-19 — Implementation evidence for the current goal

Removed ancestry tracking. Actual Go 1.27.1 buffered and verbose fixtures now
coarsen safely, including parent logging before/after a failed child and nested
return/sibling ordering. The new raw/structured/replay/gate cases failed on the
ancestry implementation, then passed after simplification. The exact JSON gate
also exposed an existing ordering bug: unattributed package derivation preceded
JSON name collection. A focused package-fail fixture reproduced it red; deriving
those reasons after completed names fixes the same facts authority, without
another parser or state. The focused parser, real-gate, saved-pass, Unicode,
linear-visit and environment selection passed (33.889s); golint and flakes passed
(0.604s and 0.056s). Full recorded suite and close remain separate pending gates.
