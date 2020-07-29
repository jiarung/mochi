package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/cache/helper"
	"github.com/cobinhood/mochi/cache/keys"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	jwtFactory "github.com/cobinhood/mochi/common/jwt"
	"github.com/cobinhood/mochi/common/logging"
	"github.com/cobinhood/mochi/common/scope-auth"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/gcp/kms"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/types"
)

const tokensWriteTimeout = 1 << 3

// OAuth2AccessTokenType is the type of token issued by our OAuth2 server.
var OAuth2AccessTokenType = "Bearer"

// ScopeAuth return a middleware that validates the scopes of each endpoints
// from `scopeMap` and JWT from users if needed. This middleware should be placed
// - AFTER `AppContextMiddleware` and `ErrorHandler`
// - BEFORE all the handler functions
//
// - Check if JWT exists
// - Get scopes of the endpoint from `scopeMap`
// - if no JWT exists
//		- Proceed only if requiring public scope
// - else
//		- Proceed if JWT is valid and one of the followings:
//			- `validScopes` contains `ScopePublic`
//			- `userScopes` contains at least one valid scope
func ScopeAuth(service cobxtypes.ServiceName, opt ...interface{}) gin.HandlerFunc {
	store := jwtFactory.NewAPIKeySecret()
	return func(ctx *gin.Context) {
		// Always allow OPTIONS to pass through.
		if ctx.Request.Method == http.MethodOptions {
			return
		}

		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			logger := logger.Get(ctx)
			logger.Error("Fail to obtain AppContext Error: %v", err)
			ctx.Abort()
			return
		}
		logger := appCtx.Logger()
		logger.SetLabel(logging.LabelApp, "scope-auth:middleware")
		defer logger.DeleteLabel(logging.LabelApp)

		var jwtStr string
		var jwtExists bool
		if len(strings.Split(ctx.GetHeader("Authorization"), ".")) == 4 {
			// API token
			vErr := validateAPIToken(appCtx, store,
				ctx.GetHeader("Authorization"))
			if vErr != nil {
				logger.Error("api token is invalid: %v", vErr)
			} else {
				jwtExists = true
			}
		} else if len(strings.Split(ctx.GetHeader("Authorization"), " ")) == 2 {
			// OAuth2 token
			authParts := strings.SplitN(ctx.GetHeader("Authorization"), " ", 2)
			authType, authCredentials := authParts[0], authParts[1]
			if authType == OAuth2AccessTokenType && len(authCredentials) > 0 {
				vErr := validateOAuth2Token(appCtx, authCredentials)
				if vErr != nil {
					logger.Error("OAuth2 access token is invalid: %v", vErr)
				} else {
					jwtExists = true
				}
			}
		} else {
			// Access token: Get authorization from header or coockie.
			jwtStr, jwtExists = extractJWT(ctx, logger)
		}

		// Get scopes of endpoint.
		validScopes, err := scopeauth.GetScopes(service,
			strings.ToUpper(ctx.Request.Method), ctx.Request.URL.Path)
		if err != nil {
			logger.Info("can't find scopes of [%s] [%s] %s. err(%s)",
				service, ctx.Request.Method, ctx.Request.URL.Path, err)
			appCtx.SetError(apierrors.ResourceNotFound)
			return
		}
		if len(validScopes) == 0 {
			logger.Error("empty scopes of [%s] %s.%s.",
				ctx.Request.Method, service, ctx.Request.URL.Path)
			ctx.Abort()
			return
		}
		appCtx.RequiredScopes = validScopes
		logger.Debug("required scopes: %s", validScopes)

		validScopeMap := make(map[types.Scope]struct{})
		hasPublicScope := false
		for _, s := range validScopes {
			if s == types.ScopePublic {
				hasPublicScope = true
			}
			validScopeMap[s] = struct{}{}
		}

		// No JWT.
		if !jwtExists {
			if hasPublicScope {
				ctx.Next()
			} else {
				logger.Error("authentication failed with no JWT found")
				appCtx.SetError(apierrors.AuthenticationError)
			}
			return
		}

		if !appCtx.IsAPIToken() && !appCtx.IsOAuth2Token() {
			skipIPCheck := false
			if len(opt) > 0 {
				skipIPCheck = opt[0].(bool)
			}
			// With JWT, validation
			userID, deviceAuthorizationID, accessTokenID, userScopes,
				devicePlatform, err := authenticateJWT(jwtStr, appCtx.RequestIP,
				skipIPCheck, appCtx.ServiceName)
			if err != nil {
				if hasPublicScope {
					ctx.Next()
				} else {
					logger.Error("authentication failed with invalid JWT. err: %+v", err)
					appCtx.SetError(apierrors.AuthenticationError)
				}
				return
			}
			logger.SetLabel(logging.LabelUserDeviceID, deviceAuthorizationID.String())

			appCtx.Platform = devicePlatform
			appCtx.UserID = userID
			appCtx.AccessTokenID = accessTokenID
			appCtx.DeviceAuthorizationID = deviceAuthorizationID
			appCtx.UserAuthorizationScopes = userScopes
		}

		logger.SetLabel(logging.LabelUserID, appCtx.UserID.String())
		logger.Debug("user scopes: %s", appCtx.UserAuthorizationScopes)

		if hasPublicScope {
			ctx.Next()
			return
		}

		// Validate scopes.
		for _, s := range appCtx.UserAuthorizationScopes {
			if _, ok := validScopeMap[s]; ok {
				// user scope is in valid scope map, proceed
				ctx.Next()
				return
			}
		}

		// User scope is not in valid, abort
		logger.Error("unauthorized scopes. user scopes [%s]. required scopes[%s]",
			appCtx.UserAuthorizationScopes, validScopes)
		appCtx.SetError(apierrors.UnauthorizedScope)
	}
}

