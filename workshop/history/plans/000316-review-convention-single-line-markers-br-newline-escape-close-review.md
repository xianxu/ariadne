# Boundary Review — ariadne#316 (whole-issue close)

| field | value |
|-------|-------|
| issue | 316 — review-convention: single-line markers + <br> newline escape |
| repo | ariadne |
| issue file | workshop/issues/000316-review-convention-single-line-markers-br-newline-escape.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03fbeaa7eed29aecbb52205372cdfaac07c256f7..6a2d2647e01aac68ab629a38c1441be0f01102a0 |
| command | sdlc close --issue 316 |
| reviewer | claude |
| timestamp | 2026-10-09T21:22:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I've read the diff and the issue, and checked which other files restate the marker grammar. Writing up the verdict now.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This is a small, prose-only change to `construct/local/fix/review-convention.md` that delivers the Spec. §3 gains a "One line" rule with the `<br>` escape, §5 gains the note that `<br>` is decoded on resolve, and there is a 2026-10-09 Revisions entry naming parley.nvim#312. Both the stat and name-status recipes show a single modified file (+14/−1). Nothing here blocks shipping. Two fixes are cheap and should land first. First, the §5 decode rule as written also applies to reference text (`X`/`D`) on reject, which can corrupt original prose that already contains a literal `<br>`, such as markdown table cells. Second, the xx-fix SKILL.md restates the grammar and the resolution steps but was not updated, so agents writing and resolving markers through that skill never see the new rule.

1. **Strengths**
   - `review-convention.md:68-77` states the rule well: writers always emit single-line markers, and readers still tolerate legacy multi-line markers. That matches the Spec's "supersedes for writers; readers may still tolerate" wording and keeps documents already in flight working.
   - The rule lives in the shared grammar, not only in parley.nvim, which is the point of the issue (agents write markers too). The Revisions entry at `:115` says why.
   - Refusing `Alt+q` on a multi-line selection (`:76-77`) closes the one way the editor could itself create a spanning marker.

2. **Critical findings:** none.

3. **Important findings**
   - **The decode rule also hits reference text on reject** (`review-convention.md:88-89`). "Any `<br>` in the result becomes a line break" applies to every resolved result, including reject-to-`X` and reject-to-`D` (§5 rows `🤖<X>{R}`, `🤖~D~`, `🤖~D~{N}`, `🤖~D~[N]`) and accept-to-`X` (`🤖<X>[H]`, `🤖<X>[H]{R}`). Reference text is quoted from one line of the prior edition, so it can never hold an escaped line break. Any `<br>` in it is a literal one already in the prose; markdown tables use `<br>` inside cells all the time. Rejecting `🤖<a<br>b>{…}` in a table row would turn the original `<br>` into a newline and break the table. A robot `{R}` meant to keep a literal `<br>` hits the same problem, and there is no escape for it. Every affected site is in §5. Fix: limit decoding to commentary content (`{R}`/`{N}`/`[N]`) and restore `<X>`/`~D~` verbatim. Either say a literal `<br>` can't be written inside commentary, or note the table-cell case.
   - **The xx-fix skill restates the grammar and was not updated** (ARCH-PURPOSE shadow-sweep; `construct/local/fix/SKILL.md:58-93` and step 4 at `:124-140`). The skill carries its own BNF (`section ::= "[" TEXT "]" | "{" TEXT "}"`) and its own resolution steps ("substitute `D` with `N`… remove the marker"). Neither says markers are single-line, that agent replies must use `<br>`, or that `<br>` is decoded when applying `N`/`R`. The issue says agents (xx-fix, Claude Code, pair) are why the rule belongs in the shared grammar, and xx-fix is that writer. Other restatements I checked:
     - `AGENTS.base.md:15` (one-line summary, links to the full grammar) is acceptable.
     - `construct/datatype/target.md:78` links to the full grammar and is acceptable.
     - `docs/vision/2026-05-22-01-…md:93-96` is a historical vision doc and is fine to leave.
     - `scripts/docflow.sh:66-73` counts markers line by line and is unaffected.

     Fix: add a "single-line; newline written `<br>`" line to the SKILL Marker Format section, and add the `<br>` decode to step 4's apply paths (or have them cite §3/§5 explicitly).

