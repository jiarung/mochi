package email

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cobinhood/mochi/types"
)

const tmplVar = `\*\|\w+\|\*`

func TestCheckVar(t *testing.T) {
	assert := assert.New(t)
	assert.Regexp(tmplVar, `*|name|*`)
	assert.Regexp(tmplVar, `*|activeLink|*`)
	assert.NotRegexp(tmplVar, `Name`)
}

func TestExecuteTemplate(t *testing.T) {
	assert := assert.New(t)
	result := executeTemplate("<*|Hello|*>", map[string]string{
		"*|Hello|*": "Hello, World",
	})
	assert.Equal("<Hello, World>", result)
}

func TestUnmarshalRequest(t *testing.T) {
	assert := assert.New(t)

	req, err := GetEmailRequest("APIKeyGeneratedRequest", []byte(`()`))
	assert.NotNilf(err, "%v", req)

	// Simple case.
	e := DepositReminderRequest{
		BaseEmailRequest: BaseEmailRequest{
			FromEmail: "aaa@ddd",
			FromName:  "aaa",
			ToName:    "dst",
			ToEmail:   "dst@ggg",
			Locale:    "US",
		},
	}
	jsonStr := fmt.Sprintf(`{
		"FromEmail":"%s",
		"FromName":"%s",
		"ToName": "%s",
		"ToEmail":"%s",
		"Locale":"%s"}`, e.FromEmail, e.FromName, e.ToName, e.ToEmail, e.Locale)
	var depositInterface Request
	depositInterface, err =
		GetEmailRequest("DepositReminderRequest", []byte(jsonStr))
	assert.Nilf(err, "%v", jsonStr)
	dRequest := depositInterface.(*DepositReminderRequest)
	assert.Equal(dRequest.FromEmail, e.FromEmail)
	assert.Equal(dRequest.FromName, e.FromName)
	assert.Equal(dRequest.ToEmail, e.ToEmail)
	assert.Equal(dRequest.ToName, e.ToName)
	assert.Equal(dRequest.Locale, e.Locale)

	// More complicated case.
	w := WithdrawVerifyRequest{
		BaseEmailRequest: BaseEmailRequest{
			FromEmail: "Qqq@asd",
			FromName:  "Qqq",
			ToName:    "dst2333",
			ToEmail:   "dst2333@qwert.com",
			Locale:    "TW",
		},
		IP:          "1.1.1.1",
		Token:       "10",
		Address:     "0x12345678",
		Currency:    "ETH",
		Amount:      "100",
		FeeCurrency: "NTD",
		FeeAmount:   "10000",
		Memo:        "fake email",
	}
	w.SetDomain("teeeest")

	jsonStr = fmt.Sprintf(`{
		"FromEmail": "%s",
		"FromName": "%s",
		"ToName": "%s",
		"ToEmail": "%s",
		"Locale": "%s",
		"IP": "%s",
		"Token": "%s",
		"Address": "%s",
		"Currency": "%s",
		"Amount": "%s",
		"FeeCurrency": "%s",
		"FeeAmount": "%s",
		"Memo": "%s",
		"Link": "%s"
		}`, w.FromEmail, w.FromName, w.ToName, w.ToEmail, w.Locale,
		w.IP, w.Token, w.Address, w.Currency,
		w.Amount, w.FeeCurrency, w.FeeAmount, w.Memo, w.Link)
	var withdrawalInterface Request
	withdrawalInterface, err =
		GetEmailRequest("WithdrawVerifyRequest", []byte(jsonStr))
	assert.Nilf(err, "%v", jsonStr)
	wRequest := withdrawalInterface.(*WithdrawVerifyRequest)
	assert.Equal(wRequest.FromEmail, w.FromEmail)
	assert.Equal(wRequest.FromName, w.FromName)
	assert.Equal(wRequest.ToEmail, w.ToEmail)
	assert.Equal(wRequest.ToName, w.ToName)
	assert.Equal(wRequest.Locale, w.Locale)
	assert.Equal(wRequest.IP, w.IP)
	assert.Equal(wRequest.Token, w.Token)
	assert.Equal(wRequest.Address, w.Address)
	assert.Equal(wRequest.Currency, w.Currency)
	assert.Equal(wRequest.Amount, w.Amount)
	assert.Equal(wRequest.FeeCurrency, w.FeeCurrency)
	assert.Equal(wRequest.FeeAmount, w.FeeAmount)
	assert.Equal(wRequest.Memo, w.Memo)
	assert.Equal(wRequest.Link, w.Link)

}

func TestGenParameter(t *testing.T) {
	assert := assert.New(t)

	// To check we didn't miss any variable in template.
	for _, req := range Requests {
		// A Hack, to fill required form.
		switch req.(type) {
		case *KYCVerifiedNotifyRequest:
			r := req.(*KYCVerifiedNotifyRequest)
			r.Level = 2
		case *SlotMachineTokenRequest:
			r := req.(*SlotMachineTokenRequest)
			r.RewardType = string(types.DailyTradingReward)
		case *TradingReminderRequest:
			r := req.(*TradingReminderRequest)
			r.TradingPairs = make([]string, 3)
			r.TokenTradingPairs = make([]string, 3)
		case *TradingReminderNoTokenRequest:
			r := req.(*TradingReminderNoTokenRequest)
			r.TradingPairs = make([]string, 3)
		}
		param := req.(Request).Parameter()
		assert.NotNilf(param, "%T", req)
		assert.NotRegexp(tmplVar, param.Template)
	}
}
