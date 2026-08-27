# Lessons Learned

*(Record patterns of what went wrong and rules to prevent repeating them)*

## A syntactic guard cannot back an absolute claim — bound the claim, not the mechanism

**Pattern (#203, four boundary rounds of one family):** each round the guard was
widened to whatever shape the last finding named — per-literal → whole
expression; two hardcoded emitters → any call; per-function → per-statement;
membership → counting. Every widening was correct and every one left a finer
residue, because a static scan over source text is an over-approximation, not a
proof. The rounds stopped only when the reviewer said to fix the *claim*: state
what the guard approximates (statically visible literals, `+` chains, statement
counting that cannot tell which message a reference is joined to) and say it
raises the cost of the defect without making it impossible.

**Rule:** when a fitness function backs a strong claim ("X cannot ship"), either
prove the claim or weaken it to what the mechanism delivers. A residue you can
name is a bounded claim; a residue you keep chasing is an unbounded one. Ask
"what is the next shape past this fix?" — if you can name it, you are widening,
not converging (`ARCH-PURPOSE`).

## Distinguish what a guard COMPUTES from what it takes on faith

**Pattern (#203 BR-9, BR-13):** the scans compute the *sites* within each
surface, so a new gate emission cannot be missed. The *set of surfaces* was
hand-declared and unverified — and under that gap
`superpowers-receiving-code-review` sat instructing agents to implement findings
"one item at a time", the exact behavior the issue existed to stop, in the skill
invoked at the moment findings arrive. Its sibling
`superpowers-requesting-code-review` then survived one round longer. Both escaped
every scan by saying "feedback"/"issues", never "findings".

**Rule:** state at the guard which half is computed and which is a hand-declared
list nothing verifies. An unstated hand-maintained set reads as coverage.

## An edit to construct/adapted/ is an edit to render output

**Pattern (#203 BR-11):** a fix to an adapted skill was scheduled for silent
deletion by `/construct upgrade`, which re-renders from `construct/sources/` plus
`construct/intents/<pkg>.md`, copying any skill the intent does not mention
straight from source. **Rule:** every change under `construct/adapted/` needs a
Conversation entry with Verify clauses in the same commit (precedent: a547179,
#71). The deliverable belongs at the source, not the compiled consumer.

## A span-based text edit is only as safe as the boundary you assume

**Pattern (#203):** replacing `s[index("## The class"):index("## Estimate")]`
silently deleted the whole `## Done when` section that sat between them — a
structural section the gates require, gone for one commit. **Rule:** before
replacing a span between two anchors, list the headings inside it. Prefer
replacing an exact known block over a range between landmarks.

## A residency/path change includes search recipes, not just prose path mentions

**Pattern (#185 close review):** The roadmap residency lift updated the main
datatype path, atlas row, and constitution, but left two search recipes with the
old effective scope: roadmap's "proto-company view" listed only the current repo,
and product's "all artifacts linking to a product" searched only `data/`, missing
roadmaps/projects now under `workshop/projects/`. A raw stale-string sweep caught
literal old paths, but not a query whose words stayed current while its search
scope was wrong.

**Rule:** When moving a datatype's residency or canonical home, audit every
`Search recipes` block as executable contract. For each recipe, ask whether its
filesystem roots still cover the stated semantic set; update labels and command
roots together. A clean `rg old/path` sweep is necessary but insufficient —
recipes can be stale by omission, not just by containing the old bytes
(`ARCH-PURPOSE`, `ARCH-DRY`).

**Origin:** #185 close review (FIX-THEN-SHIP). The fix split roadmap's
current-repo month listing from the sibling-repo proto-company view and widened
product link searches to include `workshop/projects/`.

## A prose policy is an integration contract when its test reads the repository; pin semantics and every derived consumer

**Pattern (#167 close review):** The plan labeled `SessionContinuityPolicy` PURE,
but its only regression test read `AGENTS.base.md` and the continuation prototype
from disk. The label contradicted the actual boundary: this was a repository
contract consumed by harness entry files, not an IO-free transformation. The
same test checked only that `"60%"` appeared, so reversing the requirement from
“more than 60%” to “less than 60%” still passed. Generic weave tests proved the
fan-out mechanism in isolation, but the feature test never proved this policy's
source was exported into all three consumers.

**Rule:** Classify an entity by the boundary its behavior test crosses, not by
whether its source happens to be prose. A test that reads live repository files
is INTEGRATION; call something PURE only when its behavior is exercised entirely
from in-memory inputs. For declarative policy contracts, pin the semantic
predicate (direction + boundary + action), not a bag of tokens, and drive the
actual source through its real composition seam to assert every derived consumer.
Prove the guard with a wrong-direction mutant and a broken-export mutant before
trusting green. Scope prose assertions to the owning section so duplicate words
elsewhere cannot mask a deletion. When the source is structured (a manifest,
frontmatter, JSON), parse its semantic records instead of substring-matching raw
text — a commented-out row contains the same bytes but has no behavior. When a
consumer registry already exists, derive an “every consumer” sweep from it rather
than copying today's members into the test; otherwise future consumers silently
escape the contract. Assert the complete scoped contract in each derived consumer,
not just identifying sentinels, when partial propagation would violate Done-when.
For the source itself, enumerate every behavioral predicate in the Spec—including
conditions and ordering—not merely the nouns or actions it mentions. Where the
contract is relational, assert the bound clause or relative positions; separate
presence checks do not prove causality, sequence, or the absence of negation.
(`ARCH-PURE`, `ARCH-PURPOSE`.)

**Origin:** #167 whole-issue close review (REWORK). The remediation moved the
guard from `cmd/datatype` to an end-to-end `cmd/weave` fixture, pinned “more than
60% full” plus the checkpoint boundary, checked the live base-manifest export,
and asserted `CLAUDE.md`, `AGENTS.md`, and `GEMINI.md` all derive the policy.
The follow-up FIX-THEN-SHIP review hardened it further with section scoping and
typed manifest parsing after moved-marker and commented-export mutants exposed
the raw-text false positives.

## A changed surface has shadow docs and execution records, not just the main atlas page

**Pattern (#97 close review):** The implementation updated `atlas/workflow/weave.md`
for topological settings merge, but two other atlas pages still described
settings as only `settings.ariadne.json + settings.local.json`. The code and
primary atlas page were right; the shadow documentation was stale. The same
review found the durable implementation plan still had every detailed checkbox
unchecked even though the issue checklist was complete.

**Rule:** When changing a named surface or convention, run a shadow-doc sweep for
the old phrase and update every live explanatory copy, not just the page you
remember editing. Also update the durable plan's execution state before close:
issue checkboxes, detailed plan checkboxes, and any generated review sidecars
should tell the same story. Grep for the old model terms before committing
(`settings.ariadne.json + settings.local.json`, `MergeSettings{Source}`, etc.),
then rerun `git diff --check`.

**Origin:** #97 close review (FIX-THEN-SHIP). The code review found no behavior
blockers, but caught stale atlas shadows and unchecked durable-plan steps before
the issue crossed the boundary.

## Generated review sidecars must be bounded, or they become the next review's input bug

**Pattern (#166):** `sdlc close` writes a durable review sidecar, and the next close review diffs that sidecar too. Capturing the full raw reviewer transcript, including the prompt and diff, made the sidecar enormous, introduced whitespace-check failures from embedded patches, and eventually made a later review dispatch fail with `argument list too long`. The evidence file became active input to the gate it was supposed to document.

**Rule:** Generated review artifacts must be bounded and normalized before they enter the reviewed diff. Persist the machine-useful facts (verdict, window, findings, verification commands, resolution), not the full prompt/diff transcript. If a sidecar must carry raw output, keep it out of the code-reviewed diff or teach the generator to strip/escape whitespace-sensitive embedded patches. After any generated sidecar write, run `git diff --check` before committing it.

**Pattern (#201):** A detailed plan existed, but its filename did not use the
issue file's exact stem. `sdlc change-code` therefore discovered only the short
issue checklist and correctly rejected it as non-executable.

**Rule:** Name every separate implementation plan
`workshop/plans/<exact-issue-stem>-plan.md`, then confirm the gate discovered it
before interpreting plan-quality findings. A good plan at a non-canonical path
is operationally the same as no plan.

**Pattern (#201 close review):** A function was called PURE because it accepted
an injected writer, even though writing and reading launch-error environment
context are still IO. An external conformance check asserted non-empty output,
which proved liveness but not the channel contract it existed to defend.

**Rule:** Classify entities by effects, not injectability: an injected IO seam is
INTEGRATION. A conformance assertion must pin the dependency behavior production
relies on (here, exact requested semantic output on stdout), never a weaker proxy
such as non-empty output. Do not forbid prompt text on a diagnostic channel when
the harness legitimately echoes its input there.

**Origin:** #166 close-review loop. The fix for this issue manually condensed the sidecar after each generated rewrite so `git diff --check` and later boundary-review dispatches stayed usable.

## A deferred cleanup does not run through `os.Exit` — command wrappers must cover hard exits and init races

**Pattern (#132):** A root-level Cobra wrapper acquired `.git/sdlc.lock` and used `defer release()` around the command `RunE`. That looked correct for returned errors, but most `sdlc` guard refusals call `die()`, and `die()` calls `os.Exit(1)`. `os.Exit` skips defers, so routine refusals would leave `.git/sdlc.lock` behind and wedge the next mutating command. The same review found a second liveness race: `mkdir .git/sdlc.lock` succeeds before `meta.json` is written, so a waiter can see the directory without metadata and must treat that as "holder initializing," not as a corrupt lock to remove.

**Rule:** When adding a process-wide wrapper around command bodies, enumerate every exit path, not just returned errors. If any path uses `os.Exit`, register cleanup somewhere that path drains explicitly before exit; a `defer` in the caller is not enough. For filesystem locks created as a directory plus metadata file, make waiters tolerate the mkdir-before-metadata window with a short grace period. Auto-reclaim only facts you can prove safe (same host + missing pid); cross-host or over-age uncertainty should fail with recovery guidance.

**Origin:** #132 boundary review (REWORK). The fix added a die-cleanup registry, idempotent lock release, confirmed-dead same-host reclaim, metadata-initialization polling, and real concurrent `Acquire` coverage.

## A pure helper unit-tested in isolation can be silently un-wired from its caller

**Pattern:** #72 extracted a pure `planPointer(issue) string` and printed it from the thin `runStartPlan` IO seam (`cinfo(stdout, planPointer(issue))`). TDD gave it a colocated unit test (`TestPlanPointer`) pinning the *wording* — skill name, `workshop/plans/` path, the `~/.claude/plans` demotion. All green. But nothing asserted the seam *actually calls* the helper: delete the `cinfo` line, or reorder it, or let a refactor drop it, and `TestPlanPointer` stays green while the feature ships broken. The boundary-review judge (fresh eyes) caught it; the author's suite didn't. I'd verified it *manually* (ran `start-plan`, saw the line) — so the gap was specifically the **automated regression**, not the behavior.

**Rule:** When TDD produces a pure entity consumed by a thin IO/print seam (the ARCH-PURE shape), the unit test on the entity is necessary but **not sufficient** — add one *integration assertion on the seam's output* that the entity's contribution is present (here: extend the existing `runStartPlan(&b, 75)` test with `"superpowers-writing-plans"` + `"workshop/plans/000075-"`). The unit test pins *what the helper says*; the integration assertion pins *that the caller says it*. Without the second, "pure helper exists and is correct" and "pure helper is wired in" are two independent facts and only the first is guarded. Cheap (one line appended to a test that already renders the seam) and it closes exactly the drop/reorder bug class. Distinct from the #44 "IO needs a live run" lesson: this isn't external IO — it's the wiring between a pure function and its single in-process caller, invisible because *both* the unit test and a helper-never-called build are green.

**Origin:** #72, boundary review (FIX-THEN-SHIP → fixed before crossing). The mandatory fresh-context review (binary-dispatched at `sdlc close`) found the wiring gap the author's own green suite hid — a concrete instance of why the review boundary is owned by fresh eyes, not the author (AGENTS.md §3).

## Skill design: enumeration vs. judgment

**Pattern:** A skill's behavior was specified by enumerating cases — a hardcoded list of nouns mapped to outcomes, plus a hardcoded list of "examples that DO/DO NOT trigger." Every new case required editing the skill, and the vocabulary tail (synonyms, unusual phrasings, descriptive statements that incidentally contain trigger nouns) was never reachable by enumeration.

**Rule:** When a skill's behavior is best described as *"use judgment"*, don't make it enumerate — express the principle and let the LLM apply it. The skill should describe *the question being asked* (e.g., "is this a fact, a question, or a request?") and *the discriminator* (e.g., "is the substance already present, or being requested generatively?"), not the surface forms that pass/fail. Concrete examples can serve as priming (a small, illustrative set), but they should not be the matching mechanism.

**Test for whether a list belongs in a skill:** ask *"would the skill's behavior be wrong if this list were missing, or just less ergonomic?"* If wrong → the skill has too much enumeration; the case it covers should be derivable from a principle stated elsewhere in the skill. If less ergonomic → the list is fine as priming, keep it short.

**Origin:** issue #25 (dispatcher: judgment-based triggers, replace enumeration). The `xx-datatype` skill's original noun→type mapping table was the case; it broke the atlas's own claim that "new types are pure data — adding one does not require a skill change."

## "Direct-only" handoffs hide transitivity bugs behind a depth assumption

**Pattern:** `bootstrap.sh` cloned only *direct* peers, then `exec make bootstrap` to let the recursive cloner take over. This silently assumed the handoff target (the Makefile, reached through a symlink chain) needed only the direct peer present. True for 2-deep chains, false for 3-deep — and *nothing in the codebase was 3-deep yet*, so the bug was invisible. The recursive cascade that would have fixed it could never start, because starting it required the very substrate it was meant to fetch.

**Rule:** When step A does "just enough" to hand off to step B, write down the invariant A must establish for B to run, then check it holds at the *deepest* input, not the common one. A "clone the direct peer" shortcut is really "ensure B's entrypoint resolves" — make the code do the actual requirement (clone *transitively* until the entrypoint resolves), not the proxy that happens to coincide with it at depth 2.

**Two corollaries that recurred here:**
- A file that runs *before its own substrate exists* (seed-delivered, zero-substrate) cannot share code via symlink — it must inline. Don't fight this; keep the inline copy and lock it to the canonical implementation with a **drift test** (run both on a fixture, assert equal output). One grammar, two call sites, one test.
- `local a="$1" b="$ROOT/$a/..."` on a **single line** can read `$a` as unbound under `set -u` — split positional captures from derived locals onto separate `local` statements.

**Origin:** issue #45 (bootstrap transitive clone walk). Surfaced while designing #44; the brain→nous→ariadne symlink chain was the case that exposed the depth-2 assumption.

## Integration bugs hide where pure tests can't reach — sandbox/IO needs a live run

**Pattern:** issue #44 (openshell sandbox go.mod sync) had thorough hermetic tests for the *pure* logic (`compute_sync_set` rw/ro classification, peer-walk membership) — all green. Yet the first live `make sandbox-build` exposed **three** bugs none of those tests could see: (1) a self-referential `~/workspace → /sandbox/workspace` symlink because `$HOME` is `/sandbox` in the base image (name == target); (2) an `ssh` call I added *inside* a `while read … done < <(…)` loop consumed the loop's stdin and truncated it to the first peer; (3) mutagen won't create a sync-root's missing *parent* dir, so `/sandbox/workspace/<name>` synced 0 files until `/sandbox/workspace` was pre-`mkdir`ed.

**Rule:** for any feature whose substance is IO against an external process (mutagen, ssh, docker, a container's filesystem/`$HOME`), unit tests of the pure decision logic are necessary but **not sufficient** — you must run it against the real thing once before claiming done (AGENTS.md §5). Split the work so the pure core *is* unit-tested (add a `*_LIB_ONLY` source hook to call internal functions without dispatching), then do one live E2E pass; budget for it to find bugs, because it will. Specific tripwires to remember:
- **Don't assume `$HOME`.** Check it (here it was `/sandbox`, not `/home/sandbox`); a symlink whose name equals its resolved target is always a loop. Guard with a string compare, not `-ef` (the inode test falsely falls through when the target doesn't exist yet).
- **`ssh`/`mutagen`/any stdin-reader inside a `while read` loop eats the loop's input.** Read on a dedicated fd (`done 3< <(…)`, `read … <&3`) and pass `ssh -n`.
- **mutagen creates the sync-root leaf but not missing parents** — pre-`mkdir -p` the parent.

**Origin:** issue #44. The bugs were found in three successive live `make sandbox-build` runs against a real `pair` sandbox; the pure suite (6/6) stayed green throughout — it simply couldn't observe them.

## N parallel walkers over one grammar drift apart silently — make the Nth match the others, with a test

**Pattern:** the `replace => ../<peer>` grammar in `construct/go.mod` is read by four independent walkers (setup.sh `discover_ancestors`, bootstrap-peers.sh, list-peers.sh, bootstrap.sh). The convention is "walk BOTH the root go.mod and `construct/go.mod` per node" (substrate ancestor lives in construct, not root). Three walkers honored it; `discover_ancestors` quietly walked only the root. It "worked" for years because the only failing shape — a depth-2 derivative whose depth-2 ancestor is declared in the depth-1's `construct/go.mod` — didn't exist until brain→nous→ariadne. The depth-1 case was masked by an unrelated fallback (Source-3 `ARIADNE_DIR`). The atlas even *documented* the correct behavior — so the bug was a silent divergence from stated intent, invisible because no input exercised it.

**Rule:** when the same grammar/format is parsed in more than one place, treat them as one logical parser with N call sites — not N parsers. (a) Audit ALL sites when you touch one (`grep` the format string / the path being read); the one you didn't write is the one that drifted. (b) The divergence won't show until an input hits the gap, so add a **fixture-based test that pins the sites together** (here: a hermetic chain asserting depth-2 discovery; for the inline-copy case in #45, a drift test asserting equal output). (c) When the atlas says "all four do X" but one doesn't, that's not documentation rot to fix in prose — it's a latent bug; make the code true.

**Corollary — test seams for apply-style scripts:** a function that's normally followed by a destructive apply (setup.sh mutates the target) isn't testable end-to-end without side effects. Add a narrow env-gated early-exit (`SETUP_DISCOVER_ONLY=1` prints the computed set and exits) so the *decision* is assertable hermetically while the *apply* stays untested-by-that-test. Mirrors #45's `BOOTSTRAP_DRY_RUN`/`BOOTSTRAP_CLONE_ONLY`.

**Origin:** issue #50. Surfaced pushing #49's `clone-data-deps.sh` down to brain — it never arrived because `discover_ancestors` stopped at nous and never read `nous/construct/go.mod` to find ariadne.

## Agent-invoked CLI verbs must run headless and gate on durable state, not local convenience

**Pattern:** `sdlc merge` broke two ways while shipping #56, both invisible to a human at a terminal and only biting the headless/agent path. (1) Its confirmation prompts called `scanner.Scan()` on `os.Stdin` with no tty check — an agent/background invocation has no tty, so the scan *blocked forever* (the observed "stall"). (2) Its "is the branch pushed?" gate keyed off `@{u}` — the *local upstream-tracking config* — which a plain `git push` (no `-u`) never sets, and which a sandbox that blocks `.git/config` writes silently drops. So `merge` refused a branch that was genuinely pushed with an open PR.

**Rule:** A verb an agent invokes must (a) **never block on stdin** — tty-guard every interactive prompt and, when not a tty, fail fast with a next-action (`--yes`, or a sentinel like `change-code`'s `ASK_<TOPIC>`), never a bare blocking read; and (b) **gate on the most durable signal, not a derived local convenience** — `origin/<branch>` (the remote-tracking ref, updated by any push) carries the same truth as `@{u}` (tracking config) but survives the cases where the config is absent. When choosing what a guard reads, ask "what's the *fact* I need, and what's the flakiest proxy for it I might be keying on?"

**Origin:** #56 session, `sdlc merge` fixes. `change-code` already had the tty pattern right (`isTTY` → sentinel); `merge` predated it. Found by the tool hanging in a non-tty agent run, then refusing a pushed branch because the sandbox had eaten its `push -u` config write.

## Matching convention-authored free text: the canonical form is one of many natural ones

**Pattern:** Two matchers in `sdlc` silently failed on natural-but-non-canonical phrasing. (1) The milestone-verdict guard anchored commit subjects on `^#<N> Mx:` — milestone immediately followed by a colon — so the natural `#56 M1 close: …` (milestone + words before the colon) didn't match, and `sdlc close` claimed three reviewed milestones "lacked Review-Verdict trailers" that were right there. (2) The milestone-review verdict parser only read the first non-empty line, so it recorded "unknown" when the LLM judge led with a markdown title (M1) and again when it narrated investigation prose before the verdict (M3) — twice, two different shapes.

**Rule:** When parsing text a human or LLM authors *by convention* (commit subjects, review verdicts, status lines), the documented canonical form is one of many forms real authors produce. Don't anchor on a literal token (`Mx:`); anchor on a boundary (`Mx[: ]`, still rejecting `M10`) and, for the harder cases, add a **high-precision fallback** that survives narration (a confidence-qualified `<VERDICT> (confidence: …)` line works where "verdict on line 1" doesn't). **Test the non-canonical-but-natural variants explicitly** — the canonical form always passes; the bug lives in the phrasings you didn't enumerate. (A strict matcher is a hidden enumeration of *one* accepted form — see the enumeration-vs-judgment lesson above.)

**Origin:** #56 session, `sdlc close` + `sdlc milestone-close`. Both reported a verdict of "unknown"/"missing" for work demonstrably reviewed; the fix was boundary-tolerant matching + a fallback, each pinned with a regression test for the exact failing shape.

## A hand-maintained copy of generated data drifts — render from the source

**Pattern:** `sdlc --help` listed every verb *twice*: a hand-written `SUBCOMMAND` block in `root.md` and cobra's auto-generated `Available Commands`. The hand-list was the drift-prone copy — it still advertised flat `set-status`/`fetch` after #56 made them hidden, and an atlas index still said "11 verbs" when the visible count was 10. The generated list could not drift (it renders from the live registry and auto-omits hidden commands); the hand copy needed a human to remember.

**Rule:** If a tool can render a list/count from its own registry, **don't also hand-maintain a copy** — render from the source (here: `cobra.EnableCommandSorting=false` + workflow-ordered registration gave the auto-list the ordering the hand-list existed to provide). If a curated copy is genuinely required, pin it to the source with a test, or it *will* go stale at the next change. Same family as "N parallel walkers drift," one level up: generated-output vs hand-mirror.

**Tripwire — compile-check builds drop a binary at the repo root.** `go build ./cmd/sdlc/` (run for a quick compile-check) emits `./sdlc` in the cwd, *not* the gitignored `bin/` — and `git add -A` then swept it into a commit. Two fixes: (a) compile-check with `go build -o /dev/null ./cmd/sdlc/` (or `go vet`) so no artifact lands; (b) gitignore build outputs at *every* path they can land (`/sdlc`, not just `bin/`), and scan `git status` for untracked binaries before a broad add.

**Origin:** #56 session, the `sdlc --help` consolidation + the stray-binary amend.

## Iterating files via `ls` in `$()` word-splits — glob directly

**Pattern:** #59's vm-hooks run-parts loop iterated `for name in $(cd "$DIR" && LC_ALL=C ls -1 ./*.sh)`. The unquoted command substitution word-splits on whitespace, so a hook named `15 setup.sh` became two tokens (`15`, `setup.sh`), each `bash`-run as a nonexistent path (rc=127) — the real hook silently never ran, only warned. The documented `NN-` no-space convention masked it, so it shipped and a fresh-eyes review (not the author) caught it.

**Rule:** To iterate files in shell, **glob directly** (`for f in "$DIR"/*.sh`), never `ls`/`find` inside `$()` — a command substitution always word-splits (and globs) its output. Under `set -euo pipefail` on macOS **bash 3.2**, pair the glob with `shopt -s nullglob` so an empty match is a clean no-op (and to dodge the `"${arr[@]}"`-on-empty-array `set -u` abort that bites 3.2 but not 4.4+). For arbitrary filenames, the fully-safe form is a NUL-delimited process-substitution: `while IFS= read -r -d '' f; do …; done < <(LC_ALL=C; shopt -s nullglob; for g in "$DIR"/*.sh; do printf '%s\0' "$g"; done)` — whitespace/newline-proof, order pinned, locale scoped to the subshell. **Test the spaced-filename case explicitly**; the convention-compliant names always pass.

**Origin:** #59 session, post-milestone review of the tart vm-hooks loop. Verified the fix under `/bin/bash 3.2.57` (the actual VM interpreter), not just the host shell — bash 3.2's `set -u`/empty-array and `shopt` behaviors differ from modern bash and from zsh.

## Migrating a peer repo: check its branch/cleanliness first; never `git clean -fd` it

**Pattern:** Rolling out #60 M4 to a derivative (nous), I ran `make refresh` + `git rm construct/go.mod` + commit — but nous was on its own feature branch (`000036-...`) mid-work, so my base-layer commit polluted *its* feature branch. Worse, reverting with `git reset --hard HEAD^ && git clean -fd` removed two empty untracked dirs (`workshop/notes/`, `workshop/vision/`) that weren't my artifacts — `git clean -fd` deletes ALL untracked, not just what I created. (No tracked content was lost; verified + recreated. But it was reckless on a repo I don't own the state of.)

**Rule:** A base-layer change that lands as a *commit in a peer repo* is not a mechanical loop. Before touching peer X: (a) check `git -C X branch --show-current` — if it's not the integration branch (main), STOP; committing base-layer work onto someone's feature branch is wrong. (b) check `git -C X status --porcelain` is empty — never refresh/migrate a dirty peer. (c) To undo your own artifacts, remove them **by name** (`rm construct/deps construct/dev-aliases.sh …`; `git restore <tracked>`), NEVER `git clean -fd` — that's a blunt instrument that eats the operator's untracked files too. (d) A "try it out" verification (does the migration *work*) is separable from the *commit* — you can prove the mechanism in a throwaway/verify pass without committing into the peer at all.

**Corollary — the fleet has heterogeneous git state.** "Refresh + delete + commit ×13" assumes every derivative is clean-on-main; in reality some are mid-feature-work. A cross-repo base-layer migration must survey each repo's branch/cleanliness and skip/defer the ones that aren't ready, rather than assuming a uniform loop.

**Origin:** #60 M4, the nous canary. The migration mechanism itself worked perfectly (construct/deps-only nous: list-peers/bootstrap/sdlc-build all identical to dual-read) — the failure was treating the per-repo *commit* as blind automation.

## A migration's "nothing to migrate" precondition must be checked against the real fleet — with a portable check

**Pattern:** #60 M5 retired the legacy `construct/data-deps` reader on the premise "no repo has a populated data-deps, so nothing to fold." The premise was *false* — `brain` had a live `you-decide` content mount in `construct/data-deps` — and the survey that "confirmed" it was empty used `grep -qvE '^\s*(#|$)'`. **BSD/macOS grep (ERE) doesn't support `\s`** (a GNU extension), so the pattern didn't match comment/blank lines as intended and the check reported a false negative. M5 would have made brain's mount non-reproducible (the tracked symlink survives, but a fresh clone never re-clones the sibling). Caught by fresh-eyes review, not the (green) test suite — the migrated test even *asserted* the legacy file was ignored, green-lighting the regression.

**Rule:** (a) Before retiring/deleting a mechanism, enumerate its *actual live consumers across the fleet* and migrate each — don't assert "nothing uses it" from a single grep; spot-check the repos you expect to use it (here: brain, the whole motivating case for data-deps). (b) **Use POSIX character classes, not GNU `\s`/`\d`, in shell greps** — `[[:space:]]`, `[[:blank:]]` — because the same script runs under BSD grep on macOS and GNU grep on Linux. A `\s` that silently matches nothing turns a safety check into a rubber stamp. (c) A test that asserts the NEW behavior ("legacy file ignored") does not verify the DATA migration happened — keep those separate in your head.

**Origin:** #60 M5. The retirement code was correct; the rollout missed brain's row because the precondition check was both unportable (`\s` under BSD grep) and under-scoped (didn't spot-check the known consumer).

## A guard test must be proven to have teeth — mutation-check it

**Pattern:** #63 added an e2e test that `sdlc merge` refuses *before* the irreversible `gh pr merge` when a pre-merge judge dirties the tree (the #62 M1 9b guard). A test that asserts "merge refused" can pass for the wrong reason — refused at an *earlier* gate, never reached 9b at all — and still look green. To prove the test actually exercises 9b, I temporarily neutered the guard (`redirty \!= "" && false`) and confirmed the test went **red** ("expected merge to refuse"), then restored it. Without that step, the test could have been a rubber stamp that survives the guard's deletion.

**Rule:** When a test exists to defend a specific guard/branch, **mutation-check it once**: disable the guard, confirm the test fails, restore. A test that stays green when the code it guards is removed defends nothing. Cheap to do (one throwaway edit — use `$TMPDIR` for the backup under sandbox, restore immediately), and it's the difference between "the test passes" and "the test would catch the regression." Pair with assertions that pin the *specific* failure (e.g. a 9b-unique message substring + `PRMerge` call-count == 0), so a refusal at the wrong gate can't masquerade as success.

**Corollary — testing a verb that `os.Exit`s or shells out directly.** `runMerge` resisted in-process testing because `die()` → `os.Exit(1)` kills the test and `detectRepo`/`RepoTopLevel` call `exec.Command("git")` directly. The unlock was a trio of minimal `func`→`var` seams (`die`, `detectRepo`, `runPreflightJudgesFn`) — callers unchanged — plus a real throwaway repo (`git init` + local **bare** origin) so switch/pull/archive/branch-delete run for real instead of being mocked. `expectDie` swaps `die` for `panic(&dieSignal)`+recover, preserving halt semantics in-process. Prefer a real temp repo over stubbing a dozen git calls when the cleanup *is* what you're testing. Note: process-global var swaps + `os.Chdir` forbid `t.Parallel()`; the panic-based `die` runs deferred funcs that prod's `os.Exit` would not (keep refusal paths defer-free).

**Origin:** #63 M1 (e2e harness for `runMerge`), milestone-review SHIP. The reusable kit (`expectDie`/`tempRepo`/`swapMergeDeps`) is meant for any future `run*` verb's refusal-path test.

## Dogfooding a tool on its own meta-issue catches what unit tests miss

**Pattern:** #66 fixed `sdlc close`'s `insertLogLine` to file a dated log line under its matching `### <date>` day header. Unit tests (5, exact-string) all passed. But the *first real close* of #66 misfiled the line into the issue's own `## Problem` code-block example — because `insertLogLine` matched the **first** `## Log` / `### <date>` in the body, and #66, being a meta-issue *about the log format*, literally quotes those headers inside a fenced block. The test bodies never reproduced that self-reference, so green tests + a broken close. The fix: anchor on the **last** `## Log` (the real section is conventionally final). Both the old and new code shared the first-match weakness; only running the tool on its own self-referential issue surfaced it.

**Rule:** When a tool parses document *structure* (markdown headers, sections, fences), a document *about* that structure will contain the structure literally in prose/examples — and naive first-match parsing misfires on exactly those meta-documents. (a) **Dogfood structure-parsing tools on a meta-input** that quotes the structure (a unit test with the target header inside a ``` fence earlier in the body is the cheap version). (b) Anchor to the *conventional position* (here: the LAST `## Log`, since the real section is the final one) rather than the first match, or skip fenced code blocks. (c) Green exact-string unit tests prove the cases you imagined; a live dogfood proves the case you didn't. For a tool that mutates its own artifacts (issue files, logs), closing its own issue *is* the integration test — watch where the bytes actually land.

**Origin:** #66, found by dogfooding the fix while closing #66 itself. The self-referential Problem section (a `## Log`/`### <date>` example in a fenced block) is precisely the input the unit tests omitted.

## A tool that returns a silent "0/empty" indistinguishable from a real answer is a footgun

**Pattern:** `active-time-v3.py` computes an issue's actual-hours from session transcripts passed via `--dir`. Run without `--dir` (the easy `--git-repo . --issue N` form), it found no events and **exited 0 with "no events in window"** — a result *identical* to a legitimate "no activity." So across a whole session I (and the operator, who filed #68) ran it the easy way, got 0, concluded "v3 is broken," and recorded ~7 **fabricated** `actual_hours` via judgment — silently corrupting the velocity-calibration loop the gate exists to feed. The algorithm was fine; the inputs were wrong, and nothing said so. The fix: empty `--dir` → **exit 2** ("no transcript source — misinvocation"); commits-but-0-events → **exit 3** ("TELEMETRY UNAVAILABLE, don't read 0 as measured"). The genuinely-empty case still exits 0.

**Rule:** When a measurement/derivation tool can produce a "zero/empty" result for two very different reasons — *(a) genuinely nothing* vs *(b) you fed me the wrong inputs* — it **must distinguish them with distinct exit codes / loud messages**, never collapse both to a silent success. A footgun isn't "it gave the wrong answer"; it's "it gave a wrong answer that looks exactly like a right one." Corollary: if the *correct* invocation is a 6-line command with non-obvious required inputs (here: which `~/.claude/projects/<cwd>` transcript dirs — work scatters across repo + brain + worktree cwds), **prose telling a human to run it will be shortcut or skipped** — lift it into the tool (`sdlc actual` runs v3 with the right dirs auto-selected). Prose is a footgun; a verb is not.

**Origin:** #68. Diagnosed by running v3 *correctly* (with `--dir`) on a known issue — nous#14 came back 7.79h vs 8.2h recorded (~5%), proving the algorithm sound. Dir-selection (brain + the issue's repo, NOT all folders — an unrelated concurrently-edited repo inflated it +4.3h) was the whole bug. M1 added the loud exits; M2 lifted the invocation into `sdlc actual` + close's inline suggestion.

## A contract between a prose producer and a code consumer must live in ONE referenced place, and the consumer gates on a TOKEN, not prose presence

**Pattern:** `sdlc`'s judges (LLM, prose) emit a verdict; the parser (code) gates merges on it. The contract lived only as prose on each side — each prompt hand-wrote the verdict format, and the parser independently grepped for it. They drifted: the parser only checked the *first non-empty line* for `VERDICT: CLEAN`, so a judge that wrote a title or "I've reviewed…" line first dropped to a legacy sentinel-grep that **defaulted to `failure` → blocked the merge** (forcing `--no-judge`, which kills *all* judges). The token said pass; the prose presence said fail; the parser believed the prose. A sibling parser returned `unknown` on a perfectly good review. Two independent parsers + N hand-written prompts = guaranteed drift.

**Rule:** When prose (an LLM/human producer) and code (a consumer) share a result protocol: (a) **one source of truth** — a single contract object the code embeds into the prompt verbatim (`ContractPreamble`) AND parses against, plus a human-readable mirror kept in sync by a **drift test** (assert both directions: every code token in the doc, every doc token in the code). (b) **Gate on the structured token, not prose** — read `VERDICT: <TOKEN>`, map the token to blocking/non-blocking; a non-blocking verdict *with* notes must PASS. Never gate on the presence of words like "findings"/"note". (c) **Scan robustly but guard precisely** — find the token even behind a preamble (don't be brittle), but because judges review *this very parser* and quote the contract in prose (`VERDICT: BLOCK is the generic hard block`), require a trailing precision guard (token followed by `(confidence…)` or EOL) so a quote can't shadow the real verdict — same meta-trap as [[the structure-parser-on-meta-input lesson]].

**Origin:** #70. M1 = robust token scan + the false-positive fix (proved live: a milestone-review that would've been `unknown`/`failure` parsed cleanly). M2 = `ContractPreamble` embedded by all prompts + `construct/judge-output-contract.md` + the bidirectional drift test.

## Inject what the model structurally lacks — and inject it forward (at design), not just backward (at review)

**Pattern:** Agents play good local tactics (clean function, handled edge case) but weak whole-board architecture — the payoff/cost of a structural decision shows up months downstream, so there's little training signal for it and the model can't have learned good taste there. Leaving architecture to the model's judgment fails silently. #75 made architectural principles (DRY, PURE, later shim-externals) an explicit, persistent, prompt-level scaffold: a single markered registry (`ARCH-*`, `//go:embed`'d) delivered to the planning + plan-quality + code-review prompts. Critically, the workflow had `claim` and `change-code` (the plan-quality *review* gate) but **no transition for "I'm now designing"** — so the highest-leverage moment (architecture is *decided* at plan time, while still cheap to change) had no injection point. Added `sdlc start-plan` to fill it.

**Rule:** When the model is reliably weak at a capability *because the world gives it no training signal* (architecture, long-horizon design, anything whose payoff is many turns out), don't hope it improves — **encode the human judgment as a referenced scaffold** and deliver it into the loop. Two design rules: (a) **inject forward, at the decision point, not just backward at review** — catching bad architecture in a plan (changeable) beats flagging it in a diff (built); if the workflow has no "decision point" transition, add one (a verb). (b) **One source, delivered per context** — markered entries (`ARCH-DRY`, stable semantic handles, no ordinals) in one embedded file; render the relevant *lens* (`at-plan` vs `at-review`) per consumer. A fresh-context subagent needs the full definitions delivered (a bare marker dangles); within a context, deliver-once + cite-the-marker. Pair the machine registry with the human narrative (AGENTS.md) and a **drift test** keeping them in sync (the [[one-referenced-contract lesson]] pattern).

**Origin:** #75. M1 = the registry + embed into plan-quality/review/dry-pure (authored once). M2 = `sdlc start-plan` (forward injection) + AGENTS.md workflow + the narrative-drift guard. Dogfooded: M1's own milestone-review ran through the new at-review lens.

## A gate the agent can skip isn't a gate — make the binary own it; and when you "merge" two things, hunt for other consumers before deleting

**Pattern (#69):** Two redundant per-boundary code reviews ran at every milestone — the agent's `superpowers-requesting-code-review` subagent (mandated by prose) *and* `sdlc milestone-close`'s own auto-dispatched review. The fix wasn't to pick one prompt; it was to recognize that **a review the agent is merely *told* to run is an opt-in, not a gate** — agents forget, skip "because it's simple", or vary. Moving ownership into the binary (`sdlc close`/`milestone-close` dispatch the one review themselves) makes it run every time, and lets the binary also do the cheap deterministic checks an agent forgets (boxes ticked, status flipped) before spending tokens on the LLM pass. The agent's job shrinks to "run the verb"; the verb guarantees the review.

**Rule 1 — own the gate in code, not in prose.** If a step *must* happen at a checkpoint, the checkpoint binary should perform it, not instruct the agent to. Prose mandates degrade to optional; a binary dispatch doesn't. Give it a precise `--no-<gate>` bypass (per [[inject-what-the-model-lacks]]'s sibling #67 convention) so skipping is an explicit, logged acknowledgment — not a silent omission.

**Rule 2 — procedure refers, registry defines (the two-file split).** When one prompt needs cross-cutting principles (here: the ARCH-* registry), don't paste the principle text into the prompt — that re-duplicates the registry, an ARCH-DRY violation *in the file that polices ARCH-DRY*. Keep the **procedure** (`code-review.md`: checklist, severity, verdict) separate from the **principles** (`architecture.md`), have the procedure *cite markers* (`{{ARCH_STAR}}`, expanded from the registry via one shared extractor), and co-locate the definitions at dispatch. A guardrail test that fails if a principle's defining phrase leaks into the procedure keeps the registry the sole definition site. Extends the [[one-referenced-contract lesson]] / [[inject-what-the-model-lacks]] "one source, both reference" pattern.

**Rule 3 — before deleting a "duplicate", grep for other consumers.** The plan said "drop the now-superseded `code-reviewer.md`." Implementation found a *live sibling* skill (`superpowers-subagent-driven-development`) still referenced it — so it wasn't an orphan. The root-cause fix was removing the *boundary mandate* (the redundant run), not deleting the template. Deleting on the plan's say-so would have dangled a reference. A plan written before reading every caller will over-claim what's safe to remove; verify at implementation.

**Origin:** #69 (rode on #75's registry, #70's verdict contract, #67's per-gate bypass). M1 = the one embedded reviewer + kill the double-run. M2 = `close` as a boundary + the shared `dispatchBoundaryReview`/`firstCommitReferencing`. Both milestones + the whole-issue close were reviewed *by the very reviewer they built* (M1 SHIP, M2 FIX-THEN-SHIP→fixed, issue-close SHIP) — the feature dogfooded itself.

## A DRY comment is a claim — make it true or weaken it; and pin every branch of a documented fallback

**Pattern (#58):** Extracting `issueFilePath` as the shared issue-file resolver, I wrote its doc as *"the same resolution close.go … rely on, kept in one place (ARCH-DRY)"* — but left close.go's **parallel inline glob** untouched. The comment asserted a unification that hadn't happened: two copies, one claiming to be one. The boundary review caught it — an ARCH-DRY overclaim *in the change whose whole point was ARCH-DRY*. Separately, `boundaryWindowBase`'s documented fallback to branch-start fires on **two** distinct triggers (no prior boundary at all; a prior commit that exists but lacks the `Review-Verdict:` trailer), but the first test pinned only the first trigger — the riskier "exists-but-no-trailer" over-cover path was undefended.

**Rule 1 — a comment that says "shared"/"one place"/"DRY"/"the same X uses" is a *claim about other code*, not a description of this function. Before writing it, route the other consumer through the helper (make it true), or don't write it. The moment you claim unification, grep the call sites and confirm there's exactly one.** An aspirational DRY comment is worse than none: it tells the next reader the duplication is gone, so they stop looking.

**Rule 2 — when a function documents a fallback reachable by N distinct conditions, write N tests, one per condition — not one test for "the fallback."** "No prior boundary" and "prior boundary present but malformed/missing-trailer" are different code paths through the same `return`; the second is where the safe-direction (over-cover) guarantee actually earns its keep. A single fallback test gives false coverage confidence for the sibling trigger.

**Origin:** #58 (milestone review window → prior boundary). Both fixes folded in from the SHIP boundary review before the close commit: routed close.go's locate step through `issueFilePath` (true DRY), added the 4th `MissingPriorTrailer` fixture. Same family as [[A gate the agent can skip isn't a gate]] Rule 2 (procedure refers, registry defines) — claims of single-sourcing must be verified at the call sites, not asserted in prose.

## `git add -A` / `git add <dir>/` sweeps unrelated untracked WIP — stage explicit paths

**Pattern (#77 ship):** Two separate broad-add slips in one session put files where they didn't belong. (1) My issue-close commit used `git add -A`, which swept an untracked `000079-doc-review-flow.md` (a separate in-progress issue, the operator's local-only WIP) into the #77 close commit. (2) Then `sdlc merge`'s archive step (`merge.go:421`) did `git add workshop/issues/ workshop/history/` — a *directory-wide* add — and committed that same untracked #79 onto main and pushed it. Both captured a file that had nothing to do with the change. The first I caught and amended pre-merge; the second reached `origin/main` before I noticed. Notably this is the dark twin of [[A gate the agent can skip isn't a gate]]/#78: once the merge guard was loosened to *tolerate* untracked files, a latent broad-add downstream silently *committed* them — loosening a guard makes everything it used to block reachable.

**Rule 1 — stage explicit paths, never `-A` or a bare directory, when the working tree may hold unrelated WIP.** `git add <specific files you changed>`. A repo with concurrent multi-agent / multi-issue work *always* may hold unrelated untracked files (another issue being drafted, a peer's WIP, a local-only skill). `git add -A` / `git add dir/` assumes the working tree is yours alone — it usually isn't. The cost of listing paths is trivial; the cost of committing someone's half-written work (or pushing it to main) is not.

**Rule 2 — code that commits on the user's behalf must add only the paths it touched.** A tool step that moves/generates files (archive, scaffold, sync) and then commits should `git add -- <exact paths it just wrote/removed>`, computed from what it did — never `git add <dir>/` to "catch the moves." The dir-add catches unrelated untracked neighbors too. (#80 fixes exactly this in `sdlc merge`'s archive step.)

**Rule 3 — when a broad add already happened, look before you push.** `git status --short` / `git show --stat HEAD` before pushing a commit a tool made on your behalf. The #79 leak would have been a one-line catch at `git show --stat` of the archive commit; instead it rode the push. Untracked-file scares in this session ([[pair-doctor recovery]], #79) all share the tell: a `git status` that lists files you didn't create.

**Rule 4 — when the committed output set is variable/hard to enumerate (so explicit-path staging isn't practical), guard `git add -A` with a clean-working-tree PRECHECK instead.** Some tools must `git add -A` because what they commit is a *computed* set — a re-weave's symlinks + per-harness entry files + untrack-now-ignored removals, not a fixed list. For those, make clean-before a precondition: if the target's tree is dirty *before* the tool acts, SKIP + report (never `-A`); if it was clean before, every post-action delta is provably the tool's own output, so `-A` is safe. The skip must make the run exit NON-ZERO — a skipped target is left stale, and incomplete propagation ≠ success. **And the precheck's `git status --porcelain` must pin `--untracked-files=all`** — a `status.showUntrackedFiles=no` gitconfig otherwise returns empty for untracked files, blinding the dirty-check to the exact concurrent-session file it guards against (the sibling `push.go` already pins it; share the convention via one helper, ARCH-DRY).

**Origin:** #77 ship. Caught+amended the close-commit instance pre-merge; the merge-archive instance reached main (operator chose to keep #79 there) and is filed as #80. Same hazard family as the pair-doctor stash scare earlier in the session. **Recurred #109:** `sdlc propagate-base` (new in #106, so it predated none of Rules 1–3 yet shipped without them) hit the identical sweep — `git add -A` committed a *concurrent* Claude session's uncommitted plan work in a sibling repo (parley.nvim) during a base-layer propagation; raced, resolved by luck. Fixed with Rule 4's clean-tree precheck; the boundary review then caught the config-blindable porcelain read (the `--untracked-files` pin). The recurrence is the tell that a hazard rule must be wired into the *shared mechanism* (a `commitConsumption`/`gitStatusPorcelain` helper every committing tool routes through), not re-learned per new tool.

## A test that `cd`s into a temp workspace must hard-guard it — `cd ""` falls through to the host repo

**Pattern (#79):** `docflow.test.sh` builds throwaway git repos via `mktemp -d` and `cd`s in. Under the Claude sandbox `mktemp` is *denied* → `$work` empty → `cd "$work"` is `cd ""`, which in bash **succeeds as a no-op and leaves you in the host repo**. The e2e then ran `git config user.name/email`, clobbered `README.md` to `seed`, and *committed* it as a bogus `Operator <op@example.com>` commit on the feature branch. Worse, my first cleanup fixed the *visible* damage (restored identity, deleted stray `post.md`/`two.md`) but missed the **committed** README clobber — invisible to `git status` (tree clean), and `README.md` is a base-layer file that would propagate downstream. The fresh-context boundary review caught it (FIX-THEN-SHIP); reverted by rebasing the junk commit out.

**Rule 1 — a test that creates a temp workspace and `cd`s into it must abort *before any cd/write* if the temp creation failed or came back empty.** `cd ""` returns 0 and silently strands you in `$PWD` (the real repo); every later `git init`/`config`/`commit` then mutates it. Guard `[[ -n "$work" && -d "$work" ]] || abort`, and belt-and-suspenders assert `$PWD` is under the temp root right before destructive ops. Prefer SKIP-when-no-temp over FAIL so the suite stays honest in restricted envs — but never fall through.

**Rule 2 — after a destructive-test scare, enumerate every mutation it could have made and verify each is reverted, not just the ones `git status` shows.** A clobber that got *committed* is invisible to `git status` (clean tree) — it lives only in the branch's log/diff. "Found + fixed" written into a `## Log` is itself a claim to verify: `git diff <base>..HEAD --stat` and eyeball every file before believing it. The author's post-scare relief is exactly what a fresh-context review exists to backstop.

**Origin:** #79 (docflow). Same family as [[git add -A sweeps unrelated untracked WIP]] — the shared tell is host-repo state you didn't intend to touch (a `git status`/diff listing files or commits you didn't mean to create). There the scare was *untracked*; here it was *committed and clean*, which is the more dangerous because `git status` says nothing.

## A library helper that `die()`s (os.Exit) can't be made best-effort by its caller — return errors, let severity live at the call site

**Pattern (#82 M1):** I reused `claim`'s `syncOnMain`/`syncOnBranch` from `issue new` so a freshly-filed issue auto-syncs to main. The sync was meant to be *best-effort* (the file is already written; an offline/no-origin push failure must not abort `issue new`), and I wrote `if err := sync(...); err \!= nil { warn }`. But the helpers called `die()` (os.Exit) internally on every git failure — so the "warn" branch was **dead code**: a failed push killed the whole command (and the `fetch` test, whose origin is unreachable, took the suite down with it). The same code is *fatal* for `claim` (its whole job is the sync) and *advisory* for `issue new` — but a helper that exits can only express one severity.

**Rule — a function reused by ≥2 callers with different failure tolerances must `return error`, not `die()`/`os.Exit`/`panic` internally.** Severity is the *caller's* decision: `claim` does `if err \!= nil { die(...) }` (UX unchanged), `issue new` warns. `die()` in a library hard-codes "fatal" and makes best-effort reuse impossible — and silently, because the caller's error-handling compiles fine as dead code. When extracting a shared helper from a `die()`-laden command, convert the `die()`s to `return fmt.Errorf(...)` first; the original caller re-adds the `die()` at the boundary. (Same form-vs-essence split as the merge guards: form/fatality at the edge, essence in the testable core.)

**Origin:** #82 M1. Caught by the plan-quality gate flagging the dead-code handler *plus* a real `fetch` test failure (process exit). Tests now pin both directions: `claim` fatal, `issue new` best-effort (no-origin → file created, warns, returns nil).

## `strings.TrimSpace` on a whole `git status --porcelain` blob strips the FIRST line's leading status column — field-split, don't column-slice

**Pattern (#82 M2):** porcelain is column-formatted: `XY␣path` (status in cols 0-1, path from col 3). `worktreeDirty` returns `strings.TrimSpace(string(out))` — trimming the *whole* output, which eats the leading space of the **first** entry only: `" M workshop/issues/x.md\n D y"` → `"M workshop/issues/x.md\n D y"`. A column-based parser (`parsePorcelainStatus`, `line[3:]`) then reads the first line's path as `"orkshop/..."` (off-by-one) and mis-buckets it — here, a dirty issue file got classified Blocking instead of Tracker, so the merge refused. Lines 2+ keep their leading space (they follow a `\n`), so the bug is *first-line-only* and easy to miss in tests that put the interesting line second.

**Rule — extract a porcelain path with `strings.Fields` (status = field 0, path = field 1, rename dest = last field), never fixed-column slicing, when the input may have been whole-trimmed.** Field-splitting is immune to the leading-space loss. If you must column-slice, don't whole-`TrimSpace` the blob first — trim per-line or only trailing. And test the regressing line *first* in the input, since that's the only position the trim corrupts.

**Origin:** #82 M2. Caught by the e2e (`TestRunMerge_DirtyTrackerFile_Proceeds`) — the pure `assessDirty` table test passed because its fixtures kept the leading space; only the real `worktreeDirty` path exposed it. Pinned by `TestPorcelainPaths` + a trimmed-leading-space case in `TestAssessDirty`.

## A test fake keyed on the same value-shape as the code masks format-mismatch bugs at the IO boundary — dogfood against real data

**Pattern (#76):** the `sdlc state` close-off check passes an issue's ID to a ship probe that scans `git log` for `#N` commit subjects. `IssueState.ID` is *zero-padded* (`000082`) but commit subjects use the *unpadded* number (`#82`, §12) — so the probe was grepping `#000082`, matching nothing, silently reporting every issue as "not shipped." Every unit test passed: `TestDetectDrift_CloseOff`'s fake probe was keyed on the same padded IDs the code passed it (`map[string]bool{"000051": true}`), so the fake and the code agreed on a representation that was wrong at the *real* git boundary. The bug only surfaced when I dogfooded against the live repo (a synthetic 2/2 issue for the already-shipped #82 produced *no* finding where one was obvious). The fix (unpad before the probe) was then guarded by re-keying the fake on the unpadded number — so dropping `unpadID` now fails the test.

**Rule 1 — when a fake stands in for an IO call, key it on the representation the *real* dependency uses, not the one that happens to be convenient in the test.** A fake that mirrors the caller's internal value-shape only proves the caller is self-consistent; it can't catch a mismatch between that shape and what the external system (git, gh, an API, a filename convention) actually expects. Ask "what string would *real git* receive here?" and make the fake demand exactly that. Here the tell was that both the code and the fake spoke `000082` while git speaks `82`.

**Rule 2 — a heuristic over external data is not verified until it's run against real external data once.** Green unit tests with a hand-built fake are necessary but not sufficient for an IO-boundary feature; a single dogfood pass against the live system (here: does `sdlc state` actually flag a known-shipped issue?) is what exposes representation mismatches the symmetric fake hides. Budget that dogfood step before claiming done.

**Origin:** #76 (close-off drift). Caught by dogfooding, not by the (passing) unit suite or the clean SHIP boundary review. Same family as [[A pure helper unit-tested in isolation can be silently un-wired from its caller]] — both are "the test and the code agree with each other while disagreeing with reality." There the gap was wiring; here it's value-representation at the boundary.

## A target can lie by aspiration — generalizing a proven mechanism to unbuilt siblings, marked "clarity HIGH", hides the gap instead of defending it

**Pattern (#95 → #104):** the `base-layer-mechanics` target was extracted from
the PROSE visibility fix (#99, which was actually built + verified on the parley/nous
passes). In the same breath, its `skill`, `settings`, and `file-op` slices were
written "for free" by analogy — "these compose like prose" — and the skill slice was
marked **clarity: HIGH** with a precise formula (`skills(R) = ⋃ export-skills(Lᵢ) ∪
internal-skills(Lₙ)`). None of it was built for skills: `grep` finds **no skill code
that consults visibility at all**, there are three disagreeing discovery mechanisms,
and the file-op slice's "conflict-accumulating error-monad" has zero collision logic
in `plan/`. The gap stayed invisible for two reasons: (1) only the claude target is
ever exercised, and claude routes *around* every skill gap; (2) **the target itself
signalled "solved"** — "skill: HIGH" told future-us not to look. A document meant to
defend an invariant from drift instead documented a wish and gave false confidence.
The miss was found only by tracing the actual link targets on a live multi-layer
repo (brain) during the cutover.

**Rule — a target must separate design-clarity from implementation-status, and bind
every confidence claim to a test or a verified pass.** "We understand the formula" is
not "the code honors it." When you generalize a proven mechanism to sibling artifact
types, mark them **conjecture / NOT-built** until exercised — never "HIGH" — because
a clarity label on unbuilt math is worse than silence: it actively suppresses the
audit that would catch it. Prefer a per-slice status banner ("DESIGN-ONLY", "built +
test-bound", "partial") over a single clarity grade that conflates the two. And reach
for a target only at the level you can defend with a fixture; the cross-cutting math
(the algebra) and the subsystem that instantiates it (the skill system, #104) are
different targets — collapsing them is how the subsystem's declaration/identity/
lowering/serving invariants went unowned and unbuilt.

**Origin:** #95 cutover gap analysis → #104 + the `skill-system` target. Same family
as [[A test fake keyed on the same value-shape as the code masks format-mismatch bugs]]
— "the artifact and the team agree with each other while disagreeing with reality" —
but one level up: there a fake lied about an IO boundary; here a *target* lied about
whether a subsystem exists.

## 2026-06-18 — Verify `change-code` actually created the branch before committing

`sdlc change-code --issue N` is supposed to create the feature branch (in-place by
default) after its gates. In #116 it committed the issue-file changes to `main`
("issue-sync: update issues") but did **not** leave me on a new branch — so every
subsequent `#116` code commit + a raw `git push` landed directly on `main`,
bypassing the `pr → merge` pre-merge judges, archive, and propagate.

Two compounding causes: (a) the branch creation silently didn't happen (error or
no-op after the judges), and (b) I filtered change-code's output with
`grep -vE "^#|^- |^\s"`, which hides the indented `[ok] Branch … created in place`
confirmation **and** any indented error — so I never saw that the branch step
failed.

**Rule:** after `sdlc change-code`, confirm the branch before touching code —
`git rev-parse --abbrev-ref HEAD` should show the issue branch, not `main`. Don't
over-filter sdlc's stdout/stderr to the point of dropping its branch/gate
confirmations; the indented `[ok]`/`Error:` lines are the load-bearing signal.
Recovery when work lands on main anyway: `sdlc push` (runs the pre-merge judges +
archives the done issue) then `sdlc propagate-base` — but note push judges only
see the *unpushed* window, so if the code was already `git push`ed they review an
empty diff; lean on the end-of-issue boundary review (which did see the code) for
that case.

## 2026-06-18 — Verify an issue's factual premises against ground truth before building (#118)

#118's Spec rested on two confident assertions that real-transcript inspection
disproved: (1) the subagent-dispatch tool is named `Task` — it is actually `Agent`
(`Task*` are the todo tools); building against `Task` would have produced a
silent no-op detector. (2) capped subagent spans were "the bulk of the ~3.5×
supervised overshoot" — a census of all 33 historical `Agent` spans found every
one **under** the 15-min cap, so the fix is a *demonstrated no-op* on every current
ledger row (old-engine vs new-engine returned identical actuals over identical
windows). The fix was still worth building (unit-correctness + forward-looking),
but the *rationale* in the issue + the calibration banner was wrong and would have
shipped a false "wrong-ruler explains the overshoot" story into durable docs.

**Rule:** when an issue asserts a concrete fact about the system (a tool/field
name, a JSON shape, a magnitude/causation claim like "X explains Y"), **check it
against ground truth before designing** — grep the real transcripts/data, count
the actual distribution, diff old-vs-new over the *same* window to isolate your
change from confounds (here the window-extends-to-HEAD artifact masqueraded as an
engine effect). Surface a disproven premise to the operator as a decision, not a
silent correction (#118: "build it, correct the rationale"). A plan that builds a
correct mechanism on a wrong *why* still poisons the calibration loop the mechanism
feeds. Connects to brain's `measure-before-rebuild` + `artifacts-lie-by-aspiration`.

## 2026-06-18 — Don't truncate `sdlc` judge/review output with `tail`/`grep` (#115)

`sdlc milestone-close`/`close`/`merge` print the LLM review (verdict + findings) to
stdout. Piping that through `| tail -N` (to "just see the verdict") DROPPED the
Important findings three times this session — and once the close had already mutated
state, so the findings were gone and I had to re-run `sdlc judge milestone-review
--base … --head …` (a full LLM review cycle) just to recover them.

**Rule:** capture `sdlc` judge/milestone-close/merge output in FULL — never `| tail`
or `| grep` away the findings block on the command that RUNS the judge. Filter on a
second read of the saved output file instead. Re-running an LLM judge to recover
lost findings burns a whole review cycle.

## 2026-06-18 — A plan's "remove X" step is checked against the close evidence — "inert at runtime" ≠ done (#115)

#115 M4.1 Step 2 said `rm` the retired whole-dir `construct/datatype` symlink in
every consumer. I judged the symlinks vestigial-harmless (the `datatype` binary
DAG-walks each layer's real dir, ignoring them) and RESTORED two I'd removed — but
the close `--verified` claimed "no dangling symlinks / git status clean in every
repo," which the symlinks made FALSE. The end-of-issue integration review caught it
(I1). Two ground-truth corrections also surfaced: weave's `PruneOrphans` DOES GC the
symlink once the manifest row is gone (`propagate-base` pruned them) — so both the
review's "won't be GC'd" and my "construct/ isn't a managed location" were wrong.

**Rule:** a plan step that says "remove/clean up X" is a commitment your close
`--verified` is judged against — execute it, don't rationalize skipping it because
it's inert. If you genuinely skip a planned step, change the evidence claim to match,
or the boundary review will (rightly) flag the divergence.

## 2026-06-18 — Reconcile atlas after a symbol MOVE by grepping the OLD location, not just feature prose (#115)

M1 moved `Resolve`/`ParseDeps` from `cmd/weave/internal/layer` → `pkg/layergraph`.
The atlas reconciliation updated `weave.md`'s *dynamic-skills* prose but left the
"Surface" code-MAP bullet still attributing those symbols to the old package. The
pre-merge `specs` judge caught it — atlas is the "always-current codebase map," and a
navigation pointer to a symbol's old home is exactly the drift that gate exists for.

**Rule:** when a refactor RELOCATES a symbol/package, grep `atlas/` for the OLD
path/package name (e.g. `cmd/weave/internal/layer`), not just the feature name —
code-map / "Surface" / file-pointer sections drift silently because they key on
location, not behavior. Feature-prose reconciliation alone misses them.

## 2026-06-24 — Subagent context is throwaway: instruct it to surface lessons, and persist them here (#122)

**Pattern:** Dispatched three subagents this session (a plan reviewer, an M2
implementation fork, an M2 code reviewer) and asked each for *findings + decisions +
deferrals* — but never explicitly for *reusable lessons*. Even the lessons that did
surface incidentally (the stray-binary footgun, an estimate-block gotcha) would have
evaporated: the subagent's context is discarded when it returns, AND the main session's
context is discarded at session end — the only thing that survives either boundary is
what's written to `workshop/lessons.md` or memory. The operator caught this by asking
"do you instruct subagents to report lessons back?" — I did not, and I had not been
routing the session's own lessons here either.

**Rule:** (a) Subagent prompts must explicitly request *"reusable lessons / gotchas,
separate from findings"* — findings are about the work product (task-scoped); lessons
are for future work (cross-task). The throwaway context only surrenders them if asked.
(b) The main session routes the worthy ones into this file — and running a code review
that finds mistakes (the subagent's *or* the work's) **obligates** a lessons entry per
AGENTS §4. The cross-boundary persistence is the whole point; a lesson that lives only
in a discarded context is a lesson un-learned (the consistency-prosthesis idea).

**Origin:** #122, prompted by the operator's question mid-implementation. Same family as
the brain `consistency-prosthesis` memory — coherence across time is grafted from
outside (the durable ledger), never inherent to a single context.

## 2026-06-24 — CUE authoring gotchas + the stray-`./cmd/X`-binary breaks leaf-dir tools (#122)

**Pattern:** Three CUE/build tripwires building the vocabulary layer. (1) `cue vet`
rejects list `+` concatenation in v0.11+ ("Addition of lists is superseded by
list.Concat") — the M1 vet gate failed first try on `categories.open + categories.active`.
(2) CUE **`#`-definitions don't `cue export`** — only concrete fields reach the JSON, so
the category data sdlc consumes had to be a concrete `categories:` field, with `#Status`
*derived* from it via `or()` (a definition built from the concrete, not the reverse).
(3) The stray-root-binary footgun has a sharper consequence than the [[A hand-maintained
copy of generated data drifts]] tripwire (line ~78) noted: `go build ./cmd/vocabulary`
(no `-o`) drops `./vocabulary` at the repo root, and because `vocabulary`/`datatype`
resolve `<root>/<name>/` as the leaf-local *directory*, the stray *file* makes
`MergeByName` hit ENOTDIR — so it doesn't just get swept into a commit, it **breaks the
tool** (`vet/export/check` fail with "not a directory").

**Rule:** Build CUE list unions with `list.Concat([a,b,…])`, never `+`. When a consumer
needs a value out of a CUE model, it must be a **concrete field** (definitions are
validation-only and never export) — derive the `#`-def from the concrete via `or()` so
membership is stated once. Build Go binaries into `bin/` (or `-o /dev/null` for a
compile-check), and gitignore the stray root binary at **every** `cmd/<X>` name
(`/vocabulary`, `/datatype`, not just `/sdlc`) — especially for tools that read a
same-named leaf dir, where the stray file is a functional break, not just commit noise.

**Origin:** #122 M1 (the `list.Concat` + `#`-export gaps, hit inline) and M2 (the
ENOTDIR consequence, flagged by both the implementation fork and the code review).

## 2026-06-24 — In-sandbox, SDLC judges can't reach network: --no-judge + substitute a fresh-context review; and change-code needs a ## Estimate block (#122)

**Pattern:** `sdlc change-code`/`milestone-close` auto-dispatch their LLM judges via the
`claude` CLI, which needs the network the Claude-Code sandbox blocks — so the judge
hangs or degrades, and closing `--no-judge` records `Review-Verdict: not-run`, leaving
the **mandatory §3 boundary review unrun**. Letting `not-run` stand would skip the one
review the boundary exists to guarantee. Separately, `change-code`'s estimate gate
requires a `## Estimate` fenced block (v2 primitives reconciling with `estimate_hours`),
which `sdlc issue new` does **not** scaffold — so it refuses until you add one.

**Rule:** (a) In-sandbox, close `--no-judge` to avoid the network hang, but **substitute
the mandatory review** with a fresh-context reviewer subagent against the boundary diff
window (prev-boundary..HEAD), then fix Critical/Important and **record the real verdict
in the issue Log** — don't let `not-run` be the final word on a reviewed boundary. (b)
Add the `## Estimate` block *before* `change-code`, and **derive** the total from the
itemized v2 primitives (sum design×(1+buffer) + impl×familiarity) rather than back-fitting
items to a guessed total — the estimate-quality judge exists to catch back-fitting.

**Origin:** #122 (change-code + both milestone closes ran `--no-judge`; M1/M2 reviews ran
as substitute subagents). Sibling to [[Don't truncate sdlc judge/review output]] — both
are about not losing the boundary review's signal.

## 2026-06-24 — DRY-ing an enum onto a model: rewire only category-equal literals; exposing an API ≠ enforcing it (#122 M3)

**Pattern:** Collapsing scattered status literals onto a `vocab.Issue()` model, the rewire-vs-keep line is subtle and got it wrong is easy. (1) A literal that *exactly equals a whole category* (`isTerminalStatus`'s `done|wontfix|punt` → `IsTerminal`; `!="open"` → `!IsOpen`; the `validStatuses` set → `AllStatuses()`) is a true DRY target. But a *single sub-category value* — `"working"` alone in the in-flight/contention check, `"done"` alone in the close-gate / reclose / gh-close, the `"working"` literal `claim` writes — is a **value-specific behavior**, not a category test; rewiring it to `IsActive`/`IsTerminal` silently **broadens** scope (e.g. makes `blocked` also hit a working-only path). Keep those as annotated literals. (2) The model's transition graph is *stricter* than sdlc's actual `setstatus`, which has **no transition-legality gate at all** — so wiring `CanTransition` as a hard reject is a behavior change (tightening), not a refactor.

**Rule:** Rewire a literal to a model predicate **only when it equals a whole category**; a sub-category value is value-specific — keep it, annotated with the why (the [[A hand-maintained copy of generated data drifts]] honest-grep then passes on *annotations*, not deletions). And **exposing a model API ≠ enforcing it**: ship + conformance-test `CanTransition`, but make *gating* on it an explicit operator decision — never a silent side-effect of "rewiring to the model," because it tightens previously-ungated behavior. A Done-when like "a model-forbidden transition is rejected" is a *separate enforcement decision* from "consumers read the model," and should be split out (or deferred) rather than smuggled into the rewire.

**Tactical gotchas, same session:** `vocabulary export` has **no `-o` flag** and `--noun`/`--output` are mutually exclusive — so the binding directive is `//go:generate sh -c 'vocabulary export --noun issue > issue.json'`, not `... -o issue.json` (a plan directive that names flags must be checked against the real flag set). A `switch x {case "a","b":}` → predicate needs a **tagless** `switch {case pred(x):}` (Go can't mix constant cases with function-call cases). And `gofmt` hand-written Go *before* committing — a misaligned struct-tag comment shipped in one commit and only surfaced when a later `gofmt -w` dirtied the tree.

**Origin:** #122 M3 (the `pkg/vocab` rewire). The category-vs-value split kept the rewire behavior-preserving; the tightening call left `CanTransition` exposed-but-ungated (Done-when's "rejected" deferred as an operator decision).

## 2026-06-25 — Enforcing a state-machine gate: widen the model to every legitimate flow first, and treat the test suite as the canary (#122 M4)

**Pattern:** Turning on lifecycle enforcement (gate `set-status` on `CanTransition`) is a *behavior change*, not a refactor: the formal graph was stricter than the code's previously-ungated behavior. Enforcing the M1 graph as-drawn would have wrongly rejected legitimate flows (triage `open→wontfix/punt`, resume `punt→working`), so the model had to be **widened to the real legal set first** (+6 edges), *then* gated. Turning on the gate reddened 3 existing tests — and each one was a *signal to classify*, not a nuisance to suppress: all 3 turned out to use an excluded transition (`open→blocked`, `done→open`) merely as a throwaway mutation (convenience → repointed to a legal transition), none was a genuine flow (which would have meant a missing edge). A blanket `--force` to make them green would have hidden a real model gap if one existed.

**Rule:** Before enabling enforcement of a declared state machine: (1) **widen the model to permit every legitimate transition** the system actually performs (the gate is only as right as the graph). (2) **Use the existing test suite as the canary** — when enforcement reddens a test, *classify* it: a throwaway/convenience transition → repoint to a legal one; a genuine workflow → it's a missing edge, surface it to the operator, do **not** blanket-`--force`. (3) Ship a **`--force` escape** (logged) so the gate is a guard with a pressure-relief valve, not a wall — and the friction of needing `--force` is itself the signal that the model is missing an edge. (4) **Order a guard chain general→specific** — the graph-legality check (`CanTransition`) belongs *before* value-specific guards (`→done` routing, reopen, started-stamp), so `open→done` reports the accurate "illegal transition" rather than a misleading "use sdlc close". Gate only the *arbitrary-flip surface* (`set-status`); leave verbs that perform fixed legal transitions (`claim`/`close`) ungated.

**Origin:** #122 M4 (operator chose to enforce now, not defer). The widen-then-gate order + test-suite-as-canary kept the enforcement from breaking real flows; `--force` covers the deliberately-excluded edges (`open→done`, `working→open`, `open→blocked`).

## 2026-06-25 — `milestone-close --actual` suggests CUMULATIVE; per-milestone actuals are INCREMENTS (#122)

**Pattern:** Across #122's four milestone closes, `sdlc milestone-close`'s `--actual` omit-suggestion (and `sdlc actual`) reported the issue's **cumulative** focused-hours (window anchored at the issue's first commit), but a *milestone* actual is that milestone's **increment**. Passing the increment (`cumulative − Σ(prior milestones)`) tripped the sanity-warn every time (M2 ~2.7×, M3 4.3×, M4 6.2×; note **≥10× refuses** — a long issue will eventually hit the wall). Re-derived the increment by hand each close.

**Rule:** For a multi-milestone issue, milestone actuals are increments: `this-milestone = sdlc's-cumulative-suggestion − Σ(already-recorded prior milestones)`. Expect (and ignore) the rising "Nx the measurement" warn — it compares your increment to the cumulative; it does NOT mean your number is wrong. (A possible sdlc fix: have `milestone-close` suggest the *windowed* increment — prev-boundary..HEAD — not the cumulative, mirroring how the close atlas-check already windows. Worth an issue if the ≥10× refusal ever blocks a real close.)

**Debug aside:** a test that hits the real `die()`/`os.Exit` crashes the whole `go test` binary with **no `--- FAIL` line** — just `exit status 1`. Find the culprit via `go test -v` and the **last `=== RUN` before the abrupt end**.

**Origin:** #122 M2–M4 closes (the warn recurred all three).

## 2026-06-25 — A single-source issue isn't DONE until every consumer DERIVES; "follow-up" must not offload the issue's purpose (#122)

**Pattern:** #122's whole purpose was "one source, consumers *compiled from* it, duplication *deleted*" — its Done-when even named the consumers ("categories propagate to Go/**Lua**", "compiled to consumers" — plural). At close I had wired *one* consumer (sdlc Go) + the enforcement, **hand-patched** the drifting help text, and silently reinterpreted the rest (parley Lua, the operator-prose/help surface) as "out-of-scope follow-up" — so for those surfaces `issue.cue` was *still just-documentation they don't derive from*. The duplication didn't get deleted, it **moved**. I did this *despite* having repeatedly warned — in the pensive, the `issue-lifecycle` target, and this very file — that the risk was the model becoming unenforced documentation. The boundary review caught one instance (help-text drift); I patched the line instead of reading it as "the consumer wiring is unfinished." The operator caught the rest.

**Rule:** For a single-source / DRY / "compiled to consumers" issue, **closing requires every consumer named in the goal to actually DERIVE from the source — or be explicitly de-scoped with operator sign-off.** "Follow-up" is legitimate for separable extensions, *never* for the thing that **is** the point (test: *"is the deferred work the reason this issue exists?"*). At the close gate, concretely: (a) **Done-when is the purpose-contract** — don't soften it to get the close; if it says "Go/Lua," Lua is wired or the operator agreed to split it. (b) **Shadow-sweep** — enumerate every consumer + `grep` for remaining restatements of the model; each derives or is provably gone. *"Is this just-documentation now?"* is a close gate, not a design slogan. Sweep two distinct things: the model's **data** (section lists, enums — often guardable by a containment/drift test) AND each consumer's **provenance self-claim** in prose/comments ("the single source of truth", "canonical", "hardcoded here") — the latter is NOT test-guarded, so a doc can keep *calling itself* the source after authority moved. `grep -rn "single source of truth\|source of truth\|canonical"` the touched surfaces; #145 shipped with `helptext/issue.md` still claiming to be "the single source of truth for the template" (the exact symptom the issue set out to kill) even though the section-list containment test passed — the boundary review caught it. Correcting one file's doc comment (scaffold.go) isn't enough; sweep *all* of them. **Sweep at SECTION granularity, not file granularity** — enumerating "which files mention X" and editing each file's *primary* prose still misses secondary sections within an already-touched file: a `RELATED` / "see also" cross-reference, an OPTIONS/FLAGS entry, a self-describing header. #146 (remove `close --milestone`) shipped with `helptext/milestone-close.md`'s RELATED block still saying `sdlc close — same close logic without milestone-review` — false post-change, in the *very file* the diff edited (a different line), caught only by the boundary review. So: for each touched doc, grep the whole file for the removed/renamed term AND scan its cross-ref/RELATED/see-also sections explicitly — the drift hides in the sub-section you didn't think to re-read. And extend the sweep to **command-invoking wrappers** (Makefiles/scripts), not just prose: a target that *invokes* a removed flag (`make close-issue MILESTONE=Mx` → `sdlc close --milestone`) is a harder-failing consumer than a doc that mentions it. **Corollary — fix a drift CLASS globally in ONE commit.** Under a re-review gate (the SDLC boundary review), a *partial* comment/doc fix is a treadmill: each re-close reviews only the new delta and finds the next stale instance, costing an ~8-min review per round. #146 burned THREE extra FIX-THEN-SHIP rounds (RELATED cross-ref → executable Makefile consumer → the subject file's own comments — one of which a prior pass had *rewritten wrong*) before a single global `grep -rn runClose` pass converged it, all comment-accuracy with zero behavior change. When you rename/remove a referenced symbol, grep the whole tree and fix every instance at once before re-closing. (c) **A boundary-review finding usually indicts a class, not a line** — a drifted doc means "this consumer class isn't wired," not "fix this string." (d) Keep the *project's long-term goal* in view across the whole arc, not just the current milestone's tasks. Because (a)-(c) are exactly what I *knew and still skipped*, the durable fix is to **encode** them (this entry + a memory; ideally a `sdlc close` gate that, for a single-source issue, lists consumers and asks "does each derive?") — the consistency-prosthesis applied to the *closer's* judgment, not just the designer's.

**Origin:** #122 close — operator correction ("you should have handled those as part of closing #122; you repeatedly warned of this risk and were eager to make it"). The unwired consumers were filed as parley#135 (Lua) + ariadne#125 (help-text-from-vocab), and #122's record reconciled to state what it actually delivered. Same family as [[A target can lie by aspiration]] (there a *target* over-claimed; here a *close* did) and [[A plan's "remove X" step is checked against the close evidence]] (claimed-done ≠ done).

## 2026-06-25 - A schema that only self-vets is untested against reality — run it over the whole corpus on day one

**Pattern:** #124 built an instance-conformance validator (cue-vet real issue files against
`#122`'s `#Issue`). `#Issue` had always passed `cue vet` *on itself* — but it had **never seen
a real instance**. On first contact it rejected **all 129** real issue files, for two reasons no
amount of schema-staring would surface: (1) cue's YAML loader **octal-parses leading-zero
scalars**, so unquoted `id: 000124` reads as the int `84` and `id: string` rejected every file
(fix: `id: int | string`, or quote ids); (2) a present-but-empty `estimate_hours:` parses as YAML
**null**, which `number & >0` rejected (fix: `(number & >0) | null`). Both were latent the entire
time #Issue existed as "documentation that self-vets."

**Rule:** A schema/validator is **untested until it has met real data at scale.** When you build
one against an existing corpus, the *first* test is "run it over the entire corpus" — and expect
the **schema** (not the data) to be what's wrong, in ways only real instances reveal (encoding
quirks, empty/null values, organically-grown fields). Corollary: model an existing artifact's
type **closed/strict only after** the corpus passes; default to **open** (`...`) at a fail-closed
gate so a valid-but-unmodeled field can't false-positive (and train `--no-validate`). This is
[[Measure before rebuild]] / "artifacts lie by aspiration" applied to schemas: a type that only
describes the data it was hand-written against is aspiration, not a tested contract.

**Origin:** #124 M1 — the validator engine. Surfaced by the implementation fork running the
engine over the corpus immediately (22/22 active pass; 15 legacy *done* history files correctly
flagged for missing `actual_hours` — genuine non-conformance, not a schema bug).

## 2026-06-25 - Two gotchas wiring a deterministic gate into an existing verb (#124 M2)

**Pattern A — a new side-effecting step in a shared flow fires on that flow's e2e tests.** #124
M2 added the instance-conformance gate to `sdlc merge`; that broke
`TestRunMerge_DirtyTrackerFile_Proceeds` — the e2e helper stubbed the *judge* seam but not the
*new gate* seam, so the gate ran for real inside a flow test. **Rule:** when you add a
side-effecting step to a verb that already has e2e/flow tests, audit every such test and
neutralize the new seam in their shared setup (`swapMergeDeps`), exactly as the existing
side-effecting seams (judges) are neutralized. A flow test should exercise the flow, not your
new gate — the gate has its own unit tests.

**Pattern B — building a tool binary at the repo ROOT shadows a layer-graph overlay dir.**
`go build ./cmd/vocabulary` from the repo root drops a `./vocabulary` executable, which
`resolveVocab`'s leaf-local `vocabulary/` overlay then tries to `ReadDir` → "not a directory",
breaking unrelated tests. **Rule:** build tool binaries to `bin/` or a temp dir, never the repo
root (both `vocabulary` and `sdlc` are gitignored at root precisely because a stray root binary
collides with the project-local overlay/source dirs).

**Origin:** #124 M2 — wiring the conformance gate into push/merge.

## 2026-06-25 - A "must-apply-everywhere" transform needs a guard that walks the assembled tree, not per-site tests (#125)

**Pattern:** #125 substitutes a `{{LIFECYCLE}}` placeholder in help text at *every* command-Long
load site. The plan-quality judge caught (pre-code) that the planned seam (`main.go`'s `add()`)
missed 3 of the load sites — `sdlc issue set-status`'s Long is set in `issue.go`, not via `add()` —
so it would have shipped a literal `{{LIFECYCLE}}` in the exact command the issue targets. The
durable fix isn't just "wire it correctly" — it's a test that **walks the real assembled command
tree** (`buildRoot()`) and asserts no `{{` placeholder survives in *any* command's Long. That
guard fails the instant a new command (or a new load site) forgets the seam.

**Rule:** when a transform must apply at N call sites (a substitution, a wrapper, a registration),
don't rely on per-site tests + remembering to wire each one — add ONE guard that enumerates the
*assembled* surface (the real command tree / the real route table / the real registry) and asserts
the invariant holds for every member. It catches both today's miss and tomorrow's new-site
regression. (Same family as the `{{ARCH_STAR}}` drift guard + `estimate_helptext_test.go`.)

**Also:** Bash heredocs (`cat << 'EOF'`) mangle Go's `\!` — `\!=`→`\\!=`, `\!x`→`\\!x` — even with a
quoted delimiter, yielding `illegal character U+005C`. Write Go source via the Write/Edit tools,
never a shell heredoc.

**Origin:** #125 — sdlc help text deriving lifecycle facts from the vocabulary model.

## 2026-06-25 - Verify behavioral claims against real runtime data before redesigning around them — comments go stale (#131)

**Pattern:** While spec'ing #131 (per-agent context meter from transcripts), a fresh-context spec
reviewer flagged that claude's recorded sid "never rotates on `/clear`" — citing a code comment in
`pair-cmux-title.sh` ("`/clear` rotates the file, leaving the cache pointed at the old jsonl"). I
took the comment as ground truth and redesigned the read from "use the pinned sid" to "newest
`*.jsonl` by mtime." That fix was a **regression**: the operator routinely runs multiple sessions
per cwd, and `~/.claude/projects/<enc-cwd>/` is keyed by cwd only, so newest-by-mtime aliases
co-located panes. Grepping real transcripts settled it the other way: compaction (`isCompactSummary`,
998k→47k) and even reset-to-0 events **continue writing the same pinned `--session-id` file** (context
rebuilt to 989k within one jsonl) — `/clear` does NOT rotate the file for pinned sessions. The comment
was stale (pre-`--session-id`-pinning). The pinned sid was the correct, simpler key all along.

**Rule:** A behavioral premise sourced from a **code comment** (or a reviewer quoting one) is a
hypothesis, not a fact — verify it against real runtime artifacts (logs, transcripts, on-disk state)
**before** you redesign around it, especially when the redesign trades away a property you already
have (here: exact per-pane attribution). Comments describe the code *as it was when written*; they
rot across refactors (the `--session-id` pinning post-dated this one). The check is cheap — one grep
of the actual `.jsonl` — and it caught a regression that both the author and a fresh reviewer would
otherwise have shipped. Same family as the §5 "verify before claiming" gate, applied to *inherited
assumptions* rather than your own claims.

**Origin:** #131 spec review round 3 — operator domain knowledge ("I run multiple from same cwd")
triggered the empirical check that refuted the comment-sourced premise.

## 2026-07-02 — Multi-milestone atlas gate: docs land in the milestone that introduces the surface, not front-loaded (#160)

`sdlc milestone-close` runs the §5 atlas gate over the *milestone window* (prev boundary → HEAD),
not the whole branch. So if you document a multi-milestone feature's architectural surface in an
early milestone (e.g. all the lifecycle/atlas prose in M1), the *later* milestones' windows contain
no `atlas/` change and their milestone-close **refuses at the atlas gate** — even though the feature
is well-documented overall. Two clean responses: (a) distribute atlas/docs updates to the milestone
that *introduces* each surface (M1 vocab → issue-lifecycle; M3 publish gate → pre-merge-checks), so
each window carries its own atlas change; or (b) `--no-atlas` on the milestone whose surface was
already documented upstream, with the reason in `--verified`. The whole-issue close's atlas gate uses
the *branch* window, so it sees all of it and passes regardless — the trap is milestone-scoped only.

**Origin:** #160 M2 milestone-close — the codecomplete lifecycle surface was atlas'd in M1's window
(issue-lifecycle.md, vocabulary.md), so M2's window had no atlas change and the gate fired; closed
with `--no-atlas` + rationale. Same family as [[milestone-close --actual suggests CUMULATIVE]] — both
are per-milestone-window mechanics that surprise if you reason at the whole-branch level.

## 2026-07-06 — A git-probe unit test on a non-repo temp dir exercises only the error branch (#154)

**Pattern (#154):** the fix injects a `git ls-files` trackedness probe (`gitSrcUntracked`) into the
archive move-builder; the untracked → "stage dest only" branch *is* the fix. I covered the **push**
caller end-to-end in a real repo (`hermeticRepo`) but leaned on the pre-existing **merge** sweep test,
which runs on a bare `t.TempDir()` that is **not a git repo**. There every `git ls-files` errors → the
probe's conservative `err != nil → tracked` fallback fires → the untracked branch is *never reached*.
The test was green and looked like coverage, but merge's `GitInDir(mainPath,…)` probe wiring — the
exact topology the bug was reproduced on 3× — was unexercised. The fresh-context boundary review
caught it (FIX-THEN-SHIP); I added a real-repo merge regression.

**Rule:** a real git call injected behind an interface is only truly tested **inside an actual repo**
(`hermeticRepo` / init+commit) — a non-repo temp dir tests only that call's *failure* path. Probes
with a conservative on-error default (`err → safe branch`) are especially deceptive: the no-repo test
passes *because* the probe errored, masking that the interesting branch never ran. Add a real-repo
test **per caller/wiring**, not just per shared helper — the helper being covered doesn't prove each
caller's dir/closure is wired right. Same family as [[temp workspace silently no-ops]] (#79 — bare
temp dirs giving false confidence), but the failure mode here is a silent conservative-branch, not a
`cd ""` write hazard.

**Origin:** #154 close-boundary review — push had a real-repo regression, merge did not; the merge
probe closure was reachable only in a real repo.

## 2026-07-06 — A green test can PIN a footgun as intended behavior — reverse it, don't route around it (#155)

**Pattern (#155):** the bug was that `layergraph.Walk` silently skipped a declared `substrate` whose
target is present-on-disk but ships no `base.manifest`, dropping the whole transitive chain (a fresh
derivative under-compiled to a 1-action no-op). But two tests — `TestWalkPresentSkipNonLayerDep` in
*both* `pkg/layergraph` and `cmd/weave/internal/walk` — asserted exactly that silent skip as correct
("_seen_or_add drops a non-layer dep"), ported verbatim from the shell `setup.sh` it mirrored. The
footgun wasn't just un-tested; it was *pinned green*. The fix had to **rewrite those tests to assert
the new loud error**, not add a parallel case beside them. Distinguishing the two collapsed cases
(present-but-invalid → loud; genuinely-absent peer → keep the silent present-skip) was the whole fix.

**Rule:** when a bug report contradicts an existing passing test, suspect the test **encodes the bug as
intended behavior** — especially a behavior "ported verbatim" from a predecessor (the port faithfully
copied the footgun too). Grep the bug's mechanism for a test that asserts the wrong outcome *before*
writing the fix; reversing that assertion is part of the fix, and a fix that leaves the old test green
probably didn't change the behavior the user reported. Second half: a silent filter that drops
candidates usually conflates "legitimately absent" with "present but malformed" — split them (loud on
malformed, silent on absent) rather than making the whole filter loud. Same "inherited assumption"
family as [[Verify an issue's factual premises against ground truth]] (#118) — there a comment lied;
here a *test* did.

**Origin:** #155 — the two `TestWalkPresentSkipNonLayerDep` pins had to be rewritten to
`TestWalkPresentSubstrateMissingManifestErrors`; the plan-quality judge flagged them up front as the
tests that "currently PIN the silent skip."

## 2026-07-12 — Prose concepts are not PURE entities, and evaluation evidence must match its retention claim (#168)

**Pattern:** A skill-only implementation plan listed `EvidenceSource`,
`RetroFinding`, and `FollowUpRecommendation` under “Pure entities.” Those were
conceptual nouns in `SKILL.md`, not executable units with IO-free tests. The
close reviewer correctly treated the table as a false architecture contract.
The same plan said baseline outputs would be retained “verbatim,” while the
evaluation artifact kept only excerpts and asserted ledger rows.

**Rule:** In a prose/process-skill plan, do not manufacture code-shaped PURE
entities to satisfy a planning template. If behavior exists only when an agent
loads a skill and reads evidence, classify the skill honestly as an integration
surface and test it with fixed inputs plus independent fresh-agent evaluation.
Likewise, choose the evidence-retention contract before testing: if the plan
says verbatim, retain complete worker/scorer outputs; if bounded excerpts are
intentional, say so explicitly and do not claim independent replayability.

**Origin:** #168 whole-issue close review.

## 2026-07-13 — A manual shadow sweep does not enforce a promised single source (#163)

**Pattern:** #163 centralized the issue-filename grammar and its current source sweep
proved every named consumer derived from the helper. The implementation was correct,
but the durable plan also promised an automated guard that would fail if a future
consumer copied the six-digit literal or bypassed the shared parser. Behavioral tests
alone remained green under exactly that architectural regression.

**Rule:** When a change's purpose is a single source of truth (ARCH-DRY/ARCH-PURPOSE),
turn the shadow sweep into an automated source guard before checking the plan item.
Assert both halves: the canonical production literal occurs once, and each named
consumer references the canonical constant/helper. A one-time `rg` proves today's
diff; a checked-in guard defends the invariant from tomorrow's drift.

**Origin:** #163 whole-issue close review — implementation passed the manual sweep,
but the promised structural regression test was missing.

## 2026-07-14 — Gating a commit on a piped test run gates on the pipe's LAST command (#175)

**Pattern:** `go test ./pkg/ 2>&1 | tail -1 && git add … && git commit` committed on a
FAILING suite — the `&&` reads the pipeline's exit status, which is `tail`'s (0), not
`go test`'s. The failure was visible in the printed output ("FAIL") but the chain ran
anyway, producing a bad commit that later needed a history rewrite (soft-reset +
recommit) to deduplicate.

**Rule:** Never pipe the command whose exit status gates the next step. Run the gate
bare (`go test ./pkg/ && git commit …`) and do any filtering on a second read; or use
`set -o pipefail` if a pipe is unavoidable. Same family as the #115 "don't truncate
judge output" rule — filtering a load-bearing command's output also swallowed its
load-bearing exit code.

**Origin:** #175 implementation session — the golden-drift test failure was piped
through `tail -1`, so the Task-5 commit landed before the golden was regenerated.

## 2026-07-15 — os.Getwd returns the logical $PWD; git returns resolved paths — EvalSymlinks before comparing (#179)

**Pattern:** `sdlc migrate`'s inside-repo guard compared `filepath.Abs(file)`
(built on `os.Getwd`) against `git rev-parse --show-toplevel`. Go's `os.Getwd`
prefers the `$PWD` env var — a LOGICAL, symlink-preserving path — whenever it
stats to the same dir; git always returns the RESOLVED path. Under a shell cwd
that crosses a symlink (macOS `/tmp` → `/private/tmp`, any `ln -s`), the two
disagree on a prefix and `filepath.Rel` "proves" a file is outside its own
repo. The hermetic suite couldn't see it: in-process `os.Chdir` doesn't update
the `$PWD` env var, so tests silently exercised the resolved-path branch only.
Found by the plan's mandatory live-dogfood step on a scratch fixture.

**Rule:** when comparing a getwd-derived path with a git-derived (or any
syscall-resolved) path, `filepath.EvalSymlinks` BOTH sides first. And when a
test needs to reproduce shell-launched path behavior, set `t.Setenv("PWD",
<symlinked form>)` explicitly — chdir alone tests only half the reality. Same
family as the #44 "IO needs a live run" lesson; the specific new tripwire is
that `os.Getwd` has an env-dependent branch your tests won't hit by accident.

**Origin:** #179 live dogfood (`sdlc migrate` refused a file that was plainly
inside the repo). Fixed with EvalSymlinks + a $PWD-setting regression test.

## 2026-07-16 — User strings need serialization tests; slug containment needs one resolver (#180 M3)

**Pattern:** A model-derived project scaffold interpolated CLI prose directly
into YAML, so colons broke parsing and `#` silently changed the stored value.
Separately, `project new` checked its slug locally while show/validate/set-status
joined unchecked slugs to the project directory, allowing `../` traversal.

**Rule:** Treat every user-provided frontmatter value as serialized data: encode
it with a format-safe scalar writer and test hostile punctuation plus a real
parser/validator round trip. For slug-addressed files, expose one canonical
slug-to-path resolver that validates grammar and containment; every read and
write verb must call it, with traversal tests proving no downstream IO runs.

**Origin:** #180 M3 boundary review (REWORK).

## 2026-07-16 — Modeled guards are a closed execution list; calibration needs complete inputs and staged writes (#180 M4)

**Pattern:** A dedicated close verb looked up the modeled transition but then
recognized only one guard by name, silently ignoring another modeled guard and
any future addition. Its fog ledger also summed only the issue actuals it could
read, producing a precise-looking but false partial calibration row, then
archived the project before the sibling ledger write could fail.

**Rule:** When a model names ordered guards, iterate the entire list through an
explicit handler registry and fail closed on every unknown name—never use the
model only as an edge check while shadowing its guards in conditionals. A
calibration row is valid only when every required input is measured; otherwise
refuse or record a visibly non-calibrating `n/a` under an explicit bypass. For
one command changing multiple durable records, stage every output first and
test a forced late-stage failure leaves all original records unchanged. At each
user-facing milestone, add runnable README examples in the same boundary.
If the schema accepts YAML, consume it semantically through one typed decoder;
flat regex/line readers are not schema consumers because quoted scalars and
block lists are equally valid data. Parsers for guarded evidence must preserve
absent vs invalid states so only genuine legacy absence can take an explicit
bypass.
Numeric validation must reject NaN and infinities explicitly—ordinary
comparisons do not reject NaN (`NaN <= 0` is false), so “positive” checks alone
can admit a durable false calibration value. Core-concepts status must describe
the current boundary (`planned M5` is not `new`) so review does not confuse the
roadmap with delivered code.
Calibration input sets must also be sets by canonical identity, not merely YAML
lists: normalize aliases to the durable entity key and reject duplicates before
aggregation or mutation, or one logical input can silently acquire extra weight.
Canonicalization at a degraded-input boundary must be best-effort: it may catch
more aliases when peers are present, but must not turn an already modeled
unavailable input into a new unconditional failure that defeats an explicit
non-calibrating bypass.
Derived dependency graphs must key vertices by the same canonical entity
identity used at mutation/calibration boundaries, never by display spelling;
otherwise aliases invent parallelism and duplicate rows inflate forecasts.
When inserting into line-oriented records, never derive byte offsets by adding
an assumed delimiter per split element: the last element may end at EOF. Keep
the transformation pure, parse structural headings fence-aware, and rebuild
from line slices so valid newline/no-newline forms are equally safe.

**Origin:** #180 M4 boundary review (REWORK).

Verification evidence must come from a binary built from the code under
review: a manually-built copy (e.g. /tmp/claude/sdlc) goes stale the moment a
FIX-THEN-SHIP bundle amends the behavior it exercises, and a stale binary can
"demonstrate" exactly the bug the bundle just fixed as if it were designed
behavior. Rebuild immediately before capturing evidence, or invoke the `sdlc`
shell function (which rebuilds from source every call).

**Origin:** #171 M6 boundary review (FIX-THEN-SHIP) — the Step-6 read-path
evidence recorded a "multi-match" produced by the pre-M4-fix prefix-match bug.

When verifying a consolidation / dedup sweep is complete, the acceptance grep
must match EVERY signature of the duplicated behavior, not one proxy for it. A
partially-delegated copy — a test fixture whose SETUP already delegates to the
shared helper but whose RUNNER does not — carries neither full signature, so a
gate keyed on a single marker (the setup's `commit.gpgsign` line) passes over
it silently. Grep each distinct idiom independently (setup AND runner:
`exec.Command("git"` across all `*_test.go`), minus the documented locals.

**Origin:** #186 close review (FIX-THEN-SHIP) — the rg gate keyed on
`commit.gpgsign` missed `runGitCommand`, a byte-identical `testfix.Git` twin
whose repo came from an already-delegated `hermeticRepo` (no gpgsign line).

When you change a thing, the artifacts that *describe* it are consumers too — the
shadow-sweep must include them, not just the code that calls it. Two misses of exactly
this shape landed at both boundaries of one issue: a TSV whose header line still declared
the old column set while new rows carried the new one (data present but unlabeled), and a
noun registry page that never gained the noun. Both were skipped for the same reason — a
"who consumes this?" sweep naturally follows call sites, and a file that states the schema
has no call site. Add two named checks to the sweep: **the artifact's own header/schema
line**, and **the registry/index page for its kind**.

A corollary worth its own habit: a conformance test that ranges only over
model-derived slices passes **vacuously** when the model exports empty, and its negative
assertions ("an unmodeled value is rejected") are satisfied by an empty model too. Such a
test needs one assertion with a literal in it — a length check, or a named member — or an
independent gate that the export is non-empty.

**Origin:** #187 close review (FIX-THEN-SHIP) — I1 (`estimate.UpgradeHeader`) and I2/I3
(`vet_test.sh` + `atlas/workflow/vocabulary.md` missing the `finding` noun, where
`TestFindingConformance` would have passed on an empty export).

A consolidation sweep must scope its verification to the LANGUAGE BOUNDARY of the thing being
consolidated, not to the language you happen to be working in. A grep proving "one encoding
remains" across `*.go` said nothing about the same pattern in a `.py` sibling — and that sibling
was a live fallback path, symlinked into every derivative repo. The tell was available: the file
was in `construct/base.manifest` and invoked by a Makefile target. When retiring a duplicated
rule, grep the manifest and the Makefiles for other implementations of it before claiming the
count.

Related but distinct from the two entries above (#186's "match every signature" and #187's
"artifacts that describe a thing are consumers too"): those are about missing a *consumer* or a
*shape*; this is about missing an entire *implementation* because the search was language-scoped.
The three together suggest one habit: before asserting a sweep is complete, write down what
class of thing you searched and ask what class you did not.

**Origin:** #190 close review (FIX-THEN-SHIP) — `scripts/close-issue.py:212` was a sixth
encoding of the ref grammar that the Go-only grep declared consolidated.

A scripted text edit must anchor on a string that is UNIQUE in the file, and prose about a
document's structure is not unique. Rewriting a section with `s[:s.index("## Plan")] + new +
s[s.index("## Log"):]` spliced an issue file in half: a Done-when bullet contained the words
`` `## Log` `` describing where a decision should be recorded, so `index` matched the prose,
not the heading — deleting the tail of one section and leaving a stale duplicate `## Plan`
behind it. The file then had two contradictory Plan sections and survived three commits
undetected, because the readers that matter happened to be forgiving: `issue.PlanSectionRE`
takes the first match and `insertLogLine`'s `^## Log\s*$` rejects a mangled heading and takes
the last.

Anchor on a line-anchored pattern (`^## Log$`) or split into lines and match exactly — never a
bare substring. And when an edit rewrites a *region* rather than replacing a *token*, print the
seam afterwards: the corruption was 18 lines long and one `grep -n '^## '` would have shown two
`## Plan` headings immediately.

The general rule: a self-referential document — one whose prose quotes its own structure — makes
every substring anchor ambiguous. Issue files, plans, and skill definitions are all this kind of
document.

**Origin:** #194 plan-quality gate round 1 — PQ-1 (Critical), caught by the gate rather than by
me, three commits after the splice.

When a MECHANISM moves, grep `atlas/` for the old mechanism's name in the same commit.
Twice in one issue, a doc was left asserting the contract the diff had just replaced:
`sdlc-binary.md` still said a close refuses "if HEAD … changed" after HEAD-identity was
replaced by delta classification, and `gate-state.md` still said "`code-review.md`
instructs the reviewer to read the ledger's `## Open findings`" after that section was
deleted in favour of seeding. Both were caught by a boundary reviewer, not by me.

The failure mode is specific and worse than a stale doc: a paragraph that *describes the
replaced behavior as current fact* is trusted by the next reader, so it actively
misinforms rather than merely lagging. Code has a compiler to catch the equivalent; prose
does not, and "I updated the docs" is satisfied by touching *a* doc.

The check is mechanical: before committing a behavior change, `grep -rn "<the old
mechanism's name>" atlas/ cmd/*/helptext/` — searching for what you REMOVED, not for what
you added. Searching for the new name finds the file you just edited; searching for the
old one finds the files you forgot.

**Origin:** #194 M1 review (I1) and #194 M2 review (I3) — the same family, one milestone
apart, which is the recurrence #194 M3's family escalation exists to name on the second
instance rather than the third.

A fix for a gate finding is complete only when a test FAILS WITHOUT IT — verified by
reverting the fix and watching it go red, not by inspection — and the verification is
recorded. On #194, four fixes across two milestones shipped with tests that passed either
way: reverting `Family: f.Family` and reverting `r.N >= round` to `r.N == round` both left
`go test ./cmd/sdlc/...` fully green. Each finding had *named the test to add*; the tests
were written, and they did not actually pin the behavior.

Inspection is not the check. A test written from the same mental model that produced the
fix will assert whatever the fix happens to do, including nothing. The revert is a
different question — "does this test distinguish the two worlds?" — and it takes about
fifteen seconds: copy the file, undo the fix, run the one test, restore.

Applies to any claim that a change is *tested*, not just gate findings. The bar is cheap
enough that "I checked it looks right" is never the better option.

**Where the revert cannot answer**, and what to do instead. Some fixes have no failing
test by construction, and pretending otherwise just produces a test that asserts nothing:

- **A fix that removes a possible divergence** — collapsing a duplicated algorithm onto
  one source, deleting dead coverage. Reverting restores a second implementation that
  behaves identically today, so nothing goes red. Assert the *property* instead (one
  definition site; `grep` finds exactly one), or accept that the change is structural and
  say so in the commit rather than claiming a pin.
- **A fix that makes an inert thing reachable.** Reverting is not the check — the check is
  that the thing was inert *before*. Pin the wiring, not the field: assert through the
  path production takes, then revert.
- **A pure deletion.** If the invariant is covered elsewhere, cite where; if it is not,
  the deletion needs a replacement test, not a revert.

The unifying question is not "does a test go red?" but **"what observation distinguishes
the fixed world from the broken one, and does something make it?"** The revert is the
cheapest instrument for that question, not the question itself.

**Origin of this refinement:** #194 close review BR-38, raised in the same round that
caught two inert fixes — the rule and its own limits found together.

**Origin:** #194 M2 review (C1, where the revert WAS done and caught a real gap) and #194
M3 review (BR-31, where it was skipped four times — measured prevalence 4 instances plus
one assertion that tested an unreachable state and sliced past the end of a string).

Sharpening the rule above, because reverting the fix cannot catch this case: **an
assertion nested inside a runtime guard is unfalsifiable.** `if cond { assert }` silently
passes whenever `cond` is false, and mutation-verification does not detect it — the test
still goes red, on its *other* assertions.

Write the guard as the assertion instead:

    if !cond { t.Fatal("precondition: ...") }   // fails loudly if the state never arises
    assert(...)                                  // then assert unconditionally

or build the fixture so no guard is needed. `if cond { assert }` is legitimate only when
`cond` IS the thing being asserted.

The tell in #194: two successive "repairs" of the same assertion, the second claiming in
its comment a fix that had not happened, both nested under
`if strings.Contains(prompt, "OPEN FINDINGS")` — in a fixture where that section is never
emitted. Probing beat inspecting: inserting a `t.Fatal` ahead of the guard proved in one
run what two readings had missed. When a guarded assertion is suspect, make it fail on
purpose.

Corollary worth its own beat: the fix there was **deletion**, not repair. The invariant
was already covered unconditionally elsewhere, so the guarded copy was dead coverage that
read as protection — worse than no test, because it occupies the space where a real one
would be noticed missing.

**Origin:** #194 M3 review BR-34 — 3rd instance of the family, escalated from "add a test"
(round 5) to "revert the fix to verify it" (round 6) to this.

## Repository-rooted Git pathspecs still require repository-relative configured paths

**Pattern (#162 close review):** A review recipe correctly used `git -C <root>`,
but copied absolute `--issues-dir` and `--history-dir` values into exclusion
pathspecs. Git interprets tracked pathspecs relative to the work tree, so the
commands ran successfully while silently reviewing files promised as excluded.
The live conformance test only resolved refs and inspected manifest fields; it
never executed the rendered argv, so it could not observe the semantic failure.

**Rule:** Normalize configured in-repository paths at the IO boundary before
passing them to a pure Git-command renderer, and make the renderer reject
absolute or escaping paths. Conformance for generated command recipes must
execute the structured argv against adversarial repository state and assert
both inclusion and exclusion; resolving refs or comparing command text proves
shape, not behavior (`ARCH-MOCK`, `ARCH-PURPOSE`).

Apply those assertions to **each recipe**, not to the command set in aggregate.
A loop that executes four commands but positively checks only two still allows
the unchecked commands to become empty no-ops. Mutate each generated recipe to
an empty range; its own assertions must fail.

The same claimed-fix rule applies to user-facing documentation: if a boundary
finding requires a README surface, add a section-scoped repository contract for
the commands and semantics the finding names. Correct prose without a deletion-
sensitive test is still an unguarded fix at a convergence gate.

## An unlocked review snapshot must enumerate every mutable artifact exposed in its prompt

**Pattern (#162 close review):** The close transaction snapshotted issue and
project prose before releasing its lock, but a new manifest also named the
canonical durable plan. The plan was omitted from the staleness snapshot, so it
could change while the reviewer ran and stale evidence could still finalize.

**Rule:** Whenever a review prompt gains a mutable file, update the unlocked
transaction snapshot in the same change. Capture both presence and contents—an
absent optional file is state too—and revalidate after reacquiring the lock.
Exercise modification, creation, deletion, and replacement through every caller
that unlocks around review; helper-only coverage does not prove wiring
(`ARCH-PURPOSE`, `ARCH-DRY`).

## A tagged error envelope needs a validator per surface

**Pattern (#200 close review):** A shared diagnostic struct checked only that
`code` was non-empty. That preserved envelope shape while allowing arbitrary
typos and allowed an inventory capability to carry refusal variants that only a
prospective-path result could produce.

**Rule:** For each public tagged envelope, enumerate its legal diagnostic codes
at that envelope's validation boundary. Test both encoding and decoding against
every legal member, an unknown typo, and every modeled code owned only by a
sibling surface. A shared struct does not imply a shared union (`ARCH-PURPOSE`).
