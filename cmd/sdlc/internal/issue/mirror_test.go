package issue

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRefreshMirrorStaleProjectionAndEditableProblem(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	current := strings.Replace(string(baseline), "status: open", "status: working\nstarted: 2026-09-25T14:00:00-07:00", 1)
	current = strings.Replace(current, "# Card title", "# New title", 1)
	local := strings.Replace(string(detail), "Original report.", "Locally revised report.", 1)
	got, err := RefreshMirror([]byte(local), baseline, []byte(current))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "status: working") || !strings.Contains(string(got), "# New title") || !strings.Contains(string(got), "Locally revised report.") {
		t.Fatalf("refresh: %s", got)
	}
	_, oldBody, _ := Parse(local)
	_, newBody, _ := Parse(string(got))
	if newBody != strings.Replace(oldBody, "# Card title", "# New title", 1) {
		t.Fatal("refresh changed unrelated body bytes")
	}
	if _, err := RefreshMirror(got, []byte(current), []byte(current)); err != nil {
		t.Fatal(err)
	}
}

func TestMirrorSHA256AndOIDReader(t *testing.T) {
	card, detail, err := SplitCardWithFormat([]byte(cardDetailFixture), "sha256")
	if err != nil {
		t.Fatal(err)
	}
	oid, err := MirrorBaselineOID(detail)
	if err != nil {
		t.Fatal(err)
	}
	known := sha256.Sum256(append([]byte(fmt.Sprintf("blob %d\x00", len(card))), card...))
	if oid != fmt.Sprintf("%x", known) {
		t.Fatalf("wrong Git SHA256: %s", oid)
	}
	if got, err := RefreshMirror(detail, card, card); err != nil || string(got) != string(detail) {
		t.Fatalf("SHA256 refresh: %v", err)
	}
	if _, _, err := SplitCardWithFormat([]byte(cardDetailFixture), "unknown"); err == nil {
		t.Fatal("unsupported format accepted")
	}
	if _, err := MirrorBaselineOID([]byte(cardDetailFixture)); err == nil {
		t.Fatal("missing marker accepted")
	}
}

func TestMirrorUnownedBytesAndFieldRemoval(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	current := strings.Replace(string(baseline), "estimate_hours:\n", "", 1)
	current = strings.Replace(current, "status: open", "status: working", 1)
	got, err := RefreshMirror(detail, baseline, []byte(current))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "estimate_hours:") {
		t.Fatal("removed field remained present")
	}
	unowned := "deps: [pair#17]\n# preserve this comment\ntarget: 'stable'\nflow: {kind: quick, provenance: operator}\ncustom:\n  nested: [a, b]"
	if !strings.Contains(string(got), unowned) {
		t.Fatalf("changed unowned YAML: %s", got)
	}
}

func TestMirrorEquivalentFormattingIsNotAHandEdit(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	local := strings.Replace(string(detail), "status: open", "status: 'open' # explanatory comment", 1)
	got, err := RefreshMirror([]byte(local), baseline, baseline)
	if err != nil || string(got) != local {
		t.Fatalf("same-value formatting is not an edit: %v", err)
	}
}

func FuzzRefreshMirror(f *testing.F) {
	f.Add("local text", false)
	f.Add("## Quoted\n~~~md\n# example\n~~~", true)
	f.Fuzz(func(t *testing.T, localProblem string, handEdit bool) {
		baseline, detail, err := SplitCard([]byte(cardDetailFixture))
		if err != nil {
			t.Fatal(err)
		}
		// A YAML literal scalar is branch-owned and can contain arbitrary bytes
		// only when well-formed UTF-8/Markdown; malformed documents may refuse.
		local := strings.Replace(string(detail), "Original report.", localProblem, 1)
		if handEdit {
			local = strings.Replace(local, "status: open", "status: blocked", 1)
		}
		current := strings.Replace(string(baseline), "status: open", "status: working", 1)
		got, err := RefreshMirror([]byte(local), baseline, []byte(current))
		if handEdit && err == nil {
			t.Fatal("accepted an edited protected status")
		}
		if err != nil {
			return
		}
		_, beforeBody, _ := Parse(local)
		_, afterBody, _ := Parse(string(got))
		if beforeBody != afterBody {
			t.Fatal("refresh changed branch-owned body")
		}
	})
}

func TestRefreshMirrorRefusesHandEdits(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	for field, local := range map[string]string{
		"status":         strings.Replace(string(detail), "status: open", "status: working", 1),
		"title":          strings.Replace(string(detail), "# Card title", "# Hand edit", 1),
		"estimate_hours": strings.Replace(string(detail), "estimate_hours:\n", "", 1),
		"github_issue":   strings.Replace(string(detail), "github_issue:", "github_issue: ''", 1),
	} {
		t.Run(field, func(t *testing.T) {
			_, err := RefreshMirror([]byte(local), baseline, baseline)
			var owned *OwnershipError
			if !errors.As(err, &owned) || owned.Field != field || !strings.Contains(err.Error(), "sdlc") {
				t.Fatalf("want actionable %s ownership error, got %v", field, err)
			}
		})
	}
}

func TestRefreshMirrorRefusesUnprovenBaseline(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	for name, baselineBytes := range map[string][]byte{"missing": nil, "different bytes": append(append([]byte{}, baseline...), '\n'), "invalid": []byte("bad")} {
		t.Run(name, func(t *testing.T) {
			if _, err := RefreshMirror(detail, baselineBytes, baseline); err == nil {
				t.Fatal("accepted unproven baseline")
			}
		})
	}
	if _, err := RefreshMirror([]byte(cardDetailFixture), baseline, baseline); err == nil {
		t.Fatal("accepted missing marker")
	}
	other := strings.Replace(string(baseline), "000252", "000253", 1)
	if _, err := RefreshMirror(detail, baseline, []byte(other)); err == nil {
		t.Fatal("accepted different issue")
	}
}

func TestRefreshMirrorPreservesScalarTypes(t *testing.T) {
	input := strings.Replace(cardDetailFixture, "github_issue:\n", "github_issue: 12\n", 1)
	baseline, detail, err := SplitCard([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	local := strings.Replace(string(detail), "github_issue: 12", "github_issue: '12'", 1)
	if _, err := RefreshMirror([]byte(local), baseline, baseline); err == nil {
		t.Fatal("accepted changed YAML scalar type")
	}
}
