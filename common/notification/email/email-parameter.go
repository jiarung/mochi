package email

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/cobinhood/mochi/constants"
	"github.com/cobinhood/mochi/database/fixtures"
	"github.com/cobinhood/mochi/types"
)

const timeFormat = "2006-01-02 15:04:05 UTC-07"

// Requests allocate all request defined in email.*Request to send.
var Requests = []Request{
	&APIKeyGeneratedRequest{},
	&AddWithdrawalWalletRequest{},
	&AuthFailedRequest{},
	&AuthSuccessRequest{},
	&CMTRewardNotifyRequest{},
	&ChangeEmailFailedRequest{},
	&ChangeEmailNotificationRequest{},
	&CoinOfferingNotifyRequest{},
	&ConfirmChangeEmailRequest{},
	&ConfirmDeleteAccountRequest{},
	&ConfirmDisable2FARequest{},
	&ConfirmEnable2FARequest{},
	&ConfirmRequestToDisableTwoFARequest{},
	&DeleteAccountFailedRequest{},
	&DeleteAccountNotifyRequest{},
	&DepositConfirmedRequest{},
	&DepositReminderRequest{},
	&DeviceVerifyRequest{},
	&DisableTwoFAPassRequest{},
	&DisableTwoFARejectRequest{},
	&DisableTwoFASuccessRequest{},
	&VerifyEmailRequest{},
	&EpayWithdrawVerifyRequest{},
	//&FiatDepositNotifyRequest{},
	&FiatPurchaseBankNotifyRequest{},
	&FiatPurchaseConfirmationNotifyRequest{},
	&FiatPurchaseKioskNotifyRequest{},
	&FiatWithdrawVerifyRequest{},
	&GenericRequest{},
	&KYCUpgradeFailedRequest{},
	&KYCVerifiedNotifyRequest{},
	&LiquidationWarningRequest{},
	&LoginReminderRequest{},
	&RegistrationConfirmedRequest{},
	&ResetPasswordRequest{},
	&RevokeKYCNotifyRequest{},
	&SlotMachineTokenExpiredRequest{},
	&SlotMachineTokenRequest{},
	&TradingReminderNoTokenRequest{},
	&TradingReminderRequest{},
	//&TransactionReceiptRequest{},
	&Updated2FANotifyRequest{},
	&WelcomeLetterRequest{},
	&WithdrawVerifyRequest{},
}

// genParameter generates parameter.
func (e *BaseEmailRequest) genParameter(
	id string, paramMap map[string]string) *Parameter {
	lang := getLocale(id, []string{e.Locale})
	subject := fixtures.EmailTemplates[id].TranslationMap[lang]["subject"]
	paramMap["*|subject|*"] = subject
	template, substitutions := genTemplate(id, lang, subject, paramMap)
	substitutions["year"] = fmt.Sprintf("%d", time.Now().Year())

	return &Parameter{
		To:            e.ToEmail,
		Name:          e.ToName,
		FromEmail:     e.FromEmail,
		FromName:      e.FromName,
		Subject:       subject,
		Template:      template,
		Substitutions: substitutions,
	}
}

