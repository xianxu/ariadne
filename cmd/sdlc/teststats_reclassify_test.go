package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func reclassificationFixture(t *testing.T) (string, SuiteAttempt, []byte) {
	t.Helper()
	root := t.TempDir()
	w, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	log := []byte(recordedCacheFailure(t))
	if err := os.WriteFile(w.LogPath, log, 0600); err != nil {
		t.Fatal(err)
	}
	a := storedAttemptFixture(w.ID)
	a.Slot, a.Failures, a.LogSHA256 = "ariadne:5", nil, hashBytes(log)
	if _, err := publishSuiteAttempt(root, w, a); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "attempts", a.ID+".json")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["future"] = json.RawMessage(`{"nested":[null,true,{"number":1e400}]}`)
	raw, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	a, err = decodeSuiteAttempt(raw)
	if err != nil {
		t.Fatal(err)
	}
	return root, a, raw
}

func TestSuiteReclassificationPublication(t *testing.T) {
	root, a, original := reclassificationFixture(t)
	at := time.Date(2026, 10, 21, 22, 0, 0, 0, time.UTC)
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repaired cache", at, nil); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "attempts", a.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["version"]) != "3" || string(fields["outcome"]) != `"could-not-run"` || !bytes.Contains(fields["future"], []byte("1e400")) {
		t.Fatalf("incorrect disposition or lost opaque fields: %s", raw)
	}
	var old map[string]json.RawMessage
	if err := json.Unmarshal(original, &old); err != nil {
		t.Fatal(err)
	}
	for k, v := range old {
		if k != "version" && k != "outcome" && !bytes.Equal(v, fields[k]) {
			t.Errorf("changed original fact %s: %s -> %s", k, v, fields[k])
		}
	}
	snapshot, err := readSuiteEvidence(root, 100)
	if err != nil || len(snapshot.Warnings) != 0 || len(snapshot.Attempts) != 1 || len(snapshot.Unsupported) != 0 {
		t.Fatalf("unreadable correction: %+v %v", snapshot, err)
	}
	if isFlakeEvidence(snapshot.Attempts[0]) {
		t.Fatal("could-not-run became flake evidence")
	}
	if _, _, err := loadSavedAttempt(root, a.Repo, a.SHA, a.Check, "", ""); err == nil {
		t.Fatal("could-not-run became reusable result")
	}
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repeat", at.Add(time.Hour), nil); err != nil {
		t.Fatalf("idempotent retry: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "attempts", a.ID+".json"))
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatalf("retry rewrote correction: %s %v", after, err)
	}
	// The opt-in historical-reader control consumes these actual encoder bytes,
	// without linking the new decoder into the pinned source checkout.
	if export := os.Getenv("SDLC_RECLASSIFY_EXPORT"); export != "" {
		if !filepath.IsAbs(export) {
			t.Fatal("SDLC_RECLASSIFY_EXPORT must be absolute")
		}
		if err := os.WriteFile(export, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSuiteReclassificationRequiresGenuineRerun(t *testing.T) {
	root, a, _ := reclassificationFixture(t)
	at := a.FinishedAt.Add(2 * time.Hour)
	pass := func(start time.Time) string {
		t.Helper()
		w, err := beginSuiteAttempt(root)
		if err != nil {
			t.Fatal(err)
		}
		log := []byte("ok  example/pkg 0.01s\n")
		if err := os.WriteFile(w.LogPath, log, 0600); err != nil {
			t.Fatal(err)
		}
		b := a
		b.ID, b.Exit, b.StartedAt, b.FinishedAt = w.ID, 0, start, start.Add(time.Second)
		b.WallSeconds = new(float64)
		*b.WallSeconds = 1
		b.LogSHA256 = hashBytes(log)
		if _, err := publishSuiteAttempt(root, w, b); err != nil {
			t.Fatal(err)
		}
		return w.ID
	}
	old := pass(at.Add(-time.Hour))
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repaired cache", at, nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"", old} {
		if rec, newest, err := loadSavedAttempt(root, a.Repo, a.SHA, a.Check, id, ""); err == nil {
			t.Fatalf("old pass promoted by correction: id=%q rec=%+v newest=%v", id, rec, newest)
		}
	}
	fresh := pass(at.Add(time.Minute))
	if rec, newest, err := loadSavedAttempt(root, a.Repo, a.SHA, a.Check, fresh, ""); err != nil || !newest || !rec.passed() {
		t.Fatalf("genuine rerun refused: %+v %v %v", rec, newest, err)
	}
}

func TestSuiteReclassificationInterruptedPublication(t *testing.T) {
	for _, stop := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(stop), func(t *testing.T) {
			root, a, original := reclassificationFixture(t)
			step := 0
			private := t.TempDir()
			publish := func(path string, b []byte, exclusive bool) error {
				snapshot, err := readSuiteEvidence(root, 100)
				if err != nil || len(snapshot.Attempts) != 1 || snapshot.Attempts[0].ReclassificationSHA256 != "" {
					t.Fatalf("unpublished authority vanished: %+v %v", snapshot, err)
				}
				if step == stop {
					return errors.New("injected interruption")
				}
				if err := writeSuiteReclassificationFile(private, path, b, exclusive); err != nil {
					return err
				}
				step++
				if step == 2 && stop == 2 {
					return errors.New("reply lost after replacement")
				}
				return nil
			}
			at := a.FinishedAt.Add(time.Hour)
			if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", at, publish); err == nil {
				t.Fatal("injected failure disappeared")
			}
			snapshot, err := readSuiteEvidence(root, 100)
			if err != nil || len(snapshot.Attempts) != 1 || len(snapshot.Warnings) != 0 {
				t.Fatalf("unknown publication state became absence: %+v %v", snapshot, err)
			}
			if stop < 2 {
				got, err := os.ReadFile(filepath.Join(root, "attempts", a.ID+".json"))
				if err != nil || !bytes.Equal(got, original) {
					t.Fatalf("failure authority changed before commit: %s %v", got, err)
				}
			}
			if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "retry", at.Add(time.Minute), nil); err != nil {
				t.Fatalf("retry did not reconcile exact original: %v", err)
			}
		})
	}
}

