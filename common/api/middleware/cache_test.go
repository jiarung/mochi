// FIXME(hao): use middleware_test to avoid import cycle. should be removed
// after issue resolved.
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/apps/exchange/common/api/middlewares"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/common/api/apitest"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	apierrors "github.com/cobinhood/mochi/common/api/errors"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	"github.com/cobinhood/mochi/infra/app"
)

type cacheMiddlewareSuite struct {
	suite.Suite

	r *gin.Engine
}

func (s *cacheMiddlewareSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	gin.SetMode(gin.TestMode)

	cache.Initialize(config.Cache)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	s.Require().Nil(cache.GetRedis().FlushAll())

	s.r = gin.New()
	s.r.Use(
		logger.NewLoggerMiddleware,
		middlewares.ContextLogger(),
		middleware.AppContextMiddleware(cobxtypes.Test),
		middleware.ResponseHandler,
		middlewares.PanicLogger,
	)
	s.r.GET(
		"/",
		middleware.CacheMiddlewareFunc(2, middleware.CacheMethodGlobal),
		middleware.RequireAppContext(cobxtypes.Test,
			func(appCtx *apicontext.AppContext) {
				result := struct {
					Timestamp int64 `json:"ts"`
				}{
					time.Now().UTC().UnixNano(),
				}
				time.Sleep(time.Second)
				appCtx.SetJSON(result)
			},
		),
	)

	entered := false
	s.r.GET(
		"/flaky_handler",
		middleware.CacheMiddlewareFunc(2, middleware.CacheMethodGlobal),
		middleware.RequireAppContext(cobxtypes.Test,
			func(appCtx *apicontext.AppContext) {
				if !entered {
					entered = true
					appCtx.Abort()
					appCtx.SetError(apierrors.UnexpectedError)
					return
				}
				appCtx.SetJSON(struct{ Status string }{"OK"})
			},
		),
	)
}

func (s *cacheMiddlewareSuite) TearDownSuite() {
	s.Require().Nil(cache.GetRedis().FlushAll())
	cache.Finalize()
	database.Finalize()
}

func (s *cacheMiddlewareSuite) TestMiddleware() {
	var wg sync.WaitGroup
	var res *httptest.ResponseRecorder
	go func() {
		wg.Add(1)
		defer wg.Done()
		res = apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	}()
	time.Sleep(500 * time.Millisecond)

	// should get try again later
	res2 := apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	s.Require().Equal(res2.Code, http.StatusTooManyRequests)

	// wait until request done.
	wg.Wait()

	// should get the cache
	res3 := apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	s.Require().Equal(res.Body.String(), res3.Body.String())

	// sleep until cache is expired.
	time.Sleep(2 * time.Second)

	var res4 *httptest.ResponseRecorder
	go func() {
		wg.Add(1)
		defer wg.Done()
		res4 = apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	}()
	time.Sleep(100 * time.Millisecond)

	// should get the old cache
	res5 := apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	s.Require().Equal(res.Body.String(), res5.Body.String())

	// wait until request done.
	wg.Wait()

	// should get the new cache
	res6 := apitest.PerformRequest(s.r, http.MethodGet, "/", nil)
	s.Require().Equal(res4.Body.String(), res6.Body.String())
}

func (s *cacheMiddlewareSuite) TestFailedRequestRetry() {
	var res *httptest.ResponseRecorder
	res = apitest.PerformRequest(s.r, http.MethodGet, "/flaky_handler", nil)
	s.Require().Equal(res.Code, http.StatusInternalServerError)

	// Second retry should work (not cached)
	res = apitest.PerformRequest(s.r, http.MethodGet, "/flaky_handler", nil)
	s.Require().Equal(res.Code, http.StatusOK)
}

func TestCacheMiddleware(t *testing.T) {
	suite.Run(t, new(cacheMiddlewareSuite))
}
