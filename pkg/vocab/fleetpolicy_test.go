package vocab

import (
	"fmt"
	"slices"
	"testing"
)

func TestFleetPolicyContractMetadata(t *testing.T) {
	m := FleetPolicy()
	if m.DeclarationPath != ".sdlc/fleet.json" {
		t.Fatalf("DeclarationPath = %q", m.DeclarationPath)
	}
	assertStringsEqual(t, "SupportedVersions", intsToStrings(m.SupportedVersions), []string{"1"})
	assertStringsEqual(t, "KeyKinds", m.KeyKinds, []string{"repo", "worktree", "declared-root"})
	assertStringsEqual(t, "CapacityKinds", m.CapacityKinds, []string{"bounded", "unbounded"})
	assertStringsEqual(t, "Actions", m.Actions, []string{"reject", "provision-worktree"})

	for _, version := range m.SupportedVersions {
		if !m.SupportsVersion(version) {
			t.Errorf("SupportsVersion(%d) = false for modeled version", version)
		}
	}
	for _, kind := range m.KeyKinds {
		if !m.IsKeyKind(kind) {
			t.Errorf("IsKeyKind(%q) = false for modeled kind", kind)
		}
	}
	for _, kind := range m.CapacityKinds {
		if !m.IsCapacityKind(kind) {
			t.Errorf("IsCapacityKind(%q) = false for modeled kind", kind)
		}
	}
	for _, action := range m.Actions {
		if !m.IsAction(action) {
			t.Errorf("IsAction(%q) = false for modeled action", action)
		}
	}
}

func TestFleetPolicyMembershipFailsClosed(t *testing.T) {
	m := FleetPolicy()
	if m.SupportsVersion(0) || m.SupportsVersion(2) {
		t.Error("unsupported policy version accepted")
	}
	if m.IsKeyKind("") || m.IsKeyKind("branch") {
		t.Error("unmodeled admission key kind accepted")
	}
	if m.IsCapacityKind("") || m.IsCapacityKind("elastic") {
		t.Error("unmodeled capacity kind accepted")
	}
	if m.IsAction("") || m.IsAction("queue") {
		t.Error("unmodeled on-capacity action accepted")
	}
}

func TestFleetPolicyReturnsImmutableSnapshot(t *testing.T) {
	first := FleetPolicy()
	first.SupportedVersions[0] = 999
	first.KeyKinds[0] = "mutated"

	second := FleetPolicy()
	if second.SupportsVersion(999) || second.IsKeyKind("mutated") {
		t.Fatalf("FleetPolicy exposed mutable package state: %+v", second)
	}
}

func FuzzFleetPolicyMembershipIsExactlyExportedSet(f *testing.F) {
	for _, seed := range []string{"", "repo", "worktree", "declared-root", "bounded", "unbounded", "reject", "provision-worktree", "branch", "queue"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, token string) {
		m := FleetPolicy()
		if got, want := m.IsKeyKind(token), stringIn(m.KeyKinds, token); got != want {
			t.Fatalf("IsKeyKind(%q) = %v, want %v", token, got, want)
		}
		if got, want := m.IsCapacityKind(token), stringIn(m.CapacityKinds, token); got != want {
			t.Fatalf("IsCapacityKind(%q) = %v, want %v", token, got, want)
		}
		if got, want := m.IsAction(token), stringIn(m.Actions, token); got != want {
			t.Fatalf("IsAction(%q) = %v, want %v", token, got, want)
		}
	})
}

func stringIn(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func intsToStrings(values []int) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = fmt.Sprint(value)
	}
	return result
}

func assertStringsEqual(t *testing.T, name string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
