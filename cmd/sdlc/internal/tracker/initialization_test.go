package tracker

import (
	"reflect"
	"strings"
	"testing"
)

func TestEmptyTrackerInitEligibility(t *testing.T) {
	tests := []struct {
		name   string
		change func(*MigrationInput, *MigrationManifest)
	}{
		{"published evidence", func(i *MigrationInput, m *MigrationManifest) { i.Published = []MigrationObservedFile{{}} }},
		{"tip evidence", func(i *MigrationInput, m *MigrationManifest) { i.Tips = []MigrationTipEvidence{{}} }},
		{"active", func(i *MigrationInput, m *MigrationManifest) { i.Active = []MigrationFile{{Path: "unrecognized.md"}} }},
		{"archived", func(i *MigrationInput, m *MigrationManifest) { i.Archived = []MigrationFile{{}} }},
		{"branch deletion", func(i *MigrationInput, m *MigrationManifest) { i.Branches = []MigrationBranchFile{{Raw: nil}} }},
		{"dirty", func(i *MigrationInput, m *MigrationManifest) { i.DirtyIssuePaths = []string{"dirty"} }},
		{"anchor", func(i *MigrationInput, m *MigrationManifest) { i.Anchors = map[string][]MigrationAnchor{"1": {{}}} }},
		{"cards", func(i *MigrationInput, m *MigrationManifest) { m.Cards = []MigrationCard{{}} }},
		{"conversion", func(i *MigrationInput, m *MigrationManifest) { m.Conversions = []MigrationConversion{{}} }},
		{"duplicate", func(i *MigrationInput, m *MigrationManifest) { m.Duplicates = []string{"duplicate"} }},
		{"refusal", func(i *MigrationInput, m *MigrationManifest) { m.Refusals = []MigrationRefusal{{}} }},
		{"unknown inventory", func(i *MigrationInput, m *MigrationManifest) { *i = MigrationInput{} }},
		{"different pinned inventory", func(i *MigrationInput, m *MigrationManifest) { m.Main = oidB }},
	}
	in := basicInput()
	manifest := PlanTrackerMigration(in)
	if !EmptyInitEligible(in, manifest) {
		t.Fatal("empty pinned inventory refused")
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := basicInput()
			m := PlanTrackerMigration(in)
			tc.change(&in, &m)
			if EmptyInitEligible(in, m) {
				t.Fatal("unsafe inventory authorized initialization")
			}
		})
	}
}

func TestLegacySetupHint(t *testing.T) {
	if got := LegacySetupHint(true, false); !strings.Contains(got, "sdlc issue init") || !strings.Contains(got, "local") || strings.Contains(got, "sdlc issue migrate") {
		t.Fatalf("empty hint: %q", got)
	}
	if got := LegacySetupHint(true, true); !strings.Contains(got, "sdlc issue migrate") || strings.Contains(got, "sdlc issue init") {
		t.Fatalf("nonempty hint: %q", got)
	}
	for _, hasIssues := range []bool{false, true} {
		if got := LegacySetupHint(false, hasIssues); got != "" {
			t.Fatalf("unknown inventory hint: %q", got)
		}
	}
}

func initStep(t *testing.T, s InitState, event InitEvent, want InitEffect) InitState {
	t.Helper()
	next, e, err := InitTransition(s, event)
	if err != nil || !reflect.DeepEqual(e, want) {
		t.Fatalf("%T: effect %#v want %#v, error %v", event, e, want, err)
	}
	return next
}

func TestInitializationConfirmedSequences(t *testing.T) {
	for _, disposition := range []InitDisposition{InitEmpty, InitTrackerOnly, InitRemoteReady, InitCurrent} {
		t.Run(string(disposition), func(t *testing.T) {
			s := initStep(t, NewInitialization(), BeginInit{}, InspectInitialization{})
			observed := InitInspected{Disposition: disposition, Root: oidA, Commit: oidB}
			switch disposition {
			case InitEmpty:
				observed.Root = ""
				observed.Commit = ""
				s = initStep(t, s, observed, BootstrapTracker{})
				s = initStep(t, s, BootstrapConfirmed{Root: oidA}, PublishCutover{Root: oidA})
				s = initStep(t, s, CutoverConfirmed{Commit: oidB}, FastForwardCheckout{Commit: oidB})
				s = initStep(t, s, CheckoutConfirmed{}, nil)
			case InitTrackerOnly:
				observed.Commit = ""
				s = initStep(t, s, observed, PublishCutover{Root: oidA})
				s = initStep(t, s, CutoverConfirmed{Commit: oidB}, FastForwardCheckout{Commit: oidB})
				s = initStep(t, s, CheckoutConfirmed{}, nil)
			case InitRemoteReady:
				s = initStep(t, s, observed, FastForwardCheckout{Commit: oidB})
				s = initStep(t, s, CheckoutConfirmed{}, nil)
			case InitCurrent:
				s = initStep(t, s, observed, nil)
			}
			if !s.Terminal() || s.Failure() != "" || s.Skipped() || s.Root() != oidA || s.Commit() != oidB {
				t.Fatalf("incomplete result %#v", s)
			}
		})
	}
}

