package queue

import (
	"errors"
	"fmt"
	"strings"
)

// Op is the operation an Intent carries.
type Op int

const (
	OpAdd Op = iota
	OpRemove
	OpMove
)

// Intent is one queue edit, expressed as an OPERATION rather than as file
// content. That distinction is the whole design.
//
// gitx.TrunkFile.Update retries a rejected push by re-reading the trunk and
// re-calling its transform. A transform that carries content re-pushes the same
// bytes, silently dropping whatever a peer landed in between — the lost-update
// bug ariadne#188 documents for issue ids. A transform that carries an INTENT
// replays onto the new base, so both edits survive. Reorder looked like it had
// to be content ("here is the whole reordered file") until it was expressed as
// `move X before Y`, at which point it replays exactly like add (ARCH-ORDER: one
// rule, no exception).
type Intent struct {
	Op     Op
	Ref    string
	WhyNow string
	Tag    string
	Kind   Kind

	// Anchor is the ref this one moves before/after (OpMove only).
	Anchor string
	After  bool
}

// Errors a caller renders alongside the current queue so the operator or an
// agent can re-derive the edit. They are typed rather than formatted strings
// because the verb branches on them.
var (
	// ErrAnchorMissing: the ref we were told to move relative to is gone —
	// a peer removed it between our read and our write. Refusing beats guessing
	// a position, because any guess silently reorders someone else's work.
	ErrAnchorMissing = errors.New("queue: anchor is not in the queue")
	// ErrSubjectMissing: the ref we were told to move is gone.
	ErrSubjectMissing = errors.New("queue: ref is not in the queue")
)

// Applied reports what an Apply actually did, for the verb to render. A
// converged no-op is a success worth announcing, not silence.
type Applied struct {
	Note string
}

// Validate checks every user-supplied field this intent interpolates into the
// line format. It is separate from Apply so the caller can run it BEFORE opening
// a network connection.
//
// That separation is not cosmetic. Apply is handed to TrunkFile.Update as a
// transform, and Update fetches and reads the trunk before it calls the
// transform — so validation living only inside Apply meant a malformed ref cost a
// fetch, and offline it produced an "origin unreachable" error that masked the
// real cause entirely. The claim "rejected before any git call" is only true if
// something outside the transform makes it true.
//
// The field list is an ENUMERATION, not a per-field habit: Ref, WhyNow and Tag
// are exactly the values that reach Line.String, and Tag was missed when the
// guards were added one at a time as each field appeared.
func (in Intent) Validate() error {
	if err := ValidateRef(in.Ref); err != nil {
		return err
	}
	if in.Op == OpMove {
		if err := ValidateRef(in.Anchor); err != nil {
			return fmt.Errorf("anchor: %w", err)
		}
	}
	if in.Op != OpAdd {
		return nil // remove and move interpolate no prose
	}
	if err := ValidateWhyNow(in.WhyNow); err != nil {
		return err
	}
	if err := ValidateTag(in.Tag); err != nil {
		return err
	}

	// ROUND-TRIP CHECK, not another forbidden-substring rule.
	//
	// Per-field validators can each pass while their combination renders a line
	// that re-parses to a DIFFERENT record. The live case: a ref of
	// "project:foo" clears ValidateRef, renders as "- project:foo — why", and
	// reads back as Ref "foo" of kind project — so `remove project:foo` finds
	// nothing and a re-add duplicates the entry instead of converging.
	//
	// Enumerating that as one more banned prefix would leave the next
	// combination to be discovered the same way. Rendering the record and
	// re-parsing it asks the actual question — does this survive a write and a
	// read — and it is the same technique that made ParseLine's own round-trip
	// invariant true by construction rather than by argument.
	kind := in.Kind
	if kind == KindUnspecified {
		kind = KindIssue
	}
	want := Line{Ref: in.Ref, WhyNow: in.WhyNow, Tag: in.Tag, Kind: kind, parsed: true}
	got, ok := ParseLine(want.String())
	if !ok || got != want {
		return fmt.Errorf(
			"these values render a line that reads back as something else (%q -> ref %q); "+
				"the entry could not be removed or updated afterwards",
			want.String(), got.Ref)
	}
	return nil
}

// Apply produces the new document. Pure: no git, no clock, no IO.
//
// The switch is exhaustive over Op so a new operation is a compile-time
// obligation at every site that must handle one, rather than a default branch
// that silently does nothing.
func (in Intent) Apply(d *Doc) (*Doc, Applied, error) {
	switch in.Op {
	case OpAdd:
		return in.applyAdd(d)
	case OpRemove:
		return in.applyRemove(d)
	case OpMove:
		return in.applyMove(d)
	default:
		return nil, Applied{}, fmt.Errorf("queue: unknown operation %d", in.Op)
	}
}

