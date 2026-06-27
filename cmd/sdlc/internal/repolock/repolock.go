package repolock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const LockRelPath = ".git/sdlc.lock"
const metadataFile = "meta.json"
const defaultPollInterval = 250 * time.Millisecond
const metadataInitGrace = 2 * time.Second

type Metadata struct {
	PID       int       `json:"pid"`
	Hostname  string    `json:"hostname"`
	CWD       string    `json:"cwd"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	StartedAt time.Time `json:"started_at"`
}

func Encode(m Metadata) []byte {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return []byte("{}\n")
	}
	return append(data, '\n')
}

func Decode(data []byte) (Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return Metadata{}, err
	}
	return m, nil
}

func HolderLine(m Metadata) string {
	parts := []string{}
	if m.PID > 0 {
		parts = append(parts, fmt.Sprintf("pid %d", m.PID))
	}
	if strings.TrimSpace(m.Command) != "" {
		parts = append(parts, strings.TrimSpace(m.Command))
	}
	if m.Hostname != "" {
		parts = append(parts, "on "+m.Hostname)
	}
	line := strings.Join(parts, ": ")
	if line == "" {
		line = "unknown holder"
	}
	if IsLongRunningCommand(m) {
		line += " (long-running review/ship transaction)"
	}
	return line
}

func IsLongRunningCommand(m Metadata) bool {
	fields := strings.Fields(m.Command)
	if len(fields) == 0 {
		return false
	}
	verb := fields[0]
	if verb == "sdlc" && len(fields) > 1 {
		verb = fields[1]
	}
	switch verb {
	case "change-code", "close", "milestone-close", "merge", "push":
		return true
	default:
		return false
	}
}

type ObservationKind int

const (
	ObservationActive ObservationKind = iota
	ObservationStaleMissingProcess
	ObservationStaleAge
)

type Observation struct {
	Kind    ObservationKind
	Message string
}

func Observe(m Metadata, now time.Time, host string, processAlive func(int) bool, maxAge time.Duration) Observation {
	if maxAge > 0 && !m.StartedAt.IsZero() && now.Sub(m.StartedAt) > maxAge {
		return Observation{
			Kind:    ObservationStaleAge,
			Message: fmt.Sprintf("stale sdlc repo lock at %s: holder %s exceeded %s; inspect and remove the lock only if no transaction is running", LockRelPath, HolderLine(m), maxAge),
		}
	}
	if m.Hostname == host && m.PID > 0 && processAlive != nil && !processAlive(m.PID) {
		return Observation{
			Kind:    ObservationStaleMissingProcess,
			Message: fmt.Sprintf("stale sdlc repo lock at %s: holder %s is not running on this host; inspect and remove the lock if no transaction is running", LockRelPath, HolderLine(m)),
		}
	}
	return Observation{Kind: ObservationActive}
}

type Options struct {
	GitCommonDir  string
	Command       string
	Args          []string
	Hostname      string
	PID           int
	CWD           string
	Now           func() time.Time
	ProcessAlive  func(int) bool
	Sleep         func(context.Context, time.Duration) error
	Stderr        io.Writer
	WaitTimeout   time.Duration
	StaleDuration time.Duration
	PollInterval  time.Duration
}

type Lock struct {
	dir    string
	sigCh  chan os.Signal
	stopCh chan struct{}
}

func Acquire(ctx context.Context, opts Options) (*Lock, error) {
	opts = opts.withDefaults()
	lockDir := filepath.Join(opts.GitCommonDir, "sdlc.lock")
	deadline := opts.Now().Add(opts.WaitTimeout)
	initDeadline := time.Time{}
	reported := false
	for {
		err := os.Mkdir(lockDir, 0o700)
		if err == nil {
			initDeadline = time.Time{}
			meta := Metadata{
				PID:       opts.PID,
				Hostname:  opts.Hostname,
				CWD:       opts.CWD,
				Command:   opts.Command,
				Args:      opts.Args,
				StartedAt: opts.Now(),
			}
			if werr := os.WriteFile(filepath.Join(lockDir, metadataFile), Encode(meta), 0o600); werr != nil {
				_ = os.RemoveAll(lockDir)
				return nil, fmt.Errorf("write sdlc repo lock metadata: %w", werr)
			}
			lock := &Lock{dir: lockDir}
			lock.installSignalCleanup()
			return lock, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("create sdlc repo lock %s: %w", lockDir, err)
		}

		holder, readErr := readMetadata(lockDir)
		if readErr != nil {
			if initDeadline.IsZero() {
				initDeadline = opts.Now().Add(metadataInitGrace)
			}
			if opts.Now().Before(initDeadline) {
				if err := opts.Sleep(ctx, opts.PollInterval); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("sdlc repo lock at %s is unreadable after initialization grace: %v; inspect %s and remove it only if no transaction is running", LockRelPath, readErr, lockDir)
		}
		initDeadline = time.Time{}
		obs := Observe(holder, opts.Now(), opts.Hostname, opts.ProcessAlive, opts.StaleDuration)
		if obs.Kind == ObservationStaleMissingProcess {
			if err := os.RemoveAll(lockDir); err != nil {
				return nil, fmt.Errorf("%s; additionally failed to remove stale lock %s: %w", obs.Message, lockDir, err)
			}
			reported = false
			continue
		}
		if obs.Kind != ObservationActive {
			return nil, fmt.Errorf("%s", obs.Message)
		}
		if !reported && opts.Stderr != nil {
			fmt.Fprintf(opts.Stderr, "waiting for sdlc repo lock held by %s\n", HolderLine(holder))
			reported = true
		}
		if !opts.Now().Before(deadline) {
			return nil, fmt.Errorf("timed out waiting for sdlc repo lock held by %s; inspect %s and remove it only if no transaction is running", HolderLine(holder), lockDir)
		}
		if err := opts.Sleep(ctx, opts.PollInterval); err != nil {
			return nil, err
		}
	}
}

func (l *Lock) Release() error {
	if l == nil || l.dir == "" {
		return nil
	}
	l.stopSignalCleanup()
	return os.RemoveAll(l.dir)
}

func (l *Lock) installSignalCleanup() {
	l.sigCh = make(chan os.Signal, 1)
	l.stopCh = make(chan struct{})
	signal.Notify(l.sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-l.sigCh:
			_ = os.RemoveAll(l.dir)
			if sig == syscall.SIGTERM {
				os.Exit(143)
			}
			os.Exit(130)
		case <-l.stopCh:
			return
		}
	}()
}

func (l *Lock) stopSignalCleanup() {
	if l.sigCh == nil {
		return
	}
	signal.Stop(l.sigCh)
	close(l.stopCh)
	l.sigCh = nil
	l.stopCh = nil
}

func readMetadata(lockDir string) (Metadata, error) {
	data, err := os.ReadFile(filepath.Join(lockDir, metadataFile))
	if err != nil {
		return Metadata{}, err
	}
	return Decode(data)
}

func (o Options) withDefaults() Options {
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Sleep == nil {
		o.Sleep = func(ctx context.Context, d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		}
	}
	if o.WaitTimeout <= 0 {
		o.WaitTimeout = 30 * time.Minute
	}
	if o.StaleDuration <= 0 {
		o.StaleDuration = 30 * time.Minute
	}
	if o.PollInterval <= 0 {
		o.PollInterval = defaultPollInterval
	}
	return o
}
