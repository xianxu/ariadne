package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResumeCLIValidation(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"resume"}, "give --name suite"},
		{[]string{"resume", "--name", "affected", "--attempt", "aa"}, "only suite"},
		{[]string{"resume", "--name", "suite"}, "give --attempt"},
		{[]string{"resume", "--name", "suite", "--attempt", "../aa"}, "safe parent"},
		{[]string{"reconcile"}, "give --attempt"},
		{[]string{"reconcile", "--attempt", "aa"}, "--link-digest"},
		{[]string{"reconcile", "--attempt", "aa", "--link-digest", strings.Repeat("a", 64)}, "--note"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			c := NewLocalCheckCmd()
			var out bytes.Buffer
			c.SetOut(&out)
			c.SetErr(&out)
			c.SetArgs(tc.args)
			err := c.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("CLI %v: %v; want %s", tc.args, err, tc.want)
			}
		})
	}
}

func TestResumeReconcileRefusesUnprovedOwner(t *testing.T) {
	for _, tc := range []struct {
		name   string
		child  bool
		mutate func(*ResumeLink)
		digest bool
		want   string
	}{
		{name: "live incarnation", mutate: func(l *ResumeLink) {
			l.Owner.PID = os.Getpid()
			l.Owner.PIDStarted, _ = processStartTime(os.Getpid())
			l.Owner.BootID, _ = queueBootID()
			l.Owner.Host = hostName()
		}, want: "still present"},
		{name: "other host", mutate: func(l *ResumeLink) { l.Owner.Host = "foreign-host" }, want: "host differs"},
		{name: "wrong digest", digest: true, want: "exact immutable link digest"},
		{name: "completed child", child: true, want: "committed child"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, _, store, l := resumeSavedFixture(t, tc.child)
			if tc.mutate != nil {
				tc.mutate(&l)
			}
			resumeSavedWriteJSON(t, store, l.ID, l)
			raw, err := os.ReadFile(filepath.Join(store, "attempts", l.ID+".json"))
			if err != nil {
				t.Fatal(err)
			}
			digest := resumeRecordDigest(raw)
			if tc.digest {
				digest = strings.Repeat("f", 64)
			}
			err = reconcileSavedResume(context.Background(), d, l.RerunOf, digest, "operator verified descendants stopped", "private operator")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("refusal: %v; want %s", err, tc.want)
			}
			snapshot, err := readSuiteEvidence(store, testStatsRecordLimit)
			if err != nil || len(snapshot.Reconciliations) != 0 {
				t.Fatalf("refusal mutated receipt store: %+v %v", snapshot, err)
			}
		})
	}
}

func TestResumeReconcileRequiresSteward(t *testing.T) {
	slot := 3
	d := landingDeps{caller: func() (string, *int, bool) { return "ariadne:3", &slot, true }, couchTree: func() (*int, bool) { return &slot, true }}
	c := newLocalCheckReconcileCmd(d)
	c.SetArgs([]string{"--attempt", "aa", "--link-digest", strings.Repeat("a", 64), "--note", "verified no descendants remain"})
	if err := c.Execute(); err == nil || !strings.Contains(err.Error(), ":0") {
		t.Fatalf("worker authorized reconcile: %v", err)
	}
}

func TestResumeScratchRetirement(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(fmt.Sprint(remove), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "owned-scratch")
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			calls := 0
			d := localRunDeps{runnerLifetime: &runnerArtifactLifetime{retirable: true}}
			scratch := scratchPreparation{Dir: path, Cleanup: func() {
				calls++
				if remove {
					_ = os.RemoveAll(path)
				}
			}}
			err := finishResumeScratch(d, suiteSlotRun{Exit: 0}, scratch, nil)
			if remove {
				if err != nil || calls != 1 {
					t.Fatalf("successful pass %v calls%d", err, calls)
				}
				return
			}
			if err == nil || calls != 1 {
				t.Fatalf("silent removal failure %v calls%d", err, calls)
			}
			if err = finishResumeScratch(d, suiteSlotRun{Exit: 0}, scratch, nil); err == nil || calls != 1 {
				t.Fatalf("failure retried cleanup: %v calls%d", err, calls)
			}
			if _, err = os.Stat(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResumeScratchStopsOnManagedUncertainty(t *testing.T) {
	for _, tc := range []struct {
		name      string
		run       suiteSlotRun
		err       error
		confirmed bool
		keep      bool
	}{
		{name: "canceled", run: suiteSlotRun{Exit: 0}, err: context.Canceled, confirmed: true, keep: true},
		{name: "timeout", run: suiteSlotRun{Exit: -1, TimedOut: true}, confirmed: true, keep: true},
		{name: "did not launch", run: suiteSlotRun{Exit: -1}, confirmed: true, keep: true},
		{name: "signal exit", run: suiteSlotRun{Exit: 137}, confirmed: true, keep: true},
		{name: "group uncertain", run: suiteSlotRun{Exit: 0}, keep: true},
		{name: "ordinary failed test", run: suiteSlotRun{Exit: 1}, confirmed: true},
		{name: "ordinary pass", run: suiteSlotRun{Exit: 0}, confirmed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "scratch")
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			calls := 0
			d := localRunDeps{runnerLifetime: &runnerArtifactLifetime{retirable: tc.confirmed}}
			scratch := scratchPreparation{Dir: path, Cleanup: func() { calls++; _ = os.RemoveAll(path) }}
			err := finishResumeScratch(d, tc.run, scratch, tc.err)
			if tc.keep {
				if err == nil || calls != 0 {
					t.Fatalf("uncertainty deleted scratch: err=%v calls=%d", err, calls)
				}
				if _, e := os.Stat(path); e != nil {
					t.Fatal(e)
				}
				if d.runnerLifetime.mayRetire() {
					t.Fatal("uncertainty retained cleanup permission")
				}
				if tc.err != nil && !errors.Is(err, tc.err) {
					t.Fatalf("lost cause: %v", err)
				}
			} else if err != nil || calls != 1 {
				t.Fatalf("ordinary completion: err=%v calls=%d", err, calls)
			}
		})
	}
}
