package decimal

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func (u *UtilsTestSuite) TestRoundUp() {
	testcases := []struct {
		test, base, result float64
	}{
		{123.4567, 0.001, 123.457},
		{123.4567, 0.005, 123.460},
		{123.4567, 0.01, 123.46},
		{123.4567, 0.05, 123.50},
		{123.4567, 0.1, 123.5},
		{123.4567, 0.5, 123.5},
		{123.4567, 1.0, 124},
		{123.4567, 5.0, 125},
		{123.4567, 10.0, 130},
		{123.4567, 50.0, 150},
		{123.4567, 100.0, 200},
	}

	for i := 0; i < len(testcases); i++ {
		v := testcases[i]
		c := RoundUp(
			decimal.NewFromFloat(v.test),
			decimal.NewFromFloat(v.base),
		)
		require.True(
			u.T(),
			c.Equal(
				decimal.NewFromFloat(v.result),
			),
			"Index: %v, %v != %v\n",
			i,
			c,
			v.result,
		)
	}
}

func (u *UtilsTestSuite) TestRoundDown() {
	testcases := []struct {
		test, base, result float64
	}{
		{123.4567, 0.001, 123.456},
		{123.4567, 0.005, 123.455},
		{123.4567, 0.01, 123.45},
		{123.4567, 0.05, 123.45},
		{123.4567, 0.1, 123.4},
		{123.4567, 0.5, 123.0},
		{123.4567, 1.0, 123.0},
		{123.4567, 5.0, 120.0},
		{123.4567, 10.0, 120.0},
		{123.4567, 50.0, 100.0},
		{123.4567, 100.0, 100.0},
	}

	for i := 0; i < len(testcases); i++ {
		v := testcases[i]
		c := RoundDown(
			decimal.NewFromFloat(v.test),
			decimal.NewFromFloat(v.base),
		)
		require.True(
			u.T(),
			c.Equal(
				decimal.NewFromFloat(v.result),
			),
			"Index: %v, %v != %v\n",
			i,
			c,
			v.result,
		)
	}
}

func (u *UtilsTestSuite) TestGetRoundPrecision() {
	testcases := []struct {
		price float64
		steps []float64
	}{
		{
			0.001,
			[]float64{
				0.001, 0.005, 0.01, 0.05,
				0.1, 0.5, 1.0, 5.0, 10.0, 50.0, 100.0,
			},
		},
		{
			0.03,
			[]float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 50.0, 100.0},
		},
		{
			0.1,
			[]float64{0.1, 0.5, 1.0, 5.0, 10.0, 50.0, 100.0},
		},
		{
			3.0,
			[]float64{1.0, 5.0, 10.0, 50.0, 100.0},
		},
		{
			5.0,
			[]float64{1.0, 5.0, 10.0, 50.0, 100.0},
		},
		{
			20.0,
			[]float64{10.0, 50.0, 100.0},
		},
		{
			500.0,
			[]float64{100.0},
		},
		{
			// margin case, we ensure at least one level aggregation
			1200.0,
			[]float64{1000.0},
		},
	}
	for i := 0; i < len(testcases); i++ {
		v := testcases[i]
		steps := GetRoundPrecisions(
			decimal.NewFromFloat(v.price),
		)
		require.Equal(
			u.T(), len(v.steps), len(steps),
			"Index: %v. Steps number is wrong: %v != %v",
			i,
			v.steps,
			steps,
		)

		for j := 0; j < len(steps); j++ {
			c := decimal.NewFromFloat(v.steps[j])
			require.True(
				u.T(),
				steps[j].Equal(c),
				"Index: %v, %v != %v. steps: %v != %v\n",
				i,
				c,
				steps[j],
				v.steps,
				steps,
			)
		}
	}
}

func (u *UtilsTestSuite) TestComparator() {
	a := decimal.NewFromFloat(10)
	b := decimal.NewFromFloat(20)
	u.Require().Equal(1, Comparator(b, a))
	u.Require().Equal(0, Comparator(b, b))
	u.Require().Equal(-1, Comparator(a, b))
	u.Require().Equal(1, InvertedComparator(a, b))
	u.Require().Equal(0, InvertedComparator(a, a))
	u.Require().Equal(-1, InvertedComparator(b, a))
}

func TestUtils(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func BenchmarkRoundUp(b *testing.B) {
	d, base := decimal.NewFromFloat(0.0237), decimal.NewFromFloat(0.005)
	for i := 0; i < b.N; i++ {
		RoundUp(d, base)
	}
	b.ReportAllocs()
}

func BenchmarkRoundDownAwayFromZero(b *testing.B) {
	d, base := decimal.NewFromFloat(0.0237), decimal.NewFromFloat(0.05)
	for i := 0; i < b.N; i++ {
		RoundDownAwayFromZero(d, base)
	}
	b.ReportAllocs()
}
