package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// #561: a two-car train whose cars boarded on affected records alone. The
// train's suite fails on the second car's change; the split holds it and
// lands the first. The measurement records one escape with its exact
// failure, both boardings resolved, and exactly the whole-suite executions
// that physically ran (initial, main baseline, one split probe), each with
// its saved attempt.
func TestLandingTrainMeasuresAnAffectedEscape(t *testing.T) {
	s := newTrainSetupBoarding(t, "train", affectedCheckName)
	writeTrainAttributionGraph(t, s.server)
	procedureWrite(t, s.server, "fixture.go", "package fixture\n")
	procedureWrite(t, s.server, "fixture_test.go", "package fixture\nimport \"testing\"\nfunc TestBroken(t *testing.T) {}\n")
	git(t, s.server, "add", ".")
	git(t, s.server, "commit", "-qm", "valid affected root package")
	failure := suiteFailing("example.com/train.TestBroken").Command[2]
	s.declareSuite(t, "if test -e second.go; then "+failure+"; fi; echo suite passed")
	bootstrapTrainAttributionTracker(t, s, nil)
	root := s.pushCar(t, "000002-second", "second", 2, 2)
	procedureWrite(t, root, "second.go", "package fixture\n")
	git(t, root, "add", "second.go")
	git(t, root, "commit", "-q", "--amend", "--no-edit")
	git(t, root, "push", "-q", "--force", "upstream", "000002-second")
	s.addCar(t, "000002-second", 2)

	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatalf("the clean prefix's car: %v", err)
	}
	wantSettled(t, two, landingline.OutcomeHeld)

	measures := lineEvents(t, s.line)[landingline.EventTrainMeasure]
	if len(measures) != 2 {
		t.Fatalf("measurement records %d, want start and final: %+v", len(measures), measures)
	}
	from, to := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	all, _, err := s.line.Events()
	if err != nil {
		t.Fatal(err)
	}
	stats := landingline.AggregateTrainMeasurements(all, from, to)
	if stats.Trains != 1 || stats.Measured != 1 || stats.Escaped != 1 || stats.ResolvedAffected != 2 || stats.UnresolvedAffected != 0 ||
		stats.EscapeFailures != 1 || stats.InitialRuns != 1 || stats.BaselineRuns != 1 || stats.SplitProbeRuns != 1 || stats.SplitTrains != 1 || stats.CostUnknown != 0 {
		t.Fatalf("stats %+v\n%s", stats, stats.Text)
	}
	final, err := decodeFinal(measures)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Failures) != 1 || final.Failures[0].Test != "example.com/train.TestBroken" || len(final.Incomplete) != 0 {
		t.Fatalf("final %+v", final)
	}
	// The measured executions are the physical ones: every saved attempt the
	// train made, by id, and nothing scheduled but not run.
	common := git(t, s.first, "rev-parse", "--git-common-dir")
	attempts, _, err := readSuiteAttempts(filepath.Join(common, "sdlc", "test-stats"), 100)
	if err != nil {
		t.Fatal(err)
	}
	var physical, measured []string
	for _, a := range attempts {
		if a.Check == suiteCheckName {
			physical = append(physical, a.ID)
		}
	}
	for _, x := range final.Executions {
		measured = append(measured, x.Attempt)
	}
	slices.Sort(physical)
	slices.Sort(measured)
	if !slices.Equal(physical, measured) {
		t.Fatalf("measured attempts %v, physical %v", measured, physical)
	}
	if !strings.Contains(stats.Text, "escapes 1/2 resolved affected boardings (50.0%)") {
		t.Fatalf("text %s", stats.Text)
	}
}

func decodeFinal(events []landingline.Event) (landingline.TrainMeasurement, error) {
	for _, e := range events {
		if e.Kind != landingline.EventTrainMeasure {
			continue
		}
		var m landingline.TrainMeasurement
		if err := json.Unmarshal(e.Measurement, &m); err != nil {
			return m, err
		}
		if m.Record == landingline.MeasureRecordFinal {
			return m, nil
		}
	}
	return landingline.TrainMeasurement{}, errors.New("no final measurement record")
}

// #561: a measurement append that fails warns, leaves the train's coverage
// incomplete, and returns nothing a landing could act on; the final record
// never claims completeness after a lost start.
func TestTrainMeasurementAppendFailure(t *testing.T) {
	dir := t.TempDir()
	line, err := landingline.Open(dir, landingline.Options{Poll: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	// The event log cannot be appended: its path is a directory.
	if err := os.MkdirAll(filepath.Join(dir, landingline.EventsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	c := &conductor{line: line, stderr: &stderr, id: 9, departed: time.Now()}
	c.measureDeparture([]landingline.Car{{Seq: 9, Head: strings.Repeat("a", 40)}, {Seq: 10, Head: strings.Repeat("b", 40)}})
	c.measureExecution(landingline.ExecInitial, localRunRecord{Exit: 0, AttemptID: "x", Finished: time.Now(), Duration: time.Minute}, nil)
	c.measure(landingline.MeasureFinal{At: time.Now(), Phase: "landed"})
	if !strings.Contains(stderr.String(), "start record not written") || !strings.Contains(stderr.String(), "final record not written") {
		t.Fatalf("stderr %q", stderr.String())
	}
	// Re-read what the tracker would have written: its snapshot is frozen and
	// incomplete.
	snap, err := landingline.AdvanceMeasurement(&c.meter, landingline.MeasureFinal{At: time.Now(), Phase: "other"})
	if err == nil || snap != nil {
		t.Fatalf("a finalized tracker accepted another final: %v %v", snap, err)
	}
}

// #561: an execution that gave no result (the suite could not start, or
// was cancelled) is measured as unknown with no cost; one with a result but
// no finish time has an unknown cost, never zero.
func TestMeasureExecutionUnknowns(t *testing.T) {
	dir := t.TempDir()
	line, err := landingline.Open(dir, landingline.Options{Poll: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	c := &conductor{line: line, stderr: &strings.Builder{}, id: 9, departed: time.Now()}
	c.measureDeparture([]landingline.Car{{Seq: 9, Head: strings.Repeat("a", 40)}})
	c.measureExecution(landingline.ExecInitial, localRunRecord{Exit: -1}, errors.New("could not start"))
	c.measureExecution(landingline.ExecBaseline, localRunRecord{Exit: 1, AttemptID: "b", Duration: time.Minute}, nil)
	c.measure(landingline.MeasureFinal{At: time.Now(), Phase: "aborted"})
	events, _, err := line.Events()
	if err != nil {
		t.Fatal(err)
	}
	final, err := decodeFinal(events)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Executions) != 2 || final.Executions[0].Outcome != landingline.ExecUnknown || final.Executions[0].RunSeconds != nil ||
		final.Executions[1].Outcome != landingline.ExecFailed || final.Executions[1].RunSeconds != nil || len(final.Incomplete) == 0 {
		t.Fatalf("final %+v", final)
	}
	stats := landingline.AggregateTrainMeasurements(events, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if stats.UnknownRuns != 1 || stats.CostUnknown != 1 || stats.Measured != 0 || stats.Incomplete != 1 {
		t.Fatalf("stats %+v", stats)
	}
}
