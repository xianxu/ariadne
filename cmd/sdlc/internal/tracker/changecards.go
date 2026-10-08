// changecards.go — several cards in one tracker commit (#284). A claim of a set
// of issues must land whole or not at all, and a lost race must be decided
// again over what the peer wrote: UpdateCard's fixed replacement bytes can only
// refuse a changed card, never re-judge it.
package tracker

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// ChangeCards rewrites the named cards (duplicates collapse) in one tracker commit. On every attempt
// it reads the current cards (a missing or unreadable one is absent from the
// map) and calls decide, which returns only the cards to rewrite, by ID; an
// error from decide aborts the whole set with nothing published. An empty
// decision is ErrNoChange. Each replacement must keep its card's identity.
func (r *Repository) ChangeCards(ids []string, operationToken string, trailers []string, decide func(current map[string]Record) (map[string][]byte, error), beforePush func(base, candidate string) error) error {
	if len(ids) == 0 || decide == nil {
		return errors.New("tracker change requires cards and a decision")
	}
	for _, t := range trailers {
		if strings.ContainsAny(t, "\r\n") || !strings.Contains(t, ": ") {
			return fmt.Errorf("malformed tracker trailer %q", t)
		}
	}
	if !tokenPattern.MatchString(operationToken) || beforePush == nil {
		return errors.New("tracker update requires an operation token and receipt callback")
	}
	seen := map[string]bool{}
	var sorted []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			sorted = append(sorted, id)
		}
	}
	sort.Strings(sorted)
	what := "update card"
	if len(sorted) > 1 {
		what = "update cards"
	}
	message := cardsMessage(sorted, what, operationToken, trailers...)
	return r.trunk.UpdateManyPrepared(message, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := r.read(view)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		current := map[string]Record{}
		for _, id := range sorted {
			if rec, ok := snapshot.Card(id); ok {
				current[id] = rec
			}
		}
		next, err := decide(current)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		write := map[string][]byte{}
		var replaced []replacement
		for id, raw := range next {
			rec, ok := current[id]
			if !ok {
				return gitx.TrunkWrite{}, fmt.Errorf("tracker change wrote #%s, which it did not read", id)
			}
			if bytes.Equal(rec.Raw, raw) {
				continue
			}
			card, err := issue.ParseCard(raw)
			if err != nil {
				return gitx.TrunkWrite{}, err
			}
			if card.ID != id {
				return gitx.TrunkWrite{}, errors.New("tracker update must preserve expected card identity")
			}
			replaced = append(replaced, replacement{rec, raw})
			write[rec.Path] = bytes.Clone(raw)
		}
		if len(write) == 0 {
			return gitx.TrunkWrite{}, ErrNoChange
		}
		if err := snapshot.validateReplacements(replaced); err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: write, ExactBytes: true}, nil
	}, beforePush)
}
