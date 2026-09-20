package acquire

import "github.com/xianxu/ariadne/cmd/weave/internal/staging"

// Acquisition and generation share the same durable owner/recovery rule.
type stageOwner = staging.Owner

func newStage(destination string) (string, error) { return staging.New(destination) }
func reclaimStages(destination string) error      { return staging.Reclaim(destination) }
