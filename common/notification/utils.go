package notification

import (
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/jiarung/mochi/cache/instances"
	models "github.com/jiarung/mochi/models/exchange"
)

func btcToSatoshi(btc decimal.Decimal) decimal.Decimal {
	return btc.Mul(decimal.New(1, 8))
}

func orderSatoshi(order *models.Order) (decimal.Decimal, error) {
	amount := order.Size
	pair := strings.Split(order.TradingPairID, "-")
	if len(pair) != 2 {
		return decimal.Zero, fmt.Errorf("tradingPairID wrong format %v", pair)
	}
	exchangeRate, err := instances.GetExchangeRate(pair[0])
	if err != nil {
		return decimal.Zero, err
	}
	amountBTC := amount.Mul(exchangeRate.PriceBTC)
	return btcToSatoshi(amountBTC), nil
}

// decimalToFloat doesn't care whether float exactly represent decimal, we just need a approximate value
func decimalToFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// GenerateRandomCode generate sms fix num random code
func GenerateRandomCode() (string, error) {
	code, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", code), nil
}
