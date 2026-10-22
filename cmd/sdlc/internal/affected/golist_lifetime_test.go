package affected

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGoListCancellationReclaimsDescendants(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "descendant-survived")
	// The child inherits output pipes. Killing only its parent leaves goList
	// blocked until the child finishes, and lets it write after cancellation.
	script := "#!/bin/sh\n(sleep 2; printf survived > \"$GO_LIST_MARKER\") &\nwait\n"
	if err := os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GO_LIST_MARKER", marker)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := goList(ctx, dir, "-m")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want deadline", err)
	}
	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Errorf("cancellation waited %s for orphaned output pipes", elapsed)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("descendant survived cancellation: marker error %v", err)
	}
}
