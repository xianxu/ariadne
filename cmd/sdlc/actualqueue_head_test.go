package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var queueHeadTestOID = strings.Repeat("a", 40)
var queueHeadTestParent = strings.Repeat("b", 40)
var queueHeadTestMain = strings.Repeat("c", 40)

func TestParseQueueHeadObject(t *testing.T) {
	raw := "tree " + queueHeadTestMain + "\nparent " + queueHeadTestParent + "\nauthor A <a@b> 1 +0000\ncommitter A <a@b> 1 +0000\ngpgsig signature\n continuation\n\n#3: work\n\nbody\n"
	f, err := parseQueueHeadObject(queueHeadTestOID, []byte(raw))
	if err != nil || f.OID != queueHeadTestOID || !reflect.DeepEqual(f.Parents, []string{queueHeadTestParent}) || f.Message != "#3: work\n\nbody\n" {
		t.Fatalf("fact=%+v err=%v", f, err)
	}
	for name, bad := range map[string]string{
		"missing separator":   strings.TrimSuffix(raw, "\n\n#3: work\n\nbody\n"),
		"short parent":        strings.Replace(raw, queueHeadTestParent, "abc", 1),
		"duplicate tree":      "tree " + queueHeadTestMain + "\n" + raw,
		"bad header":          strings.Replace(raw, "author A <a@b> 1 +0000", "author", 1),
		"orphan continuation": " continuation\n" + raw,
		"missing committer":   strings.Replace(raw, "committer A <a@b> 1 +0000\n", "", 1),
		"nul":                 raw + "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseQueueHeadObject(queueHeadTestOID, []byte(bad)); err == nil {
				t.Fatal("accepted malformed object")
			}
		})
	}
	if _, err := parseQueueHeadObject("abc", []byte(raw)); err == nil {
		t.Fatal("accepted abbreviated requested OID")
	}
}

func TestParseQueueHeadAuthority(t *testing.T) {
	integration := fmt.Sprintf("Merge %s into 000003-topic (sdlc merge, #332)\n\nIntegrated-Main: %s\nAgent: human\n", queueHeadTestMain, queueHeadTestMain)
	car := "Landing train 7: #3 (PR #99)\n\nLanding-Train: 7\nLanding-PR: 99\nAgent: human\n"
	reintegration := fmt.Sprintf("Merge %s into landing train 7 (sdlc merge, #449)\n\nIntegrated-Main: %s\nLanding-Train: 7\n", queueHeadTestMain, queueHeadTestMain)
	for _, tt := range []struct {
		name, message string
		parents       []string
		kind          queueHeadKind
		issues        []int
		train         int
	}{
		{"local lead", "#3: work (see #9)", nil, queueHeadDirect, []int{3}, 0},
		{"qualified chained", "sdlc: ariadne#3, #5: work", nil, queueHeadDirect, []int{3, 5}, 0},
		{"foreign lead", "pair#3: work", nil, queueHeadDirect, nil, 0},
		{"body citation", "ordinary work\n\n#3: citation", nil, queueHeadUnknown, nil, 0},
		{"sync", "issue-sync: update issues", []string{queueHeadTestParent}, queueHeadNeutral, nil, 0},
		{"archive", "archive completed issues to history", []string{queueHeadTestParent}, queueHeadNeutral, nil, 0},
		{"almost sync", "issue-sync: update issues #3", []string{queueHeadTestParent}, queueHeadUnknown, nil, 0},
		{"integration", integration, []string{queueHeadTestParent, queueHeadTestMain}, queueHeadNeutral, nil, 0},
		{"bad integration parent", integration, []string{queueHeadTestMain, queueHeadTestParent}, queueHeadUnknown, nil, 0},
		{"car", car, []string{queueHeadTestParent, queueHeadTestMain}, queueHeadTrainCar, []int{3}, 7},
		{"reintegration", reintegration, []string{queueHeadTestParent, queueHeadTestMain}, queueHeadTrainIntegration, nil, 7},
		{"bad car trailer", strings.Replace(car, "Landing-PR: 99", "Landing-PR: 98", 1), []string{queueHeadTestParent, queueHeadTestMain}, queueHeadMalformedTrain, nil, 0},
		{"duplicate train", car + "Landing-Train: 7\n", []string{queueHeadTestParent, queueHeadTestMain}, queueHeadMalformedTrain, nil, 0},
		{"car one parent", car, []string{queueHeadTestParent}, queueHeadMalformedTrain, nil, 0},
		{"train metadata cannot decorate lead", "#3: work\n\nLanding-Train: 7", nil, queueHeadMalformedTrain, nil, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := parseQueueHead(queueHeadFact{OID: queueHeadTestOID, Message: tt.message, Parents: tt.parents}, "ariadne")
			if got.Kind != tt.kind || !reflect.DeepEqual(got.Issues, tt.issues) || got.Train != tt.train {
				t.Fatalf("got %+v want kind %v issues %v train %v", got, tt.kind, tt.issues, tt.train)
			}
			if tt.kind == queueHeadNeutral || tt.kind == queueHeadTrainCar || tt.kind == queueHeadTrainIntegration {
				if got.NeedParent != queueHeadTestParent {
					t.Fatalf("parent=%s", got.NeedParent)
				}
			}
		})
	}
}

