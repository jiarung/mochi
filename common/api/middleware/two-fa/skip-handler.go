package twofactor

import "github.com/cobinhood/cobinhood-backend/types"

var _ TwoFAInformation = (*SkipTwoFAInformation)(nil)

// SkipTwoFAInformation provides skip 2fa info impl.
type SkipTwoFAInformation struct{}

// TwoFactorAuthMethod return none.
func (info *SkipTwoFAInformation) TwoFactorAuthMethod() (
	types.TwoFactorAuthMethod, error) {
	return types.TwoFactorAuthNone, nil
}

// SMSCountryCode returns empty string.
func (info *SkipTwoFAInformation) SMSCountryCode() (string, error) {
	return "", nil
}

// SMSPhoneNumber returns empty string.
func (info *SkipTwoFAInformation) SMSPhoneNumber() (string, error) {
	return "", nil
}

// TOTPSecret returns empty string.
func (info *SkipTwoFAInformation) TOTPSecret() (string, error) {
	return "", nil
}
