# Boundary Review — ariadne#203 (whole-issue close)

| field | value |
|-------|-------|
| issue | 203 — gate refusals tell the fixer to address findings, not to fix the class they belong to |
| repo | ariadne |
| issue file | workshop/issues/000203-gate-refusals-tell-the-fixer-to-address-findings-not-to-fix-the-class-they-belong-to.md |
| boundary | whole-issue close |
| milestone | — |
| window | 63ae82c087e6f292e9da39d0b348d948f5d74f57..3a508b8e62546933ea78da0c34a186cdfc1765a6 |
| command | sdlc close --issue 203 |
| reviewer | claude |
| timestamp | 2026-08-22T10:34:51-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core of the issue is genuinely delivered and I verified it independently rather than trusting the Log: all eight code sites route through the one pure formatter (reverting the wiring makes `TestEveryFixerFacingSiteRoutes` fail naming exactly those eight, including both `finalizeBoundaryReview` arms), the drift guard really goes red when `ARCH-PURPOSE` loses the discipline, and the four re-captured goldens diff byte-identically to the `architecture.md` diff (md5-equal), which substantiates the Log's judgment call about the ⛔. What blocks a clean SHIP is that this issue's own thesis — *enumerate the class mechanically, don't hand-list instances* — was applied to the Go sites and not to the two sibling classes: the **doc surfaces** were enumerated by hand and are short by three, and the **enumeration guard's class signature** is narrower than the class it claims to defend (I shipped four synthetic ninth sites past it green). Both are cheap.

### 1. Strengths

