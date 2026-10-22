package tracker

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"reflect"
	"strings"
	"testing"
)

func observed(id string, role issue.InventoryRole, oid string) MigrationObservedFile {
	root := "workshop/issues/"
	if role == issue.InventoryArchive {
		root = "workshop/history/issues/"
	}
	return MigrationObservedFile{Path: root + id + "-one.md", ID: id, Mode: "100644", OID: oid, Role: role}
}

func TestMigrationEvidenceRefusalsMatching(t *testing.T) {
	active, archived := observed("000001", issue.InventoryActive, oidA), observed("000002", issue.InventoryArchive, oidB)
	for _, tc := range []struct {
		name     string
		file     MigrationObservedFile
		ancestry MigrationAncestry
		refuse   bool
	}{
		{"related active prose", observed("000001", issue.InventoryActive, oidB), MigrationRelated, false},
		{"related active represented by archive", observed("000002", issue.InventoryActive, oidA), MigrationRelated, false},
		{"inherited unknown active", observed("000003", issue.InventoryActive, oidA), MigrationRelated, true},
		{"matching archive", archived, MigrationRelated, false},
		{"differing stale archive", observed("000002", issue.InventoryArchive, oidA), MigrationRelated, true},
		{"archive cannot match active", observed("000001", issue.InventoryArchive, oidA), MigrationRelated, true},
		{"unrelated matching active", active, MigrationUnrelated, false},
		{"unrelated differs", observed("000001", issue.InventoryActive, oidB), MigrationUnrelated, true},
		{"unrelated cannot match archive", observed("000002", issue.InventoryActive, oidB), MigrationUnrelated, true},
		{"unknown ancestry", active, "unknown", true},
		{"unknown role", observed("000001", "unknown", oidA), MigrationRelated, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := MigrationEvidenceRefusals([]MigrationObservedFile{active, archived}, []MigrationTipEvidence{{Tip: oidP, Carriers: []string{"branch"}, Ancestry: tc.ancestry, Files: []MigrationObservedFile{tc.file}}})
			if (len(got) > 0) != tc.refuse {
				t.Fatalf("refusals %+v", got)
			}
			if tc.refuse && (!strings.Contains(got[0].Subject, "branch") || got[0].Next == "") {
				t.Fatalf("missing remedy/carrier: %+v", got)
			}
		})
	}
	moved := archived
	moved.Path = "workshop/history/000002-other.md"
	if got := MigrationEvidenceRefusals([]MigrationObservedFile{archived}, []MigrationTipEvidence{{Tip: oidP, Ancestry: MigrationUnrelated, Files: []MigrationObservedFile{moved}}}); len(got) != 0 {
		t.Fatalf("path move refused: %+v", got)
	}
}

func TestMigrationEvidenceRefusalsDeterministicCarriers(t *testing.T) {
	f := observed("000001", issue.InventoryArchive, oidA)
	tips := []MigrationTipEvidence{
		{Tip: oidA, Carriers: []string{"z", "a", "a"}, Ancestry: MigrationRelated, Files: []MigrationObservedFile{f}},
		{Tip: oidB, Carriers: []string{"b"}, Ancestry: MigrationRelated, Files: []MigrationObservedFile{f}},
	}
	want := MigrationEvidenceRefusals(nil, tips)
	if len(want) != 1 || !strings.Contains(want[0].Subject, "a, b, z") {
		t.Fatalf("carriers %+v", want)
	}
	tips[0], tips[1] = tips[1], tips[0]
	if got := MigrationEvidenceRefusals(nil, tips); !reflect.DeepEqual(got, want) {
		t.Fatalf("order affects refusals: %+v != %+v", got, want)
	}
}

func TestMigrationInventoryComposesAndCannotExcuseActiveDeletion(t *testing.T) {
	in := basicInput()
	in.Active = []MigrationFile{activeFile("000001", "one", "open")}
	a := observed("000001", issue.InventoryActive, oidA)
	in.Published = []MigrationObservedFile{a, observed("000001", issue.InventoryArchive, oidB)}
	in.Tips = []MigrationTipEvidence{{Tip: oidP, Carriers: []string{"branch"}, Ancestry: MigrationRelated, Files: []MigrationObservedFile{observed("000001", issue.InventoryArchive, oidB)}}}
	in.Branches = []MigrationBranchFile{{Branch: "branch", Path: a.Path, Raw: nil}}
	m := PlanTrackerMigration(in)
	if len(m.Refusals) != 1 || m.Refusals[0].Reason != ArchivesActiveReason {
		t.Fatalf("archive excused deletion: %+v", m.Refusals)
	}
	in.Branches = nil
	in.Tips[0].Files = []MigrationObservedFile{observed("000002", issue.InventoryArchive, oidB)}
	m = PlanTrackerMigration(in)
	if len(m.Refusals) != 1 {
		t.Fatalf("evidence omitted: %+v", m.Refusals)
	}
	in.Tips = nil
	if PlanTrackerMigration(in).Digest == m.Digest {
		t.Fatal("digest ignored evidence refusal")
	}
}

func TestMigrationEvidenceMalformedMetadata(t *testing.T) {
	f := observed("000001", issue.InventoryActive, oidA)
	for _, mutate := range []func(*MigrationObservedFile){
		func(f *MigrationObservedFile) { f.ID = "000002" }, func(f *MigrationObservedFile) { f.OID = "bad" },
		func(f *MigrationObservedFile) { f.Mode = "120000" }, func(f *MigrationObservedFile) { f.Path = "../" + f.Path },
		func(f *MigrationObservedFile) { f.Role = "" },
	} {
		broken := f
		mutate(&broken)
		if got := MigrationEvidenceRefusals([]MigrationObservedFile{f}, []MigrationTipEvidence{{Tip: oidP, Ancestry: MigrationRelated, Files: []MigrationObservedFile{broken}}}); len(got) == 0 {
			t.Fatalf("malformed observed file accepted: %+v", broken)
		}
		if got := MigrationEvidenceRefusals([]MigrationObservedFile{broken}, nil); len(got) == 0 {
			t.Fatalf("malformed published file accepted: %+v", broken)
		}
	}
	if got := MigrationEvidenceRefusals(nil, []MigrationTipEvidence{{Tip: oidP, Ancestry: ""}}); len(got) == 0 {
		t.Fatal("empty unknown tip accepted")
	}
}

// Git and the repository filename grammar treat a POSIX backslash as a literal
// filename byte. It cannot hide an issue or invalidate exact published evidence.
func TestMigrationEvidencePOSIXBackslash(t *testing.T) {
	for _, role := range []issue.InventoryRole{issue.InventoryActive, issue.InventoryArchive} {
		f := observed("000001", role, oidA)
		f.Path = strings.TrimSuffix(f.Path, ".md") + "\\literal.md"
		for _, ancestry := range []MigrationAncestry{MigrationRelated, MigrationUnrelated} {
			tip := MigrationTipEvidence{Tip: oidP, Ancestry: ancestry, Files: []MigrationObservedFile{f}}
			if got := MigrationEvidenceRefusals([]MigrationObservedFile{f}, []MigrationTipEvidence{tip}); len(got) != 0 {
				t.Errorf("%s/%s matching literal backslash evidence refused: %+v", role, ancestry, got)
			}
			if got := MigrationEvidenceRefusals(nil, []MigrationTipEvidence{tip}); len(got) == 0 {
				t.Errorf("%s/%s unpublished literal backslash evidence omitted", role, ancestry)
			}
		}
	}
}
