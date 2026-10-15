package usage

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"
)

func credit(t testing.TB, raw string) CreditAmount {
	t.Helper()
	v, err := ParseCredit([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCreditExactNativePrecision(t *testing.T) {
	a := credit(t, "45.012189163599997828")
	b := credit(t, "0.9265415628571427")
	sum, err := a.Add(b)
	if err != nil || sum.String() != "45.938730726457140528" {
		t.Fatalf("sum=%s err=%v", sum.String(), err)
	}
	raw, err := json.Marshal(sum)
	if err != nil || string(raw) != sum.String() {
		t.Fatalf("json=%s err=%v", raw, err)
	}
	if credit(t, "1e-18").String() != "0.000000000000000001" {
		t.Fatal("exponent lost")
	}
	if credit(t, "0.000").String() != "0" {
		t.Fatal("zero not canonical")
	}
	if credit(t, "1.2300").Cmp(credit(t, "1.23")) != 0 {
		t.Fatal("decimal comparison")
	}
}

func TestCreditRejectsUntrustedNumberClass(t *testing.T) {
	for _, raw := range []string{"", "null", `"1"`, "-1", "NaN", "Inf", "01", "1 2", "1e9999999999999999999999999", "1e-37", strings.Repeat("9", 65), "0." + strings.Repeat("1", 37), strings.Repeat(" ", 129) + "1"} {
		if _, err := ParseCredit([]byte(raw)); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	zero := CreditAmount{}
	if zero.String() != "0" || !zero.IsZero() {
		t.Fatal("invalid zero value")
	}
}

func TestCreditJSONPreservesExactAggregate(t *testing.T) {
	a := credit(t, strings.Repeat("9", 64))
	sum, err := a.Add(a)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(sum)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CreditAmount
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded.Cmp(sum) != 0 {
		t.Fatalf("aggregate roundtrip: %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &decoded); err == nil {
		t.Fatal("null became measured zero")
	}
}

func FuzzCreditExactRoundTripAndAddition(f *testing.F) {
	for _, s := range []string{"0", "1e-18", "45.938730726457140528", "999999999999999999999999999999999999", "1e9999", "null", "-1"} {
		f.Add(s, s)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		x, e := ParseCredit([]byte(a))
		if e != nil {
			return
		}
		y, e := ParseCredit([]byte(b))
		if e != nil {
			return
		}
		xr, ok := new(big.Rat).SetString(a)
		if !ok {
			t.Fatal("accepted nonrational input")
		}
		yr, ok := new(big.Rat).SetString(b)
		if !ok {
			t.Fatal("accepted nonrational input")
		}
		z, e := x.Add(y)
		if e != nil {
			return
		}
		zr, ok := new(big.Rat).SetString(z.String())
		if !ok || zr.Cmp(new(big.Rat).Add(xr, yr)) != 0 {
			t.Fatal("inexact sum")
		}
		if x.String() != credit(t, x.String()).String() {
			t.Fatal("unstable roundtrip")
		}
		if x.Cmp(y) != xr.Cmp(yr) {
			t.Fatal("inexact comparison")
		}
	})
}
