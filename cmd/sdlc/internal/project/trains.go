package project

import "github.com/xianxu/ariadne/pkg/trains"

// The trains parser lives in pkg/trains (#350), so tl-lead renders the same
// trains the forecast and the resolver read. These names forward to it.
type (
	IssueRef  = trains.IssueRef
	Train     = trains.Train
	Edge      = trains.Edge
	TrainPlan = trains.TrainPlan
)

var ParseTrains = trains.ParseTrains

// Trains parses this doc's `## Trains` section.
func (d *Doc) Trains(defaultRepo string) TrainPlan {
	return ParseTrains(d.SectionBody("Trains"), defaultRepo)
}