// applyAdd appends, or converges when the ref is already queued.
//
// Convergence rather than refusal: two agents adding the same ref concurrently
// have the same goal, and failing the second one would turn agreement into an
// error. The newer why-now wins because it is the more recent judgement, and the
// note says so — a silent overwrite of someone's rationale would be worse than
// either.
func (in Intent) applyAdd(d *Doc) (*Doc, Applied, error) {
	// Defence in depth: the verb validates before any git call, and Apply
	// validates again because it is a public entry point a future caller could
	// reach without going through the verb.
	if err := in.Validate(); err != nil {
		return nil, Applied{}, err
	}
	kind := in.Kind
	if kind == KindUnspecified {
		kind = KindIssue // an outright add defaults to a next action
	}
	line := Line{Ref: in.Ref, WhyNow: in.WhyNow, Tag: in.Tag, Kind: kind, parsed: true}

	out := d.clone()
	if i := out.indexOf(in.Ref); i >= 0 {
		// Converge by MERGING, not replacing. A re-add that omits --tag means
		// "I did not mention the tag", not "remove the tag"; wholesale
		// replacement silently dropped it and could flip an issue line into a
		// project line, while the note claimed only the why-now had moved.
		old := out.lines[i]
		merged := old
		var changed []string
		if in.WhyNow != old.WhyNow {
			merged.WhyNow = in.WhyNow
			changed = append(changed, fmt.Sprintf("why-now (was %q)", old.WhyNow))
		}
		if in.Tag != "" && in.Tag != old.Tag {
			merged.Tag = in.Tag
			changed = append(changed, fmt.Sprintf("tag (was %q)", old.Tag))
		}
		// Only an EXPLICIT kind changes it. Unspecified means the caller said
		// nothing, and silently flipping a peer's project line to an issue line
		// on an edit that never mentioned kind is the bug this represents away.
		if in.Kind != KindUnspecified && in.Kind != old.Kind {
			merged.Kind = in.Kind
			changed = append(changed, "kind")
		}
		out.lines[i] = merged
		if len(changed) == 0 {
			return out, Applied{Note: in.Ref + " was already queued, unchanged"}, nil
		}
		return out, Applied{Note: fmt.Sprintf("%s was already queued; updated %s",
			in.Ref, strings.Join(changed, " and "))}, nil
	}
	line.cr = d.LineEnding() // match the document's endings; do not mix them
	out.lines = append(out.lines, line)
	out.trailingNewline = true
	return out, Applied{}, nil
}

// applyRemove drops the ref, converging when it is already gone.
//
// A peer removing it first achieved the same end state, so this is a success.
// Erroring would make "the work got done" indistinguishable from "the edit
// failed".
func (in Intent) applyRemove(d *Doc) (*Doc, Applied, error) {
	out := d.clone()
	i := out.indexOf(in.Ref)
	if i < 0 {
		return out, Applied{Note: fmt.Sprintf("%s was not in the queue; nothing to remove", in.Ref)}, nil
	}
	out.lines = append(out.lines[:i], out.lines[i+1:]...)
	return out, Applied{}, nil
}

// applyMove repositions the ref relative to an anchor.
//
// Both missing-ref cases REFUSE rather than converge, and the asymmetry with
// remove is deliberate: a removal that already happened reached the intended end
// state, but a move whose subject or anchor has gone has no defensible position
// to land on. Guessing one silently reorders work someone else just did.
func (in Intent) applyMove(d *Doc) (*Doc, Applied, error) {
	out := d.clone()
	from := out.indexOf(in.Ref)
	if from < 0 {
		return nil, Applied{}, fmt.Errorf("%w: %s", ErrSubjectMissing, in.Ref)
	}
	if in.Anchor == in.Ref {
		return nil, Applied{}, fmt.Errorf("queue: cannot move %s relative to itself", in.Ref)
	}
	if out.indexOf(in.Anchor) < 0 {
		return nil, Applied{}, fmt.Errorf("%w: %s (moving %s)", ErrAnchorMissing, in.Anchor, in.Ref)
	}

	line := out.lines[from]
	out.lines = append(out.lines[:from], out.lines[from+1:]...)

	// Re-find the anchor AFTER the removal: its index shifts when the moved line
	// sat above it, and computing the target from the pre-removal index is the
	// classic off-by-one here.
	at := out.indexOf(in.Anchor)
	if in.After {
		at++
	}
	out.lines = append(out.lines[:at], append([]Line{line}, out.lines[at:]...)...)
	return out, Applied{}, nil
}

// clone copies the document so Apply never mutates its input — a transform that
// mutated shared state would corrupt the base on a CAS retry.
func (d *Doc) clone() *Doc {
	out := &Doc{trailingNewline: d.trailingNewline}
	out.lines = append(out.lines, d.lines...)
	return out
}
