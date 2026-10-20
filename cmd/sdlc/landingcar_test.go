package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestParseLandingConfig(t *testing.T) {
	cases := []struct {
		body     string
		interval time.Duration
		cars     int
		err      string
	}{
		{`{}`, 0, landingline.DefaultMaxCars, ""},
		{`{"train_interval_minutes": 0}`, 0, landingline.DefaultMaxCars, ""},
		{`{"train_interval_minutes": 60, "train_max_cars": 4}`, time.Hour, 4, ""},
		{`{"train_interval_minutes": 60, "a_later_key": {"x": 1}}`, time.Hour, landingline.DefaultMaxCars, ""},
		{`{"train_interval_minutes": -1}`, 0, 0, "outside 0..1440"},
		{`{"train_interval_minutes": 1441}`, 0, 0, "outside 0..1440"},
		{`{"train_max_cars": 0}`, 0, 0, "outside 1..16"},
		{`{"train_max_wait_minutes": -1}`, 0, 0, "train_max_wait_minutes -1 is outside 0..1440"},
		{`{"train_interval_minutes": "60"}`, 0, 0, "not a JSON object"},
		{`{} {}`, 0, 0, "trailing data"},
		{`[]`, 0, 0, "not a JSON object"},
	}
	for _, c := range cases {
		s, err := parseLandingConfig([]byte(c.body))
		if c.err != "" {
			if err == nil || !strings.Contains(err.Error(), c.err) || !strings.Contains(err.Error(), landingConfigPath) {
				t.Errorf("%s: err %v, want %q naming the path", c.body, err, c.err)
			}
			continue
		}
		if err != nil || s.Interval != c.interval || s.MaxCars != c.cars {
			t.Errorf("%s: %+v %v", c.body, s, err)
		}
	}
	if s, err := parseLandingConfig([]byte(`{"train_interval_minutes": 60, "train_max_wait_minutes": 25}`)); err != nil || s.MaxWait != 25*time.Minute {
		t.Errorf("max wait: %+v %v", s, err)
	}
}

func TestReadLandingScheduleAbsentIsOff(t *testing.T) {
	git := func(args ...string) (string, error) {
		if args[0] == "ls-tree" {
			return "", nil
		}
		t.Fatalf("read a file that is absent: %v", args)
		return "", nil
	}
	s, err := readLandingSchedule(git, "upstream/main")
	if err != nil || !s.Off() {
		t.Fatalf("schedule %+v %v", s, err)
	}
	git = func(args ...string) (string, error) {
		switch args[0] {
		case "ls-tree":
			return landingConfigPath + "\n", nil
		case "show":
			if args[1] != "upstream/main:"+landingConfigPath {
				t.Fatalf("read %v, not main's tree", args)
			}
			return `{"train_interval_minutes": 30}`, nil
		}
		return "", errors.New("unexpected")
	}
	if s, err := readLandingSchedule(git, "upstream/main"); err != nil || s.Interval != 30*time.Minute {
		t.Fatalf("schedule %+v %v", s, err)
	}
}

func TestTrainAdmission(t *testing.T) {
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	sched := landingline.Schedule{Interval: time.Hour, MaxCars: 8}
	notDue := []landingline.Event{{Kind: landingline.EventTrain, NextDue: now.Add(time.Minute)}}
	ready := &landingline.Readiness{Branch: "b", Ready: true, Head: strings.Repeat("a", 40), PR: 1, At: now}
	stale := &landingline.Readiness{Branch: "b", Ready: true, Head: strings.Repeat("a", 40), PR: 1, At: now.Add(-time.Hour)}
	notReady := &landingline.Readiness{Branch: "b", Reason: "CI gate: pending", At: now}
	stranger := &landingline.Readiness{Branch: "000999-other", Ready: true, Head: strings.Repeat("a", 40), PR: 1, At: now}
	cases := []struct {
		name    string
		events  []landingline.Event
		r       *landingline.Readiness
		granted bool
		w       landingline.Waiters
		want    bool
		why     string
	}{
		{"a grant departs at once, unvouched", notDue, nil, true, landingline.Waiters{}, true, ""},
		{"not due", notDue, ready, false, landingline.Waiters{}, false, "next train at"},
		{"due, unvouched", nil, nil, false, landingline.Waiters{}, false, "not vouched"},
		{"due, stale", nil, stale, false, landingline.Waiters{}, false, "stale"},
		{"due, not ready", nil, notReady, false, landingline.Waiters{}, false, "CI gate: pending"},
		{"due, a vouch naming another branch", nil, stranger, false, landingline.Waiters{}, false, "000999-other"},
		{"due and ready", nil, ready, false, landingline.Waiters{}, true, ""},
		// #475: a full train, or a car waiting past the bound, leaves before
		// its departure time; the car itself still needs its own vouch.
		{"not due, but full", notDue, ready, false, landingline.Waiters{Ready: 8, OldestSince: now}, true, ""},
		{"not due, full, unvouched", notDue, nil, false, landingline.Waiters{Ready: 8, OldestSince: now}, false, "not vouched"},
	}
	for _, c := range cases {
		reads := 0
		admit := trainAdmission(func() landingline.Schedule { reads++; return sched }, func() time.Time { return now })()
		got, why := admit(c.events, landingline.Ticket{Seq: 1, Branch: "b"}, c.r, c.granted, c.w)
		admit(c.events, landingline.Ticket{Seq: 2, Branch: "b"}, c.r, c.granted, c.w)
		if reads != 1 {
			t.Errorf("%s: the setting was read %d times for one attempt", c.name, reads)
		}
		if got != c.want || !strings.Contains(why, c.why) {
			t.Errorf("%s: %v %q", c.name, got, why)
		}
	}
	off := trainAdmission(func() landingline.Schedule { return landingline.Schedule{} }, time.Now)()
	if ok, _ := off(nil, landingline.Ticket{Seq: 1}, nil, false, landingline.Waiters{}); !ok {
		t.Error("trains off held an unvouched merge")
	}
}

