package bench

import (
	"path/filepath"
	"testing"
)

func TestStoreTaskRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	task := Task{
		ID: "119-demo", Repo: "ariadne", BaseSHA: "abc", Created: "2026-06-19",
		Spec: "do it", Setup: []string{"go build ./..."}, Rubric: DefaultRubric(),
	}
	if err := s.WriteTask(task); err != nil {
		t.Fatal(err)
	}
	if _, err := filepath.Glob(filepath.Join(dir, "tasks", "*.md")); err != nil {
		t.Fatal(err)
	}
	got, err := s.ReadTask("119-demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseSHA != "abc" {
		t.Errorf("base_sha = %q", got.BaseSHA)
	}
}
