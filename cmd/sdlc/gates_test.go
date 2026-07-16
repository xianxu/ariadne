package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// Drift guard (#172): the gate signature catalog must stay in lockstep with the
// bypass flags each spine command actually registers. If someone adds a new
// `--no-<gate>` (or removes one) without updating the catalog, this fails — so the
// friction audit can never silently miss a gate.
//
// Keyed by (command, flag) pairs. `no-start` (claim, a workflow toggle) and the
// blanket `--force` are NOT gates and are excluded.
func TestGateCatalogMatchesRegisteredFlags(t *testing.T) {
	spine := map[string]*cobra.Command{
		"close":           NewCloseCmd(),
		"milestone-close": NewMilestoneCloseCmd(),
		"change-code":     NewChangeCodeCmd(),
		"merge":           NewMergeCmd(),
		"push":            NewPushCmd(),
		"project close":   newProjectCloseCmd(),
	}
	for cmdName, cmd := range spine {
		registered := registeredGateFlags(cmd)
		cataloged := processmanual.GateFlagsFor(cmdName)
		sort.Strings(registered)
		sort.Strings(cataloged)
		if strings.Join(registered, ",") != strings.Join(cataloged, ",") {
			t.Errorf("%s: registered gate flags %v != cataloged %v", cmdName, registered, cataloged)
		}
	}
}

// registeredGateFlags returns a command's `--no-*` flags that are bypass gates —
// excluding the non-gate `no-start` toggle and the blanket `--force`.
func registeredGateFlags(cmd *cobra.Command) []string {
	var out []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !strings.HasPrefix(f.Name, "no-") || f.Name == "no-start" {
			return
		}
		out = append(out, f.Name)
	})
	return out
}
