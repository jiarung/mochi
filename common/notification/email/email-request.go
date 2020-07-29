/*
Package email defines email request struct and wraps email deliver service.

Requests in this package are published to licensing customers. So it shouldn't
import any non-official package. Some request struct may have protected field
to let json marshaller working correctly, it should be modified by member
function e.g. SetDomain().
*/
package email

import (
	"fmt"
	"time"
)

// Parameter defines parameter for email service.
type Parameter struct {
	To            string
	Name          string
	FromEmail     string
	FromName      string
	Subject       string
	Template      string            // template of email in service
	Substitutions map[string]string // string substitutions in mail content
}

// LocaleMap transfer from frontend format to standard
var LocaleMap = map[string]string{
	"ar":      "en",
	"br":      "pt_BR",
	"de":      "de",
	"en":      "en",
	"es":      "es",
	"fr":      "fr",
	"he":      "en",
	"it":      "it",
	"ja":      "en",
	"ko":      "ko",
	"nl":      "nl",
	"ru":      "ru",
	"pt":      "pt_PT",
	"tr":      "tr",
	"vi":      "vi",
	"zh-Hans": "zh_Hans_CN",
	"zh-Hant": "zh_Hant_TW",
}

// Request defines interface of Requests. All XXXRequest should implement
// Parameter. EmailService should process emailParameter and send it.
type Request interface {
	Parameter() *Parameter
}

// BaseEmailRequest defines basic fields.
type BaseEmailRequest struct {
	FromEmail string
	FromName  string
	ToName    string
	ToEmail   string
	Locale    string
}

// VerifyEmailRequest for email verified.
type VerifyEmailRequest struct {
	BaseEmailRequest
	IP    string
	Token string

	// Protected fields.
	ActiveLink string
}

// SetDomain sets domain and generates URL link.
func (e *VerifyEmailRequest) SetDomain(domain string) {
	e.ActiveLink = fmt.Sprintf("https://%s/verify/confirmMail?token=%s",
		domain, e.Token)
}

// AuthSuccessRequest for authorization success.
type AuthSuccessRequest struct {
	BaseEmailRequest
	IP   string
	Time time.Time
}

// AuthFailedRequest for authorization failed.
type AuthFailedRequest struct {
	BaseEmailRequest
	IP   string
	Time time.Time
}

// DeviceVerifyRequest for device verification.
type DeviceVerifyRequest struct {
	BaseEmailRequest
	IP       string
	Platform string
	Optional string // append to platform if length is not zero.
	Token    string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *DeviceVerifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmDevice?token=%s",
		domain,
		e.Token,
	)
}

// ResetPasswordRequest for resetting password.
type ResetPasswordRequest struct {
	BaseEmailRequest
	IP    string
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ResetPasswordRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/resetPassword?token=%s", domain, e.Token)
}

// FiatWithdrawVerifyRequest for fiat withdraw verification.
type FiatWithdrawVerifyRequest struct {
	BaseEmailRequest
	IP       string
	Token    string
	Account  string
	Amount   string
	Fee      string
	Currency string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *FiatWithdrawVerifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmFiatWithdrawal?token=%s",
		domain, e.Token)
}

// EpayWithdrawVerifyRequest is withdraw verification for epay.
type EpayWithdrawVerifyRequest struct {
	BaseEmailRequest
	IP          string
	Token       string
	EpayEmail   string
	Amount      string
	Currency    string
	FeeAmount   string
	FeeCurrency string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *EpayWithdrawVerifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmEpayFiatWithdrawal?token=%s",
		domain, e.Token)
}

// WithdrawVerifyRequest defines withdraw verification request.
type WithdrawVerifyRequest struct {
	BaseEmailRequest
	IP          string
	Token       string
	Address     string
	Currency    string
	Amount      string
	FeeCurrency string
	FeeAmount   string
	Memo        string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *WithdrawVerifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmWithdrawal?token=%s",
		domain, e.Token)
}

