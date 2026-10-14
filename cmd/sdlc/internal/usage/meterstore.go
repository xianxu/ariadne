package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
)

const MeterMaxBytes = 64 * 1024

func MeterPath() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, e := os.UserHomeDir()
		if e != nil {
			return "", e
		}
		base = filepath.Join(home, ".local", "state")
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("XDG_STATE_HOME must be absolute")
	}
	return filepath.Join(base, "sdlc", "usage-meter.json"), nil
}
func LoadMeter(path string) (Meter, error) {
	f, e := os.Open(path)
	if os.IsNotExist(e) {
		return Meter{Version: 1}, nil
	}
	if e != nil {
		return Meter{}, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MeterMaxBytes+1))
	if e != nil {
		return Meter{}, e
	}
	if len(b) > MeterMaxBytes {
		return Meter{}, fmt.Errorf("meter exceeds %d bytes", MeterMaxBytes)
	}
	var m Meter
	if e = json.Unmarshal(b, &m); e != nil {
		return Meter{}, fmt.Errorf("read meter: %w", e)
	}
	if m.Version != 1 {
		return Meter{}, fmt.Errorf("unsupported stored meter version %d", m.Version)
	}
	if e = ValidateMeter(m); e != nil {
		return Meter{}, e
	}
	return m, nil
}

// RecordMeter serializes only the bounded account file. usage reads never take
// this lock or write a cache. Atomic rename preserves the previous observation.
func RecordMeter(ctx context.Context, path string, o Observation) (Meter, error) {
	if e := ValidateObservation(o); e != nil {
		return Meter{}, e
	}
	if e := ctx.Err(); e != nil {
		return Meter{}, e
	}
	dir := filepath.Dir(path)
	if e := os.MkdirAll(dir, 0700); e != nil {
		return Meter{}, e
	}
	// A dedicated lock parent reuses repolock's cancellation and dead-holder rules
	// without contending with repository lifecycle transactions.
	lockParent := filepath.Join(dir, "usage-meter-lock")
	if e := os.MkdirAll(lockParent, 0700); e != nil {
		return Meter{}, e
	}
	host, _ := os.Hostname()
	lock, e := repolock.Acquire(ctx, repolock.Options{GitCommonDir: lockParent, Command: "usage calibrate", Hostname: host, PID: os.Getpid(), CallerHandlesSignals: true, WaitTimeout: 10 * time.Second, ProcessAlive: meterProcessAlive})
	if e != nil {
		return Meter{}, e
	}
	defer lock.Release()
	old, e := LoadMeter(path)
	if e != nil {
		return Meter{}, e
	}
	m, e := ObserveMeter(old, o)
	if e != nil {
		return Meter{}, e
	}
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return Meter{}, e
	}
	b = append(b, '\n')
	if len(b) > MeterMaxBytes {
		return Meter{}, fmt.Errorf("meter exceeds bound")
	}
	if e = ctx.Err(); e != nil {
		return Meter{}, e
	}
	f, e := os.CreateTemp(dir, ".usage-meter-*")
	if e != nil {
		return Meter{}, e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(b)
	}
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return Meter{}, e
	}
	if e = ctx.Err(); e != nil {
		return Meter{}, e
	}
	if e = os.Rename(tmp, path); e != nil {
		return Meter{}, e
	}
	return m, nil
}
