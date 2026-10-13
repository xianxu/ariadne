---
gate: boundary-review
issue: 299
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T16:57:53-07:00"
      dispatched: "2026-10-12T16:56:35-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: dev-prepare detection regex accepts only `dev-prepare:` at column 0
          detail: construct/dev-aliases.sh:225 misses rules that name several targets (`foo dev-prepare:`) and a space before the colon, falsely matches `dev-prepare:=x`, and ignores Makefile.workflow. Document the exact accepted declaration shape in the header comment and in atlas base-layer.md.
          family: owner-declaration-detection-shape
          round: 1
        - id: BR-2
          severity: Minor
          title: atlas base-layer.md:97 joins new prose onto an existing line, making it overlong
          detail: '"...keep the plain form. The function only **builds + runs** — it" should be re-wrapped to match the surrounding line width.'
          family: doc-line-wrap
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 4511e8d7de0822369904ef96c03ff258023485e4
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T17:00:14-07:00"
      dispatched: "2026-10-12T16:59:11-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Regex now accepts multi-target and space-before-colon rules and rejects := (fixtures at dev-aliases.test.sh:79-84). Removing the ([^=]|$) guard turns the := test red. Accepted shape documented in the comment above owner_prepares (construct/dev-aliases.sh:222-226); atlas says "a dev-prepare rule", which is accurate.
          round: 2
        - id: BR-2
          disposition: addressed
          note: atlas/workflow/base-layer.md:92-101 re-wrapped, lines 75-79 chars, consistent with the paragraph.
          round: 2
      findings:
        - id: BR-3
          severity: Minor
          title: Text-regex detection of dev-prepare still false-matches recipe lines, `::=`, target-specific vars
          detail: '2nd finding in this family. The rule: inferring make declarations by regex over Makefile text leaves gaps however it is patched. Verified matches at construct/dev-aliases.sh:228 for a tab-indented recipe line "@echo run dev-prepare: now", for "dev-prepare ::= x", and for "dev-prepare: V = 1"; each emits `make -s dev-prepare`, which then fails every call. Class fix: query make''s own database (make -C owner -pRrq :) for a dev-prepare target with a rule, which also covers Makefile.workflow and other includes. Cheaper fallback: exclude tab-led lines and document ::= and target-specific vars as unsupported.'
          family: owner-declaration-detection-shape
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 1b95ae894abcc0a48d98c1d677ab4c010d01bdd6
      blocked: false
    - "n": 3
      timestamp: "2026-10-12T17:05:09-07:00"
      dispatched: "2026-10-12T17:03:53-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: owner_prepares now asks make (make -q dev-prepare) behind a grep prefilter; restoring 1b95ae89's version makes 3 new cases fail (target-specific var, recipe line, Makefile.workflow include); head passes 38/38.
          round: 3
      findings:
        - id: BR-4
          severity: Minor
          title: Prose about how dev-prepare is detected still describes the old regex
          detail: 'Instances: (a) the construct/dev-aliases.sh:15 header says "Makefile or Makefile.local" and omits Makefile.workflow, which line 222 and the atlas name; (b) the issue Plan bullet 1 says owner_prepares "greps ^dev-prepare:"; (c) Plan bullet 2 claims a test that a .PHONY mention alone is not a declaration, but none exists and make -q treats a bare .PHONY: dev-prepare as a rule (rc 0, harmless no-op). Rule: describe detection once, in the owner_prepares comment, and point other prose at it; add a Plan revision.'
          family: doc-restates-detection-rule
          round: 3
        - id: BR-5
          severity: Minor
          title: make -q runs owner Makefile parse-time side effects at shell startup, once per binary
          detail: 'Parsing the owner''s Makefiles runs $(shell ...) calls (verified: a touch fired) and can rebuild included makefiles, every time a shell sources the aliases. owner_prepares runs once per binary, not once per owner. Limited to owners whose three files mention dev-prepare. Cache the result per repo in the loop, and note in the atlas that these Makefiles should keep parsing cheap.'
          family: alias-emission-cost
          round: 3
      recipe: small-diff-review
      isolation: fenced
      reviewed: 31c214bd28e4ac07f4167701fda48c6c3e465aef
      blocked: false
