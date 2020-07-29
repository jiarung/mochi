package jwtfactory

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/common/config/secret"
	"github.com/jiarung/mochi/types"
)

type jwtType string

const (
	registrationJWT                         jwtType = "registration"
	emailVerificationJWT                    jwtType = "email_verification"
	addWithdrawalWalletEmailVerificationJWT jwtType = "add_withdrawal_wallet_email_verification"
	withdrawalFundsEmailVerificationJWT     jwtType = "withdrawal_funds_email_verification"
	fiatWithdrawalFundsEmailVerificationJWT jwtType = "fiat_withdrawal_funds_email_verification"
	epayWithdrawalFundsEmailVerificationJWT jwtType = "epay_withdrawal_funds_email_verification"
	accessTokenJWT                          jwtType = "access_token"
	deviceAuthorizationJWT                  jwtType = "device_authorization"
	deviceVerificationJWT                   jwtType = "device_verification"
	deleteAccountJWT                        jwtType = "delete_account"
	changeEmailJWT                          jwtType = "change_email"
	resetPasswordJWT                        jwtType = "reset_password"
	twoFAEnableJWT                          jwtType = "two_fa_enable"
	twoFADisableConfirmJWT                  jwtType = "two_fa_disable_confirm"
	twoFAEnableConfirmJWT                   jwtType = "two_fa_enable_confirm"
	twoFARequiredJWT                        jwtType = "two_fa_required"
	apiTokenJWT                             jwtType = "api_token"
	oauth2AuthorizationCodeJWT              jwtType = "oauth2_authorization_code"
	oauth2AccessTokenJWT                    jwtType = "oauth2_access_token"
	oauth2RefreshTokenJWT                   jwtType = "oauth2_refresh_token"
	requestToDisableTwoFAJWT                jwtType = "request_to_disable_two_fa"
)

// getJWTSecret returns JWT secret based on type and service name.
func getJWTSecret(t jwtType, s cobxtypes.ServiceName) []byte {
	serviceNameSuffixMap := map[cobxtypes.ServiceName]string{
		cobxtypes.WS:        "WS",
		cobxtypes.APIAdmin:  "ADMIN",
		cobxtypes.Test:      "TEST",
		cobxtypes.APICoblet: "COBLET",
		cobxtypes.APICobx:   "COBEX",
	}
	return []byte(secret.Get(strings.Join([]string{
		strings.ToUpper(string(t)),
		"SECRET",
		serviceNameSuffixMap[s],
	}, "_")))
}

// APIKeySecret defines api token key secret store.
type APIKeySecret struct {
	m         map[string][]byte
	latestVer string
}

// NewAPIKeySecret creates an api key secret to gen/validate api token.
func NewAPIKeySecret() *APIKeySecret {
	s := &APIKeySecret{
		m: map[string][]byte{},
	}
	// FIXME(xnum): explict init.
	// Load everything into apiKeySecretMap.
	s.latestVer = secret.Get("API_TOKEN_SECRET_LATEST_VERSION")
	count := 1
	for {
		version := fmt.Sprintf("V%d", count)
		apiSecret := secret.Get("API_TOKEN_SECRET_" + version)
		if apiSecret == "" {
			panic("API_TOKEN_SECRET_" + version + " is empty")
		}

		s.m[version] = []byte(apiSecret)
		if version == s.latestVer {
			break
		}

		count++
	}
	return s
}

func (a *APIKeySecret) latestSecret() []byte {
	return a.m[a.latestVer]
}

func (a *APIKeySecret) getSecret(version string) []byte {
	return a.m[version]
}

type jwtFactoryObj interface {
	getType() jwtType
	getClaims(expireSec int) jwt.Claims
}

// WithdrawFundsEmailVerificationObj instance for builder
type WithdrawFundsEmailVerificationObj struct {
	WithdrawID        uuid.UUID
	EmailAuthID       uuid.UUID
	ToAddress         string
	IsCobinhoodWallet bool
}

func (o WithdrawFundsEmailVerificationObj) getType() jwtType {
	return withdrawalFundsEmailVerificationJWT
}

