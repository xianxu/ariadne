---
id: 000598
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'e2a393e689aad7856b5bbabe0a2a15e260aa30c6' # card fields mirrored from issue-cards; edit via sdlc
---

# Affected planning: declared non-Go path ownership and audited reader inputs

## Problem

Pair#853 cannot safely narrow reader-certificate changes with the current affected declaration. Pair has ~560 certificates under tests/affected-readers/<Go test path>.json. Its broad tests/** full rule forces a suite for each update. Removing the rule leaves certificates unclassified. Group.go only triggers non-Go groups; it never adds Go packages. Certificates currently contain version, disposition, sha256, with no declared read paths.

## Spec

Provide declared path-to-Go-package ownership in affected.Plan: a non-Go changed path can add one or more exact Go package roots to the existing ownership union, dependency closure and group selection. Reuse the existing planner as authority (ARCH-DRY); no Pair-side duplicate planner or per-certificate test-group forest. Declaration must come from the pinned trusted base, as existing affected policy does. Preserve explicit full overrides, multi-owner union, and conservative full fallback for unclassified or missing package owners.

Support Pair's certificate identity tests/affected-readers/<test path>.json mapping to the package holding that test, including added, edited, removed and renamed certificates. The exact generic schema/representation belongs to this issue's approved design; avoid requiring a manually maintained row per test if a bounded canonical mapping can derive it. Strictly reject traversal or ambiguous mappings.

Certificates must also record their audited declared reads (repository-relative paths or explicit patterns), alongside the current disposition/hash. Define where this consumer format is owned: ariadne owns generic planner/config capability; Pair#853 owns its inventory schema migration/audit and consumer wiring unless the approved design requires an upstream seam. Declared reads must have explicit selection semantics: retain the owner's tests and all packages/groups genuinely needed by those inputs, preserving full fallbacks for policy, selector, shared harness or unclassified inputs. Do not treat reverse import closure as a substitute for black-box executable/fixture reads. local_or_literal_fixture, fixed_scanner and non_go_full_fallback remain auditable classifications; missing or invalid read declarations must never authorize skipping coverage.

Pair#853 also narrows cmd/couch/*_test.go changes, which are currently caught by cmd/couch/**. Keep production command/harness changes conservative; the input-reader audit and fixed source preflight remain in force. A test edit should select the owning command package and audited input implications, without requiring the whole repository merely because it is in the main package.

This is a prerequisite for Pair#853, not stacked work. ops:0 assigned ariadne:5 after ariadne#592 lands. Filed unclaimed as requested; no implementation or estimate implied.

## Done when

- Production affected.Plan plus config/graph integration can map a non-Go path to exact Go package owners; tests cover union, dependencies, groups, full precedence, malformed/unknown owners and no-op unmapped input.
- A real Pair certificate case (request_lifetime_test.go.json from Pair#825) qualifies against the new provider and selects its owning package; arbitrary certificates cannot silently select nothing.
- Certificate read semantics and migration boundary are documented, with invalid/older data conservatively handled and opaque future metadata preserved on rewrites.
- Policy/selector/shared harness changes retain full selection; cmd/couch test narrowing has a qualified consumer path without reducing native fixture coverage.
- Compatibility, trusted-policy provenance, deleted/renamed path behavior and bounded mapping cost are tested. No new per-file subprocess or unbounded group multiplication.

## Plan

- [ ] Design the generic ownership seam and certificate-read consumer contract; coordinate with Pair#853 before code.
- [ ] Implement and qualify the provider; land independently before Pair#853 resumes.

## Log

### 2026-10-24

Filed by pair:1 at ops:0 request. Source anchors: cmd/sdlc/internal/affected/plan.go Plan (full overrides, owners plus groups, then Dependents); config.go Group.Go is a trigger only; graph.go Graph.Owners has directory/testdata/embed ownership only. Pair source anchors: .sdlc/affected.json full tests/** and cmd/couch/**; tests/affected-go-readers-test.py load_inventory; scripts/test-affected.py run_selection fixed source preflight. No source code changed.
