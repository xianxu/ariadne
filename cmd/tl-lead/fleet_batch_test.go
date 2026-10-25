package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"testing/synctest"
)

func batchRows() []resumeRow {
	return []resumeRow{
		{Addr: "pair:4", Class: classResume},
		{Addr: "pair:3", Class: classResume, Pending: true},
		{Addr: "pair:1", Class: classResume, Pending: true},
		{Addr: "pair:2", Class: classResume, Pending: true},
	}
}

type barrierResume struct {
	base    *fakeCouch
	started chan string
	release map[string]chan struct{}
}

func (f *barrierResume) Candidates(addr string) (ResumeCandidates, error) {
	return f.base.Candidates(addr)
}
func (f *barrierResume) Resume(addr string) error {
	f.started <- addr
	<-f.release[addr]
	return f.base.Resume(addr)
}

func TestFleetResumeConcurrentWavesThroughCLI(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.live = map[string]bool{}
	c.resumeErr = nil
	c.recoveryData.Slots = nil
	f := &barrierResume{base: c}
	for i := 1; i <= 4; i++ {
		addr := fmt.Sprintf("pair:%d", i)
		c.recoveryData.Slots = append(c.recoveryData.Slots, RecoverySlot{FleetSlot: FleetSlot{Address: addr, Repo: "pair", Slot: i}, Asleep: true, Pending: i < 4})
		c.candidates[addr] = ResumeCandidates{Proven: true}
	}
	d.resume = f
	t.Chdir(cwd)
	synctest.Test(t, func(t *testing.T) {
		f.started = make(chan string, 4)
		f.release = map[string]chan struct{}{}
		for _, row := range c.recoveryData.Slots {
			f.release[row.Address] = make(chan struct{})
		}
		var out strings.Builder
		root := buildRoot(d)
		root.SetOut(&out)
		root.SetContext(t.Context())
		root.SetArgs([]string{"fleet", "resume", "--apply", "--concurrency", "2"})
		done := make(chan error, 1)
		exited := make(chan struct{})
		go func() { defer close(exited); done <- root.Execute() }()
		t.Cleanup(func() {
			for _, ch := range f.release {
				select {
				case <-ch:
				default:
					close(ch)
				}
			}
			<-exited
		})
		next := func() string {
			t.Helper()
			select {
			case addr := <-f.started:
				return addr
			case err := <-done:
				t.Fatalf("command returned before launch: %v\n%s", err, out.String())
			case <-t.Context().Done():
				t.Fatal("enclosing test canceled before launch")
			}
			return ""
		}
		first := []string{next(), next()}
		sort.Strings(first)
		if strings.Join(first, ",") != "pair:1,pair:2" {
			t.Fatal(first)
		}
		synctest.Wait()
		select {
		case addr := <-f.started:
			t.Fatalf("exceeded concurrency: %s", addr)
		default:
		}
		close(f.release["pair:2"])
		if addr := next(); addr != "pair:3" {
			t.Fatalf("pending ordering: %s", addr)
		}
		close(f.release["pair:1"])
		synctest.Wait()
		select {
		case addr := <-f.started:
			t.Fatalf("crossed pending barrier: %s", addr)
		default:
		}
		close(f.release["pair:3"])
		if addr := next(); addr != "pair:4" {
			t.Fatalf("remaining wave: %s", addr)
		}
		close(f.release["pair:4"])
		if err := <-done; err != nil {
			t.Fatalf("%v\n%s", err, out.String())
		}
		<-exited
	})
}

func admitBatch(t *testing.T, b *ResumeBatch, want string, wave int) {
	t.Helper()
	effects := advanceResumeBatch(b, resumeBatchEvent{kind: batchAdmit})
	if want == "" {
		if len(effects) != 0 {
			t.Fatalf("unexpected launch: %+v", effects)
		}
		return
	}
	if len(effects) != 1 || effects[0].Addr != want || effects[0].Wave != wave {
		t.Fatalf("launch=%+v want %s wave%d", effects, want, wave)
	}
}

