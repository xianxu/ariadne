package staging

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Metadata is outside the generated/cloned payload, so it is written before a producer can leave a
// partial output. Retries reclaim only stages whose same-host owner is dead.
// Unknown hosts, reused/live PIDs, and missing/invalid metadata are preserved.
type Owner struct {
	Version     int    `json:"version"`
	Destination string `json:"destination"`
	Host        string `json:"host"`
	PID         int    `json:"pid"`
}

func New(destination string) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(filepath.Dir(destination), "."+filepath.Base(destination)+"-weave-")
	if err != nil {
		return "", err
	}
	metadata, _ := json.Marshal(Owner{Version: 1, Destination: destination, Host: host, PID: os.Getpid()})
	if err = os.WriteFile(filepath.Join(stage, "owner.json"), metadata, 0600); err != nil {
		os.RemoveAll(stage)
		return "", err
	}
	return stage, nil
}

func Reclaim(destination string) error {
	parent := filepath.Dir(destination)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return err
	}
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	prefix := "." + filepath.Base(destination) + "-weave-"
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		stage := filepath.Join(parent, entry.Name())
		info, err := os.Lstat(filepath.Join(stage, "owner.json"))
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(stage, "owner.json"))
		if err != nil {
			continue
		}
		var owner Owner
		if json.Unmarshal(data, &owner) != nil || owner.Version != 1 || owner.Destination != destination || owner.Host != host || owner.PID <= 0 {
			continue
		}
		// ESRCH proves no process currently owns this PID. EPERM and every other
		// uncertain result preserve the stage. Concurrent setup is not supported;
		// this conservative guard only prevents recovery from removing active work.
		if err = syscall.Kill(owner.PID, 0); !errors.Is(err, syscall.ESRCH) {
			continue
		}
		if err = os.RemoveAll(stage); err != nil {
			return fmt.Errorf("reclaim interrupted stage %s: %w", stage, err)
		}
	}
	return nil
}

// Remove releases a stage returned by New after the producer has stopped.
func Remove(stage string) error { return os.RemoveAll(stage) }
