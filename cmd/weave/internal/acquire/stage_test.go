package acquire

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureReclaimsOnlyDeadOwnedStages(t *testing.T) {
	base := t.TempDir()
	src := origin(t, base, "base", "", true)
	dest := canonical(filepath.Join(base, "clone"))
	dead := exec.Command("true")
	if err := dead.Run(); err != nil {
		t.Fatal(err)
	}
	host, _ := os.Hostname()
	fixtures := []struct {
		name   string
		owner  stageOwner
		remove bool
	}{
		{"dead", stageOwner{Version: 1, Destination: dest, Host: host, PID: dead.Process.Pid}, true},
		{"live", stageOwner{Version: 1, Destination: dest, Host: host, PID: os.Getpid()}, false},
		{"foreign", stageOwner{Version: 1, Destination: dest, Host: host + "-other", PID: dead.Process.Pid}, false},
		{"other-dest", stageOwner{Version: 1, Destination: dest + "-other", Host: host, PID: dead.Process.Pid}, false},
		{"invalid", stageOwner{}, false},
	}
	for _, f := range fixtures {
		data, _ := json.Marshal(f.owner)
		put(t, filepath.Join(base, ".clone-weave-"+f.name, "owner.json"), string(data))
		put(t, filepath.Join(base, ".clone-weave-"+f.name, "checkout/partial"), "partial")
	}
	put(t, filepath.Join(base, ".clone-weave-unowned", "mine"), "authored")
	if err := Ensure(context.Background(), dest, src, true); err != nil {
		t.Fatal(err)
	}
	for _, f := range fixtures {
		_, err := os.Stat(filepath.Join(base, ".clone-weave-"+f.name))
		if f.remove != os.IsNotExist(err) {
			t.Fatalf("%s: %v", f.name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(base, ".clone-weave-unowned/mine")); err != nil {
		t.Fatal(err)
	}
}
