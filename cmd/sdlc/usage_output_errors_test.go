package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

type usageFailingWriter struct {
	err    error
	cancel context.CancelFunc
	short  bool
	calls  int
}

func (w *usageFailingWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.cancel != nil {
		w.cancel()
	}
	if w.short && len(p) > 0 {
		return len(p) - 1, nil
	}
	return 0, w.err
}

func TestUsageReportPropagatesOutputFailures(t *testing.T) {
	sentinel := errors.New("fixture output failure")
	for _, format := range []string{"text", "json"} {
		for _, mode := range []string{"write-error", "short-write", "write-and-cancel"} {
			t.Run(format+"/"+mode, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("XDG_STATE_HOME", t.TempDir())
				testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				writer := &usageFailingWriter{err: sentinel, short: mode == "short-write"}
				want := sentinel
				if writer.short {
					want = io.ErrShortWrite
				}
				if mode == "write-and-cancel" {
					writer.cancel = cancel
				}
				cmd := NewUsageCmd()
				cmd.SetOut(writer)
				cmd.SetErr(&bytes.Buffer{})
				if format == "json" {
					cmd.SetArgs([]string{"--json"})
				} else {
					cmd.SetArgs([]string{})
				}
				err := cmd.ExecuteContext(ctx)
				if writer.calls == 0 || !errors.Is(err, want) {
					t.Fatalf("%s output error=%v, want %v; writes=%d", format, err, want, writer.calls)
				}
				if mode == "write-and-cancel" && !errors.Is(err, context.Canceled) {
					t.Fatalf("output failure suppressed simultaneous cancellation: %v", err)
				}
			})
		}
	}
}

func TestUsageCalibrationOutputFailurePreservesObservation(t *testing.T) {
	sentinel := errors.New("fixture meter output failure")
	for _, mode := range []string{"write-error", "short-write", "write-and-cancel"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			writer := &usageFailingWriter{err: sentinel, short: mode == "short-write"}
			want := sentinel
			if writer.short {
				want = io.ErrShortWrite
			}
			if mode == "write-and-cancel" {
				writer.cancel = cancel
			}
			cmd := NewUsageCmd()
			cmd.SetOut(writer)
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{"calibrate", "--agent", "codex", "--unit", "credits", "--remaining", "12"})
			err := cmd.ExecuteContext(ctx)
			if writer.calls == 0 || !errors.Is(err, want) || !strings.Contains(err.Error(), "recorded") {
				t.Fatalf("meter output failure must identify recorded observation: %v writes=%d", err, writer.calls)
			}
			if mode == "write-and-cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("meter output failure suppressed simultaneous cancellation: %v", err)
			}
			path, e := usage.MeterPath()
			if e != nil {
				t.Fatal(e)
			}
			meter, e := usage.LoadMeter(path)
			if e != nil {
				t.Fatal(e)
			}
			points := usage.MeterPoints(meter, "codex", "credits")
			if len(points) != 1 || points[0].Remaining != 12 {
				t.Fatalf("successful observation was lost on output failure: %+v", meter)
			}
		})
	}
}
