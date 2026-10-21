package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CouchActors is the seam over `couch --actors --json`: which agent runs in
// each slot and its Git resting-branch observation. It needs the running Couch's socket,
// so a sandboxed call fails; callers render `?` then. Queries are bounded.
type CouchActors interface {
	Actors() (map[string]Actor, error) // by slot address, repo:N
}

// Actor is the part of one actor row tl-lead reads.
type Actor struct {
	Agent   string // claude, codex, qoder, …
	Resting bool   // on its resting Git branch; says nothing about runtime readiness
}

type execActors struct {
	bin     string
	timeout time.Duration
}

// Actor queries are light Couch operations; bound recovery observation to ten
// seconds, plus at most one second to drain pipes after cancellation.
const actorsTimeout = 10 * time.Second

func (c execActors) Actors() (map[string]Actor, error) {
	limit := orDuration(c.timeout, actorsTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.bin, "--actors", "--json")
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("couch --actors: no answer within %s: %s", limit, strings.TrimSpace(stderr.String()))
		}
		return nil, fmt.Errorf("couch --actors: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return decodeActors(stdout.Bytes())
}

func decodeActors(raw []byte) (map[string]Actor, error) {
	var r struct {
		Code, Error string
		Actors      []struct {
			Binding struct{ Slot, Agent string }
			Resting bool
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("couch --actors: unreadable reply: %v", err)
	}
	if r.Code != "ok" {
		return nil, fmt.Errorf("couch --actors refused (%s): %s", r.Code, r.Error)
	}
	out := map[string]Actor{}
	for _, a := range r.Actors {
		out[a.Binding.Slot] = Actor{Agent: a.Binding.Agent, Resting: a.Resting}
	}
	return out, nil
}
