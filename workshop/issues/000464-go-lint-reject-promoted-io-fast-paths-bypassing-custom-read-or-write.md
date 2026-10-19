---
id: 000464
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7560c220bd400d746e7aea350040b9bfc555ef58' # card fields mirrored from issue-cards; edit via sdlc
---

# Go lint: reject promoted IO fast paths bypassing custom Read or Write

## Problem

pair#700's stateful exec fixture exposed a bounded subprocess capture that embedded bytes.Buffer and defined Write. io.Copy selected the promoted ReadFrom method, bypassing the byte limit. An actual oversized fake pair response exceeded the 4 MiB cap. A named buffer field removed the promoted interface; a separate overflow bit refuses even a valid JSON prefix followed by oversized whitespace. Motivation: cmd/internal/couchcmd/archive_schedule.go and TestRunArchivePassRefusesInvalidProtocol in pair#700.

## Spec

Add a pkg/golint rule for a Go type that embeds bytes.Buffer or another io.ReaderFrom/io.WriterTo provider while defining its own Write/Read but not overriding the matching ReadFrom/WriteTo. Detect promoted fast paths that bypass wrapper enforcement. Use type/interface information where available; do not use a broad path allowlist. Diagnose how to preserve the wrapper's semantics: named fields or an explicit safe override.

## Done when

- A bounded writer embedding bytes.Buffer with its own Write but no ReadFrom is rejected; the symmetric Read/WriteTo case is rejected.
- Other embedded ReaderFrom/WriterTo providers are covered where the analyzer can resolve their method sets.
- Safe named-field composition, explicit fast-path overrides and ordinary buffer embedding without custom Read/Write are negative controls.
- A fixture reproduces io.Copy bypassing custom Write/Read, with the motivating pair#700 regression linked.
- The normal lint suite enforces the rule without unrelated ownership exemptions.

## Plan

- [ ]

## Log

Filed unclaimed by pair:7 on the hub's explicit request, receipt 9c50ae1e-4080-482e-ab5a-f930d7607101. No ariadne implementation changes made.
