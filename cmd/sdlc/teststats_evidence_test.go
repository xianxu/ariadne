package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the real saved-proof and post guards, rather than a parser that
// consumers could bypass. All records and Git state are private to this test.
func TestScopedSavedProofConsumers(t *testing.T) {
	for _, tc := range []struct {
		name, field, value string
		blocked            bool
	}{
		{"another repository", "repo", "another/repository", false},
		{"another head", "sha", strings.Repeat("f", 40), false},
		{"another named check", "check", "vm", false},
		{"same suite subject", "", "", true},
		{"linked parent", "rerun_of", "parent", true},
		{"cross-check linked parent", "rerun_of", "parent", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "skipped")
			logDir, err := localCheckLogDir(d)
			if err != nil {
				t.Fatal(err)
			}
			store := suiteStoreRoot(logDir)
			w, err := beginSuiteAttempt(store)
			if err != nil {
				t.Fatal(err)
			}
			log := []byte("ok  example.test/pkg 1.0s\n")
			if err := os.WriteFile(w.LogPath, log, 0o600); err != nil {
				t.Fatal(err)
			}
			a := storedAttemptFixture(w.ID)
			a.Repo, a.SHA, a.Exit = d.Repo, sha, 0
			a.Failures = nil
			wall := 1.0
			a.WallSeconds, a.LogSHA256 = &wall, hashBytes(log)
			if _, err := publishSuiteAttempt(store, w, a); err != nil {
				t.Fatal(err)
			}
			future := map[string]any{"version": 3, "id": "d3", "repo": d.Repo, "sha": sha, "check": "suite", "future": map[string]any{"opaque": true}}
			if tc.field != "" {
				v := tc.value
				if v == "parent" {
					v = w.ID
				}
				future[tc.field] = v
			}
			if tc.name == "cross-check linked parent" {
				future["check"] = "vm"
			}
			raw, err := json.Marshal(future)
			if err != nil {
				t.Fatal(err)
			}
			p := filepath.Join(store, "attempts", "d3.json")
			if err := os.WriteFile(p, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			// The existing supported-attempt decoder must never accept opaque
			// future contents as measurements or passing facts.
			if _, err := decodeSuiteAttempt(raw); err == nil {
				t.Fatal("opaque future evidence was accepted as a SuiteAttempt")
			}
			rec, newest, loadErr := loadSavedAttempt(store, d.Repo, sha, "suite", "", "host")
			guardErr := guardSuitePass(context.Background(), d, "suite", sha)
			if tc.blocked {
				for _, id := range []string{w.ID, "d3"} {
					_, _, err := loadSavedAttempt(store, d.Repo, sha, "suite", id, "host")
					if err == nil || !strings.Contains(err.Error(), "needs a newer sdlc") {
						t.Errorf("explicit %s selection: %v", id, err)
					}
				}
				_, reused, err := republishSavedPass(context.Background(), d, localCheck{Name: "suite"}, sha)
				if !reused || err == nil || !strings.Contains(err.Error(), "needs a newer sdlc") {
					t.Errorf("saved-pass fallback: reused=%v err=%v", reused, err)
				}
				err = postLocalCheckRecord(context.Background(), d, localRunRecord{Name: "suite", SHA: sha, Exit: 0})
				if err == nil || gh.statuses[sha][localStatusContext("suite")] == "success" {
					t.Errorf("unsupported parent authorized publication: %v", err)
				}
				for _, got := range []error{loadErr, guardErr} {
					if got == nil || !strings.Contains(got.Error(), "needs a newer sdlc") {
						t.Errorf("related unsupported proof: %v; want scoped upgrade refusal", got)
					}
				}
			} else {
				if loadErr != nil || !newest || !rec.passed() {
					t.Errorf("unrelated future record poisoned saved proof: newest=%v rec=%+v err=%v", newest, rec, loadErr)
				}
				if guardErr != nil {
					t.Errorf("unrelated future record poisoned success guard: %v", guardErr)
				}
				if loadErr == nil {
					if err := postLocalCheckRecord(context.Background(), d, rec); err != nil {
						t.Errorf("unrelated proof did not post: %v", err)
					}
					if gh.statuses[sha][localStatusContext("suite")] != "success" {
						t.Error("unrelated valid suite proof was not recorded")
					}
				}
			}
			if got, err := os.ReadFile(p); err != nil || string(got) != string(raw) {
				t.Fatalf("future bytes changed: %v", err)
			}
		})
	}
}

