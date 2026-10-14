Operation recovery contracts (#280). For each workflow verb: what it does when
it is repeated, interrupted, or answered by a lost response; how to observe its
outcome without trusting the response; and which tests prove each guarantee.

{{RECOVERY_SCOPE}}

WHY

  A coordinator that delegates work across slots (Couch "work on #N") sends
  messages over a lossy channel: a message can be lost, read late, or read
  twice, and a command's response can be lost after its effect landed. The
  channel's receipt proves delivery to a terminal, nothing more. Decide from
  durable state (`sdlc issue show N --json`, #279), and act by the verb's class.
  Nothing here promises exactly-once execution of an agent's instructions: a
  claim reserves an issue; it does not make later edits idempotent.

CLASSES

{{RECOVERY_CLASSES}}

AGENT GUIDANCE

  - Before repeating anything, observe: `sdlc issue show N --json`.
  - convergent-retry: rerun the same command; the durable state decides.
  - duplicate-safe-refusal: a repeat is refused harmlessly; read the evidence.
  - non-repeatable: never re-run to recover. Use the named path (for close:
    `sdlc issue recovery reconcile --issue N`; `sdlc issue recovery list`
    shows unfinished receipts).
  - stale or unknown observations are not negative evidence. Look again.
  - A request with no visible effect yet: look again after about 30 s before
    concluding it was lost; resend only to a convergent-retry verb.
  - Activity (a busy slot) is not progress; checkpoints are.

VERBS

{{RECOVERY_TABLE}}

EXAMPLE: SCHEDULING AN ISSUE ACROSS SLOTS

  Executed step by step by TestSchedulingExampleRuns, so it cannot drift.

{{RECOVERY_EXAMPLE}}

SEE ALSO

  sdlc <verb> --help     the verb's full contract and its proofs
  sdlc issue show --help the observation schema (#279)
  sdlc issue recovery    receipts and reconcile

MEASURED ACTUAL CORRECTIONS (#335)

  issue correct-actual retains its measured inputs in Git recovery refs before
  publication. recovery list includes pending corrections; reconcile --issue N
  resumes their tracker, calibration and mirror effects against the exact close
  generation. It probes uncertain publication and never re-measures on resume.
  An append-only calibration correction suppresses exact raw row hashes before
  trust/filter selection, so an invalidated sample cannot reappear through an
  older trusted row. Malformed correction evidence is a reported error.