func (o WithdrawFundsEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"withdraw_id":         o.WithdrawID,
		"email_auth_id":       o.EmailAuthID,
		"to_address":          o.ToAddress,
		"is_jiarung_wallet": o.IsCobinhoodWallet,
		"exp":                 time.Now().Add(time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// AddWithdrawalWalletEmailVerificationObj instance for builder
type AddWithdrawalWalletEmailVerificationObj struct {
	WalletID uuid.UUID
}

func (o AddWithdrawalWalletEmailVerificationObj) getType() jwtType {
	return addWithdrawalWalletEmailVerificationJWT
}

func (o AddWithdrawalWalletEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"wallet_id": o.WalletID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// FiatWithdrawFundsEmailVerificationObj instance for builder
type FiatWithdrawFundsEmailVerificationObj struct {
	FiatWithdrawID uuid.UUID
	MotionID       uuid.UUID
}

func (o FiatWithdrawFundsEmailVerificationObj) getType() jwtType {
	return fiatWithdrawalFundsEmailVerificationJWT
}

func (o FiatWithdrawFundsEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"motion_id":          o.MotionID,
		"fiat_withdrawal_id": o.FiatWithdrawID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second),
	}
	return
}

// EpayWithdrawFundsEmailVerificationObj instance for builder
type EpayWithdrawFundsEmailVerificationObj struct {
	EmailAuthID         uuid.UUID
	GenericWithdrawalID uuid.UUID
}

func (o EpayWithdrawFundsEmailVerificationObj) getType() jwtType {
	return epayWithdrawalFundsEmailVerificationJWT
}

func (o EpayWithdrawFundsEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"email_auth_id":         o.EmailAuthID,
		"generic_withdrawal_id": o.GenericWithdrawalID,
		"exp":                   time.Now().Add(time.Duration(expireSec) * time.Second),
	}
	return
}

// RegistrationObj instance for builder
type RegistrationObj struct {
	RegistrationID uuid.UUID
}

func (o RegistrationObj) getType() jwtType {
	return registrationJWT
}

