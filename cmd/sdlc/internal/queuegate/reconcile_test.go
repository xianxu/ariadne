package queuegate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func reconcileFixture(t *testing.T, root string) (*Store, ReconcileRequest, []byte) {
	t.Helper()
	raw := []byte(`{"boot_id":"1792451162.532897","detail":"unmanaged operation: watched owner ended; descendants require operator reconciliation","id":"2b6ca56a5f4147569665ee2f073ec211","kinds":["suite"],"managed":false,"phase":"unknown","pid":91801,"pid_started":"1792501451.112364","root":"/private/ariadne-slot1/ariadne","schema_version":1,"started_at":"2026-10-20T13:04:11.199355+00:00","workspace":"","future":{"opaque":1e9999}}`)
	sum := sha256.Sum256(raw)
	req := ReconcileRequest{ID: "2b6ca56a5f4147569665ee2f073ec211", Digest: hex.EncodeToString(sum[:]), Actor: "operator uid=501", Note: "verified no descendants remain"}
	path := filepath.Join(root, "unresolved", req.ID+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return &Store{Root: root}, req, raw
}

func absentOwner(Operation) (bool, error) { return true, nil }

func TestReconcileCrashHelper(t *testing.T) {
	root := os.Getenv("QUEUE_RECONCILE_CRASH_ROOT")
	if root == "" {
		t.Skip("helper process only")
	}
	s, req, _ := reconcileFixture(t, root)
	s.afterReconcileAudit = func() error { os.Exit(71); return nil }
	if err := s.Reconcile(context.Background(), req, absentOwner); err != nil {
		t.Fatal(err)
	}
	t.Fatal("did not crash at the durable boundary")
}

func TestReconcileCrashBetweenAuditAndUnlink(t *testing.T) {
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestReconcileCrashHelper$")
	cmd.Env = append(os.Environ(), "QUEUE_RECONCILE_CRASH_ROOT="+root)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 71 {
		t.Fatalf("crash helper: %s %v", out, err)
	}
	path := filepath.Join(root, "unresolved", "2b6ca56a5f4147569665ee2f073ec211.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("crash erased tombstone: %v", err)
	}
	logPath := filepath.Join(root, "reconcile.jsonl")
	audit, err := os.ReadFile(logPath)
	if err != nil || bytes.Count(audit, []byte("\n")) != 1 {
		t.Fatalf("crash lost durable audit: %q %v", audit, err)
	}
	sum := sha256.Sum256(raw)
	req := ReconcileRequest{ID: strings.TrimSuffix(filepath.Base(path), ".json"), Digest: hex.EncodeToString(sum[:]), Actor: "second verified actor", Note: "operator rechecked descendants"}
	s := &Store{Root: root}
	if err := s.Reconcile(context.Background(), req, absentOwner); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("retry kept tombstone: %v", err)
	}
	got, err := os.ReadFile(logPath)
	if err != nil || !bytes.Equal(got, audit) {
		t.Fatalf("retry duplicated or rewrote audit: %s %v", got, err)
	}
	if err := s.Reconcile(context.Background(), req, absentOwner); err != nil {
		t.Fatalf("lost-response retry: %v", err)
	}
	got, _ = os.ReadFile(logPath)
	if !bytes.Equal(got, audit) {
		t.Fatal("lost-response retry changed original audit")
	}
}

