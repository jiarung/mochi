package middleware

import (
	"strconv"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	apicontext "github.com/cobinhood/cobinhood-backend/common/api/context"
	apierrors "github.com/cobinhood/cobinhood-backend/common/api/errors"
	"github.com/cobinhood/cobinhood-backend/common/limiters"
	"github.com/cobinhood/cobinhood-backend/common/logging"
	"github.com/cobinhood/cobinhood-backend/common/utils"
	apiutils "github.com/cobinhood/cobinhood-backend/infra/api/utils"
)

func isLimitReachedAndSetHeader(ctx *apicontext.AppContext,
	cLimiter limiters.Limiter, key string) bool {
	ret := limiters.ReachLimitation(cLimiter, key)
	ctx.Writer().Header().Set("X-RateLimit-Limit",
		strconv.FormatInt(ret.Limit, 10))
	ctx.Writer().Header().Set("X-RateLimit-Period",
		strconv.FormatInt(ret.Seconds, 10))
	ctx.Writer().Header().Set("X-RateLimit-Reset",
		strconv.FormatInt(ret.ExpiredAt, 10))
	ctx.Writer().Header().Set("X-RateLimit-Remaining",
		strconv.FormatInt(ret.Limit-ret.Count, 10))
	return ret.Reached
}

// AuthRateLimitFunc accepts parameters for custom rate limit
// as WAF-like limiter, use user id only to create cache key
func AuthRateLimitFunc(limit int64, seconds int) gin.HandlerFunc {
	cLimiter := limiters.NewLimiter(limit, seconds)
	return func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			logging.NewLoggerTag("api:middleware:ratelimit").Error(
				"Error to get AppContext. Err: %v\n", err)
			ctx.Abort()
			return
		}
		if appCtx.IsPrivilegedIP() {
			return
		}

		key := "waf-auth-limiter:" + appCtx.UserID.String()
		if isLimitReachedAndSetHeader(appCtx, cLimiter, key) {
			appCtx.SetError(apierrors.TryAgainLater)
			return
		}
	}

}

// LimiterSelector is for selecting limiter
type LimiterSelector interface {
	SelectLimiter(db *gorm.DB, userID *uuid.UUID) limiters.Limiter
}

// URLAuthAPIRateLimitFunc accepts parameters for custom rate limit
// as WAF-like limiter, use URL path and user id to create cache key
func URLAuthAPIRateLimitFunc(limit int64, seconds int) gin.HandlerFunc {
	return urlAuthAPIRateLimitFunc(
		limiters.NewSimpleLimiterSelector(limit, seconds))
}

// URLAuthAPIVIPRateLimitFunc accepts parameters for custom rate limit
// as WAF-like limiter, use URL path and user id to create cache key
func URLAuthAPIVIPRateLimitFunc(
	limiterSelector LimiterSelector) gin.HandlerFunc {
	return urlAuthAPIRateLimitFunc(limiterSelector)
}

func urlAuthAPIRateLimitFunc(limiterSelector LimiterSelector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			logging.NewLoggerTag("api:middleware:ratelimit").Error(
				"Error to get AppContext. Err: %v\n", err)
			ctx.Abort()
			return
		}
		key := "waf-url-auth-limiter:" +
			ctx.Request.URL.String() +
			appCtx.UserID.String()

		l := limiterSelector.SelectLimiter(appCtx.DB, appCtx.UserID)

		if isLimitReachedAndSetHeader(appCtx, l, key) {
			appCtx.SetError(apierrors.TryAgainLater)
			return
		}
	}
}

// URLIPRateLimitFunc accepts parameters for custom rate limit
// as WAF-like limiter, use URL path and IP as key
func URLIPRateLimitFunc(limit int64, seconds int) gin.HandlerFunc {
	cLimiter := limiters.NewLimiter(limit, seconds)
	return func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			logging.NewLoggerTag("api:middleware:ratelimit").Error(
				"Error to get AppContext. Err: %v\n", err)
			ctx.Abort()
			return
		}

		if appCtx.IsPrivilegedIP() {
			return
		}
		if utils.IsStress() {
			return
		}

		key := "waf-url-ip-limiter:" +
			ctx.Request.URL.String() +
			apiutils.GetIPKey(ctx.Request)
		if isLimitReachedAndSetHeader(appCtx, cLimiter, key) {
			appCtx.SetError(apierrors.TryAgainLater)
			return
		}
	}
}

// WebsocketIPAPIRateLimiter is police-based middlware limits by IP 10 reqs/s
func WebsocketIPAPIRateLimiter(ctx *gin.Context) {
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		logging.NewLoggerTag("api:middleware:ratelimit").Error(
			"Error to get AppContext. Err: %v\n", err)
		ctx.Abort()
		return
	}

	if appCtx.IsPrivilegedIP() {
		return
	}

	if utils.IsStress() {
		return
	}

	if limiters.ReachWebsocketAPIIP10RPS(apiutils.GetIPKey(ctx.Request)) {
		appCtx.SetError(apierrors.TryAgainLater)
		return
	}
}

// WebsocketIPAPIRateLimiterFunc is function for utility and compatibility
func WebsocketIPAPIRateLimiterFunc() gin.HandlerFunc {
	return WebsocketIPAPIRateLimiter
}

// WebsocketConnectionAPIRateLimitFunc is function for
// utility and compatibility
func WebsocketConnectionAPIRateLimitFunc(maxConnections int) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCtx, err := apicontext.GetAppContext(c)
		if err != nil {
			logging.NewLoggerTag("api:middleware:ratelimit").Error(
				"Error to get AppContext. Err: %v\n", err)
			c.Abort()
			return
		}
		if appCtx.IsPrivilegedIP() {
			return
		}
		if utils.IsStress() {
			return
		}

		key := apiutils.GetIPKey(c.Request)
		if appCtx.IsAuthenticated() {
			key = appCtx.UserID.String()
		}

		if limiters.ReachWebsocketConnectionlimit(
			key, appCtx.RequestTag(), maxConnections) {
			appCtx.SetError(apierrors.TryAgainLater)
			return
		}

		appCtx.Logger().Debug("Valid connection. Key: %v", key)

		c.Set(limiters.WSConnLimitKey, key)
		// Conintue processing
		c.Next()

		// Clean up
		limiters.RemoveWebsocketConnection(key, appCtx.RequestTag())
	}
}
