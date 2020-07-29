// FIXME(hao): use middleware_test to avoid import cycle. should be removed
// after issue resolved.
package middleware_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/apps/exchange/common/api/middlewares"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/cache/keys"
	"github.com/cobinhood/mochi/common/aes"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/common/jwt"
	"github.com/cobinhood/mochi/common/scope-auth"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/gcp/kms"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	"github.com/cobinhood/mochi/infra/app"
	"github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/models/exchange/exchangetest"
	"github.com/cobinhood/mochi/types"
)

type MiddlewareTestSuite struct {
	suite.Suite
	router                                   *gin.Engine
	db                                       *gorm.DB
	exchangeUserAccessToken                  string
	apiTokenWithUserAccess                   string
	apiTokenWithoutUserAccess                string
	oauth2AccessTokenWithAccountAccess       string
	oauth2AccessTokenWithAccountAndKYCAccess string
	ctx                                      context.Context
	cancel                                   func()
}

// SetupSuite initializes the database,
// setups the testing router and create users with kyc data
func (s *MiddlewareTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	scopeauth.Initialize(cobxtypes.Test)
	s.db = database.GetDB(database.Default)

	// Setup HTTP Router.
	// Read request body once
	s.router = gin.Default()
	s.router.Use(logger.NewLoggerMiddleware)

	// Log Each Request / Response.
	s.router.Use(middlewares.ContextLogger())

	// Insert AppContext
	s.router.Use(middleware.AppContextMiddleware(cobxtypes.Test))

	// Register Error Handler Middleware.
	s.router.Use(middleware.ResponseHandler)

	// Error Recovery & Reporting, Must Be Registered Last.
	s.router.Use(middlewares.PanicLogger)

	handler := func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		s.Require().Nil(err)

		s.T().Log("PASSED!", ctx.Request.RequestURI)
		if appCtx.IsAPIToken() {
			ctx.Status(999)
		} else if appCtx.IsOAuth2Token() {
			ctx.Status(777)
		} else {
			ctx.Status(200)
		}
	}

	// Insert ScopAuthMiddleware
	s.router.Use(middleware.ScopeAuth("test"))

	s.router.GET("alive", handler)
	s.router.GET("ready", handler)
	v1 := s.router.Group("v1")
	{
		v1.POST("helo", handler)
		v1.GET("users", handler)
		v1.GET("users/:user_id", handler)
		v1.PUT("users/:user_id", handler)
		v1.POST("oauth2/token", handler)
		v1.GET("oauth2/resources/userinfo", handler)
		v1.GET("oauth2/resources/kyc_tiers", handler)
	}

	// prepare user
	accessToken := exchangetest.CreateUserWithAccessToken(
		s.db, []types.Role{types.RoleKYCAuditor}, "192.0.2.1")
	s.exchangeUserAccessToken = accessToken

	s.ctx, s.cancel = context.WithCancel(context.Background())

	// gen user and api token with user access token
	var err error
	s.apiTokenWithUserAccess, err = createUserAndGenAPITokenWithoutCaching(
		s.ctx, s.db, types.ScopeSlice{
			types.ScopeExchangeTradeRead,
			types.ScopeExchangeTradeWrite,
			types.ScopeExchangeAccountRead})
	s.Require().Nil(err)
	s.Require().NotZero(len(s.apiTokenWithUserAccess))

	// gen user and api token without user access token
	s.apiTokenWithoutUserAccess, err = createUserAndGenAPITokenWithoutCaching(
		s.ctx, s.db, types.ScopeSlice{
			types.ScopeExchangeTradeRead,
			types.ScopeExchangeTradeWrite})
	s.Require().Nil(err)
	s.Require().NotZero(len(s.apiTokenWithoutUserAccess))

	// gen user and OAuth2 access token with account read access
	s.oauth2AccessTokenWithAccountAccess, err =
		createUserAndGenOAuth2AccessToken(s.db, types.ScopeSlice{
			types.ScopeThirdPartyAccountRead})
	s.Require().Nil(err)
	s.Require().NotZero(len(s.oauth2AccessTokenWithAccountAccess))

	// gen user and OAuth2 access token with account and KYC read access
	s.oauth2AccessTokenWithAccountAndKYCAccess, err =
		createUserAndGenOAuth2AccessToken(s.db, types.ScopeSlice{
			types.ScopeThirdPartyAccountRead,
			types.ScopeThirdPartyKYCRead})
	s.Require().Nil(err)
	s.Require().NotZero(len(s.oauth2AccessTokenWithAccountAndKYCAccess))
}