func TestSelectQueueHeadMembership(t *testing.T) {
	direct := queueHeadStep{Kind: queueHeadDirect, Issues: []int{9}}
	known := queueHeadMembership{Known: true, Issues: []int{9}}
	unknown := queueHeadMembership{Reason: "missing object"}
	car := queueHeadStep{Kind: queueHeadTrainCar, Train: 7, Issues: []int{3}, NeedParent: queueHeadTestParent}
	prefix := queueHeadStep{Kind: queueHeadTrainCar, Train: 7, Issues: []int{5}}
	prefixMembership := queueHeadMembership{Known: true, Train: 7, Issues: []int{5}}
	different := queueHeadStep{Kind: queueHeadTrainCar, Train: 8}
	neutral := queueHeadStep{Kind: queueHeadNeutral, NeedParent: queueHeadTestParent}
	for _, tt := range []struct {
		name       string
		step       queueHeadStep
		parent     *queueHeadStep
		membership *queueHeadMembership
		known      bool
		issues     []int
		train      int
	}{
		{"direct", direct, nil, nil, true, []int{9}, 0},
		{"neutral follows first", neutral, &direct, &known, true, []int{9}, 0},
		{"neutral missing", neutral, nil, nil, false, nil, 0},
		{"neutral unresolved", neutral, &direct, &unknown, false, nil, 0},
		{"car stops before direct base", car, &direct, nil, true, []int{3}, 7},
		{"car stops before different train", car, &different, nil, true, []int{3}, 7},
		{"same train union", car, &prefix, &prefixMembership, true, []int{3, 5}, 7},
		{"same train unresolved", car, &prefix, nil, false, nil, 0},
		{"integration needs same train", queueHeadStep{Kind: queueHeadTrainIntegration, Train: 7}, &direct, &known, false, nil, 0},
		{"integration keeps prefix", queueHeadStep{Kind: queueHeadTrainIntegration, Train: 7}, &prefix, &prefixMembership, true, []int{5}, 7},
		{"neutral base is excluded", car, &neutral, &known, true, []int{3}, 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := selectQueueHeadMembership(tt.step, tt.parent, tt.membership)
			if got.Known != tt.known || !reflect.DeepEqual(got.Issues, tt.issues) || got.Train != tt.train {
				t.Fatalf("got %+v", got)
			}
		})
	}
	for _, issues := range [][]int{{3}, {1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}} {
		m := queueHeadMembership{Known: true, Train: 7, Issues: issues}
		if got := selectQueueHeadMembership(car, &prefix, &m); got.Known {
			t.Fatalf("accepted duplicate or oversized train: %+v", got)
		}
	}
}

func TestParseQueueHeadRequiresCompleteTrailerBlock(t *testing.T) {
	message := "Landing train 7: #3 (PR #99)\n\nLanding-Train: 7\nLanding-PR: 99\nthis is body prose\n"
	got := parseQueueHead(queueHeadFact{OID: queueHeadTestOID, Parents: []string{queueHeadTestParent, queueHeadTestMain}, Message: message}, "ariadne")
	if got.Kind != queueHeadMalformedTrain {
		t.Fatalf("accepted invalid trailer paragraph: %+v", got)
	}
}

func TestSelectQueueHeadNeutralCannotInheritContradictoryKnownResult(t *testing.T) {
	neutral := queueHeadStep{Kind: queueHeadNeutral}
	for _, kind := range []queueHeadKind{queueHeadUnknown, queueHeadMalformedTrain} {
		parent := queueHeadStep{Kind: kind}
		contradictory := queueHeadMembership{Known: true, Issues: []int{3}}
		if got := selectQueueHeadMembership(neutral, &parent, &contradictory); got.Known {
			t.Fatalf("inherited contradictory authority kind %v: %+v", kind, got)
		}
	}
}

