package gitx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNetworkGitBoundaries(t *testing.T) {
	dir := t.TempDir()
	script := `#!/bin/sh
printf 'partial'
printf '%s\n' "fatal: unable to access 'https://user:SECRET@git.example/repo': Could not resolve host: git.example" >&2
exit 128
`
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, runner := range []struct {
		name string
		run  func() ([]byte, error)
	}{
		{"bounded-output", func() ([]byte, error) { return boundedGitOutput(context.Background(), nil, "git", "fetch", "origin") }},
		{"input-context", func() ([]byte, error) {
			out, _, err := RunGitContext(context.Background(), "fetch", "origin")
			return out, err
		}},
	} {
		t.Run(runner.name, func(t *testing.T) {
			out, err := runner.run()
			var network *GitNetworkFailure
			if !errors.As(err, &network) || network.Host != "git.example" {
				t.Fatalf("fetch should expose network cause: %v", err)
			}
			if string(out) != "partial" {
				t.Fatalf("stdout contaminated: %q", out)
			}
		})
	}
}

func TestNetworkPublicationCause(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	candidate, err := tf.PrepareCandidate("network candidate", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"network.md": []byte("new")}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	selected, err := tf.SelectCommit(sourceFile(t, repo, "selected-network.md", "new"))
	if err != nil {
		t.Fatal(err)
	}
	cause := errors.New("process cause")
	diagnostic := []byte("fatal: unable to access 'https://user:SECRET@git.example/repo': Could not resolve host: git.example")
	network := ClassifyNetworkFailure([]string{"push"}, diagnostic, cause)
	original := runGitInContext
	t.Cleanup(func() { runGitInContext = original })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "push" {
			return nil, diagnostic, network
		}
		return original(ctx, dir, env, args...)
	}
	check := func(t *testing.T, err error) {
		t.Helper()
		var n *GitNetworkFailure
		if !errors.Is(err, ErrPublicationUncertain) || !errors.Is(err, cause) || !errors.As(err, &n) {
			t.Fatalf("lost sentinel/network/cause: %v", err)
		}
		if strings.Contains(err.Error(), "SECRET") {
			t.Fatalf("credential in rendered error: %v", err)
		}
	}
	t.Run("candidate", func(t *testing.T) {
		outcome, err := tf.PushCandidate(candidate)
		if outcome != PushUnknown {
			t.Fatal(outcome)
		}
		check(t, err)
	})
	t.Run("update-set", func(t *testing.T) {
		err := tf.UpdateMany("network set", func(*TrunkView) (TrunkWrite, error) {
			return TrunkWrite{Write: map[string][]byte{"network.md": []byte("new")}}, nil
		})
		check(t, err)
	})
	t.Run("selected-commit", func(t *testing.T) {
		result, err := tf.PublishCommit(selected)
		if result.Outcome != CommitUncertain {
			t.Fatal(result.Outcome)
		}
		check(t, err)
	})
	t.Run("bootstrap", func(t *testing.T) {
		bootstrap, _ := NewTrunkFile(repo, "origin", "new-tracker")
		_, err := bootstrap.Bootstrap(map[string][]byte{"card": []byte("new")}, "network bootstrap", func(BootstrapResult) error { return nil })
		check(t, err)
	})
	t.Run("read", func(t *testing.T) {
		err := offlineError("origin", network, diagnostic)
		var n *GitNetworkFailure
		if !errors.As(err, &n) || strings.Contains(err.Error(), "SECRET") {
			t.Fatalf("unsafe fetch error: %v", err)
		}
	})
}
