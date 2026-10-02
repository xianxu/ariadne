package observe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// boundaryRE is a review boundary: the plan, a milestone (as the plan names it) or the close.
var boundaryRE = regexp.MustCompile(`^(plan|close|` + issue.MilestoneTagPattern + `)$`)

// Validate enforces the contract's invariants, so a malformed observation can
// neither be emitted nor accepted.
func (o Observation) Validate() error {
	if o.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version %d is not %d", o.SchemaVersion, SchemaVersion)
	}
	if len(o.Issue) != 6 {
		return fmt.Errorf("issue %q is not a six-digit id", o.Issue)
	}
	if _, err := time.Parse(time.RFC3339, o.ObservedAt); err != nil {
		return fmt.Errorf("observed_at: %w", err)
	}
	reads := []struct {
		name string
		r    Read
	}{
		{"tracker", o.Tracker.Read}, {"card", o.Card.Read}, {"assignment", o.Assignment.Read},
		{"workspaces", o.Workspaces.Read}, {"branch", o.Branch.Read}, {"checkpoints", o.Checkpoints.Read},
		{"completion", o.Completion.Read}, {"landing", o.Landing.Read},
	}
	for _, rv := range o.Checkpoints.Reviews {
		reads = append(reads, struct {
			name string
			r    Read
		}{"checkpoints.reviews[" + rv.Boundary + "]", rv.Read})
	}
	for _, s := range reads {
		if err := s.r.validate(); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	if o.Workspaces.Holding == nil || o.Checkpoints.Reviews == nil {
		return errors.New("collections must be present (empty, never null)")
	}
	for name, got := range map[string]struct{ have, want Authority }{
		"card": {o.Card.Authority, AuthorityTracker}, "assignment": {o.Assignment.Authority, AuthorityTracker},
		"completion": {o.Completion.Authority, AuthorityTracker}, "landing": {o.Landing.Authority, AuthorityTracker},
		"workspaces": {o.Workspaces.Authority, AuthorityWorktree},
		"branch":     {o.Branch.Authority, AuthorityCommitted}, "checkpoints": {o.Checkpoints.Authority, AuthorityCommitted},
	} {
		if got.have != got.want {
			return fmt.Errorf("%s: authority %q, want %q", name, got.have, got.want)
		}
	}
	for _, rv := range o.Checkpoints.Reviews {
		if !boundaryRE.MatchString(rv.Boundary) {
			return fmt.Errorf("checkpoints: review boundary %q is not plan, close or a milestone", rv.Boundary)
		}
		if rv.Verdict != "" && !vocab.Verdict().IsEmitted(rv.Verdict) {
			return fmt.Errorf("checkpoints: review %s verdict %q is not a review verdict", rv.Boundary, rv.Verdict)
		}
		// A verdict is set exactly when a review artifact was read (present)
		// for a boundary that records one — the plan boundary keeps a ledger,
		// not a verdict.
		if wantVerdict := rv.State == Present && rv.Boundary != "plan"; (rv.Verdict != "") != wantVerdict {
			return fmt.Errorf("checkpoints: review %s verdict is set exactly when its artifact was read", rv.Boundary)
		}
	}
	if f := o.Checkpoints.Flow; f != nil {
		if !flow.ValidKind(f.Kind) {
			return fmt.Errorf("checkpoints: flow kind %q", f.Kind)
		}
		if !flow.ValidProvenance(f.Provenance) {
			return fmt.Errorf("checkpoints: flow provenance %q", f.Provenance)
		}
	}
	valued := func(r Read) bool { return r.State == Present || r.State == Stale }
	if (o.Assignment.Relation != "") != valued(o.Assignment.Read) {
		return errors.New("assignment: relation is set exactly when the read yielded a value")
	}
	if (o.Landing.Outcome != "") != valued(o.Landing.Read) {
		return errors.New("landing: outcome is set exactly when the read yielded a value")
	}
	if o.Landing.Outcome != "" && o.Landing.Outcome != OutcomeLanded && o.Landing.Outcome != OutcomeNotLanded {
		return fmt.Errorf("landing: unknown outcome %q", o.Landing.Outcome)
	}
	switch o.Assignment.Relation {
	case "", RelationThisWorkspace, RelationOtherWorkspace, RelationUnattributed, RelationUnknown:
	default:
		return fmt.Errorf("assignment: unknown relation %q", o.Assignment.Relation)
	}
	switch o.Assignment.ClaimantWorktree {
	case "", FateHoldsBranch, FateElsewhere, FateMissing, FateOtherMachine, FateUnknown:
	default:
		return fmt.Errorf("assignment: unknown claimant_worktree %q", o.Assignment.ClaimantWorktree)
	}
	if (o.Assignment.ClaimantWorktree != "") != (o.Assignment.Claimant != nil) {
		return errors.New("assignment: claimant_worktree is set exactly when a claimant is")
	}
	return nil
}

func (r Read) validate() error {
	switch r.State {
	case Present, Absent:
		if r.Error != "" {
			return fmt.Errorf("%s read carries an error", r.State)
		}
	case Stale, Unknown:
		if r.Error == "" {
			return fmt.Errorf("%s read must say why (error)", r.State)
		}
	default:
		return fmt.Errorf("unknown state %q", r.State)
	}
	return nil
}

// MarshalJSON emits only a valid observation.
func (o Observation) MarshalJSON() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, fmt.Errorf("observation: %w", err)
	}
	type wire Observation
	return json.Marshal(wire(o))
}

// UnmarshalJSON accepts only a valid observation with no unknown or duplicate
// keys.
func (o *Observation) UnmarshalJSON(raw []byte) error {
	if err := rejectDuplicateKeys(raw); err != nil {
		return err
	}
	type wire Observation
	var w wire
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return fmt.Errorf("observation: %w", err)
	}
	if err := Observation(w).Validate(); err != nil {
		return fmt.Errorf("observation: %w", err)
	}
	*o = Observation(w)
	return nil
}

// rejectDuplicateKeys walks the JSON token stream and refuses any object that
// names a key twice (encoding/json silently keeps the last).
func rejectDuplicateKeys(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	type frame struct {
		object bool
		keys   map[string]bool
		key    bool // next string token is a key
	}
	var stack []*frame
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("observation: %w", err)
		}
		var top *frame
		if len(stack) > 0 {
			top = stack[len(stack)-1]
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{':
				if top != nil && top.object {
					top.key = true
				}
				stack = append(stack, &frame{object: true, keys: map[string]bool{}, key: true})
				continue
			case '[':
				if top != nil && top.object {
					top.key = true
				}
				stack = append(stack, &frame{})
				continue
			default:
				stack = stack[:len(stack)-1]
				continue
			}
		case string:
			if top != nil && top.object && top.key {
				if top.keys[t] {
					return fmt.Errorf("observation: duplicate key %q", t)
				}
				top.keys[t] = true
				top.key = false
				continue
			}
		}
		if top != nil && top.object {
			top.key = true
		}
	}
}
