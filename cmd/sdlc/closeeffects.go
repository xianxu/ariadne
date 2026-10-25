package main

import (
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// closeEffects is a lexical, synchronous tail, not a durable job queue. Only
// warning-only publication goes here; authoritative writes precede this tail.
type closeEffects struct {
	after []func()
	lock  func(func() error) error
	churn churn.Report
}

func (e *closeEffects) finish() {
	for _, effect := range e.after {
		effect()
	}
}

func afterClose(f *closeFlags, effect func()) {
	if f.effects == nil {
		effect() // legacy/test callers without the phased command boundary
		return
	}
	f.effects.after = append(f.effects.after, effect)
}

// Pin the exact finalized branch tip before releasing the lock. Publication
// must not silently switch to another actor's newer floating HEAD.
func pushCloseBranch(f *closeFlags, env *trackerEnv, stderr io.Writer, verb string) {
	if f.effects == nil {
		boundaryPush(env, stderr, verb)
		return
	}
	if env.branch == "" || env.onRest() || env.branch == "main" {
		return
	}
	head, err := env.git("rev-parse", "--verify", env.branchRef()+"^{commit}")
	if err != nil {
		cwarn(stderr, fmt.Sprintf("%s: issue branch tip unavailable for push: %v", verb, err))
		return
	}
	expect, err := env.git("for-each-ref", "--format=%(objectname)", gitx.RemoteTrackingRef(env.target.Remote, env.branch))
	if err != nil {
		cwarn(stderr, fmt.Sprintf("%s: issue branch lease unavailable: %v", verb, err))
		return
	}
	publication := closeBranchPublication{root: env.root, target: env.target, branch: env.branch, head: head, expect: expect}
	afterClose(f, func() {
		validate := func(run func() error) error {
			check := func() error {
				target, err := gitx.ResolvePublicationTarget(env.ctx, env.root, env.resting)
				if err != nil {
					return err
				}
				if target != publication.target {
					return fmt.Errorf("close publication target changed; inspect it before pushing")
				}
				return run()
			}
			if f.effects.lock != nil {
				return f.effects.lock(check)
			}
			return check()
		}
		if err := validate(func() error { return nil }); err != nil {
			cwarn(stderr, fmt.Sprintf("%s: issue branch was not pushed: %v", verb, err))
			return
		}
		boundaryPushAt(env, stderr, verb, &publication, validate)
	})
}
