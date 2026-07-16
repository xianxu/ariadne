package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestProjectCommandTreeM3(t *testing.T) {
	root := buildRoot()
	project, _, err := root.Find([]string{"project"})
	if err != nil || project == root {
		t.Fatalf("project command not registered: %v", err)
	}
	for _, name := range []string{"new", "list", "show", "set-status", "validate"} {
		found, _, findErr := project.Find([]string{name})
		if findErr != nil || found == project {
			t.Errorf("project %s not registered: %v", name, findErr)
		}
	}
	for _, deferred := range []string{"close"} {
		if found, _, _ := project.Find([]string{deferred}); found != project {
			t.Errorf("project %s registered before M4", deferred)
		}
	}
}

func TestRenderLongProjectDerivesLifecycle(t *testing.T) {
	long := renderLong("project")
	if strings.Contains(long, "{{") {
		t.Fatalf("project Long has an unsubstituted placeholder:\n%s", long)
	}
	m := vocab.Project()
	for _, status := range m.AllStatuses() {
		if !strings.Contains(long, status) {
			t.Errorf("project help missing model status %q", status)
		}
		if when := m.When[status]; when != "" && !strings.Contains(long, when) {
			t.Errorf("project help missing when gloss for %q: %q", status, when)
		}
	}
}
