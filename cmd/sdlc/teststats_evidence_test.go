package main

import (
	"context"
	"encoding/json"
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