4. **Minor findings**
   - The frontmatter still says `updated: 2026-07-22` (`review-convention.md:6`); bump it to 2026-10-09 to match the new Revisions entry.
   - "parley.nvim shows the chain as `🤖[…]` / `🤖{…}`… and opens it in a thread view" (`:72-74`) describes one editor's rendering inside a grammar doc whose own "What this is NOT" section (`:22`) says rendering is tool-specific. Consider wording it as an example ("e.g.") or moving it to the Revisions rationale.
   - The Done-when clause "parley's weaved copy picks it up on the next weave" can't be checked in this window. Confirm it at the next weave.

5. **Test coverage notes:** the change is prose only, and the grammar has no executable tests in this repo; the parsers live in parley.nvim and xx-fix. Still, this doc is a behavioral contract. Parley.nvim#312 should test round-trip decoding: `<br>` in `{R}` becomes a newline on accept, and `<X>` comes back unchanged on reject even when it contains a literal `<br>`.

6. **Architectural notes**
   - **ARCH-DRY: flag.** The grammar now has two hand-maintained sources (review-convention.md and the SKILL.md BNF and resolution steps), and they have already drifted in this change. See the second Important finding.
   - **ARCH-PURE: pass.** There is no code; it's a grammar spec.
   - **ARCH-PURPOSE: flag.** The canonical doc is updated, but the xx-fix writer instructions don't follow from it. Longer term, SKILL.md should keep only the actionability rule and link to §3/§5 for everything else.

7. **Plan revision recommendations:** add a Plan item for updating `construct/local/fix/SKILL.md` (Marker Format plus step-4 decoding). If the decode scope gets narrowed, add a Revisions note to the issue saying that `<br>` decoding applies to commentary blocks only.

```findings
findings:
  - id: new
    severity: Important
    family: escape-decode-scope
    title: |
      Section 5 decodes br in every resolved result, so it can corrupt verbatim X and D text on reject
    detail: |
      review-convention.md:88-89 says any br in the result becomes a line break. Reference text (X and D) is quoted verbatim from one line, so a br in it is a literal that is already in the prose, as in markdown table cells. That affects reject-to-X and reject-to-D for the <X>{R}, ~D~, ~D~{N} and ~D~[N] rows, and accept-to-X for <X>[H] and <X>[H]{R}. Rejecting one of these would break a table. Fix: decode only commentary content ({R}, {N}, [N]), restore references verbatim, and say how a literal br in commentary is handled.
  - id: new
    severity: Important
    family: single-source-consumer-not-derived
    title: |
      The xx-fix SKILL.md grammar and resolution steps leave out the single-line rule and br decoding
    detail: |
      construct/local/fix/SKILL.md:58-93 (BNF and examples) and step 4 at about :124-140 (substitute D with N, apply R) restate the grammar and never mention single-line markers, writing replies with br, or decoding br on apply, yet xx-fix is the agent writer that motivated the issue. Other restatements I checked are acceptable as they stand: AGENTS.base.md:15 and construct/datatype/target.md:78 link to the full grammar, the docs/vision doc is historical, and scripts/docflow.sh counts markers line by line and is unaffected.
  - id: new
    severity: Minor
    family: frontmatter-updated-stamp
    title: |
      The review-convention.md frontmatter still has updated 2026-07-22 after the 2026-10-09 revision
  - id: new
    severity: Minor
    family: grammar-doc-tool-specific-rendering
    title: |
      Section 3 describes parley.nvim rendering in a grammar doc that says rendering is tool-specific
    detail: |
      Lines 72-74 conflict with the What this is NOT section at line 22. Word it as an example or move it to the Revisions rationale.
```

---

