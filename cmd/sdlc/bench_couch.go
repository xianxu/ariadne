package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

type benchCouchDeps struct {
	Command func(context.Context, ...string) ([]byte, error)
}
type benchActorBinding struct {
	Slot, Repository, Scope, Tag, Session, Nonce, Agent string
	PID                                                 int
	Start                                               string
}
type benchActorSnapshot struct {
	Binding          benchActorBinding
	Known, Supported bool
}
type benchActorsReply struct {
	Code, Error string
	Actors      []benchActorSnapshot
}
type benchPeekReply struct {
	Ref           string `json:"ref"`
	Agent         string `json:"agent"`
	WorkingPath   string `json:"working_path"`
	NativeSession string `json:"native_session"`
	Session       struct {
		Name  string `json:"name"`
		Scope string `json:"scope_key"`
		Tag   string `json:"tag"`
		Nonce string `json:"start_nonce"`
	} `json:"session"`
	Transcripts []string `json:"transcripts"`
	Unavailable string   `json:"unavailable"`
}

// Model/effort are trusted provisioning evidence here. Collection cross-checks
// the actual source logs; neither --actors nor --peek exports those settings.
type benchRuntimeActor struct {
	Identity      bench.RuntimeIdentity `json:"identity"`
	NativeSession string                `json:"native_session"`
	Transcripts   []string              `json:"transcripts"`
	PID           int                   `json:"pid"`
	ProcessStart  string                `json:"process_start"`
	Scope         string                `json:"scope"`
	Tag           string                `json:"tag"`
}

type benchMessageReceipt struct {
	Message struct {
		ID, Body string
		To       benchActorBinding
	}
	Status, Detail string
}

func readBenchDelivery(raw []byte, actor benchRuntimeActor, id, body string) (benchMessageReceipt, error) {
	var reply struct {
		Code, Error string
		Receipt     *benchMessageReceipt
	}
	if err := readBenchCouchJSON(raw, &reply); err != nil {
		return benchMessageReceipt{}, err
	}
	if reply.Code != "ok" || reply.Error != "" || reply.Receipt == nil {
		return benchMessageReceipt{}, errors.New("Couch did not return a confirmed receipt identity")
	}
	r := *reply.Receipt
	to := r.Message.To
	if r.Message.ID == "" || len(r.Message.ID) > 128 || strings.ContainsAny(r.Message.ID, "\x00\r\n") || (id != "" && r.Message.ID != id) || r.Message.Body != body || to != actor.binding() {
		return r, errors.New("Couch receipt differs from exact dispatch/recipient identity")
	}
	switch r.Status {
	case "queued", "submitted", "expired", "not-dispatched":
		return r, nil
	default:
		return r, fmt.Errorf("unsupported Couch receipt status %s", r.Status)
	}
}

func (a benchRuntimeActor) binding() benchActorBinding {
	return benchActorBinding{Slot: a.Identity.Slot, Repository: filepath.Join(a.Identity.Root, ".git"), Scope: a.Scope, Tag: a.Tag, Session: a.Identity.Session, Nonce: a.Identity.Generation, Agent: a.Identity.Agent, PID: a.PID, Start: a.ProcessStart}
}

func (d benchCouchDeps) command(ctx context.Context, args ...string) ([]byte, error) {
	if d.Command != nil {
		return d.Command(ctx, args...)
	}
	child, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(child, "couch", args...)
	processgroup.Configure(cmd)
	join := configureRunnerCleanup(cmd)
	var out, diag benchBoundedOutput
	cmd.Stdout = &out
	cmd.Stderr = &diag
	if err := errors.Join(cmd.Run(), join(), child.Err()); err != nil {
		return out.Bytes(), fmt.Errorf("bench Couch reply unconfirmed: %w: %s", err, diag.String())
	}
	return out.Bytes(), nil
}

func readBenchCouchJSON(raw []byte, target any) error {
	if len(raw) > benchRecordLimit {
		return errors.New("bench Couch reply exceeds 1 MiB")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func benchActors(ctx context.Context, d benchCouchDeps) ([]benchActorSnapshot, error) {
	raw, err := d.command(ctx, "--actors", "--json")
	if err != nil {
		return nil, err
	}
	var reply benchActorsReply
	if err := readBenchCouchJSON(raw, &reply); err != nil {
		return nil, err
	}
	if reply.Code != "ok" || reply.Error != "" || reply.Actors == nil {
		return nil, fmt.Errorf("bench Couch actors unavailable: %s %s", reply.Code, reply.Error)
	}
	return reply.Actors, nil
}

func probeBenchActor(ctx context.Context, d benchCouchDeps, slot, root string, profile bench.AgentProfile) (benchRuntimeActor, error) {
	var result benchRuntimeActor
	if strings.TrimSpace(slot) == "" || strings.ContainsAny(slot, "\x00\r\n") || !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return result, errors.New("bench probe requires exact slot and canonical root")
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil || physical != root {
		return result, errors.New("bench actor root unavailable or substituted")
	}
	actors, err := benchActors(ctx, d)
	if err != nil {
		return result, err
	}
	var selected *benchActorSnapshot
	for i := range actors {
		if actors[i].Binding.Slot == slot {
			if selected != nil {
				return result, errors.New("ambiguous bench actor binding")
			}
			selected = &actors[i]
		}
	}
	if selected == nil || !selected.Known || !selected.Supported {
		return result, errors.New("bench actor is absent, unsupported or unknown")
	}
	b := selected.Binding
	if b.Agent != profile.Agent || b.Repository != filepath.Join(root, ".git") || b.PID <= 0 || b.Start == "" || b.Scope == "" || b.Tag == "" {
		return result, errors.New("bench actor source/profile identity mismatch")
	}
	raw, err := d.command(ctx, "--peek", slot, "--json", "--transcripts", "--lines", "1")
	if err != nil {
		return result, err
	}
	var peek benchPeekReply
	if err := readBenchCouchJSON(raw, &peek); err != nil {
		return result, err
	}
	if peek.Unavailable != "" || peek.Ref != slot || peek.Agent != b.Agent || peek.WorkingPath != root || peek.Session.Name != b.Session || peek.Session.Scope != b.Scope || peek.Session.Tag != b.Tag || peek.Session.Nonce != b.Nonce || peek.NativeSession == "" || len(peek.Transcripts) == 0 || len(peek.Transcripts) > 128 {
		return result, errors.New("bench actor incarnation/root/transcript evidence changed or unavailable")
	}
	seen := map[string]bool{}
	for _, path := range peek.Transcripts {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.ContainsAny(path, "\x00\r\n") || seen[path] {
			return result, errors.New("invalid or duplicate bench transcript path")
		}
		seen[path] = true
	}
	result = benchRuntimeActor{Identity: bench.RuntimeIdentity{Slot: slot, Session: b.Session, Generation: b.Nonce, Root: root, Agent: profile.Agent, Model: profile.Model, Effort: profile.Effort}, NativeSession: peek.NativeSession, Transcripts: append([]string(nil), peek.Transcripts...), PID: b.PID, ProcessStart: b.Start, Scope: b.Scope, Tag: b.Tag}
	if err := result.Identity.Validate(); err != nil {
		return benchRuntimeActor{}, err
	}
	return result, nil
}
