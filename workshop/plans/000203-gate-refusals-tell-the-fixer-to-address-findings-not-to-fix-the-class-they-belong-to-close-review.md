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

---

## Re-review — 2026-08-22T12:33:48-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 203 — gate refusals tell the fixer to address findings, not to fix the class they belong to |
| repo | ariadne |
| issue file | workshop/issues/000203-gate-refusals-tell-the-fixer-to-address-findings-not-to-fix-the-class-they-belong-to.md |
| boundary | whole-issue close |
| milestone | — |
| window | 63ae82c087e6f292e9da39d0b348d948f5d74f57..99bb7a5ab57b120fa7231c8d8af87cfa4083f63b |
| command | sdlc close --issue 203 |
| reviewer | claude |
| timestamp | 2026-08-22T12:33:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 1's six findings are all genuinely worked, and I verified the substantive ones by reverting rather than by reading the Log: reverting the three doc routings makes `TestEveryFixerFacingHelptextRoutes` fail naming exactly `close.md:13`, `close.md:59`, `milestone-close.md:38` (BR-1); all four BR-2 shapes plus a fifth I constructed (split literal, package-level const, `resolve` verb, second unrouted line in a routing builder) now fail the enumeration guard; and stripping the routing from `classifyFallback` turns `TestGatePathStderrCarriesRoutingLine` red with real stderr (BR-4). The real tree is fully covered — a wide scan for `finding` + any imperative across non-test `cmd/sdlc/*.go` turns up nothing unrouted that isn't already a reasoned exclusion. What holds a clean SHIP is that the guard-widening was again done at the shapes the finding *named*: two further shapes of the same class ship green (a fixer-facing message split across adjacent literals **outside** a `cwarn`/`die` call, and a routing ref inside a pass-1-claimed call being credited to an unrelated unrouted literal in the same func), so `atlas/workflow/gate-state.md:127`'s "a ninth refusal site cannot ship unrouted" is still stronger than the mechanism. That is the third round of `guard-narrower-than-claimed-class`, so the deliverable is the rule, not two more patches.

### 1. Strengths

- `cmd/sdlc/gatefindings_test.go:113-127` — pass 1's literal **concatenation** is real and load-bearing, not cosmetic: my synthetic split-literal `cwarn` (the tree's own prevailing style, `close.go:1194` being three literals) fails correctly where round 1's version passed.
- The doc guard's two exclusions are **structural, not line-keyed** (`isQuotedOutput` ≥6-space blocks, `referenceSections`), so they can't rot as the helptext is edited. I confirmed `close.md`'s convergence-line examples fall out via the indent rule rather than a hardcode, and that a genuine directive in those files still fires.
- `cmd/sdlc/gatefindings_test.go:33-46` — the drift guard asserts the *destination* still carries what the citation promises (`CLASS`/`enumeration`/`family:` in the `ARCH-PURPOSE` entry), not merely that the marker string appears. That's the half most such guards omit.
- `TestGatePathStderrCarriesRoutingLine` (`:355`) picks a seam with an existing driver and says so, and it rides the same `judge.Run` var the rest of the package stubs (`boundaryledger_test.go:244`, `changecode_test.go:153`) — consistent with the house ARCH-MOCK seam rather than a new one.
- `atlas/workflow/architecture-principles.md:18-29` routes to the registry and never restates the `ARCH-PURPOSE` body — the shadow-sweep over the registry's consumers (`architecture.go`, `judge_test.go`, four goldens, `gatefindings.go`, `AGENTS.base.md`) comes back clean; nothing hand-maintains a second copy of the model.

### 2. Critical findings

None.

### 3. Important findings

**I-1 — Two more shapes of the enumeration guard's class ship green; this is the 3rd round of `guard-narrower-than-claimed-class`.** (`cmd/sdlc/gatefindings_test.go:113`, `:131-149`)

Verified in a scratch copy, both green:

| shape | construction | guard |
|---|---|---|
| F | fixer-facing message split across adjacent literals in a call that is **not** `cwarn`/`die` (e.g. `fmt.Fprintf`) | **green** ✗ |
| H | a func whose `cwarn` routes (claimed by pass 1) *and* which builds a second, unrouted fixer-facing line | **green** ✗ |

