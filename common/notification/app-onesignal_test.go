package notification

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/cobinhood-backend/types"
)

type AppOnesignalTestSuite struct {
	suite.Suite
	app *AppStruct
}

func (s *AppOnesignalTestSuite) mockOnesignalServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Header check.
			s.Require().Equal("Basic "+s.app.APIKey,
				r.Header.Get("Authorization"), r.Header.Get("Authorization"))
			s.Require().Equal("application/json; charset=utf-8",
				r.Header.Get("Content-Type"))
			// Read body.
			b, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			s.Require().Nil(err)
			var msg struct {
				AppID    string                             `json:"app_id"`
				Headings map[types.OneSignalLanguage]string `json:"headings"`
				Contents map[types.OneSignalLanguage]string `json:"contents"`
				Filters  []struct {
					Field    string `json:"field,omitempty"`
					Key      string `json:"key,omitempty"`
					Relation string `json:"relation,omitempty"`
					Value    string `json:"value,omitempty"`
					Operator string `json:"operator,omitempty"`
				} `json:"filters"`
				Data interface{} `json:"data"`
			}
			s.Require().Nil(json.Unmarshal(b, &msg))
			s.Require().True(onesignalContentsCheck(msg.Headings, msg.Contents))
			// Mock response.
			// Reference: https://goo.gl/gz94cW
			res, err := json.Marshal(struct {
				ID         string      `json:"id"`
				Recipients int         `json:"recipients"`
				Errors     interface{} `json:"errors"`
			}{
				uuid.NewV4().String(),
				1,
				nil,
			})
			s.Require().Nil(err)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(res)
		}))
}

func (s *AppOnesignalTestSuite) testSendOnesignal(appReq *AppRequest) {
	mockServer := s.mockOnesignalServer()
	defer mockServer.Close()
	// Set request url to mock server url.
	oneSignalRestURL = mockServer.URL
	s.Require().Nil(s.app.Send(context.Background(), appReq))
}

func (s *AppOnesignalTestSuite) testBatchSendOnesignal(appReq *BatchAppRequest) {
	mockServer := s.mockOnesignalServer()
	defer mockServer.Close()
	// Set request url to mock server url.
	oneSignalRestURL = mockServer.URL
	_, _, err := s.app.BatchSend(context.Background(), appReq)
	s.Require().Nil(err)
}