---

# Gate ledger — ariadne#299 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T16:57:53-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `owner-declaration-detection-shape` dev-prepare detection regex accepts only `dev-prepare:` at column 0
  construct/dev-aliases.sh:225 misses rules that name several targets (`foo dev-prepare:`) and a space before the colon, falsely matches `dev-prepare:=x`, and ignores Makefile.workflow. Document the exact accepted declaration shape in the header comment and in atlas base-layer.md.
- **BR-2** [Minor] `doc-line-wrap` atlas base-layer.md:97 joins new prose onto an existing line, making it overlong
  "...keep the plain form. The function only **builds + runs** — it" should be re-wrapped to match the surrounding line width.

## Round 2 — 2026-10-12T17:00:14-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Regex now accepts multi-target and space-before-colon rules and rejects := (fixtures at dev-aliases.test.sh:79-84). Removing the ([^=]|$) guard turns the := test red. Accepted shape documented in the comment above owner_prepares (construct/dev-aliases.sh:222-226); atlas says "a dev-prepare rule", which is accurate.
- BR-2 — addressed — atlas/workflow/base-layer.md:92-101 re-wrapped, lines 75-79 chars, consistent with the paragraph.

### Raised

- **BR-3** [Minor] `owner-declaration-detection-shape` Text-regex detection of dev-prepare still false-matches recipe lines, `::=`, target-specific vars
  2nd finding in this family. The rule: inferring make declarations by regex over Makefile text leaves gaps however it is patched. Verified matches at construct/dev-aliases.sh:228 for a tab-indented recipe line "@echo run dev-prepare: now", for "dev-prepare ::= x", and for "dev-prepare: V = 1"; each emits `make -s dev-prepare`, which then fails every call. Class fix: query make's own database (make -C owner -pRrq :) for a dev-prepare target with a rule, which also covers Makefile.workflow and other includes. Cheaper fallback: exclude tab-led lines and document ::= and target-specific vars as unsupported.

## Round 3 — 2026-10-12T17:05:09-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — owner_prepares now asks make (make -q dev-prepare) behind a grep prefilter; restoring 1b95ae89's version makes 3 new cases fail (target-specific var, recipe line, Makefile.workflow include); head passes 38/38.

### Raised

- **BR-4** [Minor] `doc-restates-detection-rule` Prose about how dev-prepare is detected still describes the old regex
  Instances: (a) the construct/dev-aliases.sh:15 header says "Makefile or Makefile.local" and omits Makefile.workflow, which line 222 and the atlas name; (b) the issue Plan bullet 1 says owner_prepares "greps ^dev-prepare:"; (c) Plan bullet 2 claims a test that a .PHONY mention alone is not a declaration, but none exists and make -q treats a bare .PHONY: dev-prepare as a rule (rc 0, harmless no-op). Rule: describe detection once, in the owner_prepares comment, and point other prose at it; add a Plan revision.
- **BR-5** [Minor] `alias-emission-cost` make -q runs owner Makefile parse-time side effects at shell startup, once per binary
  Parsing the owner's Makefiles runs $(shell ...) calls (verified: a touch fired) and can rebuild included makefiles, every time a shell sources the aliases. owner_prepares runs once per binary, not once per owner. Limited to owners whose three files mention dev-prepare. Cache the result per repo in the loop, and note in the atlas that these Makefiles should keep parsing cheap.

## Open findings

- **BR-4** [Minor] `doc-restates-detection-rule` Prose about how dev-prepare is detected still describes the old regex
- **BR-5** [Minor] `alias-emission-cost` make -q runs owner Makefile parse-time side effects at shell startup, once per binary