F is reachable: `cmd/sdlc` non-test files hold ~200 `fmt.Fprint*` calls and ~70 adjacent-literal concatenations. H is reachable in the exact functions this issue touches — `finalizeBoundaryReview` already has two claimed routing refs, so an added builder line there has two credits to spend.

Per the family escalation, **do not patch F and H.** State the rule: *every fixer-facing message must be matched as a whole string-valued expression, and each match must be attributed to a routing reference in its own syntactic unit.* Two concrete forms of that one rule — (a) fold `+`-joined literal chains (`ast.BinaryExpr`) into one text before matching **wherever they occur**, not only inside two hardcoded emitter names; (b) in pass 2, count only routing refs that are *not* inside a pass-1-claimed call. Better still, collapse the two name-keyed passes into one pass over string-valued expressions attributed to the nearest enclosing statement, which makes F and H the same case. The narrow `fixerFacingMessage` verb list is shared by both guards, so the same rule covers the doc scan. If you'd rather not strengthen the mechanism, then per PQ-2's own words don't keep the strong claim over the weak one: soften `atlas/workflow/gate-state.md:127` and Done-when 4 to what the two passes actually deliver, and record F/H as the declared boundary. `ARCH-PURPOSE` at-review.

### 4. Minor findings

- **`helptext/change-code.md:37`'s `ARCH-PURPOSE` citation is unguarded** — I removed it and no test failed. Neither that paragraph nor the one before it carries a directive verb, so `fixerFacingMessage` never matches and the doc guard never reaches it. `close.md:79`'s citation *is* guarded (removing it fires at `close.md:70` via the next-paragraph rule). This is why BR-3 is disposed `not-addressed` below rather than re-raised: half of it landed. Done-when 5 can regress silently on the `change-code.md` half.
- **The issue's own "Doc surfaces:" line is now stale** — `workshop/issues/000203-*.md` still reads "`helptext/close.md` FINDING FAMILIES and `helptext/change-code.md` THE PLAN GATE", i.e. two, where five paragraphs across three files were delivered. Done-when 5 likewise names two files and omits `milestone-close.md`. The Log narrates the change but the enumeration section — the section whose whole point is being mechanical — was left hand-maintained and wrong. 2nd in `hand-enumerated-class`; the rule is to have that section *route* at the guard (`TestEveryFixerFacingSiteRoutes` / `TestEveryFixerFacingHelptextRoutes` compute the class) instead of re-listing sites that drift.
- **`construct/adapted/superpowers-receiving-code-review/SKILL.md` is an unruled sibling and actively contradicts this issue.** Its response pattern ends `6. IMPLEMENT: One item at a time, test each` — the per-site patching #203 exists to stop — and it is the repo's canonical *reception* surface, invoked exactly when a gate hands findings over. It escapes the doc scan only because it says "feedback"/"items", never "finding". 2nd in `undocumented-scan-boundary`; the rule is that the class boundary is declared per-glob but the **set of globs** never is, so Non-goals rules out siblings from memory (doc-review) rather than from an enumeration.
- `cmd/sdlc/gatefindings_test.go:344` — `describe(fset, pos, path)` is a very generic name for a package-`main` test helper; `describePos` would age better.

### 5. Test coverage notes

Coverage is now three-layered and each layer goes red on its own regression, which is the `#194` bar: source scan (Go), source scan (docs), and one end-to-end stderr pin. Remaining gaps, all noted above rather than separately raised:

- The e2e pin covers `classifyFallback` only. The `d.Block` ledger arm of `runPlanQualityJudge`, `finalizeBoundaryReview`'s two arms, and `formatFixThenShipProtocol` are source-pinned but never observed on a writer. The test's own comment is honest about this ("one path, not the class"), and the scans do cover the class — so this is acceptable as-is, *provided* I-1 is resolved, since the scans are the thing carrying that weight.
- No test pins `fixTheClassNote()`'s `"\n  "` shape. A regression to `" "` would leave every one-line refusal running the routing text onto the same line and every test would stay green. One `strings.HasPrefix(fixTheClassNote(), "\n  ")` assertion in the drift guard covers it.

### 6. Architectural notes

