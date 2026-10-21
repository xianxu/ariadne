package landingline

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const TrainControlVersion = 1

var (
	ErrTrainDropPending    = errors.New("landing train: accepted drop pending")
	ErrTrainMergeStarted   = errors.New("landing train: merge operation already started")
	ErrTrainCleanupPending = errors.New("landing train: accepted drop cleanup pending")
	ErrTrainControlStale   = errors.New("landing train: control does not match live head generation")
)

// TrainMember pins both the composed head and its owning checkout. Unknown
// fields survive every read/modify/write, including nested request identities.
type TrainMember struct {
	Seq       int64                      `json:"seq"`
	Issue     int                        `json:"issue"`
	Branch    string                     `json:"branch"`
	Head      string                     `json:"head"`
	PR        int                        `json:"pr"`
	OwnerRoot string                     `json:"owner_root"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (m TrainMember) Car() Car {
	return Car{Seq: m.Seq, Issue: m.Issue, Branch: m.Branch, Head: m.Head, PR: m.PR}
}
func sameTrainMember(a, b TrainMember) bool { return a.Car() == b.Car() && a.OwnerRoot == b.OwnerRoot }

type TrainDrop struct {
	Train  int64                      `json:"train"`
	Member TrainMember                `json:"member"`
	Reason string                     `json:"reason"`
	At     time.Time                  `json:"at"`
	Extra  map[string]json.RawMessage `json:"-"`
}

// Publication is fenced before external writes. Started means train PR
// creation was attempted; intended/previous tips are pinned before a push.
// Cleanup leases the exact positively observed ref tip, even before PR creation.
type TrainPublication struct {
	Started        bool                       `json:"started"`
	Repository     string                     `json:"repository"`
	Remote         string                     `json:"remote"`
	Branch         string                     `json:"branch"`
	RecoveryBranch string                     `json:"recovery_branch"`
	ExpectedTip    string                     `json:"expected_tip"`
	PreviousTip    string                     `json:"previous_tip"`
	PR             int                        `json:"pr"`
	Extra          map[string]json.RawMessage `json:"-"`
}
type TrainCleanup struct {
	Pending     bool                       `json:"pending"`
	Publication TrainPublication           `json:"publication"`
	Extra       map[string]json.RawMessage `json:"-"`
}
type TrainControl struct {
	Version      int                        `json:"version"`
	Train        int64                      `json:"train"`
	Phase        Phase                      `json:"phase"`
	MergeStarted bool                       `json:"merge_started"`
	Members      []TrainMember              `json:"members"`
	Drops        []TrainDrop                `json:"drops"`
	Publication  TrainPublication           `json:"publication"`
	Cleanup      TrainCleanup               `json:"cleanup"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (c TrainControl) CurrentDrops() []TrainDrop {
	var drops []TrainDrop
	for _, d := range c.Drops {
		if d.Train == c.Train {
			drops = append(drops, d)
		}
	}
	return drops
}

type DropReceipt struct {
	Control  TrainControl
	Drop     TrainDrop
	Repeated bool
	// NotificationError follows an accepted safety write. It never turns that
	// durable acceptance into a refusal; admission retries the notification.
	NotificationError error
}

func trainPhaseValid(p Phase) bool {
	switch p {
	case PhaseForming, PhaseIntegrating, PhaseTesting, PhaseBisecting, PhaseAttributing, PhaseReintegrating, PhaseChecking, PhaseLanding, PhaseMerging, PhaseLanded, PhaseAbandoned, PhaseRefused, PhaseUnconfirmed:
		return true
	}
	return false
}
func validateTrainMember(m TrainMember) error {
	if m.Seq <= 0 || m.Issue <= 0 || strings.TrimSpace(m.Branch) == "" || m.PR <= 0 || !validTrainOID(m.Head) {
		return errors.New("invalid pinned train member")
	}
	return nil
}
func validTrainOID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validateTrainControl(c TrainControl) error {
	if c.Version != TrainControlVersion || c.Train <= 0 || !trainPhaseValid(c.Phase) {
		return errors.New("invalid or unsupported train control identity/version/phase")
	}
	if (c.Phase == PhaseMerging || c.Phase == PhaseLanded || c.Phase == PhaseUnconfirmed) && !c.MergeStarted {
		return errors.New("merge phase without irreversible start marker")
	}
	seqs, issues := map[int64]bool{}, map[int]bool{}
	for _, m := range c.Members {
		if err := validateTrainMember(m); err != nil {
			return err
		}
		if seqs[m.Seq] || issues[m.Issue] {
			return errors.New("ambiguous train membership")
		}
		seqs[m.Seq], issues[m.Issue] = true, true
	}
	seen := map[int64]bool{}
	for _, d := range c.Drops {
		if err := validateTrainMember(d.Member); err != nil {
			return err
		}
		if d.Train <= 0 || d.At.IsZero() || strings.TrimSpace(d.Reason) == "" || seen[d.Member.Seq] {
			return errors.New("invalid or duplicate pinned drop")
		}
		seen[d.Member.Seq] = true
	}
	if len(c.CurrentDrops()) > 0 && c.MergeStarted {
		return errors.New("accepted drop and merge start cannot coexist")
	}
	for _, p := range []TrainPublication{c.Publication, c.Cleanup.Publication} {
		if p.PR < 0 || p.Started && (p.Repository == "" || p.Remote == "" || p.Branch == "" || !validTrainOID(p.ExpectedTip)) {
			return errors.New("invalid train publication identity")
		}
		if p.PreviousTip != "" && !validTrainOID(p.PreviousTip) {
			return errors.New("invalid previous train tip")
		}
	}
	return nil
}

func decideTrainDrop(c TrainControl, train int64, member TrainMember, reason string, at time.Time) (TrainControl, TrainDrop, bool, error) {
	if err := validateTrainControl(c); err != nil {
		return c, TrainDrop{}, false, err
	}
	for _, d := range c.Drops {
		if d.Train == train && sameTrainMember(d.Member, member) {
			return c, d, true, nil
		}
	}
	if train != c.Train {
		return c, TrainDrop{}, false, ErrTrainControlStale
	}
	if c.MergeStarted {
		return c, TrainDrop{}, false, fmt.Errorf("%w; train %d is %s", ErrTrainMergeStarted, c.Train, c.Phase)
	}
	if c.Phase == PhaseForming || c.Phase == PhaseIntegrating {
		return c, TrainDrop{}, false, fmt.Errorf("train %d is %s; candidates are not yet composed, no drop accepted", c.Train, c.Phase)
	}
	if c.Phase.Terminal() {
		return c, TrainDrop{}, false, fmt.Errorf("train %d is %s, with no droppable composed car", c.Train, c.Phase)
	}
	i := slices.IndexFunc(c.Members, func(m TrainMember) bool { return sameTrainMember(m, member) })
	if i < 0 || strings.TrimSpace(reason) == "" || at.IsZero() {
		return c, TrainDrop{}, false, errors.New("drop does not pin a current composed member and reason")
	}
	d := TrainDrop{Train: train, Member: c.Members[i], Reason: reason, At: at}
	n := c
	n.Drops = append(slices.Clone(c.Drops), d)
	n.Cleanup.Pending = true
	n.Cleanup.Publication = c.Publication
	// Refresh known publication values without erasing future cleanup-only
	// metadata recorded by a newer reader in this same train generation.
	n.Cleanup.Publication.Extra = mergeExtra(c.Cleanup.Publication.Extra, c.Publication.Extra)
	return n, d, false, nil
}

func decideTrainMerge(c TrainControl, tr Train) (TrainControl, error) {
	if c.Train != tr.ID {
		return c, ErrTrainControlStale
	}
	if len(c.CurrentDrops()) > 0 {
		return c, ErrTrainDropPending
	}
	if tr.Phase != PhaseMerging || !tr.MergeStarted {
		return c, errors.New("merge fence requires the computed merging transition")
	}
	if c.Phase.Terminal() {
		return c, errors.New("terminal train cannot start merge")
	}
	n := c
	n.Phase = tr.Phase
	n.MergeStarted = true
	return n, nil
}

func trainControlFields(c *TrainControl) map[string]any {
	return map[string]any{"version": &c.Version, "train": &c.Train, "phase": &c.Phase, "merge_started": &c.MergeStarted, "members": &c.Members, "drops": &c.Drops, "publication": &c.Publication, "cleanup": &c.Cleanup}
}
func trainMemberFields(m *TrainMember) map[string]any {
	return map[string]any{"seq": &m.Seq, "issue": &m.Issue, "branch": &m.Branch, "head": &m.Head, "pr": &m.PR, "owner_root": &m.OwnerRoot}
}
func trainDropFields(d *TrainDrop) map[string]any {
	return map[string]any{"train": &d.Train, "member": &d.Member, "reason": &d.Reason, "at": &d.At}
}
func trainPublicationFields(p *TrainPublication) map[string]any {
	return map[string]any{"started": &p.Started, "repository": &p.Repository, "remote": &p.Remote, "branch": &p.Branch, "recovery_branch": &p.RecoveryBranch, "expected_tip": &p.ExpectedTip, "previous_tip": &p.PreviousTip, "pr": &p.PR}
}
func trainCleanupFields(c *TrainCleanup) map[string]any {
	return map[string]any{"pending": &c.Pending, "publication": &c.Publication}
}
func marshalTrainObject(base any, extra map[string]json.RawMessage, fields map[string]any) ([]byte, error) {
	data, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	return encodeWithExtra(data, extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}

// Every schema field is present in v1 writes. In particular, omission of a
// nested irreversible marker is unreadable state, never a false default.
func decodeTrainObject(data []byte, fields map[string]any) (map[string]json.RawMessage, error) {
	extra, err := decodeStrict(data, fields)
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for k := range fields {
		if _, ok := raw[k]; !ok {
			return nil, fmt.Errorf("train control missing %s", k)
		}
	}
	return extra, nil
}
func DecodeTrainControl(data []byte) (TrainControl, error) {
	var c TrainControl
	if len(data) > maxTicket {
		return c, errors.New("train control exceeds size limit")
	}
	extra, err := decodeTrainObject(data, trainControlFields(&c))
	if err != nil {
		return c, err
	}
	c.Extra = extra
	return c, validateTrainControl(c)
}
func (c TrainControl) MarshalJSON() ([]byte, error) {
	type plain TrainControl
	if c.Members == nil {
		c.Members = []TrainMember{}
	}
	if c.Drops == nil {
		c.Drops = []TrainDrop{}
	}
	return marshalTrainObject(plain(c), c.Extra, trainControlFields(&c))
}
func (m TrainMember) MarshalJSON() ([]byte, error) {
	type plain TrainMember
	return marshalTrainObject(plain(m), m.Extra, trainMemberFields(&m))
}
func (m *TrainMember) UnmarshalJSON(data []byte) error {
	extra, err := decodeTrainObject(data, trainMemberFields(m))
	m.Extra = extra
	return err
}
func (d TrainDrop) MarshalJSON() ([]byte, error) {
	type plain TrainDrop
	return marshalTrainObject(plain(d), d.Extra, trainDropFields(&d))
}
func (d *TrainDrop) UnmarshalJSON(data []byte) error {
	extra, err := decodeTrainObject(data, trainDropFields(d))
	d.Extra = extra
	return err
}
func (p TrainPublication) MarshalJSON() ([]byte, error) {
	type plain TrainPublication
	return marshalTrainObject(plain(p), p.Extra, trainPublicationFields(&p))
}
func (p *TrainPublication) UnmarshalJSON(data []byte) error {
	extra, err := decodeTrainObject(data, trainPublicationFields(p))
	p.Extra = extra
	return err
}
func (c TrainCleanup) MarshalJSON() ([]byte, error) {
	type plain TrainCleanup
	return marshalTrainObject(plain(c), c.Extra, trainCleanupFields(&c))
}
func (c *TrainCleanup) UnmarshalJSON(data []byte) error {
	extra, err := decodeTrainObject(data, trainCleanupFields(c))
	c.Extra = extra
	return err
}
