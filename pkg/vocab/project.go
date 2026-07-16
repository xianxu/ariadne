package vocab

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:generate sh -c "vocabulary export --noun project > project.json"

//go:embed project.json
var projectJSON []byte

// ProjectModel is the read-only, parsed `project` noun (ariadne#180): the
// lifecycle funnel, per-status semantics, discovery, and creation scaffold.
// Derived from construct/vocabulary/project.cue at generate time; never
// hand-edited. The single Go read of the project vocabulary — verbs, gates,
// and helptext all derive from here.
type ProjectModel struct {
	Categories map[string][]string `json:"categories"`
	When       map[string]string   `json:"when"`
	// Disc reuses the issue noun's Discovery shape; Plans stays empty —
	// projects have no plan sidecars.
	Disc      Discovery    `json:"discovery"`
	Lifecycle []Transition `json:"lifecycle"`
	Scaf      Scaffold     `json:"scaffold"`
}

// projectCategoryOrder is the project noun's category ordering for AllStatuses:
// the funnel left-to-right, terminal last.
var projectCategoryOrder = []string{"forming", "committed", "executing", "terminal"}

var projectModel = mustLoadProject()

func mustLoadProject() *ProjectModel {
	var m ProjectModel
	if err := json.Unmarshal(projectJSON, &m); err != nil {
		panic(fmt.Sprintf("vocab: corrupt embedded project.json (run `make vocab-embed`): %v", err))
	}
	return &m
}

// Project returns the embedded `project` model.
func Project() *ProjectModel { return projectModel }

// Discovery returns the project noun's location model (home/glob/archive), so
// consumers derive artifact locations from the model instead of hardcoding them.
func (m *ProjectModel) Discovery() Discovery { return m.Disc }

// Sections returns the ordered creation-template body sections, so
// `sdlc project new` derives the section list from the model.
func (m *ProjectModel) Sections() []Section { return m.Scaf.Sections }

// InitialStatus returns the status a newly-created project carries — the first
// member of the `forming` category (the funnel's entry point). Falls back to
// "ideation" only if a corrupt model defines no forming status (mustLoadProject
// already panics on corrupt JSON, so this is a belt-and-suspenders guard).
func (m *ProjectModel) InitialStatus() string {
	forming := m.Categories["forming"]
	if len(forming) == 0 {
		return "ideation"
	}
	return forming[0]
}

// IsTerminal reports whether s is a closed status (done/dropped).
func (m *ProjectModel) IsTerminal(s string) bool { return inCat(m.Categories, "terminal", s) }

// IsExecuting reports whether s is in the live portfolio (executing/paused).
func (m *ProjectModel) IsExecuting(s string) bool { return inCat(m.Categories, "executing", s) }

// IsForming reports whether s is pre-baseline (ideation/defined).
func (m *ProjectModel) IsForming(s string) bool { return inCat(m.Categories, "forming", s) }

// AllStatuses returns every status, funnel-ordered:
// forming → committed → executing → terminal.
func (m *ProjectModel) AllStatuses() []string {
	return allStatuses(m.Categories, projectCategoryOrder)
}

// CanTransition reports whether the lifecycle declares a from→to edge.
func (m *ProjectModel) CanTransition(from, to string) bool {
	return canTransition(m.Lifecycle, from, to)
}

// LegalTransitions returns the statuses `from` may legally transition to, in
// lifecycle order, de-duplicated — for rendering refusal messages.
func (m *ProjectModel) LegalTransitions(from string) []string {
	return legalTransitions(m.Lifecycle, from)
}

// TransitionFor returns the declared from→to edge, or nil when the lifecycle
// has none — the guard runner's lookup surface: the returned edge's Guards
// name the registry entries `sdlc project set-status`/`close` must run.
func (m *ProjectModel) TransitionFor(from, to string) *Transition {
	for i := range m.Lifecycle {
		if m.Lifecycle[i].From == from && m.Lifecycle[i].To == to {
			return &m.Lifecycle[i]
		}
	}
	return nil
}

// RenderLifecycleHelp renders the model-derived lifecycle reference (STATUSES +
// LEGAL TRANSITIONS) for the `sdlc project` help text.
func (m *ProjectModel) RenderLifecycleHelp() string {
	return renderLifecycleHelp(m.AllStatuses(), m.When, m.Lifecycle)
}

// StatusNames joins the status set with sep, in AllStatuses order.
func (m *ProjectModel) StatusNames(sep string) string {
	return strings.Join(m.AllStatuses(), sep)
}
