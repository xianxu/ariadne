package main

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The train's measurement effects (#561). The conductor supplies facts to
// the pure tracker (landingline.AdvanceMeasurement) and appends the records
// it returns to the line's event log. Measurement is evidence only: a failed
// append warns and leaves coverage incomplete, and nothing here changes a
// landing decision.

// measure advances the tracker and performs the append it asks for.
func (c *conductor) measure(e landingline.MeasurementEvent) {
	rec, err := landingline.AdvanceMeasurement(&c.meter, e)
	if err != nil {
		fmt.Fprintf(c.stderr, "landing train %d: measurement: %v\n", c.id, err)
	}
	if rec == nil {
		return
	}
	appendErr := c.appendMeasurement(*rec)
	if rec.Record == landingline.MeasureRecordStart {
		c.measure(landingline.MeasureStartAppended{Err: appendErr})
		return
	}
	if appendErr != nil {
		fmt.Fprintf(c.stderr, "landing train %d: measurement final record not written (%v); its start reads as incomplete\n", c.id, appendErr)
	}
}

func (c *conductor) appendMeasurement(m landingline.TrainMeasurement) error {
	raw, err := landingline.EncodeMeasurement(m)
	if err != nil {
		return err
	}
	err = c.line.Record(landingline.Event{Kind: landingline.EventTrainMeasure, Seq: c.id, Train: c.id, Measurement: raw})
	if err != nil && m.Record == landingline.MeasureRecordStart {
		fmt.Fprintf(c.stderr, "landing train %d: measurement start record not written (%v); this train's coverage is incomplete\n", c.id, err)
	}
	return err
}

// measureDeparture fixes the original cars and the proofs they boarded on.
func (c *conductor) measureDeparture(cars []landingline.Car) {
	measured := make([]landingline.MeasuredCar, len(cars))
	for i, car := range cars {
		measured[i] = landingline.MeasuredCar{Seq: car.Seq, Issue: car.Issue, Head: car.Head, Proof: landingline.ProofUnknown}
		if p, ok := c.proofs[car.Seq]; ok && p.Head == car.Head {
			measured[i].Proof = p.Check
		}
	}
	c.measure(landingline.MeasureDeparture{Train: c.id, Departed: c.departed, Cars: measured})
}

// measureTransition feeds the reducer's confirmed facts: a passing suite
// covers its prefix's cars, and an attributable split that the reducer
// accepted names its culprit with the exact qualifying failures.
func (c *conductor) measureTransition(before, after landingline.Train, ev landingline.TrainEvent) {
	switch ev := ev.(type) {
	case landingline.SuiteDone:
		if ev.Outcome == landingline.SuitePass && ev.Prefix >= 1 && ev.Prefix <= len(before.Cars) {
			c.measure(landingline.MeasurePrefixPassed{Seqs: carSeqsOf(before.Cars[:ev.Prefix])})
		}
	case landingline.SplitAssessed:
		if ev.Verdict.Kind == landingline.AttributionAttributable && after.Culprit != 0 {
			c.measure(landingline.MeasureCulprit{Seq: after.Culprit, Failures: ev.Verdict.Escapes})
		}
	}
}

// measureExecution records one managed whole-suite execution from its
// returned evidence. err means it gave no result: no cost, coverage
// incomplete.
func (c *conductor) measureExecution(kind landingline.ExecutionKind, rec localRunRecord, err error) {
	x := landingline.ExecutionSummary{Kind: kind, Outcome: landingline.ExecUnknown, Attempt: rec.AttemptID}
	if err == nil && rec.Exit != -1 {
		x.Outcome = landingline.ExecFailed
		if rec.passed() {
			x.Outcome = landingline.ExecPassed
		}
		if !rec.Finished.IsZero() {
			run, queue := rec.Duration.Seconds(), rec.QueueWait.Seconds()
			x.RunSeconds, x.QueueSeconds = &run, &queue
		}
	}
	c.measure(landingline.MeasureExecution{Summary: x, Finished: rec.Finished})
}

func carSeqsOf(cars []landingline.Car) []int64 {
	seqs := make([]int64, len(cars))
	for i, car := range cars {
		seqs[i] = car.Seq
	}
	return seqs
}

// boardedProofs are the proofs bound to the vouches boarding read: the
// conductor's own fresh readiness and each waiter's sidecar (#561). An
// unbound or missing proof is left out and measures as unknown.
func boardedProofs(self int64, own landingline.Readiness, ready map[int64]landingline.Readiness) map[int64]landingline.BoardingProof {
	out := map[int64]landingline.BoardingProof{}
	for seq, r := range ready {
		if p, ok := r.AcceptedProof(); ok {
			out[seq] = p
		}
	}
	delete(out, self) // its own fresh readiness, not its last vouch, is what it rides on
	if p, ok := ownProof(own); ok {
		out[self] = p
	}
	return out
}

// ownProof is the conductor's own readiness proof. That readiness was read
// fresh, never written by Vouch, so its stamp is its own observation time.
func ownProof(r landingline.Readiness) (landingline.BoardingProof, bool) {
	if r.Proof != nil {
		p := *r.Proof
		p.Stamp = r.At
		r.Proof = &p
	}
	return r.AcceptedProof()
}
