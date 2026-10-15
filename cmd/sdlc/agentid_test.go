package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// #367: the verb prints the trailer sdlc stamps (TestMain pins SDLC_AGENT_ID).
func TestAgentIDPrintsTrailer(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"agent-id"}, "Agent: test/test-model/-\n"},
		{[]string{"agent-id", "--json"}, `"trailer":"Agent: test/test-model/-"`},
	} {
		root := buildRoot()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs(tc.args)
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Errorf("%v printed %q, want %q", tc.args, out.String(), tc.want)
		}
	}
}