func TestSuiteReclassificationProofMutations(t *testing.T) {
	for _, kind := range []string{"audit missing", "audit truncated", "log changed", "log missing", "audit actor", "audit reason", "audit time", "audit original", "audit classifier", "record head", "record exit"} {
		t.Run(kind, func(t *testing.T) {
			root, a, _ := reclassificationFixture(t)
			if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", a.FinishedAt.Add(time.Hour), nil); err != nil {
				t.Fatal(err)
			}
			record := filepath.Join(root, "attempts", a.ID+".json")
			audit := filepath.Join(root, "reclassified", a.ID+".json")
			log := filepath.Join(root, "failed", a.ID+".log")
			mutate := func(path, key, value string) {
				b, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(b, &fields); err != nil {
					t.Fatal(err)
				}
				fields[key] = json.RawMessage(value)
				b, err = json.Marshal(fields)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch kind {
			case "audit missing":
				if err := os.Remove(audit); err != nil {
					t.Fatal(err)
				}
			case "audit truncated":
				if err := os.WriteFile(audit, []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			case "log changed":
				if err := os.WriteFile(log, []byte("=== RUN TestActual\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "log missing":
				if err := os.Remove(log); err != nil {
					t.Fatal(err)
				}
			case "record head":
				mutate(record, "sha", `"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`)
			case "record exit":
				mutate(record, "exit", "0")
			default:
				key, value := "actor", `"ariadne:4"`
				switch kind {
				case "audit reason":
					key, value = "reason", `" "`
				case "audit time":
					key, value = "at", `"0001-01-01T00:00:00Z"`
				case "audit original":
					key, value = "original", `"e30="`
				case "audit classifier":
					key, value = "classifier", "2"
				}
				mutate(audit, key, value)
				b, err := os.ReadFile(audit)
				if err != nil {
					t.Fatal(err)
				}
				digest, _ := json.Marshal(hashBytes(b))
				mutate(record, "reclassification_sha256", string(digest))
			}
			snapshot, err := readSuiteEvidence(root, 100)
			if err != nil {
				t.Fatal(err)
			}
			if len(snapshot.Attempts) != 0 || len(snapshot.Warnings) == 0 && len(snapshot.Unsupported) == 0 {
				t.Fatalf("unverified proof became absence/authority: %+v", snapshot)
			}
			if kind == "record head" && len(snapshot.Warnings) == 0 {
				t.Fatal("moved subject concealed the original head's unavailable authority")
			}
			if _, _, err := loadSavedAttempt(root, a.Repo, a.SHA, a.Check, "", ""); err == nil {
				t.Fatal("damaged proof authorized saved result")
			}
		})
	}
}

func TestSuiteReclassificationBoundedReadCounts(t *testing.T) {
	for _, n := range []int{8, 16} {
		root := t.TempDir()
		for range n {
			source, a, _ := reclassificationFixture(t)
			if err := publishSuiteReclassification(source, a.ID, a.Repo, "ariadne:5", "repair", a.FinishedAt.Add(time.Hour), nil); err != nil {
				t.Fatal(err)
			}
			for _, dir := range []string{"attempts", "reclassified", "failed"} {
				ext := ".json"
				if dir == "failed" {
					ext = ".log"
				}
				b, err := os.ReadFile(filepath.Join(source, dir, a.ID+ext))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, dir, a.ID+ext), b, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		opens := 0
		snapshot, err := readSuiteEvidenceWithOpen(root, n+1, suiteAttemptReaderBudget, func(path string) (*os.File, error) { opens++; return os.Open(path) })
		if err != nil || len(snapshot.Attempts) != n || opens != 3*n+1 {
			t.Fatalf("N=%d opens=%d attempts=%d err=%v", n, opens, len(snapshot.Attempts), err)
		}
	}
}

func TestSuiteReclassificationAuditRetryAndRetention(t *testing.T) {
	root, a, _ := reclassificationFixture(t)
	at := a.FinishedAt.Add(time.Hour)
	private := t.TempDir()
	err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", at, func(path string, b []byte, exclusive bool) error {
		if !exclusive {
			return errors.New("interrupted before record")
		}
		return writeSuiteReclassificationFile(private, path, b, true)
	})
	if err == nil {
		t.Fatal("injection did not interrupt")
	}
	auditPath := filepath.Join(root, "reclassified", a.ID+".json")
	b, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatal(err)
	}
	fields["future"] = json.RawMessage(`{"opaque":1e400}`)
	b, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auditPath, b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "retry", at.Add(time.Minute), nil); err != nil {
		t.Fatal(err)
	}
	retained, err := os.ReadFile(auditPath)
	if err != nil || !bytes.Equal(retained, b) {
		t.Fatalf("retry rewrote unknown audit facts: %s %v", retained, err)
	}
	now := time.Now()
	old := now.Add(-suiteAttemptKeep - time.Hour)
	if err := os.Chtimes(filepath.Join(root, "attempts", a.ID+".json"), old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{auditPath, filepath.Join(root, "attempts", a.ID+".json"), filepath.Join(root, "failed", a.ID+".log")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expired recovery evidence retained: %s %v", path, err)
		}
	}
}

