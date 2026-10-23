package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func typedAttemptJSON(t *testing.T) map[string]any {
	t.Helper()
	raw, err := json.Marshal(storedAttemptFixture("a1"))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["version"] = 4
	fields["runner_contract_sha256"] = strings.Repeat("c", 64)
	fields["failures"] = []string{"arbitrary legacy Go failure", "script:scripts/check.sh", "lua:tests/spec.lua#case / punctuation: # exact"}
	fields["known_flakes"] = []any{map[string]any{"test": "github.com/x/pkg.TestA", "issue": "000001"}, map[string]any{"test": "script:scripts/check.sh", "issue": "000002"}, map[string]any{"test": "lua:tests/spec.lua#case / punctuation: # exact", "issue": "000003"}}
	return fields
}
func encodeTypedAttempt(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTypedAttemptV4MixedIdentities(t *testing.T) {
	for _, check := range []string{"suite", "affected", "suite-rerun"} {
		t.Run(check, func(t *testing.T) {
			fields := typedAttemptJSON(t)
			fields["check"] = check
			if check == "suite-rerun" {
				fields["rerun_of"] = "b2"
			}
			raw := encodeTypedAttempt(t, fields)
			got, err := decodeSuiteAttempt(raw)
			if err != nil {
				t.Fatalf("valid typed record: %v", err)
			}
			if len(got.Failures) != 3 || len(got.KnownFlakes) != 3 || got.Failures[2] != fields["failures"].([]string)[2] {
				t.Fatalf("changed typed facts: %+v", got)
			}
			roundtrip, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(roundtrip, []byte(`"runner_contract_sha256":"`+strings.Repeat("c", 64)+`"`)) {
				t.Fatalf("contract hash lost: %s", roundtrip)
			}
		})
	}
}

func TestTypedAttemptV4RejectsMalformed(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"missing contract":  func(m map[string]any) { delete(m, "runner_contract_sha256") },
		"null contract":     func(m map[string]any) { m["runner_contract_sha256"] = nil },
		"short contract":    func(m map[string]any) { m["runner_contract_sha256"] = "ab" },
		"nonhex contract":   func(m map[string]any) { m["runner_contract_sha256"] = strings.Repeat("z", 64) },
		"number contract":   func(m map[string]any) { m["runner_contract_sha256"] = 1 },
		"non test check":    func(m map[string]any) { m["check"] = "lint" },
		"reclassification":  func(m map[string]any) { m["reclassification_sha256"] = strings.Repeat("d", 64) },
		"could not run":     func(m map[string]any) { m["outcome"] = "could-not-run" },
		"bare Lua":          func(m map[string]any) { m["failures"] = []string{"lua:tests/spec.lua"} },
		"invalid script":    func(m map[string]any) { m["failures"] = []string{"script:../check.sh"} },
		"oversize Go":       func(m map[string]any) { m["failures"] = []string{strings.Repeat("a", 513)} },
		"oversize Lua case": func(m map[string]any) { m["failures"] = []string{"lua:spec.lua#" + strings.Repeat("a", 2049)} },
		"oversize path":     func(m map[string]any) { m["failures"] = []string{"lua:" + strings.Repeat("a", 509) + ".lua#case"} },
		"null failure":      func(m map[string]any) { m["failures"] = []any{nil} },
		"bad issue": func(m map[string]any) {
			m["known_flakes"] = []any{map[string]any{"test": "script:check.sh", "issue": "1"}}
		},
		"null known name": func(m map[string]any) { m["known_flakes"] = []any{map[string]any{"test": nil, "issue": "000001"}} },
		"too many known":  func(m map[string]any) { m["known_flakes"] = make([]any, 10001) },
	} {
		t.Run(name, func(t *testing.T) {
			fields := typedAttemptJSON(t)
			mutate(fields)
			if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, fields)); err == nil {
				t.Fatal("accepted invalid typed attempt")
			}
		})
	}
	raw := string(encodeTypedAttempt(t, typedAttemptJSON(t)))
	for name, bad := range map[string]string{
		"alias":     strings.Replace(raw, `"runner_contract_sha256":`, `"Runner_Contract_SHA256":`, 1),
		"duplicate": strings.Replace(raw, `"version":4`, `"version":4,"runner_contract_sha256":"`+strings.Repeat("c", 64)+`"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeSuiteAttempt([]byte(bad)); err == nil {
				t.Fatal("accepted noncanonical typed field")
			}
		})
	}
}

func TestTypedAttemptVersionIsolationAndBounds(t *testing.T) {
	// A well-formed v3 audit envelope must retain its exact old shape.
	v3 := typedAttemptJSON(t)
	v3["version"], v3["outcome"] = 3, "could-not-run"
	v3["reclassification_sha256"] = strings.Repeat("d", 64)
	delete(v3, "runner_contract_sha256")
	delete(v3, "failures")
	delete(v3, "known_flakes")
	if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, v3)); err != nil {
		t.Fatalf("valid legacy audit: %v", err)
	}
	v3["runner_contract_sha256"] = strings.Repeat("c", 64)
	if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, v3)); err == nil {
		t.Fatal("v3 audit accepted a typed contract")
	}
	for _, version := range []int{1, 2, 3} {
		for _, field := range []string{"runner_contract_sha256", "failures", "known_flakes"} {
			t.Run(string(rune('0'+version))+"/"+field, func(t *testing.T) {
				m := typedAttemptJSON(t)
				m["version"] = version
				delete(m, "known_flakes")
				delete(m, "runner_contract_sha256")
				m["failures"] = []string{"legacy emitted Go failure"}
				switch field {
				case "runner_contract_sha256":
					m[field] = strings.Repeat("c", 64)
				case "failures":
					m[field] = []string{"script:check.sh"}
				case "known_flakes":
					m[field] = []any{map[string]any{"test": "lua:spec.lua#case", "issue": "000001"}}
				}
				if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, m)); err == nil {
					t.Fatal("accepted typed semantics in legacy version")
				}
			})
		}
	}
	m := typedAttemptJSON(t)
	m["failures"] = []string{"lua:" + strings.Repeat("a", 508) + ".lua#" + strings.Repeat("x", 2048)}
	if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, m)); err != nil {
		t.Fatalf("maximum canonical Lua identity (2565 bytes): %v", err)
	}
	for _, version := range []int{1, 2} {
		m := typedAttemptJSON(t)
		m["version"] = version
		delete(m, "runner_contract_sha256")
		delete(m, "known_flakes")
		m["failures"] = []string{strings.Repeat("a", 512)}
		if _, err := decodeSuiteAttempt(encodeTypedAttempt(t, m)); err != nil {
			t.Fatalf("legacy arbitrary name: %v", err)
		}
	}
	if suiteAttemptVersion != 2 || suiteAttemptReadVersion != 4 {
		t.Fatalf("writer/read versions = %d/%d", suiteAttemptVersion, suiteAttemptReadVersion)
	}
}

func TestTypedAttemptV4RawSnapshotAndRetention(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "attempts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := typedAttemptJSON(t)
	m["future"] = json.RawMessage(`{"number":1e400,"opaque":[true,null]}`)
	m["known_flakes"].([]any)[1].(map[string]any)["future"] = json.RawMessage(`{"opaque":true}`)
	raw := append(encodeTypedAttempt(t, m), '\n')
	path := filepath.Join(dir, "a1.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := readSuiteEvidence(root, 10)
	if err != nil || len(snapshot.Attempts) != 1 || len(snapshot.Warnings) != 0 {
		t.Fatalf("typed snapshot: attempts=%d unsupported=%d warnings=%v error=%v", len(snapshot.Attempts), len(snapshot.Unsupported), snapshot.Warnings, err)
	}
	if snapshot.recordDigests["a1"] != resumeRecordDigest(raw) {
		t.Fatal("snapshot did not hash exact original bytes")
	}
	if err := pruneSuiteAttempts(root, storedAttemptFixture("a1").FinishedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, raw) {
		t.Fatalf("retention rewrote raw additive fields: %s %v", after, err)
	}
}

// A typed-capable attempt never loses its runner binding through the legacy
// startup correction writer; its ordinary producer already retains startup
// environment failures outside attempts.
func TestTypedAttemptCannotDowngradeThroughStartupAudit(t *testing.T) {
	a := storedAttemptFixture("a1")
	a.Version = 4
	a.RunnerContractSHA256 = strings.Repeat("a", 64)
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = encodeReclassifiedAttempt(raw, strings.Repeat("b", 64)); err == nil {
		t.Fatal("typed contract downgraded into legacy audit disposition")
	}
	if _, err = planSuiteReclassification(a, a.Repo, "ariadne:0", "private startup evidence", []byte("go: creating work dir: no space left on device\n")); err == nil {
		t.Fatal("typed attempt accepted by legacy startup correction")
	}
}

func TestTypedAttemptRejectedZeroExitKeepsImmutableLog(t *testing.T) {
	root := t.TempDir()
	publish := func(outcome string, log []byte) SuiteAttempt {
		t.Helper()
		w, err := beginSuiteAttempt(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(w.LogPath, log, 0600); err != nil {
			t.Fatal(err)
		}
		a := storedAttemptFixture(w.ID)
		a.Version = 4
		a.RunnerContractSHA256 = strings.Repeat("f", 64)
		a.Exit = 0
		a.Failures = nil
		a.Outcome = outcome
		a.LogSHA256 = resumeRecordDigest(log)
		path, err := publishSuiteAttempt(root, w, a)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		a, err = decodeSuiteAttempt(raw)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	firstLog := []byte("{\"sdlc_unit\":1\n")
	failed := publish("unattributable-failure", firstLog)
	publish("passed", []byte("ok example.org/pkg 0.1s\n"))
	raw, err := os.ReadFile(suiteAttemptLogFile(root, failed))
	if err != nil || !bytes.Equal(raw, firstLog) || failed.LogPath != filepath.Join("failed", failed.ID+".log") {
		t.Fatalf("rejected zero-exit proof overwritten: path=%s log=%s err=%v", failed.LogPath, raw, err)
	}
}
