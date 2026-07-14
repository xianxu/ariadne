# Boundary Review — ariadne#173 (milestone M3)

| field | value |
|-------|-------|
| issue | 173 — introspect ingest codex transcripts |
| repo | ariadne |
| issue file | workshop/issues/000173-introspect-ingest-codex-transcripts.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | e3f89feb3712c57e8be99765989691635dbee985..HEAD |
| command | sdlc milestone-close --issue 173 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-07-14T07:23:19-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3's two code fixes are correct and, unusually for a "measurement" milestone, I could verify them independently against the real corpus rather than take the Log's word: 40 of 592 codex rollouts carry `forked_from_id`, 39 of those 40 replay their parent's opening user message, and the run caches reproduce the atlas's headline numbers exactly (pre-fix 823 substantial moments with 112 friction/198 endorsements → post-fix 202 with 12 friction/8 redirects). Seven unit suites pass. What blocks SHIP is not the code but the **atlas spec**, which is Task 9's actual deliverable and the single source #172's Go reader derives from: it defines a fork as any rollout whose meta has `forked_from_id`/`parent_thread_id`/`agent_nickname` and claims all of them replay the parent — but 79 rollouts in the corpus carry `parent_thread_id`+`agent_nickname` *without* `forked_from_id`, carry one meta, do **not** replay anything, and are processed by the Python (contributing 8 of the 202 substantial moments). A Go implementor following that sentence would skip all 119 and diverge from Python's 40 — precisely the cross-language drift the spec-level DRY exists to prevent (ARCH-DRY). Two operator-facing docs also still assert the claim M3 disproved.