// An unreadable setting is the serial line (warned once), never a refusal:
// a merge with trains off must not depend on the config's readability.
func TestLandingScheduleSourceFailsToSerial(t *testing.T) {
	var stderr strings.Builder
	git := func(args ...string) (string, error) { return "", errors.New("unknown revision upstream/main") }
	oldTTL := landingScheduleTTL
	landingScheduleTTL = 0 // every call reads, so the warning is what's cached
	t.Cleanup(func() { landingScheduleTTL = oldTTL })
	src := landingScheduleSource(git, "upstream", "sample", &stderr)
	if s := src(); !s.Off() {
		t.Fatalf("schedule %+v", s)
	}
	src()
	if n := strings.Count(stderr.String(), "landing serially"); n != 1 {
		t.Errorf("warned %d times:\n%s", n, stderr.String())
	}
	if got := landingMainRef(git, "upstream"); got != "upstream/main" {
		t.Errorf("main ref %q", got)
	}
	if got := landingMainRef(git, ""); got != "origin/main" {
		t.Errorf("fallback main ref %q", got)
	}
	cfg := func(args ...string) (string, error) { return "upstream\n", nil }
	if got := landingMainRef(cfg, ""); got != "upstream/main" {
		t.Errorf("configured main ref %q", got)
	}
}

func TestLandingPRLabelMarksLocalGates(t *testing.T) {
	pr := `[{"number":7,"state":"OPEN","headRefName":"b","headRefOid":"` + strings.Repeat("a", 40) + `","baseRefName":"main","baseRefOid":"` + strings.Repeat("b", 40) +
		`","url":"https://github.com/test/repo/pull/7","isCrossRepository":false,"headRepository":{"name":"repo","nameWithOwner":"test/repo"},"headRepositoryOwner":{"login":"test"},"labels":[{"name":"bug"},{"name":"local-gates-verified"}]}]`
	got, err := parseLandingPRs([]byte(pr), "test/repo", "b")
	if err != nil || len(got) != 1 || !got[0].LocalGatesVerified {
		t.Fatalf("%+v %v", got, err)
	}
	plain := strings.Replace(pr, `{"name":"local-gates-verified"}`, `{"name":"local-gates-verified-later"}`, 1)
	if got, err := parseLandingPRs([]byte(plain), "test/repo", "b"); err != nil || got[0].LocalGatesVerified {
		t.Fatalf("a near-miss label counted: %+v %v", got, err)
	}
}

// carFixture is a pushed issue branch with an open PR at its head.
func carFixture(t *testing.T) (landingTarget, *landingFakeGH, string) {
	t.Helper()
	roots, _, gh := landingFixture(t, 1)
	target := landingTarget{Root: roots[1], Remote: "upstream", Repo: "test/repo"}
	return target, gh, gh.pr.HeadOID
}

func passGates() carGates {
	ok := func(context.Context, landingPR) error { return nil }
	return carGates{publish: ok, ci: ok}
}

