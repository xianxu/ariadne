package reviewprepare

import "testing"

func TestParse(t *testing.T) {
	for _, raw := range []string{`{"version":1,"command":["make","build"]}`, `{"version":1,"command":["echo","x"],"timeout_minutes":30}`} {
		got, err := Parse([]byte(raw))
		if err != nil || got.Version != 1 || len(got.Command) != 2 || got.TimeoutMinutes < 1 {
			t.Fatalf("Parse(%s)=%+v,%v", raw, got, err)
		}
	}
	got, _ := Parse([]byte(`{"version":1,"command":["true"]}`))
	if got.TimeoutMinutes != 10 {
		t.Fatal(got)
	}
	for _, raw := range []string{``, `null`, `{}`, `[]`, `{"version":2,"command":["x"]}`, `{"version":1,"command":null}`, `{"version":1,"command":[]}`, `{"version":1,"command":[""]}`, `{"version":1,"command":["x",""]}`, `{"version":1,"command":["x",null]}`, `{"version":1,"command":["x"],"timeout_minutes":null}`, `{"version":1,"command":["x"],"timeout_minutes":0}`, `{"version":1,"command":["x"],"timeout_minutes":31}`, `{"version":1,"command":["x"],"extra":1}`, `{"version":1,"command":["x"]} {}`, `{"version":1,"version":1,"command":["x"]}`, `{"Version":1,"command":["x"]}`} {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}
func FuzzParse(f *testing.F) {
	f.Add([]byte(`{"version":1,"command":["make","build"]}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, b []byte) {
		c, e := Parse(b)
		if e != nil {
			return
		}
		if c.Version != 1 || c.TimeoutMinutes < 1 || c.TimeoutMinutes > 30 || len(c.Command) == 0 {
			t.Fatalf("invalid %+v", c)
		}
		for _, s := range c.Command {
			if s == "" {
				t.Fatal("empty argv")
			}
		}
	})
}
