---
gate: boundary-review
issue: 203
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-22T10:34:51-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
          detail: |-
            A paragraph-level scan of cmd/sdlc/helptext/*.md for findings-word + directive-verb finds
            close.md:13 (POST-VERDICT PROTOCOL, "fix the findings NOW" — twin of the routed close.go:1809),
            close.md:59 ("Fix the findings, or have the next review dispose them" — twin of the routed
            close.go:1194), and milestone-close.md:38 ("fix the findings before committing"; the file was
            not touched at all). Spec bullet 2 says EVERY surface routes. The Go class got a mechanical
            scan that found 8 where a hand pass found 4; the doc class got a hand pass. ARCH-PURPOSE.
          family: hand-enumerated-class
          round: 1
        - id: BR-2
          severity: Important
          title: Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
          detail: |-
            Verified by appending unrouted sites to cmd/sdlc/state.go and re-running the guard. Blind to
            (a) a message split across adjacent string literals — the tree's own prevailing style, e.g.
            close.go:1194 is 3 literals and changecode.go:554 became 2 in this diff; (b) a package-level
            const, since pass 2 walks only FuncDecl bodies; (c) a directive verb outside "fix "/"address "/
            "review above"; and (d) a second unrouted line inside an already-routing builder func, because
            pass 2 is function-granular — the exact thing pass 1's own comment declares insufficient.
            atlas/workflow/gate-state.md:127 and Done-when 4 claim "a ninth refusal site cannot ship
            unrouted". Fix: concatenate a call's literals before matching; make pass 2 count-based.
          family: guard-narrower-than-claimed-class
          round: 1
        - id: BR-3
          severity: Minor
          title: New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
          detail: |-
            helptext/close.md:79 and helptext/change-code.md:37 cite ARCH-PURPOSE with no drift guard;
            TestFixTheClassLine_RoutesToArchPrinciples covers only fixTheClassLine, and
            TestArchitecture_NarrativeRoutesToArchPrinciples reads AGENTS.md only.
          family: guard-narrower-than-claimed-class
          round: 1
        - id: BR-4
          severity: Minor
          title: No test asserts a real gate path's stderr actually contains the routing line
          detail: |-
            Coverage is source-level plus string-level with nothing joining them. changecode_test.go:364
            already drives the blocking runPlanQualityJudge path with a stub judge; swapping ioDiscard()
            for a bytes.Buffer and asserting on "ARCH-PURPOSE" is a two-line end-to-end pin.
          family: coverage-stops-at-the-seam
          round: 1
        - id: BR-5
          severity: Minor
          title: The "\n  " join prefix is hand-written at six of the eight call sites
          detail: |-
            changecode.go:554,572,722, close.go:1238, judge.go:172, milestoneclose.go:626 each spell the
            same newline+indent. A fixTheClassNote() returning the pre-joined form would leave only the
            two genuinely different indents (close.go:1194, close.go:1809).
          family: repeated-join-prefix
          round: 1
        - id: BR-6
          severity: Minor
          title: The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
          detail: |-
            gatefindings_test.go:70 globs "./*.go", excluding subpackages; one comment naming that as the
            deliberate class boundary would prevent a wrong assumption. cmd/doc-review/review.go:125 prints
            "triage each finding" to the main agent — the same framing in another binary, deserving an
            explicit in-or-out ruling.
          family: undocumented-scan-boundary
          round: 1
      blocked: true
---

# Gate ledger — ariadne#203 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-22T10:34:51-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `hand-enumerated-class` Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
  A paragraph-level scan of cmd/sdlc/helptext/*.md for findings-word + directive-verb finds
  close.md:13 (POST-VERDICT PROTOCOL, "fix the findings NOW" — twin of the routed close.go:1809),
  close.md:59 ("Fix the findings, or have the next review dispose them" — twin of the routed
  close.go:1194), and milestone-close.md:38 ("fix the findings before committing"; the file was
  not touched at all). Spec bullet 2 says EVERY surface routes. The Go class got a mechanical
  scan that found 8 where a hand pass found 4; the doc class got a hand pass. ARCH-PURPOSE.
- **BR-2** [Important] `guard-narrower-than-claimed-class` Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
  Verified by appending unrouted sites to cmd/sdlc/state.go and re-running the guard. Blind to
  (a) a message split across adjacent string literals — the tree's own prevailing style, e.g.
  close.go:1194 is 3 literals and changecode.go:554 became 2 in this diff; (b) a package-level
  const, since pass 2 walks only FuncDecl bodies; (c) a directive verb outside "fix "/"address "/
  "review above"; and (d) a second unrouted line inside an already-routing builder func, because
  pass 2 is function-granular — the exact thing pass 1's own comment declares insufficient.
  atlas/workflow/gate-state.md:127 and Done-when 4 claim "a ninth refusal site cannot ship
  unrouted". Fix: concatenate a call's literals before matching; make pass 2 count-based.
- **BR-3** [Minor] `guard-narrower-than-claimed-class` New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
  helptext/close.md:79 and helptext/change-code.md:37 cite ARCH-PURPOSE with no drift guard;
  TestFixTheClassLine_RoutesToArchPrinciples covers only fixTheClassLine, and
  TestArchitecture_NarrativeRoutesToArchPrinciples reads AGENTS.md only.
- **BR-4** [Minor] `coverage-stops-at-the-seam` No test asserts a real gate path's stderr actually contains the routing line
  Coverage is source-level plus string-level with nothing joining them. changecode_test.go:364
  already drives the blocking runPlanQualityJudge path with a stub judge; swapping ioDiscard()
  for a bytes.Buffer and asserting on "ARCH-PURPOSE" is a two-line end-to-end pin.
- **BR-5** [Minor] `repeated-join-prefix` The "\n  " join prefix is hand-written at six of the eight call sites
  changecode.go:554,572,722, close.go:1238, judge.go:172, milestoneclose.go:626 each spell the
  same newline+indent. A fixTheClassNote() returning the pre-joined form would leave only the
  two genuinely different indents (close.go:1194, close.go:1809).
- **BR-6** [Minor] `undocumented-scan-boundary` The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
  gatefindings_test.go:70 globs "./*.go", excluding subpackages; one comment naming that as the
  deliberate class boundary would prevent a wrong assumption. cmd/doc-review/review.go:125 prints
  "triage each finding" to the main agent — the same framing in another binary, deserving an
  explicit in-or-out ruling.

## Open findings

- **BR-1** [Important] `hand-enumerated-class` Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
- **BR-2** [Important] `guard-narrower-than-claimed-class` Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
- **BR-3** [Minor] `guard-narrower-than-claimed-class` New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
- **BR-4** [Minor] `coverage-stops-at-the-seam` No test asserts a real gate path's stderr actually contains the routing line
- **BR-5** [Minor] `repeated-join-prefix` The "\n  " join prefix is hand-written at six of the eight call sites
- **BR-6** [Minor] `undocumented-scan-boundary` The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
