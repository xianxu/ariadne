package reviewbox

import (
	"reflect"
	"strings"
	"testing"
)

func TestMainAuthority(t *testing.T) {
	sha := strings.Repeat("a", 40)
	sha2 := strings.Repeat("b", 40)
	slotConfig := "branch.main.remote origin\nbranch.main.merge refs/heads/main\nremote.origin.fetch +refs/heads/*:refs/remotes/origin/*\n"
	slotRefs := sha + " refs/heads/main\n" + sha2 + " refs/remotes/origin/main\n" + sha + " refs/remotes/origin/000447-x\n"
	for _, tc := range []struct {
		name, config, refs string
		want               authority
		err                bool
	}{
		{"a slot: upstream, fetch refspec and the main refs, not other branches", slotConfig, slotRefs, authority{
			config:  [][2]string{{"branch.main.remote", "origin"}, {"branch.main.merge", "refs/heads/main"}, {"remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"}},
			refs:    [][2]string{{"refs/heads/main", sha}, {"refs/remotes/origin/main", sha2}},
			remotes: []string{"origin"},
		}, false},
		{"no authority: nothing to copy", "", "", authority{}, false},
		{"two remotes with a main: both, each with its fetch", "remote.origin.fetch o\nremote.up.fetch u\nremote.other.fetch x\n",
			sha + " refs/remotes/origin/main\n" + sha2 + " refs/remotes/up/main\n", authority{
				config:  [][2]string{{"remote.origin.fetch", "o"}, {"remote.up.fetch", "u"}},
				refs:    [][2]string{{"refs/remotes/origin/main", sha}, {"refs/remotes/up/main", sha2}},
				remotes: []string{"origin", "up"},
			}, false},
		{"a local upstream names no remote", "branch.main.remote .\nbranch.main.merge refs/heads/trunk\n", sha + " refs/heads/main\n", authority{
			config: [][2]string{{"branch.main.remote", "."}, {"branch.main.merge", "refs/heads/trunk"}},
			refs:   [][2]string{{"refs/heads/main", sha}},
		}, false},
		{"a URL never passes", "remote.origin.url git@github.com:x/y.git\n", "", authority{}, true},
		{"an insteadOf never passes", "url.https://x.insteadof git@x\n", "", authority{}, true},
		{"a push setting never passes", "remote.origin.pushurl git@github.com:x/y.git\n", "", authority{}, true},
		{"a malformed ref row refuses", "", "nothex refs/heads/main\n", authority{}, true},
		{"a valueless config row refuses", "branch.main.remote\n", "", authority{}, true},
	} {
		got, err := mainAuthority(tc.config, tc.refs)
		if (err != nil) != tc.err {
			t.Errorf("%s: err %v", tc.name, err)
			continue
		}
		if !tc.err && !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\n got %+v\nwant %+v", tc.name, got, tc.want)
		}
	}
}
