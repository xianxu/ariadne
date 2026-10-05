---
id: 000292
status: open
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: 'a820cacfbb1b8d45c7de7ca218872a2813d9c1ef' # card fields mirrored from issue-cards; edit via sdlc
---

# Rename doc-review's fresh-context review to fact review; xx-fix runs one per review session

## Problem

The operator drives prose reviews from the pair review pane with Alt+Return.
That always reaches the co-authoring agent, which carries the whole
conversation, so claims in the doc only get checked by the same context that
helped write them. `doc-review` (the separate, read-only, second-vendor
fact-and-reference check) runs only if the operator remembers to ask for a
"fresh context review" in chat. It isn't part of the normal review loop, and
the name "fresh context review" says how it runs, not what it's for.

Wanted: Alt+Return should trigger the appropriate review, and every review
session should include at least one fact-check pass that tightens the doc's
claims before it ships.

## Spec

- **Name it "fact review".** In `doc-review`'s help, report header
  (`# Fact Review of …`), `review_kind` frontmatter, and in the xx-fix skill.
  Keep "fresh context review" / "fresh review" / "second-agent review" as
  accepted trigger phrases so existing habits still work.
- **xx-fix incorporates it.** In a docflow review session, the agent runs at
  least one fact review (`doc-review <file>`) before ship, without waiting to
  be asked: e.g. on the first Alt+Return round once the draft's claims have
  settled, or when a round adds new factual claims, cited figures, or links.
  It triages the report and proposes accepted corrections as normal handoff
  records; rejected findings get a one-line reason in the round body.
- **Ship guard (soft).** If no fact review ran on the branch, the agent says
  so when asked to ship and offers to run one first; the operator can decline.
- Docs-only prose with no factual claims can skip it; the agent says why.

## Done when

- `doc-review --help` and its report call it a fact review; old trigger phrases
  still work.
- The xx-fix skill (construct/local/fix/SKILL.md) defines when a fact review
  runs within a review session, how findings flow into handoff records, and the
  pre-ship check.
- A review session that ships records at least one fact review in its round
  bodies (or an explicit skip reason).

## Plan

- [ ]

## Log

### 2026-10-05