func TestSelectQueueHeadTrainRejectsMalformedBase(t *testing.T) {
	car := queueHeadStep{Kind: queueHeadTrainCar, Train: 7, Issues: []int{3}}
	malformed := queueHeadStep{Kind: queueHeadMalformedTrain}
	if got := selectQueueHeadMembership(car, &malformed, nil); got.Known {
		t.Fatalf("accepted malformed base: %+v", got)
	}
}

func TestParseQueueHeadRejectsConflictingTrailerCase(t *testing.T) {
	message := "Landing train 7: #3 (PR #99)\n\nLanding-Train: 7\nlanding-train: 8\nLanding-PR: 99\n"
	got := parseQueueHead(queueHeadFact{OID: queueHeadTestOID, Parents: []string{queueHeadTestParent, queueHeadTestMain}, Message: message}, "ariadne")
	if got.Kind != queueHeadMalformedTrain {
		t.Fatalf("accepted conflicting trailer case: %+v", got)
	}
}

func TestSelectQueueHeadMembershipTrainAtCarLimit(t *testing.T) {
	issues := make([]int, maxTrainCars-1)
	for i := range issues {
		issues[i] = i + 2
	}
	step := queueHeadStep{Kind: queueHeadTrainCar, Train: 7, Issues: []int{1}}
	parent := queueHeadStep{Kind: queueHeadTrainCar, Train: 7}
	membership := queueHeadMembership{Known: true, Train: 7, Issues: issues}
	got := selectQueueHeadMembership(step, &parent, &membership)
	if !got.Known || len(got.Issues) != maxTrainCars {
		t.Fatalf("rejected complete at-limit prefix: %+v", got)
	}
	got.Issues[1] = 99
	if membership.Issues[0] != 2 {
		t.Fatal("reducer mutated parent membership")
	}
}

func TestParseQueueHeadObjectSHA256(t *testing.T) {
	oid, parent, tree := strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)
	raw := fmt.Sprintf("tree %s\nparent %s\nauthor A <a@b> 1 +0000\ncommitter A <a@b> 1 +0000\n\n#3: work\n", tree, parent)
	got, err := parseQueueHeadObject(oid, []byte(raw))
	if err != nil || got.Bytes != len(raw) || !reflect.DeepEqual(got.Parents, []string{parent}) {
		t.Fatalf("got %+v err %v", got, err)
	}
	raw = strings.Replace(raw, parent, queueHeadTestParent, 1)
	if _, err := parseQueueHeadObject(oid, []byte(raw)); err == nil {
		t.Fatal("accepted mixed object formats")
	}
}

func TestParseQueueHeadOrdinaryArchiveProvenanceIsNotTrain(t *testing.T) {
	message := "#3: issue: archive after a merge outside sdlc\n\nLanding-PR: ariadne#99\nLanding-Head: " + queueHeadTestMain + "\n"
	got := parseQueueHead(queueHeadFact{OID: queueHeadTestOID, Parents: []string{queueHeadTestParent}, Message: message}, "ariadne")
	if got.Kind != queueHeadDirect || !reflect.DeepEqual(got.Issues, []int{3}) {
		t.Fatalf("ordinary archive provenance treated as train: %+v", got)
	}
}

func TestSelectQueueHeadNeutralSharesImmutableMembership(t *testing.T) {
	parentStep := queueHeadStep{Kind: queueHeadDirect, Issues: []int{3, 5, 7}}
	parent := selectQueueHeadMembership(parentStep, nil, nil)
	neutral := queueHeadStep{Kind: queueHeadNeutral, NeedParent: queueHeadTestParent}
	// A fork of neutral descendants must retain one immutable issue set rather
	// than multiplying a large direct lead by the number of graph nodes.
	for range 100 {
		inherited := selectQueueHeadMembership(neutral, &parentStep, &parent)
		if !inherited.Known || len(inherited.Issues) != len(parent.Issues) {
			t.Fatalf("lost membership: %+v", inherited)
		}
		if &inherited.Issues[0] != &parent.Issues[0] {
			t.Fatal("neutral inheritance copied the immutable issue set")
		}
	}
}
