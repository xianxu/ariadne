package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/internal/jsonstrict"
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
	out, err := docRepositoryOutput(ctx, "", c.bin, "--resume-candidates", addr, "--json")
	if err != nil {
		if ctx.Err() != nil {
			return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates %s: no answer within %s: %w", addr, limit, err)
		}
		return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates %s: %w", addr, err)
	}
	return decodeCandidates(out)
}

func decodeCandidates(raw []byte) (ResumeCandidates, error) {
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates: unreadable reply: %w", err)
	}
	var in struct {
		Proven *struct {
			NativeID string `json:"native_id"`
		}
		Incomplete *bool
		Reasons    []string
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return ResumeCandidates{}, fmt.Errorf("couch --resume-candidates: unreadable reply: %v", err)
	}
	incomplete := in.Incomplete == nil || *in.Incomplete
	if in.Incomplete == nil {
		in.Reasons = append(in.Reasons, "missing completeness evidence")
	}
	return ResumeCandidates{Proven: in.Proven != nil && strings.TrimSpace(in.Proven.NativeID) != "", Incomplete: incomplete, Reasons: in.Reasons}, nil
}

// slotClass is what fleet resume makes of one slot.
type slotClass int

const (
	classRunning    slotClass = iota // a live agent: left alone
	classSkipped                     // no positive asleep evidence: no mutation
	classResume                      // explicitly asleep, complete proven conversation
	classAmbiguous                   // needs the operator: never guessed
	classUnknown                     // a read failed: never treated as down
	classCandidates                  // explicitly asleep: candidates decide (planner-internal)
	classExcluded                    // visible with --all, never a mutation candidate
)

var classNames = map[slotClass]string{classRunning: "running", classSkipped: "skipped", classResume: "resume", classAmbiguous: "ambiguous", classUnknown: "unknown", classExcluded: "excluded"}

type resumeRow struct {
	Addr    string
	Class   slotClass
	Why     string
	Pending bool
}

// resumeInputs is everything planResume reads, already fetched. Each
// slot's asleep and pending evidence comes from one Couch recovery report.
type resumeInputs struct {
	Scope     hubScope
	All       bool
	Slots     []RecoverySlot
	Actors    map[string]Actor
	ActorsErr error
}

// planResume classifies every slot that resolves to this hub alone. A tied
// or unclaimed slot is ambiguous, since the order that breaks it is Couch's
// command line; a failed read makes the affected slots unknown. Slots of
// other hubs are counted, not listed. Explicitly asleep workers come back as
// classCandidates: finishResume decides them once their candidates are read.
func planResume(in resumeInputs) (rows []resumeRow, others, unclaimed int) {
	for _, evidence := range in.Slots {
		sl := evidence.FleetSlot
		addr := sl.Address
		if sl.Slot == 0 {
			if in.All {
				rows = append(rows, resumeRow{addr, classExcluded, "hub slot", false})
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
					rows = append(rows, resumeRow{addr, classExcluded, why, false})
				}
			case containsStr(own.Owners, in.Scope.self):
				rows = append(rows, resumeRow{addr, classAmbiguous, why, false})
			default:
				others++
				if in.All {
					rows = append(rows, resumeRow{addr, classExcluded, why, false})
				}
			}
			continue
		} else if who != in.Scope.self {
			others++
			if in.All {
				rows = append(rows, resumeRow{addr, classExcluded, "belongs to " + who, false})
			}
			continue
		}
		if in.ActorsErr != nil {
			rows = append(rows, resumeRow{addr, classUnknown, "", false})
			continue
		}
		if _, live := in.Actors[addr]; live {
			rows = append(rows, resumeRow{addr, classRunning, "", false})
			continue
		}
		if evidence.Unknown != "" {
			rows = append(rows, resumeRow{addr, classUnknown, evidence.Unknown, false})
			continue
		}
		if !evidence.Asleep {
			rows = append(rows, resumeRow{addr, classSkipped, "no explicit asleep mark (not asleep or older Couch)", false})
			continue
		}
		why := strings.Join(sl.Claims, ", ")
		if why == "" {
			why = "explicitly asleep"
		}
		rows = append(rows, resumeRow{addr, classCandidates, why, evidence.Pending})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Addr < rows[j].Addr })
	return rows, others, unclaimed
}

