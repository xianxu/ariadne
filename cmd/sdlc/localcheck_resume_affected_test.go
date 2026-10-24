package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// resumeAffectedPassed is resumeCorePassed for an affected parent: its link
// is version 5 and fences the selection the child ran (#586).
func resumeAffectedPassed(t *testing.T) ResumeProofInput {
	in := resumeCorePassed(t)
	in.Parent.Check = affectedCheckName
	in.Link.Check, in.Link.Version = affectedCheckName, 5
	in.Link.Selection = &ResumeSelection{GoPkgs: []string{"example.org/p"}, Groups: []string{"a", "b"}}
	return in
}

// #586 D4: a link's version is its check's, nothing else.
func TestResumeEnvelopeVersionsByCheck(t *testing.T) {
	for _, tc := range []struct {
		version int
		check   string
		ok      bool
	}{{3, suiteCheckName, true}, {5, affectedCheckName, true}, {3, affectedCheckName, false}, {5, suiteCheckName, false}, {4, affectedCheckName, false}, {5, "lint", false}} {
		err := resumeEnvelope(tc.version, "bb", "x/y", strings.Repeat("a", 40), tc.check, "aa")
		if (err == nil) != tc.ok {
			t.Errorf("v%d %s: %v", tc.version, tc.check, err)
		}
	}
}

// #586: an affected link's selection is required, strict, and round-trips;
// a suite link carries none.
func TestResumeSelectionStrict(t *testing.T) {
	in := resumeAffectedPassed(t)
	raw, _ := json.Marshal(in.Link)
	got, _, err := decodeResumeEvidence(raw, "bb.json")
	if err != nil || got.Selection == nil || strings.Join(got.Selection.Groups, " ") != "a b" {
		t.Fatalf("decode %+v %v", got, err)
	}
	for name, sel := range map[string]string{
		"missing":      "",
		"null":         `null`,
		"null groups":  `{"go_pkgs":["p"],"groups":null}`,
		"missing pkgs": `{"groups":["a"]}`,
		"unsorted":     `{"go_pkgs":[],"groups":["b","a"]}`,
		"duplicate":    `{"go_pkgs":[],"groups":["a","a"]}`,
		"space":        `{"go_pkgs":[],"groups":["a b"]}`,
		"control":      `{"go_pkgs":[],"groups":["a\u0007"]}`,
		"empty name":   `{"go_pkgs":[""],"groups":[]}`,
		"both empty":   `{"go_pkgs":[],"groups":[]}`,
	} {
		var m map[string]json.RawMessage
		_ = json.Unmarshal(raw, &m)
		if sel == "" {
			delete(m, "selection")
		} else {
			m["selection"] = json.RawMessage(sel)
		}
		b, _ := json.Marshal(m)
		if _, _, err := decodeResumeEvidence(b, "bb.json"); err == nil {
			t.Errorf("%s: accepted %s", name, b)
		}
	}
	suite := resumeCoreLink(resumeCoreFixture(t))
	suite.Selection = &ResumeSelection{GoPkgs: []string{}, Groups: []string{"a"}}
	if b, _ := json.Marshal(suite); func() bool { _, _, err := decodeResumeEvidence(b, "bb.json"); return err == nil }() {
		t.Error("a version 3 link carried a selection")
	}
}

// #586: an affected parent resumes through a version 5 link only, and the
// link's check must be the parent's.
func TestResolveResumeProofAffected(t *testing.T) {
	in := resumeAffectedPassed(t)
	if p := resolveResumeProof(in); p.State != ResumePassed || p.Outcome.Kind != flakes.KnownFlakePassed {
		t.Fatalf("affected: %+v", p)
	}
	eligible := resumeCoreFixture(t)
	eligible.Parent.Check = affectedCheckName
	if p := resolveResumeProof(eligible); p.State != ResumeEligible {
		t.Fatalf("affected parent without a link: %+v", p)
	}
	v3 := resumeAffectedPassed(t)
	v3.Link.Version, v3.Link.Selection = 3, nil
	if p := resolveResumeProof(v3); p.State != ResumeRefused {
		t.Fatalf("v3 link on an affected parent: %+v", p)
	}
	mismatch := resumeCorePassed(t)
	mismatch.Parent.Check = affectedCheckName
	if p := resolveResumeProof(mismatch); p.State != ResumeRefused {
		t.Fatalf("suite link on an affected parent: %+v", p)
	}
	lint := resumeCoreFixture(t)
	lint.Parent.Check = "lint"
	if p := resolveResumeProof(lint); p.State != ResumeRefused {
		t.Fatalf("lint parent: %+v", p)
	}
}

// #586 D4: why version 5. Before #586 a reader falls back from an
// unreadable resume record to the unsupported-proof envelope, which admits
// only versions above suiteAttemptReadVersion: a version 5 affected link is
// fenced proof for its own parent and head, while a version 3 one would be
// an invalid record, which marks the whole snapshot incomplete.
func TestOldReaderFencesAffectedLinkPerHead(t *testing.T) {
	in := resumeAffectedPassed(t)
	raw, _ := json.Marshal(in.Link)
	p, err := decodeUnsupportedProof(raw, "bb.json")
	if err != nil || p.Check != affectedCheckName || p.RerunOf != "aa" || p.SHA != in.Parent.SHA || p.Version != 5 {
		t.Fatalf("v5 link as unsupported proof: %+v %v", p, err)
	}
	in.Link.Version = 3
	raw, _ = json.Marshal(in.Link)
	if _, err := decodeUnsupportedProof(raw, "bb.json"); err == nil {
		t.Fatal("a v3 affected link would have been fenced, not invalid")
	}
}

// #586: a version 5 receipt binds its version 5 link.
func TestResumeReceiptAffected(t *testing.T) {
	in := resumeAffectedPassed(t)
	in.Child = nil
	l := in.Link
	in.Receipt = &ResumeReconciliation{Version: 5, Kind: "resume-reconciliation", ID: "ee", Repo: l.Repo, SHA: l.SHA, Check: affectedCheckName, RerunOf: l.RerunOf, LinkID: l.ID, LinkSHA256: in.LinkSHA256, ChildID: l.ChildID, Actor: "uid=1", Note: "stopped", Outcome: "interrupted", FinishedAt: l.CreatedAt.Add(1)}
	if p := resolveResumeProof(in); p.State != ResumeInterrupted {
		t.Fatalf("v5 receipt: %+v", p)
	}
	raw, _ := json.Marshal(in.Receipt)
	if _, r, err := decodeResumeEvidence(raw, "ee.json"); err != nil || r == nil {
		t.Fatalf("decode v5 receipt: %v", err)
	}
	in.Receipt.Version = 3
	if p := resolveResumeProof(in); p.State != ResumeRefused {
		t.Fatalf("v3 receipt on an affected link: %+v", p)
	}
}
