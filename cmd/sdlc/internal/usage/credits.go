package usage

import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// CreditAmount is an immutable, exact native credit measurement. Its zero value
// is measured zero; an unavailable measurement is represented by a nil pointer.
// No binary float participates in parsing, normalization or accumulation.
type CreditAmount struct{ decimal string }

var creditNumber = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// ParseCredit accepts bounded nonnegative JSON numbers. Precision limits are
// checked before math/big sees an exponent or allocates its coefficient.
func ParseCredit(raw []byte) (CreditAmount, error) {
	bad := fmt.Errorf("invalid or out-of-bounds reported credits")
	if len(raw) == 0 || len(raw) > 128 {
		return CreditAmount{}, bad
	}
	s := string(bytes.TrimSpace(raw))
	if !creditNumber.MatchString(s) {
		return CreditAmount{}, bad
	}
	mantissa, exponent := s, 0
	if at := strings.IndexAny(s, "eE"); at >= 0 {
		mantissa = s[:at]
		var err error
		exponent, err = strconv.Atoi(s[at+1:])
		if err != nil || exponent < -128 || exponent > 128 {
			return CreditAmount{}, bad
		}
	}
	digits := strings.ReplaceAll(mantissa, ".", "")
	if len(strings.TrimLeft(digits, "0")) > 64 {
		return CreditAmount{}, bad
	}
	scale := -exponent
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		scale += len(mantissa) - dot - 1
	}
	if scale > 36 {
		return CreditAmount{}, bad
	}
	if scale < 0 {
		scale = 0
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return CreditAmount{}, bad
	}
	v, err := canonicalCredit(r, scale, 64)
	if err != nil {
		return CreditAmount{}, bad
	}
	return v, nil
}

func canonicalCredit(r *big.Rat, scale, digits int) (CreditAmount, error) {
	s := r.FloatString(scale)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if len(strings.TrimLeft(strings.ReplaceAll(s, ".", ""), "0")) > digits {
		return CreditAmount{}, fmt.Errorf("credit aggregate bound exceeded")
	}
	return CreditAmount{decimal: s}, nil
}

func (c CreditAmount) String() string {
	if c.decimal == "" {
		return "0"
	}
	return c.decimal
}
func (c CreditAmount) IsZero() bool                 { return c.String() == "0" }
func (c CreditAmount) MarshalJSON() ([]byte, error) { return []byte(c.String()), nil }
func (c CreditAmount) rat() *big.Rat                { r, _ := new(big.Rat).SetString(c.String()); return r }
func (c CreditAmount) Cmp(d CreditAmount) int       { return c.rat().Cmp(d.rat()) }
func (c CreditAmount) Add(d CreditAmount) (CreditAmount, error) {
	scale := func(v CreditAmount) int {
		s := v.String()
		if n := strings.IndexByte(s, '.'); n >= 0 {
			return len(s) - n - 1
		}
		return 0
	}
	n := scale(c)
	if m := scale(d); m > n {
		n = m
	}
	return canonicalCredit(new(big.Rat).Add(c.rat(), d.rat()), n, 80)
}

// Float64 is for approximate wall-clock rates and meter projections only.
func (c CreditAmount) Float64() float64 { f, _ := c.rat().Float64(); return f }
