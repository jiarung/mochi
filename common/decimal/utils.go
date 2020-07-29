package decimal

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// GetRoundPrecisions returns round steps to aggregate orderbook for reducing data
// transfer
func GetRoundPrecisions(price decimal.Decimal) []decimal.Decimal {
	modBase := price
	if price.LessThan(decimal.NewFromFloat(1.0)) {
		modBase = decimal.New(1, int32(-GetPrecision(price)))
	} else {
		modBase = decimal.New(1, int32(len(fmt.Sprint(price.IntPart()))-1))
	}

	step1, step2 := decimal.NewFromFloat(5.0), decimal.NewFromFloat(2.0)
	if !modBase.Add(modBase.Mul(decimal.NewFromFloat(4.0))).Equal(modBase.Mul(step1)) {
		step1, step2 = step2, step1
	}

	steps := []decimal.Decimal{modBase}
	for modBase.LessThan(decimal.NewFromFloat(100.0)) {
		modBase = modBase.Mul(step1)
		newV, _ := decimal.NewFromString(modBase.String()) // remove hidden exponent info
		steps = append(steps, newV)
		step1, step2 = step2, step1
	}
	return steps
}

// RoundUp given `d` according to base
func RoundUp(d, base decimal.Decimal) decimal.Decimal {
	sign := base.Sign()

	if sign == 0 {
		return decimal.Zero
	}

	r := d.Mod(base)
	if r.Sign() == 0 {
		return d
	} else if sign > 0 {
		return d.Sub(r).Add(base)
	} else {
		return d.Sub(r).Sub(base)
	}
}

// RoundDown given `d` according to base
func RoundDown(d, base decimal.Decimal) decimal.Decimal {
	sign := base.Sign()

	if sign == 0 {
		return decimal.Zero
	}
	return d.Sub(d.Mod(base))
}

// RoundDownAwayFromZero given `d` away from zero if necessary.
func RoundDownAwayFromZero(d, base decimal.Decimal) decimal.Decimal {
	v := RoundDown(d, base)
	if v.Sign() == 0 {
		return base
	}
	return v
}

// GetScientificNotation string
func GetScientificNotation(d decimal.Decimal) string {
	if d.Exponent() == 0 {
		base := len(fmt.Sprint(d.IntPart())) - 1
		div := decimal.New(1, int32(base))
		return fmt.Sprintf("%vE%v", d.Div(div), base)
	}
	return fmt.Sprintf("%vE%v", d.Coefficient(), d.Exponent())
}

// GetPrecision returns the precision of the decimal.
func GetPrecision(d decimal.Decimal) int {
	s := d.String()
	idx := strings.IndexRune(s, '.')
	if idx == -1 {
		return 0
	}
	return len(s) - idx - 1
}

// Comparator is a decimal comparator for gods.utils.comparator.Comparator.
// It is equivalent to a.Cmp(b). Useful for sorting in ascending order.
// If any of given input is not decimal, it will panic in type assertion.
func Comparator(a, b interface{}) int {
	aAsserted := a.(decimal.Decimal)
	bAsserted := b.(decimal.Decimal)

	return aAsserted.Cmp(bAsserted)
}

// InvertedComparator is a decimal comparator for gods.utils.comparator.Comparator.
// It is equivalent to b.Cmp(a). Useful for sorting in descending order.
// If any of given input is not decimal, it will panic in type assertion.
func InvertedComparator(a, b interface{}) int {
	aAsserted := a.(decimal.Decimal)
	bAsserted := b.(decimal.Decimal)

	return bAsserted.Cmp(aAsserted)
}