// AddWithdrawalWalletRequest is add withdrawal wallet verification request.
type AddWithdrawalWalletRequest struct {
	BaseEmailRequest
	IP       string
	Token    string
	Currency string
	Address  string
	Memo     string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *AddWithdrawalWalletRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmAddWithdrawalWallet?token=%s",
		domain, e.Token)
}

// ConfirmEnable2FARequest is verification of confirm enable 2fa.
type ConfirmEnable2FARequest struct {
	BaseEmailRequest
	IP    string
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ConfirmEnable2FARequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmEnable2FA?token=%s",
		domain, e.Token)
}

// RegistrationConfirmedRequest is confirm registration.
type RegistrationConfirmedRequest struct {
	BaseEmailRequest
	IP string
}

// Updated2FANotifyRequest is notification of 2FA updated.
type Updated2FANotifyRequest struct {
	BaseEmailRequest
	IP string
}

// KYCVerifiedNotifyRequest is notification of KYC verified.
type KYCVerifiedNotifyRequest struct {
	BaseEmailRequest
	Level int
	Time  time.Time
}

// KYCUpgradeFailedRequest is notification of KYC upgrade failed.
type KYCUpgradeFailedRequest struct {
	BaseEmailRequest
	Level  int
	Reason string
	Time   time.Time
}

// APIKeyGeneratedRequest is notification of API Key generated.
type APIKeyGeneratedRequest struct {
	BaseEmailRequest
	IP    string
	Time  time.Time
	Label string
}

// ConfirmDisable2FARequest is verification of confirm disable 2FA.
type ConfirmDisable2FARequest struct {
	BaseEmailRequest
	IP    string
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ConfirmDisable2FARequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmDisable2FA?token=%s",
		domain, e.Token)
}

// ConfirmChangeEmailRequest is verification of confirm change email.
type ConfirmChangeEmailRequest struct {
	BaseEmailRequest
	OldEmail string
	NewEmail string
	IP       string
	Token    string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ConfirmChangeEmailRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmChangeEmail?token=%s",
		domain, e.Token)
}

// ChangeEmailNotificationRequest is notification of email changed.
type ChangeEmailNotificationRequest struct {
	BaseEmailRequest
	NewEmail string
	Time     time.Time
}

// ChangeEmailFailedRequest is notification of email changing failed.
type ChangeEmailFailedRequest struct {
	BaseEmailRequest
	Reason string
	Time   time.Time
}

// ConfirmDeleteAccountRequest is verification of deleting account.
type ConfirmDeleteAccountRequest struct {
	BaseEmailRequest
	IP    string
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ConfirmDeleteAccountRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmDeleteAccount?token=%s",
		domain, e.Token)
}

// DeleteAccountNotifyRequest defines account deleted notification request.
type DeleteAccountNotifyRequest struct {
	BaseEmailRequest
	Time time.Time
}

// DeleteAccountFailedRequest defines delete account failed notification
// request.
type DeleteAccountFailedRequest struct {
	BaseEmailRequest
	Reason string
	Time   time.Time
}

// SlotMachineTokenRequest defines slot machine token request.
type SlotMachineTokenRequest struct {
	BaseEmailRequest
	RewardType string
	Count      int
	Expires    time.Time
}

// SlotMachineTokenExpiredRequest defines slot mahcine token expired request.
type SlotMachineTokenExpiredRequest struct {
	BaseEmailRequest
	Count int
}

// LiquidationWarningRequest defines liquidation warning request.
type LiquidationWarningRequest struct {
	BaseEmailRequest
	TradingPairs []string
}

// DepositConfirmedRequest defines deposit confirmed request.
type DepositConfirmedRequest struct {
	BaseEmailRequest
	CurrencyID  string
	Amount      string
	TxHash      string
	FromAddress string
	ToAddress   string
}

// DisableTwoFAPassRequest defines disable two fa passwd request.
type DisableTwoFAPassRequest struct {
	BaseEmailRequest
	DelayTimeHour string
}

// DisableTwoFARejectRequest defines disable two fa rejected request.
type DisableTwoFARejectRequest struct {
	BaseEmailRequest
	Reason string
}

