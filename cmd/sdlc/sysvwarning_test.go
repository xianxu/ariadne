package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/sysvshm"
)

func TestSysVFooterDoesNotChangeOutcomeAndSurvivesReplay(t *testing.T) {
	r := sysvshm.Report{Before: 1, After: 2, Candidates: []sysvshm.Candidate{{Segment: sysvshm.Segment{ID: 7, Creator: 42, Size: 56}}}}
	var out bytes.Buffer
	writeSysVReport(&out, "suite", "head", r)
	for _, want := range []string{"new segments appeared during this run, not attributed", "id=7", "creator=42", "command=\"unknown\""} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, &out)
		}
	}
	rec := localRunRecord{Name: "suite", SHA: "head", Exit: 0, Tail: strings.Split(strings.TrimSpace(out.String()), "\n"), Scope: strings.Repeat("long", 80)}
	if !rec.passed() || rec.state() != "success" || !strings.HasPrefix(rec.statusDescription(), "SysV warning") {
		t.Fatalf("%+v %s", rec, rec.statusDescription())
	}
	writeSysVReport(&out, "suite", "head", sysvshm.Report{})
	rec.Tail = strings.Split(strings.TrimSpace(out.String()), "\n")
	if strings.Contains(rec.statusDescription(), "SysV warning") {
		t.Fatal("earlier child-looking footer overrode final clean footer")
	}
}

type fakeSysVObserver struct {
	segments            []sysvshm.Segment
	captures, witnesses int
}

func (f *fakeSysVObserver) Snapshot(context.Context) sysvshm.Snapshot {
	f.captures++
	if f.captures == 1 {
		return sysvshm.Snapshot{}
	}
	return sysvshm.Snapshot{Segments: append([]sysvshm.Segment(nil), f.segments...)}
}
func (f *fakeSysVObserver) WitnessGroup(context.Context, int) map[int]sysvshm.Witness {
	f.witnesses++
	return nil
}

func TestSysVManagedRunWarningIsHashedWithoutChangingPass(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "acquired")
	f := &fakeSysVObserver{segments: []sysvshm.Segment{{ID: 7, UID: os.Geteuid(), Creator: 42, Size: 56}}}
	d.sysvObserver = f
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "zellij", Command: []string{"sh", "-c", "echo ok"}, TimeoutMinutes: 1}, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("%+v %v %s", rec, err, stderr)
	}
	if f.captures != 2 || f.witnesses != 1 {
		t.Fatalf("captures=%d witnesses=%d", f.captures, f.witnesses)
	}
	if !strings.HasPrefix(rec.statusDescription(), "SysV warning") || !strings.Contains(stderr.String(), "not attributed") {
		t.Fatalf("%+v %s", rec, stderr)
	}
	hash, tail, err := hashAndTail(rec.LogPath, localCheckTailLines)
	if err != nil || hash != rec.LogSHA256 || !hasSysVWarning(tail, rec.Name, sha) {
		t.Fatalf("hash=%s saved=%s err=%v tail=%v", hash, rec.LogSHA256, err, tail)
	}
}