// Legacy projections are test-only: production callers must handle the scoped
// snapshot, so adding a new caller cannot silently drop unsupported authority.
func readSuiteAttempts(root string, limit int) ([]SuiteAttempt, []string, error) {
	s, err := readSuiteEvidence(root, limit)
	return s.Attempts, s.Warnings, err
}
func readSuiteAttemptsBudget(root string, limit int, bytes int64) ([]SuiteAttempt, []string, error) {
	s, err := readSuiteEvidenceBudget(root, limit, bytes)
	return s.Attempts, s.Warnings, err
}

func futureProofFixture(id string) []byte {
	return []byte(`{"version":3,"id":"` + id + `","repo":"x/ariadne","sha":"` + strings.Repeat("a", 40) + `","check":"suite","opaque":{"number":1e400,"array":[true,null]}}`)
}

func TestDecodeUnsupportedProof(t *testing.T) {
	valid := futureProofFixture("d3")
	p, err := decodeUnsupportedProof(valid, "d3.json")
	if err != nil || p.ID != "d3" || string(p.Raw) != string(valid) {
		t.Fatalf("opaque identity: %+v %v", p, err)
	}
	valid[0] = ' '
	if p.Raw[0] != '{' {
		t.Fatal("raw evidence aliases caller memory")
	}
	valid = futureProofFixture("d3")
	for name, raw := range map[string]string{
		"supported":        strings.Replace(string(valid), `"version":3`, `"version":2`, 1),
		"negative":         strings.Replace(string(valid), `"version":3`, `"version":-3`, 1),
		"version type":     strings.Replace(string(valid), `"version":3`, `"version":"3"`, 1),
		"missing repo":     strings.Replace(string(valid), `"repo":"x/ariadne",`, "", 1),
		"null repo":        strings.Replace(string(valid), `"repo":"x/ariadne"`, `"repo":null`, 1),
		"alias":            strings.Replace(string(valid), `"sha":`, `"SHA":`, 1),
		"duplicate nested": strings.Replace(string(valid), `"number":1e400`, `"number":1,"number":2`, 1),
		"trailing":         string(valid) + `{}`,
		"invalid repo":     strings.Replace(string(valid), `x/ariadne`, `x/ariadne/extra`, 1),
		"short sha":        strings.Replace(string(valid), strings.Repeat("a", 40), "aaaa", 1),
		"bad check":        strings.Replace(string(valid), `"check":"suite"`, `"check":"../suite"`, 1),
		"self parent":      strings.Replace(string(valid), `"version":3`, `"version":3,"rerun_of":"d3"`, 1),
		"parent type":      strings.Replace(string(valid), `"version":3`, `"version":3,"rerun_of":[]`, 1),
		"parent null":      strings.Replace(string(valid), `"version":3`, `"version":3,"rerun_of":null`, 1),
		"oversize":         string(valid) + strings.Repeat(" ", suiteAttemptMaxJSON),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeUnsupportedProof([]byte(raw), "d3.json"); err == nil {
				t.Fatal("accepted ambiguous identity")
			}
		})
	}
	if _, err := decodeUnsupportedProof(valid, "d4.json"); err == nil {
		t.Fatal("accepted filename mismatch")
	}
}

