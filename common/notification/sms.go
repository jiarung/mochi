package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	mathRand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jiarung/mochi/common/config/secret"
	"github.com/jiarung/mochi/common/config/thirdparty"
)

type mapWithExist map[string]interface{}

func (m mapWithExist) exist(key string) (exist bool) {
	_, exist = m[key]
	return
}

// SMSConfig defines config of sms.
type SMSConfig struct {
	TwilioEnabled bool
	TwilioConfig

	NexmoEnabled bool
	NexmoConfig
}

// SMSStruct wraps sms struct.
type SMSStruct struct {
	SMSConfig

	twilio *Twilio

	nexmo *Nexmo
}

// SMS returns an inited SMSSturct.
func SMS() *SMSStruct {
	// FIXME(xnum): poor design.
	return NewSMS(SMSConfig{
		TwilioEnabled: true,
		TwilioConfig: TwilioConfig{
			AccountSID:    thirdparty.TwilioAccountSid(),
			AuthToken:     secret.Get("TWILIO_AUTH_TOKEN"),
			CountryPhones: thirdparty.TwilioCountryPhone(),
			FromPhoneUS1:  thirdparty.TwilioFromPhoneUs1(),
		},
		NexmoEnabled: true,
		NexmoConfig: NexmoConfig{
			APIKey:          thirdparty.NexmoApiKey(),
			APISecret:       secret.Get("NEXMO_API_SECRET"),
			CountryCodeList: thirdparty.NexmoCountryCodeList(),
		},
	})
}

// NewSMS creates SMS.
func NewSMS(cfg SMSConfig) *SMSStruct {
	s := &SMSStruct{SMSConfig: cfg}
	s.twilio = NewTwilio(cfg.TwilioConfig)
	s.nexmo = NewNexmo(cfg.NexmoConfig)
	return s
}

// SendTo sends SMS using Twilio API and return the response.
func (s *SMSStruct) SendTo(toCountry, toPhoneNum, msg string) (err error) {
	if !s.TwilioEnabled && !s.NexmoEnabled {
		return
	}

	if !strings.HasPrefix(toCountry, "+") {
		toCountry = "+" + toCountry
	}
	to := toCountry + toPhoneNum

	if s.NexmoEnabled {
		if s.nexmo.IsSupportCountry(toCountry) {
			return s.sendSMSNexmo(to, msg)
		}

		if s.twilio.IsBlackListPhone(to) {
			return s.sendSMSNexmo(to, msg)
		}

		// for analysing Nexmo's stability
		mathRand.Seed(time.Now().UnixNano())
		if mathRand.Intn(100) < 10 {
			return s.sendSMSNexmo(to, msg)
		}
	}

	if !s.TwilioEnabled {
		return errors.New("parameter is not suitable for sms service")
	}

	client := &http.Client{}
	resp, err := client.Do(s.twilio.Request(to, toCountry, msg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		e := decoder.Decode(&data)
		if e != nil {
			return fmt.Errorf("parse response error: %+v", e)
		}
		return fmt.Errorf("Get error response: %+v", data)
	}

	_, err = ioutil.ReadAll(resp.Body)
	return
}

func (s *SMSStruct) sendSMSNexmo(to string, msg string) (err error) {
	client := &http.Client{}
	// FIXME(xnum): what's cobsms?
	resp, err := client.Do(s.nexmo.Request(to, "cobsms", msg))
	if err != nil {
		return fmt.Errorf("send sms error phoneNum<%s>: %v", to, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	type smsResp struct {
		Messages []struct {
			Status string `json:"status"`
		} `json:"messages"`
	}
	res := smsResp{}
	json.Unmarshal(bodyBytes, &res)
	if len(res.Messages) <= 0 || res.Messages[0].Status != "0" {
		err = fmt.Errorf("send sms error res: %s", string(bodyBytes))
		return
	}

	return
}
