# Lessons

Compact guidance for Ariadne's SDLC, generated contracts, and base-layer work.
Detailed incidents remain in their owning issue or review artifact.

- #503: an injected readiness callback's entry counter is not evidence that its
  caller published readiness. Cancel only after observing the persisted effect;
  force callback-in-flight and cancel-before-publication orderings with channels.
  A repository sweep found only this fixture, so the TL chose deterministic
  controls over a one-site lint rule.

- A card owner may hold more than one issue. Before a numbered command commits
  or deletes HEAD's branch, compare it with the branch derived from that issue's
  card; ownership alone cannot identify the checkout. Remote deletion must
  lease the exact tip already kept in the archive. Close may admit a renamed
  recovery branch only at the exact canonical issue tip; ancestry or a similar
  name is not proof that it holds only that issue's work (#471).
- A preflight tip check cannot authorize a later destructive ref update: Git
  hooks or another checkout may move the ref in between. Delete a local branch
  with `git update-ref -d <ref> <archived-head>` and a remote branch with an
  exact-head lease, then test a move at the last intervening hook. Keep the
  separate checked-out-worktree refusal: `update-ref` does not enforce it, while
  `git branch -D` did. When branch admission changes, update README's usage
  examples as well as help and atlas
  (#471 BR-1/BR-2).

- Before counting joined records, check that the identity is unique across the
  whole cohort. Per-record card/event/Git validity did not prevent two issues
  from claiming one PR, branch or merge and doubling a weekly denominator.
  Reject every claimant on a collision, and test each identity in both input
  orders (#392 BR-7).
- A metric with a physical window and a cross-window cohort must classify
  ownership across the entire cohort before using the window for diagnostics.
  An ambiguous attempt just before the week must still null the headline ratio;
  otherwise a partial numerator can look exact (#392).

- When a bug fix adds a diagnostic flag or revises an identity policy, enumerate
  the full input matrix in Done when and committed fixtures, then check the
  README for the new user-facing command. A motivating fixture alone can pass
  while adjacent identity forms remain undocumented or untested (#403).
  Atlas coverage does not replace README usage: shell-owner changes also need
  the exact alias pickup/relaunch step beside the diagnostic command (#461 BR-1).
- A diagnostic flag that promises stable machine-readable output needs a
  byte-for-byte stdout comparison and exit-status check with the flag off and
  on. Capture stderr separately; discarding stdout proves only the message,
  not the interface contract (#403 BR-4).

- Correcting a liveness fake is not proof that reclamation is safe. Pause a
  stale observation until another holder acquires, then assert its generation
  survives. Serialize reapers and revalidate before the destructive rename;
  faithful scheduling exposed a production race behind the original test flake
  (#413).

- CLI fakes must locate named option operands and model read-only queries.
  Adding `-trimpath` shifted Go's `-o` operand, while `go env GOFLAGS` was
  incorrectly counted as a build. Preserve assertions about real build effects
  while extending the fake's command model (#386).

## Contracts and single sources

- A prose policy becomes an integration contract when code reads the repository.
  Put the contract in one referenced source, give it a machine-readable token,
  and make consumers gate on that token rather than prose presence.
- Generate projections from the source model. A hand-maintained copy of generated
  data, an enum registry, a plan table, or an atlas list drifts silently.
- A declaration nothing reads is a comment with a type. Trace every field,
  capability, and status through construction, persistence, and consumers.
- A target can lie by aspiration. Mark only mechanisms actually implemented and
  tested; do not generalize a high-clarity proof to unbuilt siblings.
- When merging or removing concepts, hunt every consumer before deleting a
  symbol. A follow-up issue must not offload the current issue's purpose.

- Moving code to a new package, or adding a `cmd/` binary, changes two
  populations nothing recomputes. Source guards that scan a directory list
  silently stop covering the moved code: their exemptions name the old file,
  and their scope check walks only the old tree. Prose that enumerates the
  binaries (README, atlas) goes stale. Before the first review, grep the guards
  for the old path and the docs for a sibling binary's name (#350 BR-1).

## Guards and proof

- Suite output contains quoted subprocess output too. Preserve indentation before
  tokenizing control records. Raw child output can overlap real ancestry even
  with correct four-space levels: prefer a proven top-level failure to an
  invented leaf. Keep exact leaves only from structured evidence, and expose
  the possible subtest-only registration miss. Validate the whole leading
  whitespace prefix with the tokeniser's Unicode category, accepting only ASCII
  spaces; initial spaces alone miss a following tab or NBSP. Use the same
  column-zero boundary for package attribution, unknown-failure reasons and
  timing facts. Derive unidentified-package failures after collecting structured
  names. Test real Go output, retained-log readers and the actual retry (#476).

- Whenever artifact creation gains a fallback root, enumerate that root in
  stale cleanup too, including after the original root becomes usable again.
  Test stale removal and fresh preservation at the repaired location; normal
  exit cleanup alone does not cover a killed creator. Enumerate sibling
  artifact families in the same repair round: fixing checkouts alone missed
  per-run directories under the same former fallback (#477 BR-1, BR-2).

- Temp fallback must replace inherited child TMPDIR and GOTMPDIR and repair
  checkout creation before child startup. Probe actual write access, not mode
  bits. Keep startup environment evidence outside code-suite results so a
  same-head repair can pass without hiding an earlier genuine failure (#477).

- A test that invalidates inherited temp variables must supply its fallback
  candidate through the IO seam, not depend on the host /tmp or user-temp
  directory remaining writable. Complete unrelated checkout setup first, then
  test real child env/writes and a denied-selector no-child/no-success control.
  Reproduce the permission boundary with a real sandbox (#487).

- If a contained reviewer cannot launch a nested OS sandbox, capture host-side
  acceptance as measured evidence: source revision, profile digest/content,
  probe source, denied host writes, allowed fixture writes, test commands and
  exit statuses. A prose claim that sandbox tests passed is not independently
  reviewable evidence (#487 BR-1).

- Keep the real checkout stable while cmd/sdlc package tests run: its TestMain
  snapshots HEAD, branch and porcelain status. A concurrent commit or a new
  dirty path invalidates the run even when every assertion passes (#472).

- Native CLI fixtures that assert quiet output must explicitly isolate live
  runtime context at their shared subprocess boundary. Set inherited Couch
  context in the regression so CI outside Couch also detects a leak; keep
  dedicated Couch diagnostic tests explicit (#461 suite failure).

- When an approved revision changes acceptance, synchronize the current Done when
  clause and preserve the original wording in Revisions. Leaving contradictory
  current criteria makes the close evidence stale even when the revised live
  observation is valid (#418 BR-1).

- Choose focused tests by the changed helper's callers, not just a shared test
  name prefix. The landing selection change also affected IntegrationPRSync;
  omitting that caller's controls delayed an obsolete diagnostic expectation
  until the full suite (#418).

- A strict identity check is defeated when a fallback or polling loop drops one
  of its binding fields. Carry the current branch head through every PR selector,
  retain competing OPEN records regardless of head, and test integration/recovery
  with historical completed records still in the external service (#418).

- A selected repository path is data, but Git pathspecs are patterns. Pass
  selected filenames with literal pathspec syntax at every Git operation,
  including status and cleanup; a wildcard filename otherwise expands to
  unrelated dirty files (#240 BR-5).
- Cleanup guidance must account for Git tracking state: `git restore` cannot
  remove an untracked file before a merge. Test the printed command against a
  real Git checkout, including a wildcard filename beside another match
  (#240 post-review).
- A cleanup that reads bytes and later restores or removes a path has a race
  between its check and its effect. A rename and recheck do not close the race:
  an editor can still write through an already-open descriptor immediately
  before the renamed inode is unlinked. Model every interrupted effect and keep
  the old inode recoverable for as long as late writes are promised safe; test
  writes at the final destructive boundary, not only after the move (#240 BR-1,
  BR-6).

- Before shortening a command's transaction lock, enumerate every path that can
  run a suite and every ref-writing fetch, including legacy and recovery paths.
  The #370 plan review caught a second merge path whose local CI fallback would
  otherwise lose its lock entirely. Probe a second same-repo command against
  the real lock during both first and repeated suite runs.
- A validation window needs one pinned commit from the first gate through the
  irreversible effect. Re-reading a remote ref after an unlocked gate can make
  later CI checks authorize code the earlier gate never saw (#370 BR-1).
- After polling an external service for a pushed head, fetch and check main
  again before an expensive suite. Main may advance during registration, and
  detecting it only after the suite wastes the full run (#370 BR-2).
- Recovery preconditions must account for effects a completed earlier phase
  deletes. A merged PR can remove its remote branch before archive resumes;
  prove the shipped local head against main before demanding a remote ref
  (#370 BR-3).
- A concurrent test fixture owns its worker through teardown. On every exit,
  release blockers, cancel waits and join the worker before restoring cwd or
  global adapters; a signal write alone leaves a race (#370 BR-4).
- A context-free compatibility wrapper is still production code to a source
  guard. Remove unused wrappers or route their nil context through the one
  bounded fallback; keep fixture aliases in test files. A targeted behavior
  test will not exercise a whole-package source guard (#332 full suite).
- A reentrant resource acquisition must return the original owner's identity,
  not merely permission to continue. The second local suite run reacquired its
  lock but omitted the holder pid from its child environment, so nested make
  waited on its own process (#332 BR-15). Test repeated runs with the real
  stateful lock helper; a fake that always says "skipped" cannot expose this.
- Run a recorded suite in a fresh checkout before treating it as landing
  evidence. An ignored woven skill directory made a process-manual test pass
  in the slot and fail in the clean suite worktree; assertions about generated
  projections must be conditional or backed by an explicit fixture (#332).
- When a stateful fake models a remote push, update its advertised PR head
  from the pushed ref in every fixture that can integrate main. A stale fake
  turns a successful integration into a two-minute registration timeout
  (#332 full suite).
- An override of execution settings must not override a mandatory gate's
  requirement. Test omitted and explicitly false flags through the production
  gate, as well as the pure derivation (#332 BR-8).
- A failed parent lookup does not prove a root commit. Read positive parent
  evidence and propagate operational errors; test those errors (#332 BR-3).
- CommandContext alone does not bound a subprocess whose descendants retain
  output pipes. Test cancellation with a real descendant and bound cleanup and
  pipe draining (#332 BR-5).
- A termination assertion must expire before the fixture child can exit
  naturally; a two-minute wait for a one-minute sleeper proves no cleanup.
  Mutation-check both timeout and parent-exit cases (#332 BR-10).

- In a rejection table, each row must assert that it was refused by its own
  check (for example, the error names the field). A shared invariant failing
  first, such as a nil collection, makes every row pass vacuously (#279 BR-8).
- A scripted edit (search and replace) must assert that its anchor matched.
  An unmatched replace silently does nothing, and the work log then claims a
  change that never happened (#279 BR-8).
  Anchor on content, never on a line number: a `sed -e 'Ns/…/…/'` whose line
  drifted by one matched nothing, twice in one issue (#325).

- A predicate that collapses several refusals into a bool forces its caller to
  guess the reason, and the guess becomes a wrong message. Return the reason as
  an error and share the one probe between callers (#296 BR-2).
- When one effect is added at several call sites, each site needs a test that
  fails when that site alone is removed. A mutation that removes them all at
  once passes as long as any one survives, and a later boundary can mask an
  earlier one (#286 BR-2, BR-6).
- When a fix changes one reader of an outcome (a stored flag), compute the
  outcome once and have every reader consume it: the reported line, the waiver
  stamp and the convergence summary too. A fix to one reader leaves the others
  reporting the old answer (#300 BR-9).
- Ticking a Plan row checks it: every test the row names exists under that
  name and asserts each clause, and a step not yet done (a close) sits in its
  own unticked row (#300, three findings in one family).
- An effect a function performs before a fallible step belongs in its error
  too. Returning it only in a result that callers drop on error loses it: weave
  moved an edited file aside, then failed, and nothing named where it went
  (#281 BR-1). Wrap every later error with the effect, once, in the function.
- When delegating to a transaction that keeps partial progress on error (weave
  refresh), read what happened from the state it acts on, not from its exit
  status. A failure can follow real effects (#303 BR-1).
- Changing who a guard admits changes every sentence that states the old rule.
  Grep the policy's words (help, recovery catalog, comments, refusal text) across
  every verb that offers the guard's next action, not only the guard's own help:
  the restore help still told a refused user to claim (#274 BR-1). Then state the
  policy once (a help placeholder) and add a wording test over help, catalog and
  atlas, or the next review finds the next copy (#274 BR-4).
- Do not `filepath.Glob` a path joined from a directory you did not write: `[`, `*`
  and `?` in the repository path become wildcards and a present file reads as
  absent. List the directory and match the base name (#235).
- A step that reruns to converge must not lease on local evidence of remote
  state another flow may have changed since: a remote-tracking ref kept for a
  branch the landing deleted made every rerun's leased push refuse (#305 BR-1).
  Refresh the evidence from the remote first, and decide every refusal before
  the first mutation (switching branches included, BR-3). When a needed
  mutation (a fast-forward) must come first, every refusal after it names it,
  through one helper, and checks that do not need it move above it (BR-6).
  Check a restore's postcondition before publishing or reporting success, and
  make it test the step's effect, which the wrong starting state cannot
  already satisfy: a landed branch has the live details too (BR-5, BR-7).
- A guard must fail closed on malformed, unknown, unsupported, and ambiguous input.
  A catalogued flag may waive only the refusal named by that catalog entry.
- Scope a guard's failure to the input it actually reads. A fail-closed check that
  decodes a whole record trips on an unrelated malformed field, and the failure
  changes behavior for records the guard was never about (#326 BR-1).
- A syntactic guard cannot support an absolute semantic claim. Bound the claim to
  the language actually inspected and state what remains outside it.
- Distinguish what a guard computes from what it takes on faith. Negative scans
  and blind reads cannot prove absence; preserve query failure as failure.
- A guard test must have teeth: mutation-check the guard and make the test invoke
  the production entry point. A helper test or a happy end state can be vacuous.
- Enumerate the real population from the tree, not from the finding's examples or
  a comment. When a helper is extracted, prove its caller still traverses it.
- A gate the agent can skip is not a gate. Put the critical transition in the
  binary and make its refusal actionable without a bypass.
- Optional evidence stays optional only when unavailable is distinguishable from
  unsupported and from a negative result.

- When a new path replaces an old one, list every case the old one decided and say
  what the new one does with each. The old peer-write had a brain row ("write in
  place, nous commits it"); the origin publish classified a brain's legacy project
  as a peer and silently dropped its tick (#308 BR-7). A decision table being
  deleted is the checklist for its replacement.

- When a verb starts writing an artifact that only a person used to write, re-read
  every existing reader of that artifact. Readers written for the old producer may
  now match cases they never saw: once the binary committed a `--no-judge`
  milestone's not-run trailer, the legacy boundary grep took a skipped review as a
  review boundary (#304 BR-2).

- Define a main-relative check on the branch patch (`merge-base(main,HEAD)..HEAD`) or
  on the reviewed patch replayed onto today's main, never on commits after an anchor.
  Merging or rebasing main changes commits, not the patch. A commit count read every
  integration as unreviewed work, and an ancestry test silently disowned a rebased
  close (#304).
- A per-process cache of remote state needs the local view in its key. Keyed on the
  checkout alone, it returned the main of the first call after a later fetch (#304).

- Refs under `refs/` are shared by every linked worktree (every slot). Never judge
  whether a ref is live from one checkout's files: delete exactly what this operation
  ended, or decide from shared state such as the tracker's cards (#304 close review).
- A source guard that greps for a call must skip its own file. The guard names
  the call it looks for, so it satisfies itself for its own directory, and that
  directory's real TestMain goes unguarded. Prove a guard by deleting the call
  it protects and watching it go red (#324 BR-5).
- A wait inside a test binary spends `go test`'s timeout: cmd/go kills the binary
  at `-timeout` plus a minute from its start, however long it queued. Bound the
  wait below that and shrink the binary's own timeout by it (#324).
- Evidence read from a shared list must be matched by provenance, not only by
  name. `gh pr checks` mixes commit statuses and check runs, so a gate keyed on
  a check's name accepts any job that happens to carry it. Say which kind of
  record counts (a commit status: no workflow, no Actions job link) and test a
  look-alike of the other kind (#345 BR-2).
- A tightening fix needs a test that the loose version fails. #345 narrowed
  `--rev` from "any commit on a remote branch" to "a branch tip", but only the
  unpushed case was tested, which both versions refuse; the pushed-but-not-tip
  case is the one that tells them apart (#345 BR-3).

- A check that needs a resource for only some inputs reads it lazily, at the
  first input that needs it. #348's issue validate read the area declaration
  up front, so a repository with no main failed every issue, including ones
  carrying no area at all (BR-1). Pass a memoized reader, not the value, and
  type its errors so a gate can still fail closed when the resource is needed.
- Values a writer puts into YAML go through one encoder that quotes what YAML
  would read as another type. #348 wrote areas unquoted, so a declared `2026`
  or `true` became a number or a bool that the card's string check refused,
  while the gate's decoder read it as a string (BR-2). Test a round trip with
  type-ambiguous values.

- Prove that a test builds the state it names, before trusting it as a
  regression test. #361's first zombie-group test used `wait4(WNOWAIT)` to wait
  without reaping, but Darwin ignores `WNOWAIT`, so the child was reaped and the
  test passed with the fix reverted. Assert the precondition the kernel reports
  (EPERM for a zombie-only group), then revert the fix and watch the test fail.

## Paths, processes, and integration

- A before/after measurement needs a baseline binary built from the baseline
  commit. In a slot, `sdlc` is a shell function that rebuilds from the slot's
  working tree on every call, so "the installed binary" is whatever is
  checked out. Build the old version from `git archive <sha>` (#254).

- The test hermeticity guard's "leaked .git/sdlc.lock" can be another slot's
  verb: every slot of a repository shares one Git common dir, so the lock is
  fleet-wide. Rerun the named tests alone before blaming the diff (#307).

- Resolve a verb's repository from the path it was given (`--issues-dir`,
  `--repo`), never from the process cwd. A cwd fallback reads, or fetches, the
  wrong repository silently, including the developer's own checkout during
  `go test` (#279 BR-1).

- A best-effort projection must not add a refusal to the verb that hosts it.
  If a cosmetic refresh's input cannot be read, the output should degrade to
  the unrefreshed bytes and stay deterministic for any proof; it should not
  fail closed and wedge a landing (#275).

- A changed path or residency includes search recipes, generated references,
  comments, and execution records. Sweep the old location, not only feature prose.
- A peer migration starts by checking branch, cleanliness, and the real fleet;
  never clean another repo or assume “nothing to migrate” from one local view.
- Command wrappers that call `os.Exit` cannot be made best-effort by their caller.
  Return errors from libraries; let the command boundary choose severity and
  perform deferred cleanup, including init races and hard exits.
- Agent-invoked CLI verbs must run headless and gate on durable state, not local
  convenience or an interactive editor.
- A read-only check must not resolve writer-side configuration (a publication
  target, an upstream) just to learn where to read. CI checkouts are detached with
  no local branch config; take the source explicitly from the caller (#266).
- A live integration test must exercise filesystem, Git, process, or network
  behavior with real data. A fake keyed to the same value shape masks IO bugs.
- A silent `0` or empty result is a footgun when it is indistinguishable from a
  real answer. Return status and provenance that explain what was observed.
- `cd` into a temp workspace must hard-guard the path; an empty variable must not
  fall through to the host repository. Glob paths directly instead of using
  `ls` in command substitution.

## Review, plans, and artifacts

- When a fix round changes a contract, grep every restatement of it (atlas,
  README, the help page of every verb whose behavior changed, plan steps and
  tables, Revisions) and update them in the same commit. Registered flags are
  checked against their FLAGS section by `TestEveryFlagAppearsInItsHelp`. A fixed rule documented the old way reads as two rules (#277).
- Absence is not evidence of a transfer. Authorize a state change from
  positive evidence written by the actor entitled to make it, never from
  "the other party no longer holds X" (#277 BR-9).

- Acquisition and later discovery must accept the same topology; test re-entry
  from every provisioned repository shape, including linked feature worktrees.
- Independent clones need clone-specific feature-worktree destinations; a fleet
  path keyed only by repository and branch collides across environments.

- New CLI surfaces need discoverable README usage as well as detailed atlas and
  embedded help documentation; check all three before closing.
- User-runnable procedures need the same README discovery link, even when they
  introduce no command. Preservation tests must snapshot both source and destination;
  snapshot after intentional fixture movement rather than skipping that ref.
- Before a boundary, read the actual verdict and reconcile the plan against the
  committed tree. “Inert at runtime” is not the same as “removed”; an unticked
  row can disable the guard that proves it.
- Multi-milestone inventories distinguish delivered entities from future files
  and modifications. Separate completed implementation checklists from the
  acceptance gate whose successful execution is still pending.
- Check generated review artifacts against the complete boundary range before
  committing; an unstaged-only whitespace check misses already committed prose.
- Review snapshots enumerate every mutable artifact in their prompt and remain
  bounded. Do not truncate judge output with `tail` or `grep`.
- A close claim names evidence at the corrected mutation boundary. A plan revision
  edits every affected row, signature, citation, and status field.
- Atlas updates land in the milestone that introduces the surface. After a symbol
  moves, grep the old location and refresh downstream links and search recipes.
- Record actual hours from the measurement tool. Commit issue/plan syncs before
  long-running reviews so design survives compaction.
- Stage explicit paths; never let `git add -A` sweep unrelated untracked WIP.
  Verify `change-code` created the intended branch before committing.

- A durable plan's Core concepts table is the current model, not history. A
  commit that adds, removes or renames an entity, or that lands a scope
  revision, updates the table and entity notes in that same commit (removed
  entities marked `deleted`). It also appends a Revisions entry that supersedes
  by name every earlier claim it voids, including ticked rows whose tests no
  longer exist. Before a close, grep the plan for each cited test name and
  identifier against the tree (#350, family plan-core-concepts-drift).

## Base-layer and data sharp edges

- Preserve raw Git path output through adapters; trimming whitespace before the
  path parser loses valid path bytes. Detect lexical environment candidates
  before canonicalization so a symlink cannot turn a refusal into generic setup.
- Private dependency scans must accept every name acquisition accepts. Global
  fleet backup/dot-directory heuristics are not a private-repository contract.

- A presence probe that follows symlinks reads a dangling link as absent. Once
  absence triggers a fallback, that silent skip becomes a bypass of an invalid
  target. Probe presence with Lstat and fail toward present (#230 BR-1).
- When "no answer" has an expected meaning and a failure meaning (not a worktree
  vs Git could not say), type the expected one and fail closed on the rest. A
  caller that treats every error as the expected case acts on a guess (#230 BR-2).
- When one rule becomes shared (one resolver for every consumer), enumerate the
  consumers from the input's readers (every caller of the parser), not from the
  plan's list. weave refresh re-resolved rows itself and broke from a linked
  worktree after compile was fixed (#230 BR-8).
- State a list of a component's consumers once, with the grep that regenerates
  it, and point every other mention at it. Three hand-kept copies of the
  resolver's callers drifted the moment a fix added one (#230, third finding in
  the doc-overclaims family).
- A lint that recognizes a construct by enumerating its contexts (what may
  precede a command) misses the next context each review. When a review finds a
  second one, invert the rule: match every occurrence and enumerate only the
  provable exceptions (#342: if/while/sudo -u/time -p after {, sudo, time, VAR=x).
- Before a lint encodes a portability rule, measure each form on the target. #353
  assumed every array operator failed on bash 3.2 under `set -u`; measured, only the
  bare `${a[@]}`/`${a[*]}` (and the bash-4 case operators) do, so the first rule
  over-flagged substrings and pattern removal.
- `--no-atlas` is a claim to verify, not a default. Before passing it, grep the atlas
  for the changed component's name and behavior words; twice (#355, #360) the close
  review found the atlas restating the old behavior after `--no-atlas`.
- A behavior change to a surface the README documents updates the README, even
  with no new command: a now-false README sentence is a finding (#230 BR-3).
- Validate external identifier grammar before interpreting sentinel values. An
  all-zero Git OID is unborn only at a valid hash length; malformed zeros must
  not bypass validation by becoming a null value.
- Editing `construct/adapted/` changes rendered downstream instructions. Inspect
  generated output and downstream consumers before landing a base-layer change.
- A span edit is only safe if its boundary is explicit and tested. A pathspec is
  repository-relative even when the configured root is absolute.
- `strings.TrimSpace` on a whole porcelain-status blob removes the first status
  column; parse fields instead of slicing columns.
- CUE schemas need whole-corpus validation, not only self-vetting examples. State
  machine gates must cover every legitimate flow before tightening a status enum.
- User strings, slugs, and symlinked paths need serialization and canonical-path
  tests. Resolve logical and physical paths before comparing them.
- Generated review sidecars and durable records need bounded sizes and explicit
  framing; raw transcripts do not belong in source artifacts.

## Working rule

When designing a gate, name the source of truth, the exact closed grammar, the
production transition it controls, and the mutation that would make its test red.
The simplest durable authority beats a clever scan of consequences.

- Git candidate reachability proves remote state, not which caller won: concurrent
  writers can construct identical commits. Reservation retries must distinguish
  actual ref updates, up-to-date responses, confirmed rejection, and uncertainty.
- Race conformance must exercise both client stale-lease rejection and server
  receive-pack ref-lock rejection; both occur in real concurrent Git publication.

- Review cancellation must be wired from actual CLI signals through command
  contexts to owned subprocesses. A background-context entry point can orphan
  a reviewer after its repository lock is released, despite unit cancellation tests.
- Before recording any review verdict, revalidate the complete prepared read set,
  including absent plans and ledger generations; rejected verdicts also write state.

- When reusing a publication guard, inspect its hidden destination and selection
  assumptions. Pass the pinned configured trunk and owned close records; an
  issue-body diff can omit independently published review evidence.
- Bind GitHub identity to Git's effective fetch and push URLs, including rewrites.
  Revalidate the selected branch and head as well as destination configuration.
- Archive retries need an intact complete generation: preserve compatible prose,
  refuse changed lifecycle metadata or missing owned artifacts, and verify the
  exact reachable archive commit before deleting local recovery refs.

- Scope cleanliness to effects: integration/switching need a clean issue checkout,
  while proven ref-only cleanup on rest can preserve new staged and unstaged work.
  Test both immediate post-switch edits and retries after each cleanup effect.

- A cancelled or output-limited Git command has no trustworthy predicate exit code;
  keep those errors distinct from completed `exit 1` absence/ancestry results.
- Match writer limits to reader limits, including framing overhead; a successful
  write must not make the next authoritative read refuse its own data.
- A bounded writer must not accidentally expose an embedded `ReadFrom` fast path
  that bypasses its `Write` limit. Test the real subprocess/`io.Copy` boundary.
- Authoritative Git blob identities require exact bytes: bypass clean filters on
  writes and archive export filters on reads; test caller and global attributes.
- Rebase overwrite checks must cover paths touched by intermediate replay commits,
  including files deleted again before HEAD, not only the target and final trees.
- When a procedure promises two history comparisons, show and test both Git
  commands. State which local tracking evidence may be stale before calling
  commits unpublished.
- Cross-repository requirements need a named issue and review boundary in the
  owning repository; a single repository's pinned diff cannot prove a peer edit.
- Tests for executable agent instructions must consume the authored instruction
  when practical; duplicating its commands in a fixture cannot catch a missing
  step in the guide.
- When a new operator phrase is added to an agent guide, update the existing
  README entry point for that workflow in the same change.
- Before submitting a milestone, enumerate the non-test files added in its window
  (`git diff --name-status BASE HEAD | grep '^A' | grep -v _test`) and reconcile
  them with the plan's Core concepts table and each Task "Files:" list; reviewing
  the inventory by memory missed six entities and two phantom files (#252 M2).
- When a verb's behaviour changes, `rg` the verb's name across
  `cmd/sdlc/helptext/` and `atlas/` and classify every hit (still true / now
  false / legacy until cutover) in the same change: per-verb help is the workflow
  contract agents read, and fixing only the pages you remember left four stale
  descriptions (#252 M2).
- Checks run before effects: a verb's dry-run and refusal paths must leave every
  file unchanged, so compute refreshes in memory and write only after the last
  check, including building any receipt or operation that validates its input
  (#252 M2: a stale-mirror refresh was written before refusals ran).
- An observation error is never evidence of absence: a Git probe distinguishes
  exit 1 (false) from any other failure (#252 M2).
- Stage explicit paths, never a directory (`git add cmd/sdlc` committed a 9 MB
  build artifact and an unrelated gofmt rewrite in #252 M3); build binaries to
  `-o` outside the source tree.
- A landing's identity is what the forge confirmed (the PR's merge commit), not
  ancestry of the branch's commits: squash and rebase rewrite them (#252 M3).
- A deferred effect replays inputs pinned when it was decided, never the live
  worktree a later commit may have changed (#252 M3) — and replays them
  three-way: only where the target still holds the version the decision saw.
  A later committed edit is newer than the pin; overwriting it is data loss
  (#252 M3 BR-30). Every pinned input the replay does not apply is reported
  (path + reason, durably where possible), never dropped silently.
- A cleanup path shared by landing and abandonment completes nothing: gate each
  completion on the confirmed landing itself, not on reaching the step after it
  (#252 M3 BR-29: removing an unmerged worktree marked its cards done).
- Every identifier in a ticked plan row greps to shipped code. When a name
  changes in implementation, rename it in the row (planned name in parentheses)
  as well as in Revisions (#252, 4th inventory-drift finding).
- Fixtures must not normalize what production callers don't: resolving a temp
  root's symlinks up front hid close comparing /tmp and /private/tmp paths
  until an e2e used the raw slot paths (#252 M4). Compare filesystem paths
  canonically on both sides (`canonRoot`).
- Before designing a transformation of existing records, run the strict
  parser over the real population read-only: a fleet probe showed a fifth of
  issue files failing it and duplicate IDs, which reshaped the migration
  (#252 M4). Then rehearse the whole command on a disposable copy.
- An artifact that pins another's identity (a blob OID, a digest) is built by
  one constructor from the source's final bytes, after its last mutation — and
  the pure plan asserts pin == identity(final). Deriving the pin alongside the
  first version and mutating the source afterwards (binding a close to a card)
  left main naming a blob the tracker never held (#252 M4 BR-34).
- An end-to-end test that skips a gate (`--no-judge`, `NoJudge: true`) also
  skips every bug behind it: the #252 slot-cycle e2e and the first smoke cycle
  both bypassed the publish gate, so the transfer guard refusing the owner's
  own edits on a direct push surfaced only in the second smoke cycle. Run the
  deterministic gates in e2e; stub only the LLM.
- "Both sides make the identical change, so the merge is clean" holds only
  until either side edits that region again; durable convergence needs the
  change as a shared ancestor (merge it), not a byte-identical copy. #252's
  first reconcile committed main's conversion bytes and conflicted at the
  first mirror refresh (#252, leftover-branch matrix).
- When a rollout changes shape (big-bang → incremental), re-audit every
  premise the old shape let you drop: #252 M2 made verbs tracker-only for a
  simultaneous cutover, the per-repo rollout silently needed them legacy-capable,
  and only a live soak found it. For "behaves like the old binary" claims, test
  differentially against the old binary (`testdata/legacy-equivalence.sh`).
- A fix to a helper shared by both modes needs checking in both: #252's
  tracker-mode `TrunkRef` fix (local main behind the trunk) silently broke
  legacy mode (local main ahead of it), caught only by re-running
  `legacy-equivalence.sh` before close. When a fix picks one ref over another,
  enumerate every ordering of the two (behind, ahead, diverged) and test each.
- A mode decision is itself a read, and must degrade like one: #252's
  legacy/tracked selector asked the remote with no offline fallback, so a
  verb that never needed the network before started needing it just to learn
  which code path to run. When adding a gate in front of an existing local
  verb, ask what it does offline, and test it with an unreachable remote.
- Never slice `git status` output at fixed columns when the runner trims:
  the first entry's leading status space (" M path") is data, and trimming
  shifts its path by one. Read status as `--porcelain=v1 -z` through
  `gitx.ParseStatusZ` from untrimmed output. Hit twice in one day (#255's
  migrate dirty list; #259's close evidence silently dropped the gate ledger).
- A generated block inside a tracked file (weave's `.gitignore` block) is
  shared across checkouts, but the inventory that justifies each line is
  per-checkout. Rebuilding the block from local state alone erased the lines
  another checkout's compile committed (#263). Remove only what local
  provenance proves you own. And before adding a migration exception, check
  in git history where the legacy data actually lived: #263's first draft
  exempted legacy entries "inside the block", which never existed there.
- A fleet-wide "all migrated" claim must come from a sweep, not a worklist:
  #241's log declared every ariadne-layer repo adopted while parley.nvim and
  the three brains had never been touched (the Spec's "brain repos retain
  their commit rhythm" was misread as an exclusion). Before claiming fleet
  completion, enumerate every repo under the workspace and check the end
  state directly (inventory present, managed block, no owned link still
  tracked, one workflow include). Same for "no atlas surface": search the
  atlas for the component's name, not just the change's vocabulary (#264).
- A multi-pass writer to one shared file must let only the pass that owns a
  migration perform it: weave's data pass stripped the legacy ignore list
  that only the later artifacts pass could replace, so a refused artifacts
  pass left generated files exposed (#264).

- A protocol compatibility exception needs behavioral probes for both admission
  and refusal: confirmed old mode, restored state, retargeted state, and unknown
  provenance. New-format happy paths cannot prove an old-format exception.

- Don't infer whose work a commit is from git topology. Neither ancestry (a
  parent and its child share the parent's commits) nor a branch name (the
  owner can commit elsewhere) says who owns it; read the `#N` subject tag.
  #272's first start-plan guard refused a parent once it advanced past its
  child's fork, and #274's handoff guard refuses owner edits made off the
  owner's branch.
- A Revisions entry for one changed row leaves the others stale. When
  implementation diverges from a plan's Core-concepts table, check every row
  against the tree (does the path exist, does the symbol exist) at each
  boundary. The same goes for every backticked identifier in the Core-concepts
  prose and the Plan rows, not only the table cells. Rewrite the table to current reality in one Revisions entry
  rather than patching the row a reviewer named. #280 drifted twice: M1
  revised one row, and M2's review found four more absent files.
- "Unreachable because guard X refuses first" is a claim about every caller,
  not the ones you had in mind. Before skipping a branch on that argument,
  list the function's callers and confirm each one passes through X. #288 M1
  skipped unreadable cards in `ownedCompletions` as "behind transferguard",
  but landing and recovery reach it without the guard.
- When a rule has to be remembered at every call site, the fourth miss means
  the API should enforce it. #288's "an unreadable card is unknown, never
  absent" was re-found at a new `Records.Get` caller in three review rounds;
  unexporting `Get` (only `Require` remains) ended the family.
- Joining a shared authority's records (one tracker, read by every clone) onto
  per-checkout rows: compute "unmatched" against all rows, and deduplicate by
  the authority's key. Per-checkout matching reports each clone's live records
  as the other's orphans (#288 `dangling_claims`).
- Carry a read's outcome on the value it produced; never infer it from the
  value's emptiness. "No issue association" is both "looked up, no match" and
  "lookup failed" — #289 M1 judged the second from the first until the row
  recorded the lookup's error next to its result (as `claims_state` does).
- A read sdlc and weave both perform on slot state lives once in `pkg/`;
  `cmd/sdlc` never re-implements `cmd/weave/internal` behavior. #289 re-derived
  weave's marker list, `construct/deps` reader and placement rule, each weaker
  (lexical paths, symlink-following, FIFO-blocking) until moved to
  `pkg/workspace` / `pkg/layergraph` and shared.
- Compare a failing suite against main through the same runner. #283's main
  baseline ran `scripts/test-shard.py` directly, while the branch ran `make
  test`, which exports `Makefile.workflow`'s cwd-relative `WF_ISSUES_DIR`. A
  test that died only under make looked branch-caused until the env var was
  printed. When a guard says "not here", log cwd, the resolved top level and
  every `WF_*` override before suspecting the diff.
- A verb's surface is its help text, its atlas section and its recovery
  contract (`internal/recovery/catalog.go` Effects, Preconditions, Repeat,
  LostResponse, Ends, Proofs, and the `page.go` example text). Update all
  three in the same commit that changes the verb's behaviour, including a
  review's fix round: #284 M2's fix made publish bring main in, but the help
  and atlas still said "refused" until the next review. Sweep it in the same
  window that changes the verb. #283 changed claim and start-plan semantics and
  updated help, atlas and AGENTS, but the catalog still said claim flips
  open → working and start-plan "pushes nothing" until the fourth close
  review. The contract test proves only that the named tests exist, not that
  the prose is true.
- Every remedy a refusal names must run from the refused state, and a test
  must run it from there. #284 M2's republish refused "main moved; `sdlc
  claim` fast-forwards the rest", but the dirty file that caused the refusal
  blocks exactly that fast-forward. Where the remedy is mechanical, have the
  verb perform it (publish now merges main in itself). Where it isn't, name a
  step that works from the refused state (resolve the markers, then rerun).
- A note or log line written before a step that can fail must be convergent:
  the error's "rerun the same command" otherwise duplicates it (#284 M2:
  `unclaim --note`).
- A rerun that promises to finish a step must find its target in state that
  outlives the step it follows. #284 M3's takeover spent the release, the only
  record of the tip to check out, in the same card write whose response could
  be lost. The rerun then had nothing to finish at. Derive the target from
  what survives (the remote branch), and give every effect boundary that
  promises a rerun a lost-response test.
- When two readers of one record must agree (a split and the row that
  describes it), put the "is this record usable" predicate in ONE function
  that both call, and filter nowhere else. #234 M1's review found the
  split dropping unparseable rounds that the row metadata still counted, and
  then found a second filter (`BoundaryAll`) left in the pure grouping step
  after the first was unified. Each extra filter site is a disagreement
  waiting for an input that only one of them excludes.
- A reviewer instruction is only as good as the input it is shown. Before adding
  "if X changed, return REWORK" to a review prompt, check that X is inside what the
  window's commands show the reviewer. #183's fix-review told the reviewer to judge
  `## Spec` changes, but `reviewCommands` excludes `workshop/issues/`, so the rule
  could never fire. Where the input sits outside the window, compute the fact
  deterministically and pass it as a review note.
- Code and tests run inside agent sandboxes (macOS Seatbelt): exec of `ps` is
  denied, and `kill(pid, 0)` on any process outside the sandbox answers EPERM.
  Ask the kernel without exec'ing a tool, and read EPERM as "exists"; only ESRCH
  means gone. #330 found both: a cancellation test failed in every sandboxed slot
  (and two siblings ignored the `ps` error and passed vacuously), and the repo
  lock reclaimed a live peer slot's `.git/sdlc.lock` as "dead". An error ignored
  from an environment probe is a check that silently passes; fail on it instead.
- A ref you compute against is not always a ref a person can type. #269's tracker
  main came back as a pinned SHA (`TrunkView.Ref()`), and the first version put
  it into the `git merge <ref>` remedy and parsed a remote out of it for the
  push-state check. Keep two values: the ref the probe reads and the name the
  message shows.
- A prompt sentence that describes the reviewer's environment ("you run in a
  disposable checkout, revert there") must be rendered only by callers that
  provide that environment. #204 edited the shared code-review template, and
  the unisolated `sdlc judge milestone-review` path started telling reviewers
  to revert in the operator's tree. Make it a substitution keyed on the
  caller's fact (`PromptInput.Isolated`), and pin both renderings.
- A retry inherits every piece of mutable state the first attempt touched.
  When an attempt may mutate a resource (a review checkout), give the retry a
  reset hook and test it with residue left by attempt 1 (#204 BR-2).
- A word like "read-only" in help, flag usage or deprecation text is a claim
  about enforcement. Before writing it, name the mechanism that enforces it for
  every caller the text covers. #310 called every judge read-only while claude's
  Bash stayed unrestricted outside the #204 fence. Rewording the one paragraph
  the review cited left the same claim in four other strings. Grep the diff for
  the word and fix every string in one pass.
- A settings or test change is documented in more places than the file you
  edit: the atlas, the file's `$comment`, the test's header, the issue Log. In
  #334, three review rounds in a row each found one more stale description
  (an sdlc-binary fence note, the guard-test summary, a "gh stays bare" line).
  In the commit that changes the behavior, grep the name (`gh`, the test name,
  the setting key) across `atlas/`, comments and the issue, and fix every hit.
  Also: write `excludedCommands` entries as `X *`; a bare word matches only the
  verb with no arguments.
- Recovery advice in an error message is a plan the operator will execute, so
  walk it through the verb's own gates before writing it. #196's first stranded
  message said "commit the files, then re-run `sdlc merge`". The re-run then
  refuses at step 4 (branch ahead of origin) or at #148 (a merged PR whose
  branch has new commits). For each remedy a message names, check that the next
  verb it points to still accepts the state the remedy leaves.
- A scratch directory from a command substitution can come back empty: `mktemp`
  fails in the sandbox when it falls back to the system temp dir. Then `cd $D`
  goes to `$HOME` and the next write lands in the operator's home. In #299 an
  ad-hoc probe wrote `~/Makefile` this way, the same failure as #339's reviewer
  creating `~/.git`. Write `D="$(mktemp -d "$TMPDIR/x.XXXXXX")" || exit 1; cd
  "${D:?}"`, and never `cd` to an unchecked variable before writing files.
- Hoisting the checks of sequential steps ahead of the first step's writes
  changes what each check reads. In #265, the artifacts check moved before the
  data pass and read the pre-data inventory, so a path moving from a data mount
  to a generated output refused forever; the old order had converged it. When
  you move a check earlier, enumerate what every earlier step changes that the
  check reads (records, created paths, removed paths), project each into the
  check, and give every projection branch a test that fails with that branch
  disabled. Round 2 found the removed-path branch had no such test.
- Before writing a matcher for a repository convention (a commit-subject
  anchor, a branch name, an ID format), grep the internal packages for the
  existing one. #311 wrote its own `^#N` regex while `gitx.IsShippedWorkSubject`
  already owned that anchor plus the bookkeeping-verb filter, so the copy
  would have accepted a range holding only `#N: close` commits.
- Don't edit or commit in a checkout while its `make test` runs. The cmd/sdlc
  hermeticity guard compares the repo before and after, so your own edits read
  as "a test wrote the real repo" and every shard fails although every test
  passed (#329). Run the full suite only when nothing else writes to the tree.
- A handoff against a file someone else keeps editing (an ops project's Log)
  goes stale as a saved patch within hours. Hand off the regeneration script
  and keep the patch only as a preview (#329 D10).

- Adding binary sidecars under a document root changes existing subtree-reader
  workload even if the document parser ignores them. Select document metadata
  before loading blobs; per-attachment limits do not bound aggregate reads (#337 PQ-1).
- Multi-file writes need an explicit distinction between unchanged preflight
  refusal and recoverable post-effect failure. Rollback may remove only bytes
  this attempt owns, while restart must re-establish evidence (#337 PQ-2).

- Authorization for publishing new bytes belongs to the pinned destination
  snapshot, not a stale working-tree configuration. Re-evaluate on every retry
  and test remote-only tightening across all admission paths (#337 BR-1).
- Git can transform bytes at every materialization path, including an archived
  source later renamed into a safe destination. Check both paths before effects
  and verify actual restored bytes against pinned content (#337 BR-2).
- Classify a named function by its actual dependencies: snapshot loading, Git
  or clock reads make a proof integration even when its planner is pure (#337 BR-3).
- A fake external CLI response does not verify production argv against the real
  service. Supply an opt-in read-only conformance check and its cadence (#337 BR-4).

- A caller sweep is not complete until its inventory includes every direct
  checkout/restore followed by commit. Verify that inventory against code search;
  shared helper coverage did not imply issue restore used it (#337 BR-2 repeat).
- Do not strip destination host identity before an authorization query. Ambient
  CLI host settings can select a different private repository with the same
  owner/name; test conflicting host settings through the real seam (#337 BR-5).
- An index remains authoritative when its storage directory is absent. Skipping
  validation based only on directory history misses body-only dangling links
  (#337 BR-6).
- A prepared durable candidate may predate its current validator. Validate the
  complete pinned candidate at publication and uncertain-outcome recovery, not
  only during preparation; absence of asset pins is not a body-index proof
  (#337 BR-7).
- Every claim a durable contract makes needs a test that proves it: each
  half of an OR rule, each recovery-catalog field, each ticked plan box. In
  #344 the ref-containment half of `Leavable` (the half the crash recovery
  relied on) could be mutated to false with every test still green, and two
  review rounds found unproven claims in the same family. Before closing,
  mutate each branch of a new rule once and list the test that fails.

- Classify a function by all production effects reachable through its callbacks, not only its direct calls or name. An injected reader still makes a planner INTEGRATION when the production binding reads Git; trace every callback before declaring PURE (#337 BR-3).

- When an authorized acceptance revision changes required checks, update the current Done when criteria as well as appending its revision history. Label completed local verification separately from pending PR CI; do not leave two competing acceptance contracts (#358 BR-1).

- Polling deadlines must cover inflight IO and reject results that arrive after
  expiry, including green. Exclude separately budgeted suite time explicitly;
  give cleanup a fresh bounded context after cancellation (#332 BR-12).
- A pure action selector does not own authoritative state. Route production
  outcomes through typed state/event transitions, including error, recovery and
  retry paths; test the counter and ordering invariants by mutation (#332 BR-14).
- Bound the complete subprocess graph, including indirect shared Git helpers.
  Cancellation during push may follow remote publication: preserve the local
  commit and reconcile evidence on retry, rather than rolling back (#332 BR-13).

- A deadline is only complete when it reaches every subprocess on the actual
  call graph. Enumerate nested probes, replay helpers, tracker fallbacks,
  archive reads and final diff measurements, then test cancellation with a real
  sleeping command at each independent seam (#332 BR-12/BR-13 round 7).
- A new consumer configuration hook needs a README entry as well as atlas details: name its default convention, override placement, and failure behavior (#333 BR-1).

- Tests of environment-controlled output must own every relevant flag, including flags that override the one being tested. Codex's NO_COLOR broke forced-color assertions; inherited CLICOLOR_FORCE broke plain output (#363). Exercise the runner with conflicting ambient settings.

- Plan test strategies by named risky function, adversarial input class and a mechanical guard; case inventories alone hide missing invariants. For append-only diagnostics, a per-attempt cap is not a retention policy: state permanent cost or bounded supersession and measure at the writer (#364 PQ-1/PQ-2).

- In subprocess seams that execute on a heartbeat goroutine, test callbacks must return an error or use Errorf rather than Fatal: Fatal exits only that goroutine and strands the caller waiting for its result (#364 test development).

- Git diff is a normalized report, not proof of byte preservation. A guard that promises unchanged tracked inputs must fingerprint actual bytes/types/modes independently of clean/text conversion and assume-unchanged/skip-worktree flags; disabling one diff feature only fixes an instance (#364 BR-1).

- A pure admission test does not fulfill a promised lifecycle proof. For dispute
  flows, drive the real close/dispute/dispatch path and check both evidence delivery
  and non-advancement on invalid responses; mutation-check the integration wiring.


- A convergent retry must prove every authoritative effect after private receipt
  cleanup, including the effective calibration operation and current projection;
  an unchanged card alone cannot prove completion (#335).
- A calibration schema migration must preserve raw sample identities. Upgrade
  headers under the same writer lock and atomically include the first new row;
  test both fresh append and recovery of an already-written row (#335).
- Cancellation tests should separate cleanup latency from natural completion by
  a wide margin; subsecond scheduling assertions are brittle in sharded suites.
  Keep the cancellation bound while making the fake command much slower (#335).

- Admit delegated activity by original dispatch provenance before merging it
  into parent activity. Clipping a merged span by a parent's later eligibility
  loses the distinction between release, reclaim and continuous relocation
  (#335 BR-1).
- Authoritative exclusion must survive consumer selection across the full issue
  history. Test multiple eligible older rows, not only a sole target; filtering
  the excluded newest row alone can resurrect a superseded sample (#335 BR-2).
- Diagnostic requirements need production caller regressions. A correct
  comparator unit test cannot detect a removed close-time warning call; mutate
  the wiring and include the incremental-milestone control (#335 BR-3).

- Installing a new authoritative generation must retire metadata bound to its
  predecessor in the same transition. Test re-close after both numeric repair
  and exclusion, plus same-generation landing preservation (#335 BR-4).

- A successful repair must show the measurement's reconstruction basis and
  quality diagnostics as well as the number and hashes. Exercise preview and
  execution through the command path, and reuse measurement rendering (#335 BR-5).

- Bind accounting by recorded immutable dispatch facts and exact identities. A
  disposable directory name is not proof; an exact captured session must outrank
  an interrupted earlier root-only receipt when a checkout is reused (#365).
- Scope completeness to the denominator being projected. Global response dedup
  can coexist with per-agent coverage; another agent's unknown format must not
  invalidate a complete account calibration (#365).
- Live CLI metadata includes protocol noise before authoritative identity.
  Distinguish known queue records and startup banners from the first actual
  session metadata, and regression-test the observed provider formats (#365).

- Accounting recovery refs should retain their document, not pin the reviewed
  code graph. Test rebase/prune while metadata survives; generic recovery object
  retention can silently change code lifecycle semantics (#365).
- Keep separately bounded metadata outside a byte-measured diagnostic envelope.
  Adding a sidecar block after that envelope changes the reported retained bytes
  even when the diagnostic renderer is untouched (#365).
- New CLI flags and recovery commands need README coverage alongside atlas and help; enumerate all three before close (#366 BR-1).
- When a fenced reviewer cannot run a regression, record an exact-head host rerun, environment qualification and log path; distinguish unverifiable review execution from a demonstrated implementation failure (#366 BR-2).

- A successful process-group signal is not a fork barrier. Forceful teardown
  must cover members missed during traversal; test a successful partial first
  signal with a real descendant holding inherited pipes (#369).
- After a successful cancellation signal, preserve exec.Cmd.Cancel's nil
  result while reconciling group retirement; returning ProcessDone afterward
  can suppress the context error for a zero-exit parent (#369).
- A deadline may legitimately expire before a subprocess's readiness marker.
  Use explicit readiness for cancellation-after-start coverage rather than
  requiring a short-budget test to prove OS scheduling latency (#369).

- A child's 'released' message and even process Wait are completion evidence,
  not proof that a producer lease is available. Probe the resource within the
  existing settle budget; keep the child alive while asserting release so a
  missing-close mutation cannot pass on implicit exit cleanup (#369).
- test2json assigns assertion output to the exact failed subtest. A summary
  that prints only its parent discards the fault; preserve failed child output
  and regression-test it with recorded events (#369).

- File metadata is not a read budget: enforce limit+1 at the reader for both admission and post-effect rereads, and test growth after metadata validation with measured consumption rather than only checking the final error (#240 BR-8).

- A log-format contract needs fixtures for every named producer format, including structured events; test malformed records and failure identities as well as successful timings (#376 BR-1).
- Deduplicate evidence by the kind of check it covers, not only by commit SHA; a smoke result must not hide a suite log (#376 BR-2).
- Every temporary publication file needs an owner and a crash cleanup path. Test dead-owner reclamation alongside an active writer that must survive (#376 BR-4).
- Go's JSON struct decoder accepts field names case-insensitively even with unknown-field rejection. Validate exact persisted keys before decoding evidence that can change a success or failure outcome (#376 BR-5).
- A format matrix must cover both outcomes: a failing package can still report a duration, and omitting it biases slow-unit statistics (#376 BR-6).
- Apply declared input bounds before every consumer of the input. A parser that checks only total bytes can still hand an oversized line to regex and failure extraction (#376 BR-7).
- Exact JSON key validation does not prevent null from becoming a zero-valued nested scalar. Validate required fields, including nested measurement fields, before struct decoding and prove rejected rows never enter aggregates (#376 BR-8).
- Optional telemetry must not become a precondition for the operation it observes. Test failures both before execution (store reservation) and after execution (publication), and assert the primary result and status still complete (#376 BR-9).
- Reconcile all applicable cumulative evidence before declaring response spend
  complete: a newer format or reset must not erase a higher earlier observation.
  Snapshots diagnose gaps without being added to response totals (#365 BR-1).
- Invalid accounting evidence is uncertainty, not absence. Share artifact decoding,
  report validation/parse failures, and quarantine usable reviewer identities so
  invalid receipts cannot silently reclassify spend as worker (#365 BR-2).
- Reconcile the durable plan's concept map with final files and PURE/IO ownership
  before close; planned file names are not delivered architecture (#365 BR-3).
- A direct dispatcher test cannot prove gate wiring. Drive production boundary,
  plan and estimate entry points through retained receipts to usage, including
  interrupted outcomes; kill omitted-attachment mutations (#365 BR-4).

- Fake one selected agent and guard every fallback agent executable in production
  dispatch tests. An omitted explicit-agent flag must fail locally rather than
  launch a real provider from a hermetic fixture (#365 BR-4).

- Enumerate cumulative counter scopes as well as formats. Compare each turn's
  high-water evidence with that turn's response union; unrelated turns must not
  mask a missing segment (#365 BR-1).

- Discovery and decoding must share one immutable executing-identity parser.
  Preserve failed first metadata as uncertainty; later copied headers cannot
  repair it. Exercise the production discovery→decoder pair (#365 BR-5).
- Load exclusion evidence before filtering requested results. Reviewer identity
  must survive cross-issue/repository selection and private-receipt cleanup;
  test the committed artifact after the private copy is collected (#365 BR-6).

- Issue numbers are repository-local. Qualify diagnostic scope by the same
  repository identity used for ownership, even when exclusion evidence must be
  loaded fleet-wide; include same-number foreign-repo controls (#365 BR-6).

- A nested measurement may package cancellation as a status. Check context at
  the collector/command success boundary and preserve simultaneous output
  failures. Test late IO cancellation, not only entry cancellation (#365).
- Propagate summary/renderer writes, including short writes. After a durable
  calibration succeeds, an output error must say the observation was recorded;
  never pretend it rolled back or discard the original IO cause (#365).

- An immutable identity is also an ordering contract: account only after it is
  resolved, or buffer and reconcile every dependent counter and timestamp.
  Missing identity and pre-identity records must leave spend visibly incomplete;
  test early/late/absent headers through discovery and decoding (#365 BR-5).

- A valid process receipt is not a complete accounting denominator. Reconcile
  every relevant attempt with usable reported evidence or an explicit spend gap;
  test the collector and calibrated forecast, not only binding helpers (#365 BR-7).
- Canonicalize receipt phase-prefix copies before counting unfinished attempts.
  Confirmed launch failure is nonspending; missing launch evidence is uncertainty
  and cannot be collapsed into that outcome (#365 BR-7).

- Individually valid state records are not necessarily compatible copies. Reuse
  the writer's transition contract to collapse prefixes, and require uncontested
  evidence for nonspending exemptions. Test all phase/outcome pairs, rewrites and
  persistence orders; contradictory launch proof must survive (#365 BR-7).
- Convenience branch tracking is a durable shared-config effect. Before using
  push --set-upstream, trace the consumers; explicit remote/refspec publication
  can avoid both the protected config write and its later cleanup. Distinguish
  required ref retirement from ancillary metadata, and name an automatic retry
  owner for any retained residue (#374 plan review).

- Generated recovery commands must preserve the request’s semantic intent. An ambiguous exclusion that printed a numeric-repair retry lost --exclude; test both intents through the production refusal before publishing recovery guidance (#384).

- A failure reproduced on main proves shared exposure, not intermittency. Flake-based landing exceptions need same-exact-head fail→pass evidence and a named issue; persistent-main and suspected failures stay distinct and blocking (#390 BR-1).
- A native agent marker establishes the invoking stack, not its session. Bind
  session metadata to validated launch identity and workspace; newer activity
  cannot resolve competing sessions. Sampled prefix/tail metadata cannot rule
  out a contradiction in the skipped middle (#385 PQ-1).
- Optional native measurements require optional subtotals too. An empty sum over
  no supported observations is not measured zero, and an incomplete group must
  retain its independently known subtotal (#385).

- Unsupported provider fields still need format validation. Keep unavailable
  units null, but diagnose changes to their documented placeholders across every
  counter field; do not silently bless a new accounting format (#385 BR-1).
- A user-facing measurement change includes the README entry point and its atlas
  anchors. Sweep links when renaming a mapped section, alongside help and atlas
  updates (#385 BR-2).

- Failed evidence discovery needs an explicit unresolved outcome. A source is not
  proven unrelated merely because its owner has not been parsed yet; test parse,
  read and bound failures before metadata, then test production adoption against
  a positive surviving-source baseline (#385 BR-3).

- Provider discovery must distinguish standalone sessions from native auxiliary
  layouts. Unclassifiable files need visible skipped-source diagnostics, not a
  fleet-wide close failure; reserve adoption refusal for positively owned broken
  evidence and test a real nested subagent fixture through close adoption (#391).

- Shared rules in command help are executable contracts. If both roles must
  read the same section, embed one shared source and compose each output;
  independent wording copies will drift even when both have phrase guards
  (#404 BR-1).
- Run the pre-close full suite through `sdlc local-check run --name suite`
  on a pushed head so landing can reuse recorded evidence. A raw `make test`
  pass does not publish the suite status that the landing gate requires.

- A help-contract phrase guard must include the required action or restriction,
  not merely its subject. Reverse each approval, escalation and agreement
  clause independently and verify the corresponding output guard fails
  (#406 BR-1).

- When an approved revision extends acceptance, append concrete Done-when
  criteria in the same checkpoint. A Revision explaining the extension does
  not make the acceptance checklist complete (#406 BR-2).

- Keep the real worktree steady during direct cmd/sdlc test runs: its final
  hermeticity guard compares repository state across the whole invocation.
  Concurrent agent edits look like test leakage. Finish edits first, or run
  verification through the isolated local-check checkout (#408).

- #408: A publication invariant checked only at command preflight does not protect saved candidates or recovery. Put the rule in the existing shared publication admission hook and exercise first push, moved-base retry and durable receipt replay. Filter discovery by the canonical filename grammar before classifying an ID family; an ID prefix alone also matches backups.

- #408 close round 2: A shared publication hook must enforce the complete eligibility rule, including current card state. Moving only path absence into the hook leaves preflight status eligibility stale on retry and receipt recovery. Define the conjunction before implementing its checks.

- A permission-denied publication is not necessarily safe to ignore: ownership reconciliation may already displace an output or remove its old backup. Defer the effect before those mutations, retain original evidence, and test the real denied operation plus the supported recovery command on an issue branch (#411).
- The real-repo hermeticity guard applies to targeted cmd/sdlc runs too. Keep branch/HEAD/status fixed until they exit; recorded local-check suites use a detached checkout and avoid that coupling (#411 verification).
- Protect an operator-owned destination independently of its current declaration: removing a declaration still triggers retirement, and alternate/generated writers can reach the same path. Verify preserved ownership and recovery evidence plus eventual operator cleanup across those transitions (#411 BR-1).
- When review generalizes a behavior, update its atlas description in the same change; a merge-only description became stale when protection expanded to a destination across retirement and alternate writers (#411 BR-4).

- Signal ownership must match the lifetime of cancellable resources. A command-wide NotifyContext can swallow termination in context-free Git/network calls; scope it to owned queue/suite effects, bound cleanup independently, and test the unchanged production handler through blocked preamble and bookkeeping calls (#416 BR-1).
- Shared staging needs a TL grant even when the target is clean or already staged: a different issue may be using its binary while awaiting a live check. Confirm ownership before replacing :0, then clear and rebuild the full main tool set before reporting it free (#416 staging error; #421 tracks enforcement).
- Signal tests must state the child's inherited disposition. Background suites can start with SIGINT ignored, and Go preserves that default; establish SIG_DFL before exec when asserting default termination, while keeping the production handler intact (#416 full-suite fixture failure).
- Build verification must treat missing or malformed provenance as unknown, never as the safe default. A corrupt native executable can make a build-info reader return its ordinary "not Go" error, so classify the file header before skipping it; test invalid stamps and corrupt containers (#417 BR-1).
- A marker passed to a build tool is still an argument to that tool. Restrict it to the intended target grammar, and test options and variable assignments as well as shell metacharacters (#417 BR-2).
- When a workflow adds more external process steps, put the process call behind one injectable boundary and test order, failure state and retry with a stateful fake, alongside a real-tool conformance fixture (#417 BR-3).
- When classifying executable formats around a standard-library parser, match its full supported format set, including cross-platform formats. A host-only ELF/Mach-O/PE list let corrupt XCOFF and Plan 9 Go binaries bypass verification (#417 BR-1 round 2).
- A retained measurement must include every transcript needed to replay its evidence, even when correction removes that transcript's entire activity; test the zero-contribution case with another contributing source (#420 BR-1).
- A yielded tool is work evidence only after a terminal result. Running polls must preserve its pending association, and an untrackable running result must contribute no completed span (#420 BR-2).
- When adding a measurement method, enumerate every method consumer beyond validation and display; calibration trust and correction paths need explicit regression rows for legacy and every ownership-proven method (#420 BR-3).
- When native tool evidence spans a yielded call, test the wrapper's observed polling syntax end to end; a valid resumed Codex transcript in an old date folder was present, but assignment-form `write_stdin` polls left its suite call unpaired (#420 live preview).
- A returned poll proves only that the poll completed. Keep a yielded operation pending after transport errors, empty results and unknown-process responses; require explicit process-exit evidence before assigning positive work to that operation (#420 BR-2 round 2).
- When a measurement's authoritative producer records exact timing, use that record directly. Reconstructing the same interval by pairing transcript polls added call-ID ambiguity and repeated review failures; retain transcript evidence only for independently completed overlapping work, and mark unprovable yielded work or legacy timing as unknown (#420 BR-2 round 3).
- A metric need not assign a number to every overlap. If result decoding repeatedly decides whether waiting time becomes work, use the producer's exact queue interval, zero that interval, and refuse samples with a non-poll call inside it. This avoids granting time from incomplete or harness-specific tool results (#420 BR-2 round 4).
- For a conservative queue metric, prefer a bounded undercount over classifying nested tool calls: zero the producer-recorded wait without inspecting calls. Before calibrating, both close and correction must refuse any relevant or unclassifiable attempt record with incomplete evidence; a reader warning is not just display text (#420 BR-4/BR-5 round 5).
- An inferred legacy queue start cannot prove non-overlap: `started_at` follows acquisition by an unrecorded setup delay, so even a record started after release or a measurement end can hide a wait inside it. Check legacy uncertainty before interval filtering and test both boundaries through close and correction (#420 BR-5 round 6).
- A negative-path integration test must prove its fixture reached the intended guard. Validate generated records with the production decoder and match the selected repository and workspace before asserting a refusal; mutation-test the guard in isolation when another validation layer can produce the same error (#420 BR-5 round 7).

- Read quality does not establish value presence: a stale snapshot can truthfully contain no completion binding. Check the parsed binding identity independently and cover fresh/stale × present/absent whenever read-quality predicates guard optional values (#423 BR-1).

- Independently sampled timestamps are not stable identity. Join retained review evidence by reviewed commit, admit legacy ordering only with a fully anchored compatible sequence, and keep uncertain rounds unknown. History labels follow entry positions. Pin mutable refs before reading related artifacts; otherwise even matching sequences may mix generations (#423 BR-2/PQ-1).

- Measure which workflow phase paid for a rerun before adding ownership. A pre-close branch suite does not own the landing line: close its tested branch, then let merge integrate current main under its existing turn. Repeatedly chasing newer main before close can spend suites that the owned final integration would coalesce; ordinary docs-only publication already inherits eligible suite records (#425).

- When changed embedded help is the shipped behavior, characterize its execution mechanism and protect the rendered instructions with focused contract checks. Synchronize README usage alongside help and atlas in the same review window.

- Keep Agent and Co-Authored-By trailers together in the final Git message paragraph. A blank line between them makes the earlier Agent line ordinary body text, so trailer readers report it missing even when a grep finds the literal line.
- Special-case advice for an evidence refusal must classify the full blocker set, not the first diagnostic. Preserve all selector warnings through the error boundary and test mixed blocker orders before suggesting an irreversible exclusion (#428 BR-1).
- A numeric metric over retained logs needs an explicit history floor proving its whole cohort is still observable. Test expired records, every retry in an admission chain, and legacy rows through one shared time selector before calling coverage complete; restrict source collection to the requested cohort so unrelated old failures cannot null it (#392 BR-1/2/3/6).

- For a set of overlapping source claims, record every later claimant of a blocked physical file as unknown too; vary root order and add a third claimant in tests so a first-conflict-only fix cannot leave a false complete zero (#402 BR-1).
- An immutable dated artifact with an observation timestamp needs an existence check before recollection and a race check after publication. Test retries with an advancing clock, and validate persisted numeric semantics before accepting either a retry or prior-week comparison (#402 BR-2/BR-3).
- Persist observation validity independently of comparison status. A valid device reading can have an unknown weekly delta after a reboot, missing prior, or timing gap; serializing that status as the only validity signal prevents every later week from recovering. Test three-week sequences across the real JSON boundary (#402 BR-6).
- A syntax-only provenance rule must invalidate every fact used by a finding when a local assignment can change it, including conditional and indirect writes. An explicit interface destination also cannot inherit a concrete wire type from its initializer. Enumerate these paths in red fixtures before treating the matcher as conservative (#440 BR-1/BR-2).
- If a syntax-only rule cannot model aliases, discard its proof when a value escapes rather than tracking only the overwrite forms already seen. Include range bindings and closure captures in the write/escape matrix, and explicitly document the resulting false negatives (#440 BR-2 round 2).
- A temporal syntax proof can still fail when an escaped variable binding survives a fresh value assignment. For a narrow lint, prefer a whole-function eligibility rule (one constructor write, receiver-only uses, no address/closure escape) over restoring state after assignment; test the address-alias sequence at runtime (#440 BR-2 round 3).
- A module-wide source index must map an exact module import to the root directory as well as slash-descendant imports to subdirectories. Test root and subpackage identities with default and explicit aliases, lookalike prefixes, and local shadowing before claiming cross-file coverage (#440 BR-4).
- A whole-function write count does not prove local ownership: a package-global binding can be reset by a helper outside that function. Require the sole constructor to declare the binding locally, and test same-file and cross-file global resets alongside a positive local shadow (#440 BR-2 round 5).
- A direct-call lint can silently depend on statement syntax if each AST position has a separate visitor. Enumerate unconditional direct-call positions and test equivalent `var`, short-declaration, assignment, and return forms before claiming a rule covers the call (#440 BR-5).
- A lexical proof of statement order fails when `goto` can enter after the flag that established it. For a bounded syntax-only lint, reject goto-bearing functions as unproven and test both unconditional and conditional jumps against a straight-line control (#440 BR-7).
- A stateful fake that copies a production decision (a merge, a cap, a status refusal) makes the test pass whatever the production code does. Put the decision in one pure function that production and the fake both call, and test it directly on a real input. Check the regression test is real by deleting the decision (#394 M1 BR-2: replacing the card merge with `tests := add` kept every test green).
- A new kind of record stored beside an existing kind needs its own discriminator (a distinct name or type), not a side field every reader must remember to check. Readers that select the old kind then exclude it by default. Sweeping readers one by one missed one each round (#394 M2 BR-13, then BR-19: a rerun row saved as check `suite` was reused as the head's suite pass).
- When another issue's adopted actual is exclusion authority, derive occupied time from the same counted Build/Converge phases, not all attributed runs; verification Tail is not adopted. Decide temporal and slot relevance before demanding replay evidence, and clear exclusion options in diagnostic-only populations (#443 BR-1–BR-3).
- When a relevance decision needs an identity proof, parse every provenance source into one canonical-or-unknown type at entry. Validating only the claimant leaves retained source pins free to become false string mismatches; test a noncanonical saved pin end to end and guard any later string-based engine boundary (#443 BR-5).
- For an agent-facing protocol, test the decision at the artifact boundary: a rule that says "open a side doc" should have a before/after case whose link and index are checked, plus a mutation that tries the old action. Keyword-only assertions can pass while the example still teaches the wrong behavior (#437 and #444).
- When validating a before/after corpus that supports renames, compare moved outputs to the existing move projection before applying same-path emission checks. A same-path lookup mislabels preserved content as newly written; test a renamed parent, its children, and an orphan displaced by the move (#444 BR-1).

- #452 plan review: a cache of time-dependent authority must invalidate at every authority-derived temporal transition, including future activation as well as expiry; unchanged file metadata does not imply unchanged validity. Name the production functions and adversarial test strategies in plans instead of enumerating prose test cases.

- #452 cancellation: spawning a provider in its own process group requires an explicit SIGTERM unwind path in the Python owner; default signal exit skips finally cleanup. Test cancellation while the child is in flight, not only after provider return. Synchronize PID records on complete contents, not file existence.

- #452 BR-1: a live-file snapshot is not a lease on its pathname. All admission readers must tolerate a concurrently released ticket disappearing before open, including legacy owners that do not take the reader mutex. Pin the scan/release interleaving with real flocks.
- #452 BR-2: validate persisted identity with the same pure rule as caller identity before filesystem use. Absolute strings alone are insufficient: NUL, encoding, canonical form, bounds and field types must degrade visibly to ordinary FIFO rather than crash every reader.

- A faster test fixture does not prove independent timeout budgets. When moving
  a deadline from a batch to each item, inject command execution, observe distinct
  live contexts and cancellation between items, and demonstrate red with a shared
  context across the actual environment loop. Keep the production hang ceiling
  unchanged; healthy real scripts can pass both the bug and the fix (#448 BR-1).
- A new helper flag still needs README usage: state its output framing and whether
  it executes work, alongside the architectural atlas entry (#448 BR-2).

- Registry expansion changes consumers outside the registry package. Sweep
  rendered-header counts across Go tests and generated docs before selecting
  checks. Derive header expectations from the registry/shared renderer; reserve
  literal sets for deliberate policy tripwires, and name those exceptions
  accurately in the atlas (#454 BR-1/BR-2). The test-hardcoded-arch-count catalog
  rule catches fixed numeric architecture-header Contains assertions.

- A value handed to two goroutines (a closure, a cache, a flag) must be safe
  for concurrent use. Fixing the one the review named left the next one racing
  (#449 BR-1, then BR-10). When a design starts a goroutine beside another
  loop, list everything both receive, check each one, and add a `-race` test
  that drives both at once.
- A file another process wrote is a claim. Give it one check function (here
  `Readiness.For`) and route every reader through it, the display included. A
  reader that skips the check shows forged state as real (#449 BR-3, then the
  status reader). Read slow admission inputs outside the line's mutex, and cache
  the slow-changing ones.
- At a tolerant persisted-record boundary, validate known YAML node tags before
  struct decoding: `yaml.Node.Decode` can coerce numbers and booleans into
  strings. Traverse unknown JSON members with `UseNumber` so a valid future
  value such as `1e400` stays opaque; known numeric fields still decode into
  their declared types. Test both behaviors at the read boundary (#460 BR-1/BR-2).
- **#453 — synchronize acceptance on scope transfer:** when the hub moves
  acceptance steps to a peer issue, update the current Done-when checklist in
  the same checkpoint and preserve the original criteria plus ownership in
  Revisions. A revision note alone leaves two contradictory close contracts.
  This is a semantic scope decision; the existing close freshness judge owns
  enforcement rather than a lexical Go lint rule.

- #465: count subprocesses as well as wall time when optimizing shell fixtures on a shared macOS host. Isolate each scan to relevant owners, batch independent cases, reuse immutable successful sentinels, and select concurrency order with readiness barriers. Pure path splitting needs real dirname/basename edge conformance plus an N/2N zero-exec guard; keep partial/failed output fixtures independent.

- #465 CI: reuse executable contents by copying the immutable template when the production path renames outputs. Hardlinking source and installed output makes them the same file; GNU mv refuses it, and shared timestamp changes weaken build-isolation coverage even when BSD mv permits it.

- #465 BR-1/BR-2: replacing a clock wait must preserve the ordering the regression proves. Give start, edit and completed output independently ordered timestamps; mutate away restoration to confirm failure. When approved acceptance changes, update Done when alongside Revisions so the current contract names every required guard.

- #465 BR-3: fixture reuse must preserve observed output and status for every CLI mode. Never substitute a supplied status or another mode's baseline. Enumerate every reused result, regenerate after relevant input/context changes, and mutate ordinary/strict/verbose behavior independently.
- **#458 — review the observable command and its design table:** a function
  that invokes an injected Git callback is integration logic even when the
  callback is passed as a value. Test the final CLI rows for default and verbose
  behavior with real ancestry, and carry a reason when uncertain evidence keeps
  a row visible. Document new flags in README as well as atlas.
- **#458 BR-5/BR-6 — card ownership only:** issue Markdown, including dedicated
  Log lines, cannot establish ownership of a Go test. Presentation syntax is too
  ambiguous for a decision that hides a failure. Use only an exact fully
  qualified name in the terminal card's `flake_tests`; legacy failures stay
  visible until retention expires (ARCH-SECURE).

- **#473 M1 BR-1/BR-7 — test selection must over-select:** when a classifier maps a file to the code its change can
  break, collect every claimant (here the package whose dir holds the file, each `testdata/` owner, and each package
  that embeds it, plus every matching group), and fall back to the full run only when there are none. Two shortcuts
  each skipped tests: "nearest ancestor directory" let a root package claim every path, and "first owner wins"
  missed a second embedder. A third shortcut, ignoring docs before asking who owns them, skipped a package that
  embeds a `.md`. Ask "who claims this path?" first, and drop a path only when nothing claims it. Prefer the
  toolchain's exact facts (`go list` EmbedFiles) to a directory heuristic, and add a table row for a file with two
  owners.
- #470 verification: cmd/sdlc's source-checkout hermeticity guard pins HEAD for the entire test run. Do not commit while a targeted source-checkout run remains active: even passing tests fail that guard when HEAD moves. The recorded local-check suite uses a detached checkout and can overlap source bookkeeping safely. The existing hermeticity guard already enforces this class; no additional source lint is needed.

- #470: when extending an output callback to another mode, preserve existing mode prefixes and sweep tests by the message/callback as well as filenames. The existing readiness fixture caught the train-prefix change; the grant fixture now asserts both serial and train prefixes. These observable prose contracts are defended by those integration assertions rather than a generic lexical lint rule.

- #470 BR-1: acceptance should preserve behavioral guarantees rather than promise unchanged tests when a feature intentionally changes those tests' expected behavior. Keep the current Done-when synchronized with the approved policy, and preserve the superseded promise in Revisions. This is semantic acceptance wording, defended by the close review rather than lexical lint.

- #480 measurement harness review: process-leader exit does not prove its process group is gone; cancellation must check group survival independently and exercise a TERM-resistant descendant. A pinned HEAD also needs a clean-source check before and after measurement. Local evidence-harness tests enforce these contracts; no production Go lint surface is added for this temporary Python tool.
- #481 plan review: a PID-lifetime watcher does not own descendant completion. Before making quiescence authoritative, name the component that owns cancellation and cleanup; unmanaged children stay unknown rather than borrowing the managed runner's guarantee. Real owner-death/child-survival barriers defend this semantic lifetime boundary; a lexical source lint cannot prove descendant ownership.

- #481 M1 review: a durable cancellation intent must remain authoritative across every later write, owner death and retirement. Test the gaps between publications and repeated recovery, including diagnostic retention before intent cleanup. Check context inside bounded transaction loops; mutex-acquisition deadlines alone do not bound a command. These semantic crash/deadline properties are guarded by real-store regression tests rather than lexical lint.

- #481 M2 BR-5: a global admission mutex cannot contain a blocking secondary mutex acquisition. Try the local resource without blocking, release the global mutex on contention, and retry with the same ticket. Test a real held allocation lock and prove pause/cancellation finish before releasing it. The callback's dynamic lock context crosses Go/Python abstractions, so the real-lock contention tests enforce this class; a source-name lint would only mirror the present implementation.

- Bounded shutdown cleanup still receives the caller context: preserve its
  values with context.WithoutCancel, then apply the cleanup timeout. Do not mint
  context.Background inside a new runtime; run TestVerbContextsReachTrackerReads
  with targeted runtime checks (#481 final suite).

- #494: a train outcome that says landed precedes cleanup and cannot establish archive success. Reuse exact archive-transition proof before accepting a sibling's generic sweep archive. For tracked train fixtures, use prepared addressable slots, persist separate card/detail mirrors, and advance the prepared resting baseline past cutover before asserting the post-return sweep. Synthetic car clones exercise boarding but cannot substitute for real durable merge callers.
- #499 BR-1: a context-scoped cache is coherent only if every authority reader enters it. Sweep direct Snapshot consumers along the production call chain, including transferred-details checks before selection, and retain a production-path test that changes authority between them. A fake gate can verify cache lifetime while missing the raw-reader bypass. Cache lifetime and consumer provenance are semantic properties; the interleaving regression is the enforcement rather than a function-name lint.

- #499 root binding: when routing a publication guard through a shared reader, preserve the guard's original repository authority explicitly. A configured details directory is a data location, not the publication repository selector. Test a legacy peer directory beside a tracked checkout and ensure it cannot turn an existing ownership refusal into acceptance. This semantic provenance boundary is enforced by the production guard regression rather than lexical lint.

- #494 integration: a final-cleanup barrier does not force a waiting car to consume its train outcome; a concurrent readiness poll may observe MERGED first. A regression for the positive-outcome path must order that observation too. Model a lagging listing until the outcome is consumed, retain the actual merge/archive effects, and mutation-check the original failure. This semantic interleaving is defended by the real fixture and negative controls, rather than a lexical lint rule.
