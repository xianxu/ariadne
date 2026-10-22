package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestQueueHelpUsesNoticeOnlyDrainContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"help", []string{"queue", "--help"}},
		{"noargs", []string{"queue"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := buildRoot()
			registered, _, err := root.Find([]string{"queue"})
			if err != nil || registered == root {
				t.Fatalf("queue registration missing: %v", err)
			}
			// Preserve the actual root's rendered help wiring, replacing only queue's
			// dependencies so any storage, authority or live inspection is a failure.
			d := queueFixture(t)
			d.store = func() (*queuegate.Store, error) {
				t.Fatal("queue help accessed queue storage")
				return nil, nil
			}
			d.legacy = func(context.Context, []queuegate.Operation) ([]queuegate.Activity, []string) {
				t.Fatal("queue help inspected live queue state")
				return nil, nil
			}
			d.authority.caller = func() (string, *int, bool) {
				t.Fatal("queue help inspected caller authority")
				return "", nil, false
			}
			d.authority.couchTree = func() (*int, bool) {
				t.Fatal("queue help inspected Couch authority")
				return nil, false
			}
			queue := newQueueCmd(d)
			queue.Long, queue.Short = registered.Long, registered.Short
			root.RemoveCommand(registered)
			root.AddCommand(queue)
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			help := strings.Join(strings.Fields(strings.ReplaceAll(out.String(), "`", "")), " ")
			for _, pointer := range []string{"tl-worker --help", "couch --skill"} {
				if !strings.Contains(help, pointer) {
					t.Errorf("queue help lacks authoritative drain pointer %q", pointer)
				}
			}
			for _, obsolete := range []string{
				"drain notice formatting and acknowledgments",
				"acknowledge the authenticated drain ID",
				"couch --drain --ack",
			} {
				if strings.Contains(help, obsolete) {
					t.Errorf("queue help retains obsolete formal drain acknowledgment %q", obsolete)
				}
			}
		})
	}
}
