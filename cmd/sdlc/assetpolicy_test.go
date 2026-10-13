package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// visibilityGH persists host state between queries, allowing a repo to change
// visibility without rebuilding the caller or hiding state in call expectations.
type visibilityGH struct {
	stubGH
	file string
}

func (g visibilityGH) RepoVisibility(_ context.Context, repo string) (string, error) {
	raw, err := os.ReadFile(g.file)
	if err != nil {
		return "", err
	}
	var state map[string]string
	if err = json.Unmarshal(raw, &state); err != nil {
		return "", err
	}
	v, ok := state[repo]
	if !ok {
		return "", errors.New("unavailable")
	}
	return v, nil
}

func TestAssetPolicyVisibilityAndExplicitPolicy(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root, "https://github.com/example/assets.git")
	git(t, root, "config", "branch.main.remote", "origin")
	git(t, root, "config", "branch.main.merge", "refs/heads/main")
	state := filepath.Join(t.TempDir(), "visibility.json")
	previous := ghClient
	ghClient = visibilityGH{file: state}
	t.Cleanup(func() { ghClient = previous })
	for _, v := range []string{"PUBLIC", "PRIVATE", "unknown"} {
		if err := os.WriteFile(state, []byte(`{"example/assets":"`+v+`"}`), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := resolveAssetPolicy(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		if got.Allowed != (v == "PRIVATE") {
			t.Fatalf("visibility %s: %+v", v, got)
		}
		if got.Source == "" {
			t.Fatal("policy omitted provenance")
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "workshop"), 0755); err != nil {
		t.Fatal(err)
	}
	policy := filepath.Join(root, assetPolicyPath)
	if err := os.WriteFile(policy, []byte(`{"version":1,"policy":"allowed"}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := resolveAssetPolicy(context.Background(), root)
	if err != nil || !got.Allowed || !strings.Contains(got.Source, assetPolicyPath) {
		t.Fatalf("explicit: %+v %v", got, err)
	}
	if err := os.WriteFile(policy, []byte(`{"version":1,"policy":"allowd"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = resolveAssetPolicy(context.Background(), root); err == nil || !strings.Contains(err.Error(), "operator") {
		t.Fatalf("invalid policy remedy: %v", err)
	}
}

func TestAssetPolicyUnknownNeverGrantsAndSymlinkRefuses(t *testing.T) {
	root := t.TempDir()
	got, err := resolveAssetPolicy(context.Background(), root)
	if err != nil || got.Allowed {
		t.Fatalf("unknown: %+v %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "workshop"), 0755); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(other, []byte(`{"version":1,"policy":"allowed"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, filepath.Join(root, assetPolicyPath)); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveAssetPolicy(context.Background(), root); err == nil {
		t.Fatal("symlink policy granted permission")
	}
}
