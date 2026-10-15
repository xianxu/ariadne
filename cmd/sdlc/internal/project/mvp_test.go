package project

import (
	"errors"
	"reflect"
	"testing"
)

func TestMVPProgress(t *testing.T) {
	status := map[string]bool{"ariadne#1": true, "pair#2": false, "ariadne#3": true}
	lookup := func(ref string) (bool, error) {
		done, ok := status[ref]
		if !ok {
			return false, errors.New("no such issue")
		}
		return done, nil
	}
	m := MVPProgress([]string{"ariadne#1", "pair#2", "gone#9", "ariadne#3"}, lookup)
	if !reflect.DeepEqual(m.Landed, []string{"ariadne#1", "ariadne#3"}) || !reflect.DeepEqual(m.Open, []string{"pair#2"}) || len(m.Errors) != 1 {
		t.Fatalf("got %+v", m)
	}
	if m.Done() {
		t.Error("open and unreadable refs: not done")
	}
	if !MVPProgress([]string{"ariadne#1"}, lookup).Done() {
		t.Error("all landed: done")
	}
	if MVPProgress([]string{"gone#9"}, lookup).Done() || MVPProgress(nil, lookup).Done() {
		t.Error("an unreadable ref or an empty scope is never done")
	}
}