func TestUnsupportedProofBlocks(t *testing.T) {
	base := UnsupportedProof{ID: "d3", Repo: "x/ariadne", SHA: strings.Repeat("a", 40), Check: "suite"}
	subject := proofSubject{Repo: base.Repo, SHA: base.SHA, Check: "suite", ID: "a1"}
	for _, tc := range []struct {
		name   string
		change func(*UnsupportedProof, *proofSubject)
		want   bool
	}{
		{"same subject", func(*UnsupportedProof, *proofSubject) {}, true},
		{"other repo", func(p *UnsupportedProof, _ *proofSubject) { p.Repo = "x/other" }, false},
		{"other head", func(p *UnsupportedProof, _ *proofSubject) { p.SHA = strings.Repeat("b", 40) }, false},
		{"other named check", func(p *UnsupportedProof, _ *proofSubject) { p.Check = "vm" }, false},
		{"affected shares suite authority", func(p *UnsupportedProof, _ *proofSubject) { p.Check = "affected" }, true},
		{"rerun shares suite authority", func(p *UnsupportedProof, _ *proofSubject) { p.Check = "suite-rerun" }, true},
		{"named check is independent", func(_ *UnsupportedProof, s *proofSubject) { s.Check = "vm" }, false},
		{"linked parent", func(p *UnsupportedProof, _ *proofSubject) { p.RerunOf = "a1"; p.Check = "vm" }, true},
		{"different parent", func(p *UnsupportedProof, _ *proofSubject) { p.RerunOf = "b2" }, false},
		{"implicit selection cannot skip link", func(p *UnsupportedProof, s *proofSubject) { p.RerunOf = "b2"; s.ID = "" }, true},
		{"explicit unsupported ID", func(p *UnsupportedProof, s *proofSubject) { s.ID = p.ID; s.SHA = strings.Repeat("b", 40) }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, s := base, subject
			tc.change(&p, &s)
			if got := unsupportedProofBlocks(p, s); got != tc.want {
				t.Fatalf("relation=%v want %v", got, tc.want)
			}
			var snapshot SuiteEvidenceSnapshot
			snapshot.addUnsupported(p)
			if got := snapshot.proofError(s) != nil; got != tc.want {
				t.Fatalf("indexed authority=%v want %v", got, tc.want)
			}
		})
	}
}

func TestReadSuiteEvidenceBudget(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "attempts")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name string, raw []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a := storedAttemptFixture("a1")
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	write("a1.json", raw)
	write("d3.json", futureProofFixture("d3"))
	snapshot, err := readSuiteEvidence(root, 10)
	if err != nil || len(snapshot.Attempts) != 1 || len(snapshot.Unsupported) != 1 || len(snapshot.Warnings) != 0 {
		t.Fatalf("mixed snapshot: %+v %v", snapshot, err)
	}
	if len(snapshot.flakeAttempts()) != 0 || len(snapshot.coverageNotes("x/ariadne")) != 1 || !snapshot.knownSuiteHeads()[a.SHA] {
		t.Fatal("unavailable proof became flake or legacy evidence")
	}
	other := storedAttemptFixture("b2")
	other.SHA = strings.Repeat("b", 40)
	snapshot.Attempts = append(snapshot.Attempts, other)
	if got := snapshot.flakeAttempts(); len(got) != 1 || got[0].ID != "b2" {
		t.Fatalf("unrelated proof lost: %+v", got)
	}
	for name, limit := range map[string]int64{"byte budget": 1, "record budget": 1 << 20} {
		t.Run(name, func(t *testing.T) {
			count := 10
			if name == "record budget" {
				count = 1
			}
			got, err := readSuiteEvidenceBudget(root, count, limit)
			if err != nil || len(got.Warnings) == 0 {
				t.Fatalf("budget invented completeness: %+v %v", got, err)
			}
		})
	}
	write("e4.json", []byte(`{"version":3,"id":"e4"}`))
	got, err := readSuiteEvidence(root, 10)
	if err != nil || len(got.Warnings) != 1 || len(got.Attempts) != 1 || len(got.Unsupported) != 1 {
		t.Fatalf("unscopable read: %+v %v", got, err)
	}
}