- **ARCH-DRY — pass.** One formatter, one pre-joined variant, eight consumers; BR-5's six hand-spelled indents are gone and the two survivors are argued at `gatefindings.go:33-38`. No second copy of the principle anywhere — the atlas file and both helptext files route and cite.
- **ARCH-PURE — pass.** `fixTheClassLine`/`fixTheClassNote` are constant-returning pure functions tested without IO. The scan guards do filesystem IO, but they are meta-tests over sources where that is intrinsic, not business logic in a shell.
- **ARCH-MOCK — pass.** No new external dependency. The e2e test overrides the existing `judge.Run` seam with `t.Cleanup` restore, matching `boundaryledger_test.go`'s convention; production and test flow share that boundary.
- **ARCH-PURPOSE — flag (I-1).** The shadow-sweep over the code and doc consumers is complete and I re-derived it independently. What is not complete is the *durability* claim: the guard was widened to the four shapes BR-2 enumerated, and two more shapes of the same class remain. That is the diff's own newly-written at-review lens applied to itself — "a fix that resolves the site a prior finding named while enumerable siblings of the same class remain in the tree: that is the instance, not the class." Encouragingly, the *first-order* purpose (the routing itself) was delivered at class level, twice; the recursion is one level down each round, which is convergence, not stall.
- **Forward note.** `golden_test.go:37`'s ⛔ still doesn't cover a deliberate registry edit, and the atlas now documents the carve-out in prose rather than in the guard. It will recur on the next `ARCH-*` edit; worth its own issue.

### 7. Plan revision recommendations

The plan lives in the issue. Two `## Revisions` entries:

1. **"The class (enumeration)" — the Doc surfaces line is stale and hand-maintained.** It names two surfaces; five paragraphs across three files were delivered (`close.md` POST-VERDICT PROTOCOL, `close.md` MODES ledger bullet, `close.md` FINDING FAMILIES, `change-code.md` THE PLAN GATE, `milestone-close.md`). Replace the hand list with a pointer at `TestEveryFixerFacingHelptextRoutes` as the authority, and update Done-when 5 to include `milestone-close.md`.
2. **Done-when 4 / `atlas/workflow/gate-state.md:127` — reconcile the claim with the mechanism.** "a ninth refusal site cannot ship unrouted" and "asserting every match routes" both overstate what the two passes deliver: shapes F and H demonstrably ship green. Record either the strengthening (whole-expression matching + unclaimed-only routing refs) or the two shapes as the guard's declared boundary.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestEveryFixerFacingHelptextRoutes verified red on revert, naming exactly close.md:13, close.md:59, milestone-close.md:38.
  - id: BR-2
    disposition: addressed
    note: |
      All four named shapes plus a fifth I constructed now fail the guard; verified in a scratch copy.
  - id: BR-3
    disposition: not-addressed
    note: |
      close.md:79's citation is guarded; change-code.md:37's is not — removing it fails no test, since neither it nor its predecessor paragraph carries a directive verb.
  - id: BR-4
    disposition: addressed
    note: |
      Verified red: stripping fixTheClassNote from classifyFallback fails TestGatePathStderrCarriesRoutingLine on real stderr.
  - id: BR-5
    disposition: addressed
    note: |
      fixTheClassNote owns the join at six sites; the two remaining indents are argued at gatefindings.go:33.
  - id: BR-6
    disposition: addressed
    note: |
      Scan boundary documented at the test; doc-review ruled OUT in Non-goals with reasoning.