func TestCarReadiness(t *testing.T) {
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	t.Run("ready at its pushed head", func(t *testing.T) {
		target, gh, head := carFixture(t)
		r, _ := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, passGates(), now)
		if !r.Ready || r.Head != head || r.PR != 1 || r.Alone || r.Branch != landingTestBranch || !r.At.Equal(now) {
			t.Fatalf("%+v", r)
		}
		if err := (landingline.Readiness{Version: landingline.ReadinessVersion, Seq: 1, Branch: r.Branch, Head: r.Head, PR: r.PR, Ready: true, At: r.At}).Valid(1); err != nil {
			t.Fatalf("the vouch would not be accepted: %v", err)
		}
	})
	t.Run("a local commit not pushed is not ready", func(t *testing.T) {
		target, gh, _ := carFixture(t)
		procedureWrite(t, target.Root, "more.txt", "x\n")
		git(t, target.Root, "add", ".")
		git(t, target.Root, "commit", "-qm", "unpushed")
		r, _ := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, passGates(), now)
		if r.Ready || !strings.Contains(r.Reason, "not pushed at its head") {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("the PR must be open at the head", func(t *testing.T) {
		target, gh, _ := carFixture(t)
		gh.pr.HeadOID = strings.Repeat("c", 40)
		r, _ := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, passGates(), now)
		if r.Ready || !strings.Contains(r.Reason, "no open PR") {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("a failing gate is not ready, with its reason", func(t *testing.T) {
		target, gh, _ := carFixture(t)
		for name, g := range map[string]carGates{
			"publish gate": {publish: func(context.Context, landingPR) error { return errors.New("HEAD moved since review\nmore") }, ci: passGates().ci},
			"CI gate":      {publish: passGates().publish, ci: func(context.Context, landingPR) error { return errors.New("needs recorded local checks: suite") }},
		} {
			r, _ := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, g, now)
			if r.Ready || !strings.HasPrefix(r.Reason, name) || strings.Contains(r.Reason, "\n") {
				t.Errorf("%s: %+v", name, r)
			}
		}
	})
	t.Run("a PR already merged at the head reports merged", func(t *testing.T) {
		target, gh, _ := carFixture(t)
		gh.pr.State = "MERGED"
		r, merged := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, passGates(), now)
		if r.Ready || !merged {
			t.Fatalf("%+v merged %v", r, merged)
		}
	})
	t.Run("a local-gates PR rides alone", func(t *testing.T) {
		target, gh, _ := carFixture(t)
		gh.pr.LocalGatesVerified = true
		r, _ := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, passGates(), now)
		if !r.Ready || !r.Alone || !strings.Contains(r.Reason, "local-gates") {
			t.Fatalf("%+v", r)
		}
	})
}

// With trains on, a merge first waits for its own readiness, vouches it, and
// is admitted once due and ready; the vouching goroutine ends with the wait.
// The car turns ready only after the line reported it held as not ready, so
// the order is fixed by the test, not by the scheduler.
func TestEnterLandingLineVouchesWhileWaiting(t *testing.T) {
	target, _, head := carFixture(t)
	oldPoll, oldEvery := landingLinePoll, carVouchEvery
	landingLinePoll, carVouchEvery = time.Millisecond, time.Millisecond
	t.Cleanup(func() { landingLinePoll, carVouchEvery = oldPoll, oldEvery })
	stderr := &watchWriter{want: "held: train due; not ready: CI gate: pending", seen: make(chan struct{})}
	var calls atomic.Int64
	readiness := func(context.Context) (landingline.Readiness, bool) {
		calls.Add(1)
		select {
		case <-stderr.seen:
			return landingline.Readiness{Branch: landingTestBranch, Head: head, PR: 1, Ready: true, At: landingNow()}, false
		default:
			return landingline.Readiness{Branch: landingTestBranch, Reason: "CI gate: pending", At: landingNow()}, false
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sched := func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour, MaxCars: 8} }
	turn, err := enterLandingLine(ctx, execGitRunner{}, target, landingTestBranch, sched, readiness, stderr)
	if err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	defer turn.leave(nil)
	out := stderr.String()
	for _, want := range []string{"landing train: not ready: CI gate: pending", "landing train: " + stderr.want, "landing train: ready at"} {
		if !strings.Contains(out, want) {
			t.Errorf("stderr lacks %q:\n%s", want, out)
		}
	}
	n := calls.Load()
	time.Sleep(20 * time.Millisecond)
	if got := calls.Load(); got != n {
		t.Errorf("vouching continued after admission: %d → %d calls", n, got)
	}
}

// watchWriter is a goroutine-safe stderr that closes seen once want appears.
type watchWriter struct {
	mu   sync.Mutex
	b    strings.Builder
	want string
	seen chan struct{}
	done bool
}

