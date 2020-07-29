package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	jsonBuilder "github.com/cobinhood/mochi/common/encode/json"
	models "github.com/cobinhood/mochi/models/exchange"
)

const (
	tagName = "from_server"
)

// FBConfig defines config of FB.
type FBConfig struct {
	OfflineEventSetID string `config:"FbOfflineEventSetId"`

	AccessToken string `config:"FbBusinessAccessToken"`
}

// FBStruct provides namespace for facebook functions.
type FBStruct struct {
	cfg      FBConfig
	endpoint string
}

// NewFB returns a FBStruct.
func NewFB(cfg FBConfig) *FBStruct {
	fb := &FBStruct{cfg: cfg}
	fb.endpoint = fmt.Sprintf("https://graph.facebook.com/v3.1/%s/events",
		cfg.OfflineEventSetID)
	return fb
}

// ServiceName returns its concrete service name.
func (fb *FBStruct) ServiceName() string {
	return "Facebook offline conversion"
}

// GenSignUp create signup event data for Facebook offline conversion
func (fb *FBStruct) GenSignUp(
	registration *models.Registration) (*url.Values, error) {
	v := url.Values{}
	v.Set("upload_tag", tagName)
	v.Set("access_token", fb.cfg.AccessToken)

	s, err := jsonBuilder.Array(
		jsonBuilder.Object(
			jsonBuilder.Attr("event_time", time.Now().Unix()),
			jsonBuilder.Attr("event_name", "CompleteRegistration"),
			jsonBuilder.Attr("match_keys", jsonBuilder.Object(
				jsonBuilder.Attr("extern_id", registration.UserID.String()),
			)),
		),
	).Marshal()

	if err != nil {
		return nil, err
	}
	v.Set("data", string(s))
	return &v, nil
}

// GenDeposited create deposited event data for Facebook offline conversion
func (fb *FBStruct) GenDeposited(
	deposit *models.Deposit) (*url.Values, error) {
	v := url.Values{}
	v.Set("upload_tag", tagName)
	v.Set("access_token", fb.cfg.AccessToken)

	s, err := jsonBuilder.Array(
		jsonBuilder.Object(
			jsonBuilder.Attr("event_time", time.Now().Unix()),
			// fb only accept nine predefined event_name:
			// "ViewContent"、"Search"、"AddToCart"、"AddToWishlist"、"InitiateCheckout"、
			// "AddPaymentInfo"、"Purchase"、"Lead"、"CompleteRegistration"
			// but "Deposited" isn't involved
			// so we should use "Other" as event_name and
			// set the event_name we want in custom_data
			jsonBuilder.Attr("event_name", "Other"),
			jsonBuilder.Attr("match_keys", jsonBuilder.Object(
				jsonBuilder.Attr("extern_id", deposit.UserID.String()),
			)),
			jsonBuilder.Attr("custom_data", jsonBuilder.Object(
				jsonBuilder.Attr("custom_event_name", "Deposited"),
				// Why do we put "value" field on custom data ?
				// Because if we put it on higher level, fb will complain
				// that we should add "currency" field, but fb doesn't accept "BTC"
				// (It only accepts ISO 4217 currency code).
				jsonBuilder.Attr("value", decimalToFloat(btcToSatoshi(deposit.BTCValue))),
			)),
		),
	).Marshal()

	if err != nil {
		return nil, err
	}
	v.Set("data", string(s))
	return &v, nil
}

// GenOrderCompleted create order completed event data for Facebook offline conversion
func (fb *FBStruct) GenOrderCompleted(
	order *models.Order) (*url.Values, error) {
	v := url.Values{}
	v.Set("upload_tag", tagName)
	v.Set("access_token", fb.cfg.AccessToken)

	satoshi, err := orderSatoshi(order)
	if err != nil {
		return nil, fmt.Errorf("get satoshi from order fail err %v", err)
	}

	s, err := jsonBuilder.Array(
		jsonBuilder.Object(
			jsonBuilder.Attr("event_time", time.Now().Unix()),
			jsonBuilder.Attr("event_name", "Other"),
			jsonBuilder.Attr("match_keys", jsonBuilder.Object(
				jsonBuilder.Attr("extern_id", order.UserID.String()),
			)),
			jsonBuilder.Attr("custom_data", jsonBuilder.Object(
				jsonBuilder.Attr("custom_event_name", "OrderCompleted"),
				jsonBuilder.Attr("value", decimalToFloat(satoshi)),
			)),
		),
	).Marshal()

	if err != nil {
		return nil, fmt.Errorf("marshal: %v", err)
	}
	v.Set("data", string(s))
	return &v, nil
}

// Send send data to Facebook offline conversion
func (fb *FBStruct) Send(ctx context.Context, v *url.Values) error {
	client := &http.Client{}
	request, err := http.NewRequest(
		http.MethodPost,
		fb.endpoint,
		nil,
	)

	if err != nil {
		return fmt.Errorf("create http.NewRequest fail. err: %v", err)
	}

	request.URL.RawQuery = v.Encode()
	resp, err := client.Do(request.WithContext(ctx))

	if err != nil {
		return fmt.Errorf("Send fail. err: %v", err)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var bodyJSON map[string]interface{}
	err = json.Unmarshal(bodyBytes, &bodyJSON)

	if err != nil {
		return fmt.Errorf("Send fail. cannot parse response. err: %v", err)
	} else if bodyJSON["error"] != nil {
		return fmt.Errorf("Send fail. response: %v", string(bodyBytes))
	}
	return nil
}
