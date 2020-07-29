package notification

import (
	"net/http"
	"net/url"
	"strings"
)

var (
	nexmoAPIUrl = "https://rest.nexmo.com/sms/json"
)

// NexmoConfig defines config of nexmo.
type NexmoConfig struct {
	APIKey          string
	APISecret       string
	CountryCodeList string
}

// Nexmo wraps nexmo service.
type Nexmo struct {
	NexmoConfig
	countryCodeMap mapWithExist
}

// NewNexmo creates nexmo.
func NewNexmo(cfg NexmoConfig) *Nexmo {
	n := &Nexmo{
		NexmoConfig:    cfg,
		countryCodeMap: make(mapWithExist),
	}

	CountryCodeList := strings.Split(n.CountryCodeList, ",")
	for _, element := range CountryCodeList {
		n.countryCodeMap[element] = nil
	}
	return n
}

// Request creates http request.
func (n *Nexmo) Request(to, from, msg string) *http.Request {
	v := url.Values{}
	v.Set("api_key", n.APIKey)
	v.Set("api_secret", n.APISecret)
	v.Set("from", from)
	v.Set("to", to)
	v.Set("text", msg)
	rb := *strings.NewReader(v.Encode())

	req, _ := http.NewRequest(http.MethodPost, nexmoAPIUrl, &rb)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// IsSupportCountry returns the country is supported.
func (n *Nexmo) IsSupportCountry(toCountry string) bool {
	return n.countryCodeMap.exist(toCountry)
}