## Re-review — 2026-10-09T21:25:32-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 316 — review-convention: single-line markers + <br> newline escape |
| repo | ariadne |
| issue file | workshop/issues/000316-review-convention-single-line-markers-br-newline-escape.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03fbeaa7eed29aecbb52205372cdfaac07c256f7..7467f2a0c11061506fa3be2564bda3adc1a3be27 |
| command | sdlc close --issue 316 |
| reviewer | claude |
| timestamp | 2026-10-09T21:25:32-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Inspection is done. I'm writing up the verdict: two prior findings have fixes and two are still open.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The main correctness problem from round 1 (BR-1) is fixed. `review-convention.md` §5 now decodes `<br>` only in turn text (`R`, `N`) and keeps anchor text (`X`, `D`) verbatim. xx-fix now also teaches the single-line writing rule (`SKILL.md:89-93`) and repeats the decode scope in its bulk-resolution section (`SKILL.md:406-408`). One gap from BR-2 is left. xx-fix's main per-marker resolution path is step 4 (`SKILL.md:137-146`). It substitutes `N` and applies `R`, but neither the step nor the "One line" paragraph tells the agent to decode `<br>` there. Bulk resolution is now right, but the everyday `🤖~D~[N]` and "apply R" path would paste a literal `<br>` into the prose. The two Minor findings from round 1 (BR-3 and BR-4) were not touched. None of this blocks the gate, and all of it is cheap to fix.

**1. Strengths**
- `review-convention.md:89-92`: the decode scope is stated once in §5, with the reason (tables legitimately contain `<br>`). This matches how the reference vs. commentary roles are defined in §2.
- `review-convention.md:68-78`: writers must emit single-line markers, and readers may still accept legacy multi-line markers. That avoids breaking documents already in flight.
- `SKILL.md:89-93`: the writing rule sits next to the BNF, which is where an agent writing markers will read it.
- The Revisions entry (`:118`) records why the rule lives in the grammar doc and not in one editor.

**2. Critical findings:** none.

**3. Important findings:** none new. BR-2 is re-opened in its disposition below.

**4. Minor findings:** none new. BR-3 and BR-4 are still open (see dispositions).

**5. Test coverage notes:** the window changes only Markdown contracts. There is no executable test surface in ariadne for these documents. Decode behavior will need tests in the consumers (the parley.nvim resolver). Neither document says how to write a literal `<br>` inside commentary, because there is no escape for it. That is acceptable for now, but it should be written down.

**6. Architectural notes**
- **ARCH-DRY: pass.** `SKILL.md` restates the rules but links back to the grammar for the full contract, and the restatements agree with it.
- **ARCH-PURE: pass.** There is no code in this window.
- **ARCH-PURPOSE: flag (BR-2 residual).** The consumer that motivated the issue (xx-fix) derives the rule in its writing and bulk paths, but not in its primary resolve path (step 4).

**7. Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      review-convention.md:89-92 now decodes only R and N and keeps X and D verbatim; SKILL.md:406-408 matches. Still missing: how a literal br inside commentary is written (no escape exists). Worth one sentence.
  - id: BR-2
    disposition: not-addressed
    note: |
      The writing rule (SKILL.md:89-93) and the bulk-resolution decode (:406-408) were added, but step 4 at SKILL.md:137-146 (substitute D with N, apply R) still never says to decode br, and the One line paragraph at :89 does not mention decoding either. Fix: state the decode-on-apply rule once in the :89 paragraph so step 4 and bulk resolution both inherit it.
  - id: BR-3
    disposition: not-addressed
    note: |
      review-convention.md:6 is still updated 2026-07-22, but the newest Revisions entry is 2026-10-09.
  - id: BR-4
    disposition: not-addressed
    note: |
      review-convention.md:72-75 still describes how parley.nvim renders markers in the grammar section, which conflicts with :22. Present it as an example or move it into the 2026-10-09 Revisions rationale.