// finishResume decides a down slot with work from its candidates read.
func finishResume(r resumeRow, c ResumeCandidates, err error) resumeRow {
	switch {
	case err != nil:
		r.Class, r.Why = classUnknown, r.Why+"; candidates unreadable: "+err.Error()
	case c.Incomplete:
		r.Class, r.Why = classUnknown, r.Why+"; candidates incomplete: "+strings.Join(c.Reasons, "; ")
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

// fleetResume captures one report and plans worker waves. The bounded batch
// owns admission and stop decisions; uncertainty never permits an automatic retry.
type resumeOptions struct {
	All         bool
	Concurrency int
}

func fleetResume(ctx context.Context, out io.Writer, d deps, cwd string, apply bool, options ...resumeOptions) error {
	opt := resumeOptions{Concurrency: 3}
	if len(options) > 0 {
		opt = options[0]
	}
	if opt.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be a positive integer")
	}
	all := opt.All
	scope, err := readHubScope(cwd)
	if err != nil {
		return err
	}
	if !scope.enabled {
		return fmt.Errorf("no .hub at repository root: run from a team's ops repo")
	}
	if d.recovery == nil {
		return fmt.Errorf("Couch recovery provider unavailable; inspect couch --recover-plan-from-sdlc")
	}
	fleet, err := d.recovery.Recovery(ctx, cwd)
	if err != nil {
		return err
	}
	in := resumeInputs{Scope: scope, All: all, Slots: fleet.Slots}
	in.Actors, in.ActorsErr = d.actors.Actors()
	rows, others, unclaimed := planResume(in)
	for i, r := range rows {
		if r.Class == classCandidates {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("resume planning canceled before dispatch: %w", err)
			}
			c, err := d.resume.Candidates(r.Addr)
			if canceled := ctx.Err(); canceled != nil {
				return fmt.Errorf("resume planning canceled before dispatch: %w", canceled)
			}
			rows[i] = finishResume(r, c, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("resume planning canceled before dispatch: %w", err)
	}
	fmt.Fprintln(out, invisibleLayers)
	if in.ActorsErr != nil {
		fmt.Fprintf(out, "couch --actors unreadable: %v; run tl-lead as a bare call if sandbox exclusions apply\n", in.ActorsErr)
	}
	for _, cl := range []slotClass{classResume, classAmbiguous, classUnknown, classRunning, classSkipped, classExcluded} {
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
		batch, err := newResumeBatch(rows, opt.Concurrency)
		if err != nil {
			return err
		}
		var waves [2][]string
		for _, launch := range batch.plan() {
			waves[launch.Wave-1] = append(waves[launch.Wave-1], launch.Addr)
		}
		fmt.Fprintf(out, "wave 1 (pending work): %s\n", orNone(strings.Join(waves[0], ", ")))
		fmt.Fprintf(out, "wave 2 (remaining asleep): %s\n", orNone(strings.Join(waves[1], ", ")))
		fmt.Fprintf(out, "dry run: --apply resumes explicitly asleep workers, pending wave before remaining wave; concurrency %d per invocation\n", opt.Concurrency)
		return nil
	}
	return runResumeBatch(ctx, out, d, scope.self, rows, opt.Concurrency)
}

func newFleetCmd(d deps) *cobra.Command {
	fleet := &cobra.Command{
		Use:   "fleet",
		Short: "The slots this hub owns",
	}
	var apply, all bool
	var concurrency int
	resume := &cobra.Command{
		Use:   "resume [--apply] [--all] [--concurrency N]",
		Short: "Resume this hub's explicitly asleep workers (only when the operator asks; never at startup)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return fleetResume(cmd.Context(), cmd.OutOrStdout(), d, cwd, apply, resumeOptions{All: all, Concurrency: concurrency})
		},
	}
	resume.Flags().BoolVar(&apply, "apply", false, "resume explicitly asleep workers in two waves (default: a dry run)")
	resume.Flags().BoolVar(&all, "all", false, "show every slot; never expand automatic resume authority")
	resume.Flags().IntVar(&concurrency, "concurrency", 3, "maximum simultaneous worker resumes per invocation; must be positive")
	fleet.AddCommand(resume)
	return fleet
}
