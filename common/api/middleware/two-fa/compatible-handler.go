package twofactor

import (
	"fmt"

	"github.com/jiarung/gorm"
	"github.com/satori/go.uuid"

	apicontext "github.com/jiarung/mochi/common/api/context"
	models "github.com/jiarung/mochi/models/exchange"
	"github.com/jiarung/mochi/types"
)

// TwoFARequiredAPI API name which need 2FA
type TwoFARequiredAPI string

// TwoFACompatibleHandler defines a interface which is compatible to two factor
// authentication process. To have 2FA compatible endpoint, implement the interface
// and wrap the handler with the `TwoFAHandler` function to return a gin.HandlerFunc
type TwoFACompatibleHandler interface {
	// TwoFARequiredAPI return the name of API for handler registration.
	TwoFARequiredAPI() TwoFARequiredAPI

	// IsTwoFAOptional returns if the two factor authentication is optional.
	IsTwoFAOptional(appCtx *apicontext.AppContext) bool

	//  PreValidate validates the gin.Context in the front of the handler.
	// If `res` is not nil, `TwoFAHandler` will return `res` and end this
	// handler. Otherwise, just return and the `TwoFAHandler` will continue.
	PreValidate(appCtx *apicontext.AppContext) (res interface{}, errCode string, err error)

	// GetOptionalInfo is used to get optional info for cache in column `optional_info`.
	// Implement this method if the info is needed in `ConfirmHandler`.
	GetOptionalInfo(appCtx *apicontext.AppContext,
		tx *gorm.DB, twoFAType types.TwoFactorAuthMethod) (optStr *string, err error)

	// GetUserID returns the user ID of the user that is using the API.
	GetUserID(appCtx *apicontext.AppContext) (userID uuid.UUID, err error)

	// GetTwoFAInformationOfUser returns the two FA information of the API.
	GetTwoFAInformationOfUser(appCtx *apicontext.AppContext,
		userID uuid.UUID) (TwoFAInformation, error)

	// ExecWithout2FA executes in the `TwoFAHandler` if
	// - 2FA is bypassed
	// - `IsTwoFAOptional` returns true and 2FA method is None
	ExecWithout2FA(appCtx *apicontext.AppContext) (
		res interface{}, errCode string, err error)

	//  ConfirmHandler is registered and used in `/confirm_two_factor_authentication` endpoint to make
	// final confirm with `optionalInfo`
	ConfirmHandler(appCtx *apicontext.AppContext, optionalInfo *string,
		twoFAType types.TwoFactorAuthMethod, authID uuid.UUID) (res interface{},
		errCode string, err error)
}

// TwoFAInformation defines the interface that specifies the two factor authentication
// information for TwoFACompatibleHandler.
type TwoFAInformation interface {
	// TwoFactorAuthMethod returns the two factor authentication method.
	TwoFactorAuthMethod() (types.TwoFactorAuthMethod, error)

	// SMSCountryCoder returns the country code for sms 2FA type.
	SMSCountryCode() (string, error)

	// SMSPhoneNumber returns the phone number for sms 2FA type.
	SMSPhoneNumber() (string, error)

	// TOTPSecret returns the totp secret for totp 2FA type.
	TOTPSecret() (string, error)
}

// UserTwoFAInformation defines a struct that implements the TwoFAInformation
// interface and returns the 2FA information of the user account.
type UserTwoFAInformation struct {
	Method      types.TwoFactorAuthMethod
	CountryCode *string
	PhoneNumber *string
	TotpSecret  *string
}

// NewUserTwoFAInformation returns a new UserTwoFAInformation from appCtx and
// userID
func NewUserTwoFAInformation(appCtx *apicontext.AppContext, userID uuid.UUID) (*UserTwoFAInformation, error) {
	user := models.User{}
	result := appCtx.DB.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("Get user<%s> error: %v", userID.String(), result.Error)
	}
	info := UserTwoFAInformation{
		Method: user.TwoFactorAuthMethod,
	}

	switch user.TwoFactorAuthMethod {
	case types.TwoFactorAuthSMS:
		userSMSAuth := models.SMSAuth{}
		result := appCtx.DB.Model(&user).Related(&userSMSAuth)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to get sms auth of user. err(%s)",
				result.Error)
		}
		info.CountryCode = &userSMSAuth.CountryCode
		info.PhoneNumber = &userSMSAuth.PhoneNumber
	case types.TwoFactorAuthTOTP:
		userTOTP := models.TOTPAuth{}
		result := appCtx.DB.Model(&user).Related(&userTOTP)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to get totp auth of user. err(%s)",
				result.Error)
		}
		info.TotpSecret = &userTOTP.TOTPSecret
	case types.TwoFactorAuthNone:
	default:
		return nil, fmt.Errorf("unexpected two factor method. %s",
			user.TwoFactorAuthMethod)
	}

	return &info, nil
}

// TwoFactorAuthMethod implements TwoFAInformation interface.
func (info *UserTwoFAInformation) TwoFactorAuthMethod() (types.TwoFactorAuthMethod, error) {
	if !info.Method.IsValid() {
		return "", fmt.Errorf("invalid two factor method. %s", info.Method)
	}
	return info.Method, nil
}

// SMSCountryCode implements TwoFAInformation interface.
func (info *UserTwoFAInformation) SMSCountryCode() (string, error) {
	if info.Method != types.TwoFactorAuthSMS {
		return "", fmt.Errorf("invalid two factor method. %s", info.Method)
	}
	if info.CountryCode == nil {
		return "", fmt.Errorf("nil country code")
	}
	return *info.CountryCode, nil
}

// SMSPhoneNumber implements TwoFAInformation interface.
func (info *UserTwoFAInformation) SMSPhoneNumber() (string, error) {
	if info.Method != types.TwoFactorAuthSMS {
		return "", fmt.Errorf("invalid two factor method. %s", info.Method)
	}
	if info.PhoneNumber == nil {
		return "", fmt.Errorf("nil phone number")
	}
	return *info.PhoneNumber, nil
}

// TOTPSecret implements TwoFAInformation interface.
func (info *UserTwoFAInformation) TOTPSecret() (string, error) {
	if info.Method != types.TwoFactorAuthTOTP {
		return "", fmt.Errorf("invalid two factor method. %s", info.Method)
	}
	if info.TotpSecret == nil {
		return "", fmt.Errorf("nil totp secret")
	}
	return *info.TotpSecret, nil
}
