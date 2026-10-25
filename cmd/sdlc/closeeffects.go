package main

import (
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
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
	afterClose(f, func() { boundaryPushAt(env, stderr, verb, head, f.effects.lock) })
}