func TestResumeBatchBoundAndWaveBarrier(t *testing.T) {
	for _, limit := range []int{1, 3} {
		b, err := newResumeBatch(batchRows(), limit)
		if err != nil {
			t.Fatal(err)
		}
		admitBatch(t, b, "pair:1", 1)
		if limit == 1 {
			admitBatch(t, b, "", 0)
			advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:1"})
		}
		admitBatch(t, b, "pair:2", 1)
		if limit == 1 {
			advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:2"})
		}
		admitBatch(t, b, "pair:3", 1)
		admitBatch(t, b, "", 0)
		// Completion order cannot let remaining-asleep work cross the barrier.
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:3"})
		if limit == 3 {
			admitBatch(t, b, "", 0)
			advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:2"})
			admitBatch(t, b, "", 0)
			advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:1"})
		}
		admitBatch(t, b, "pair:4", 2)
		if b.finished() {
			t.Fatal("returned before last terminal receipt")
		}
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:4"})
		if !b.finished() || len(b.untouched()) != 0 {
			t.Fatal("did not finish exact set")
		}
	}
}

func TestResumeBatchStopsAndJoins(t *testing.T) {
	for _, failure := range []error{errors.New("unclassified"), errUncertain, context.Canceled} {
		b, err := newResumeBatch(batchRows(), 3)
		if err != nil {
			t.Fatal(err)
		}
		admitBatch(t, b, "pair:1", 1)
		admitBatch(t, b, "pair:2", 1)
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:1", err: failure})
		admitBatch(t, b, "", 0)
		if b.finished() {
			t.Fatal("failed call dropped in-flight operation")
		}
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:2"})
		if !b.finished() || len(b.untouched()) != 2 {
			t.Fatalf("untouched=%v", b.untouched())
		}
	}
	b, _ := newResumeBatch(batchRows(), 3)
	advanceResumeBatch(b, resumeBatchEvent{kind: batchCancel, err: context.Canceled})
	admitBatch(t, b, "", 0)
	if !b.finished() || len(b.untouched()) != 4 {
		t.Fatal("canceled batch admitted work")
	}
}

func TestResumeBatchRejectsInvalidInputs(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, err := newResumeBatch(batchRows(), n); err == nil {
			t.Fatalf("accepted limit%d", n)
		}
	}
	if _, err := newResumeBatch([]resumeRow{{Addr: "pair:1", Class: classResume}, {Addr: "pair:1", Class: classResume}}, 3); err == nil {
		t.Fatal("accepted duplicate operation")
	}
}

type batchCodedError struct {
	key     resumeFailureKey
	message string
}

func (e batchCodedError) Error() string { return e.message }
func (e batchCodedError) failureKey() (resumeFailureKey, bool) {
	return e.key, e.key.Code != "" || e.key.Diagnostic != ""
}

func TestResumeBatchStableFailureIdentityAndReplay(t *testing.T) {
	b, _ := newResumeBatch(batchRows(), 3)
	admitBatch(t, b, "pair:1", 1)
	admitBatch(t, b, "pair:2", 1)
	key := resumeFailureKey{Status: "refused", Code: "not-offered"}
	first := batchCodedError{key: key, message: "request-A pair:1"}
	second := batchCodedError{key: key, message: "request-B pair:2"}
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:1", err: first})
	// A replay/foreign result cannot count as the second shared failure.
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:1", err: first})
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "foreign", err: first})
	admitBatch(t, b, "pair:3", 1)
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:2", err: second})
	admitBatch(t, b, "", 0)
	if b.finished() {
		t.Fatal("shared failure abandoned in-flight call")
	}
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:3"})
	if !b.finished() || len(b.untouched()) != 1 || b.resultError() == nil {
		t.Fatal("missing failed/untouched outcomes")
	}
}

func TestResumeBatchDistinctCodedFailuresContinue(t *testing.T) {
	b, _ := newResumeBatch(batchRows(), 1)
	for i, addr := range []string{"pair:1", "pair:2", "pair:3"} {
		admitBatch(t, b, addr, 1)
		key := resumeFailureKey{Status: "refused", Code: string(rune('a' + i))}
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: addr, err: batchCodedError{key: key, message: addr}})
	}
	admitBatch(t, b, "pair:4", 2)
	advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: "pair:4"})
	if !b.finished() || b.resultError() == nil || len(b.untouched()) != 0 {
		t.Fatal("isolated failures lost partial progress")
	}
}

