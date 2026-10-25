package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Producer-shaped schema-1 report. The mark is nested in the thread row,
// not at the slot level; an unclaimed worker is still eligible.
const asleepRecoveryFixture = `{"schema_version":1,"rows":[{
"address":"pair:2","path":"/work/pair-slot2/pair",
"git":{"source":"sdlc","branch":"main-slot2","state":"resting","dirty":"no","unlanded":"no","operation":"no"},
"claims":{"quality":"present"},
"agent":{"state":"parked","threads":1,"rows":[{"repo_scope":"scope","tag":"thread","state":"parked","asleep_at":"2026-10-24T00:41:07.123-07:00"}]}}]}`

func TestDecodeRecoveryEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, raw                string
		asleep, pending, unknown bool
	}{
		{"unclaimed asleep", asleepRecoveryFixture, true, false, false},
		{"old producer", strings.Replace(asleepRecoveryFixture, `,"asleep_at":"2026-10-24T00:41:07.123-07:00"`, "", 1), false, false, false},
		{"active claim", strings.Replace(asleepRecoveryFixture, `"quality":"present"`, `"quality":"present","active":"pair#2"`, 1), true, true, false},
		{"inactive claim", strings.Replace(asleepRecoveryFixture, `"quality":"present"`, `"quality":"present","inactive":["pair#1"]`, 1), true, false, false},
		{"issue branch", strings.Replace(asleepRecoveryFixture, `"state":"resting"`, `"state":"open-issue"`, 1), true, true, false},
		{"dirty", strings.Replace(asleepRecoveryFixture, `"dirty":"no"`, `"dirty":"yes"`, 1), true, true, false},
		{"unlanded", strings.Replace(asleepRecoveryFixture, `"unlanded":"no"`, `"unlanded":"yes"`, 1), true, true, false},
		{"invalid time", strings.Replace(asleepRecoveryFixture, "2026-10-24T00:41:07.123-07:00", "yesterday", 1), false, false, true},
		{"null time", strings.Replace(asleepRecoveryFixture, `"2026-10-24T00:41:07.123-07:00"`, "null", 1), false, false, true},
		{"claims unknown", strings.Replace(asleepRecoveryFixture, `"quality":"present"`, `"quality":"unknown"`, 1), true, false, true},
		{"git unknown", strings.Replace(asleepRecoveryFixture, `"dirty":"no"`, `"dirty":"unknown"`, 1), true, false, true},
		{"conflicting threads", strings.Replace(asleepRecoveryFixture, `"threads":1`, `"threads":2`, 1), false, false, true},
		{"not parked", strings.ReplaceAll(asleepRecoveryFixture, `"state":"parked"`, `"state":"live"`), false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := decodeRecovery([]byte(tc.raw))
			if err != nil || len(s.Slots) != 1 {
				t.Fatalf("decode: %+v %v", s, err)
			}
			r := s.Slots[0]
			if r.Address != "pair:2" || r.Asleep != tc.asleep || r.Pending != tc.pending || (r.Unknown != "") != tc.unknown {
				t.Fatalf("projection %+v; want asleep=%v pending=%v unknown=%v", r, tc.asleep, tc.pending, tc.unknown)
			}
		})
	}
}

