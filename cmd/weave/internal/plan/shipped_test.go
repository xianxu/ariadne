package plan

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
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
