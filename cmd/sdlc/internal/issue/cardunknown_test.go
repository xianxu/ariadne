package issue

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func unknownCardValues(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	fm, _, err := Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := yaml.Unmarshal([]byte(fm), &root); err != nil {
		t.Fatal(err)
	}
	claimant := root["claimant"].(map[string]any)
	tracker := root["tracker"].(map[string]any)
	completion := tracker["completion"].(map[string]any)
	return map[string]any{
		"top":        root["future_top"],
		"claimant":   claimant["future_claimant"],
		"tracker":    tracker["future_tracker"],
		"completion": completion["future_completion"],
	}
}

// A newer binary may add fields while an older sdlc merge remains alive. The
// older process must keep those bytes when it changes unrelated known fields.
func TestOlderCardReaderPreservesUnknownFieldsThroughSetters(t *testing.T) {
	card, err := SetCardClaimant([]byte(claimantCard), testClaimant())
	if err != nil {
		t.Fatal(err)
	}
	card, err = SetCardCompletion(card, testCompletion("close-1"))
	if err != nil {
		t.Fatal(err)
	}
	raw := string(card)
	raw = strings.Replace(raw, "---\n\n# Ownership", "future_top: {nested: [one, two]}\n---\n\n# Ownership", 1)
	raw = strings.Replace(raw, "    operator: Xian Xu\n", "    operator: Xian Xu\n    future_claimant: {nested: [one, two]}\n", 1)
	raw = strings.Replace(raw, "    version: 1\n", "    version: 1\n    future_tracker: {nested: [one, two]}\n", 1)
	raw = strings.Replace(raw, "        token: close-1\n", "        token: close-1\n        future_completion: {nested: [one, two]}\n", 1)
	for _, field := range []string{"future_top", "future_claimant", "future_tracker", "future_completion"} {
		if !strings.Contains(raw, field+":") {
			t.Fatalf("fixture omitted %s:\n%s", field, raw)
		}
	}
	if _, err := ParseCard([]byte(raw)); err != nil {
		t.Fatalf("old reader refused a newer card: %v", err)
	}
	wantUnknown := unknownCardValues(t, []byte(raw))
	updates := []struct {
		name string
		fn   func([]byte) ([]byte, error)
	}{
		{"scalar", func(b []byte) ([]byte, error) { return SetCardField(b, "updated", "2026-10-19") }},
		{"claimant", func(b []byte) ([]byte, error) { return SetCardClaimant(b, testClaimant()) }},
		{"completion", func(b []byte) ([]byte, error) {
			c := testCompletion("close-1")
			c.LandedCommit = strings.Repeat("e", 40)
			return SetCardCompletion(b, c)
		}},
		{"handoff", func(b []byte) ([]byte, error) { return SetCardHandoff(b, testHandoff()) }},
		{"title", func(b []byte) ([]byte, error) { return SetCardTitle(b, "New title") }},
		{"release", func(b []byte) ([]byte, error) {
			return SetCardRelease(b, &Release{By: ReleasedBy(testClaimant())})
		}},
		{"abandoned", func(b []byte) ([]byte, error) { return SetCardAbandoned(b, &Abandoned{}) }},
		{"actual correction", func(b []byte) ([]byte, error) {
			c := testCompletion("close-1")
			inputs := `{"version":1,"repository":"` + c.Repository + `","code_head":"` + c.ReviewedHEAD + `"}`
			record := ActualCorrection{Version: 1, Token: "repair-1", CompletionToken: c.Token, Repository: c.Repository, EvidenceCommit: c.EvidenceCommit, ReviewedHEAD: c.ReviewedHEAD, PreviousActual: "1", NewActual: "2", Reason: "correction", Inputs: inputs, MeasurementDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(inputs)))}
			return SetCardActualCorrection(b, record)
		}},
	}
	for _, update := range updates {
		card, err = update.fn([]byte(raw))
		if err != nil {
			t.Fatalf("%s: %v", update.name, err)
		}
		for _, field := range []string{"future_top", "future_claimant", "future_tracker", "future_completion"} {
			if !strings.Contains(string(card), field+":") {
				t.Errorf("%s dropped %s:\n%s", update.name, field, card)
			}
		}
		if got := unknownCardValues(t, card); !reflect.DeepEqual(got, wantUnknown) {
			t.Errorf("%s changed future values: got %#v, want %#v", update.name, got, wantUnknown)
		}
	}
	cleared, err := ClearCardClaimant([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cleared), "claimant:") || !strings.Contains(string(cleared), "future_top:") || !strings.Contains(string(cleared), "future_tracker:") || !strings.Contains(string(cleared), "future_completion:") {
		t.Fatalf("clear claimant changed unrelated future fields:\n%s", cleared)
	}
}

func TestMergeUnknownCardNodesKeepsFutureTreesWithoutRestoringRemovedKnownRecords(t *testing.T) {
	for _, future := range []string{"42", "[one, two]", "{deep: {leaf: [one, two]}}"} {
		t.Run(future, func(t *testing.T) {
			var old, next yaml.Node
			before := "version: 1\nfuture: " + future + "\ncompletion: {token: old, future: " + future + "}\nrelease: {branch: old}\n"
			after := "version: 1\ncompletion: {token: new}\n"
			if err := yaml.Unmarshal([]byte(before), &old); err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal([]byte(after), &next); err != nil {
				t.Fatal(err)
			}
			mergeUnknownCardNodes(old.Content[0], next.Content[0], reflect.TypeOf(trackerEnvelope{}))
			out, err := yaml.Marshal(next.Content[0])
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"future: " + future, "token: new"} {
				if !strings.Contains(string(out), want) {
					t.Errorf("missing %q:\n%s", want, out)
				}
			}
			if strings.Contains(string(out), "release:") || strings.Contains(string(out), "token: old") {
				t.Errorf("known removed or replaced value was resurrected:\n%s", out)
			}
		})
	}
}

func TestReleaseRewritePreservesUnknownGrandchild(t *testing.T) {
	card, err := SetCardRelease([]byte(claimantCard), &Release{By: ReleasedBy(testClaimant())})
	if err != nil {
		t.Fatal(err)
	}
	raw := strings.Replace(string(card), "        repository: github.com/xianxu/ariadne\n", "        repository: github.com/xianxu/ariadne\n        future_by: {nested: [one, two]}\n", 1)
	if !strings.Contains(raw, "future_by:") {
		t.Fatal("fixture omitted nested future member")
	}
	out, err := SetCardRelease([]byte(raw), &Release{By: ReleasedBy(testClaimant())})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "future_by: {nested: [one, two]}") {
		t.Fatalf("release rewrite dropped unknown grandchild:\n%s", out)
	}
}

func TestParseCardRejectsMalformedKnownTrackerRecord(t *testing.T) {
	card, err := SetCardCompletion([]byte(claimantCard), testCompletion("close-1"))
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"missing token":   strings.Replace(string(card), "        token: close-1\n", "", 1),
		"wrong head type": strings.Replace(string(card), "        reviewed_head: "+strings.Repeat("c", 40), "        reviewed_head: [bad]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCard([]byte(raw)); err == nil {
				t.Fatal("card decode admitted malformed known tracker record")
			}
		})
	}
}
