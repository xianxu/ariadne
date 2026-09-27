// migration.go — the pure plan of the one-time tracker migration (#252).
//
// PlanTrackerMigration turns an inventory of a legacy repository into a
// deterministic manifest: the cards the tracker starts with, the details main
// converts, and every reason the cutover must not happen yet. It performs no IO;
// the command gathers the inventory and applies the manifest, and the digest
// lets the operator apply exactly the plan they reviewed.
package tracker

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// MigrationFile is one details file read from the pinned main tree.
type MigrationFile struct {
	Path string // repository-relative
	Raw  []byte
}

// MigrationBranchFile is an issue file a branch changed relative to its merge
// base with main (Raw is nil when the branch deleted it).
type MigrationBranchFile struct {
	Branch, Path string
	Raw          []byte
}

// MigrationAnchor is a legacy close found for a codecomplete issue: the commit
// that recorded codecomplete on Ref, and whether code landed after it. OnMain
// marks main's own history, where a codecomplete is usually the legacy
// publication of a branch's close rather than a close of its own.
type MigrationAnchor struct {
	Ref, Anchor, Parent string
	CodeAfter, OnMain   bool
}

// MigrationInput is the inventory the command gathers.
type MigrationInput struct {
	Repository, ObjectFormat, Main string
	Active, Archived               []MigrationFile
	Branches                       []MigrationBranchFile
	DirtyIssuePaths                []string // "<worktree>: <path>" with uncommitted issue edits
	Anchors                        map[string][]MigrationAnchor
}

// MigrationCard is one card the tracker is bootstrapped with.
type MigrationCard struct {
	ID, Path, Source string
	Raw              []byte
	Inferences       []string
}

// MigrationConversion is one active details file main rewrites.
type MigrationConversion struct {
	Path string
	Raw  []byte
}

// MigrationRefusal is a reason the cutover cannot happen, with its remedy.
type MigrationRefusal struct {
	Subject, Reason, Next string
}

// MigrationManifest is the deterministic migration plan.
type MigrationManifest struct {
	Repository, ObjectFormat, Main string
	Cards                          []MigrationCard
	Conversions                    []MigrationConversion
	Duplicates                     []string
	Refusals                       []MigrationRefusal
	Digest                         string
}

// TrackerFiles is the tree the tracker is bootstrapped with.
func (m MigrationManifest) TrackerFiles() map[string][]byte {
	files := map[string][]byte{ManifestPath: ManifestBytes()}
	for _, c := range m.Cards {
		files[c.Path] = c.Raw
	}
	return files
}

// MigrationToken names the completion binding a migration imports for #id.
func MigrationToken(id string) string { return "migrate-" + id }

