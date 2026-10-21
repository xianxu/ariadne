package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const reconcileTestID = "2b6ca56a5f4147569665ee2f073ec211"

// Preserve the incident's schema and unknown fields; PID/start are owned by
// this test, so kernel checks can never depend on an unrelated machine process.
func reconcileCommandFixture(t *testing.T, live bool) (queueCommandDeps, *queuegate.Store, string, []byte, string) {
	t.Helper()
	started := "1792501451.112364"
	if live {
		var err error
		started, err = processStartTime(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
	}
	raw := []byte(fmt.Sprintf(`{"boot_id":"1792451162.532897","detail":"unmanaged operation %s: watched owner ended; descendants require operator reconciliation","id":"%s","kinds":["suite"],"managed":false,"phase":"unknown","pid":%d,"pid_started":%q,"root":"/private/incident/ariadne-slot1/ariadne","schema_version":1,"started_at":"2026-10-20T13:04:11.199355+00:00","workspace":"","future":{"large":1e9999,"opaque":[null,true]}}`, reconcileTestID, reconcileTestID, os.Getpid(), started))
	s := &queuegate.Store{Root: t.TempDir()}
	path := filepath.Join(s.Root, "unresolved", reconcileTestID+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	d := queueFixture(t)
	d.store = func() (*queuegate.Store, error) { return s, nil }
	sum := sha256.Sum256(raw)
	return d, s, path, raw, hex.EncodeToString(sum[:])
}

func TestQueueReconcileIncident(t *testing.T) {
	d, s, path, raw, digest := reconcileCommandFixture(t, false)
	if err := s.Pause(context.Background(), "operator-hold", "keep this owner", false); err != nil {
		t.Fatal(err)
	}
	pause, err := os.ReadFile(filepath.Join(s.Root, "pause.json"))
	if err != nil {
		t.Fatal(err)
	}
	op, err := queuegate.DecodeOperation(raw)
	if err != nil {
		t.Fatal(err)
	}
	interrupted, err := queuegate.EncodeInterruptionRecord(queuegate.InterruptionRecord{SchemaVersion: 1, ID: op.ID, Outcome: "interrupted", Operation: op, Reason: "keep this evidence", RecordedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	ip := filepath.Join(s.Root, "interrupted", op.ID+".json")
	if err := os.MkdirAll(filepath.Dir(ip), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ip, interrupted, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"reconcile", "--id", op.ID, "--digest", digest, "--note", "operator verified no descendants; quoted $(false) is data"}
	if out, err := executeQueue(t, d, args...); err != nil || !strings.Contains(out, "reconciled") {
		t.Fatalf("recovery: %q %v", out, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("tombstone survived: %v", err)
	}
	auditPath := filepath.Join(s.Root, "reconcile.jsonl")
	audit, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(audit), "\n") != 1 {
		t.Fatalf("audit is not one record: %q", audit)
	}
	var record struct {
		SchemaVersion int       `json:"schema_version"`
		ID            string    `json:"id"`
		Digest        string    `json:"digest"`
		Original      []byte    `json:"original"`
		Actor         string    `json:"actor"`
		Note          string    `json:"note"`
		At            time.Time `json:"at"`
	}
	if err := json.Unmarshal(audit, &record); err != nil {
		t.Fatal(err)
	}
	if record.SchemaVersion != 1 || record.ID != op.ID || record.Digest != digest || string(record.Original) != string(raw) || record.Actor == "" || record.Note != args[len(args)-1] || record.At.IsZero() {
		t.Fatalf("audit lost evidence: %+v", record)
	}
	if _, err := executeQueue(t, d, args...); err != nil {
		t.Fatalf("replay: %v", err)
	}
	after, err := os.ReadFile(auditPath)
	if err != nil || string(after) != string(audit) {
		t.Fatalf("replay changed audit: %s %v", after, err)
	}
	after, err = os.ReadFile(filepath.Join(s.Root, "pause.json"))
	if err != nil || string(after) != string(pause) {
		t.Fatal("recovery changed pause owners")
	}
	after, err = os.ReadFile(ip)
	if err != nil || string(after) != string(interrupted) {
		t.Fatal("recovery changed interruption evidence")
	}
	snapshot, err := s.Snapshot(context.Background())
	if err != nil || !snapshot.Complete || len(snapshot.Active) != 0 || len(snapshot.PauseOwners) != 1 {
		t.Fatalf("not reconciled: %+v %v", snapshot, err)
	}
}

func TestQueueReconcileRefusals(t *testing.T) {
	for _, kind := range []string{"wrong digest", "missing note", "live owner", "live lease", "operation residue", "worker", "held tombstone"} {
		t.Run(kind, func(t *testing.T) {
			d, s, path, raw, digest := reconcileCommandFixture(t, kind == "live owner")
			note := "operator verified no descendants"
			if kind == "wrong digest" {
				digest = strings.Repeat("0", 64)
			}
			if kind == "missing note" {
				note = " "
			}
			if kind == "worker" {
				d.authority.caller = func() (string, *int, bool) { n := 3; return "repo:3", &n, true }
			}
			if kind == "live lease" || kind == "operation residue" || kind == "held tombstone" {
				held := path
				if kind != "held tombstone" {
					held = filepath.Join(s.Root, "operations", reconcileTestID+".json")
					if err := os.MkdirAll(filepath.Dir(held), 0700); err != nil {
						t.Fatal(err)
					}
				}
				f, err := os.OpenFile(held, os.O_CREATE|os.O_RDWR, 0600)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if kind != "operation residue" {
					if err := flock.Exclusive(f, true); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := executeQueue(t, d, "reconcile", "--id", reconcileTestID, "--digest", digest, "--note", note); err == nil || strings.Contains(err.Error(), "unknown flag") || strings.Contains(err.Error(), "unknown command") {
				t.Fatalf("refusal did not exercise reconciliation: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != string(raw) {
				t.Fatalf("refusal changed tombstone: %v", err)
			}
			if _, err := os.Stat(filepath.Join(s.Root, "reconcile.jsonl")); !os.IsNotExist(err) {
				t.Fatalf("refusal created audit: %v", err)
			}
		})
	}
}

func TestQueueReconcileOwnerHelper(t *testing.T) {
	if os.Getenv("QUEUE_RECONCILE_OWNER_HELPER") != "1" {
		t.Skip("helper process only")
	}
	fmt.Println("ready")
	time.Sleep(testfix.HangCeiling)
}

func TestQueueReconcileKernelOwnerConformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestQueueReconcileOwnerHelper$")
	cmd.Env = append(os.Environ(), "QUEUE_RECONCILE_OWNER_HELPER=1")
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.WaitDelay = 100 * time.Millisecond
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	if line, err := bufio.NewReader(pipe).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper readiness: %q %v", line, err)
	}
	start, err := processStartTime(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	op := queuegate.Operation{PID: cmd.Process.Pid, PIDStarted: start}
	if absent, err := queueOwnerAbsent(op); err != nil || absent {
		t.Fatalf("live child: absent=%t %v", absent, err)
	}
	op.PIDStarted = "a different incarnation"
	if absent, err := queueOwnerAbsent(op); err != nil || !absent {
		t.Fatalf("reused PID: absent=%t %v", absent, err)
	}
	op.PIDStarted = start
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	if absent, err := queueOwnerAbsent(op); err != nil || !absent {
		t.Fatalf("waited-for dead child: absent=%t %v", absent, err)
	}
}
