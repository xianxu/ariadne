---
id: 000276
status: open
deps: [pair#353]
github_issue:
created: 2026-09-30
updated: 2026-09-30
estimate_hours:
card_mirror: '29811fedd6eadbb05d2af7a0eb7237da73fca44c' # card fields mirrored from issue-cards; edit via sdlc
---

# Add Couch skill prelude

## Problem

pair#353 ships the Couch peer-messaging skill in `couch --skill`, but agents
in Ariadne-managed repositories currently need an explicit reminder to load it.
The operator requested the simple Ariadne prelude as the remaining integration.

## Spec

Add a thin skill prelude through Ariadne's existing local/binary-skill pattern.
It tells an agent coordinating with live Couch slots to run `couch --skill`
and follow the returned instructions. Couch remains the source of truth for
slot discovery, messaging, receipts, safe delivery and loop-control conventions;
do not duplicate that protocol in Ariadne. Propagate the prelude through the
normal managed skill distribution so fresh agents can discover it without a
manual operator prompt. Loading the skill must not initiate messages or work.

## Done when

- The managed Couch skill prelude delegates to `couch --skill`.
- A freshly started agent in a downstream Ariadne-managed repository discovers
  the prelude and can load the binary-provided skill.
- Normal weave/distribution checks pass; the prelude contains no copied runtime
  protocol and handles an unavailable Couch binary with a clear explanation.

## Plan

- [ ] Add the prelude using the existing binary-skill pattern and distribution.
- [ ] Verify downstream discovery and loading; document the setup if needed.

## Log

### 2026-09-30

Filed during pair#353 close after operator acceptance of direct send/reply and
family dispatch smoke. This is the remaining skill-discovery integration, not
another messaging implementation. Depends on pair#353's binary skill provider.