func TestSuiteReclassificationRetryBindsOriginalRawDigest(t *testing.T) {
	root, a, raw := reclassificationFixture(t)
	private := t.TempDir()
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", a.FinishedAt.Add(time.Hour), func(path string, b []byte, exclusive bool) error {
		if !exclusive {
			return errors.New("interrupted before record")
		}
		return writeSuiteReclassificationFile(private, path, b, true)
	}); err == nil {
		t.Fatal("injection did not interrupt")
	}
	// Even a whitespace-only rewrite is a different original byte generation.
	path := filepath.Join(root, "attempts", a.ID+".json")
	if err := os.WriteFile(path, append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "retry", a.FinishedAt.Add(2*time.Hour), nil); err == nil {
		t.Fatal("interrupted recovery accepted a different raw original digest")
	}
	if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, append(raw, '\n')) {
		t.Fatalf("refused recovery changed original: %s %v", got, err)
	}
}

func TestSuiteReclassificationPruneInterruptionReclaimsAudit(t *testing.T) {
	root, a, _ := reclassificationFixture(t)
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", a.FinishedAt.Add(time.Hour), nil); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	old := now.Add(-suiteAttemptKeep - time.Hour)
	for _, dir := range []string{"attempts", "reclassified"} {
		if err := os.Chtimes(filepath.Join(root, dir, a.ID+".json"), old, old); err != nil {
			t.Fatal(err)
		}
	}
	attempt := filepath.Join(root, "attempts", a.ID+".json")
	audit := filepath.Join(root, "reclassified", a.ID+".json")
	if err := pruneSuiteAttemptsWithRemove(root, now, func(path string) error {
		if path == attempt {
			if err := os.Remove(path); err != nil {
				return err
			}
			return errors.New("pruner interrupted after attempt unlink")
		}
		return os.Remove(path)
	}); err == nil {
		t.Fatal("pruner did not expose injected interruption")
	}
	if _, err := os.Stat(audit); err != nil {
		t.Fatalf("proof removed before interrupted attempt removal completed: %v", err)
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audit); !os.IsNotExist(err) {
		t.Fatalf("orphan audit survives future pruning forever: %v", err)
	}
}

