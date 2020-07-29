package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	models "github.com/jiarung/mochi/models/exchange"
)

const (
	gaEndPoint = "https://www.google-analytics.com/collect"
)

// GAStruct provides namespace for google ananlytics functions.
type GAStruct struct {
	// thirdparty.GaTrackingId())
	TrackingID string `config:"GaTrackingId"`
}

// ServiceName returns its concrete service name.
func (ga *GAStruct) ServiceName() string {
	return "Google Analytics"
}

// GenSignUp create signup event data for Google Analytic
// For more infomation about Google Measurement Protocol,
// please reference to
// https://developers.google.com/analytics/devguides/collection/protocol/v1/reference
func (ga *GAStruct) GenSignUp(registration *models.Registration) *url.Values {
	v := url.Values{}
	v.Set("t", "event")
	v.Set("v", "1")
	v.Set("uid", registration.UserID.String())
	v.Set("tid", ga.TrackingID)
	v.Set("ec", "Account")
	v.Set("ea", "SignUp")
	return &v
}

// GenDeposited create deposited event data for Google Analytics
func (ga *GAStruct) GenDeposited(deposit *models.Deposit) *url.Values {
	v := url.Values{}
	v.Set("t", "event")
	v.Set("v", "1")
	v.Set("uid", deposit.UserID.String())
	v.Set("tid", ga.TrackingID)
	v.Set("ec", "Balance")
	v.Set("ea", "Deposited")
	v.Set("ev", btcToSatoshi(deposit.BTCValue).StringFixed(0))
	return &v
}

// GenOrderCompleted create order completed event data for Google Analytics
func (ga *GAStruct) GenOrderCompleted(order *models.Order) (*url.Values, error) {
	v := url.Values{}
	v.Set("t", "event")
	v.Set("v", "1")
	v.Set("uid", order.UserID.String())
	v.Set("tid", ga.TrackingID)
	v.Set("ec", "Order")
	v.Set("ea", "Completed")
	satoshi, err := orderSatoshi(order)
	if err != nil {
		return nil, fmt.Errorf("get satoshi from order fail err %v", err)
	}
	v.Set("ev", satoshi.StringFixed(0))
	return &v, nil
}

// Send send data to Google Analytics
func (ga *GAStruct) Send(ctx context.Context, v *url.Values) error {
	client := &http.Client{}
	request, err := http.NewRequest(
		http.MethodPost,
		gaEndPoint,
		strings.NewReader(v.Encode()),
	)

	if err != nil {
		return fmt.Errorf("create http.NewRequest fail. err: %v", err)
	}

	resp, err := client.Do(request.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("Send fail err %v response %v", err, resp)
	}

	return nil
}
