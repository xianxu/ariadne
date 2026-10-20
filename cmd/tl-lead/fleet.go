package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// CouchResume is the seam over Couch's per-slot resume verbs (couch
// --skill): `--resume-candidates repo:N --json` (seconds; read-only) and
// `--resume repo:N` (heavy; one at a time). Tests use a stateful fake.
type CouchResume interface {
	Candidates(addr string) (ResumeCandidates, error)
	Resume(addr string) error
}

// ResumeCandidates is the part of a --resume-candidates reply tl-lead reads.
type ResumeCandidates struct {
	Proven     bool // a proven conversation: plain --resume resumes it
	Incomplete bool
	Reasons    []string
}

// errUncertain marks a resume whose outcome Couch did not confirm.
var errUncertain = errors.New("outcome uncertain")

type execResume struct {
	bin string
	// candidates and resume bound each call; zero means the defaults below.
	candidates, resume time.Duration
}

func orDuration(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

// Couch's cost classes: --resume-candidates takes seconds, --resume is heavy.
const (
	candidatesTimeout = time.Minute
	// Couch polls receipts for five minutes; allow admission and process teardown.
	resumeTimeout = 6 * time.Minute
)

func (c execResume) Candidates(addr string) (ResumeCandidates, error) {
	limit := orDuration(c.candidates, candidatesTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	out, err := exec.CommandContext(ctx, c.bin, "--resume-candidates", addr, "--json").Output()
	if err != nil {
		var ee *exec.ExitError
		if ctx.Err() != nil {
			return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates %s: no answer within %s", addr, limit)
		}
		if errors.As(err, &ee) {
			return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates %s: %s", addr, strings.TrimSpace(string(ee.Stderr)))
		}
		return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates %s: %v", addr, err)
	}
	return decodeCandidates(out)
}

func decodeCandidates(raw []byte) (ResumeCandidates, error) {
	var in struct {
		Proven *struct {
			NativeID string `json:"native_id"`
		}
		Incomplete bool
		Reasons    []string
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates: unreadable reply: %v", err)
	}
	return ResumeCandidates{Proven: in.Proven != nil && in.Proven.NativeID != "", Incomplete: in.Incomplete, Reasons: in.Reasons}, nil
}

// slotClass is what fleet resume makes of one slot.
type slotClass int

const (
	classRunning    slotClass = iota // a live agent: left alone
	classIdle                        // down, no work in flight: left down
	classResume                      // down, work in flight, a proven conversation
	classAmbiguous                   // needs the operator: never guessed
	classUnknown                     // a read failed: never treated as down
	classCandidates                  // down with work: its candidates decide (planner-internal)
	classExcluded                    // visible with --all, never a mutation candidate
)

var classNames = map[slotClass]string{classRunning: "running", classIdle: "idle", classResume: "resume", classAmbiguous: "ambiguous", classUnknown: "unknown", classExcluded: "excluded"}

type resumeRow struct {
	Addr  string
	Class slotClass
	Why   string
}

// resumeInputs is everything planResume reads, already fetched. Each
// slot's claims come from the inventory (FleetSlot.Claims).
type resumeInputs struct {
	Scope     hubScope
	All       bool
	Slots     []FleetSlot
	Actors    map[string]Actor
	ActorsErr error
}

// planResume classifies every slot that resolves to this hub alone. A tied
// or unclaimed slot is ambiguous, since the order that breaks it is Couch's
// command line; a failed read makes the affected slots unknown. Slots of
// other hubs are counted, not listed. Down slots with work come back as
// classCandidates: finishResume decides them once their candidates are read.
func planResume(in resumeInputs) (rows []resumeRow, others, unclaimed int) {
	for _, sl := range in.Slots {
		addr := sl.Address
		if sl.Repo == in.Scope.self && sl.Slot == 0 {
			if in.All {
				rows = append(rows, resumeRow{addr, classExcluded, "this hub"})
			}
			continue
		}
		own := in.Scope.owner(sl)
		if who, sole := own.sole(); !sole {
			why := "unclaimed: Couch gives it to the first hub on its command line"
			if !own.Unclaimed {
				why = "claimed by " + strings.Join(own.Owners, " and ") + ": the first on Couch's command line owns it"
			}
			switch {
			case own.Unclaimed:
				unclaimed++
				if in.All {
					rows = append(rows, resumeRow{addr, classAmbiguous, why})
				}
			case containsStr(own.Owners, in.Scope.self):
				rows = append(rows, resumeRow{addr, classAmbiguous, why})
			default:
				others++
				if in.All {
					rows = append(rows, resumeRow{addr, classExcluded, why})
				}
			}
			continue
		} else if who != in.Scope.self {
			others++
			if in.All {
				rows = append(rows, resumeRow{addr, classExcluded, "belongs to " + who})
			}
			continue
		}
		if in.ActorsErr != nil {
			rows = append(rows, resumeRow{addr, classUnknown, ""})
			continue
		}
		if _, live := in.Actors[addr]; live {
			rows = append(rows, resumeRow{addr, classRunning, ""})
			continue
		}
		if !sl.ClaimsKnown {
			rows = append(rows, resumeRow{addr, classUnknown, "claims unreadable: " + sl.ClaimsWhy})
			continue
		}
		if len(sl.Claims) == 0 {
			rows = append(rows, resumeRow{addr, classIdle, ""})
			continue
		}
		rows = append(rows, resumeRow{addr, classCandidates, strings.Join(sl.Claims, ", ")})
	}
	return rows, others, unclaimed
}

// finishResume decides a down slot with work from its candidates read.
func finishResume(r resumeRow, c ResumeCandidates, err error) resumeRow {
	switch {
	case err != nil:
		r.Class, r.Why = classUnknown, r.Why+"; candidates unreadable: "+err.Error()
	case c.Proven:
		r.Class = classResume
	default:
		r.Class, r.Why = classAmbiguous, r.Why+"; no proven conversation (choose: couch --resume "+r.Addr+" --candidate KEY, or --reboot)"
	}
	return r
}

func containsStr(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// invisibleLayers is said on every run: what Couch knows that tl-lead can't.
const invisibleLayers = "note: Couch runtime moves (couch --move-slot since Couch started) are not visible here; a moved slot is classed by .hub"

// fleetResume reads, plans, prints, and with apply resumes the resume class
// one slot at a time. Every unsuccessful outcome stops the batch; uncertainty
// permits only read-only inspection, never an automatic mutation retry.
func fleetResume(out io.Writer, d deps, cwd string, apply bool, allFlags ...bool) error {
	all := len(allFlags) > 0 && allFlags[0]
	scope, err := readHubScope(cwd)
	if err != nil {
		return err
	}
	if !scope.enabled {
		return fmt.Errorf("no .hub at repository root: run from a team's ops repo")
	}
	fleet, err := d.sdlc.FleetInventory(cwd)
	if err != nil {
		return err
	}
	in := resumeInputs{Scope: scope, All: all, Slots: fleet.Slots}
	in.Actors, in.ActorsErr = d.actors.Actors()
	rows, others, unclaimed := planResume(in)
	for i, r := range rows {
		if r.Class == classCandidates {
			c, err := d.resume.Candidates(r.Addr)
			rows[i] = finishResume(r, c, err)
		}
	}
	fmt.Fprintln(out, invisibleLayers)
	if in.ActorsErr != nil {
		fmt.Fprintf(out, "couch --actors unreadable: %v; run tl-lead as a bare call if sandbox exclusions apply\n", in.ActorsErr)
	}
	for _, cl := range []slotClass{classResume, classAmbiguous, classUnknown, classRunning, classIdle, classExcluded} {
		for _, r := range rows {
			if r.Class == cl {
				fmt.Fprintf(out, "%-9s %s  %s\n", classNames[cl], r.Addr, r.Why)
			}
		}
	}
	if others > 0 {
		fmt.Fprintf(out, "(%d slot(s) belong to other hubs)\n", others)
	}
	if unclaimed > 0 && !all {
		fmt.Fprintf(out, "(%d unclaimed slot(s); --all to list)\n", unclaimed)
	}
	if !apply {
		fmt.Fprintln(out, "dry run: --apply resumes the resume class, one slot at a time")
		return nil
	}
	var todo []string
	for _, r := range rows {
		if r.Class == classResume {
			todo = append(todo, r.Addr)
		}
	}
	fmt.Fprintf(out, "resuming %s\n", orNone(strings.Join(todo, ", ")))
	for i, addr := range todo {
		fmt.Fprintf(out, "resuming  %s; waiting for Couch's terminal receipt\n", addr)
		err := d.resume.Resume(addr)
		if err == nil {
			fmt.Fprintf(out, "resumed   %s\n", addr)
			continue
		}
		if errors.Is(err, errUncertain) {
			fmt.Fprintf(out, "uncertain %s: %v\n", addr, err)
			actors, aerr := readActors(d.actors)
			switch {
			case aerr != nil:
				fmt.Fprintf(out, "observation unavailable: %v\n", aerr)
			case actors[addr].Agent != "":
				fmt.Fprintf(out, "observed %s running; operation outcome remains uncertain\n", addr)
			default:
				fmt.Fprintf(out, "observed no actor at %s; operation outcome remains uncertain\n", addr)
			}
			fmt.Fprintf(out, "inspect the request with couch --operation-status ID (from the diagnostics) and couch --peek %s before a manual resume or switcher action\n", addr)
		} else {
			fmt.Fprintf(out, "refused   %s: %v\n", addr, err)
		}
		fmt.Fprintf(out, "stopped at %s; not attempted: %s\n", addr, orNone(strings.Join(todo[i+1:], ", ")))
		return fmt.Errorf("fleet resume stopped at %s: %w", addr, err)
	}

	return nil
}

func newFleetCmd(d deps) *cobra.Command {
	fleet := &cobra.Command{
		Use:   "fleet",
		Short: "The slots this hub owns",
	}
	var apply, all bool
	resume := &cobra.Command{
		Use:   "resume [--apply] [--all]",
		Short: "Resume this hub's down slots that have work in flight (only when the operator asks; never at startup)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return fleetResume(cmd.OutOrStdout(), d, cwd, apply, all)
		},
	}
	resume.Flags().BoolVar(&apply, "apply", false, "resume the resume class, one slot at a time (default: a dry run)")
	resume.Flags().BoolVar(&all, "all", false, "show every slot; never expand automatic resume authority")
	fleet.AddCommand(resume)
	return fleet
}
