Reserve an open issue card on the tracker (open → working).

  sdlc claim --issue N

An issue is claimable only once its creation is complete: its card is on the
`issue-tracker` branch AND its details file has landed on main. A card-only or
local-only issue is still being written by its creator; claim refuses it and
names the next action (`sdlc issue move-detail --issue N` from the creating
checkout, or merging the branch that filed it). Readiness is checked against
fresh main, and checked again immediately before the reservation is pushed.

Claim changes only the card: status → working, `updated`, and a first
`started` stamp. It is a compare-and-swap on the card blob it read, published
through the resting branch's configured upstream remote. Nothing is committed
or pushed to main, and the resting branch is never edited. In a checkout that
holds the details on a feature branch, the details' card mirror is refreshed;
a hand-edited mirrored field is left alone and reported.

A competing claim has one winner; the loser sees "not open" or a changed card.
Every non-open status refuses, including a repeat by the original worker —
continue started work without reclaiming. No estimate is required; change-code
owns the implementation gates. Next: `sdlc start-plan --issue N` moves design
onto the issue's own branch.

FLAGS

  --issue <n>           required issue ID
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues (details home)
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history
  --dry-run             check readiness and describe the reservation; change nothing

EXAMPLES

  sdlc claim --issue 31
  sdlc claim --issue 31 --dry-run

RELATED

  sdlc issue new          file an issue (card on the tracker, details locally)
  sdlc issue move-detail  complete creation by publishing initial details to main
  sdlc start-plan         prepare the issue branch and deliver design principles
  sdlc issue set-status   guarded lifecycle transition on the card
