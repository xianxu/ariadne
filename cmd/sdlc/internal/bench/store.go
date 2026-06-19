package bench

import (
	"os"
	"path/filepath"
)

// Store is the IO shell for the workshop/benchmarks/ tree. Format logic lives in
// the pure Render*/Parse* functions it calls; the Store itself only does
// filesystem access.
type Store struct{ root string } // e.g. "workshop/benchmarks"

func NewStore(root string) *Store { return &Store{root: root} }

func (s *Store) tasksDir() string { return filepath.Join(s.root, "tasks") }
func (s *Store) runsDir() string  { return filepath.Join(s.root, "runs") }

// WriteTask persists a frozen task. Tasks are immutable, so this is normally
// called once (by `sdlc bench freeze`).
func (s *Store) WriteTask(t Task) error {
	if err := os.MkdirAll(s.tasksDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.tasksDir(), t.ID+".md"), []byte(RenderTask(t)), 0o644)
}

// ReadTask loads a frozen task by id.
func (s *Store) ReadTask(id string) (Task, error) {
	b, err := os.ReadFile(filepath.Join(s.tasksDir(), id+".md"))
	if err != nil {
		return Task{}, err
	}
	return ParseTask(string(b))
}
