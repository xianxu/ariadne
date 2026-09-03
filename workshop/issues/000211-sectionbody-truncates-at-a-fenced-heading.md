---
id: 000211
status: open
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# SectionBody truncates at a fenced heading

## Problem

`issue.SectionBody` (`cmd/sdlc/internal/issue/section.go:15`) extracts a
section with:

    (?ms)^## <heading>\s*\n(.*?)(?:^## |\z)

The terminator `^## ` matches anywhere at line start — including **inside a
fenced code block**. So a section that quotes markdown containing a `##`
heading is silently cut off at that line.

Hit live on ariadne#208, whose Spec quotes the registry entry it adds:

    ## Spec
    ... 35 words of prose ...
    ```markdown
    ## ARCH-SECURE — Name what is trusted     <-- SectionBody stops here
    ...
    ```
    ... several hundred more words, invisible ...

`sdlc change-code` refused with `` `## Spec` has 35 words; need ≥ 50 `` against a
Spec of ~700 words. The gate fired for a real-looking reason that was false,
which is the worst kind: the operator's choices are to `--force` past a guard
that is working correctly on bad input, or to contort the document until the
parser can read it.

**Silent under-reading is the general risk.** `SectionBody` is not only the
structural gate's input — it backs section extraction wherever issue bodies are
read. A section that *looks* present and non-empty but is truncated at its first
quoted heading can make a check pass on a fragment just as easily as fail on one.
Quoting markdown in an issue is not exotic here: any issue that adds prose to a
registry, a skill, a helptext or a datatype does it.

## Spec

Make the section terminator fence-aware: `^## ` ends a section only when it is
not inside a fenced code block. Track fence state (``` and ~~~, including longer
runs and info strings) while scanning lines, rather than extending the regex —
fenced-block nesting is not a regular language and the next quoted-heading
variant would reopen this.

Scope is the extraction primitive, not its callers: every consumer inherits the
fix. Confirm which checks change behavior on the existing corpus before landing,
so a section that has been silently truncated does not start failing a gate it
used to pass for the wrong reason.

## Done when

- A section whose body contains a fenced block with `## ` inside is extracted in
  full; `sdlc change-code --issue 208` passes `spec-present` on the real word
  count with no `--force`.
- Fence forms covered by table test: ``` and ~~~, longer runs, info strings
  (```markdown), an unterminated fence, and a `## ` that is genuinely the next
  section immediately after a closed fence.
- `sdlc issue validate --all` over `workshop/issues/` + `workshop/history/` is
  run before and after; any issue whose section presence/length verdict CHANGES
  is listed in the Log with the reason.

## Plan

- [ ] Replace the regex terminator with a fence-aware line scan in `SectionBody`.
- [ ] Table-test the fence forms above, including the adversarial "closed fence
      then a real heading" case.
- [ ] Diff `issue validate --all` verdicts across the corpus before/after; record
      the changed set.

## Log

### 2026-09-02

Found while running `sdlc change-code --issue 208`, an issue whose Spec quotes
the two registry entries it adds — so the parser stopped at the first quoted
`## ARCH-SECURE` and reported 35 words against a ~700-word Spec. #208 was
unblocked by adding a summary paragraph *before* the fence, which the Spec wanted
anyway; the parser bug is untouched and lives here.
