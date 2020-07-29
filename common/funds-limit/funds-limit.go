// Code generated by cobctl. This go file is generated automatically, DO NOT EDIT.

package fundslimit

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Unlimited returns unlimited values in database (-1)
func Unlimited() decimal.Decimal {
	return decimal.NewFromFloat(-1)
}

var dailyFiatDepositLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 4000,
	3: 15000,
}

// GetDailyFiatDepositLimit returns the daily fiat deposit limit of kycLevel
func GetDailyFiatDepositLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := dailyFiatDepositLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var dailyFiatWithdrawalLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 4000,
	3: 15000,
}

// GetDailyFiatWithdrawalLimit returns the daily fiat withdrawal limit of kycLevel
func GetDailyFiatWithdrawalLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := dailyFiatWithdrawalLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var monthlyFiatDepositLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 60000,
	3: 200000,
}

// GetMonthlyFiatDepositLimit returns the monthly fiat deposit limit of kycLevel
func GetMonthlyFiatDepositLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := monthlyFiatDepositLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var monthlyFiatWithdrawalLimit = map[int]float64{
	0: 0,
	1: 0,
	2: 60000,
	3: 200000,
}

// GetMonthlyFiatWithdrawalLimit returns the monthly fiat withdrawal limit of kycLevel
func GetMonthlyFiatWithdrawalLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := monthlyFiatWithdrawalLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var dailyCryptoDepositLimit = map[int]float64{
	0: -1,
	1: -1,
	2: -1,
	3: -1,
}

// GetDailyCryptoDepositLimit returns the daily crypto deposit limit of kycLevel
func GetDailyCryptoDepositLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := dailyCryptoDepositLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var dailyCryptoWithdrawalLimit = map[int]float64{
	0: 0,
	1: 3,
	2: 100,
	3: 100,
}

// GetDailyCryptoWithdrawalLimit returns the daily crypto withdrawal limit of kycLevel
func GetDailyCryptoWithdrawalLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := dailyCryptoWithdrawalLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var monthlyCryptoDepositLimit = map[int]float64{
	0: -1,
	1: -1,
	2: -1,
	3: -1,
}

// GetMonthlyCryptoDepositLimit returns the monthly crypto deposit limit of kycLevel
func GetMonthlyCryptoDepositLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := monthlyCryptoDepositLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}

var monthlyCryptoWithdrawalLimit = map[int]float64{
	0: 0,
	1: 90,
	2: 3000,
	3: 3000,
}

// GetMonthlyCryptoWithdrawalLimit returns the monthly crypto withdrawal limit of kycLevel
func GetMonthlyCryptoWithdrawalLimit(kycLevel int) (decimal.Decimal, error) {
	limit, ok := monthlyCryptoWithdrawalLimit[kycLevel]
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected kyc level [%d]", kycLevel)
	}
	return decimal.NewFromFloat(limit), nil
}