```

---

## Re-review — 2026-10-09T21:27:14-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 316 — review-convention: single-line markers + <br> newline escape |
| repo | ariadne |
| issue file | workshop/issues/000316-review-convention-single-line-markers-br-newline-escape.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03fbeaa7eed29aecbb52205372cdfaac07c256f7..4d6ca449e9c552003574c7abea799004ef7984d8 |
| command | sdlc close --issue 316 |
| reviewer | claude |
| timestamp | 2026-10-09T21:27:14-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All three open prior findings are fixed in the pinned range. The one-line rule and the `<br>` decoding now appear in the grammar and in both xx-fix resolution paths, the frontmatter date is bumped, and §3 now gives the rendering rationale as an example. Nothing here is Critical. Two Important gaps remain and both are cheap to fix. First, the issue's Spec and Plan still describe the pre-BR-1 design: `<br>` encoded inside `<>`, decoded in all resolved text, and only one file touched. Second, nothing says how to write a literal `<br>` inside turn text, so accepting a robot edit to a markdown table row would split the row. Both should be fixed before close.

1. **Strengths**
   - `construct/local/fix/review-convention.md:87-91`: decoding is limited to text from `[]`/`{}` turns, and `X`/`D` anchors are restored verbatim. It gives the concrete reason (tables can contain `<br>`), which settles BR-1 properly.
   - `construct/local/fix/SKILL.md:89-94`, `:137-139`, `:409-411`: xx-fix now states the writer rule and the decode step on both of its resolution paths (per-marker step 4 and bulk §6). The agent writer that motivated the issue is covered.
   - The Revisions entry at `review-convention.md:117` gives the reason the rule belongs in the shared grammar: agents write markers too.
   - Legacy tolerance ("read them, but write single-line ones") keeps documents already in progress working.

2. **Critical findings:** none.

3. **Important findings**
   - **Issue Spec/Plan don't match what was delivered** (`workshop/issues/000316-…md` Spec and Plan sections).
     - The Spec still says a newline in a `[]`/`{}`/`<>` block is written `<br>`, and that §5 decodes "the resolved text". The code, after BR-1, excludes anchors from both encoding and decoding.
     - The Plan lists only `review-convention.md`, but the range also changes `construct/local/fix/SKILL.md` (BR-2 scope).
     - The `## Log` has an empty `### 2026-10-09` heading.
     - Fix: add a `## Revisions` entry (see §7) and fill in the Log.
   - **A literal `<br>` can't be written in turn text** (`review-convention.md:68-71`, `:87-91`; `SKILL.md:89-92`, `:137-139`, `:409-411`).
     - Decoding turns every `<br>` in `R`/`N` into a line break, and no escape exists for a literal one.
     - Failure case: `🤖<| a<br>b | c |>{| a<br>b | d |}` → accepting it breaks the table row across lines. The doc itself admits tables legitimately contain `<br>` (`:91`).
     - **This is the 2nd finding in family `escape-decode-scope`.** The rule covering both: *decoding runs only on text the encoder produced, and must be invertible — every token the decoder rewrites needs a spelling for its literal form.*
     - Fix it once in the grammar, either by defining a literal spelling (e.g. `&lt;br>` stays as is) or by stating the limitation explicitly. Then have both SKILL.md decode sites point to it rather than restating it.

4. **Minor findings**
   - `review-convention.md:75-77`: §3 (grammar) prescribes editor behaviour: "render it as broken" and "`Alt+q` on a selection spanning lines is refused".
     - **This is the 2nd finding in family `grammar-doc-tool-specific-rendering`.**
     - Rule: §3 states only the grammar; any editor or gesture behaviour goes in the tool-labelled §4/§5 sections. That means moving the `Alt+q` sentence into §4 and dropping "render it as broken".
   - `SKILL.md:92`: the line runs past the file's wrap width (reflow nit).

