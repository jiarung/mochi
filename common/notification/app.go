package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/common/utils"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/types"
)

// AppStruct defines struct of app.
type AppStruct struct {
	OneSignalConfig
	Tag     AppTag
	Service string `config:"AppNotificationService"`
}

// AppRequest defines the sending message which generates from template.
type AppRequest struct {
	ID         string
	Subject    map[types.OneSignalLanguage]string
	Message    map[types.OneSignalLanguage]string
	UserID     string
	ReturnData interface{}
}

// BatchAppRequest defines the batch sending message.
type BatchAppRequest struct {
	Subjects   map[types.OneSignalLanguage]string
	Contents   map[types.OneSignalLanguage]string
	ReturnData interface{}
	URL        *string
	Filters    []map[string]string
	Options    []types.DevicePlatform

	Schedule *struct {
		SendAfter   time.Time
		IsOptimized bool
	}
}

// ServiceName returns its concrete service name.
func (a *AppStruct) ServiceName() string {
	return a.Service
}

// Send calls push notification service with req.
func (a *AppStruct) Send(ctx context.Context, req *AppRequest) (err error) {
	switch utils.Environment() {
	case utils.Production:
	case utils.Staging:
		for k := range req.Subject {
			req.Subject[k] += " [staging]"
		}
	case utils.Development:
		for k := range req.Subject {
			req.Subject[k] += " [dev]"
		}
	case utils.LocalDevelopment:
		for k := range req.Subject {
			req.Subject[k] += " [localdev]"
		}
		return
	case utils.CI, utils.Stress:
		return
	}

	switch a.Service {
	case OnesignalSymbol:
		return a.OneSignalConfig.Send(ctx, req)
	default:
		return errors.New("Invalid app service type")
	}
}

// GenDepositConfirmed generates deposit confirmed message and packs into
// return value.
func (a *AppStruct) GenDepositConfirmed(userID uuid.UUID,
	deposit *models.Deposit, service cobxtypes.ServiceName) (*AppRequest, error) {
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(userID.String(), service),
		Subject: map[types.OneSignalLanguage]string{
			types.English: "Deposit Notice",
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Your %s %s deposit has been received!",
				deposit.Amount.String(),
				deposit.CurrencyID),
		},
		UserID: userID.String(),
		ReturnData: map[string]string{
			"target": "/wallet",
		},
	}

	return req, nil
}

// GenWithdrawalConfirmed generates withdraw confirmed message and packs into
// return value.
func (a *AppStruct) GenWithdrawalConfirmed(userID uuid.UUID,
	withdrawal *models.Withdrawal, service cobxtypes.ServiceName) (
	*AppRequest, error) {
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(userID.String(), service),
		Subject: map[types.OneSignalLanguage]string{
			types.English: "Withdraw Notice",
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Your %s %s withdrawal has been sent out!",
				withdrawal.Amount.String(),
				withdrawal.CurrencyID),
		},
		UserID: userID.String(),
		ReturnData: map[string]string{
			"target": "/wallet",
		},
	}
	return req, nil
}

// GenOrderFilled generates order filled message and packs into
// return value.
func (a *AppStruct) GenOrderFilled(order models.Order) (*AppRequest, error) {
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(order.UserID.String(), cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Order of %s is completed",
				order.TradingPairID),
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Your %s order of %s has been completed",
				order.Type,
				order.TradingPairID),
		},
		UserID: order.UserID.String(),
	}

	return req, nil
}

// GenOrderPartiallyFilled generates order update message and packs into
// return value.
func (a *AppStruct) GenOrderPartiallyFilled(order models.Order) (*AppRequest, error) {
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(order.UserID.String(), cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Order of %s is partially filled",
				order.TradingPairID),
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Your %s order of %s has been partially filled",
				order.Type,
				order.TradingPairID),
		},
		UserID: order.UserID.String(),
	}

	return req, nil
}

// GenPriceAlert generates price alert message and packs into
// return value.
func (a *AppStruct) GenPriceAlert(userID uuid.UUID, tradingPairID string,
	price decimal.Decimal) (*AppRequest, error) {
	req := &AppRequest{
		// No price alert on Coblet for now.
		ID: a.Tag.GetNotificationTag(userID.String(), cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf("%s @ %s", tradingPairID, price),
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"Price for %s has reached %s",
				tradingPairID,
				price.String()),
		},
		UserID: userID.String(),
		ReturnData: map[string]string{
			"target": fmt.Sprintf("/trade/%s", tradingPairID),
		},
	}

	return req, nil
}

// GenReferralMissionCompleted generates referral mission completed message and
// packs into return value.
func (a *AppStruct) GenReferralMissionCompleted(user models.User) (
	*AppRequest, error) {
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(user.ID.String(), cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English:            "Congrats!",
			types.ChineseTraditional: "一起開心去領獎拉～",
			types.ChineseSimplified:  "好开心呀",
		},
		Message: map[types.OneSignalLanguage]string{
			types.English:            "You just got a referral bonus. Go to the candy machine to win rewards!",
			types.ChineseTraditional: "恭喜你完成推薦任務並獲得糖果拉霸機券，馬上去試試手氣！",
			types.ChineseSimplified:  "恭喜你完成推荐任务并获得糖果拉霸机券，马上去试试手气！",
		},
		UserID: user.ID.String(),
		ReturnData: map[string]string{
			"target": "/campaign/candy-machine",
		},
	}

	return req, nil
}

