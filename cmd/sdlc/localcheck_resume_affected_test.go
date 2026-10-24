package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
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

// resumeAffectedFixture is a saved affected attempt that failed only on the
// registered example.org/p.TestOne and has no link yet (#586).
func resumeAffectedFixture(t *testing.T) (localRunDeps, string, string) {
	t.Helper()
	d, sha, store, link := resumeSavedFixture(t, false)
	if err := os.Remove(filepath.Join(store, "attempts", link.ID+".json")); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(store, "attempts", "aa.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parent SuiteAttempt
	if err := json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	parent.Check = affectedCheckName
	resumeSavedWriteJSON(t, store, parent.ID, parent)
	return d, sha, store
}

// affectedResumeCheck is the affected declaration whose groups command
// records the env it ran under in marker, then runs group.
func affectedResumeCheck(marker, group string) localCheck {
	return localCheck{Name: affectedCheckName, Command: []string{"sh", "-c", `printf '%s|%s' "$SDLC_AFFECTED_GO_PKGS" "$SDLC_AFFECTED_GROUPS" > ` + marker + "; " + group}, TimeoutMinutes: 30, Weight: 1}
}

func plannedSelection(sel affected.Selection, calls *int) resumeSelectionFunc {
	return func(context.Context) (affected.Selection, error) { *calls++; return sel, nil }
}

const resumeTestOnePass = `printf '%s\n' '{"Action":"pass","Package":"example.org/p","Test":"TestOne"}' '{"Action":"pass","Package":"example.org/p","Elapsed":1}'`

// #586 Done when: an affected attempt that failed only on a registered flake
// and stopped before its groups resumes: the test reruns, every selected
// group runs, and the head records an affected pass under a version 5 link.
func TestResumeAffectedRunsFlakesThenGroups(t *testing.T) {
	d, sha, store := resumeAffectedFixture(t)
	asked := scriptRerun(t, resumeTestOnePass)
	marker := filepath.Join(t.TempDir(), "groups")
	calls := 0
	sel := plannedSelection(affected.Selection{GoPkgs: []string{"example.org/p"}, Groups: []string{"a", "b"}}, &calls)
	rec, err := resumeSaved(context.Background(), d, affectedResumeCheck(marker, "exit 0"), sel, sha, "aa")
	if err != nil || !rec.passed() || rec.Name != affectedCheckName || rec.Outcome != flakes.KnownFlakePassed || rec.ResumeParent != "aa" || rec.ResumeChild == "" || rec.Scope != "1 package, groups a b" {
		t.Fatalf("resume: %+v %v", rec, err)
	}
	if got, _ := os.ReadFile(marker); string(got) != "|a b" || strings.Join(*asked, ",") != "example.org/p.TestOne" || calls != 1 {
		t.Fatalf("groups ran with %q, reran %v, planned %d times", got, *asked, calls)
	}
	if gh := d.GH.(*ciFakeGH); gh.statuses[sha][localStatusContext(affectedCheckName)] != "success" || gh.statuses[sha][localStatusContext(suiteCheckName)] != "" {
		t.Fatalf("statuses %v", gh.statuses[sha])
	}
	s, err := readSuiteEvidence(store, testStatsRecordLimit)
	if err != nil || len(s.Resumes) != 1 || s.Resumes[0].Version != 5 || s.Resumes[0].Check != affectedCheckName || strings.Join(s.Resumes[0].Selection.Groups, " ") != "a b" {
		t.Fatalf("link %+v %v", s.Resumes, err)
	}
	// A repeat reposts from the proof: nothing plans or runs again.
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	again, err := resumeSaved(context.Background(), d, affectedResumeCheck(marker, "exit 0"), sel, sha, "aa")
	if _, statErr := os.Stat(marker); err != nil || !again.passed() || calls != 1 || statErr == nil {
		t.Fatalf("repeat: %+v %v calls=%d", again, err, calls)
	}
}

// #586 Done when: an unregistered failure still refuses, before any plan,
// link, child or status.
func TestResumeAffectedUnregisteredFailureRefuses(t *testing.T) {
	d, sha, store := resumeAffectedFixture(t)
	d.FlakeRegistry = gateRegistry(t, "example.org/p.TestOther")
	calls := 0
	_, err := resumeSaved(context.Background(), d, affectedResumeCheck(filepath.Join(t.TempDir(), "m"), "exit 0"), plannedSelection(affected.Selection{Groups: []string{"a"}}, &calls), sha, "aa")
	s, _ := readSuiteEvidence(store, testStatsRecordLimit)
	if err == nil || calls != 0 || len(s.Resumes) != 0 || len(d.GH.(*ciFakeGH).statuses[sha]) != 0 {
		t.Fatalf("err %v, planned %d, links %d, statuses %v", err, calls, len(s.Resumes), d.GH.(*ciFakeGH).statuses[sha])
	}
}

// #586: the flake passes but a group fails: no pass is recorded.
func TestResumeAffectedGroupFailureRecordsNoPass(t *testing.T) {
	d, sha, _ := resumeAffectedFixture(t)
	scriptRerun(t, resumeTestOnePass)
	calls := 0
	rec, err := resumeSaved(context.Background(), d, affectedResumeCheck(filepath.Join(t.TempDir(), "m"), "exit 1"), plannedSelection(affected.Selection{Groups: []string{"a"}}, &calls), sha, "aa")
	if (err == nil && rec.passed()) || d.GH.(*ciFakeGH).statuses[sha][localStatusContext(affectedCheckName)] == "success" {
		t.Fatalf("a failing group landed: %+v %v", rec, err)
	}
}

// #586 D3: a selection that is now full, or empty, runs nothing and points
// at the suite; a suite resume of an affected parent names the right check.
func TestResumeAffectedRefusesFullOrEmptySelection(t *testing.T) {
	for name, sel := range map[string]affected.Selection{"full": {Full: true, Reason: "go.mod changed"}, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			d, sha, store := resumeAffectedFixture(t)
			calls := 0
			_, err := resumeSaved(context.Background(), d, affectedResumeCheck(filepath.Join(t.TempDir(), "m"), "exit 0"), plannedSelection(sel, &calls), sha, "aa")
			s, _ := readSuiteEvidence(store, testStatsRecordLimit)
			if err == nil || !strings.Contains(err.Error(), "run sdlc local-check run --name suite --rev "+sha) || len(s.Resumes) != 0 {
				t.Fatalf("%v links=%d", err, len(s.Resumes))
			}
		})
	}
	d, sha, _ := resumeAffectedFixture(t)
	if _, err := resumeSaved(context.Background(), d, suiteCheck("true"), nil, sha, "aa"); err == nil || !strings.Contains(err.Error(), "resume --name affected") {
		t.Fatalf("suite resume of an affected parent: %v", err)
	}
}

// #586: an aged terminal affected bundle (version 5 link) retires like a
// suite one; a link whose check is not its parent's is never a witness.
func TestResumeRetentionAffectedBundle(t *testing.T) {
	root, link, now := retentionFixtureOf(t, true, resumeAffectedPassed(t))
	if _, err := pruneResumeBundles(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); !os.IsNotExist(err) {
		t.Fatalf("aged affected bundle kept: %v", err)
	}
	mismatch := resumeAffectedPassed(t)
	mismatch.Parent.Check = suiteCheckName
	root, link, now = retentionFixtureOf(t, true, mismatch)
	if _, err := pruneResumeBundles(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); err != nil {
		t.Fatalf("a link bound to another check's parent was retired: %v", err)
	}
}
