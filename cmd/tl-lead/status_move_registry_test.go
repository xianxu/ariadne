package main

import (
	"testing"
)

func TestMoveRegistrationReply(t *testing.T) {
	good := `{"Code":"ok","Registration":{"ID":"identity","File":"/a.md","Owner":"ops:0","Future":{"x":1e400}},"Future":1e400}`
	r, err := decodeMoveRegistration([]byte(good))
	if err != nil || r == nil || r.ID != "identity" {
		t.Fatalf("%+v %v", r, err)
	}
	for _, text := range []string{
		`{"Code":"ok"}`, `{"Code":"ok","Registration":{}}`,
		`{"Code":"ok","Code":"ok","Registration":null}`,
		`{"Code":"ok","Registration":{"ID":3,"File":"/a.md","Owner":"ops:0"}}`,
		`{"Code":"ok","Registration":{"ID":"a","ID":"b","File":"/a.md","Owner":"ops:0"}}`,
		`{"Code":"unavailable","Registration":null}`, `{"Code":"ok","Registration":null} garbage`,
	} {
		if _, err := decodeMoveRegistration([]byte(text)); err == nil {
			t.Errorf("accepted %s", text)
		}
	}
	if r, err := decodeMoveRegistration([]byte(`{"Code":"ok","Registration":null}`)); r != nil || err != nil {
		t.Fatalf("absence: %+v %v", r, err)
	}
}
