package fundslimit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cdecimal "github.com/cobinhood/mochi/common/decimal"
)

type FundsLimitTestSuite struct {
	suite.Suite
}

var testDailyCryptoDepositLimit = map[int]float64{
	0: -1,
	1: -1,
	2: -1,
	3: -1,
}

func (s FundsLimitTestSuite) TestDailyCryptoDepositLimit() {
	for level, limit := range testDailyCryptoDepositLimit {
		result, err := GetDailyCryptoDepositLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testMonthlyFiatDepositLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 60000,
	3: 200000,
}

func (s FundsLimitTestSuite) TestMonthlyFiatDepositLimit() {
	for level, limit := range testMonthlyFiatDepositLimit {
		result, err := GetMonthlyFiatDepositLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testMonthlyCryptoDepositLimit = map[int]float64{
	0: -1,
	1: -1,
	2: -1,
	3: -1,
}

func (s FundsLimitTestSuite) TestMonthlyCryptoDepositLimit() {
	for level, limit := range testMonthlyCryptoDepositLimit {
		result, err := GetMonthlyCryptoDepositLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testMonthlyCryptoWithdrawalLimit = map[int]float64{
	0: 0,
	1: 90,
	2: 3000,
	3: 3000,
}

func (s FundsLimitTestSuite) TestMonthlyCryptoWithdrawalLimit() {
	for level, limit := range testMonthlyCryptoWithdrawalLimit {
		result, err := GetMonthlyCryptoWithdrawalLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testDailyFiatDepositLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 4000,
	3: 15000,
}

func (s FundsLimitTestSuite) TestDailyFiatDepositLimit() {
	for level, limit := range testDailyFiatDepositLimit {
		result, err := GetDailyFiatWithdrawalLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testMonthlyFiatWithdrawalLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 60000,
	3: 200000,
}

func (s FundsLimitTestSuite) TestMonthlyFiatWithdrawalLimit() {
	for level, limit := range testMonthlyFiatWithdrawalLimit {
		result, err := GetMonthlyFiatWithdrawalLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testDailyFiatWithdrawalLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 4000,
	3: 15000,
}

func (s FundsLimitTestSuite) TestDailyFiatWithdrawalLimit() {
	for level, limit := range testDailyFiatWithdrawalLimit {
		result, err := GetDailyFiatWithdrawalLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

var testDailyCryptoWithdrawalLimit = map[int]float64{
	0: 0,
	1: 3,
	2: 100,
	3: 100,
}

func (s FundsLimitTestSuite) TestDailyCryptoWithdrawalLimit() {
	for level, limit := range testDailyCryptoWithdrawalLimit {
		result, err := GetDailyCryptoWithdrawalLimit(level)
		require.Nil(s.T(), err)
		cdecimal.RequireEqual(s.T(), decimal.NewFromFloat(limit), result)
	}
}

func TestFundsLimit(t *testing.T) {
	suite.Run(t, new(FundsLimitTestSuite))
}
