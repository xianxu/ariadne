package golint

import "testing"

func TestTestSweepPrivateRoot(t *testing.T) {
	for _, tc := range []struct {
		body string
		want int
	}{
		{`root:=t.TempDir(); sweepStaleTemp(root,true,now,fake)`, 0},
		{`root:=t.TempDir(); child:=filepath.Join(root,"tmp"); sweepStaleTemp(child,true,now,fake)`, 0},
		{`sweepStaleTemp(t.TempDir(),true,now,fake)`, 0},
		{`root:=os.TempDir(); sweepStaleTemp(root,true,now,fake)`, 1},
		{`scratch:=prepareScratch(); sweepStaleTemp(filepath.Dir(scratch.Dir),true,now,fake)`, 1},
		{`root:=t.TempDir(); root=os.TempDir(); sweepStaleTemp(root,true,now,fake)`, 1},
		{`root:=t.TempDir(); if uncertain {root=os.TempDir()}; sweepStaleTemp(root,true,now,fake)`, 1},
		{`root:=t.TempDir(); alias:=root; alias=live; sweepStaleTemp(alias,true,now,fake)`, 1},
		{`root:=t.TempDir(); alias:=root; sweepStaleTemp(alias,true,now,fake)`, 0},
		{`root:=t.TempDir(); sweepStaleTemp(filepath.Dir(root),true,now,fake)`, 1},
		{`root:=t.TempDir(); sweepStaleTemp(filepath.Join(root,".."),true,now,fake)`, 1},
		{`root:=t.TempDir(); sweepStaleTemp(filepath.Join(root,name),true,now,fake)`, 1},
		{`root:=t.TempDir(); collect:=sweepStaleTemp; collect(runtimeRoot,true,now,fake)`, 1},
		{`root:=t.TempDir(); {root:=os.TempDir();sweepStaleTemp(root,true,now,fake)}; sweepStaleTemp(root,true,now,fake)`, 1},
		{`root:=t.TempDir(); mutate(&root); sweepStaleTemp(root,true,now,fake)`, 1},
		{`root:=t.TempDir(); for root=range roots{};sweepStaleTemp(root,true,now,fake)`, 1},
		{`other(os.TempDir(),true,now,fake)`, 0},
	} {
		got, err := Check("fixture_test.go", []byte("package p\nimport(\"testing\";\"os\";\"path/filepath\")\nfunc f(t *testing.T){"+tc.body+"}"), []Rule{ruleOf(t, "test-sweep-private-root")})
		if err != nil || len(got) != tc.want {
			t.Errorf("%s: got%v err%v want%d", tc.body, got, err, tc.want)
		}
	}
}