func (w *watchWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.b.Write(p)
	if !w.done && strings.Contains(w.b.String(), w.want) {
		w.done = true
		close(w.seen)
	}
	return len(p), nil
}

func (w *watchWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

// The vouching loop consults the schedule each round: while trains are off it
// reads no gates, and once they turn on (main changed during the wait) it
// starts vouching, so a long wait can still board.
func TestVouchingStartsWhenTrainsTurnOn(t *testing.T) {
	oldEvery := carVouchEvery
	carVouchEvery = time.Millisecond
	t.Cleanup(func() { carVouchEvery = oldEvery })
	dir := t.TempDir()
	line, _, err := openLandingLine(dir, nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: "b"})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	var rounds, gateReads atomic.Int64
	sched := func() landingline.Schedule {
		if rounds.Add(1) <= 5 {
			return landingline.Schedule{}
		}
		return landingline.Schedule{Interval: time.Hour, MaxCars: 8}
	}
	readiness := func(context.Context) (landingline.Readiness, bool) {
		if rounds.Load() <= 5 {
			t.Error("gates read while trains were off")
		}
		gateReads.Add(1)
		return landingline.Readiness{Branch: "b", Head: strings.Repeat("a", 40), PR: 1, Ready: true, At: landingNow()}, false
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := vouchWhileWaiting(ctx, line, h, sched, readiness, func() { t.Error("reported merged") }, io.Discard)
	deadline := time.Now().Add(10 * time.Second)
	for gateReads.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	ready, err := line.Readiness()
	if err != nil || !ready[h.Ticket().Seq].Ready {
		t.Fatalf("no vouch after trains turned on: %v %v", ready, err)
	}
}

// The schedule source is shared by the vouching goroutine and the admission
// poll; run with -race, this is that sharing. It also reads at most once per
// TTL however often it is asked.
func TestLandingScheduleSourceIsSharedAndCached(t *testing.T) {
	var reads atomic.Int64
	git := func(args ...string) (string, error) {
		reads.Add(1)
		return "", errors.New("unreadable")
	}
	src := landingScheduleSource(git, "upstream", "sample", io.Discard)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				src()
			}
		}()
	}
	wg.Wait()
	if n := reads.Load(); n != 1 {
		t.Fatalf("400 calls within one TTL ran %d config reads", n)
	}
}

// landedCommit resolves a merged PR's landing for every mergeCommit GitHub
// may report for a car merged by reachability (#449, PQ-3).
func TestLandedCommitAcceptsEveryMergeCommitVariant(t *testing.T) {
	head, train, other := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
	onMain := map[string]bool{head: true, train: true}
	in := func(oid string) bool { return onMain[oid] }
	for name, c := range map[string]struct {
		merge, want string
		ok          bool
	}{
		"the train's merge commit":   {train, train, true},
		"a null mergeCommit":         {"", head, true},
		"the car's own head":         {head, head, true},
		"a merge commit not on main": {other, head, true},
	} {
		got, ok := landedCommit(landingPR{State: "MERGED", HeadOID: head, MergeOID: c.merge}, in)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: %q %v", name, got, ok)
		}
	}
	if _, ok := landedCommit(landingPR{State: "MERGED", HeadOID: other, MergeOID: other}, in); ok {
		t.Error("a PR with nothing on main resolved")
	}
	if _, ok := landedCommit(landingPR{State: "OPEN", HeadOID: head, MergeOID: train}, in); ok {
		t.Error("an open PR resolved")
	}
}

// Only GitHub's decoder and landedCommit read the raw mergeCommit; every
// consumer reads the resolved LandedOID (ARCH-PURPOSE: the class, not a site).
func TestOnlyLandedCommitReadsMergeOID(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "ghlanding.go" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), ".MergeOID") {
			t.Errorf("%s reads MergeOID; read LandedOID (resolved by landedCommit)", f)
		}
	}
}

// The train and the line take their GitHub seam from where sdlc merge wires
// it (landing.go); none of their files reaches for the global client.
func TestTrainCodeTakesItsGitHubSeam(t *testing.T) {
	files, err := filepath.Glob("landing*.go")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, f := range files {
		if f == "landing.go" || strings.HasSuffix(f, "_test.go") {
			continue // landing.go is where sdlc merge wires the seam
		}
		checked++
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "ghClient") {
			t.Errorf("%s reads the global ghClient; take the seam as a parameter", f)
		}
	}
	if checked < 5 {
		t.Fatalf("checked only %d landing files; the glob is stale", checked)
	}
}