**Architecture markers.** ARCH-DRY — **flag**: the fork rule is stated inaccurately in the one artifact that *is* the shared source (Important #1), and codex wire-format knowledge (`session_meta`, `forked_from_id`) now lives in `normalize.py` rather than `agent_codex.py`, which its own docstring says "owns codex wire-format knowledge" (Minor #6). ARCH-PURE — **pass**: `_output_is_error` is pure and fixture-tested with no IO; `process_codex_file` is the declared IO seam and is tested with real temp files and no mocks, exactly as the plan's Core Concepts prescribed. ARCH-PURPOSE — **pass**: M3 delivers the issue's headline answer with traceable evidence and does not settle for the easy win; the two fixes that surfaced mid-dogfood were folded in rather than deferred, and the finding is reported honestly (the fixes only strengthen the "no").

## 1. Strengths

- **The fork rule is empirically right, not just plausible.** `normalize.py:419-427` keys off the first meta and skips `forked_from_id`. Verified over all 592 rollouts: 40 forked files, 38 with the two-meta shape, 39/40 sharing the parent's first user message. The rule the code implements matches the corpus.
- **The finding is traceable to artifacts.** Filtering `20260714T000000-codex-m3-fixed` to substantial segments (assistant_message_count ≥ 15) yields exactly the atlas's 8 redirects / 12 frictions / 28 endorsements over 54 raw sessions, and the pre-fix run yields exactly the claimed 112 friction / 198 endorsements. The numbers in the Log are reproducible, which is what makes a measurement milestone reviewable at all.
- **`test_codex_fork_replay_skipped` (test_normalize.py:120-138) pins the bug shape, not the implementation** — the fixture puts the replayed parent meta *second* with an explicit "must NOT win" comment, so it fails if someone reverts to last-meta-wins. `test_function_call_output_benign_nonzero_no_error` table-drives four real benign shapes (rg no-match, ls, command-not-found, no-matches) rather than one token case.
- **`codex_forks_skipped` is threaded into `run.json` and stderr** (`normalize.py:528-529, 554`) — a filter that drops 40 files announces itself in the run record instead of silently changing counts.
- **`--agent both` works end-to-end** — I ran it (61 segments, correctly tagged 36 codex / 25 claude), so the mixed-corpus Done-when is real capability even though the dogfood ran the two agents separately.

## 2. Critical findings

None.

## 3. Important findings

**1. `atlas/workflow/introspect.md:167-182` — the fork spec conflates two distinct codex concepts, and the sub-agent half is factually wrong (ARCH-DRY).**

Line 169 defines the trap as "A forked/sub-agent rollout (pair, parley.nvim multi-agent runs; meta has `forked_from_id`/`parent_thread_id`/`agent_nickname`) **replays the parent session's entire transcript**", then line 180 states the rule as skipping only `forked_from_id`. Corpus evidence contradicts the first sentence:

| meta shape | files | metas/file | replays parent? | code |
|---|---|---|---|---|
| `forked_from_id` (+`parent_thread_id`,`agent_nickname`) | 40 | 38 have 2 | yes (39/40 share parent's first user msg) | skipped |
| `parent_thread_id`+`agent_nickname`, no `forked_from_id` | 79 | 1 | **no** (0/79; they have no `user_message` events at all) | **processed** |
| neither | 473 | 1 | n/a | processed |

Failure scenario: #172's Go `process-manual --session` implements the spec as written, skips all 119 rollouts, and its per-session counts diverge from introspect's by 79 sessions — the exact cross-language drift the spec-level DRY point was created to prevent. Secondarily, those 79 nickname'd sub-agent rollouts contribute 8 of M3's 202 substantial moments (all edit-after-edit), i.e. agent-orchestration signal that the spec's own rationale ("forks are agent-orchestration") says shouldn't count.

Fix sketch: split the section into two named properties — **fork-replay** (`forked_from_id` → two metas → replays parent → MUST skip, 40/592) and **sub-agent thread** (`parent_thread_id`/`agent_nickname` alone → one meta → own content, no user turns → currently processed, 79/592), with the counts. Then decide explicitly whether sub-agent threads should also be dropped (they can only produce edit-after-edit/friction, never redirects/endorsements) and say so; either answer is defensible, but the spec must state which one so both languages agree.

**2. `construct/local/introspect/SKILL.md:69-71` — the operator-facing doc still asserts the claim M3 disproved.**

> "Note: codex sessions carry MORE friction signal than claude (non-zero exit codes vs claude's harness is_error flag) — some benign (grep-no-match); the cluster walkthrough (Stage 5) filters those."

Both halves are now false: M3's whole finding is that codex does *not* carry more signal (112 → 12, ~95% artifact), and benign exits are filtered in the adapter (`agent_codex._output_is_error`), not by the Stage-5 walkthrough. An operator reading Stage 1 gets steered by a note the same milestone refuted. Fix: replace with the M3 conclusion (codex friction ≈ Claude's after the hint gate; the adapter filters benign non-zero exits) and link the atlas section.

**3. `atlas/index.md:17` — still describes introspect as "postmortem mining of past **Claude** transcripts".**

Agent-neutrality is the issue's entire purpose, and the index is the map's entry point — the one line a reader hits first still says Claude-only. M3 is the last gate before close. Fix: "past agent transcripts (Claude + codex)".

## 4. Minor findings

- `agent_codex.py:18-20` — module docstring still documents the **old** gate ("`Process exited with code N`, N != 0, **or** an `error:`/`exit code` prefix + friction hint"), where a non-zero exit alone sufficed. The function docstring below it is correct; the header contradicts it.
- `agent_codex.py:35-36` and `atlas/workflow/introspect.md:159` — "mirroring Claude's gate" mischaracterizes Claude. `agent_claude._result_is_error` is `is_err_flag OR (starts_with_error AND has_hint)` — the flag path needs **no** hint. Codex now requires a hint always, so it is deliberately *stricter*, not a mirror. The justification (codex's non-zero exit is not equivalent to a harness-set flag) is sound; just say that instead.
- `SKILL.md:87` — the `run.json` field list doesn't mention the new `codex_forks_skipped`.
- `SKILL.md:89` — "A flat `events.jsonl` stream will be added in M2/M3…" is a stale forward-reference; M2/M3 are done and the NormEvent stream stayed in-memory by design.
- `atlas/workflow/introspect.md:189` — "Over 54 root codex sessions" is imprecise: 54 is the count of *substantial* (amc≥15) root sessions; the corpus has 552 root sessions of 592 rollouts. As written it understates the corpus 10× and disagrees with the issue Log's own "592 rollouts, 552 root sessions".
- `normalize.py:419-431` — `session_meta`/`forked_from_id`/`git.branch` are codex wire-format knowledge living outside `agent_codex.py`, whose docstring claims it "owns codex wire-format knowledge" (ARCH-DRY, cohesion). One line today, so not worth churning now — but see the architectural note below.

## 5. Test coverage notes

- The two fixes are covered at the right altitude: the friction gate purely (no IO), the fork skip through the real IO seam with temp files and no mocks — consistent with the plan's declared test surface.
- **Gap matching Important #1:** no test asserts a `parent_thread_id`/`agent_nickname`-only rollout is **not** skipped. That's the behavior distinguishing the two concepts and the one the spec currently gets wrong; a three-line fixture (single meta, `parent_thread_id` set, no `forked_from_id` → `skipped is False`) would pin it and stop a future "skip all sub-agents" edit from silently dropping 79 sessions.
- `test_parity.py` still passes, but it cannot catch the Claude/codex gate asymmetry: its codex fixture output contains "operation not permitted" (a hint) and its Claude fixture sets `is_error` *and* includes a hint. A Claude tool result with `is_error=True` and no hint fires on Claude and not on codex. That asymmetry is intentional and doesn't affect M3's conclusion (Claude run-3 had 0 friction moments), but the keystone test currently overstates the neutrality it proves — worth a comment there rather than a code change.

## 6. Architectural notes for upcoming work

- **The fork/sub-agent classification wants to be one pure function in the adapter.** Once Important #1 is resolved there will be three meta kinds (root / sub-agent / fork-replay), and the decision rule is codex format knowledge sitting in `normalize.py`. A pure `codex_meta_kind(meta) -> str` in `agent_codex.py` would put format knowledge back where the module docstring says it lives, be directly unit-testable without tempfiles, and give the atlas spec a one-to-one code referent for #172 to mirror. Cheap now, and the growing enum is the trigger — not worth doing while it's a single `if`.
- **The spec-level DRY point needs a drift guard, not just prose.** #173 has now discovered three codex format properties (double-representation, derived `is_error`, fork-replay) *after* the spec was written, each time from the real corpus. Since the atlas section is load-bearing for a second implementation in another language, consider a `target` (per AGENTS.md §1) capturing "codex format spec == both readers' behavior" so the next discovery has somewhere to land, and reference it from #172.

## 7. Plan revision recommendations

The plan's 2026-07-14 Revisions entry claims more than the code delivers on one point. Recommend appending:

```
### 2026-07-14 — M3 review: fork vs sub-agent are two properties, not one

The M3 revision (and the atlas spec) described the fork trap as "meta has
forked_from_id/parent_thread_id/agent_nickname → replays the parent". Corpus
evidence says these are TWO distinct codex concepts:
  - fork-replay: forked_from_id (40/592), two session_meta, replays the parent
    transcript → skipped by process_codex_file. The 66%-inflation source.
  - sub-agent thread: parent_thread_id/agent_nickname without forked_from_id
    (79/592), one session_meta, own content, no user_message events → NOT a
    replay and NOT skipped; contributes 8 of M3's 202 substantial moments
    (edit-after-edit only).
The code's forked_from_id rule is correct; the SPEC over-claimed. Corrected in
atlas/workflow/introspect.md so #172's Go reader skips 40, not 119. Whether
sub-agent threads should also be excluded (agent-orchestration, per the same
rationale) is an open decision recorded here, not silently implied.
```