func validateAPIToken(appCtx *apicontext.AppContext,
	store *jwtFactory.APIKeySecret, token string) (err error) {
	err = jwtFactory.ValidateCOBSecret(token, store)
	if err != nil {
		return
	}

	claimMap, err := jwtFactory.ParseJWTPayload(token)
	if err != nil {
		return
	}

	// check scope
	scopesFromClaim, sExist := claimMap["scope"].([]interface{})
	if !sExist {
		err = fmt.Errorf("scope not exist")
		return
	}
	userScopes := make([]types.Scope, 0)
	for _, scope := range scopesFromClaim {
		str, ok := scope.(string)
		if ok {
			s := types.Scope(str)
			userScopes = append(userScopes, s)
		}
	}

	// check user id
	userIDStr := claimMap["user_id"].(string)
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return
	}

	// check api token
	apiTokenIDStr, exist := claimMap["api_token_id"].(string)
	if !exist {
		// it should be panic if api_token_id is not exist in payload
		panic(fmt.Errorf("api_token_id not exist"))
	}
	apiTokenID, err := uuid.FromString(apiTokenIDStr)
	if err != nil {
		return
	}

	// create shared key
	apiTokenKey := keys.GetAPITokenKeyByUserStr(userIDStr)

	// check secret
	secret, err := appCtx.Cache.GetFieldOfMap(apiTokenKey, keys.APITokenSecretKey)
	if err != nil {
		errorCode := cache.ParseCacheErrorCode(err)
		if errorCode != cache.ErrNilKey {
			return
		}

		// lazy loading
		var client *kms.Client
		client, err = kms.NewDefaultClient(appCtx, kms.KeyAPIToken)
		if err != nil {
			return
		}
		// set secret
		apiSecret := models.APISecret{}
		err = appCtx.DB.Where("user_id = ?", userID).First(&apiSecret).Error
		if err != nil {
			return
		}

		var secretBytes []byte
		secretBytes, err = client.Decrypt(apiSecret.Secret)
		if err != nil {
			return
		}
		secret = string(secretBytes)

		err = appCtx.Cache.SetFieldOfMap(apiTokenKey, "secret", secret)
		if err != nil {
			return
		}
	}

	// valid token with secret
	partialToken := strings.Join(strings.Split(token, ".")[:3], ".")
	_, _, err = jwtFactory.BuildWithSecret(jwtFactory.APITokenObj{}, secret).
		Validate(partialToken, appCtx.ServiceName)
	if err != nil {
		return
	}

	// check cached data and lazy loading with db query
	_, err = appCtx.Cache.GetFieldOfMap(apiTokenKey, apiTokenIDStr)
	if err != nil {
		errorCode := cache.ParseCacheErrorCode(err)
		if errorCode != cache.ErrNilKey {
			return
		}

		lockKey := apiTokenKey + "_lock"
		// lock while updating tokens
		appCtx.Cache.Lock(lockKey, appCtx.RequestTag(), tokensWriteTimeout)
		defer appCtx.Cache.UnLock(lockKey, appCtx.RequestTag())

		// set api tokens
		apiTokens := []models.APIToken{}
		err = appCtx.DB.Where("user_id = ? AND revoked_at IS NULL",
			userID).Find(&apiTokens).Error
		if err != nil {
			return
		}

		var isTokenExisting bool
		for _, at := range apiTokens {
			id := at.ID.String()
			err = appCtx.Cache.SetFieldOfMap(apiTokenKey, id, "")
			if err != nil {
				return
			}

			if !isTokenExisting {
				isTokenExisting = id == apiTokenIDStr
			}
		}

		if !isTokenExisting {
			err = fmt.Errorf("user<%s> api token<%s> error: %v",
				userIDStr, apiTokenIDStr, err)
			return
		}
	}

	appCtx.UserAuthorizationScopes = userScopes
	appCtx.UserID = &userID
	appCtx.APITokenID = &apiTokenID
	appCtx.Logger().SetLabel(logging.LabelAuthMethod, "api_token")
	return nil
}