- `cmd/sdlc/gatefindings_test.go:96` — pass 1's **call granularity** is the right call and its stated reason holds up: reverting the wiring names `close.go:1194` and `close.go:1237` separately, which a function-level check on `finalizeBoundaryReview` would not have.
- `cmd/sdlc/gatefindings_test.go:33` — the drift guard asserts the *destination* still carries what the citation promises, not just that the marker string appears. Stripping the six-line block from `architecture.md` fails it on both `"CLASS"` and `"family:"`. That is the half most reviewers omit.
- `cmd/sdlc/gatefindings.go:1-25` — the header is load-bearing, not decorative: it states why this is a file (the `gatepersist.go` precedent), why it routes instead of restating (#128), and why `ArchitectureBlock`'s dangling-pointer warning does not bind at these sites. That last distinction is the non-obvious one and it is argued, not asserted.
- Golden re-capture verified: the diff on `dry`/`milestone-review`/`plan-quality`/`pure` `.prompt` is md5-identical to the `architecture.md` diff — the Log's "exactly the twelve added lines and nothing else" is true. `estimate-quality`/`plan`/`specs` legitimately carry no registry.
- The `AGENTS.base.md` non-goal checks out: it routes to `sdlc arch-principles` and never restates `ARCH-PURPOSE`, so downstream repos inherit the extension for free (`AGENTS.base.md:87-92`).

### 2. Critical findings

None.

### 3. Important findings

**I-1 — The doc-surface class was hand-enumerated; a mechanical scan finds three more, two of which are doc twins of code sites this diff routed.** (`cmd/sdlc/helptext/close.md:13`, `close.md:59`, `cmd/sdlc/helptext/milestone-close.md:38`)

The issue's own "Doc surfaces" line names exactly two sections (`close.md` FINDING FAMILIES, `change-code.md` THE PLAN GATE) and records — as the issue's thesis — that "a hand enumeration found seven and was still wrong; the mechanical scan found eight." The code side got that scan. The doc side did not. A paragraph-level scan of `cmd/sdlc/helptext/*.md` for *findings-word + directive-verb* surfaces three unrouted blocks:

- `close.md:13` "POST-VERDICT PROTOCOL (#174) … **fix the findings NOW**" — the doc twin of `close.go:1809`, which this diff routed.
- `close.md:59-61` "**Fix the findings**, or have the next review dispose them explicitly" — the doc twin of `close.go:1194`, which this diff routed.
- `milestone-close.md:38-40` "**fix the findings** before committing, bundle them into the one milestone-close commit" — `milestone-close.md` was not touched at all.

Spec bullet 2 says "**Every** surface that hands findings to the fixer routes to that statement and cites the marker." Fix sketch: append the same one-clause route (`— see ARCH-PURPOSE (`sdlc arch-principles`)`) to those three blocks, or record them in the issue as explicit *reasoned* exclusions. Either is fine; leaving the enumeration silently short is not, because that is the defect the issue exists to fix. `ARCH-PURPOSE` at-review.

**I-2 — The enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green.** (`cmd/sdlc/gatefindings_test.go:60`, `:96`, `:120`)

`atlas/workflow/gate-state.md:127` claims "a ninth refusal site cannot ship unrouted," and Done-when 4 makes the same claim. I appended four synthetic unrouted sites to `cmd/sdlc/state.go` and re-ran the guard:

| synthetic site | shape | guard |
|---|---|---|
| A | single literal, `cwarn` | **FAIL** ✓ (reproduces `state.go:458`) |
| B | message split across two adjacent literals | **green** ✗ |
| C | package-level `const`, outside any `FuncDecl` | **green** ✗ |
| D | directive verb `resolve` (outside `fix `/`address `/`review above`) | **green** ✗ |

Case B is not hypothetical — it is the tree's prevailing style: `close.go:1194` is a three-literal concatenation, and `changecode.go:554` became a two-literal one *in this diff*. It only passes today because "address " and "dispose" happen to land in the same first literal.

Separately, pass 2 (`:96`) checks `referencesRoutingLine(fn.Body)` at **function** granularity — the exact thing pass 1's own comment declares insufficient. I added a second unrouted fixer-facing string to `formatFixThenShipProtocol` (which already routes once) and the guard stayed green.

Fix sketch, all cheap: (a) in pass 1, apply `fixerFacingLiteral` to the *concatenation* of the call's literals, not each one; (b) in pass 2, require `count(routing refs) >= count(matching literals)` in the func rather than a boolean; (c) either widen `hasDirective` or state in the test doc that the vocabulary is the deliberate class boundary. If you'd rather not strengthen the mechanism, soften the atlas + Done-when claim instead — but per PQ-2's own words, don't keep the strong claim over the weak mechanism. `ARCH-PURPOSE` / `ARCH-DRY`.

### 4. Minor findings

- `cmd/sdlc/gatefindings.go:28` — the `"\n  "` join prefix is hand-written at six call sites (`changecode.go:554,572,722`, `close.go:1238`, `judge.go:172`, `milestoneclose.go:626`). A second `fixTheClassNote()` returning the pre-joined form would leave only the two genuinely different indents (`close.go:1194`, `close.go:1809`).
- The new `ARCH-PURPOSE` citations in `helptext/close.md:79` and `helptext/change-code.md:37` are unguarded — only `fixTheClassLine()` is covered by the drift guard, and `TestArchitecture_NarrativeRoutesToArchPrinciples` reads `AGENTS.md` only. A rename of the marker would leave those two dangling silently.
- `cmd/sdlc/gatefindings_test.go:70` — `filepath.Glob("./*.go")` scopes the scan to the `cmd/sdlc` package dir. That is almost certainly right, but it is undocumented; one comment line naming it as the class boundary would stop a future reader from assuming subpackages are covered.
- `cmd/doc-review/review.go:125` prints "triage **each** finding" to the main agent — the same per-site framing, in a different binary. Probably out of scope, but it is the one in-tree sibling outside `cmd/sdlc` and deserves an explicit in-or-out ruling rather than silence.

### 5. Test coverage notes

Both guards are real and I confirmed both go red on their own regression, which is the bar the `#194` claimed-fixes rule sets. Gaps worth noting:

- There is **no assertion that a real gate path's stderr actually contains the line.** Coverage is source-level (`TestEveryFixerFacingSiteRoutes`) plus string-level (`TestFixTheClassLine_RoutesToArchPrinciples`), with nothing joining them. `changecode_test.go:364` already drives the blocking `runPlanQualityJudge` path with a stub judge; swapping `ioDiscard()` for a `bytes.Buffer` and asserting `strings.Contains(buf.String(), "ARCH-PURPOSE")` is a two-line end-to-end pin. Cheap, and it would catch a future refactor that routes in source but drops the emission.
- Helptext changes are covered only by `TestCloseEmbedded`-style existence checks; nothing pins the new routing prose. Given I-1, a scan-based helptext guard (mirroring the Go one) would close the doc class the same way the Go class was closed — and would have caught I-1 itself.

### 6. Architectural notes

- **ARCH-DRY — pass.** One formatter, eight consumers; the header argues the extraction against the `gatepersist.go` precedent. The thesis sentence now appears in five places (`fixTheClassLine`, two helptext files, two atlas files), but all five *route and cite* rather than restate the enumerate-and-sweep procedure, which is the #128 pattern working as designed. Minor nit at §4.
- **ARCH-PURE — pass.** `fixTheClassLine()` is a constant-returning pure function, tested without IO by the drift guard. The enumeration guard does filesystem IO, but it is a meta-test over sources where that is intrinsic, not logic buried in a shell.
- **ARCH-MOCK — pass, n/a.** No external binary or service surface is introduced; the judge dispatch rides the existing injected seam (`stubJudgeName`), which the new tests do not disturb.
- **ARCH-PURPOSE — flag (I-1, I-2).** The shadow-sweep over the *code* consumers is complete and verified. The sweep over the two sibling classes is not: the doc consumers were hand-enumerated and are short by three, and the guard that was supposed to make the code enumeration durable has four demonstrated in-class blind spots. Both are the diff's own newly-added at-review lens applied to itself: "a fix that resolves the site a prior finding named while enumerable siblings of the same class remain in the tree: that is the instance, not the class."
- Forward note: the `⛔ NEVER re-run -update-golden` comment in `golden_test.go:37` genuinely does not cover a deliberate registry edit. The Log correctly scopes that out as a different class — but it will recur on the next `ARCH-*` edit, so it is worth its own issue rather than a Log note.

### 7. Plan revision recommendations

The plan lives in the issue (`workshop/issues/000203-*.md`), per the artifact-tier judgment call in the Log. Two `## Revisions` entries:

1. **"The class (enumeration)" — the doc-surface half was hand-enumerated.** The Go half was scanned mechanically and found 8 where a hand pass found 4; the "Doc surfaces:" line was written by hand and names 2 where a paragraph-level scan of `cmd/sdlc/helptext/*.md` finds 5 (`close.md:13`, `close.md:59`, `close.md:68`, `change-code.md`, `milestone-close.md:38`). Record the scan command, the full list, and the in-or-out ruling for each.
2. **Done-when 4 — reconcile the claim with the mechanism.** "asserting every match routes" currently overstates what `fixerFacingLiteral` + the two-pass granularity delivers. Either record the strengthening (per-call literal concatenation, count-based pass 2) or state the four known blind spots as the guard's declared boundary. `atlas/workflow/gate-state.md:127`'s "a ninth refusal site cannot ship unrouted" needs the same treatment.

```findings
findings:
  - id: new
    severity: Important
    family: hand-enumerated-class
    title: |
      Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
    detail: |
      A paragraph-level scan of cmd/sdlc/helptext/*.md for findings-word + directive-verb finds
      close.md:13 (POST-VERDICT PROTOCOL, "fix the findings NOW" — twin of the routed close.go:1809),
      close.md:59 ("Fix the findings, or have the next review dispose them" — twin of the routed
      close.go:1194), and milestone-close.md:38 ("fix the findings before committing"; the file was
      not touched at all). Spec bullet 2 says EVERY surface routes. The Go class got a mechanical
      scan that found 8 where a hand pass found 4; the doc class got a hand pass. ARCH-PURPOSE.
  - id: new
    severity: Important
    family: guard-narrower-than-claimed-class
    title: |
      Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
    detail: |
      Verified by appending unrouted sites to cmd/sdlc/state.go and re-running the guard. Blind to
      (a) a message split across adjacent string literals — the tree's own prevailing style, e.g.
      close.go:1194 is 3 literals and changecode.go:554 became 2 in this diff; (b) a package-level
      const, since pass 2 walks only FuncDecl bodies; (c) a directive verb outside "fix "/"address "/
      "review above"; and (d) a second unrouted line inside an already-routing builder func, because
      pass 2 is function-granular — the exact thing pass 1's own comment declares insufficient.
      atlas/workflow/gate-state.md:127 and Done-when 4 claim "a ninth refusal site cannot ship
      unrouted". Fix: concatenate a call's literals before matching; make pass 2 count-based.
  - id: new
    severity: Minor
    family: guard-narrower-than-claimed-class
    title: |
      New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
    detail: |
      helptext/close.md:79 and helptext/change-code.md:37 cite ARCH-PURPOSE with no drift guard;
      TestFixTheClassLine_RoutesToArchPrinciples covers only fixTheClassLine, and
      TestArchitecture_NarrativeRoutesToArchPrinciples reads AGENTS.md only.
  - id: new
    severity: Minor
    family: coverage-stops-at-the-seam
    title: |
      No test asserts a real gate path's stderr actually contains the routing line
    detail: |
      Coverage is source-level plus string-level with nothing joining them. changecode_test.go:364
      already drives the blocking runPlanQualityJudge path with a stub judge; swapping ioDiscard()
      for a bytes.Buffer and asserting on "ARCH-PURPOSE" is a two-line end-to-end pin.
  - id: new
    severity: Minor
    family: repeated-join-prefix
    title: |
      The "\n  " join prefix is hand-written at six of the eight call sites
    detail: |
      changecode.go:554,572,722, close.go:1238, judge.go:172, milestoneclose.go:626 each spell the
      same newline+indent. A fixTheClassNote() returning the pre-joined form would leave only the
      two genuinely different indents (close.go:1194, close.go:1809).
  - id: new
    severity: Minor
    family: undocumented-scan-boundary
    title: |
      The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
    detail: |
      gatefindings_test.go:70 globs "./*.go", excluding subpackages; one comment naming that as the
      deliberate class boundary would prevent a wrong assumption. cmd/doc-review/review.go:125 prints
      "triage each finding" to the main agent — the same framing in another binary, deserving an
      explicit in-or-out ruling.
```