func TestSuiteReclassificationPruneRefusalKeepsProof(t *testing.T) {
	root, a, _ := reclassificationFixture(t)
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", a.FinishedAt.Add(time.Hour), nil); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	old := now.Add(-suiteAttemptKeep - time.Hour)
	attempt := filepath.Join(root, "attempts", a.ID+".json")
	if err := os.Chtimes(attempt, old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttemptsWithRemove(root, now, func(path string) error {
		if path == attempt {
			return os.ErrPermission
		}
		return os.Remove(path)
	}); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("prune refusal missing: %v", err)
	}
	for _, path := range []string{attempt, filepath.Join(root, "reclassified", a.ID+".json"), filepath.Join(root, "failed", a.ID+".log")} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("proof dependency removed while authority remains: %s %v", path, err)
		}
	}
}

func TestSuiteReclassificationPruneDuringPublicationPreservesProof(t *testing.T) {
	root, a, _ := reclassificationFixture(t)
	now := time.Now()
	old := now.Add(-suiteAttemptKeep - time.Hour)
	if err := os.Chtimes(filepath.Join(root, "attempts", a.ID+".json"), old, old); err != nil {
		t.Fatal(err)
	}
	private := t.TempDir()
	if err := publishSuiteReclassification(root, a.ID, a.Repo, "ariadne:5", "repair", now, func(path string, b []byte, exclusive bool) error {
		if err := pruneSuiteAttempts(root, now); err != nil {
			return err
		}
		return writeSuiteReclassificationFile(private, path, b, exclusive)
	}); err != nil {
		t.Fatalf("pruner interfered with recovery publication: %v", err)
	}
	audit := filepath.Join(root, "reclassified", a.ID+".json")
	if _, err := os.Stat(audit); err != nil {
		t.Fatalf("committed audit lost before its retention window: %v", err)
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audit); err != nil {
		t.Fatalf("fresh orphan audit lost: %v", err)
	}
	if err := os.Chtimes(audit, old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audit); err != nil {
		t.Fatalf("proof pruned while fresh correction remains: %v", err)
	}
	if err := os.Chtimes(filepath.Join(root, "attempts", a.ID+".json"), old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audit); !os.IsNotExist(err) {
		t.Fatalf("expired audit from prune/publication race survives: %v", err)
	}
}

func FuzzEncodeReclassifiedAttempt(f *testing.F) {
	a := storedAttemptFixture("a1")
	a.Failures, a.Slot, a.LogPath = nil, "ariadne:5", "failed/a1.log"
	original, err := json.Marshal(a)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(original)
	f.Fuzz(func(t *testing.T, raw []byte) {
		encoded, err := encodeReclassifiedAttempt(raw, strings.Repeat("a", 64))
		if err != nil {
			return
		}
		var before, after map[string]json.RawMessage
		if err := json.Unmarshal(raw, &before); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &after); err != nil {
			t.Fatal(err)
		}
		for k, v := range before {
			if k == "version" || k == "outcome" {
				continue
			}
			var left, right bytes.Buffer
			if err := json.Compact(&left, v); err != nil {
				t.Fatal(err)
			}
			if err := json.Compact(&right, after[k]); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(left.Bytes(), right.Bytes()) {
				t.Fatalf("lost original field %q", k)
			}
		}
		decoded, err := decodeSuiteAttempt(encoded)
		if err != nil || decoded.Exit <= 0 || decoded.TimedOut || decoded.Version != 3 {
			t.Fatalf("correction grants unexpected authority: %+v %v", decoded, err)
		}
	})
}