func (o RegistrationObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"registration_id": o.RegistrationID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// EmailVerificationObj instance for builder
type EmailVerificationObj struct {
	RegistrationID uuid.UUID
}

func (o EmailVerificationObj) getType() jwtType {
	return emailVerificationJWT
}

func (o EmailVerificationObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"registration_id": o.RegistrationID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// DeviceAuthorizationObj instance for builder
type DeviceAuthorizationObj struct {
	LoginID               uuid.UUID
	DeviceAuthorizationID uuid.UUID
}

func (o DeviceAuthorizationObj) getType() jwtType {
	return deviceAuthorizationJWT
}

func (o DeviceAuthorizationObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"login_id":                o.LoginID,
		"device_authorization_id": o.DeviceAuthorizationID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// DeviceVerificationObj instance for builder
type DeviceVerificationObj struct {
	LoginID               uuid.UUID
	DeviceAuthorizationID uuid.UUID
}

func (o DeviceVerificationObj) getType() jwtType {
	return deviceVerificationJWT
}

func (o DeviceVerificationObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"login_id":                o.LoginID,
		"device_authorization_id": o.DeviceAuthorizationID,
		"exp":                     time.Now().Add(time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// AccessTokenObj instance for builder
type AccessTokenObj struct {
	UserID                uuid.UUID
	AccessTokenID         uuid.UUID
	DeviceAuthorizationID uuid.UUID
	Platform              types.DevicePlatform
	LoginCount            int64
}

func (o AccessTokenObj) getType() jwtType {
	return accessTokenJWT
}

func (o AccessTokenObj) getClaims(expireSec int) (claims jwt.Claims) {
	c := jwt.MapClaims{
		"user_id":                 o.UserID,
		"access_token_id":         o.AccessTokenID,
		"device_authorization_id": o.DeviceAuthorizationID,
		"platform":                o.Platform,
		"login_count":             o.LoginCount,
	}
	if o.Platform == types.DeviceWeb {
		c["exp"] = time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix()
	}

	return c
}

// TwoFAEnableObj instance for builder
type TwoFAEnableObj struct {
	TotpURL   string
	AuthID    string
	TokenType string
}

func (o TwoFAEnableObj) getType() jwtType {
	return twoFAEnableJWT
}

func (o TwoFAEnableObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"totp_url": o.TotpURL,
		"auth_id":  o.AuthID,
		"type":     o.TokenType,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// TwoFAEnableConfirmObj instance for builder
type TwoFAEnableConfirmObj struct {
	AuthID    string
	TokenType string
}

func (o TwoFAEnableConfirmObj) getType() jwtType {
	return twoFAEnableConfirmJWT
}

func (o TwoFAEnableConfirmObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"auth_id": o.AuthID,
		"type":    o.TokenType,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// TwoFADisableConfirmObj instance for build
type TwoFADisableConfirmObj struct {
	UserID string
	AuthID string
}

func (o TwoFADisableConfirmObj) getType() jwtType {
	return twoFADisableConfirmJWT
}

func (o TwoFADisableConfirmObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"user_id": o.UserID,
		"auth_id": o.AuthID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// TwoFARequiredObj instance for builder
type TwoFARequiredObj struct {
	API       string
	TwoFAType string
	AuthID    uuid.UUID
}

func (o TwoFARequiredObj) getType() jwtType {
	return twoFARequiredJWT
}

func (o TwoFARequiredObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"api":     o.API,
		"type":    o.TwoFAType,
		"auth_id": o.AuthID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// ResetPasswordObj instance for builder
type ResetPasswordObj struct {
	ResetPasswordID string
	UserID          string
}

func (o ResetPasswordObj) getType() jwtType {
	return resetPasswordJWT
}

func (o ResetPasswordObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"reset_password_id": o.ResetPasswordID,
		"user_id":           o.UserID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

//ChangeEmailEmailVerificationObj instance for builder
type ChangeEmailEmailVerificationObj struct {
	ChangeEmailID string
	UserID        string
	FromNewEmail  bool
}

func (o ChangeEmailEmailVerificationObj) getType() jwtType {
	return changeEmailJWT
}

func (o ChangeEmailEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"change_email_id": o.ChangeEmailID,
		"user_id":         o.UserID,
		"from_new_email":  o.FromNewEmail,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// DeleteAccountEmailVerificationObj instance for builder
type DeleteAccountEmailVerificationObj struct {
	DeleteAccountID string
	UserID          string
}

func (o DeleteAccountEmailVerificationObj) getType() jwtType {
	return fiatWithdrawalFundsEmailVerificationJWT
}

func (o DeleteAccountEmailVerificationObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"delete_account_id": o.DeleteAccountID,
		"user_id":           o.UserID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// APITokenObj instance for builder
type APITokenObj struct {
	UserID     uuid.UUID
	APITokenID uuid.UUID
	Scope      []types.Scope
}

func (o APITokenObj) getType() jwtType {
	return apiTokenJWT
}

func (o APITokenObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"api_token_id": o.APITokenID,
		"user_id":      o.UserID,
		"scope":        o.Scope,
	}
	return
}

// OAuth2AuthorizationCodeObj instance for builder
type OAuth2AuthorizationCodeObj struct {
	AuthorizationCodeID uuid.UUID
}

func (o OAuth2AuthorizationCodeObj) getType() jwtType {
	return oauth2AuthorizationCodeJWT
}

func (o OAuth2AuthorizationCodeObj) getClaims(
	expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"oauth2_authorization_code_id": o.AuthorizationCodeID,
		"exp":                          time.Now().Add(time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// OAuth2AccessTokenObj instance for builder
type OAuth2AccessTokenObj struct {
	AccessTokenID uuid.UUID
	ClientID      uuid.UUID
	UserID        uuid.UUID
	Scope         []types.Scope
}

func (o OAuth2AccessTokenObj) getType() jwtType {
	return oauth2AccessTokenJWT
}

func (o OAuth2AccessTokenObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"oauth2_access_token_id": o.AccessTokenID,
		"client_id":              o.ClientID,
		"user_id":                o.UserID,
		"scope":                  o.Scope,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return claims
}

// OAuth2RefreshTokenObj instance for builder
type OAuth2RefreshTokenObj struct {
	RefreshTokenID uuid.UUID
	ClientID       uuid.UUID
	UserID         uuid.UUID
	Scope          []types.Scope
}

func (o OAuth2RefreshTokenObj) getType() jwtType {
	return oauth2RefreshTokenJWT
}

func (o OAuth2RefreshTokenObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"oauth2_refresh_token_id": o.RefreshTokenID,
		"client_id":               o.ClientID,
		"user_id":                 o.UserID,
		"scope":                   o.Scope,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// RequestToDisableTwoFAObj instance for builder
type RequestToDisableTwoFAObj struct {
	UserID                                uuid.UUID
	SelfieWithPhotoIdentificationCacheKey string
	EmailAuthID                           uuid.UUID
}

func (o RequestToDisableTwoFAObj) getType() jwtType {
	return requestToDisableTwoFAJWT
}

func (o RequestToDisableTwoFAObj) getClaims(expireSec int) (claims jwt.Claims) {
	claims = jwt.MapClaims{
		"user_id": o.UserID,
		"selfie_with_photo_identification_cache_key": o.SelfieWithPhotoIdentificationCacheKey,
		"email_auth_id": o.EmailAuthID,
		"exp": time.Now().Add(
			time.Duration(expireSec) * time.Second).Unix(),
	}
	return
}

// Method the instance for builder
type Method struct {
	obj    jwtFactoryObj
	Secret string
}

// Gen generate JWT token after build
func (m *Method) Gen(serviceName cobxtypes.ServiceName,
	expireSec int) (token string, err error) {
	// use secret from BuildWithSecret or config
	var secret []byte
	if m.Secret != "" {
		secret = []byte(m.Secret)
	} else {
		secret = getJWTSecret(m.obj.getType(), serviceName)
	}

	if len(secret) <= 0 {
		err = fmt.Errorf("empty secret")
		return
	}

	claims := m.obj.getClaims(expireSec)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = jwtToken.SignedString(secret)
	return
}

// GenWithCOBSecret generate JWT token with COB secret after build
// [HEADER].[PAYLOAD].[SIGNATURE].[V1:COB-SIGNATURE]
func (m *Method) GenWithCOBSecret(serviceName cobxtypes.ServiceName,
	store *APIKeySecret, expireSec int) (token string,
	err error) {
	jwtToken, err := m.Gen(serviceName, expireSec)
	if err != nil {
		return
	}

	mac := hmac.New(sha256.New, store.latestSecret())
	_, err = mac.Write([]byte(jwtToken))
	if err != nil {
		return
	}
	signature := hex.EncodeToString(mac.Sum(nil))
	token = jwtToken + "." + store.latestVer + ":" + signature
	return
}

// Validate validate JWT token after build
func (m *Method) Validate(token string, serviceName cobxtypes.ServiceName) (
	claims jwt.MapClaims, expired bool, err error) {
	// use secret from BuildWithSecret or config
	var secret []byte
	if m.Secret != "" {
		secret = []byte(m.Secret)
	} else {
		secret = getJWTSecret(m.obj.getType(), serviceName)
	}

	if len(secret) <= 0 {
		err = fmt.Errorf("empty secret")
		return
	}

	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ",
				token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		validationErr := err.(*jwt.ValidationError)
		expired = validationErr.Errors == jwt.ValidationErrorExpired
		return
	}

	var ok bool
	if claims, ok = t.Claims.(jwt.MapClaims); ok && t.Valid {
		return
	}

	return nil, expired, errors.New("invalid token")
}

// Build build method from implemented instance
func Build(obj jwtFactoryObj) (method *Method) {
	method = &Method{
		obj: obj,
	}
	return
}

// BuildWithSecret build method with secret from implemented instance
func BuildWithSecret(obj jwtFactoryObj,
	secret string) (method *Method) {
	method = &Method{
		obj:    obj,
		Secret: secret,
	}
	return
}

// ParseJWTPayload parse JWT payload
func ParseJWTPayload(token string) (claimMap map[string]interface{}, err error) {
	parts := strings.Split(token, ".")

	payload := parts[1]
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decodeBytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return
	}

	err = json.Unmarshal(decodeBytes, &claimMap)
	if err != nil {
		return
	}

	return
}

// ValidateCOBSecret validate JWT token with COB secret.
// [HEADER].[PAYLOAD].[SIGNATURE].[V1:COB-SIGNATURE]
func ValidateCOBSecret(token string, store *APIKeySecret) (err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 4 {
		err = fmt.Errorf("invalid cob secret part length: %d", len(parts))
		return
	}

	jwtToken := parts[0] + "." + parts[1] + "." + parts[2]

	lastParts := strings.Split(parts[3], ":")
	if len(lastParts) != 2 {
		err = fmt.Errorf("invalid cob secret last part length: %d", len(parts))
		return
	}
	secretVersion := lastParts[0]
	signatureForCheck := lastParts[1]

	mac := hmac.New(sha256.New, store.getSecret(secretVersion))
	_, err = mac.Write([]byte(jwtToken))
	if err != nil {
		return
	}
	signatureForVerify := hex.EncodeToString(mac.Sum(nil))

	if signatureForCheck != signatureForVerify {
		err = fmt.Errorf("signature is not correct")
		return
	}

	return nil
}