5. **Test coverage notes:** the range changes only Markdown contracts. No executable xx-fix test harness exists in this repo for marker resolution, and `scripts/docflow.sh` counts markers line by line, so it is unaffected. A fixture of a multi-line `R` and a table-row anchor through accept/reject would catch the `<br>` gap above in parley.nvim's implementation.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The three restatements of the decode scope list different letters (`R,N` / `N,R,H` / `R,Y,N`), but each matches its local table's names under the same stated rule. Make sure the fix for the `<br>` gap lands in the grammar first.
   - **ARCH-PURE: pass.** There is no code; it doesn't apply.
   - **ARCH-PURPOSE: pass.** The shadow sweep found consumers that either derive from the grammar (xx-fix, both paths) or only link to it (`AGENTS.base.md`, `construct/datatype/target.md`, `atlas/workflow/ledger-landscape.md`). parley.nvim takes it on its next weave, as the Done-when says.
   - **Docs gate:** nothing new a user runs or types, so no README change is due. The atlas only links to the convention, so no change is needed there either.

7. **Plan revision recommendations**
   - `## Revisions` — 2026-10-09: anchors (`<X>`, `~D~`) are never `<br>`-encoded and must quote a single line. Decoding applies only to `[]`/`{}` turn text, which supersedes the Spec's `<>` clause and "decode the resolved text" (BR-1). Scope grew to `construct/local/fix/SKILL.md` so xx-fix writes single-line markers and decodes on both resolution paths (BR-2). Tick a matching Plan item.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      SKILL.md:89-94 adds the one-line writer rule; step 4 (:137-139) and bulk resolution (:409-411) both decode br in turn text and keep X/D verbatim.
  - id: BR-3
    disposition: addressed
    note: |
      review-convention.md:6 now reads updated 2026-10-09.
  - id: BR-4
    disposition: addressed
    note: |
      Lines 72-74 now say an editor (example: thread view) with no parley-specific claim; the leftover Alt+q/render clause is raised separately as a Minor.
findings:
  - id: new
    severity: Important
    family: spec-plan-drift-unrecorded
    title: |
      Issue Spec/Plan still describe br encoding inside anchors and a single-file scope; no Revisions entry, empty Log
    detail: |
      Spec says a newline in []/{}/<> is written br and section 5 decodes the resolved text, but the delivered design excludes anchors (BR-1). The Plan lists only review-convention.md, yet SKILL.md was changed (BR-2). Add a Revisions entry plus a plan item and fill in the 2026-10-09 Log heading.
  - id: new
    severity: Important
    family: escape-decode-scope
    title: |
      No way to write a literal br in turn text; accepting a robot edit to a table row containing br splits the row
    detail: |
      2nd in family escape-decode-scope. Rule: decoding runs only on encoder-produced text and must be invertible, so every token the decoder rewrites needs a literal spelling. Instances: review-convention.md:68-71 and :87-91; SKILL.md:89-92, :137-139, :409-411. Define the literal spelling (or state the limitation) once in the grammar and have both SKILL.md decode sites point to it.
  - id: new
    severity: Minor
    family: grammar-doc-tool-specific-rendering
    title: |
      Grammar section 3 still prescribes editor behaviour (render legacy as broken; Alt+q multi-line selection refused)
    detail: |
      2nd in family. Rule: section 3 states only the grammar; editor and gesture behaviour belongs in the tool-labelled sections 4 and 5. Instance: review-convention.md:75-77. Move the Alt+q sentence into section 4 and drop render-as-broken.
