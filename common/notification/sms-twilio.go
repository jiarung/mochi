package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	twilioURL = "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json"
	// FIXME(xnum): maybe configurable.
	twilioPhoneBlacklist = mapWithExist{
		"+2348059007341": nil,
		"+447412974963":  nil,
		"+447577663237":  nil,
		"+375295971657":  nil,
		"+375297848843":  nil,
		"+84941156950":   nil,
		"+37127186741":   nil,
		"+886955656737":  nil,
		"+380637137872":  nil,
		"+6281370658851": nil,
		"+584141287352":  nil,
		"+79107037730":   nil,
		"+14159334532":   nil,
		"+79619112729":   nil,
		"+5521985300517": nil,
		"+447424026212":  nil,
	}
)

// TwilioConfig defines config of twilio.
type TwilioConfig struct {
	AccountSID    string
	AuthToken     string
	CountryPhones string
	FromPhoneUS1  string
}

// Twilio wraps twilio service.
type Twilio struct {
	TwilioConfig
	url             string
	countryPhoneMap mapWithExist
}

// NewTwilio creates Twilio.
func NewTwilio(cfg TwilioConfig) *Twilio {
	t := &Twilio{TwilioConfig: cfg,
		countryPhoneMap: make(mapWithExist),
	}
	t.url = fmt.Sprintf(twilioURL, cfg.AccountSID)

	if len(t.CountryPhones) > 0 {
		twilioCountryPhoneConfig := strings.Split(t.CountryPhones, ",")
		for _, c := range twilioCountryPhoneConfig {
			cc := strings.Split(c, ":")
			if len(cc) != 2 {
				panic(fmt.Sprintf("unexpected config: %s", c))
			}
			t.countryPhoneMap[cc[0]] = cc[1]
		}
	}

	return t
}

// Request creates http request.
func (t *Twilio) Request(to, country, msg string) *http.Request {
	var from string
	if t.countryPhoneMap.exist(country) {
		from = t.countryPhoneMap[country].(string)
	} else {
		from = t.FromPhoneUS1
	}

	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Body", msg)
	rb := *strings.NewReader(v.Encode())

	req, _ := http.NewRequest(http.MethodPost, t.url, &rb)
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// IsBlackListPhone returns to is in black list.
func (t *Twilio) IsBlackListPhone(to string) bool {
	return twilioPhoneBlacklist.exist(to)
}
