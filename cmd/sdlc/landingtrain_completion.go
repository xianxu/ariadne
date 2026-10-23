package main

import (
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// Keep the conductor's authoritative train until its completion is committed.
func (c *conductor) commitTrainCompletion(tr *landingline.Train, next landingline.Train, cars []landingline.Car, stop func()) error {
	var err error
	if next.Phase == landingline.PhaseMerging && !tr.MergeStarted {
		_, err = c.line.StartTrainMerge(c.handle, next)
	} else {
		_, err = c.line.UpdateTrain(c.handle, next)
	}
	if errors.Is(err, landingline.ErrTrainDropPending) {
		return c.applyTrainDrops(tr, cars, stop)
	}
	if err != nil {
		return err
	}
	*tr = next
	return nil
}

func (c *conductor) abortTrain(tr *landingline.Train, cars []landingline.Car, stop func(), cause error) error {
	why := "the train stopped: " + firstLine(cause.Error())
	next, effects, terr := landingline.Transition(*tr, landingline.Aborted{Reason: why})
	if terr != nil {
		return errors.Join(cause, terr)
	}
	// Abort is also a completion: commit before releasing cars or closing
	// their PR, and let a drop accepted first use the unchanged riding set.
	if err := c.commitTrainCompletion(tr, next, cars, stop); err != nil {
		return errors.Join(cause, err)
	}
	for _, e := range effects {
		if rel, ok := e.(landingline.Release); ok {
			for _, s := range rel.Seqs {
				c.settle(carBySeq(cars, s), landingline.OutcomeReleased, why)
			}
		} else {
			_, _ = c.execute(e, cars, "")
		}
	}
	// Cars the machine no longer rides (none, normally) still get a word.
	for _, car := range cars {
		c.settle(car, landingline.OutcomeReleased, why)
	}
	return cause
}
