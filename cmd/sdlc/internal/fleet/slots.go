package fleet

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Verdict is a checkout's or slot's readiness to take new work (#289). It is
// an observation: an action that reuses a slot re-checks it at action time.
type Verdict string

const (
	VerdictReady         Verdict = "ready"          // clean, settled, nothing unlanded or claimed
	VerdictHoldsWork     Verdict = "holds-work"     // work to resume: unlanded commits, an open issue, a claim
	VerdictUnknown       Verdict = "unknown"        // a probe it depends on failed; never ready
	VerdictMissing       Verdict = "missing"        // a declared checkout is absent
	VerdictNeedsRecovery Verdict = "needs-recovery" // dirty, mid-operation or detached
)

var verdictOrder = []Verdict{VerdictReady, VerdictHoldsWork, VerdictUnknown, VerdictMissing, VerdictNeedsRecovery}

func (v Verdict) rank() int {
	for i, o := range verdictOrder {
		if o == v {
			return i
		}
	}
	return -1
}

// Worst is the slot fold: the highest-ranked verdict.
func Worst(vs ...Verdict) Verdict {
	worst := VerdictReady
	for _, v := range vs {
		if v.rank() > worst.rank() {
			worst = v
		}
	}
	return worst
}

// MemberVerdict is one checkout's verdict with its reason codes and, for
// failed probes, their errors.
type MemberVerdict struct {
	Verdict Verdict  `json:"verdict"`
	Reasons []string `json:"reasons"`
	Errors  []string `json:"errors,omitempty"`
}

// JudgeCheckout judges one checkout row against its resting branch (pure).
// Facts that show needs-recovery decide it even when another probe failed;
// otherwise a failed probe the verdict depends on makes it unknown;
// otherwise unlanded commits, an open issue or this machine's claim hold
// work; otherwise it is ready.
func JudgeCheckout(row TreeRow, resting string) MemberVerdict {
	var recovery, holds, probes, errs []string
	probe := func(what, err string) { probes, errs = append(probes, "probe:"+what), append(errs, err) }
	f := row.Facts
	if !f.Available {
		probe("facts", f.Error)
	} else {
		if f.DirtyCount != nil && *f.DirtyCount > 0 {
			recovery = append(recovery, "dirty")
		}
		if f.Operation != "" {
			recovery = append(recovery, "operation:"+f.Operation)
		}
		if f.OperationError != "" {
			probe("operation", f.OperationError)
		}
		if !f.BaseAvailable || f.Ahead == nil {
			probe("base", f.BaseError)
		} else if *f.Ahead > 0 {
			holds = append(holds, "unlanded-commits")
		}
	}
	if row.Detached {
		recovery = append(recovery, "detached")
	}
	if row.IssuesError != "" {
		probe("issue", row.IssuesError) // the lookup failed; no match is not a failure
	}
	if row.Branch != "" && row.Branch != resting {
		for _, a := range row.Issues {
			if !vocab.Issue().IsTerminal(a.DeclaredStatus) {
				holds = append(holds, "open-issue:"+a.Ref)
			}
		}
	}
	if claimsCarryValue(row.ClaimsState) {
		for _, c := range row.Claims {
			holds = append(holds, "claimed:"+c.Ref)
		}
	} else if row.ClaimsState != ClaimsAbsent {
		probe("claims", row.ClaimsError)
	}
	switch {
	case len(recovery) > 0:
		return MemberVerdict{Verdict: VerdictNeedsRecovery, Reasons: append(recovery, probes...), Errors: errs}
	case len(holds) > 0 && onlyClaimsUnread(probes):
		// A claim could only add holds-work; the work already held decides.
		return MemberVerdict{Verdict: VerdictHoldsWork, Reasons: holds}
	case len(probes) > 0:
		return MemberVerdict{Verdict: VerdictUnknown, Reasons: probes, Errors: errs}
	case len(holds) > 0:
		return MemberVerdict{Verdict: VerdictHoldsWork, Reasons: holds}
	}
	return MemberVerdict{Verdict: VerdictReady, Reasons: []string{}}
}

func onlyClaimsUnread(probes []string) bool {
	return len(probes) == 1 && probes[0] == "probe:claims"
}

// SlotHost is a slot's host checkout, found from its row's path (#289):
// a fleet primary is repo:0; the canonical slot path of the row's own
// repository is repo:N.
type SlotHost struct {
	Repo     string
	Slot     int
	Address  string
	Resting  string
	HostPath string
	EnvRoot  string // numbered slots only
}

// discoverSlots finds every slot host among the rows (pure), sorted by
// repository then slot number.
func discoverSlots(rows []TreeRow, fleetRoot string) []SlotHost {
	var hosts []SlotHost
	for _, row := range rows {
		if row.Bare {
			continue
		}
		repo := filepath.Base(row.RepoRoot)
		if row.TreePath == row.RepoRoot && filepath.Dir(row.RepoRoot) == fleetRoot {
			hosts = append(hosts, SlotHost{Repo: repo, Address: repo + ":0", Resting: workspace.RestingBranch(0), HostPath: row.TreePath})
			continue
		}
		env := filepath.Dir(row.TreePath)
		n, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(env), repo+"-slot"))
		if err != nil || n < 1 {
			continue
		}
		if want, err := workspace.SlotPath(fleetRoot, repo, n); err != nil || want != row.TreePath {
			continue
		}
		hosts = append(hosts, SlotHost{Repo: repo, Slot: n, Address: repo + ":" + strconv.Itoa(n),
			Resting: workspace.RestingBranch(n), HostPath: row.TreePath, EnvRoot: env})
	}
	sort.Slice(hosts, func(i, j int) bool {
		if hosts[i].Repo != hosts[j].Repo {
			return hosts[i].Repo < hosts[j].Repo
		}
		return hosts[i].Slot < hosts[j].Slot
	})
	return hosts
}

