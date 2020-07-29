package decimal

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// AssertEqual provides a utility for tests to compare 2 decimals.
// Equivalent to assert.True(suite.T(), expected.Equal(actual)).
func AssertEqual(t *testing.T, expected, actual decimal.Decimal) bool {
	return assert.True(t, expected.Equal(actual), "%s (actual) ≠ %s (expected)",
		actual, expected)
}

// RequireEqual provides a utility for tests to compare 2 decimals.
// Equivalent to suite.Require().True(expected.Equal(actual)).
func RequireEqual(t *testing.T, expected, actual decimal.Decimal) {
	if !AssertEqual(t, expected, actual) {
		t.FailNow()
	}
}
