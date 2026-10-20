package main

import (
	"bytes"
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
	resumeTimeout     = 3 * time.Minute
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

func (c execResume) Resume(addr string) error {
	limit := orDuration(c.resume, resumeTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.bin, "--resume", addr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	switch {
	case ctx.Err() != nil:
		return fmt.Errorf("%w: no answer within %s", errUncertain, limit)
	case err != nil:
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return fmt.Errorf("did not run: %v", err) // couch missing, say: nothing was attempted
		}
		return fmt.Errorf("refused (%v): %s", err, orNone(strings.TrimSpace(stderr.String())))
	}
	return nil
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
)

var classNames = map[slotClass]string{classRunning: "running", classIdle: "idle", classResume: "resume", classAmbiguous: "ambiguous", classUnknown: "unknown"}

type resumeRow struct {
	Addr  string
	Class slotClass
	Why   string
}

// resumeInputs is everything planResume reads, already fetched. Each
// slot's claims come from the inventory (FleetSlot.Claims).
type resumeInputs struct {
	Self      string
	Hubs      []hub
	Slots     []FleetSlot
	Actors    map[string]Actor
	ActorsErr error
}

// planResume classifies every slot that resolves to this hub alone. A tied
// or unclaimed slot is ambiguous, since the order that breaks it is Couch's
// command line; a failed read makes the affected slots unknown. Slots of
// other hubs are counted, not listed. Down slots with work come back as
// classCandidates: finishResume decides them once their candidates are read.
func planResume(in resumeInputs) (rows []resumeRow, others int) {
	for _, sl := range in.Slots {
		addr := sl.Address
		if sl.Repo == in.Self && sl.Slot == 0 {
			continue // this hub itself
		}
		own := ownerOf(sl.Repo, sl.Slot, in.Hubs)
		if who, ok := own.sole(); !ok {
			why := "unclaimed: Couch gives it to the first hub on its command line"
			if !own.Unclaimed {
				why = "claimed by " + strings.Join(own.Owners, " and ") + ": the first on Couch's command line owns it"
			}
			if own.Unclaimed || containsStr(own.Owners, in.Self) {
				rows = append(rows, resumeRow{addr, classAmbiguous, why})
			} else {
				others++
			}
			continue
		} else if who != in.Self {
			others++
			continue
		}
		if in.ActorsErr != nil {
			rows = append(rows, resumeRow{addr, classUnknown, "couch --actors unreadable: " + in.ActorsErr.Error()})
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
		work := sl.Claims
		if len(work) == 0 {
			rows = append(rows, resumeRow{addr, classIdle, ""})
			continue
		}
		rows = append(rows, resumeRow{addr, classCandidates, strings.Join(work, ", ")})
	}
	return rows, others
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
// one slot at a time, reporting each outcome. A refusal never stops the rest;
// an uncertain outcome is reported, never retried: peek, then rerun (a slot
// already resumed is running and skipped).
func fleetResume(out io.Writer, d deps, cwd string, apply bool) error {
	self, workspace, err := hereHub(cwd)
	if err != nil {
		return err
	}
	hubs, err := readHubs(workspace)
	if err != nil {
		return err
	}
	fleet, err := d.sdlc.FleetInventory(cwd)
	if err != nil {
		return err
	}
	in := resumeInputs{Self: self, Hubs: hubs, Slots: fleet.Slots}
	in.Actors, in.ActorsErr = d.actors.Actors()
	rows, others := planResume(in)
	for i, r := range rows {
		if r.Class == classCandidates {
			c, err := d.resume.Candidates(r.Addr)
			rows[i] = finishResume(r, c, err)
		}
	}
	fmt.Fprintln(out, invisibleLayers)
	for _, cl := range []slotClass{classResume, classAmbiguous, classUnknown, classRunning, classIdle} {
		for _, r := range rows {
			if r.Class == cl {
				fmt.Fprintf(out, "%-9s %s  %s\n", classNames[cl], r.Addr, r.Why)
			}
		}
	}
	if others > 0 {
		fmt.Fprintf(out, "(%d slot(s) belong to other hubs)\n", others)
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
	for _, addr := range todo {
		switch err := d.resume.Resume(addr); {
		case err == nil:
			fmt.Fprintf(out, "resumed   %s\n", addr)
		case errors.Is(err, errUncertain):
			fmt.Fprintf(out, "uncertain %s: %v; peek the slot before rerunning\n", addr, err)
		default:
			fmt.Fprintf(out, "refused   %s: %v\n", addr, err)
		}
	}
	return nil
}

func newFleetCmd(d deps) *cobra.Command {
	fleet := &cobra.Command{
		Use:   "fleet",
		Short: "The slots this hub owns",
	}
	var apply bool
	resume := &cobra.Command{
		Use:   "resume [--apply]",
		Short: "Resume this hub's down slots that have work in flight (only when the operator asks; never at startup)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return fleetResume(cmd.OutOrStdout(), d, cwd, apply)
		},
	}
	resume.Flags().BoolVar(&apply, "apply", false, "resume the resume class, one slot at a time (default: a dry run)")
	fleet.AddCommand(resume)
	return fleet
}
