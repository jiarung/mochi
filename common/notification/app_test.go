package notification

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/models/exchange"
	"github.com/jiarung/mochi/types"
)

type AppTestSuite struct {
	AppOnesignalTestSuite
}

func (s *AppTestSuite) SetupSuite() {
	s.app = &AppStruct{
		OneSignalConfig: OneSignalConfig{
			APIKey: "ABC",
			AppID:  "DEF",
		},
		Tag: AppTag{
			TagSecret: "ABABABAB",
		},
		Service: OnesignalSymbol,
	}
}

func (s *AppTestSuite) genTestBatchRequest() *BatchAppRequest {
	return &BatchAppRequest{
		Subjects: map[types.OneSignalLanguage]string{
			types.English: "test",
		},
		Contents: map[types.OneSignalLanguage]string{
			types.English: "test",
		},
		Filters: []map[string]string{
			map[string]string{
				"field":    "tag",
				"key":      "notification_tag",
				"relation": "=",
				"value": s.app.Tag.GetNotificationTag(uuid.NewV4().String(),
					cobxtypes.APICobx),
			},
			map[string]string{
				"operator": "OR",
			},
			map[string]string{
				"field":    "tag",
				"key":      "notification_tag",
				"relation": "=",
				"value": s.app.Tag.GetNotificationTag(uuid.NewV4().String(),
					cobxtypes.APICobx),
			},
			map[string]string{
				"operator": "OR",
			},
			map[string]string{
				"field":    "tag",
				"key":      "notification_tag",
				"relation": "=",
				"value": s.app.Tag.GetNotificationTag(uuid.NewV4().String(),
					cobxtypes.APICobx),
			},
		},
	}
}

func (s *AppTestSuite) TestSend() {
	var (
		appReq *AppRequest
		err    error
	)
	// Deposit confirmed.
	appReq, err = s.app.GenDepositConfirmed(uuid.NewV4(),
		&exchange.Deposit{}, cobxtypes.APICobx)
	s.Require().Nil(err)
	s.testSendOnesignal(appReq)
	// Withdrawal confirmed.
	appReq, err = s.app.GenWithdrawalConfirmed(uuid.NewV4(),
		&exchange.Withdrawal{}, cobxtypes.APICobx)
	s.Require().Nil(err)
	s.testSendOnesignal(appReq)
	// Order update.
	appReq, err = s.app.GenOrderFilled(exchange.Order{UserID: uuid.NewV4()})
	s.Require().Nil(err)
	s.testSendOnesignal(appReq)
	// Price alert.
	appReq, err = s.app.GenPriceAlert(uuid.NewV4(), "ETH-USDT",
		decimal.NewFromFloat(100))
	s.Require().Nil(err)
	s.testSendOnesignal(appReq)
	// Referral mission completed.
	appReq, err = s.app.GenReferralMissionCompleted(
		exchange.User{ID: uuid.NewV4()})
	s.Require().Nil(err)
	s.testSendOnesignal(appReq)
}

func (s *AppTestSuite) TestBatchSend() {
	s.testBatchSendOnesignal(s.genTestBatchRequest())
}

func TestApp(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
