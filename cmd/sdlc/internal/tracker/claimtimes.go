// claimtimes.go — when each card was last claimed (#284), read from tracker
// history: every card write is a commit naming its operation, so the newest
// claim-kind commit touching a card is when its current owner took it.
package tracker

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// ClaimLogFormat is the `git log` --format whose output ClaimTimes reads
// (with --name-only): a record separator, the committer time, the message.
const ClaimLogFormat = "%x01%cI%x00%B%x00"

// The owner-setting operations, as their tokens' verbs (#284). Call sites mint
// tokens from these names, so claim ages cannot lose a verb a writer uses.
const (
	OpClaim    = "claim"    // claim, takeover included
	OpReclaim  = "reclaim"  // the operator's transfer
	OpRelocate = "relocate" // the owner's own `sdlc move`
	opAdopt    = "adopt"    // pre-#284 cards: recognised in history, no longer minted
)

// claimKinds are the operations that set a card's owner.
var claimKinds = map[string]bool{OpClaim: true, OpReclaim: true, OpRelocate: true, opAdopt: true}

var operationLine = regexp.MustCompile(`(?m)^Tracker-Operation: ([A-Za-z0-9][A-Za-z0-9._-]*)$`)

// ClaimTimes reads a newest-first tracker log stream in ClaimLogFormat and
// returns, for each card path in want, the time of the newest commit that
// both touched it and was a claim-kind operation. It stops reading as soon
// as every wanted path has its time, so its cost is bounded by the oldest
// live claim, not the whole history. Malformed records are skipped. Pure:
// its only input is the stream.
func ClaimTimes(r io.Reader, want map[string]bool) (map[string]time.Time, error) {
	out := map[string]time.Time{}
	if len(want) == 0 {
		return out, nil
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	sc.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		start := 0
		if len(data) > 0 && data[0] == 0x01 {
			start = 1
		}
		if i := strings.IndexByte(string(data[start:]), 0x01); i >= 0 {
			return start + i, data[start : start+i], nil
		}
		if atEOF && len(data) > start {
			return len(data), data[start:], nil
		}
		return 0, nil, nil
	})
	for sc.Scan() {
		fields := strings.SplitN(sc.Text(), "\x00", 3)
		if len(fields) != 3 {
			continue
		}
		when, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[0]))
		if err != nil {
			continue
		}
		m := operationLine.FindStringSubmatch(fields[1])
		if m == nil {
			continue
		}
		kind, _, _ := strings.Cut(m[1], "-")
		if !claimKinds[kind] {
			continue
		}
		for _, name := range strings.Split(fields[2], "\n") {
			name = strings.TrimSpace(name)
			if want[name] {
				if _, done := out[name]; !done {
					out[name] = when
				}
			}
		}
		if len(out) == len(want) {
			return out, nil
		}
	}
	return out, sc.Err()
}

// ClaimTimes reads this tracker's history for when each wanted card's current
// owner took it (#284), stopping once all are found; a failed read is an
// error, never an empty answer.
func (r *Repository) ClaimTimes(want map[string]bool) (map[string]time.Time, error) {
	if len(want) == 0 {
		return map[string]time.Time{}, nil
	}
	stream, finish, err := r.trunk.HistoryStream(ClaimLogFormat, vocab.Issue().Discovery().Cards)
	if err != nil {
		return nil, err
	}
	times, readErr := ClaimTimes(stream, want)
	if err := finish(len(times) == len(want)); err != nil {
		return nil, err
	}
	return times, readErr
}