func TestDecodeRecoveryStrictEnvelope(t *testing.T) {
	for _, raw := range []string{
		`null`, `{}`, `{"schema_version":1}`, `{"schema_version":1,"rows":null}`,
		`{"schema_version":1,"rows":[{"address":"pair:2"},{"address":"pair:2"}]}`,
		strings.Replace(asleepRecoveryFixture, `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(asleepRecoveryFixture, `"schema_version":1`, `"schema_version":1,"SCHEMA_VERSION":1`, 1),
		strings.Replace(asleepRecoveryFixture, `"threads":1`, `"threads":"1"`, 1),
		strings.Replace(asleepRecoveryFixture, `"tag":"thread"`, `"tag":"one","TAG":"two"`, 1),
		strings.Replace(asleepRecoveryFixture, `"address":"pair:2"`, `"address":"pair:02"`, 1),
		asleepRecoveryFixture + ` {}`,
		asleepRecoveryFixture[:len(asleepRecoveryFixture)-1],
	} {
		if _, err := decodeRecovery([]byte(raw)); err == nil {
			t.Errorf("accepted unsafe report: %s", raw)
		}
	}
	// Unknown fields are opaque: a future numeric value need not fit float64.
	future := strings.Replace(asleepRecoveryFixture, `"claims":`, `"future":{"x":1e400},"claims":`, 1)
	if _, err := decodeRecovery([]byte(future)); err != nil {
		t.Fatal(err)
	}
}

func TestExecRecoveryReportBoundary(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "couch")
	script := "#!/bin/sh\n[ \"$#\" -eq 1 ] && [ \"$1\" = '--recover-plan-from-sdlc' ] || exit 29\npwd > \"$0.cwd\"\nexec cat \"$0.payload\"\n"
	if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(raw string) {
		t.Helper()
		if err := os.WriteFile(bin+".payload", []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(asleepRecoveryFixture)
	s, err := (execRecovery{bin: bin}).Recovery(t.Context(), root)
	if err != nil || len(s.Slots) != 1 || !s.Slots[0].Asleep {
		t.Fatalf("%+v %v", s, err)
	}
	cwd, err := os.ReadFile(bin + ".cwd")
	if err != nil || strings.TrimSpace(string(cwd)) != root {
		t.Fatalf("cwd=%q err=%v", cwd, err)
	}
	write(strings.Repeat(" ", docRepositoryOutputLimit+1) + asleepRecoveryFixture)
	if _, err := (execRecovery{bin: bin}).Recovery(t.Context(), root); err == nil || !strings.Contains(err.Error(), "output exceeds limit") {
		t.Fatalf("oversized report: %v", err)
	}
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho 'old Couch does not support this report' >&2\nexit 2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := (execRecovery{bin: bin}).Recovery(t.Context(), root); err == nil || !strings.Contains(err.Error(), "old Couch") {
		t.Fatalf("unavailable report: %v", err)
	}
}

// A claim is pending-work evidence, never evidence of a shutdown asleep mark.
func TestFleetResumeDoesNotInferAsleepFromClaims(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	for i := range c.recoveryData.Slots {
		c.recoveryData.Slots[i].Asleep = false
	}
	var out bytes.Buffer
	_ = fleetResume(t.Context(), &out, d, cwd, true)
	if got := c.attempts(); len(got) != 0 {
		t.Fatalf("resumed without positive asleep evidence: %v\n%s", got, out.String())
	}
}

func TestFleetResumeRecoveryCaptureCounts(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			d, c, cwd := fleetFixture(t)
			d.sdlc = nil // resume must not re-derive claims through another inventory
			c.resumeErr = nil
			c.recoveryData.Slots = nil
			for i := 0; i < n; i++ {
				addr := fmt.Sprintf("pair:%d", 100+i)
				c.recoveryData.Slots = append(c.recoveryData.Slots, RecoverySlot{FleetSlot: FleetSlot{Address: addr, Repo: "pair", Slot: 100 + i}, Asleep: true, Pending: i%2 == 0})
				c.candidates[addr] = ResumeCandidates{Proven: true}
			}
			var out bytes.Buffer
			if err := fleetResume(t.Context(), &out, d, cwd, true); err != nil {
				t.Fatal(err)
			}
			if c.recoveryReads != 1 || c.actorsReads != 1 || c.candReads != n || len(c.attempts()) != n {
				t.Fatalf("N=%d: recovery=%d actors=%d candidates=%d attempted=%d", n, c.recoveryReads, c.actorsReads, c.candReads, len(c.attempts()))
			}
		})
	}
}

func TestFleetResumeRequiresCompleteProvenCandidate(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.resumeErr = nil
	for i := range c.recoveryData.Slots {
		c.recoveryData.Slots[i].Asleep = c.recoveryData.Slots[i].Address == "pair:2"
	}
	c.candidates["pair:2"] = ResumeCandidates{Proven: true, Incomplete: true, Reasons: []string{"store unreadable"}}
	var out bytes.Buffer
	_ = fleetResume(t.Context(), &out, d, cwd, true)
	if len(c.attempts()) != 0 || !strings.Contains(out.String(), "candidates incomplete") {
		t.Fatalf("%v\n%s", c.attempts(), out.String())
	}
}
