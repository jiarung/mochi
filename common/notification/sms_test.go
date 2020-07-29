package notification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/cobinhood-backend/common/config/misc"
	"github.com/cobinhood/cobinhood-backend/common/utils"
)

type SMSTestSuite struct {
	suite.Suite
	sms *SMSStruct
	cfg SMSConfig
}

func (s *SMSTestSuite) mockNexmoServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Header check.
			s.Require().Equal("application/x-www-form-urlencoded",
				r.Header.Get("Content-Type"))
			// Parse form.
			s.Require().Nil(r.ParseForm())
			s.Require().Equal(s.cfg.APIKey, r.Form.Get("api_key"))
			s.Require().Equal(s.cfg.APISecret, r.Form.Get("api_secret"))
			s.Require().NotEqual("", r.Form.Get("from"))
			s.Require().NotEqual("", r.Form.Get("to"))
			s.Require().NotEqual("", r.Form.Get("text"))
			// Mock response.
			// Reference: https://developer.nexmo.com/api/sms
			// Error code: https://developer.nexmo.com/api/sms#errors
			res, err := json.Marshal(struct {
				MessageCount int         `json:"message-count"`
				Messages     interface{} `json:"messages"`
			}{
				1,
				[]struct {
					To               string `json:"to"`
					MessageID        string `json:"message-id"`
					Status           string `json:"status"`
					RemainingBalance string `json:"remaining-balance"`
					MessagePrice     string `json:"message-price"`
					Network          string `json:"network"`
				}{
					{
						"447700900000",
						"0A0000000123ABCD1",
						"0",
						"3.14159265",
						"0.03330000",
						"12345",
					},
				},
			})
			s.Require().Nil(err)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(res)
		}))
}

func (s *SMSTestSuite) mockTwilioServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Header check.
			s.Require().Equal("application/x-www-form-urlencoded",
				r.Header.Get("Content-Type"))
			// Basic auth.
			sid, token, ok := r.BasicAuth()
			s.Require().True(ok)
			s.Require().Equal(s.cfg.AccountSID, sid)
			s.Require().Equal(s.cfg.AuthToken, token)
			// Parse form.
			s.Require().Nil(r.ParseForm())
			s.Require().NotEqual("", r.Form.Get("From"))
			s.Require().NotEqual("", r.Form.Get("To"))
			s.Require().NotEqual("", r.Form.Get("Body"))
			// Mock response.
			// Example: https://www.twilio.com/blog/2014/06/sending-sms-from-your-go-app.html
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
		}))
}

func (s *SMSTestSuite) TestSendTo() {
	s.cfg = SMSConfig{
		TwilioEnabled: true,
		TwilioConfig: TwilioConfig{
			AccountSID:    "ABCD",
			AuthToken:     "EDCRFV",
			CountryPhones: "+886:987",
			FromPhoneUS1:  "+1:111",
		},
		NexmoEnabled: true,
		NexmoConfig: NexmoConfig{
			APIKey:          "AAA",
			APISecret:       "CDE",
			CountryCodeList: "886,1",
		},
	}
	s.sms = NewSMS(s.cfg)
	env := misc.ServerEnvironment()
	misc.SetServerEnvironment(utils.EnvDevelopmentTag)

	// Mock nexmo server.
	mockNexmoServer := s.mockNexmoServer()
	defer mockNexmoServer.Close()
	nexmoAPIUrl = mockNexmoServer.URL

	// Mock twilio server.
	mockTwilioServer := s.mockTwilioServer()
	defer mockTwilioServer.Close()
	twilioURL = mockTwilioServer.URL

	// Test nexmo.
	err := s.sms.SendTo("91", "123", "test")
	s.Require().Nil(err)

	// Test twilio.
	err = s.sms.SendTo("886", "123", "test")
	s.Require().Nil(err)

	misc.SetServerEnvironment(env)
}

func TestSMS(t *testing.T) {
	suite.Run(t, new(SMSTestSuite))
}