func TestScopedEvidenceProjectionGrowth(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			var s SuiteEvidenceSnapshot
			for i := 0; i < n; i++ {
				id := fmt.Sprintf("%04x", i)
				p := UnsupportedProof{ID: id, Repo: "x/repo", SHA: strings.Repeat("a", 40), Check: "suite", RerunOf: fmt.Sprintf("%04x", i+n)}
				s.addUnsupported(p)
				s.Attempts = append(s.Attempts, SuiteAttempt{ID: p.RerunOf, Repo: p.Repo, SHA: p.SHA, Check: p.Check})
			}
			// The index has one entry per parent. Each projection does at most one
			// ID and one parent lookup, independent of the number of sibling records.
			if len(s.byID) != n || len(s.byParent) != n || len(s.bySubject) != 1 {
				t.Fatalf("index cardinality: %+v", s)
			}
			if got := s.flakeAttempts(); len(got) != 0 {
				t.Fatalf("linked proof leaked: %d", len(got))
			}
		})
	}
}

func TestReadSuiteEvidenceSingleScan(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "attempts")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < n; i++ {
				id := fmt.Sprintf("%04x", i)
				if err := os.WriteFile(filepath.Join(dir, id+".json"), futureProofFixture(id), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			opens := map[string]int{}
			snapshot, err := readSuiteEvidenceWithOpen(root, n+1, suiteAttemptReaderBudget, func(path string) (*os.File, error) { opens[path]++; return os.Open(path) })
			if err != nil || len(snapshot.Unsupported) != n || len(snapshot.Warnings) != 0 || len(opens) != n+1 {
				t.Fatalf("bounded single scan: opens=%v snapshot=%+v error=%v", opens, snapshot, err)
			}
			for path, count := range opens {
				if count != 1 {
					t.Fatalf("rescanned %s %d times", path, count)
				}
			}
			for i := 0; i < 2*n; i++ {
				_ = snapshot.proofError(proofSubject{Repo: "x/ariadne", SHA: strings.Repeat("a", 40), Check: "suite"})
			}
			if len(opens) != n+1 {
				t.Fatal("proof projection performed IO")
			}
		})
	}
}

func TestScopedLinkedProofUsesParentCheck(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(logDir)
	var parent string
	for _, check := range []string{"suite", "vm"} {
		w, err := beginSuiteAttempt(store)
		if err != nil {
			t.Fatal(err)
		}
		log := []byte("ok example.test/pkg 1s\n")
		if err := os.WriteFile(w.LogPath, log, 0o600); err != nil {
			t.Fatal(err)
		}
		a := storedAttemptFixture(w.ID)
		a.Repo, a.SHA, a.Check, a.Exit, a.Failures = d.Repo, sha, check, 0, nil
		wall := 1.0
		a.WallSeconds, a.LogSHA256 = &wall, hashBytes(log)
		if _, err := publishSuiteAttempt(store, w, a); err != nil {
			t.Fatal(err)
		}
		if check == "vm" {
			parent = w.ID
		}
	}
	raw, err := json.Marshal(map[string]any{"version": 3, "id": "d3", "repo": d.Repo, "sha": sha, "check": "suite", "rerun_of": parent})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "attempts", "d3.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	rec, newest, err := loadSavedAttempt(store, d.Repo, sha, "suite", "", "host")
	if err != nil || !newest || !rec.passed() {
		t.Fatalf("link to VM blocked unrelated suite: newest=%v err=%v", newest, err)
	}
	if err := postLocalCheckRecord(context.Background(), d, rec); err != nil || gh.statuses[sha][localStatusContext("suite")] != "success" {
		t.Fatalf("unrelated suite publication: %v", err)
	}
	_, _, err = loadSavedAttempt(store, d.Repo, sha, "vm", "", "host")
	if err == nil || !strings.Contains(err.Error(), "needs a newer sdlc") {
		t.Fatalf("VM parent authority lost: %v", err)
	}
}
