package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// reviewConfigPath is a repo's review configuration (#310), repo-local like
// workshop/issue-tracker.json: the base layer never carries it.
const reviewConfigPath = "workshop/review.json"

// reviewConfig is the file's shape.
type reviewConfig struct {
	BoundaryReviewer judge.ReviewerMode                `json:"boundary_reviewer"`
	ReviewerPairs    map[judge.AgentCLI]judge.AgentCLI `json:"reviewer_pairs"`
}

// loadReviewerPolicy reads root's reviewer policy. An absent file is `same`; an
// unreadable or invalid one is an error naming the file.
func loadReviewerPolicy(root string) (judge.ReviewerPolicy, error) {
	path := filepath.Join(root, filepath.FromSlash(reviewConfigPath))
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return judge.ReviewerPolicy{}, nil
	}
	if err != nil {
		return judge.ReviewerPolicy{}, fmt.Errorf("read %s: %w", reviewConfigPath, err)
	}
	var c reviewConfig
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return judge.ReviewerPolicy{}, fmt.Errorf("parse %s: %w", reviewConfigPath, err)
	}
	p := judge.ReviewerPolicy{Mode: c.BoundaryReviewer, Pairs: c.ReviewerPairs}
	if err := p.Validate(); err != nil {
		return judge.ReviewerPolicy{}, fmt.Errorf("%s: %w", reviewConfigPath, err)
	}
	return p, nil
}
