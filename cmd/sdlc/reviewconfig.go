package main

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// reviewConfigPath is a repo's review configuration (#310); the loader is
// pkg/reviewpolicy's (#441), shared with doc-review.
const reviewConfigPath = reviewpolicy.ConfigPath

// loadReviewerPolicy reads root's reviewer policy. An absent file is
// judge.DefaultReviewerMode; an unreadable or invalid one is an error naming
// the file.
func loadReviewerPolicy(root string) (judge.ReviewerPolicy, error) { return reviewpolicy.Load(root) }