findings:
  - id: new
    severity: Important
    family: guard-narrower-than-claimed-class
    title: |
      Third round of the same family — two further shapes of the guard's class ship green; state the rule, do not patch F and H
    detail: |
      Verified green in a scratch copy: (F) a fixer-facing message split across adjacent literals in a
      call that is not cwarn/die — pass 2 never concatenates, and cmd/sdlc non-test files hold ~200
      fmt.Fprint* calls and ~70 adjacent-literal concatenations; (H) a func whose cwarn routes (claimed
      by pass 1) and which also builds a second unrouted fixer-facing line — countRoutingRefs credits the
      claimed call's ref to the unclaimed literal, and finalizeBoundaryReview already carries two such
      credits. Rounds 1 and 2 each widened to the shapes the finding named. The RULE: match every
      fixer-facing message as a whole string-valued expression (fold +-joined BinaryExpr chains wherever
      they occur, not only inside two hardcoded emitter names) and attribute each match to a routing ref
      in its own syntactic unit (count only unclaimed refs in pass 2). One pass over string-valued
      expressions attributed to the nearest enclosing statement makes F and H the same case, and the same
      rule covers the doc scan, which shares fixerFacingMessage. Alternative, equally valid at class
      level: soften atlas/workflow/gate-state.md:127 and Done-when 4 to the declared boundary rather than
      keeping the strong claim over the weak mechanism. ARCH-PURPOSE.
  - id: new
    severity: Minor
    family: hand-enumerated-class
    title: |
      The issue's own "Doc surfaces" enumeration is stale — names 2, delivered 5 across 3 files
    detail: |
      Second in this family. The section whose point is being mechanical is still hand-maintained: it
      names close.md FINDING FAMILIES and change-code.md THE PLAN GATE, while BR-1's fix routed five
      paragraphs across close.md, change-code.md and milestone-close.md; Done-when 5 omits
      milestone-close.md. Do not just correct the list — the rule is that a hand-written enumeration of a
      class the guards now compute must ROUTE at the guard (the #128 pattern this issue applies
      everywhere else) instead of restating sites that drift. Same rule covers the "eight sites" table,
      which will go stale the moment a ninth lands.
  - id: new
    severity: Minor
    family: undocumented-scan-boundary
    title: |
      superpowers-receiving-code-review is an unruled sibling that actively contradicts the issue's thesis
    detail: |
      Second in this family. construct/adapted/superpowers-receiving-code-review/SKILL.md is the repo's
      canonical findings-RECEPTION surface — invoked exactly when a gate hands findings over — and its
      response pattern ends "6. IMPLEMENT: One item at a time, test each", the per-site patching #203
      exists to stop. It escapes the doc scan only because it says "feedback"/"items", never "finding".
      The rule: each guard declares its own glob, but the SET of globs (which instruction surfaces are in
      the class) is never declared, so Non-goals rules siblings in or out from memory — doc-review by
      name, this one not at all. Declare the surface set, then rule each member.
```

---

## Re-review — 2026-08-22T12:50:38-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 203 — gate refusals tell the fixer to address findings, not to fix the class they belong to |
| repo | ariadne |
| issue file | workshop/issues/000203-gate-refusals-tell-the-fixer-to-address-findings-not-to-fix-the-class-they-belong-to.md |
| boundary | whole-issue close |
| milestone | — |
| window | 63ae82c087e6f292e9da39d0b348d948f5d74f57..311673e7e9796297ec1396b1deb9d24a0dd75ec5 |
| command | sdlc close --issue 203 |
| reviewer | claude |
| timestamp | 2026-08-22T12:50:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Cleaning up complete. Here's the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 3 finally answered `guard-narrower-than-claimed-class` at the class level rather than widening again, and I verified that independently rather than trusting the Log: the two-pass name-keyed structure is gone, and in a scratch worktree shapes F (split literals in a non-`cwarn` call) and H (a second unrouted line in an already-routing func) both go red, along with a package-level const and an off-list verb — four shapes, one rule, one shorter guard. The helptext guard is non-vacuous (a probe paragraph in `MODES` fires at `close.md:24`), BR-9's `SKILL.md` fix goes red on revert, and the shadow-sweep over every live `ARCH-PURPOSE` consumer comes back clean — no hand-maintained restatement outside the registry and its `-update-golden` derivatives. Two things hold a clean SHIP, both cheap. First, the stated rule was implemented as *membership* (`referencesRoutingLine`) where BR-7 stated it as *counting* — so two fixer-facing messages sharing one statement with one routing ref still ship green, which I demonstrated; the class answer is either the 5-line counting change (verified: zero false positives on the tree) or softening the absolute claim in `atlas/workflow/gate-state.md:127`, because no syntactic approximation can back "cannot ship unrouted". Second, and new: BR-9's fix landed on `construct/adapted/…/SKILL.md`, which is *render output* — the next `/construct upgrade superpowers` + `promote` reconstructs it from `sources/` + `intents/superpowers.md`, and `receiving-code-review` appears nowhere in that intent transcript, so the GENERALIZE step is deleted by design.

### 1. Strengths

- `cmd/sdlc/gatefindings_test.go:200-266` — the rewrite is a genuine class-level answer, not a third widening. One synthetic probe file carrying shape F, shape H, a package-level const, and the verb `triage` produced exactly four violations naming all four. `stringLiterals`, `enclosingFunc`, `countMatchingLiterals`, `countRoutingRefs` and the `claimed` bookkeeping are all deleted — shorter *and* stronger, which is what "state the rule" is supposed to buy.
- `cmd/sdlc/close.go:1808-1814` — the merge into one `append` with the comment naming *why* ("the routing line must share a statement with the message it routes") is the honest shape, not a test-appeasing edit. The guard caught this in the tree, which is the guard doing its job on its own author.
- The doc guard's exclusions are structural, and I confirmed non-vacuity: an unrouted directive dropped into `close.md`'s `MODES` fires at `close.md:24`. `isQuotedOutput` and `referenceSections` are classes, not line numbers, so they don't rot.
- BR-9's catch is the sharpest thing in the issue and the delivery path checks out: `.claude/skills/superpowers-receiving-code-review` is a **symlink** into `construct/adapted/`, and per `construct/skill/construct/SKILL.md` derivatives reach the same dir via the weave layer walk — so the GENERALIZE step is live for this repo and downstream, not shelved. `TestReceivingCodeReviewSkillGeneralizes` goes red on revert on both assertions.
- `## The class (enumeration)` deleting its own tables and routing at the guards (`workshop/issues/000203-*.md:79-103`) is the right answer to BR-8 — it keeps only the evidence that doesn't drift (2 → 7 → 8), which is the part worth recording.

