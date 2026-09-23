package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// reviewArtifact is the exact content/presence consumed by a prepared review.
// Supplied issue/project bytes retain the preparation's view, not a later read.
type reviewArtifact struct {
	path    string
	present bool
	text    string
}

type preparedReview struct {
	root, gitDir, branch, head string
	artifacts                  []reviewArtifact
}

func captureReviewArtifact(path string) (reviewArtifact, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return reviewArtifact{path: path, present: true, text: string(data)}, nil
	}
	if os.IsNotExist(err) {
		return reviewArtifact{path: path}, nil
	}
	return reviewArtifact{}, fmt.Errorf("read review artifact %s: %w", path, err)
}

// capturePreparedReview pins Git identity and every optional artifact, including
// absence. Read failures are never evidence that an artifact does not exist.
func capturePreparedReview(head string, supplied []reviewArtifact, paths ...string) (preparedReview, error) {
	s, err := captureReviewIdentity()
	if err != nil {
		return preparedReview{}, err
	}
	if head != "" && head != s.head {
		return preparedReview{}, fmt.Errorf("review anchor changed during preparation: %s -> %s", head, s.head)
	}
	seen := map[string]reviewArtifact{}
	add := func(a reviewArtifact) error {
		absolute, err := filepath.Abs(a.path)
		if err != nil {
			return err
		}
		a.path = filepath.Clean(absolute)
		if previous, ok := seen[a.path]; ok {
			if previous != a {
				return fmt.Errorf("inconsistent prepared bytes for %s", a.path)
			}
			return nil
		}
		seen[a.path] = a
		s.artifacts = append(s.artifacts, a)
		return nil
	}
	for _, a := range supplied {
		if err := add(a); err != nil {
			return preparedReview{}, err
		}
	}
	for _, path := range paths {
		a, err := captureReviewArtifact(path)
		if err != nil {
			return preparedReview{}, err
		}
		if err := add(a); err != nil {
			return preparedReview{}, err
		}
	}
	return s, nil
}

func captureReviewIdentity() (preparedReview, error) {
	query := func(args ...string) (string, error) {
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			return "", fmt.Errorf("resolve prepared review Git identity (%s): %w", args[0], err)
		}
		return strings.TrimSuffix(string(out), "\n"), nil
	}
	root, err := query("rev-parse", "--show-toplevel")
	if err != nil {
		return preparedReview{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return preparedReview{}, err
	}
	gitDir, err := query("rev-parse", "--absolute-git-dir")
	if err != nil {
		return preparedReview{}, err
	}
	gitDir, err = filepath.EvalSymlinks(gitDir)
	if err != nil {
		return preparedReview{}, err
	}
	head, err := query("rev-parse", "--verify", "HEAD")
	if err != nil {
		return preparedReview{}, err
	}
	branchOut, err := exec.Command("git", "symbolic-ref", "--quiet", "HEAD").Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
			return preparedReview{}, fmt.Errorf("resolve review branch: %w", err)
		}
	}
	return preparedReview{root: root, gitDir: gitDir, branch: strings.TrimSuffix(string(branchOut), "\n"), head: head}, nil
}

// comparePreparedReview owns the read-set decision; gathering current facts is
// separate so the same content/presence policy serves every review caller.
func compareReviewIdentity(want, got preparedReview) error {
	if want.root != got.root || want.gitDir != got.gitDir {
		return fmt.Errorf("review repository/worktree identity changed")
	}
	if want.branch != got.branch {
		return fmt.Errorf("review branch changed from %q to %q", want.branch, got.branch)
	}
	return nil
}

func comparePreparedReview(want, got preparedReview) error {
	if err := compareReviewIdentity(want, got); err != nil {
		return err
	}
	if len(want.artifacts) != len(got.artifacts) {
		return fmt.Errorf("review artifact set changed")
	}
	for i, a := range want.artifacts {
		current := got.artifacts[i]
		if current.path != a.path {
			return fmt.Errorf("review artifact identity changed")
		}
		if current.present != a.present {
			if current.present {
				return fmt.Errorf("%s appeared", a.path)
			}
			return fmt.Errorf("%s disappeared", a.path)
		}
		if current.text != a.text {
			return fmt.Errorf("%s changed", a.path)
		}
	}
	return nil
}

func (s preparedReview) observe() (preparedReview, error) {
	current, err := captureReviewIdentity()
	if err != nil {
		return preparedReview{}, err
	}
	for _, a := range s.artifacts {
		got, err := captureReviewArtifact(a.path)
		if err != nil {
			return preparedReview{}, err
		}
		current.artifacts = append(current.artifacts, got)
	}
	return current, nil
}

func (s preparedReview) validateIdentityAndArtifacts() error {
	current, err := s.observe()
	if err != nil {
		return err
	}
	return comparePreparedReview(s, current)
}

func (s preparedReview) validateExact() error {
	current, err := s.observe()
	if err != nil {
		return err
	}
	if err := comparePreparedReview(s, current); err != nil {
		return err
	}
	if current.head != s.head {
		return fmt.Errorf("review HEAD changed from %s to %s", s.head, current.head)
	}
	return nil
}
