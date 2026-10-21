package main

import (
	"bytes"
	"context"
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
