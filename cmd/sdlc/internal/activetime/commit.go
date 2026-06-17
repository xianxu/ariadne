package activetime

import "time"

// Commit is a window commit with its tracked-issue subject refs (deduped,
// order-preserving). Time is the author date (%aI). Commits define segment
// boundaries; each non-suffix segment is anchored by the commit at its end.
type Commit struct {
	Time    time.Time
	SHA     string // short (7)
	Subject string
	Issues  []string
}
