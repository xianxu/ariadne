package plan

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

const shippedIssue = "construct/generated/vocabulary/issue.json"

func shipRow(v intent.Visibility, p string) intent.Intent {
	return intent.Intent{Kind: intent.Ship, Visibility: v, Source: p, Target: p}
}

// ShippedPaths takes the selected `ship` rows: every layer's exports and the
// leaf's internals; an ancestor's internal row never reaches the leaf, and an
// unsafe path refuses.
func TestShippedPaths(t *testing.T) {
	layers := []layer.Layer{
		{Path: "/base", Intents: []intent.Intent{shipRow(intent.Export, "a/exported.json"), shipRow(intent.Internal, "a/base-only.json")}},
		{Path: "/leaf", Intents: []intent.Intent{shipRow(intent.Internal, shippedIssue), {Kind: intent.Symlink, Source: "x", Target: "x"}}},
	}
	got, err := ShippedPaths(layers)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a/exported.json", shippedIssue}; !reflect.DeepEqual(got, want) {
		t.Errorf("ShippedPaths = %v, want %v", got, want)
	}
	for _, bad := range []string{"../x.json", "/abs.json"} {
		_, err := ShippedPaths([]layer.Layer{{Intents: []intent.Intent{shipRow(intent.Internal, bad)}}})
		if err == nil || !strings.Contains(err.Error(), "ship") {
			t.Errorf("ship %q: err = %v, want a refusal", bad, err)
		}
	}
}

// ShipStaged marks each declared staged file and adds exactly one sidecar
// holding its digest; an undeclared output is untouched and a declared path
// nothing produces comes back unmatched.
func TestShipStaged(t *testing.T) {
	root := t.TempDir()
	stage, actions := stagedFixture(t, root)
	defer RemoveGenerationStage(stage)
	missing := "construct/generated/vocabulary/missing.json"
	got, unmatched := ShipStaged(actions, []string{shippedIssue, missing})
	if !reflect.DeepEqual(unmatched, []string{missing}) {
		t.Errorf("unmatched = %v", unmatched)
	}
	if len(got) != len(actions)+1 {
		t.Fatalf("got %d actions, want %d (one sidecar)", len(got), len(actions)+1)
	}
	var marked int
	for _, a := range got[:len(actions)] {
		s, ok := a.(stagedOutput)
		if !ok {
			t.Fatalf("%T is not staged", a)
		}
		w, _ := s.Action.(WriteFile)
		if (w.Path == shippedIssue) != (s.Sidecar != "") {
			t.Errorf("%s: Sidecar = %q", w.Path, s.Sidecar)
		}
		if s.Sidecar != "" {
			marked++
		}
	}
	side, ok := got[len(actions)].(WriteFile)
	if marked != 1 || !ok || side.Path != SidecarPath(shippedIssue) || side.Content != digest([]byte("generated issue.json"))+"\n" {
		t.Errorf("marked=%d sidecar=%+v", marked, got[len(actions)])
	}
}

// The ownership matrix for a shipped output: (output bytes) × (sidecar) ×
// (inventory: none, a cold checkout; or recorded from an earlier apply) ×
// (declared or not). The guard: weave overwrites only what it can prove it
// wrote (the inventory, or an output matching its committed sidecar);
// otherwise a cold checkout refuses and a warm one displaces, as before.
func TestShippedOutputOwnership(t *testing.T) {
	const generated = "generated issue.json"
	sum := func(s string) string { return digest([]byte(s)) + "\n" }
	for _, tc := range []struct {
		name            string
		declared, warm  bool
		output, sidecar string // "" sidecar: none on disk
		want            string // regenerated | refused | displaced
	}{
		{"cold parley case: committed pair from an older vocabulary", true, false, "old", sum("old"), "regenerated"},
		{"cold hand edit", true, false, "edited", sum("old"), "refused"},
		{"cold, no sidecar", true, false, "old", "", "refused"},
		{"cold, malformed sidecar", true, false, "old", "not a digest\n", "refused"},
		{"cold, undeclared", false, false, "old", sum("old"), "refused"},
		{"warm, pulled a newer committed pair", true, true, "pulled", sum("pulled"), "regenerated"},
		{"warm hand edit", true, true, "edited", sum(generated), "displaced"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			stage, actions := stagedFixture(t, root)
			defer RemoveGenerationStage(stage)
			if tc.declared {
				actions, _ = ShipStaged(actions, []string{shippedIssue})
			}
			if tc.warm {
				managedApply(t, root, actions, ScopeArtifacts)
			}
			managedWrite(t, root, shippedIssue, tc.output)
			side := SidecarPath(shippedIssue)
			if tc.sidecar != "" {
				managedWrite(t, root, side, tc.sidecar)
			}
			report, err := ApplyManaged(weavefs.OSFS{}, root, actions, ScopeArtifacts)
			switch tc.want {
			case "refused":
				if err == nil || !strings.Contains(err.Error(), "occupied by something weave did not write") {
					t.Fatalf("err = %v, want the refusal", err)
				}
				if got := managedRead(t, root, shippedIssue); got != tc.output {
					t.Errorf("a refused apply changed the output to %q", got)
				}
			case "regenerated":
				if err != nil {
					t.Fatal(err)
				}
				if got := managedRead(t, root, shippedIssue); got != generated {
					t.Errorf("output = %q, want the generated bytes", got)
				}
				if got := managedRead(t, root, side); got != sum(generated) {
					t.Errorf("sidecar = %q, want the new digest", got)
				}
				if len(report.Displaced) != 0 || !slices.Contains(report.Regenerated, shippedIssue) {
					t.Errorf("report = %+v, want issue.json regenerated and nothing displaced", report)
				}
			case "displaced":
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Contains(report.Displaced, shippedIssue) {
					t.Errorf("report = %+v, want the hand edit displaced", report)
				}
				if got := managedRead(t, root, DisplacedPath(shippedIssue)); got != tc.output {
					t.Errorf("displaced copy = %q, want the hand edit", got)
				}
			}
		})
	}
}