// DisableTwoFASuccessRequest defines disable two fa successesd request.
type DisableTwoFASuccessRequest struct {
	BaseEmailRequest
	Link string
}

// ConfirmRequestToDisableTwoFARequest defines request to disable two fa
// verification request.
type ConfirmRequestToDisableTwoFARequest struct {
	BaseEmailRequest
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *ConfirmRequestToDisableTwoFARequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf(
		"https://%s/verify/confirmRequestToDisableTwoFA?token=%s",
		domain, e.Token)
}

// WelcomeLetterRequest defines welcome request.
type WelcomeLetterRequest struct {
	BaseEmailRequest
}

// DepositReminderRequest defines deposit reminder request.
type DepositReminderRequest struct {
	BaseEmailRequest
}

// LoginReminderRequest defines login reminder request.
type LoginReminderRequest struct {
	BaseEmailRequest
}

// TradingReminderRequest defines trading reminder request.
type TradingReminderRequest struct {
	BaseEmailRequest
	Token             string
	TradingPairs      []string
	TokenTradingPairs []string
}

// TradingReminderNoTokenRequest defines trading reminder without token request.
type TradingReminderNoTokenRequest struct {
	BaseEmailRequest
	TradingPairs []string
}

// RevokeKYCNotifyRequest defines revoke kyc notification request.
type RevokeKYCNotifyRequest struct {
	BaseEmailRequest
	Reason      string
	OriginLevel int
	TargetLevel int
}

// CMTRewardNotifyRequest defines cmt reward dispatched notification request.
type CMTRewardNotifyRequest struct {
	BaseEmailRequest

	// Protected fields.
	Link string
}

// SetDomain sets domain and generates URL link.
func (e *CMTRewardNotifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/wallet", domain)
}

// CoinOfferingNotifyRequest defines coin offering notification request.
type CoinOfferingNotifyRequest struct {
	BaseEmailRequest
	Token string

	// Protected fields.
	Link string
}

// SetDomain sets domain.
func (e *CoinOfferingNotifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/token-offering/history", domain)
}

// FiatPurchaseKioskNotifyRequest defines fiat purchase kiosk notification
// request.
type FiatPurchaseKioskNotifyRequest struct {
	BaseEmailRequest
	Amount        string
	CurrencyID    string
	PaymentMethod string
	PaymentCode   string
	Link          string
}

// SetDomain sets domain.
func (e *FiatPurchaseKioskNotifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/purchase-cryptocurrency/history", domain)
}

// FiatPurchaseBankNotifyRequest defines fiat purchase bank notification
// request.
type FiatPurchaseBankNotifyRequest struct {
	BaseEmailRequest
	Amount        string
	CurrencyID    string
	PaymentMethod string
	BankAccount   string
	Link          string
}

// SetDomain sets domain.
func (e *FiatPurchaseBankNotifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/purchase-cryptocurrency/history", domain)
}

// FiatPurchaseConfirmationNotifyRequest defines fiat purchase confirmmation
// notification request.
type FiatPurchaseConfirmationNotifyRequest struct {
	BaseEmailRequest
	Amount     string
	CurrencyID string
	Link       string
}

// SetDomain sets domain.
func (e *FiatPurchaseConfirmationNotifyRequest) SetDomain(domain string) {
	e.Link = fmt.Sprintf("https://%s/wallet/history", domain)
}

// TransactionReceiptRequest defines transaction recipt request.
type TransactionReceiptRequest struct {
	BaseEmailRequest
	OrderID         string
	Currency        string
	Side            string
	Size            string
	Price           string
	Total           string
	PaymentCurrency string
	CompletedAt     time.Time
}

// FiatDepositNotifyRequest defines fiat deposit notification request.
type FiatDepositNotifyRequest struct {
	BaseEmailRequest
	IP       string
	Amount   string
	Fee      string
	Currency string
}

// GenericRequest defines generic request for admin api.
type GenericRequest struct {
	BaseEmailRequest
	Subject       string
	Template      string
	Substitutions map[string]string
}