// ViolentPriceMovementInfo defines struct for violent price movement.
type ViolentPriceMovementInfo struct {
	Pair            string
	UserID          string
	IsRising        bool
	PriceNotifyType types.PriceNotifyType
	ChangeRate      string
	EventID         string
	Experiment      string
	Group           string
}

// GenViolentPriceChangesAlert generates messages to notify user price suddenly
// change.
func (a *AppStruct) GenViolentPriceChangesAlert(
	info ViolentPriceMovementInfo) (*AppRequest, error) {

	var sideStr string
	var sideZhTwStr string
	var sideEmoji string
	if info.IsRising {
		sideStr = "+"
		sideZhTwStr = "上漲"
		sideEmoji = "📈"
	} else {
		sideStr = "-"
		sideZhTwStr = "下跌"
		sideEmoji = "📉"
	}
	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(info.UserID, cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"🚨 %v price %v%v%% in an hour %v",
				strings.Replace(info.Pair, "-USDT", "", -1),
				sideStr,
				info.ChangeRate,
				sideEmoji,
			),
			types.ChineseTraditional: fmt.Sprintf(
				"🚨 %v價格於一小時內%v%v%% %v",
				strings.Replace(info.Pair, "-USDT", "", -1),
				sideZhTwStr,
				info.ChangeRate,
				sideEmoji,
			),
		},
		Message: map[types.OneSignalLanguage]string{
			types.English:            fmt.Sprint("Check out details>>"),
			types.ChineseTraditional: fmt.Sprint("查看詳情>>"),
		},
		UserID: info.UserID,
		ReturnData: map[string]interface{}{
			"target": fmt.Sprintf("/trade/%v", info.Pair),
			"forecast_info": map[string]string{
				"trading_pair_id": info.Pair,
				"type":            "px move",
				"experiment":      info.Experiment,
				"group":           info.Group,
			},
		},
	}

	return req, nil
}

// MACDCrossInfo defines struct for MACD cross.
type MACDCrossInfo struct {
	Pair            string
	UserID          string
	PriceNotifyType types.PriceNotifyType
	CrossType       types.CrossType
	TimePeriod      types.Timeframe
	EventID         string
	Experiment      string
	Group           string
}

// GenMACDCrossAlert generates messages to notify user MACD cross happened.
func (a *AppStruct) GenMACDCrossAlert(
	info MACDCrossInfo) (*AppRequest, error) {

	var crossTypeStr string
	var crossTypeZhTwStr string
	var timePeriodStr string
	var timePeriodZhTwStr string
	var sideEmoji string
	if info.CrossType == types.DeathCross {
		crossTypeStr = "Death Cross"
		crossTypeZhTwStr = "死亡交叉訊號"
		sideEmoji = "📉"
	} else if info.CrossType == types.GoldenCross {
		crossTypeStr = "Golden Cross"
		crossTypeZhTwStr = "黃金交叉訊號"
		sideEmoji = "📈"
	} else {
		return nil, errors.New("unknown cross type")
	}

	if info.TimePeriod == types.OneDay {
		timePeriodStr = "Daily"
		timePeriodZhTwStr = "日線"
	} else if info.TimePeriod == types.OneHour {
		timePeriodStr = "Hourly"
		timePeriodZhTwStr = "小時線"
	}

	req := &AppRequest{
		ID: a.Tag.GetNotificationTag(info.UserID, cobxtypes.APICobx),
		Subject: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf("🚨 MACD %v of %v %v",
				crossTypeStr,
				strings.Replace(info.Pair, "-USDT", "", -1),
				sideEmoji,
			),
			types.ChineseTraditional: fmt.Sprintf("🚨 %v出現MACD%v %v",
				strings.Replace(info.Pair, "-USDT", "", -1),
				crossTypeZhTwStr,
				sideEmoji,
			),
		},
		Message: map[types.OneSignalLanguage]string{
			types.English: fmt.Sprintf(
				"%v MACD of %v formed a %v, check out details >>",
				timePeriodStr,
				strings.Replace(info.Pair, "-USDT", "", -1),
				crossTypeStr),
			types.ChineseTraditional: fmt.Sprintf(
				"%v%v出現MACD%v, 查看詳情>>",
				strings.Replace(info.Pair, "-USDT", "", -1),
				timePeriodZhTwStr,
				crossTypeZhTwStr,
			),
		},
		UserID: info.UserID,
		ReturnData: map[string]interface{}{
			"target": fmt.Sprintf("/trade/%v", info.Pair),
			"forecast_info": map[string]string{
				"trading_pair_id": info.Pair,
				"type":            string(info.PriceNotifyType),
				"experiment":      info.Experiment,
				"group":           info.Group,
				"event_id":        info.EventID,
			},
		},
	}

	return req, nil
}

// BatchSend calls push notification service with req.
func (a *AppStruct) BatchSend(ctx context.Context, req *BatchAppRequest) (
	notificationID string, recipients int, err error) {
	switch utils.Environment() {
	case utils.Production:
	case utils.Staging:
		for k := range req.Subjects {
			req.Subjects[k] += " [staging]"
		}
	case utils.Development:
		for k := range req.Subjects {
			req.Subjects[k] += " [dev]"
		}
	case utils.LocalDevelopment:
		for k := range req.Subjects {
			req.Subjects[k] += " [localdev]"
		}
		return
	case utils.CI:
		return
	}

	switch a.Service {
	case OnesignalSymbol:
		return a.OneSignalConfig.BatchSend(ctx, req)
	default:
		return "", 0, errors.New("Invalid app service type")
	}
}
