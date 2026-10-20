---
id: 000496
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9725f593e7d90716624e9263c81ff7e70dc6d64d' # card fields mirrored from issue-cards; edit via sdlc
---

# golint: reject paired views of one live inventory

## Problem

pair#711 reproduced a fleet messaging failure with two individually valid native inventories: Zellij --short observed a selected session; --no-formatting observed the same selected session plus an unrelated new one. Comparing their name sets refused the healthy sender/recipient. The direct collector called both sequentially; the poller called both concurrently. One validated full reply already contains the needed names and states.

## Spec

Add a narrowly evidenced Go lint catalog rule for production collectors that request mutually exclusive full/short views of the same live native inventory inside one function, initially the concrete Zellij list-sessions --short / --no-formatting call shape from pair#711. Recognize literal argument lists at the actual query call, including calls nested in goroutines, and report both sites. Keep any native-command/view facts in one declared owner/catalog; do not infer snapshot coherence merely from arbitrary variable names or ban all multiple queries. Independent targeted checks and generation/effect fences remain legal. Scope exemptions to explicit reviewed catalog context, not ownership or broad paths. State the static rule's limits rather than claiming it proves arbitrary IO coherence.

## Done when

- The before-#711 sequential and concurrent collectors and equivalent literal formatting are rejected.
- A single complete listing, repeated independent admission fences, and unrelated native commands are accepted; fixture/conformance tests retain their intentional native view coverage.
- The rule lives in pkg/golint's shared catalog with the incident cited and has positive/negative controls.

## Plan

## Log

Filed by pair:7 while implementing pair#711; unclaimed follow-up, no peer source changes. Root behavior is fixed and defended by stateful native-query count/churn tests in pair#711; this issue adds static prevention for the recognizable mechanical class.