// SlotDeclaration is what a numbered slot's host declares (DeclaredMembers).
type SlotDeclaration struct {
	Members []MemberDecl
	Errors  map[string]string // declaring directory -> why its construct/deps failed
}

// Member roles in a slot.
const (
	RoleHost       = "host"
	RoleDependency = "dependency"
)

// SlotMember is one checkout of a slot and its own verdict.
type SlotMember struct {
	Role          string `json:"role"`
	Path          string `json:"path"`
	RestingBranch string `json:"resting_branch"`
	Branch        string `json:"branch,omitempty"`
	MemberVerdict
}

// Slot is one workspace's readiness: its address, members and the worst
// member verdict. An observation: an action reusing the slot re-checks it.
type Slot struct {
	Address         string       `json:"address"`
	Repo            string       `json:"repo"`
	Slot            int          `json:"slot"`
	EnvironmentRoot string       `json:"environment_root,omitempty"`
	RestingBranch   string       `json:"resting_branch"`
	Verdict         Verdict      `json:"verdict"`
	Members         []SlotMember `json:"members"`
}

// AssembleSlots judges every slot's members and folds them (pure). A :0 slot
// is its host alone; a numbered slot adds its declared dependencies, each
// resting on main (weave's scoped clone policy).
func AssembleSlots(hosts []SlotHost, decls map[string]SlotDeclaration, rows []TreeRow) []Slot {
	byPath := map[string]TreeRow{}
	for _, row := range rows {
		byPath[row.TreePath] = row
	}
	slots := make([]Slot, 0, len(hosts))
	for _, h := range hosts {
		d := decls[h.HostPath]
		judge := func(role, path, resting string) SlotMember {
			m := SlotMember{Role: role, Path: path, RestingBranch: resting}
			row, ok := byPath[path]
			if !ok {
				m.MemberVerdict = MemberVerdict{Verdict: VerdictUnknown, Reasons: []string{"probe:checkout"}, Errors: []string{path + " is not a Git checkout the inventory could read"}}
			} else {
				m.Branch, m.MemberVerdict = row.Branch, JudgeCheckout(row, resting)
			}
			if err, failed := d.Errors[path]; failed {
				m.MemberVerdict = withProbe(m.MemberVerdict, "deps", "construct/deps: "+err)
			}
			return m
		}
		s := Slot{Address: h.Address, Repo: h.Repo, Slot: h.Slot, EnvironmentRoot: h.EnvRoot, RestingBranch: h.Resting}
		s.Members = append(s.Members, judge(RoleHost, h.HostPath, h.Resting))
		if h.Slot > 0 {
			for _, decl := range d.Members {
				switch decl.State {
				case MemberMissing:
					s.Members = append(s.Members, SlotMember{Role: RoleDependency, Path: decl.Path, RestingBranch: workspace.RestingBranch(0),
						MemberVerdict: MemberVerdict{Verdict: VerdictMissing, Reasons: []string{"missing"}}})
				case MemberOutside:
					s.Members = append(s.Members, SlotMember{Role: RoleDependency, Path: decl.Path, RestingBranch: workspace.RestingBranch(0),
						MemberVerdict: MemberVerdict{Verdict: VerdictUnknown, Reasons: []string{"probe:membership"}, Errors: []string{decl.Path + " is declared outside the environment " + h.EnvRoot}}})
				default:
					s.Members = append(s.Members, judge(RoleDependency, decl.Path, workspace.RestingBranch(0)))
				}
			}
		}
		verdicts := make([]Verdict, len(s.Members))
		for i, m := range s.Members {
			verdicts[i] = m.Verdict
		}
		s.Verdict = Worst(verdicts...)
		slots = append(slots, s)
	}
	return slots
}

// withProbe adds a failed probe to a verdict: it cannot stay ready or
// holds-work on an unread fact; needs-recovery (already known) stands.
func withProbe(v MemberVerdict, what, err string) MemberVerdict {
	v.Reasons = append(append([]string(nil), v.Reasons...), "probe:"+what)
	v.Errors = append(append([]string(nil), v.Errors...), err)
	if v.Verdict == VerdictReady || v.Verdict == VerdictHoldsWork {
		v.Verdict = VerdictUnknown
	}
	return v
}

func (s Slot) validate() error {
	if s.Address == "" || s.Repo == "" || s.RestingBranch == "" {
		return errors.New("address, repo and resting branch are required")
	}
	if len(s.Members) == 0 || s.Members[0].Role != RoleHost {
		return errors.New("the host is the first member")
	}
	verdicts := make([]Verdict, len(s.Members))
	for i, m := range s.Members {
		if m.Role != RoleHost && m.Role != RoleDependency || (i > 0) != (m.Role == RoleDependency) {
			return fmt.Errorf("member %s: invalid role %q", m.Path, m.Role)
		}
		if m.Verdict.rank() < 0 {
			return fmt.Errorf("member %s: invalid verdict %q", m.Path, m.Verdict)
		}
		if m.Reasons == nil || (m.Verdict != VerdictReady) != (len(m.Reasons) > 0) {
			return fmt.Errorf("member %s: reasons are required exactly when not ready", m.Path)
		}
		verdicts[i] = m.Verdict
	}
	if s.Verdict != Worst(verdicts...) {
		return fmt.Errorf("verdict %s is not the worst member verdict %s", s.Verdict, Worst(verdicts...))
	}
	return nil
}