func (s *MiddlewareTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
	scopeauth.Finalize()
}

func (s *MiddlewareTestSuite) TestNonLoggedInUser() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	request = httptest.NewRequest(http.MethodGet, "/alive", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)
}

func (s *MiddlewareTestSuite) TestLoggedInUser() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	request = httptest.NewRequest(http.MethodGet, "/alive", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", s.exchangeUserAccessToken)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)
}

func (s *MiddlewareTestSuite) TestAPITokenWithUserAccess() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	request = httptest.NewRequest(http.MethodGet, "/alive", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)
}

func (s *MiddlewareTestSuite) TestAPITokenWithoutUserAccess() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", s.apiTokenWithoutUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", s.apiTokenWithoutUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", s.apiTokenWithoutUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(999, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)
}

func (s *MiddlewareTestSuite) TestRevokedAPIToken() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	err := revokeAPIToken(s.apiTokenWithUserAccess)
	s.Require().Nil(err)

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(200, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", s.apiTokenWithUserAccess)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)
}

func (s *MiddlewareTestSuite) TestOAuth2AccessTokenWithAccountAccess() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	authorization := "Bearer " + s.oauth2AccessTokenWithAccountAccess

	request = httptest.NewRequest(http.MethodGet, "/alive", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)
}

func (s *MiddlewareTestSuite) TestOAuth2AccessTokenWithAccountAndKYCAccess() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	authorization := "Bearer " + s.oauth2AccessTokenWithAccountAndKYCAccess

	request = httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/users/anything", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPut, "/v1/users/arbitrary_thing", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusForbidden, recorder.Code)

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(777, recorder.Code)
}

