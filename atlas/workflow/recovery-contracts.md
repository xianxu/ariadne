# Recovery contracts (#280)

What each workflow verb does when it is repeated, interrupted, or answered by
a lost response, how to observe its outcome, and which tests prove it. It is
the rulebook a coordinator uses once Couch's lossy channel has sent "work on
#N". Observations (#279) say what happened; contracts say what an agent may do
next.

**Scope:** repositories with an issue tracker (#252). In a legacy repository
(no issue-tracker branch) these guarantees are unknown, because its verbs
publish details to main directly. Every generated surface renders
`recovery.Scope`.

- **Single source:** `cmd/sdlc/internal/recovery`.
  - `Catalog` holds one `Contract` per verb or verb family: class, effects,
    evidence, preconditions, repeat, lost response, when guarantees end, and
    proofs naming tests. `Exempt` lists the mutating commands outside the
    issue workflow, each with its reason.
  - `attachRecoveryContracts` (main.go) appends each contract's section to
    its verb's `--help`, in one pass over the command tree.
- **Classes:**
  - `read-only`
  - `duplicate-safe-refusal`: a repeat is refused and changes nothing.
  - `convergent-retry`: a repeat reaches the same end state.
  - `non-repeatable`: a repeat is a new effect, so recover through the named
    path.
  - Nothing is called idempotent wholesale. A claim reserves an issue; it
    does not make later edits idempotent.
- **Proven, or marked unproven:** `TestRecoveryContractsAreProven` derives
  the required verbs from the command tree (every repo-lock-annotated
  command, plus start-plan and issue show). It fails on a verb with neither
  a contract nor an exemption, on a stale exemption, on a missing help
  section, and on any named proof that no test declares. A claim with no
  test renders as **UNPROVEN** (treat as unknown).
- **Single-card verbs:** claim, adopt, relocation, reclaim and the card
  setters publish through one seam, `cardPublish` (pinned by
  `TestCardPublishCallers`). A lost acknowledgement gets one shared message:
  "rerun the same command — the card decides" (`uncertainCardWrite`).
  Lost-ack, race, duplicate and generation cases each name their tests in
  the catalog.
- **The topic:** `sdlc help recovery` renders the scope, the classes, agent
  guidance, one entry per contract, and the scheduling example. The topic is
  `helptext/recovery.md` plus the `{{RECOVERY_*}}` placeholders in
  `renderLong`; the rendering lives in `recovery/page.go`.
- **The executable example:** `recovery.Example` is data: actor, command,
  expectations as dot paths into `sdlc issue show N --json`, the "otherwise"
  action, a `TrackerUnreachable` condition, and an `IfVerdict` condition
  (FIX-THEN-SHIP's fix-and-reconcile steps). `TestSchedulingExampleRuns`
  runs it from the same data, once per legal close verdict (SHIP and
  FIX-THEN-SHIP). An expectation must hold for every successful run, so one
  that holds for only one variant fails. The fixture's own values (quick
  flow, stubbed verdict) are checked separately, never as guidance.
  - The recipient runs each command. A convergent-retry command is
    delivered twice.
  - A coordinator in another checkout checks each expectation with
    `recovery.Lookup`.
  - The only additions to the documented commands are fixture switches:
    no judges or estimate, a given actual, no separate worktree, no atlas.
