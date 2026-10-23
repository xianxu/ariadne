package landingline

import (
	"fmt"
	"time"
)

// TrainRampReviewWindow is how much measured train history a cap increase
// waits for (#561): a complete week of the repository's own trains.
const TrainRampReviewWindow = 7 * 24 * time.Hour

// SerialLandingPrerequisite is how a repository requires a full suite on
// every landing again: there is no per-call serial flag, only main's guarded
// train setting (#561).
const SerialLandingPrerequisite = "Serial landing (one full suite per landing, no train) needs the TL's guarded main change setting train_interval_minutes to 0 in " + ConfigPath + " first."

// ConfigPath is the train setting's file on main (#449), under the
// local-gates guard.
const ConfigPath = ".sdlc/landing.json"

// FormatTrainRampPolicy is the one statement of how a train cap ramps (#561):
// help and status both print it, so no repository improvises its own rule.
// It never enables trains or raises a cap; the guarded main configuration is
// the only authority.
func FormatTrainRampPolicy() string {
	days := int(TrainRampReviewWindow / (24 * time.Hour))
	return fmt.Sprintf("Train cap ramp: a repo without train_max_cars boards %d cars per train. "+
		"Raise the cap only by hand, through a guarded change to %s on main, after reviewing a complete %d-day window "+
		"of its measurements: escapes over resolved affected boardings (the denominator), unresolved and unmeasured coverage, and split costs. "+
		"Nothing raises a cap or enables trains automatically. %s",
		DefaultMaxCars, ConfigPath, days, SerialLandingPrerequisite)
}