func TestReconcileUnsafeEvidence(t *testing.T) {
	for _, kind := range []string{"unknown owner", "live owner", "missing probe", "corrupt log", "torn log", "audit is directory", "wrong phase", "missing incarnation", "unsafe id", "symlink", "changed record after audit"} {
		t.Run(kind, func(t *testing.T) {
			s, req, raw := reconcileFixture(t, t.TempDir())
			path := filepath.Join(s.Root, "unresolved", req.ID+".json")
			probe := absentOwner
			switch kind {
			case "unknown owner":
				probe = func(Operation) (bool, error) { return false, errors.New("identity unavailable") }
			case "live owner":
				probe = func(Operation) (bool, error) { return false, nil }
			case "missing probe":
				probe = nil
			case "corrupt log":
				if err := os.WriteFile(filepath.Join(s.Root, "reconcile.jsonl"), []byte("{}\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "torn log":
				if err := os.WriteFile(filepath.Join(s.Root, "reconcile.jsonl"), []byte("{\"schema_version\":1"), 0600); err != nil {
					t.Fatal(err)
				}
			case "audit is directory":
				if err := os.Mkdir(filepath.Join(s.Root, "reconcile.jsonl"), 0700); err != nil {
					t.Fatal(err)
				}
			case "wrong phase":
				raw = bytes.Replace(raw, []byte(`"phase":"unknown"`), []byte(`"phase":"running"`), 1)
			case "missing incarnation":
				raw = bytes.Replace(raw, []byte(`"pid_started":"1792501451.112364"`), []byte(`"pid_started":""`), 1)
			case "unsafe id":
				req.ID = "../escape"
			case "symlink":
				target := filepath.Join(s.Root, "must-not-delete.json")
				if err := os.WriteFile(target, raw, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			case "changed record after audit":
				s.afterReconcileAudit = func() error { return errors.New("simulated crash") }
				if err := s.Reconcile(context.Background(), req, absentOwner); err == nil {
					t.Fatal("injected crash passed")
				}
				s.afterReconcileAudit = nil
				raw = append(raw, ' ')
			}
			if kind != "symlink" {
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "wrong phase" || kind == "missing incarnation" {
				sum := sha256.Sum256(raw)
				req.Digest = hex.EncodeToString(sum[:])
			}
			if err := s.Reconcile(context.Background(), req, probe); err == nil {
				t.Fatal("unsafe evidence reconciled")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, raw) {
				t.Fatalf("refusal changed record: %s %v", after, err)
			}
		})
	}
}

func TestReconcileAuditCodecPreservesFutureInput(t *testing.T) {
	_, req, raw := reconcileFixture(t, t.TempDir())
	record := reconcileRecord{SchemaVersion: 1, ID: req.ID, Digest: req.Digest, Original: raw, Actor: req.Actor, Note: req.Note, At: time.Now().UTC()}
	original, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	future := append(append([]byte(nil), original[:len(original)-1]...), []byte(`,"future":{"opaque":1e9999}}`)...)
	decoded, err := decodeReconcileRecord(future)
	if err != nil || !bytes.Equal(decoded.Original, raw) {
		t.Fatalf("newer audit refused or original lost: %v", err)
	}
	for _, bad := range [][]byte{
		bytes.Replace(original, []byte(`"actor":`), []byte(`"actor":null,"shadow":`), 1),
		bytes.Replace(original, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1),
		bytes.Replace(original, []byte(`"schema_version":1`), []byte(`"schema_version":"1"`), 1),
		bytes.Replace(original, []byte(`"schema_version":1`), []byte(`"schema_version":2`), 1),
		bytes.Replace(original, []byte(req.Digest), []byte(strings.Repeat("0", 64)), 1),
	} {
		if _, err := decodeReconcileRecord(bad); err == nil {
			t.Fatalf("invalid audit accepted: %s", bad)
		}
	}
	raw = bytes.Replace(raw, []byte(`"opaque":1e9999`), []byte(`"opaque":"`+strings.Repeat("x", 900000)+`"`), 1)
	sum := sha256.Sum256(raw)
	record.Original = raw
	record.Digest = hex.EncodeToString(sum[:])
	large, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if len(large) <= MaxJSON {
		t.Fatal("test did not exceed the operation's JSON bound")
	}
	if _, err := decodeReconcileRecord(large); err != nil {
		t.Fatalf("valid base64 audit refused: %v", err)
	}
}

func TestReconcilePreservesCancellationEvidence(t *testing.T) {
	s, req, _ := reconcileFixture(t, t.TempDir())
	if err := s.Pause(context.Background(), "keep-owner", "operator hold", false); err != nil {
		t.Fatal(err)
	}
	intent := []byte(`{"schema_version":1,"id":"` + req.ID + `","reason":"operator cancelled this suite","future":{"opaque":true}}`)
	if err := atomicWrite(s.intentPath(req.ID), intent); err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), req, absentOwner); err != nil {
		t.Fatal(err)
	}
	retained, err := os.ReadFile(filepath.Join(s.Root, "interrupted", req.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	record, err := DecodeInterruptionRecord(retained)
	if err != nil || record.ID != req.ID || record.Outcome != "interrupted" {
		t.Fatalf("lost interruption: %+v %v", record, err)
	}
	snapshot, err := s.Snapshot(context.Background())
	if err != nil || !snapshot.Complete || len(snapshot.Active) != 0 || len(snapshot.PauseOwners) != 1 {
		t.Fatalf("orphan intent fabricated activity: %+v %v", snapshot, err)
	}
	after, err := os.ReadFile(filepath.Join(s.Root, "interrupted", req.ID+".json"))
	if err != nil || !bytes.Equal(after, retained) {
		t.Fatal("status rewrote interrupted evidence")
	}
}
