// Package observe is the read-only workflow observation of one issue (#279):
// who owns it and where its work is, which checkpoints passed, and whether it
// landed — each section naming its source and authority, and saying whether
// the read was present, absent, stale or unknown.
//
// The package is pure: collectors in cmd/sdlc read the existing authorities
// (tracker card, committed artifacts, worktrees) into Inputs; Assemble derives
// every judgment from them. Nothing here performs IO or records state.
package observe

// SchemaVersion is the observation contract's version. Consumers reject any
// other version.
const SchemaVersion = 1

// State is a section's read quality — never its value. Present: the authority
// was read and holds the evidence. Absent: it was read and holds none. Stale:
// read from a last-fetched tracker. Unknown: the read failed (Error says why);
// unknown is never collapsed into absent.
type State string

const (
	Present State = "present"
	Absent  State = "absent"
	Stale   State = "stale"
	Unknown State = "unknown"
)

// Authority says what a section's evidence can prove. Tracker: claim, status,
// completion and landing. Committed: checkpoints recorded on the issue's branch
// or archived on main. Worktree: activity only — dirty files and unpushed
// commits are not progress, and a working card proves a claim, not execution.
type Authority string

const (
	AuthorityTracker   Authority = "tracker"
	AuthorityCommitted Authority = "committed"
	AuthorityWorktree  Authority = "worktree"
)

// Read is the read-quality header every section carries.
type Read struct {
	State  State  `json:"state"`
	Source string `json:"source,omitempty"`
	Error  string `json:"error,omitempty"`
}

type Observation struct {
	SchemaVersion int         `json:"schema_version"`
	Issue         string      `json:"issue"`
	ObservedAt    string      `json:"observed_at"`
	Tracker       Tracker     `json:"tracker"`
	Card          Card        `json:"card"`
	Assignment    Assignment  `json:"assignment"`
	Workspaces    Workspaces  `json:"workspaces"`
	Branch        Branch      `json:"branch"`
	Checkpoints   Checkpoints `json:"checkpoints"`
	Completion    Completion  `json:"completion"`
	Landing       Landing     `json:"landing"`
}

// Tracker is the issue tracker read itself: Ref is the tracker commit the
// answer was read at.
type Tracker struct {
	Read
	Ref      string `json:"ref,omitempty"`
	RefError string `json:"ref_error,omitempty"` // the cards were read, but their tracker commit could not be named
}

type Card struct {
	Read
	Authority Authority `json:"authority"`
	Revision  string    `json:"revision,omitempty"` // the card blob OID (`sdlc reclaim --expect`)
	Status    string    `json:"status,omitempty"`
	Title     string    `json:"title,omitempty"`
	Started   string    `json:"started,omitempty"`
	Updated   string    `json:"updated,omitempty"`
}

// Relation is the recorded owner judged against the queried checkout.
type Relation string

const (
	RelationThisWorkspace  Relation = "this-workspace"
	RelationOtherWorkspace Relation = "other-workspace"
	RelationUnattributed   Relation = "unattributed"
	RelationUnknown        Relation = "unknown"
)

// WorktreeFate is where the recorded owner's worktree stands now.
type WorktreeFate string

const (
	FateHoldsBranch  WorktreeFate = "holds-branch"
	FateElsewhere    WorktreeFate = "elsewhere"     // a local worktree on another branch
	FateMissing      WorktreeFate = "missing"       // not a worktree of this repository here
	FateOtherMachine WorktreeFate = "other-machine" // never probed
	FateUnknown      WorktreeFate = "unknown"
)

type Claimant struct {
	Operator    string `json:"operator"`
	Machine     string `json:"machine"`
	MachineName string `json:"machine_name"`
	Workspace   string `json:"workspace,omitempty"`
	Worktree    string `json:"worktree"`
	Repository  string `json:"repository"`
}

type Assignment struct {
	Read
	Authority        Authority    `json:"authority"`
	Claimant         *Claimant    `json:"claimant,omitempty"`
	Relation         Relation     `json:"relation,omitempty"`
	RelationError    string       `json:"relation_error,omitempty"` // why the relation is unknown
	ClaimantWorktree WorktreeFate `json:"claimant_worktree,omitempty"`
}

type Workspace struct {
	Path       string `json:"path"`
	Address    string `json:"address,omitempty"`
	Branch     string `json:"branch"`
	Head       string `json:"head"`
	DirtyCount int    `json:"dirty_count"`
	Ahead      int    `json:"ahead"`
	Behind     int    `json:"behind"`
	IsClaimant bool   `json:"is_claimant"`
	Error      string `json:"error,omitempty"` // its activity facts could not be read
}

// Workspaces are the local worktrees holding the issue branch — activity only.
type Workspaces struct {
	Read
	Authority Authority   `json:"authority"`
	Holding   []Workspace `json:"holding"`
}

type Branch struct {
	Read
	Authority          Authority `json:"authority"`
	Ref                string    `json:"ref,omitempty"`
	Head               string    `json:"head,omitempty"`
	CommitsAheadOfMain int       `json:"commits_ahead_of_main"`
	LastCommitAt       string    `json:"last_commit_at,omitempty"`
}

type Flow struct {
	Kind       string `json:"kind"`
	Provenance string `json:"provenance"`
}

type Plan struct {
	Total  int `json:"total"`
	Ticked int `json:"ticked"`
}

// Review is one review boundary's evidence: its verdict from the review
// artifact and the boundary's open blocking findings from its gate ledger.
type Review struct {
	Read
	Boundary     string `json:"boundary"` // "plan", "close", "M1", …
	Verdict      string `json:"verdict,omitempty"`
	Window       string `json:"window,omitempty"`
	At           string `json:"at,omitempty"`
	OpenBlocking int    `json:"open_blocking"`
}

type Checkpoints struct {
	Read
	Authority Authority `json:"authority"`
	Flow      *Flow     `json:"flow,omitempty"`
	Plan      Plan      `json:"plan"`
	Reviews   []Review  `json:"reviews"`
}

type Completion struct {
	Read
	Authority      Authority `json:"authority"`
	Token          string    `json:"token,omitempty"`
	EvidenceCommit string    `json:"evidence_commit,omitempty"`
	ReviewedHead   string    `json:"reviewed_head,omitempty"`
	LandedCommit   string    `json:"landed_commit,omitempty"`
}

// Outcome is whether the issue's work landed.
type Outcome string

const (
	OutcomeLanded    Outcome = "landed"
	OutcomeNotLanded Outcome = "not-landed"
)

type Landing struct {
	Read
	Authority    Authority `json:"authority"`
	Outcome      Outcome   `json:"outcome,omitempty"`
	LandedCommit string    `json:"landed_commit,omitempty"`
	Archived     string    `json:"archived,omitempty"`
	ArchiveError string    `json:"archive_error,omitempty"` // the archive lookup failed; Archived is then unknown
}
