package landingline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

const (
	EventWaitStart = "wait_start"
	EventWaitBind  = "wait_bind"
	EventWaitEnd   = "wait_end"
)

// WaitIdentity pins the invocation independently of a recyclable PID or branch.
type WaitIdentity struct {
	Repo        string `json:"repo"`
	Root        string `json:"root"`
	Slot        string `json:"slot"`
	Head        string `json:"head"`
	OperationID string `json:"operation_id"`
	PID         int    `json:"pid"`
	PIDStarted  string `json:"pid_started"`
	Branch      string `json:"branch"`
}
type WaitFact struct {
	Version   int           `json:"version"`
	Identity  WaitIdentity  `json:"identity"`
	Conductor *WaitIdentity `json:"conductor,omitempty"`
}
type WaitRecord struct {
	Event  Event
	Digest string
}
type WaitInterval struct {
	Seq          int64
	Issue        int
	Identity     WaitIdentity
	Start, End   time.Time
	Trains       []int64
	RecordDigest string
}
type WaitIncomplete struct {
	Seq      int64
	Issue    int
	Identity WaitIdentity
	Reason   string
}

func (id WaitIdentity) valid() error {
	if id.Repo == "" || id.Root == "" || !filepath.IsAbs(id.Root) || filepath.Clean(id.Root) != id.Root || id.Slot == "" || id.OperationID == "" || id.PID <= 0 || id.PIDStarted == "" || id.Branch == "" {
		return errors.New("missing canonical invocation identity")
	}
	if address, err := workspace.ParseAddress(id.Slot); err != nil || address.String() != id.Slot {
		return errors.New("slot is not a canonical workspace address")
	}
	for _, value := range []string{id.Repo, id.Root, id.Slot, id.Head, id.OperationID, id.PIDStarted, id.Branch} {
		if len(value) > 4096 {
			return errors.New("invocation identity capacity exceeded")
		}
	}
	if id.Head != strings.ToLower(id.Head) {
		return errors.New("head is not lowercase hexadecimal")
	}
	if len(id.Head) != 40 && len(id.Head) != 64 {
		return errors.New("head is not full length")
	}
	if _, err := hex.DecodeString(id.Head); err != nil {
		return errors.New("head is not hexadecimal")
	}
	return nil
}
func (id *WaitIdentity) UnmarshalJSON(data []byte) error {
	_, err := decodeStrict(data, map[string]any{"repo": &id.Repo, "root": &id.Root, "slot": &id.Slot, "head": &id.Head, "operation_id": &id.OperationID, "pid": &id.PID, "pid_started": &id.PIDStarted, "branch": &id.Branch})
	return err
}
func (w *WaitFact) UnmarshalJSON(data []byte) error {
	_, err := decodeStrict(data, map[string]any{"version": &w.Version, "identity": &w.Identity, "conductor": &w.Conductor})
	return err
}
func validWaitEvent(e Event) error {
	if e.Wait == nil || e.Wait.Version != 1 || e.Seq <= 0 || e.Issue <= 0 || e.At.IsZero() {
		return errors.New("missing wait fact, version, ticket, issue or time")
	}
	if err := e.Wait.Identity.valid(); err != nil {
		return err
	}
	switch e.Kind {
	case EventWaitStart, EventWaitEnd:
		if e.Wait.Conductor != nil || e.Train != 0 {
			return errors.New("endpoint carries conductor")
		}
	case EventWaitBind:
		if e.Wait.Conductor == nil || e.Train <= 0 || e.Train == e.Seq || e.PR <= 0 {
			return errors.New("binding lacks distinct conductor or verified PR")
		}
		c := *e.Wait.Conductor
		if err := c.valid(); err != nil {
			return err
		}
		if c.Repo != e.Wait.Identity.Repo || c.OperationID == e.Wait.Identity.OperationID || (c.PID == e.Wait.Identity.PID && c.PIDStarted == e.Wait.Identity.PIDStarted) {
			return errors.New("binding conductor is foreign or same invocation")
		}
	default:
		return errors.New("unsupported wait event")
	}
	return nil
}

// DecodeWaitEvent strictly decodes known wait facts; callers select the namespace.
func DecodeWaitEvent(raw []byte) (Event, error) {
	var e Event
	if len(raw) > maxTicket {
		return e, errors.New("wait event capacity exceeded")
	}
	_, err := decodeStrict(raw, map[string]any{"kind": &e.Kind, "at": &e.At, "seq": &e.Seq, "issue": &e.Issue, "branch": &e.Branch, "pr": &e.PR, "train": &e.Train, "wait": &e.Wait})
	if err != nil {
		return e, err
	}
	return e, validWaitEvent(e)
}

