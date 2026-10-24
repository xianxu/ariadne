---
id: 000590
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'd0973cbc6b730d6361f2cb8a8b53af7b1b170455' # card fields mirrored from issue-cards; edit via sdlc
---

# Qualify recursive JSON lint against independent oracles and DOM materializers

## Problem

Ariadne pkg/golint json-recursive-scanner reports proven false positives in Pair while adopting 5a17f789619d for pair#838. The matcher treats recursive Token + More + any string-key map as duplicate-validator authority. It flags both an independent differential-test oracle and a standalone DOM materializer. Its advice names cmd/sdlc/internal/jsonstrict, which downstream Go modules cannot import.

## Spec

Qualify/narrow the matcher using the two counterexamples without weakening actual duplicate-validator detection or adding broad filename/package exemptions. Clarify downstream-accessible remediation. Preserve the independent oracle and standalone sample boundaries; consumers must not delete independent checks or add unwanted production dependencies to satisfy a rule.

## Done when

- Fixtures reproducing both false positives are accepted, while current true duplicate-key scanner fixtures remain rejected.
- Diagnostics recommend an accessible shared authority or accurately describe ownership instead of an internal Go import inaccessible downstream.
- Counted/existing matcher tests and required checks pass; landed revision sent to pair:3 so exactly two temporary per-site allows in pair#861 can be removed.

## Plan

- [ ]

## Log

Pair#861 qualification found51 new catalog findings. Three actual duplicate validators will be consolidated; these two are different roles:

1. cmd/internal/strictjson/decode_equivalence_test.go legacyScanValue (line46), used exclusively by legacyRejectDuplicateKeys and FuzzRejectDuplicateKeysEquivalent. It is the independent legacy Token oracle for an optimized byte scanner; replacing it with the implementation under test makes the differential test tautological. Semantics include exact key equality, invalid Unicode replacement and float overflow.
2. examples/mock-to-real/internal/jsonvalue/value.go value(d *json.Decoder, depth int) (any,error) (line28), a generic recursive DOM materializer returning map/array/scalar values with json.Number, exact keys and depth64. Nested module example.com/mock-to-real deliberately has no production Pair dependency (see catalogue/catalogue.go concluding boundary note).

Pair retains the complete observed lint log in its #838 evidence and filed adoption #861. TL explicitly authorized reasoned per-site lint:allow only at these two declarations, naming this upstream issue, pending its correction. These are temporary proven-false-positive annotations, not blanket waivers. Please notify pair:3 on landing.