func TestResumeCoordinatorStopJoinsEveryStartedCall(t *testing.T) {
	for _, mode := range []string{"cancel", "shared failure", "uncertain", "unclassified"} {
		t.Run(mode, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				base := &fakeCouch{live: map[string]bool{}, resumeErr: map[string]error{}}
				f := &barrierResume{base: base, started: make(chan string, 4), release: map[string]chan struct{}{}}
				for _, row := range batchRows() {
					f.release[row.Addr] = make(chan struct{})
				}
				switch mode {
				case "shared failure":
					for _, addr := range []string{"pair:1", "pair:2"} {
						base.resumeErr[addr] = &resumeFailure{receipt: resumeReceipt{ID: addr, Status: "refused", Code: "not-offered"}, details: addr}
					}
				case "uncertain":
					base.resumeErr["pair:1"] = fmt.Errorf("%w: request-one", errUncertain)
				case "unclassified":
					base.resumeErr["pair:1"] = errors.New("did not run")
				}
				var out strings.Builder
				done := make(chan error, 1)
				exited := make(chan struct{})
				go func() {
					defer close(exited)
					done <- runResumeBatch(ctx, &out, deps{resume: f, actors: base}, "ops", batchRows(), 2)
				}()
				t.Cleanup(func() {
					for _, ch := range f.release {
						select {
						case <-ch:
						default:
							close(ch)
						}
					}
					<-exited
				})
				next := func() string {
					t.Helper()
					select {
					case addr := <-f.started:
						return addr
					case err := <-done:
						t.Fatalf("returned before launch: %v", err)
					}
					return ""
				}
				first := []string{next(), next()}
				sort.Strings(first)
				if strings.Join(first, ",") != "pair:1,pair:2" {
					t.Fatal(first)
				}
				if mode == "cancel" {
					cancel()
				} else {
					close(f.release["pair:1"])
				}
				last := "pair:2"
				if mode == "shared failure" {
					if addr := next(); addr != "pair:3" {
						t.Fatalf("first stable failure must permit progress: %s", addr)
					}
					close(f.release["pair:2"])
					last = "pair:3"
				}
				synctest.Wait()
				select {
				case err := <-done:
					t.Fatalf("abandoned active call: %v", err)
				default:
				}
				select {
				case addr := <-f.started:
					t.Fatalf("launched after stop: %s", addr)
				default:
				}
				if mode == "cancel" {
					close(f.release["pair:1"])
				}
				close(f.release[last])
				err := <-done
				<-exited
				if err == nil || !strings.Contains(out.String(), "worker waves incomplete") || strings.Contains(out.String(), "worker waves 1 and 2 complete") {
					t.Fatalf("wrong handoff: %v\n%s", err, out.String())
				}
				if mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if mode == "uncertain" && !errors.Is(err, errUncertain) {
					t.Fatal(err)
				}
				want := "not attempted: pair:3, pair:4"
				if mode == "shared failure" {
					want = "not attempted: pair:4"
				}
				if !strings.Contains(out.String(), want) {
					t.Fatalf("missing untouched set %q:\n%s", want, out.String())
				}
			})
		})
	}
}

func TestFleetResumeScopesExecutableCancellation(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.recoveryErr = errors.New("private report unavailable")
	started, stopped := 0, 0
	d.resumeContext = func(ctx context.Context) (context.Context, context.CancelFunc) {
		started++
		return ctx, func() { stopped++ }
	}
	t.Chdir(cwd)
	root := buildRoot(d)
	root.SetContext(t.Context())
	root.SetArgs([]string{"fleet", "resume"})
	if err := root.Execute(); err == nil || started != 1 || stopped != 1 {
		t.Fatalf("context lifecycle on report refusal: err=%v start=%d stop=%d", err, started, stopped)
	}
}
