Show a repository's landing line, and grant or revoke fleet-blocking priority
(#393).

THE LINE

  Each repository has one landing line. `sdlc merge` waits in it, so one merge
  at a time integrates and lands, and each landing runs one integration suite
  instead of one per main movement. The line lives in the repository's git
  common dir (`sdlc/landing/`), which every slot shares; it needs no Couch.

  A waiting merge holds a ticket, a file it keeps an flock on; the admitted
  merge holds `head.lock`. Every flock dies with its process, so a killed
  merge leaves the line by itself. Admission order: fleet-blocking tickets
  first, then normal ones, first come first served within each class. A
  running landing is never interrupted.

  sdlc landing [status] [--repo DIR]
      The landing now, the waiting merges in admission order, live grants, and
      the ready-to-landed delay p50/p90 over the last 30 days.

LANDING TRAINS (#449)

  `.sdlc/landing.json` on main turns the line into trains:

    {"train_interval_minutes": 60, "train_max_cars": 8, "train_max_wait_minutes": 30}

  The file is absent by default, and an interval of 0 keeps the serial line
  above. It is under the local-gates guard, so the operator verifies every
  change to it.

  {{TRAIN_RAMP_POLICY}}

  With trains on:

  - A waiting merge checks its own publish and CI gates, read-only, on its
    pushed head, and vouches it ready (`ready/<seq>`). Where main declares
    .sdlc/affected.json (#473), a recorded `affected` run stands in for
    `suite` here: the train runs the full suite on its prefix before anything
    lands. A docs-only branch boards on its docs-only `affected` record.
    Prepare a car with `sdlc local-check run --name affected --boarding`: a
    full selection refuses there rather than running a per-car suite (#561),
    and names its repair and the full-suite alternative. A passing `suite`
    record on the head still boards.
  - A merge departs only when the train is due and its vouch is current. The
    train is due:
    - on the first train;
    - at once when train_max_cars cars are ready (#475);
    - when the oldest ready car has waited train_max_wait_minutes since its
      first ready vouch (0 or absent: no bound);
    - at once after a departure that left ready cars behind;
    - otherwise at the next departure: the repository's next wall-clock slot
      after the last train (with a machine plan), or one interval after it.
    An early departure never moves later slots. A lone or fleet-blocking
    landing is not a departure: it never pushes the next train back.
  - The admitted head conducts. It boards every waiting merge whose vouch
    holds and whose remote tip and open PR are still at the vouched head, up
    to train_max_cars. It builds `landing-train/<id>`: main, then each car's
    exact head, with the instance and duplicate-ID gates run per prefix.
  - It opens one train PR, runs one recorded suite and CI on it, and merges
    it once. GitHub marks every car's PR merged. Each car's own sdlc merge is
    told (`outcomes/<seq>`) and finishes its cleanup.
  - Holds and splits:
      - A car that conflicts with the train is held.
      - An unexplained suite failure is bisected: the culprit is held, the
        clean prefix lands, and the cars behind it ride the next train at
        once.
      - A known flake (#394) gets its rerun and never splits.
      - Main failing the same way refuses the train, as does red CI on the
        train PR.
  - A fleet-blocking grant departs at once and lands alone. So does a PR
    carrying the local-gates label.
  - A car that finds its own PR merged with no word from its train finishes
    as a resume.

  `sdlc landing status` shows the train setting, the next departure and why
  (slot, interval, max wait, full, cars left ready), each
  waiting merge's readiness, and the last trains (result, cars, train PR,
  split suites, culprit).

  It also shows the train filter over the last seven days (#561), the
  measurements the cap ramp rule reviews. Each train records a start and a
  final measurement in the landing log: the proof each car boarded on
  (affected, suite or unknown), which boardings a passing prefix resolved,
  which an exact attributed split confirmed as escapes (a failure the
  train's suite found in a package the car changed, its affected run having
  passed), and every whole-suite run it made (initial, main baseline, split
  probe, reintegration) with its run and queue time. The line reads:
    escapes E/R resolved affected boardings (rate, or unknown at R=0);
    unresolved, suite and unknown-proof boardings beside it; trains
    measured, incomplete (a start without a final, a lost append, a
    bound) and unmeasured (before #561); the split's extra suites (baseline
    plus probes) with run, queue and elapsed wall time, and reintegration
    suites apart.
  Registered flakes, untouched or unknown impact, a red main and released
  cars are never escapes. The weekly test-stats snapshot carries the same
  numbers and line as train_filter. Measurements never decide a landing.

DROP A COMPOSED CAR (#526)

  sdlc landing drop --issue N [--reason WHY] [--repo DIR]

  The car's owning checkout/workspace or the same repository's :0 steward may
  drop its exact composed car. The actual caller is checked independently of
  --repo; entering :0 from a worker does not grant steward authority.

  Acceptance durably excludes the pinned car, stops and joins the old train's
  work, closes its obsolete PR/ref and lets surviving runners re-form. New-head
  validation runs normally. The dropped runner exits held with the reason and
  its exact branch recovery; it stays out until explicit owner retry.

  Merge starts when the green-CI result atomically admits the Merge effect,
  before that effect's final fetch/API. From then on, drop refuses and names the
  exact phase and irreversible start marker, including unconfirmed outcomes.
  A refusal does not claim the car could still be removed.

  Status shows accepted drops and cleanup/recomposition pending. Pending cleanup
  survives a dead conductor and blocks every successor's merge until positive
  old-PR/ref closure proof; resume the printed original conductor merge command.
  Missing, malformed or unsupported control refuses without accepting a drop.
  Conductors started with older binaries do not support this protocol.

THE MACHINE PLAN (#475)

  Repositories with trains on share one machine's suite slots. The machine
  plan places each on a wall-clock grid (epoch + offset + k * interval) so
  heavy train suites don't collide: for example ariadne at :00 on both slots,
  pair and parley.nvim together at :15. It is one file beside the suite lock,
  ${XDG_STATE_HOME:-~/.local/state}/ariadne/landing-plan.json (or
  $WF_TEST_LOCK_FILE's directory), and it belongs to the steward.

    sdlc landing plan            the plan computed from today's costs (a dry run)
    sdlc landing plan --write    replace the plan file

  Each repository's reservation is its suite's p90 wall time over 7 days
  (test-stats), rounded up to 5 minutes plus 5, times its declared suite
  weight (no samples: 15 minutes). The heaviest goes first, at the earliest
  offset that keeps every minute within the machine's slots ($WF_TEST_SUITES,
  default 2). --write keeps the file's unknown members and other
  repositories' entries, and refuses to replace a plan it can't read. Rerun it
  after a cost changes (a suite's shard count, a repository turning trains on).
  A repository without an entry departs one interval after its last train; an
  unreadable plan does too, warned once.

FLEET-BLOCKING PRIORITY

  sdlc landing grant --issue N --reason WHY [--hours H] [--repo DIR]
  sdlc landing revoke --issue N [--reason WHY] [--repo DIR]

  A grant puts issue N's merge ahead of waiting normal ones at the next
  admission. While any valid grant exists, normal merges wait even before the
  granted issue queues. Granted issues keep FIFO among themselves; a running
  landing continues. Status and wait output name each grant, reason, expiry and
  the revoke command (run it in this repository's checkout). Expiry or
  revocation releases the hold. It waives no gate. Only the operator's shell or a :0 slot (the
  TL's hub) may grant or revoke; a worker slot, or a :0 checkout entered from
  one ($COUCH_TREE), refuses. A grant lasts at most 24 hours (default 4) and
  ends when the issue lands. The grants log is the only source of priority,
  and the slot rule is checked again when it is read. Trust is cooperative:
  this guards against agents' mistakes on one machine, not against forgery.

SUITE QUEUE PRIORITY (#452)

  The same active grant puts the issue's suite ahead of waiting normal suites
  at the next sufficient free machine capacity. Running suites continue;
  weight is respected and priority suites are FIFO. Expiry or revocation
  returns a waiting suite to its original normal FIFO position. An expired
  grant confers nothing. No extra suite grant is needed. A composed train's
  suite inherits any active grant on a car in its exact tested prefix, including
  one granted while queued. Conflicting/removed cars and later split cars confer
  nothing; revoke and expiry restore normal FIFO.

  Recorded suites bind the original checkout's issue branch to the tested
  HEAD; main diagnostics and other checks stay normal. Direct make test uses
  its supplying workflow's sdlc provider. Unavailable or invalid grant evidence
  prints a priority-unavailable warning and uses normal FIFO for every suite;
  priority never prevents a suite running. Old sticky helper waiters drain
  naturally during rollout; they are never killed or unlocked.

RETENTION

  `events.jsonl` (enqueue, admit with what it displaced, landed, refused,
  train and car outcomes) and `grants.jsonl` keep 30 days. A sidecar in
  `ready/` or `outcomes/` goes with its ticket. A daily sweep never changes which grants are
  valid. Malformed lines are skipped and counted.
