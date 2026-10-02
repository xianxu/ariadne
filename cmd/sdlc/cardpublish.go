// cardpublish.go — the one publication path for verbs whose whole effect is a
// single card compare-and-swap (#280): claim, adopt, relocation, reclaim and
// the card setters. One seam means one place a lost-acknowledgement test
// injects, and one wording for what an uncertain outcome means.
package main

import (
	"errors"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// cardPublish publishes next over the card read as expected. Its callers are
// pinned by TestCardPublishCallers.
var cardPublish = func(env *trackerEnv, expected tracker.Record, next []byte, token string, trailers []string, beforePush func(base, candidate string) error) error {
	if beforePush == nil {
		beforePush = func(string, string) error { return nil }
	}
	return env.repo.UpdateCardWithTrailers(expected, next, token, trailers, beforePush)
}

// uncertainCardWrite turns a lost publication response into the recovery
// action every single-card verb shares: rerun it, and the card decides.
func uncertainCardWrite(err error, rerun string) error {
	if errors.Is(err, gitx.ErrPublicationUncertain) {
		return fmt.Errorf("%w\n      the change may or may not have published; rerun the same command (%s) — the card decides", err, rerun)
	}
	return err
}