func TestInitializationStopsWithoutConfirmation(t *testing.T) {
	inspecting := initStep(t, NewInitialization(), BeginInit{}, InspectInitialization{})
	bootstrapping := initStep(t, inspecting, InitInspected{Disposition: InitEmpty}, BootstrapTracker{})
	publishing := initStep(t, bootstrapping, BootstrapConfirmed{Root: oidA}, PublishCutover{Root: oidA})
	completing := initStep(t, publishing, CutoverConfirmed{Commit: oidB}, FastForwardCheckout{Commit: oidB})
	for _, s := range []InitState{inspecting, bootstrapping, publishing, completing} {
		for _, reason := range []string{"failed", "canceled", "push acknowledgment unknown"} {
			stopped := initStep(t, s, InitFailed{Reason: reason}, nil)
			if !stopped.Terminal() || stopped.Failure() != reason || stopped.Root() != s.Root() || stopped.Commit() != s.Commit() {
				t.Fatalf("lost confirmed evidence: %#v", stopped)
			}
			events := []InitEvent{BeginInit{}, BootstrapConfirmed{Root: oidA}, CutoverConfirmed{Commit: oidB}, CheckoutConfirmed{}, InitFailed{Reason: reason}}
			for _, event := range events {
				next, e, err := InitTransition(stopped, event)
				if err == nil || e != nil || next != stopped {
					t.Fatalf("stopped state advanced on %T", event)
				}
			}
		}
	}
	for _, disposition := range []InitDisposition{InitSkip, InitRefuse} {
		s := initStep(t, inspecting, InitInspected{Disposition: disposition, Reason: "observed reason"}, nil)
		if !s.Terminal() || s.Skipped() != (disposition == InitSkip) || (disposition == InitRefuse && s.Failure() != "observed reason") {
			t.Fatalf("bad terminal disposition %#v", s)
		}
	}
}

func TestInitializationRejectsOutOfOrderAndInvalidIdentities(t *testing.T) {
	start := NewInitialization()
	inspecting := initStep(t, start, BeginInit{}, InspectInitialization{})
	bootstrapping := initStep(t, inspecting, InitInspected{Disposition: InitEmpty}, BootstrapTracker{})
	publishing := initStep(t, bootstrapping, BootstrapConfirmed{Root: oidA}, PublishCutover{Root: oidA})
	completing := initStep(t, publishing, CutoverConfirmed{Commit: oidB}, FastForwardCheckout{Commit: oidB})
	done := initStep(t, completing, CheckoutConfirmed{}, nil)
	states := []InitState{start, inspecting, bootstrapping, publishing, completing, done}
	events := []InitEvent{BeginInit{}, InitInspected{Disposition: InitEmpty}, BootstrapConfirmed{Root: oidA}, CutoverConfirmed{Commit: oidB}, CheckoutConfirmed{}}
	// Every ordinary event has exactly one valid predecessor. This matrix catches
	// skipped and duplicate phases without coupling the assertion to phase names.
	for si, s := range states {
		for ei, event := range events {
			if si == ei {
				continue
			}
			next, e, err := InitTransition(s, event)
			if err == nil || e != nil || next != s {
				t.Fatalf("state %d accepted out-of-order event %T", si, event)
			}
		}
	}
	for _, bad := range []string{"", "short", strings.Repeat("0", 40), strings.Repeat("A", 40)} {
		for _, tc := range []struct {
			s InitState
			e InitEvent
		}{
			{bootstrapping, BootstrapConfirmed{Root: bad}},
			{publishing, CutoverConfirmed{Commit: bad}},
			{inspecting, InitInspected{Disposition: InitTrackerOnly, Root: bad}},
			{inspecting, InitInspected{Disposition: InitRemoteReady, Root: oidA, Commit: bad}},
			{inspecting, InitInspected{Disposition: InitCurrent, Root: bad, Commit: oidB}},
		} {
			next, e, err := InitTransition(tc.s, tc.e)
			if err == nil || e != nil || next != tc.s {
				t.Fatalf("accepted identity %q in %T", bad, tc.e)
			}
		}
	}
	for _, event := range []InitEvent{nil, InitInspected{Disposition: "unknown"}, InitInspected{Disposition: InitEmpty, Root: oidA}, InitFailed{}} {
		next, e, err := InitTransition(inspecting, event)
		if err == nil || e != nil || next != inspecting {
			t.Fatalf("accepted invalid event %#v", event)
		}
	}
}
