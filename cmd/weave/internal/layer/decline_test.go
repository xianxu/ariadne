package layer

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
)

func row(k intent.Kind, target string) intent.Intent {
	return intent.Intent{Kind: k, Source: target, Target: target}
}

func TestDeclineDropsInheritedRowsUnderThePath(t *testing.T) {
	foundation := Layer{Name: "ariadne", Intents: []intent.Intent{
		row(intent.Seed, ".github/workflows/merge-check.yml"),
		row(intent.Seed, ".github/workflows/local-gates-guard.yml"),
		row(intent.Seed, "bootstrap.sh"),
		row(intent.Symlink, ".github/workflowsX"), // a sibling sharing the prefix
		row(intent.Prose, "AGENTS.base.md"),       // no single target: never declined
		{Kind: intent.Merge, Source: "s.json", Target: ".github/workflows/settings.json"},
	}}
	leaf := Layer{Name: "ops", Intents: []intent.Intent{
		row(intent.Decline, ".github/workflows"),
		row(intent.Seed, ".github/workflows/own.yml"), // the leaf's own row is never declined
	}}
	got := Decline([]Layer{foundation, leaf})

	wantKept := []intent.Intent{
		row(intent.Seed, "bootstrap.sh"),
		row(intent.Symlink, ".github/workflowsX"),
		row(intent.Prose, "AGENTS.base.md"),
	}
	if !reflect.DeepEqual(got[0].Intents, wantKept) {
		t.Fatalf("foundation intents = %v, want %v", got[0].Intents, wantKept)
	}
	if !reflect.DeepEqual(got[1].Intents, leaf.Intents) {
		t.Fatalf("leaf intents changed: %v", got[1].Intents)
	}
	want := []Declination{{Path: ".github/workflows", Removed: []intent.Intent{
		row(intent.Seed, ".github/workflows/merge-check.yml"),
		row(intent.Seed, ".github/workflows/local-gates-guard.yml"),
		// A merge into a declined path is removed too, and named in the
		// report, never dropped silently.
		{Kind: intent.Merge, Source: "s.json", Target: ".github/workflows/settings.json"},
	}}}
	if !reflect.DeepEqual(got[1].Declined, want) {
		t.Fatalf("declined = %+v, want %+v", got[1].Declined, want)
	}
}

func TestDeclineInAnAncestorIsIgnored(t *testing.T) {
	foundation := Layer{Intents: []intent.Intent{
		row(intent.Decline, ".github/workflows"),
		row(intent.Seed, ".github/workflows/merge-check.yml"),
	}}
	middle := Layer{Intents: []intent.Intent{row(intent.Seed, ".github/workflows/x.yml")}}
	leaf := Layer{}
	got := Decline([]Layer{foundation, middle, leaf})
	if len(got[0].Intents) != 2 || len(got[1].Intents) != 1 || got[2].Declined != nil {
		t.Fatalf("an ancestor's decline took effect: %+v", got)
	}
}

func TestDeclineMatchingNothingIsRecordedEmpty(t *testing.T) {
	leaf := Layer{Intents: []intent.Intent{row(intent.Decline, ".github/workflows/")}}
	got := Decline([]Layer{{}, leaf})
	want := []Declination{{Path: ".github/workflows"}}
	if !reflect.DeepEqual(got[1].Declined, want) {
		t.Fatalf("declined = %+v, want %+v", got[1].Declined, want)
	}
}

func TestDeclineExactTargetMatches(t *testing.T) {
	foundation := Layer{Intents: []intent.Intent{row(intent.Seed, "bootstrap.sh")}}
	leaf := Layer{Intents: []intent.Intent{row(intent.Decline, "bootstrap.sh")}}
	got := Decline([]Layer{foundation, leaf})
	if len(got[0].Intents) != 0 || len(got[1].Declined[0].Removed) != 1 {
		t.Fatalf("exact-path decline missed: %+v", got)
	}
}
