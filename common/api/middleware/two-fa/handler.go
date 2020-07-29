package twofactor

import (
	"fmt"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/satori/go.uuid"

	"github.com/cobinhood/mochi/api-cobx/message-code"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/common/config/misc"
	jwtFactory "github.com/cobinhood/mochi/common/jwt"
	"github.com/cobinhood/mochi/common/logging"
	"github.com/cobinhood/mochi/common/notification"
	"github.com/cobinhood/mochi/database"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/types"
)

var (
	twoFATokenExpiredTime = time.Duration(TwoFaTokenExpiredTimeoutSecond) * time.Second
	twoFASendSMSTimeout   = time.Duration(TwoFaSendSmsTimeoutSecond) * time.Second

	// ConfirmedTwoFAMap is the map for 2FA required endpoint.
	ConfirmedTwoFAMap = map[TwoFARequiredAPI]TwoFACompatibleHandler{}
)

const (
	// ResetPasswordAPI reset password API name,
	ResetPasswordAPI TwoFARequiredAPI = "reset_password"
	// ChangePasswordAPI change password API name.
	ChangePasswordAPI TwoFARequiredAPI = "change_password"
	// LoginAPI login API name.
	LoginAPI TwoFARequiredAPI = "login"
	// SubmitKYCPhoneNumber submit kyc phone number api name.
	SubmitKYCPhoneNumber TwoFARequiredAPI = "submit_kyc_phone_number"
	// AddWithdrawalAddressAPI add withdrawal address API name
	AddWithdrawalAddressAPI TwoFARequiredAPI = "add_withdrawal_address"
	// WithdrawFundsAPI withdraw funds API name
	WithdrawFundsAPI TwoFARequiredAPI = "withdrawal_funds"
	// FiatWithdrawalFundsAPI add fiat withdrawal API name.
	FiatWithdrawalFundsAPI TwoFARequiredAPI = "fiat_withdrawal_funds"
	// EpayWithdrawalAPI add epay withdrawal API name.
	EpayWithdrawalAPI TwoFARequiredAPI = "epay_withdrawal"
	// GenAPITokenAPI gen API token API name.
	GenAPITokenAPI TwoFARequiredAPI = "gen_api_token"
	// Disable2FAAPI disable 2FA API name.
	Disable2FAAPI TwoFARequiredAPI = "disable_two_factor_authentication"
	// DeleteAccountAPI delete account API name.
	DeleteAccountAPI TwoFARequiredAPI = "delete_account"
	// ChangeEmailAPI change email API name.
	ChangeEmailAPI TwoFARequiredAPI = "change_email"
	// DisableOAuth2ClientAPI disable OAuth2 client API name.
	DisableOAuth2ClientAPI TwoFARequiredAPI = "disable_oauth2_client"
)

var actionMap = map[TwoFARequiredAPI]struct {
	sms  types.SMSAuthAction
	totp types.TOTPAuthAction
}{
	ResetPasswordAPI:        {types.SMSResetPassword, types.TOTPResetPassword},
	ChangePasswordAPI:       {types.SMSChangePassword, types.TOTPChangePassword},
	LoginAPI:                {types.SMSLoginAuthorization, types.TOTPLoginAuthorization},
	SubmitKYCPhoneNumber:    {types.SMSSubmitKYCPhoneNumber, types.TOTPTwoFactorRequired},
	AddWithdrawalAddressAPI: {types.SMSAddWithdrawalAddress, types.TOTPAddWithdrawalAddress},
	WithdrawFundsAPI:        {types.SMSWithdrawFunds, types.TOTPWithdrawFunds},
	FiatWithdrawalFundsAPI:  {types.SMSFiatWithdrawFunds, types.TOTPFiatWithdrawFunds},
	EpayWithdrawalAPI:       {types.SMSEpayWithdrawFunds, types.TOTPEpayWithdrawFunds},
	GenAPITokenAPI:          {types.SMSGenAPIToken, types.TOTPGenAPIToken},
	Disable2FAAPI:           {types.DisableSMSTwoFactor, types.DisableTOTPTwoFactor},
	DeleteAccountAPI:        {types.SMSDeleteAccount, types.TOTPDeleteAccount},
	ChangeEmailAPI:          {types.SMSChangeEmail, types.TOTPChangeEmail},
	DisableOAuth2ClientAPI:  {types.SMSDisableOAuth2Client, types.TOTPDisableOAuth2Client},
}

// TwoFARequiredSuccess returns an HTTP success response object
func TwoFARequiredSuccess(response interface{}) interface{} {
	return struct {
		TwoFA interface{} `json:"2fa"`
	}{TwoFA: response}
}