### 2. Critical findings

None.

### 3. Important findings

**I-1 — `referencesRoutingLine` is membership where the stated rule is counting; two messages in one statement, one routing ref, ships green.** (`cmd/sdlc/gatefindings_test.go:187-189`, `:352-364`)

**This is the 4th finding in family `guard-narrower-than-claimed-class`.** Per the escalation I am not asking for a patch, and this is not a new shape — it is the *un-implemented half of the rule BR-7 already stated*: "attribute each match to a routing ref in its own syntactic unit **(count only unclaimed refs)**". Round 3 implemented the unit and dropped the count. Demonstrated green in a scratch copy:

```go
lines = append(lines,
    "1. Fix the findings NOW, before committing this close.",
    "   "+fixTheClassLine(),
    "2. Any remaining findings — address them one at a time.",   // ← unrouted, guard green
)
```

That is reachable in `formatFixThenShipProtocol` specifically, because round 3's own fix put the routing ref *into* that append — the statement now has a credit to spend on any line added beside it. It is shape H exactly one granularity down (function → statement).

The rule, and the reason to stop here: **a syntactic approximation cannot back an absolute claim.** Counting closes this residue, but the next one (two refs both joined to one message, second message unrouted) is one level further down, and so on without bound. So the durable fix is at the *claim*, not the mechanism: `atlas/workflow/gate-state.md:127` still says "so a ninth refusal site cannot ship unrouted", and the paragraph beneath it explains that "per-function attribution lets an already-routing func carry a second unrouted line for free" — a sentence that is still true of the shipped mechanism with "function" replaced by "statement". State what the guard approximates (whole `+`-folded string expressions, attributed per statement, matched on statically visible literals) and the claim becomes true.

Minimum disposition: one sentence at `gate-state.md:127` naming the approximation. I also prototyped the counting change and it is free — `perStmt[m.stmt]++` compared against a `countRoutingRefs(m.stmt)`, ~5 lines, **zero false positives on the real tree** (only my probe fired) and it closes the residue. Both together is the strongest answer; either alone disposes the finding. `ARCH-PURPOSE`.

**I-2 — BR-9's fix edits render output, not the source it derives from; the next `/construct upgrade` deletes it.** (`construct/adapted/superpowers-receiving-code-review/SKILL.md:24-37`)

`construct/adapted/` is generated. `construct/skill/construct/SKILL.md` (`/construct promote`, step 4) is **delete-then-copy**: `rm -rf construct/adapted/superpowers-*/` then copy from staging. Staging is rendered by `/construct upgrade` from `construct/sources/superpowers/<version>/` plus `construct/intents/superpowers.md`, and — step 2 — *"skills not mentioned in the intent → copy new source as-is."* `grep -n "receiving-code-review" construct/intents/superpowers.md` returns **nothing**. So the GENERALIZE step, the ARCH-PURPOSE routing, and the `family:`-as-worklist paragraph are all scheduled for deletion by the substrate's own pipeline.

