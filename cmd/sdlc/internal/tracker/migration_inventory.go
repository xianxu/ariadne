package tracker

import (
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// MigrationObservedFile is metadata from a pinned issue document.
type MigrationObservedFile struct {
	Path, ID, Mode, OID string
	Role                issue.InventoryRole
}

// MigrationAncestry records the adapter's known relationship to main.
type MigrationAncestry string

const (
	MigrationRelated   MigrationAncestry = "related"
	MigrationUnrelated MigrationAncestry = "unrelated"
)

// MigrationTipEvidence keeps all diagnostic names for an immutable tree.
type MigrationTipEvidence struct {
	Tip      string
	Carriers []string
	Ancestry MigrationAncestry
	Files    []MigrationObservedFile
}

// MigrationEvidenceRefusals checks non-main evidence without importing records.
func MigrationEvidenceRefusals(published []MigrationObservedFile, tips []MigrationTipEvidence) []MigrationRefusal {
	type key struct {
		role    issue.InventoryRole
		id, oid string
	}
	type problem struct{ path, reason string }
	problems := map[problem]map[string]bool{}
	add := func(p, reason string, carriers []string) {
		k := problem{p, reason}
		if problems[k] == nil {
			problems[k] = map[string]bool{}
		}
		for _, c := range carriers {
			problems[k][c] = true
		}
	}
	valid := func(f MigrationObservedFile) bool {
		id, _, ok := issue.ParseFilename(path.Base(f.Path))
		return fs.ValidPath(f.Path) && !strings.ContainsAny(f.Path, "\\\x00") && ok && id == f.ID &&
			(f.Role == issue.InventoryActive || f.Role == issue.InventoryArchive) &&
			(f.Mode == "100644" || f.Mode == "100755") && validReceiptOID(f.OID)
	}
	ids := map[string]bool{}
	publishedKeys := map[key]bool{}
	for _, f := range published {
		if !valid(f) {
			add(f.Path, "malformed published issue inventory metadata", []string{"published main"})
			continue
		}
		ids[f.ID] = true
		publishedKeys[key{f.Role, f.ID, f.OID}] = true
	}
	for _, tip := range tips {
		carriers := tip.Carriers
		if len(carriers) == 0 {
			carriers = []string{tip.Tip}
		}
		if tip.Ancestry != MigrationRelated && tip.Ancestry != MigrationUnrelated {
			add(tip.Tip, "unknown issue inventory ancestry", carriers)
			continue
		}
		for _, f := range tip.Files {
			if !valid(f) {
				add(f.Path, "malformed observed issue inventory metadata", carriers)
				continue
			}
			switch {
			case f.Role == issue.InventoryArchive:
				if !publishedKeys[key{issue.InventoryArchive, f.ID, f.OID}] {
					add(f.Path, "archive issue record is not published on main", carriers)
				}
			case !ids[f.ID]:
				add(f.Path, "issue ID exists only outside main", carriers)
			case tip.Ancestry == MigrationUnrelated && !publishedKeys[key{issue.InventoryActive, f.ID, f.OID}]:
				add(f.Path, "unrelated active issue record is not published on main", carriers)
			}
		}
	}
	var refusals []MigrationRefusal
	for p, names := range problems {
		carriers := make([]string, 0, len(names))
		for name := range names {
			carriers = append(carriers, name)
		}
		sort.Strings(carriers)
		refusals = append(refusals, MigrationRefusal{
			Subject: p.path + " (on " + strings.Join(carriers, ", ") + ")", Reason: p.reason,
			Next: "resolve the named legacy evidence: publish, align or drop it under the old workflow, then re-run the dry run before cutover",
		})
	}
	sort.Slice(refusals, func(i, j int) bool {
		if refusals[i].Subject != refusals[j].Subject {
			return refusals[i].Subject < refusals[j].Subject
		}
		return refusals[i].Reason < refusals[j].Reason
	})
	return refusals
}
