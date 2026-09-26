package tracker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// IssueRecord composes one issue from its two homes (#252): the tracker card is
// authoritative for card-owned fields; the details in this checkout carry the
// rest. Either half may be absent — a card-only issue is still being created,
// and a repository without a tracker (pre-migration) has details only.
type IssueRecord struct {
	ID         string
	Card       *Record // nil without a card
	DetailPath string  // absolute; "" when this checkout has no details
	DetailFM   string
	DetailBody string
	DetailErr  error // unreadable or malformed details, reported rather than skipped
	// DetailUnreadable separates an IO failure from malformed content.
	DetailUnreadable bool
	tracked          bool // composed from a repository that has a tracker
	// Duplicate marks a second details file for an ID already seen: listed so
	// ambiguity stays visible, never returned by Get.
	Duplicate bool
}

var cardOwned = func() map[string]bool {
	owned := map[string]bool{}
	for _, f := range vocab.Issue().CardFields() {
		if f.Kind != "title" {
			owned[f.Name] = true
		}
	}
	return owned
}()

// Field returns a frontmatter field from the half that owns it. A card-owned
// field of an issue with a card is never read from the (possibly stale) mirror.
func (r IssueRecord) Field(name string) (string, bool) {
	if cardOwned[name] && r.Card != nil {
		return issue.GetField(r.Card.Card.Frontmatter, name)
	}
	if cardOwned[name] && r.tracked {
		return "", false // tracked repository, no card: unknown, not the mirror
	}
	if r.DetailPath == "" || r.DetailErr != nil {
		return "", false
	}
	return issue.GetField(r.DetailFM, name)
}

func (r IssueRecord) Status() string {
	s, _ := r.Field("status")
	return s
}

// Title is the card's canonical title, else the details' first H1.
func (r IssueRecord) Title() string {
	if r.Card != nil {
		return r.Card.Card.Title
	}
	for _, line := range strings.Split(r.DetailBody, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// FetchMode says how fresh the cards must be.
type FetchMode int

const (
	// Fresh fetches the tracker and fails when it cannot: gates that authorize
	// a write must not act on a stale card.
	Fresh FetchMode = iota
	// PreferFresh fetches when possible and otherwise reads the last-fetched
	// tracker, marking the result stale — for read-only views.
	PreferFresh
)

// Records is one composed, pinned view. Tracker is false for a repository whose
// publication remote has no issue-tracker branch: its details are the record.
type Records struct {
	Tracker bool
	Stale   bool
	Ref     string
	list    []IssueRecord
	byID    map[string]int
}

func (rs Records) All() []IssueRecord { return append([]IssueRecord(nil), rs.list...) }
func (rs Records) Get(id string) (IssueRecord, bool) {
	i, ok := rs.byID[id]
	if !ok {
		return IssueRecord{}, false
	}
	return rs.list[i], true
}

// LoadRecords joins the tracker cards (via repo; nil means no publication
// remote is configured) with the details under detailsDir.
func LoadRecords(ctx context.Context, repo *Repository, detailsDir string, mode FetchMode) (Records, error) {
	var rs Records
	var snap Snapshot
	if repo != nil {
		exists, err := repo.Initialized()
		switch {
		case err != nil && mode == Fresh:
			return rs, err
		case err != nil:
			local, ok, lerr := repo.LocalSnapshot()
			if lerr != nil {
				return rs, errors.Join(err, lerr)
			}
			rs.Stale, rs.Tracker, snap = true, ok, local
		case exists:
			if snap, err = repo.Snapshot(); err != nil {
				return rs, err
			}
			rs.Tracker = true
		}
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return rs, err
		}
	}
	byID := map[string]*IssueRecord{}
	if rs.Tracker {
		rs.Ref = snap.Ref()
		for _, c := range snap.Records() {
			card := c
			byID[c.ID] = &IssueRecord{ID: c.ID, Card: &card, tracked: true}
		}
	}
	matches, err := filepath.Glob(filepath.Join(detailsDir, "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"))
	if err != nil {
		return rs, err
	}
	var dups []*IssueRecord
	for _, p := range matches {
		id, slug, ok := issue.ParseFilename(filepath.Base(p))
		if !ok || slug == "" { // the inventory requires a slug
			continue
		}
		rec := byID[id]
		if rec == nil {
			rec = &IssueRecord{ID: id, tracked: rs.Tracker}
			byID[id] = rec
		}
		if rec.DetailPath != "" {
			dup := &IssueRecord{ID: id, Card: rec.Card, tracked: rs.Tracker, Duplicate: true}
			readDetails(dup, p)
			dups = append(dups, dup)
			continue
		}
		readDetails(rec, p)
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rs.byID = map[string]int{}
	for _, id := range ids {
		rs.byID[id] = len(rs.list)
		rs.list = append(rs.list, *byID[id])
		for _, d := range dups {
			if d.ID == id {
				rs.list = append(rs.list, *d)
			}
		}
	}
	return rs, nil
}

// readDetails fills one record's details half; errors are recorded, not skipped.
func readDetails(rec *IssueRecord, p string) {
	rec.DetailPath = p
	raw, err := os.ReadFile(p)
	if err != nil {
		rec.DetailErr, rec.DetailUnreadable = err, true
		return
	}
	rec.DetailFM, rec.DetailBody, rec.DetailErr = issue.Parse(string(raw))
}

// RepositoryFor opens the tracker of the repository at root through its
// resting branch's publication target. Without a usable target the repository
// can have no tracker — unless one was fetched already, which means the
// configuration broke: that is an error, never a quiet fallback to mirrors.
// nil, nil means "no tracker": its details are the record.
func RepositoryFor(ctx context.Context, root, restingBranch string) (*Repository, error) {
	if _, err := gitx.RunGit("-C", root, "rev-parse", "--git-dir"); err != nil {
		return nil, nil // not a repository: nothing can have been tracked
	}
	target, err := gitx.ResolvePublicationTarget(ctx, root, restingBranch)
	if err == nil {
		return NewRepository(ctx, root, target.Remote)
	}
	fetched, ferr := gitx.RunGit("-C", root, "for-each-ref", "--format=%(refname)", "refs/remotes/*/"+vocab.Issue().Discovery().Tracker)
	if ferr == nil && len(fetched) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf("%s has a fetched issue tracker but its publication target is unusable: %w", root, err)
}