// PlanTrackerMigration computes the manifest for in. It never fails: every
// problem becomes a refusal, so one dry run reports all of them.
func PlanTrackerMigration(in MigrationInput) MigrationManifest {
	m := MigrationManifest{Repository: in.Repository, ObjectFormat: in.ObjectFormat, Main: in.Main}
	refuse := func(subject, reason, next string) {
		m.Refusals = append(m.Refusals, MigrationRefusal{Subject: subject, Reason: reason, Next: next})
	}
	type archived struct {
		file    MigrationFile
		created string
	}
	active := map[string][]MigrationFile{}
	older := map[string][]archived{}
	for _, f := range in.Active {
		id, _, ok := issue.ParseFilename(path.Base(f.Path))
		if !ok {
			continue
		}
		active[id] = append(active[id], f)
	}
	for _, f := range in.Archived {
		id, _, ok := issue.ParseFilename(path.Base(f.Path))
		if !ok {
			continue
		}
		created := ""
		if fm, _, err := issue.Parse(string(f.Raw)); err == nil {
			created, _ = issue.GetField(fm, "created")
		}
		older[id] = append(older[id], archived{f, created})
	}
	cards := map[string]MigrationCard{}
	for id, files := range active {
		if len(files) > 1 {
			for _, f := range files {
				refuse(f.Path, fmt.Sprintf("%d active details share ID %s", len(files), id), "renumber all but one before cutover")
			}
			continue
		}
		f := files[0]
		card, mirrored, inferences, err := issue.MigrateActiveDetails(f.Raw, in.ObjectFormat)
		if err != nil {
			refuse(f.Path, err.Error(), "fix the details under the old workflow, then re-run the dry run")
			continue
		}
		parsed, err := issue.ParseCard(card)
		if err != nil {
			refuse(f.Path, err.Error(), "fix the details under the old workflow, then re-run the dry run")
			continue
		}
		if parsed.ID != id {
			refuse(f.Path, fmt.Sprintf("frontmatter id %s disagrees with the filename's %s", parsed.ID, id), "correct the id before cutover")
			continue
		}
		if status, _ := issue.GetField(parsed.Frontmatter, "status"); status == "codecomplete" {
			bound, reason := bindLegacyClose(card, id, in)
			if reason != "" {
				refuse(f.Path, reason, "land the issue, or reopen and re-close it under the old workflow, before cutover")
				continue
			}
			card = bound
		}
		_, slug, _ := issue.ParseFilename(path.Base(f.Path))
		cards[id] = MigrationCard{ID: id, Path: CardPath(id, slug), Source: f.Path, Raw: card, Inferences: inferences}
		m.Conversions = append(m.Conversions, MigrationConversion{Path: f.Path, Raw: mirrored})
		for _, a := range older[id] {
			m.Duplicates = append(m.Duplicates, fmt.Sprintf("%s: card from %s; archived %s keeps its own record", id, f.Path, a.file.Path))
		}
	}
	for id, files := range older {
		if _, taken := active[id]; taken {
			continue
		}
		sort.Slice(files, func(i, j int) bool {
			if files[i].created != files[j].created {
				return files[i].created > files[j].created
			}
			return files[i].file.Path > files[j].file.Path
		})
		seed := files[0].file
		card, inferences, err := issue.ArchivedCard(seed.Raw, path.Base(seed.Path), in.ObjectFormat)
		if err != nil {
			refuse(seed.Path, err.Error(), "fix the archived details, then re-run the dry run")
			continue
		}
		_, slug, _ := issue.ParseFilename(path.Base(seed.Path))
		cards[id] = MigrationCard{ID: id, Path: CardPath(id, slug), Source: seed.Path, Raw: card, Inferences: inferences}
		for _, other := range files[1:] {
			m.Duplicates = append(m.Duplicates, fmt.Sprintf("%s: card from %s; archived %s keeps its own record", id, seed.Path, other.file.Path))
		}
	}
	// Stacked branches and their remote-tracking copies repeat one problem;
	// each (path, reason) is one refusal naming every branch that carries it.
	type branchProblem struct{ path, reason, next string }
	carriers := map[branchProblem][]string{}
	for _, b := range in.Branches {
		id, _, ok := issue.ParseFilename(path.Base(b.Path))
		if !ok || b.Raw == nil {
			continue // a branch's own archive move or an unrelated file
		}
		var p branchProblem
		card, known := cards[id]
		switch {
		case !known && len(active[id]) == 0 && len(older[id]) == 0:
			p = branchProblem{b.Path, "issue exists only on a branch", "publish it to main (legacy `sdlc issue sync --push`) or drop it, before cutover"}
		case len(active[id]) == 0:
			p = branchProblem{b.Path, "a branch edits an issue main has archived", "land or drop the branch's edit before cutover"}
		case known:
			if _, err := issue.ReconcileLegacyDetails(b.Raw, card.Raw, in.ObjectFormat); err != nil {
				p = branchProblem{b.Path, "unpublished card fields: " + err.Error(), "publish them to main (legacy `sdlc issue sync --push`) or revert them, before cutover"}
			}
		}
		if p.path != "" {
			carriers[p] = append(carriers[p], b.Branch)
		}
	}
	for p, branches := range carriers {
		sort.Strings(branches)
		refuse(p.path+" (on "+strings.Join(branches, ", ")+")", p.reason, p.next)
	}
	for _, d := range in.DirtyIssuePaths {
		refuse(d, "uncommitted issue edits", "commit and publish, or discard, them before cutover")
	}
	for _, c := range cards {
		m.Cards = append(m.Cards, c)
	}
	sort.Slice(m.Cards, func(i, j int) bool { return m.Cards[i].ID < m.Cards[j].ID })
	sort.Slice(m.Conversions, func(i, j int) bool { return m.Conversions[i].Path < m.Conversions[j].Path })
	sort.Strings(m.Duplicates)
	sort.Slice(m.Refusals, func(i, j int) bool {
		if m.Refusals[i].Subject != m.Refusals[j].Subject {
			return m.Refusals[i].Subject < m.Refusals[j].Subject
		}
		return m.Refusals[i].Reason < m.Refusals[j].Reason
	})
	m.Digest = m.digest()
	return m
}

// bindLegacyClose binds a codecomplete card to its legacy close: exactly one
// distinct branch anchor commit, with no code after it on the ref that carries
// it. Main's anchor is the close only when no branch carries one (a close made
// directly on main); otherwise it is the legacy publication of the branch's.
func bindLegacyClose(card []byte, id string, in MigrationInput) ([]byte, string) {
	candidates := []MigrationAnchor{}
	for _, a := range in.Anchors[id] {
		if !a.OnMain {
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		candidates = in.Anchors[id]
	}
	distinct := map[string]MigrationAnchor{}
	for _, a := range candidates {
		if a.CodeAfter {
			return nil, fmt.Sprintf("codecomplete, but %s has code after its close %s", a.Ref, short(a.Anchor))
		}
		distinct[a.Anchor] = a
	}
	if len(distinct) != 1 {
		return nil, fmt.Sprintf("codecomplete, but %d legacy close commits were found (exactly one is provable)", len(distinct))
	}
	for _, a := range distinct {
		bound, err := issue.SetCardCompletion(card, issue.Completion{Token: MigrationToken(id), Repository: in.Repository,
			ReviewedHEAD: a.Parent, EvidenceCommit: a.Anchor})
		if err != nil {
			return nil, "codecomplete binding: " + err.Error()
		}
		return bound, ""
	}
	return nil, "unreachable"
}

func short(oid string) string {
	if len(oid) > 10 {
		return oid[:10]
	}
	return oid
}

// digest identifies the plan by what it would do: repository, pinned main,
// every card and conversion by content, and every refusal.
func (m MigrationManifest) digest() string {
	type entry struct{ Path, SHA string }
	sum := func(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
	var doc struct {
		Repository, ObjectFormat, Main string
		Cards, Conversions             []entry
		Duplicates                     []string
		Refusals                       []MigrationRefusal
	}
	doc.Repository, doc.ObjectFormat, doc.Main = m.Repository, m.ObjectFormat, m.Main
	for _, c := range m.Cards {
		doc.Cards = append(doc.Cards, entry{c.Path, sum(c.Raw)})
	}
	for _, c := range m.Conversions {
		doc.Conversions = append(doc.Conversions, entry{c.Path, sum(c.Raw)})
	}
	doc.Duplicates, doc.Refusals = m.Duplicates, m.Refusals
	raw, _ := json.Marshal(doc)
	return sum(raw)[:16]
}
