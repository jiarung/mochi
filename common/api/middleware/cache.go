package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jiarung/mochi/cache"
	apicontext "github.com/jiarung/mochi/common/api/context"
	apierrors "github.com/jiarung/mochi/common/api/errors"
)

const noWritten = -1

// CacheMethod is a function type used to control the scope of caching.
type CacheMethod func(writer io.Writer, appCtx *apicontext.AppContext) error

// CacheMethodGlobal can be used at endpoints whose response doesn't change on
// different clients. The cached content should be valid for all clients.
func CacheMethodGlobal(
	writer io.Writer, appCtx *apicontext.AppContext) error {

	// Do nothing if global cache is used.
	return nil
}

// CacheMethodByRequestIP can be used at endpoints whose response may vary on
// clients from different IP addresses.
func CacheMethodByRequestIP(
	writer io.Writer, appCtx *apicontext.AppContext) error {

	// Encode IP address by its binary representation.
	_, err := writer.Write(appCtx.RequestRawIP)
	return err
}

// CacheMethodByAuthorization can be used at endpoints whose response is
// specific to single user and may be different on different sessions.
func CacheMethodByAuthorization(
	writer io.Writer, appCtx *apicontext.AppContext) error {

	// Encode authorization info if it is provided.
	_, err := io.WriteString(writer, appCtx.Request().Header.Get("Authorization"))
	return err
}

type cacheKey struct {
	tsKey    string
	valueKey string
}

func generateCacheKey(
	appCtx *apicontext.AppContext, method CacheMethod) (*cacheKey, error) {

	h := md5.New()

	err := method(h, appCtx)
	if err != nil {
		return nil, err
	}

	// Encode by url.
	_, err = io.WriteString(h, appCtx.Request().URL.String())
	if err != nil {
		return nil, err
	}

	keyPrefix := fmt.Sprintf("%s:cache:middleware:%s",
		string(appCtx.ServiceName), hex.EncodeToString(h.Sum(nil)))
	return &cacheKey{keyPrefix + ":ts", keyPrefix + ":value"}, nil
}

// CacheMiddlewareFunc caches response if requeset succeeds.
func CacheMiddlewareFunc(seconds int, method CacheMethod) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Only cache the GET handlers.
		if ctx.Request.Method != http.MethodGet {
			ctx.Next()
			return
		}

		appCtx, err := apicontext.GetAppContext(ctx)
		if err != nil {
			ctx.Abort()
			return
		}
		logger := appCtx.Logger()

		cacheSec := seconds + 20
		var key *cacheKey
		for {
			resetTsCache := func(k string) {
				err = cache.GetRedis().Set(k, time.Now().Unix(), cacheSec)
				if err != nil {
					logger.Error("failed to set cache timestamp. err: %v", err)
				}
			}

			key, err = generateCacheKey(appCtx, method)
			if err != nil {
				logger.Error("failed to generate cache key. err: %v", err)
				break
			}
			ts, err := cache.GetRedis().GetInt(key.tsKey)
			if err != nil {
				// No cached ts is found. Set the ts and fallback to handler
				// handling.
				logger.Debug("failed to get cached timestamp. err: %v", err)
				resetTsCache(key.tsKey)
				break
			}

			// Set ts to current timestamp and go into the handler. The next
			// incoming request will get the old cache before the handler
			// ends handling request.
			if ts < int(time.Now().Unix())-seconds {
				logger.Debug("cache timeout")
				resetTsCache(key.tsKey)
				break
			}

			// Within the cache period, return the cached value.
			value, err := cache.GetRedis().Get(key.valueKey)
			if err != nil {
				// No cache found for the very first "2nd" request of one
				// handler. The ts is set and the handler hasn't ended.
				logger.Warn("failed to get cached value. err: %v", err)
				appCtx.SetError(apierrors.TryAgainLater)
				return
			}

			// Always respond the cached value whether it is expired or not.
			appCtx.SetResp(gin.MIMEJSON, []byte(value.(string)))
			appCtx.SetIgnoreAndAbort()
			return
		}

		// Forward to handler.
		ctx.Next()

		// return for error or no key.
		if appCtx.IsAborted() || key == nil {
			err = cache.GetRedis().Delete(key.tsKey)
			if err != nil {
				logger.Warn("failed to delete cache timestamp. err: %v", err)
			}
			return
		}

		data, err := json.Marshal(appCtx.JSON())
		if err != nil {
			logger.Error("json marshal failed. err: %s", err)
			return
		}
		err = cache.GetRedis().Set(key.valueKey, string(data), cacheSec)
		if err != nil {
			logger.Error("failed to write cache. err: %v", err)
		}
	}
}