func (s *MiddlewareTestSuite) TestRevokedOAuth2AccessToken() {
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	err := revokeOAuth2AccessToken(s.oauth2AccessTokenWithAccountAndKYCAccess)
	s.Require().Nil(err)

	authorization := "Bearer " + s.oauth2AccessTokenWithAccountAndKYCAccess

	request = httptest.NewRequest(http.MethodPost, "/v1/oauth2/token", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(200, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/userinfo", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)

	request = httptest.NewRequest(http.MethodGet, "/v1/oauth2/resources/kyc_tiers", nil)
	request.Header.Set("Authorization", authorization)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusUnauthorized, recorder.Code)
}

func createUserAndGenAPITokenWithoutCaching(ctx context.Context, db *gorm.DB,
	scopes types.ScopeSlice) (token string, err error) {
	user := exchangetest.CreateTestingUser(db, 1)[0]
	client, err := kms.NewDefaultClient(ctx, kms.KeyAPIToken)
	if err != nil {
		return
	}

	var secret string
	apiSecret := exchange.APISecret{}
	result := db.Where("user_id = ?", user.ID).First(&apiSecret)
	if result.Error != nil && !result.RecordNotFound() {
		err = result.Error
		return
	} else if result.RecordNotFound() {
		var aesBytes aes.Key
		aesBytes, err = aes.GenerateAESKey(32)
		if err != nil {
			return
		}

		secret = base64.StdEncoding.EncodeToString(aesBytes)
		apiSecret.UserID = user.ID
		apiSecret.Secret, err = client.Encrypt([]byte(secret))
		if err != nil {
			return
		}

		result = db.Create(&apiSecret)
		if result.Error != nil {
			err = result.Error
			return
		}
	}

	secretBytes, err := client.Decrypt(apiSecret.Secret)
	if err != nil {
		return
	}

	secret = string(secretBytes)

	apiToken := exchange.APIToken{
		Label:  "test",
		Scopes: scopes,
		UserID: user.ID,
		TwoFactorAuthConfirmationBase: exchange.TwoFactorAuthConfirmationBase{
			TwoFactorAuthMethod: types.TwoFactorAuthNone,
		},
	}
	err = db.Create(&apiToken).Error
	if err != nil {
		return
	}

	store := jwtfactory.NewAPIKeySecret()
	token, err = jwtfactory.BuildWithSecret(jwtfactory.APITokenObj{
		UserID:     apiToken.UserID,
		APITokenID: apiToken.ID,
		Scope:      apiToken.Scopes,
	}, secret).GenWithCOBSecret(cobxtypes.Test, store, 1999)
	if err != nil {
		return
	}

	hiddenSignature := strings.Split(token, ".")[3][:10] + "*****"
	err = db.Model(exchange.APIToken{}).Where("id = ?", apiToken.ID).
		Update("signature", hiddenSignature).Error
	if err != nil {
		return
	}

	return
}

func revokeAPIToken(token string) (err error) {
	claims, err := jwtfactory.ParseJWTPayload(token)
	if err != nil {
		return
	}

	apiTokenID, err := uuid.FromString(claims["api_token_id"].(string))
	if err != nil {
		return
	}

	userID := claims["user_id"].(string)

	err = database.GetDB(database.Default).Model(exchange.APIToken{}).Where("id = ?", apiTokenID).
		Update("revoked_at", time.Now()).Error
	if err != nil {
		return
	}

	err = cache.GetRedis().RemoveFieldFromMap(
		keys.GetAPITokenKeyByUserStr(userID), apiTokenID.String())
	if err != nil {
		return
	}

	return
}

func createUserAndGenOAuth2AccessToken(db *gorm.DB, scopes types.ScopeSlice) (
	token string, err error) {

	users := exchangetest.CreateTestingUser(db, 2)
	appUser := users[0]
	appOwner := users[1]

	oauth2Client := exchange.OAuth2Client{
		ID:            uuid.NewV4(),
		SecretType:    types.OAuth2SecretTypeNone,
		OwnerID:       appOwner.ID,
		Public:        true,
		Enabled:       true,
		AllowedScopes: scopes,
	}
	err = db.Create(&oauth2Client).Error
	if err != nil {
		return
	}

	timeout := time.Hour
	issuedAt := time.Now()
	expireAt := issuedAt.Add(timeout)

	oauth2AccessToken := exchange.OAuth2Token{
		UserID:   appUser.ID,
		ClientID: oauth2Client.ID,
		Scopes:   scopes,
		Type:     types.OAuth2AccessToken,
		IssuedAt: issuedAt,
		ExpireAt: expireAt,
	}
	err = db.Create(&oauth2AccessToken).Error
	if err != nil {
		return
	}

	token, err = jwtfactory.Build(jwtfactory.OAuth2AccessTokenObj{
		AccessTokenID: oauth2AccessToken.ID,
		ClientID:      oauth2AccessToken.ClientID,
		UserID:        appUser.ID,
		Scope:         scopes,
	}).Gen(cobxtypes.Test, int(timeout.Seconds()))
	return
}

func revokeOAuth2AccessToken(token string) (err error) {
	claims, _, err := jwtfactory.Build(
		jwtfactory.OAuth2AccessTokenObj{}).Validate(token, cobxtypes.Test)
	if err != nil {
		return
	}

	oauth2AccessTokenIDStr := claims["oauth2_access_token_id"].(string)
	oauth2AccessTokenID, err := uuid.FromString(oauth2AccessTokenIDStr)
	if err != nil {
		return
	}

	err = database.GetDB(database.Default).Model(exchange.OAuth2Token{}).Where("id = ?",
		oauth2AccessTokenID).Update("revoked_at", time.Now()).Error
	if err != nil {
		return
	}

	err = cache.GetRedis().Delete("oauth2_access_token:" + oauth2AccessTokenIDStr)
	if err != nil {
		return
	}

	return
}

func TestMiddleware(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