This is the issue's own principle turned one level out: the deliverable was made at the compiled consumer instead of the source that compiles it — a hand-maintained restatement that does not derive. The precedent is exact and recent: `a547179` (#71), the last ARCH-registry change that touched an adapted skill, appended a `## Conversation 7` with a `### Verify` block to `construct/intents/superpowers.md` in the same commit. Fix sketch: append `## Conversation 8 (2026-08-22): receiving-code-review — GENERALIZE before IMPLEMENT` with the user/AI exchange and verify clauses mirroring `TestReceivingCodeReviewSkillGeneralizes`'s three substrings. `TestReceivingCodeReviewSkillGeneralizes` does make the loss loud rather than silent, which is why this is Important and not Critical — but a red test with no intent record leaves the next reader reconstructing the wording from three `strings.Contains` calls. `ARCH-PURPOSE`.

### 4. Minor findings

- The emitted routing line is 120 chars; at the FIX-THEN-SHIP block's 9-space indent that is 129, against ~83 for every other line in the same block (`close.go:1812`). It will hard-wrap in an 80/100-col terminal mid-sentence. Splitting after the colon would match the surrounding rhythm.
- `fixerFacingSurfaces` (`gatefindings_test.go:52-84`) is a floating comment, not a declaration — nothing named `fixerFacingSurfaces` exists, so the issue's `## The class (enumeration)` routes to a grep target rather than a symbol. It works, but a `var fixerFacingSurfaces = []string{...}` consumed by the globs would make the declared set load-bearing instead of advisory.

### 5. Test coverage notes

Four guards, and I verified each goes red on its own regression, which is the #194 bar: the source scan (probe file, 4/4 shapes), the doc scan (probe paragraph → `close.md:24`), the skill pin (revert → 2 failures), and the e2e stderr pin. `go test ./...` is green repo-wide, `gofmt -l` and `go vet` clean. Remaining gaps:

- The counting residue at I-1 — the only demonstrated hole, and the scans are what carry the durability claim, so it matters more than its size suggests.
- Still no assertion on `fixTheClassNote()`'s `"\n  "` shape (noted last round). A regression to `" "` runs the routing text onto the refusal line at six sites and every test stays green. One `strings.HasPrefix` in the drift guard covers it.
- `TestGatePathStderrCarriesRoutingLine` remains one path (`classifyFallback`); its own comment says so honestly, and the scans cover the class. Acceptable as-is.

### 6. Architectural notes

- **ARCH-DRY — pass.** One formatter, one pre-joined variant, eight consumers; `fixerFacingMessage` is shared by both scans rather than duplicated per guard. The hand-rolled `walk` in `fixerFacingMatches` reimplements parent-tracking that `astutil` would give, but stdlib-only is the house style here and the implementation is correct — I traced the push/pop discipline and the `c == n` re-entry.
- **ARCH-PURE — pass.** `fixTheClassLine`/`fixTheClassNote` are constant-returning, tested without IO. `foldStringExpr`/`isStringExpr`/`fixerFacingMessage`/`isQuotedOutput`/`sectionHeading` are all pure and independently exercisable; the filesystem IO sits in the two test bodies, which is intrinsic to a source scan.
- **ARCH-MOCK — pass.** No new external dependency. The e2e test overrides the existing `judge.Run` seam with `t.Cleanup` restore, matching `boundaryledger_test.go` / `changecode_test.go`; production and test flow share that boundary.
- **ARCH-PURPOSE — flag (I-1, I-2).** The first-order shadow-sweep is complete and I re-derived it: every live `ARCH-PURPOSE` consumer (`archprinciples.go`, `sdlc-binary.md`, `processmanual/session.go`, `lessons.md`, `AGENTS.base.md`, three helptext files, four goldens) *cites* rather than restates, so the registry is genuinely the single source. What remains is one level out in each direction — the guard's claim exceeds the guard (I-1), and the substrate edit sits downstream of its own generator (I-2). The trajectory is convergent: round 1 found four shapes, round 2 two, round 3 one, and the round-3 answer was structural rather than another patch.
- **Forward note.** `golden_test.go:37`'s ⛔ still doesn't cover a deliberate registry edit, and the atlas now documents that carve-out in prose (`architecture-principles.md:15-17`) rather than in the guard. It will recur on the next `ARCH-*` edit — worth its own issue.

### 7. Plan revision recommendations

