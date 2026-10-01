---
id: 000276
status: working
deps: [pair#353]
github_issue:
created: 2026-09-30
updated: 2026-09-30
estimate_hours:
card_mirror: '8a6b659d6b7a79a0686595a714c4bf0c601e6afb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-30T22:11:29-07:00
flow: {kind: quick, provenance: inferred, spec: "121d93ce", done: "e14868a9"}
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

Design: a static pointer skill at `construct/local/couch/SKILL.md`, following
the `construct/local/sdlc` (`xx-sdlc`) precedent. weave lowers every
`construct/local/<name>` to `.claude/skills/xx-<name>` and
`.agents/skills/xx-<name>`. Derivatives inherit it through the weave layer
walk, so no manifest change is needed.

A dynamic (weave-generated) skill was rejected for two reasons. Its body would
be a copy of `couch --skill`, which the spec forbids. Its generator would also
fail `weave compile` on any machine without Couch, while the prelude must
degrade with an explanation instead.

- [x] Add `construct/local/couch/SKILL.md`. Its trigger description covers
      coordinating with live slots and interpreting an incoming Couch message.
      Its body says to run `couch --skill` and follow the output; to explain
      and stop if `couch` is missing; and that loading sends nothing.
- [x] Run `weave compile` in ariadne. Generated: `xx-couch` links, `.gitignore`
      entries, and the process-manual entries.
- [x] Verify downstream with a throwaway derivative linked to this checkout.
      `weave compile` must produce `.claude/skills/xx-couch`, and its
      `SKILL.md` must resolve.
- [x] Atlas: add the skill to the skill list, if one exists beyond the
      generated process manual.

## Log

### 2026-09-30

Filed during pair#353 close after operator acceptance of direct send/reply and
family dispatch smoke. This is the remaining skill-discovery integration, not
another messaging implementation. Depends on pair#353's binary skill provider.

- Claimed and started planning in :1. The binary-skill precedent is
  `construct/local/sdlc`, a static pointer to `sdlc --help`. Couch at
  `/Users/xianxu/workspace/pair/bin/couch` provides `couch --skill`
  (`name: couch`).
- Added `construct/local/couch/SKILL.md`. `weave compile` had to run outside
  the sandbox, which denies writes to `.claude/settings.json` and
  `.claude/skills`. It lowered `.claude/skills/xx-couch` and
  `.agents/skills/xx-couch`, and added the `.gitignore` entries. This live
  session listed `xx-couch` as soon as the link appeared.
- Regenerated `atlas/workflow/process-manual.md` with `sdlc process-manual`. It
  now lists `xx-couch`, and the regeneration also picked up earlier help-text
  changes that had never been regenerated.
- Downstream check: a throwaway derivative in the scratchpad ran
  `weave link <this checkout>` then `weave compile`. Both `xx-couch` links
  resolve to `construct/local/couch`, and `SKILL.md` reads through them.
- `go test ./cmd/weave/... ./pkg/...` passes, and `make weave-drift-check`
  passes.
- No hand-maintained skill list exists outside the generated manual. Added an
  `atlas/index.md` entry.