// ReduceWaitEvidence describes whole invocations, including released/preboarding gaps.
// Facts never prove source idleness and never become accounting subtraction inputs.
func ReduceWaitEvidence(records []WaitRecord) ([]WaitInterval, []WaitIncomplete) {
	type state struct {
		id         WaitIdentity
		issue      int
		start, end *Event
		binds      []Event
		digests    map[string]bool
		rawDigests map[string]bool
		bad        string
	}
	states := map[int64]*state{}
	for _, r := range records {
		e := r.Event
		if !strings.HasPrefix(e.Kind, "wait_") {
			continue
		}
		s := states[e.Seq]
		if s == nil {
			s = &state{issue: e.Issue, digests: map[string]bool{}, rawDigests: map[string]bool{}}
			if e.Wait != nil {
				s.id = e.Wait.Identity
			}
			states[e.Seq] = s
		}
		if decoded, err := hex.DecodeString(r.Digest); err != nil || len(decoded) != sha256.Size || r.Digest != strings.ToLower(r.Digest) {
			s.bad = "missing or malformed raw record digest"
			continue
		}
		if err := validWaitEvent(e); err != nil {
			s.bad = err.Error()
			continue
		}
		if s.id != e.Wait.Identity || s.issue != e.Issue {
			s.bad = "conflicting invocation identity"
		}
		b, _ := json.Marshal(e)
		canonical := string(b)
		if r.Digest != "" {
			s.rawDigests[r.Digest] = true
		}
		if s.digests[canonical] {
			continue
		}
		s.digests[canonical] = true
		switch e.Kind {
		case EventWaitStart:
			if s.start != nil {
				s.bad = "conflicting start"
			} else {
				v := e
				s.start = &v
			}
		case EventWaitEnd:
			if s.end != nil {
				s.bad = "conflicting end"
			} else {
				v := e
				s.end = &v
			}
		case EventWaitBind:
			s.binds = append(s.binds, e)
		}
	}
	var out []WaitInterval
	var notes []WaitIncomplete
	for _, seq := range slices.Sorted(slices.Values(keysWait(states))) {
		s := states[seq]
		if s.start == nil || s.end == nil {
			if s.bad == "" {
				s.bad = "missing owner start or end"
			}
		} else if s.end.At.Before(s.start.At) {
			s.bad = "end precedes start"
		}
		trains := map[int64]bool{}
		bindings := map[int64]Event{}
		for _, b := range s.binds {
			if s.start != nil && b.At.Before(s.start.At) || s.end != nil && b.At.After(s.end.At) {
				s.bad = "binding outside owner endpoints"
			}
			if old, ok := bindings[b.Train]; ok && (!reflect.DeepEqual(old.Wait.Conductor, b.Wait.Conductor) || old.PR != b.PR || !old.At.Equal(b.At)) {
				s.bad = "conflicting conductor binding"
			}
			bindings[b.Train] = b
			trains[b.Train] = true
		}
		if s.bad != "" {
			notes = append(notes, WaitIncomplete{seq, s.issue, s.id, s.bad})
			continue
		}
		if len(trains) == 0 {
			continue
		}
		var ds []string
		for d := range s.rawDigests {
			ds = append(ds, d)
		}
		slices.Sort(ds)
		sum := sha256.Sum256([]byte(strings.Join(ds, "\n")))
		var ts []int64
		for tr := range trains {
			ts = append(ts, tr)
		}
		slices.Sort(ts)
		out = append(out, WaitInterval{seq, s.issue, s.id, s.start.At, s.end.At, ts, hex.EncodeToString(sum[:])})
	}
	return out, notes
}
func keysWait[V any](m map[int64]V) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// BeginWait records before polling. Its held resource is required, never a PID guess.
func (l *Line) BeginWait(h *Handle) error {
	return l.withMutex(func() error {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.file == nil || h.waitStarted {
			return errors.New("wait start requires a fresh held ticket")
		}
		if h.ticket.WaitIdentity == nil {
			return errors.New("wait start lacks invocation identity")
		}
		e := Event{Kind: EventWaitStart, At: l.o.Now(), Seq: h.ticket.Seq, Issue: h.ticket.Issue, Wait: &WaitFact{Version: 1, Identity: *h.ticket.WaitIdentity}}
		if err := l.appendWaitLocked(e); err != nil {
			return err
		}
		h.waitStarted = true
		return nil
	})
}

// EndWait runs after readiness joins, before Release, including admitted heads.
func (l *Line) EndWait(h *Handle) error {
	return l.withMutex(func() error {
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.waitStarted || h.waitEnded || (h.file == nil && h.head == nil) {
			return errors.New("wait end requires its live owner resource")
		}
		h.ticket.WaitEnded = true
		if h.file != nil {
			data, err := json.Marshal(h.ticket)
			if err != nil {
				return err
			}
			if err = h.file.Truncate(0); err != nil {
				return err
			}
			if _, err = h.file.WriteAt(data, 0); err != nil {
				return err
			}
		}
		e := Event{Kind: EventWaitEnd, At: l.o.Now(), Seq: h.ticket.Seq, Issue: h.ticket.Issue, Wait: &WaitFact{Version: 1, Identity: *h.ticket.WaitIdentity}}
		if err := l.appendWaitLocked(e); err != nil {
			return err
		}
		h.waitEnded = true
		return nil
	})
}

// BindWait attaches an independently verified car to this exact held head generation.
func (l *Line) BindWait(conductor *Handle, car Car) error {
	return l.withMutex(func() error {
		if err := l.requireHeadLocked(conductor); err != nil {
			return err
		}
		head := conductor.ticket
		if head.WaitIdentity == nil {
			return errors.New("conductor lacks invocation identity")
		}
		live, err := l.scan()
		if err != nil {
			return err
		}
		for _, t := range live {
			if t.Seq != car.Seq {
				continue
			}
			if t.WaitEnded || t.WaitIdentity == nil || t.WaitIdentity.Head != car.Head || t.Issue != car.Issue {
				return errors.New("verified car does not match live invocation")
			}
			e := Event{Kind: EventWaitBind, At: l.o.Now(), Seq: t.Seq, Issue: t.Issue, PR: car.PR, Train: head.Seq, Wait: &WaitFact{Version: 1, Identity: *t.WaitIdentity, Conductor: head.WaitIdentity}}
			return l.appendWaitLocked(e)
		}
		return ErrGone
	})
}
func (l *Line) appendWaitLocked(e Event) error {
	if err := validWaitEvent(e); err != nil {
		return err
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if len(b) > maxTicket {
		return fmt.Errorf("wait event capacity exceeded")
	}
	return appendLogLocked(l, EventsFile, e, KeepEvents)
}
