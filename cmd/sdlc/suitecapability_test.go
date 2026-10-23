package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// A private resource ledger models acquisition, cleanup and failures, rather
// than claiming success from a particular sequence of mock calls.
type capabilityFixture struct {
	bootErr, mkdirErr, listenErr, closeErr, removeErr error
	dir, listening                                    bool
	cancel                                            context.CancelFunc
}
type capabilityCloser struct{ f *capabilityFixture }

func (c capabilityCloser) Close() error { c.f.listening = false; return c.f.closeErr }
func (f *capabilityFixture) probe() suiteCapabilityProbe {
	return suiteCapabilityProbe{
		bootID: func() (string, error) { return "boot", f.bootErr },
		mkdir: func() (string, error) {
			if f.mkdirErr == nil {
				f.dir = true
			}
			return "/private/probe", f.mkdirErr
		},
		listen: func(string) (io.Closer, error) {
			if f.cancel != nil {
				f.cancel()
			}
			if f.listenErr != nil {
				return nil, f.listenErr
			}
			f.listening = true
			return capabilityCloser{f}, nil
		},
		remove: func(string) error {
			if f.listening {
				return errors.New("removed live listener")
			}
			f.dir = false
			return f.removeErr
		},
	}
}

func TestSuiteCapabilityProbe(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fixture capabilityFixture
		want    error
	}{
		{"ready", capabilityFixture{}, nil},
		{"boot denied", capabilityFixture{bootErr: fmt.Errorf("sysctl: %w", syscall.EPERM)}, syscall.EPERM},
		{"scratch denied", capabilityFixture{mkdirErr: syscall.EACCES}, syscall.EACCES},
		{"bind denied", capabilityFixture{listenErr: &os.SyscallError{Syscall: "bind", Err: syscall.EPERM}}, syscall.EPERM},
		{"bind ordinary error", capabilityFixture{listenErr: syscall.EADDRINUSE}, syscall.EADDRINUSE},
		{"close failure", capabilityFixture{closeErr: syscall.EIO}, syscall.EIO},
		{"remove failure", capabilityFixture{removeErr: syscall.EIO}, syscall.EIO},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := probeSuiteCapabilities(context.Background(), tc.fixture.probe())
			if (err == nil) != (tc.want == nil) || tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
			if tc.fixture.dir || tc.fixture.listening {
				t.Fatal("owned resources leaked")
			}
		})
	}
}

func TestSuiteCapabilityCancellation(t *testing.T) {
	for _, before := range []bool{true, false} {
		ctx, cancel := context.WithCancel(context.Background())
		f := capabilityFixture{}
		if before {
			cancel()
		} else {
			f.cancel = cancel
		}
		err := probeSuiteCapabilities(ctx, f.probe())
		cancel()
		if !errors.Is(err, context.Canceled) || f.dir || f.listening {
			t.Fatalf("before=%v err=%v resources=%+v", before, err, f)
		}
	}
}

func TestSuiteCapabilityAdvice(t *testing.T) {
	for _, cause := range []error{syscall.EPERM, syscall.EACCES, errors.New("quoted operation not permitted")} {
		err := &suiteEnvironmentError{Check: "affected", Head: "abc", Cause: cause,
			Diagnosis: suiteStartupDiagnosis{Resource: "suite host capability"}}
		out := err.Error()
		if strings.Contains(out, "whole Bash call") != errors.Is(cause, syscall.EPERM) {
			t.Fatalf("advice inferred from text, not kernel error: %s", out)
		}
		if strings.Contains(out, "; log ") {
			t.Fatalf("preflight invented a log path: %s", out)
		}
		// Main's canonical command keeps affected retries in boarding mode;
		// the preflight's whole-call advice must preserve that mode too.
		if !strings.Contains(out, "sdlc local-check run --name affected --boarding --rev abc") {
			t.Fatal(out)
		}
	}
}

func TestSuiteCapabilityShortSocket(t *testing.T) {
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), strings.Repeat("long", 50)))
	root := filepath.Join(t.TempDir(), strings.Repeat("scratch", 30))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_TEST_SCRATCH", root)
	if err := suiteHostCapabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "scratch-link")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_TEST_SCRATCH", alias)
	if err := suiteHostCapabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
}
