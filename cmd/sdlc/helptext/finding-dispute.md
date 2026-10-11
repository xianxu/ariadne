Decline a gate finding, with a rationale and a counter-measurement (#202). The
dispute is recorded as the finding's `disputed` disposition in the gate's
ledger (`<issue>-plan-gate.md` for a PQ-n id, `<issue>-close-gate.md` for BR-n),
which this verb commits by itself on the issue branch (`#N: dispute ID`). It
never pushes.

  sdlc finding dispute --issue 202 --id BR-40 \
    --rationale "the byte path is asserted by the full suite, not -run TestPTY" \
    --evidence "go test ./... with the mutation: TestEditorLoopCtrlCExitsZero exit = 9, want 0"

WHEN. A finding is wrong, or correct but not worth its fix (the fix breaks a
pinned invariant, say). Fix it if you can: disputing costs evidence, fixing
costs a diff. The evidence is a measurement you RE-RAN — the command and what
it showed — or the invariant and the test that pins it. An opinion is not
evidence.

WHAT IT DOES TO THE GATE.
  - An Important or Minor finding stops blocking.
  - A Critical (appeal-only: the hard-blocking set) is recorded and put to
    the judge, but keeps blocking until the judge withdraws it.
  - The next review round lists each live dispute with its rationale and
    evidence. The judge answers it: `withdrawn` (the evidence holds) or
    `not-addressed` (re-raised; the finding blocks again, and you fix it —
    the CLASS it names, not the site: ARCH-PURPOSE (`sdlc arch-principles`)
    — or dispute it again with new evidence). A dispute it leaves
    unanswered stands.
  - It is loud. This verb warns, every gate line reports "N of M finding(s)
    settled by implementer DISPUTE, not a fix", and `sdlc close` records the
    counts in the calibration ledger (gate_disputed, boundary_disputed).

REFUSALS. Each names its reason, before anything is written:
  - --rationale or --evidence missing or blank;
  - an id that is not PQ-n or BR-n, or that the ledger never issued;
  - a finding already disputed and not yet answered by a review round;
  - a finding the judge already settled (addressed or withdrawn);
  - not on the issue's branch; a repository without the issue tracker.

{{OWNERSHIP_GATE}}

FLAGS
  --issue N           the issue whose gate raised the finding (required).
  --id ID             the finding's id as the gate printed it (required).
  --rationale TEXT    why it is wrong or not worth fixing (required).
  --evidence TEXT     the counter-measurement (required).
  --issues-dir DIR    issue details directory (default workshop/issues).
  --plans-dir DIR     gate ledger directory (default workshop/plans, or
                      WF_PLANS_DIR).
