package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	"github.com/cobinhood/mochi/common/logging"
)

var (
	nonceHeader   = "nonce"
	ignoreMethods = map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodOptions: true,
	}
	ignorePaths = map[string]bool{
		"/v1/oauth2/token":       true,
		"/v1/fiat/epay/redirect": true,
	}
	// prevent previous nonce set too large or last value forgot
	nonceTimeout = 86400
)

// prepare redis key, decouple from handlerfunc
func prepareNonceCacheKey(pk string) string {
	return fmt.Sprintf("api:middleware:nonce:%v", pk)
}

// NonceMiddleware is a handler function that limits request sequence
// to prevent duplicated request.
func NonceMiddleware(ctx *gin.Context) {

	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		logging.NewLoggerTag(ctx.GetString(logging.LabelTag)).Error(
			"Fail to obtain AppContext Error: %v\n", err)
		ctx.Abort()
		return
	}

	if _, ok := ignoreMethods[ctx.Request.Method]; ok {
		return
	}

	if _, ok := ignorePaths[ctx.Request.URL.Path]; ok {
		return
	}

	// Ignore nonce checking for epay deposit callback api
	if strings.HasPrefix(ctx.Request.URL.Path,
		"/v1/fiat/epay/deposit_callback/") {
		return
	}

	// Ignore nonce checking for epay in callback api
	if strings.HasPrefix(ctx.Request.URL.Path,
		"/v1/trading/callback/epay/") {
		return
	}

	// Ignore nonce checking for fiat general callbacks
	if strings.HasPrefix(ctx.Request.URL.Path,
		"/v1/trading/callbacks/") {
		return
	}

	userNonce, err := strconv.ParseInt(ctx.Request.Header.Get(nonceHeader), 10, 64)
	if err != nil {
		appCtx.Logger().Warn("Not numeric nonce. Error: %v\n", err)
		appCtx.SetError(apierrors.InvalidNonce)
		return
	}

	var userID uuid.UUID
	if userID, err = appCtx.GetUserID(); err != nil {
		// For public endpoints such as `login`.
		return
	}

	var key string
	if appCtx.IsAPIToken() {
		key = prepareNonceCacheKey(appCtx.APITokenID.String())
	} else {
		key = prepareNonceCacheKey(fmt.Sprintf("%s:%s", userID.String(), ctx.Request.URL.Path))
	}
	nonceLock := fmt.Sprintf("lock%v%v", key, userNonce) // versioned key
	rCli, release := appCtx.Cache.GetConn()
	defer release()
	if v, err := rCli.Do(
		"SET", nonceLock, appCtx.RequestTag(),
		"EX", nonceTimeout, "NX"); v != "OK" || err != nil {
		appCtx.Logger().Error("Duplicated Nonce: %v. Error: %v\n", userNonce, err)
		appCtx.SetError(apierrors.InvalidNonce)
		return
	}
	defer appCtx.Cache.Delete(nonceLock)

	if rawNonce, err := appCtx.Cache.Get(key); err == nil { // got token
		storedNonce, err := strconv.ParseInt(rawNonce.(string), 10, 64)
		if err != nil || userNonce <= storedNonce {
			appCtx.Logger().Error(
				"key: <%v>: Stored nonce (%v) >= userNonce (%v). Error: %v\n",
				key, storedNonce, userNonce, err,
			)
			appCtx.SetError(apierrors.InvalidNonce)
			return
		}
	}

	appCtx.Cache.Set(key, userNonce, nonceTimeout)
}

// NonceMiddlewareFunc returns NonceMiddlware for utility or compactibiliy
func NonceMiddlewareFunc() gin.HandlerFunc {
	return NonceMiddleware
}

// CleanNonceWithContext is shortcut function to clean stored nonce
func CleanNonceWithContext(ctx *gin.Context) {
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		logging.NewLoggerTag(ctx.GetString(logging.LabelTag)).Error(
			"Fail to obtain AppContext Error: %v\n", err)
		ctx.Abort()
		return
	}
	userID, _ := appCtx.GetUserID()
	appCtx.Cache.Delete(prepareNonceCacheKey(userID.String()))
}