type twoFARequiredRes struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type twoFARequiredMsg struct {
	MessageCode string `json:"message_code"`
}

// TwoFAHandler return 2FA required handler with implemented struct
func TwoFAHandler(
	handler TwoFACompatibleHandler, bypass2FA bool) middleware.AppHandlerFunc {
	if _, duplicate := ConfirmedTwoFAMap[handler.TwoFARequiredAPI()]; duplicate {
		panic(fmt.Sprintf("api has already registered: %v", handler.TwoFARequiredAPI()))
	}

	smsSender := notification.SMS()
	ConfirmedTwoFAMap[handler.TwoFARequiredAPI()] = handler
	return func(appCtx *apicontext.AppContext) {
		logger := appCtx.Logger()
		logger.SetLabel(logging.LabelApp, "api:middleware:twoFARequired():"+
			string(handler.TwoFARequiredAPI()))

		validateRes, errCode, validateErr := handler.PreValidate(appCtx)
		if len(errCode) > 0 || validateErr != nil {
			if len(errCode) > 0 {
				appCtx.SetError(errCode)
			}

			if validateErr != nil {
				logger.Error("validate error: %v", validateErr)
			}

			appCtx.Abort()
			return
		}

		if validateRes != nil {
			appCtx.SetJSON(TwoFARequiredSuccess(validateRes))
			return
		}

		if !bypass2FA {

			userID, getUserIDErr := handler.GetUserID(appCtx)
			if getUserIDErr != nil {
				logger.Error("get UserID error: %v", getUserIDErr)
				appCtx.Abort()
				return
			}

			twoFAInfo, err := handler.GetTwoFAInformationOfUser(appCtx, userID)
			if err != nil {
				logger.Error("get two factor auth info error : %v", err)
				appCtx.Abort()
				return
			}

			twoFAMethod, err := twoFAInfo.TwoFactorAuthMethod()
			if err != nil {
				logger.Error("get two factor auth method error : %v", err)
				appCtx.Abort()
				return
			}

			switch twoFAMethod {
			case types.TwoFactorAuthSMS:
				smsAuth := models.SMSAuth{}
				result := appCtx.DB.Raw(`
				SELECT id FROM sms_auth
				WHERE user_id = ? AND created_at >= ? LIMIT 1`,
					userID, time.Now().Add(-twoFASendSMSTimeout)).Scan(&smsAuth)
				if result.Error != nil && !result.RecordNotFound() {
					logger.Error("Error while get sms_auth by user<%v>: %v",
						userID, result.Error)
					appCtx.Abort()
					return
				}

				if !uuid.Equal(smsAuth.ID, uuid.Nil) {
					appCtx.SetError(
						apierrors.WaitForCooldownTime, twoFASendSMSTimeout.String())
					return
				}

				countryCode, err := twoFAInfo.SMSCountryCode()
				if err != nil {
					logger.Error("get sms country code error : %v", err)
					appCtx.Abort()
					return
				}

				phoneNumber, err := twoFAInfo.SMSPhoneNumber()
				if err != nil {
					logger.Error("get sms phone number error : %v", err)
					appCtx.Abort()
					return
				}

				code, err := notification.GenerateRandomCode()
				if err != nil {
					logger.Error("generate sms random code error: %v", err)
					appCtx.Abort()
					return
				}
				var toCountryCode string
				var toPhoneNum string
				var authID uuid.UUID
				txErr := database.Transaction(appCtx.DB, func(tx *gorm.DB) error {
					info, err := handler.GetOptionalInfo(appCtx, tx, types.TwoFactorAuthSMS)
					if err != nil {
						return err
					}
					smsAuth := &models.SMSAuth{
						AuthBase: models.AuthBase{
							Timestamp:        time.Now(),
							UserID:           &userID,
							IPAddress:        appCtx.RequestIP,
							AuthFailureCount: 0,
							ExpireAt:         time.Now().Add(twoFATokenExpiredTime),
							OptionalInfo:     info,
						},
						Action:      actionMap[handler.TwoFARequiredAPI()].sms,
						SMSCode:     code,
						PhoneNumber: phoneNumber,
						CountryCode: countryCode,
					}
					result = tx.Create(&smsAuth)
					if result.Error != nil {
						return result.Error
					}

					authID = smsAuth.ID
					toCountryCode = smsAuth.CountryCode
					toPhoneNum = smsAuth.PhoneNumber
					return nil
				})

				if txErr != nil {
					logger.Error("Error while sms transaction. Error: %v", txErr)
					appCtx.Abort()
					return
				}

				go func() {
					if err := smsSender.SendTo(
						toCountryCode,
						toPhoneNum,
						code+" is your COBINHOOD verification code."); err != nil {
						logger.Error("Error while sending SMS Auth. Error: %v", err)
						return
					}
				}()
				token, err := jwtFactory.Build(jwtFactory.TwoFARequiredObj{
					API:       string(handler.TwoFARequiredAPI()),
					TwoFAType: string(types.TwoFactorAuthSMS),
					AuthID:    authID,
				}).Gen(appCtx.ServiceName,
					misc.TwoFaTokenExpiredTimeoutInt(),
				)
				if err != nil {
					logger.Error("Error while generate sms JWT. Error: %v", err)
					appCtx.Abort()
					return
				}
				res := twoFARequiredRes{
					Type:  string(types.TwoFactorAuthSMS),
					Token: token,
				}
				appCtx.SetJSON(TwoFARequiredSuccess(res))
				return
			case types.TwoFactorAuthTOTP:
				totpAuth := models.TOTPAuth{}
				result := appCtx.DB.Raw(`
					SELECT id FROM totp_auth
					WHERE user_id = ? AND created_at >= ? LIMIT 1`,
					userID, time.Now().Add(-twoFASendSMSTimeout)).Scan(&totpAuth)
				if result.Error != nil && !result.RecordNotFound() {
					logger.Error("Error while get totp_auth by user<%v>: %v",
						userID, result.Error)
					appCtx.Abort()
					return
				}

				if !uuid.Equal(totpAuth.ID, uuid.Nil) {
					appCtx.SetError(
						apierrors.WaitForCooldownTime, twoFASendSMSTimeout.String())
					return
				}

				totpSecret, err := twoFAInfo.TOTPSecret()
				if err != nil {
					logger.Error("get sms phone number error : %v", err)
					appCtx.Abort()
					return
				}

				var authID uuid.UUID
				txErr := database.Transaction(appCtx.DB, func(tx *gorm.DB) error {
					info, err := handler.GetOptionalInfo(appCtx, tx, types.TwoFactorAuthTOTP)
					if err != nil {
						return err
					}

					totpAuth := &models.TOTPAuth{
						AuthBase: models.AuthBase{
							Timestamp:        time.Now(),
							UserID:           &userID,
							IPAddress:        appCtx.RequestIP,
							AuthFailureCount: 0,
							ExpireAt:         time.Now().Add(twoFATokenExpiredTime),
							OptionalInfo:     info,
						},
						Action:     actionMap[handler.TwoFARequiredAPI()].totp,
						TOTPSecret: totpSecret,
					}
					result := tx.Create(&totpAuth)
					if result.Error != nil {
						return result.Error
					}

					authID = totpAuth.ID
					return nil
				})

				if txErr != nil {
					logger.Error("Error while totp transaction. Error: %v", txErr)
					appCtx.Abort()
					return
				}

				token, err := jwtFactory.Build(jwtFactory.TwoFARequiredObj{
					API:       string(handler.TwoFARequiredAPI()),
					TwoFAType: string(types.TwoFactorAuthTOTP),
					AuthID:    authID,
				}).Gen(appCtx.ServiceName,
					misc.TwoFaTokenExpiredTimeoutInt(),
				)
				if err != nil {
					logger.Error("Error while generate totp JWT. Error: %v", err)
					appCtx.Abort()
					return
				}
				res := twoFARequiredRes{
					Type:  string(types.TwoFactorAuthTOTP),
					Token: token,
				}
				appCtx.SetJSON(TwoFARequiredSuccess(res))
				return
			case types.TwoFactorAuthNone:
				if !handler.IsTwoFAOptional(appCtx) {
					appCtx.SetJSON(TwoFARequiredSuccess(
						twoFARequiredMsg{MessageCode: messagecode.TwoFARequired}))
					return
				}
			default:
				logger.Error("unexpected type: %v", twoFAMethod)
				return
			}
		}

		res, errCode, err := handler.ExecWithout2FA(appCtx)
		if len(errCode) > 0 || err != nil {
			if len(errCode) > 0 {
				appCtx.SetError(errCode)
			}

			if err != nil {
				logger.Error("ExecWithout2FA() error: %v", err)
			}

			appCtx.Abort()
			return
		}
		appCtx.SetJSON(res)
	}
}

// GetTwoFASendSMSTimeout returns twoFASendSMSTimeout.
func GetTwoFASendSMSTimeout() time.Duration {
	return twoFASendSMSTimeout
}

// SetTwoFASendSMSTimeout sets twoFASendSMSTimeout.
func SetTwoFASendSMSTimeout(timeout time.Duration) {
	twoFASendSMSTimeout = timeout
}
