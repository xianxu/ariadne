package main

import (
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Publication policy is chosen once for this compile, never retained in inventory.
type compilePublicationPolicy uint8

const (
	publishAllOutputs compilePublicationPolicy = iota
	deferProtectedSettings
	deferPlanningOutputs
)

type deferredSeed struct {
	path   string
	action plan.Action
}

type seedDeferralPlan struct {
	actions []plan.Action
	seeds   []deferredSeed
}

// planDeferredSeeds suppresses the whole seed class, independent of Git tracking.
// Deferring before publication avoids both rollback and index-observation races.
func planDeferredSeeds(actions []plan.Action) seedDeferralPlan {
	result := seedDeferralPlan{actions: make([]plan.Action, 0, len(actions))}
	for _, action := range actions {
		var dst string
		switch a := action.(type) {
		case plan.Seed:
			dst = a.Dst
		case plan.SeedOnce:
			dst = a.Dst
		default:
			result.actions = append(result.actions, action)
			continue
		}
		result.actions = append(result.actions, plan.DeferredOutput{Path: dst})
		result.seeds = append(result.seeds, deferredSeed{path: dst, action: action})
	}
	return result
}

// deferredSeedOutputs only observes destinations. DeferredOutput keeps any prior
// managed evidence intact; seeds themselves remain consumer-owned, not inventoried.
func deferredSeedOutputs(fs weavefs.FS, root string, actions []plan.Action, out io.Writer) ([]plan.Action, error) {
	deferred := planDeferredSeeds(actions)
	reported := map[string]bool{}
	for _, seed := range deferred.seeds {
		current, err := plan.OutputCurrent(fs, root, seed.action)
		if err != nil {
			return nil, fmt.Errorf("inspect deferred seed %s: %w", seed.path, err)
		}
		if !current && !reported[seed.path] {
			if _, err := fmt.Fprintf(out, "weave: seeded file %s is behind ariadne; deferred during planning. Run weave compile on a dedicated consumer issue branch and publish regeneration as its own change; local-gates files require operator verification.\n", seed.path); err != nil {
				return nil, err
			}
			reported[seed.path] = true
		}
	}
	return deferred.actions, nil
}