func validateOAuth2Token(appCtx *apicontext.AppContext, token string) error {
	claims, _, err := jwtFactory.Build(
		jwtFactory.OAuth2AccessTokenObj{}).Validate(token, appCtx.ServiceName)
	if err != nil {
		return err
	}

	oauth2AccessTokenIDStr, found := claims["oauth2_access_token_id"].(string)
	if !found {
		return fmt.Errorf("oauth2_access_token_id could not be found")
	}
	oauth2AccessTokenID, err := uuid.FromString(oauth2AccessTokenIDStr)
	if err != nil {
		return err
	}

	clientIDStr, found := claims["client_id"].(string)
	if !found {
		return fmt.Errorf("client_id could not be found")
	}
	clientID, err := uuid.FromString(clientIDStr)
	if err != nil {
		return err
	}

	userIDStr, found := claims["user_id"].(string)
	if !found {
		return fmt.Errorf("user_id could not be found")
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return err
	}

	scopes, found := claims["scope"].([]interface{})
	if !found {
		return fmt.Errorf("scope could not be found")
	}

	oauth2TokenKey := "oauth2_access_token:" + oauth2AccessTokenIDStr
	found, err = appCtx.Cache.Exist(oauth2TokenKey)
	if err != nil {
		return err
	}

	if !found {
		oauth2Token := models.OAuth2Token{}
		result := appCtx.DB.Where("id = ? AND type = ? AND revoked_at IS NULL",
			oauth2AccessTokenID, types.OAuth2AccessToken).First(&oauth2Token)
		if result.Error != nil {
			return result.Error
		}

		expireSec := int(oauth2Token.ExpireAt.Sub(time.Now()) / time.Second)
		appCtx.Cache.Set(oauth2TokenKey, 1, expireSec)
	}

	userScopes := make([]types.Scope, 0)
	for _, scope := range scopes {
		scopeStr, ok := scope.(string)
		if ok {
			userScopes = append(userScopes, types.Scope(scopeStr))
		}
	}

	appCtx.UserAuthorizationScopes = userScopes
	appCtx.UserID = &userID
	appCtx.OAuth2TokenID = &oauth2AccessTokenID
	appCtx.OAuth2ClientID = clientID
	appCtx.Logger().SetLabel(logging.LabelAuthMethod, "oauth2_access_token")
	return nil
}

