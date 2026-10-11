Report the converge factor k = converge / build, per model, flow and review
recipe, from the calibration ledger (#234). Read-only.

WHY THIS EXISTS

  An estimate prices the work up to "done as far as I can tell". It cannot
  price the review-and-fix loop: nobody budgets time for bugs they have not
  written yet. So every measured actual is split at the issue's boundary
  reviews (see `sdlc actual --help`):

  - build is design and implementation up to each boundary's first review
    dispatch;
  - converge is the review-and-fix loop after it.

  Drift compares the estimate with build only. Converge is calibrated as a
  factor instead, which this report shows.

WHAT IT READS

  The newest row of each issue that carries a split (split_source dispatch,
  finalize or no-review) with positive build. Rows written before #234
  (split_source -) and typed-actual rows (manual) have no split and are left
  out. Quick-flow rows count: they have no estimate, but they do have
  converge. Historical rows have no split until ariadne#323's backfill.

OUTPUT

  One line per (flow, model, recipe) group: n issues, Σbuild, Σconverge and k.

FLAGS

  --brain-dir <path>  brain root holding the ledger (default ../brain;
                      $WF_CALIB_LEDGER overrides the ledger path)
