package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestResumeCLIValidation(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"resume"}, "give --name suite"},
		{[]string{"resume", "--name", "affected", "--attempt", "aa"}, "only suite"},
		{[]string{"resume", "--name", "suite"}, "give --attempt"},
		{[]string{"resume", "--name", "suite", "--attempt", "../aa"}, "safe parent"},
		{[]string{"reconcile"}, "give --attempt"},
		{[]string{"reconcile", "--attempt", "aa"}, "--link-digest"},
		{[]string{"reconcile", "--attempt", "aa", "--link-digest", strings.Repeat("a", 64)}, "--note"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			c := NewLocalCheckCmd()
			var out bytes.Buffer
			c.SetOut(&out)
			c.SetErr(&out)
			c.SetArgs(tc.args)
			err := c.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("CLI %v: %v; want %s", tc.args, err, tc.want)
			}
		})
	}
}