func extractJWT(ctx *gin.Context, logger logging.Logger) (jwtString string, exists bool) {
	var jwtStr string
	jwtStr = ctx.GetHeader("Authorization")
	if len(jwtStr) != 0 {
		jwtString = jwtStr
		exists = true
		return
	}

	var err error
	jwtStr, err = ctx.Cookie("Authorization")
	if err == nil && len(jwtStr) != 0 {
		jwtString = jwtStr
		exists = true
		return
	}

	jwtString = ""
	exists = false
	return
}

func authenticateJWT(jwtStr string, requestIP string, skipIPCheck bool,
	serviceName cobxtypes.ServiceName) (userID *uuid.UUID,
	deviceAuthorizationID *uuid.UUID, accessTokenID *uuid.UUID,
	userScopes []types.Scope, devicePlatform types.DevicePlatform, err error) {
	claims, isExpired, delErr := jwtFactory.Build(jwtFactory.AccessTokenObj{}).
		Validate(jwtStr, serviceName)
	if delErr != nil || isExpired {
		userID = nil
		deviceAuthorizationID = nil
		accessTokenID = nil
		userScopes = nil
		if isExpired {
			err = fmt.Errorf("token <%v> expired", accessTokenID)
		} else {
			err = delErr
		}
		return
	}

	userIDFromClaim, uExist := claims["user_id"].(string)
	accessTokenIDFromClaim, aExist := claims["access_token_id"].(string)
	devAuthIDFromClaim, dExist := claims["device_authorization_id"].(string)
	platformFromClaim, pExist := claims["platform"].(string)
	if !uExist || !aExist || !dExist || !pExist {
		userID = nil
		deviceAuthorizationID = nil
		accessTokenID = nil
		userScopes = nil
		err = fmt.Errorf(
			"invalid claims [user_id: %v, access_token_id: %v,"+
				" device_authorization_id: %v, platform: %v]",
			uExist, aExist, dExist, pExist,
		)
		return
	}

	accessTokenPayload := &helper.AccessTokenPayload{}
	if redisErr := accessTokenPayload.Get(
		database.GetDB(database.Default), accessTokenIDFromClaim); redisErr != nil {
		userID = nil
		deviceAuthorizationID = nil
		accessTokenID = nil
		userScopes = nil
		err = fmt.Errorf(
			"can't validate token on cache with key: <token:%v>. Err: %+v",
			accessTokenIDFromClaim, redisErr,
		)
		return
	}

	if !skipIPCheck && platformFromClaim == "Web" &&
		requestIP != accessTokenPayload.IP {
		delErr = cache.GetRedis().Delete(
			keys.GetAccessTokenCacheKey(accessTokenIDFromClaim))
		if delErr != nil {
			userID = nil
			deviceAuthorizationID = nil
			accessTokenID = nil
			userScopes = nil
			err = fmt.Errorf("Delete token with key: <token:%v>. Err: %+v",
				accessTokenIDFromClaim, delErr)
			return
		}

		result := database.GetDB(database.Default).Model(models.AccessToken{}).Where("id = ?",
			accessTokenIDFromClaim).Update("revoked_at", time.Now())
		if result.Error != nil || result.RowsAffected != 1 {
			err = fmt.Errorf("revoke access token <%v> error: %v",
				accessTokenIDFromClaim, result.Error)
			return
		}

		userID = nil
		deviceAuthorizationID = nil
		accessTokenID = nil
		userScopes = nil
		err = fmt.Errorf("JWT IP is different from request IP. JWT IP (%s). RequestIP (%s)",
			accessTokenPayload.IP, requestIP)
		return
	}

	userIDValue := uuid.FromStringOrNil(userIDFromClaim)
	userID = &userIDValue
	deviceAuthorizationIDValue := uuid.FromStringOrNil(devAuthIDFromClaim)
	deviceAuthorizationID = &deviceAuthorizationIDValue
	accessTokenIDValue := uuid.FromStringOrNil(accessTokenIDFromClaim)
	accessTokenID = &accessTokenIDValue
	for _, r := range accessTokenPayload.Roles {
		userScopes = append(userScopes, types.GetScopesOfRole(r)...)
	}
	devicePlatform = types.DevicePlatform(platformFromClaim)
	err = nil
	return
}