// Parameter returns Parameter.
func (e *VerifyEmailRequest) Parameter() *Parameter {
	id := "EMAIL_VERIFICATION"

	paramMap := map[string]string{
		"*|name|*":        e.ToName,
		"*|ip|*":          e.IP,
		"*|active_link|*": e.ActiveLink,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *AuthSuccessRequest) Parameter() *Parameter {
	id := "AUTH_SUCCESS"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
		"*|time|*": e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *AuthFailedRequest) Parameter() *Parameter {
	id := "AUTH_FAILED"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
		"*|time|*": e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DeviceVerifyRequest) Parameter() *Parameter {
	id := "DEVICE_VERIFICATION"
	platformFullName := e.Platform
	if len(e.Optional) > 0 {
		platformFullName = fmt.Sprintf("%s (%s)", platformFullName, e.Optional)
	}
	paramMap := map[string]string{
		"*|name|*":     e.ToName,
		"*|ip|*":       e.IP,
		"*|platform|*": html.EscapeString(platformFullName),
		"*|link|*":     e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ResetPasswordRequest) Parameter() *Parameter {
	id := "RESET_PASSWORD"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
		"*|link|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *FiatWithdrawVerifyRequest) Parameter() *Parameter {
	id := "FIAT_WITHDRAW_VERIFICATION"
	paramMap := map[string]string{
		"*|name|*":     e.ToName,
		"*|ip|*":       e.IP,
		"*|token|*":    e.Link,
		"*|amount|*":   e.Amount,
		"*|fee|*":      e.Fee,
		"*|currency|*": e.Currency,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *EpayWithdrawVerifyRequest) Parameter() *Parameter {
	id := "FIAT_WITHDRAW_VERIFICATION"
	amountcurrency := fmt.Sprintf("%s %s", e.Amount, e.Currency)
	feeamountcurrency := fmt.Sprintf("%s %s", e.FeeAmount, e.FeeCurrency)
	paramMap := map[string]string{
		"*|name|*":           e.ToName,
		"*|ip|*":             e.IP,
		"*|account|*":        e.EpayEmail,
		"*|amount|*":         amountcurrency,
		"*|fee|*":            feeamountcurrency,
		"*|payment_method|*": "Epay",
		"*|link|*":           e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *WithdrawVerifyRequest) Parameter() *Parameter {
	id := "WITHDRAW_VERIFICATION"
	feeAmountCurrency := fmt.Sprintf("%s %s", e.FeeAmount, e.FeeCurrency)
	paramMap := map[string]string{
		"*|name|*":     e.ToName,
		"*|ip|*":       e.IP,
		"*|currency|*": e.Currency,
		"*|address|*":  e.Address,
		"*|amount|*":   e.Amount,
		"*|fee|*":      feeAmountCurrency,
		"*|memo|*":     e.Memo,
		"*|link|*":     e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *AddWithdrawalWalletRequest) Parameter() *Parameter {
	id := "ADD_WITHDRAWAL_WALLET_VERIFICATION"
	paramMap := map[string]string{
		"*|name|*":     e.ToName,
		"*|ip|*":       e.IP,
		"*|currency|*": e.Currency,
		"*|address|*":  e.Address,
		"*|memo|*":     e.Memo,
		"*|link|*":     e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ConfirmEnable2FARequest) Parameter() *Parameter {
	id := "CONFIRM_ENABLE_TWO_FA"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
		"*|link|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *RegistrationConfirmedRequest) Parameter() *Parameter {
	id := "REGISTRATION_CONFIRMED"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *Updated2FANotifyRequest) Parameter() *Parameter {
	id := "2FA_UPDATED_NOTIFY"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|ip|*":   e.IP,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *KYCVerifiedNotifyRequest) Parameter() *Parameter {
	id := fmt.Sprintf("ACCOUNT_UPGRADE_TO_LEVEL_%d", e.Level)
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|time|*": e.Time.UTC().Format(timeFormat),
	}
	lang := getLocale(id, []string{e.Locale})
	subject := fmt.Sprintf(
		fixtures.EmailTemplates[id].TranslationMap[lang]["subject"], e.Level)
	template, substitutions := genTemplate(id, lang, subject, paramMap)

	return &Parameter{
		To:            e.ToEmail,
		Name:          e.ToName,
		FromEmail:     e.FromEmail,
		FromName:      e.FromName,
		Subject:       subject,
		Template:      template,
		Substitutions: substitutions,
	}
}

// Parameter returns Parameter.
func (e *KYCUpgradeFailedRequest) Parameter() *Parameter {
	id := "ACCOUNT_UPGRADE_FAILED"
	paramMap := map[string]string{
		"*|name|*":   e.ToName,
		"*|level|*":  fmt.Sprintf("%d", e.Level),
		"*|reason|*": html.EscapeString(e.Reason),
		"*|time|*":   e.Time.UTC().Format(timeFormat),
	}

	lang := getLocale(id, []string{e.Locale})
	subject := fmt.Sprintf(
		fixtures.EmailTemplates[id].TranslationMap[lang]["subject"], e.Level)
	template, substitutions := genTemplate(id, lang, subject, paramMap)

	return &Parameter{
		To:            e.ToEmail,
		Name:          e.ToName,
		FromEmail:     e.FromEmail,
		FromName:      e.FromName,
		Subject:       subject,
		Template:      template,
		Substitutions: substitutions,
	}
}

// Parameter returns Parameter.
func (e *APIKeyGeneratedRequest) Parameter() *Parameter {
	id := "API_KEY_GENERATED"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|ip|*":    e.IP,
		"*|time|*":  e.Time.UTC().Format(timeFormat),
		"*|label|*": e.Label,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ConfirmDisable2FARequest) Parameter() *Parameter {
	id := "CONFIRM_DISABLE_TWO_FA"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|ip|*":    e.IP,
		"*|token|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ConfirmChangeEmailRequest) Parameter() *Parameter {
	id := "CONFIRM_CHANGE_EMAIL"
	paramMap := map[string]string{
		"*|name|*":      e.ToName,
		"*|ip|*":        e.IP,
		"*|old_email|*": e.OldEmail,
		"*|new_email|*": e.NewEmail,
		"*|token|*":     e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ChangeEmailNotificationRequest) Parameter() *Parameter {
	id := "VERIFY_CHANGE_EMAIL"
	paramMap := map[string]string{
		"*|name|*":      e.ToName,
		"*|new_email|*": e.NewEmail,
		"*|time|*":      e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ChangeEmailFailedRequest) Parameter() *Parameter {
	id := "REJECT_CHANGE_EMAIL"
	paramMap := map[string]string{
		"*|name|*":   e.ToName,
		"*|reason|*": html.EscapeString(e.Reason),
		"*|time|*":   e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ConfirmDeleteAccountRequest) Parameter() *Parameter {
	id := "CONFIRM_DELETE_ACCOUNT"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|ip|*":    e.IP,
		"*|token|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DeleteAccountNotifyRequest) Parameter() *Parameter {
	id := "VERIFY_DELETE_ACCOUNT"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|time|*": e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DeleteAccountFailedRequest) Parameter() *Parameter {
	id := "REJECT_DELETE_ACCOUNT"
	paramMap := map[string]string{
		"*|name|*":   e.ToName,
		"*|reason|*": html.EscapeString(e.Reason),
		"*|time|*":   e.Time.UTC().Format(timeFormat),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *SlotMachineTokenRequest) Parameter() *Parameter {
	id := "SLOT_MACHINE_TOKEN"
	rewardType := types.RewardType(e.RewardType)
	if !rewardType.IsValid() {
		return nil
	}
	reason := ""
	switch rewardType {
	case types.DailyTradingReward:
		reason = "You have successfully completed a transaction on COBINHOOD."
	case types.RefereeKYCLevelTwoReward:
		reason = "You have successfully became a member of the COBINHOOD family."
	case types.ReferrerKYCLevelTwoReward:
		reason = "You have successfully invited a friend to join COBINHOOD."
	}
	paramMap := map[string]string{
		"*|name|*":   e.ToName,
		"*|count|*":  fmt.Sprintf("%d", e.Count),
		"*|reason|*": html.EscapeString(reason),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *SlotMachineTokenExpiredRequest) Parameter() *Parameter {
	id := "SLOT_MACHINE_TOKEN_EXPIRED"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|count|*": fmt.Sprintf("%d", e.Count),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *LiquidationWarningRequest) Parameter() *Parameter {
	id := "LIQUIDATION_WARNING"
	pairs := strings.Join(e.TradingPairs, ", ")
	maintenance := fmt.Sprintf("%.2f", constants.PositionMarginWarnRatio/
		constants.PositionMarginRatio*100)
	paramMap := map[string]string{
		"*|name|*":        e.ToName,
		"*|pairs|*":       pairs,
		"*|maintenance|*": maintenance,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DepositConfirmedRequest) Parameter() *Parameter {
	id := "DEPOSIT_CONFIRMED"
	paramMap := map[string]string{
		"*|name|*":     e.ToName,
		"*|currency|*": e.CurrencyID,
		"*|amount|*":   e.Amount,
		"*|tx|*":       e.TxHash,
		"*|from|*":     e.FromAddress,
		"*|to|*":       e.ToAddress,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DisableTwoFAPassRequest) Parameter() *Parameter {
	id := "DISABLE_TWO_FA_REQUEST_PASS"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|hour|*": e.DelayTimeHour,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DisableTwoFARejectRequest) Parameter() *Parameter {
	id := "DISABLE_TWO_FA_REQUEST_REJECT"
	paramMap := map[string]string{
		"*|name|*":   e.ToName,
		"*|reason|*": html.EscapeString(e.Reason),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DisableTwoFASuccessRequest) Parameter() *Parameter {
	id := "DISABLE_TWO_FA_REQUEST_SUCCESS"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|link|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *ConfirmRequestToDisableTwoFARequest) Parameter() *Parameter {
	id := "CONFIRM_REQUEST_TO_DISABLE_TWO_FA"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|link|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *WelcomeLetterRequest) Parameter() *Parameter {
	id := "WELCOME"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *DepositReminderRequest) Parameter() *Parameter {
	id := "DEPOSIT_REMINDER"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *LoginReminderRequest) Parameter() *Parameter {
	id := "LOGIN_REMINDER"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *TradingReminderRequest) Parameter() *Parameter {
	id := "TRADING_REMINDER"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|tp1|*":   e.TradingPairs[0],
		"*|tp2|*":   e.TradingPairs[1],
		"*|tp3|*":   e.TradingPairs[2],
		"*|ttp1|*":  e.TokenTradingPairs[0],
		"*|ttp2|*":  e.TokenTradingPairs[1],
		"*|ttp3|*":  e.TokenTradingPairs[2],
		"*|token|*": e.Token,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *TradingReminderNoTokenRequest) Parameter() *Parameter {
	id := "TRADING_REMINDER_NO_TOKEN"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|tp1|*":  e.TradingPairs[0],
		"*|tp2|*":  e.TradingPairs[1],
		"*|tp3|*":  e.TradingPairs[2],
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *RevokeKYCNotifyRequest) Parameter() *Parameter {
	id := "KYC_REVOCATION"
	paramMap := map[string]string{
		"*|name|*":           e.ToName,
		"*|reason|*":         html.EscapeString(e.Reason),
		"*|original_level|*": fmt.Sprintf("%d", e.OriginLevel),
		"*|target_level|*":   fmt.Sprintf("%d", e.TargetLevel),
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *CMTRewardNotifyRequest) Parameter() *Parameter {
	id := "CMT_REWARD_NOTIFY"
	paramMap := map[string]string{
		"*|name|*": e.ToName,
		"*|link|*": e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *CoinOfferingNotifyRequest) Parameter() *Parameter {
	id := "COIN_OFFERING_NOTIFICATION"
	paramMap := map[string]string{
		"*|name|*":  e.ToName,
		"*|token|*": e.Token,
		"*|link|*":  e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *FiatPurchaseKioskNotifyRequest) Parameter() *Parameter {
	id := "FIAT_PURCHASE_KIOSK"
	paramMap := map[string]string{
		"*|number|*": e.Amount,
		"*|token|*":  e.CurrencyID,
		"*|way|*":    e.PaymentMethod,
		"*|code|*":   e.PaymentCode,
		"*|link|*":   e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *FiatPurchaseBankNotifyRequest) Parameter() *Parameter {
	id := "FIAT_PURCHASE_BANK"
	paramMap := map[string]string{
		"*|number|*":  e.Amount,
		"*|token|*":   e.CurrencyID,
		"*|way|*":     e.PaymentMethod,
		"*|account|*": e.BankAccount,
		"*|link|*":    e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *FiatPurchaseConfirmationNotifyRequest) Parameter() *Parameter {
	id := "FIAT_PURCHASE_CONFIRMATION"
	paramMap := map[string]string{
		"*|number|*": e.Amount,
		"*|token|*":  e.CurrencyID,
		"*|link|*":   e.Link,
	}
	return e.genParameter(id, paramMap)
}

// Parameter returns Parameter.
func (e *TransactionReceiptRequest) Parameter() *Parameter {
	// FIXME(xnum): old implement is for coblet.
	return nil
}

// Parameter returns Parameter.
func (e *FiatDepositNotifyRequest) Parameter() *Parameter {
	// FIXME(xnum): old implement is out of dated.
	return nil
}

// Parameter returns Parameter.
func (e *GenericRequest) Parameter() *Parameter {
	return &Parameter{
		To:            e.ToEmail,
		Name:          e.ToName,
		FromEmail:     e.FromEmail,
		FromName:      e.FromName,
		Subject:       e.Subject,
		Template:      e.Template,
		Substitutions: e.Substitutions,
	}
}