1. **`## The class (enumeration)` — record the guard's declared approximation.** The section now correctly routes at the guards, but the guards' *boundary* is stated only in `gate-state.md`'s claim, which overstates it (I-1). Add one line: what the scan matches (`+`-folded string expressions, attributed per statement, statically visible literals only) and what it therefore does not.
2. **`## Non-goals` — rule `construct/intents/superpowers.md` in or out.** The issue rules on `cmd/doc-review` and `AGENTS.base.md` explicitly; the substrate's intent transcript is the one surface the diff touched *without* a ruling, and it is the source `adapted/` derives from (I-2).

```findings
dispose:
  - id: BR-3
    disposition: not-addressed
    note: |
      Re-verified: deleting the ARCH-PURPOSE paragraph at change-code.md:37 fails no test — neither it nor its neighbours carry a directive verb, so fixerFacingMessage never reaches them. Class fix: a guard that every ARCH-* marker cited in helptext exists in the registry.
  - id: BR-7
    disposition: addressed
    note: |
      Verified in a scratch worktree: shapes F and H both go red, along with a package-level const and an off-list verb; the two-pass name-keyed structure is deleted, not widened.
  - id: BR-8
    disposition: addressed
    note: |
      The section it named now routes at the guards. Note: Plan items 4-5 still hand-name "the eight sites" and two doc surfaces, but those are checked-off historical steps rather than a live enumeration, and `## Done when` was removed entirely along with its restatements.
  - id: BR-9
    disposition: addressed
    note: |
      Surface set declared; SKILL.md guard verified red on revert on both assertions; the .claude/skills symlink confirms the edit reaches the loaded skill. Its durability is a separate new finding.
findings:
  - id: new
    severity: Important
    family: guard-narrower-than-claimed-class
    title: |
      Statement attribution is membership, not counting — two messages sharing one statement with one routing ref ship green
    detail: |
      4th in this family, and not a new shape: it is the half of BR-7's own stated rule that round 3
      dropped ("count only unclaimed refs"). Verified green in a scratch copy — an append carrying
      "1. Fix the findings NOW..." + fixTheClassLine() + "2. Any remaining findings — address them"
      passes. Reachable in formatFixThenShipProtocol specifically, because round 3's fix put the
      routing ref into that same append. THE RULE, and why to stop widening here: a syntactic
      approximation cannot back an absolute claim — counting closes this residue, the next (two refs
      both joined to one message) is one level further, without bound. So fix the CLAIM:
      atlas/workflow/gate-state.md:127 still says "a ninth refusal site cannot ship unrouted", and the
      paragraph under it describes this exact defect with "function" where "statement" now belongs.
      State what the guard approximates (whole +-folded string expressions, per-statement attribution,
      statically visible literals only). I also prototyped the counting change: ~5 lines, closes the
      residue, ZERO false positives on the real tree — free, but optional; the claim is the deliverable.
      ARCH-PURPOSE.
  - id: new
    severity: Important
    family: fix-at-consumer-not-source
    title: |
      BR-9's SKILL.md fix edits render output; /construct upgrade regenerates it from an intent transcript that never mentions the skill
    detail: |
      construct/adapted/ is generated. /construct promote step 4 is delete-then-copy (rm -rf
      construct/adapted/superpowers-*/ then copy staging), and /construct upgrade renders staging from
      construct/sources/<version>/ + construct/intents/superpowers.md, where "skills not mentioned in
      the intent → copy new source as-is". grep for receiving-code-review in that transcript returns
      nothing, so the GENERALIZE step, the ARCH-PURPOSE routing, and the family:-as-worklist paragraph
      are all scheduled for deletion by the substrate's own pipeline. The precedent is exact: a547179
      (#71), the last ARCH-registry change touching an adapted skill, appended a Conversation entry
      with Verify clauses in the same commit. TestReceivingCodeReviewSkillGeneralizes makes the loss
      loud rather than silent, which is why this is Important not Critical — but a red test with no
      intent record leaves the next reader reconstructing wording from three strings.Contains calls.
      This is the issue's own principle one level out: the deliverable landed at the compiled consumer
      instead of the source it derives from. ARCH-PURPOSE.
  - id: new
    severity: Minor
    family: refusal-line-width
    title: |
      The routing line is 120 chars — 129 at the FIX-THEN-SHIP indent, against ~83 for every neighbouring line
    detail: |
      close.go:1812 places it in a block whose widest existing line is 83. It hard-wraps mid-sentence in
      an 80/100-col terminal. Splitting after the colon would match the surrounding rhythm.
```