```

---

## Re-review — 2026-10-09T21:29:42-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 316 — review-convention: single-line markers + <br> newline escape |
| repo | ariadne |
| issue file | workshop/issues/000316-review-convention-single-line-markers-br-newline-escape.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03fbeaa7eed29aecbb52205372cdfaac07c256f7..e68c507c9dbeaae98d6ba761b463b0147ddc8e29 |
| command | sdlc close --issue 316 |
| reviewer | claude |
| timestamp | 2026-10-09T21:29:42-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Inspection is done: I've read the whole diff, the issue file, and checked the other places that mention markers. Writing up the verdict now.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The grammar target now has the one-line rule, `<br>` and `\<br>` encoding limited to turn text, anchors kept verbatim, an updated stamp and a Revisions entry. The issue's Spec, Plan, Log and Revisions match the delivered design, so BR-5 is resolved. Two prior findings are still open:
- **BR-6 (Important):** the two decode sites in xx-fix `SKILL.md` (per-marker step 4 and bulk resolution) each restate the decode rule with only `<br>` and leave out `\<br>`. An agent following either step literally still splits a table row on accept, which is the bug BR-6 described.
- **BR-7 (Minor):** §3 still prescribes editor behaviour.

Both fixes are small edits to text.

1. **Strengths**
   - `review-convention.md:68-79`: the anchor/turn split (turns encoded, `<X>`/`~D~` verbatim) is stated clearly and is invertible now that `\<br>` exists.
   - `review-convention.md:90-94`: §5 now limits decoding to `R`/`N` and gives the reason (table `<br>` in anchors).
   - Shadow-sweep passes: `AGENTS.base.md:15` and `construct/datatype/target.md:78` point to the grammar file instead of restating it. `atlas/workflow/docflow.md` and `scripts/docflow.sh` only count markers and don't parse their contents, so the one-line rule doesn't affect them.
   - The issue Revisions entry records the scope growth (BR-1, BR-2, BR-6).

2. **Critical:** none.

3. **Important**
   - BR-6 is still open (`not-addressed` below). In `SKILL.md`, `:138-140` and `:410-412` say only "turn each `<br>` into a real line break" and don't mention `\<br>`. This is the 3rd instance of family `escape-decode-scope`. The general rule: the decode rule (which tokens, which slots) is defined once, in `SKILL.md`'s "One line" paragraph (`:89-95`, which points to the grammar's §3/§5). Each resolution site then says "decode turn text per **One line** above" and doesn't restate it. The slot lists also disagree today: step 4 says `N, R, H`, bulk says `R, Y, N`, and §5 says `R, N`. A single definition removes that drift too.

4. **Minor**
   - BR-7 is still open: `review-convention.md:76-78` still says "render it as broken" and "`Alt+q` on a selection spanning lines is refused". These belong in §4 (Alt+q) or should be dropped.
   - The issue Log has an empty duplicate `### 2026-10-09` heading after the round entries. Delete it.
   - `SKILL.md:90` is one overlong line compared with the wrapped prose around it.

5. **Test coverage:** this is a prose-and-prompt contract change with no runtime code in ariadne. The codec is in parley.nvim#312 and is outside this window. The "weaved copy picks it up" Done-when clause can't be checked from this repo.

6. **Architecture**
   - **ARCH-DRY: flag.** The decode rule is restated three times in `SKILL.md`, and the copies disagree (see BR-6).
   - **ARCH-PURE: pass.** Docs only; no IO logic.
   - **ARCH-PURPOSE: pass.** Every consumer either points to the grammar or carries the rule. The only gap is the incomplete restatements under BR-6.

7. **Plan revisions:** none needed beyond deleting the empty Log heading.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      Spec now says turn-only decode, anchors verbatim, SKILL.md in scope; Plan item for SKILL.md; Revisions entry; Log filled (stray empty heading noted Minor).
  - id: BR-6
    disposition: not-addressed
    note: |
      Grammar defines the literal \<br>, but SKILL.md step 4 (138-140) and bulk (410-412) restate decode as only <br>; following either literally still splits a table row on accept. Point both sites at the single One line definition and stop restating it (3rd in escape-decode-scope).
  - id: BR-7
    disposition: not-addressed
    note: |
      review-convention.md:76-78 still says render legacy as broken and Alt+q multi-line refused, inside the grammar section 3.
findings:
  - id: new
    severity: Minor
    family: spec-plan-drift-unrecorded
    title: |
      Issue Log carries an empty duplicate 2026-10-09 heading
    detail: |
      The second ### 2026-10-09 heading under Log has no body; delete it. It is the only instance in the window.
```
