package middleware

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/common/api/apitest"
	"github.com/cobinhood/cobinhood-backend/common/limiters"
	"github.com/cobinhood/cobinhood-backend/common/logging"
	"github.com/cobinhood/cobinhood-backend/database"
	"github.com/cobinhood/cobinhood-backend/database/exchangedb"
	"github.com/cobinhood/cobinhood-backend/infra/api/middleware/logger"
	"github.com/cobinhood/cobinhood-backend/infra/api/utils"
	"github.com/cobinhood/cobinhood-backend/infra/app"
)

const (
	testIPRequestCount      = 20
	testConnectRequestCount = 2
	testMaxConnectionCount  = 1
	testConnectDelayUnit    = 10
	testUserRequestCount    = 20
)

var delayCount int32

func connectionIDGenerator(ctx *gin.Context) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	ctx.Set(logging.LabelTag, strconv.Itoa(r1.Intn(1000000)))
}

func mockConnectionDelay(ctx *gin.Context) {
	count := atomic.AddInt32(&delayCount, 1)
	time.Sleep(
		time.Millisecond * time.Duration(testConnectDelayUnit) * time.Duration(count),
	)
	ctx.Next()
}

type mockLimiterSelector struct {
	m map[string]limiters.Limiter
}

func (s *mockLimiterSelector) SelectLimiter(db *gorm.DB,
	userID *uuid.UUID) limiters.Limiter {
	return s.m[userID.String()]
}

func getTestEngine() *gin.Engine {
	r := gin.Default()
	r.Use(logger.NewLoggerMiddleware)
	r.Use(
		ResponseHandler,
		connectionIDGenerator,
		AppContextMiddleware(cobxtypes.Test),
		setUserIDfromRequestHeader, // borrow from nonce_test.go
	)
	r.GET("/ws/1", WebsocketIPAPIRateLimiterFunc(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	r.GET("/ws/2",
		mockConnectionDelay,
		WebsocketConnectionAPIRateLimitFunc(testMaxConnectionCount),
		func(ctx *gin.Context) {
			// fake websocket long connection
			time.Sleep(time.Millisecond * time.Duration(100))
			ctx.JSON(http.StatusOK, gin.H{"success": true})
		})
	r.GET("/cip", URLIPRateLimitFunc(1, 2), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	r.GET("/cauth", URLAuthAPIRateLimitFunc(2, 1), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	r.GET("/auth1", AuthRateLimitFunc(2, 1), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	r.GET("/auth2", AuthRateLimitFunc(2, 1), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})

	l := mockLimiterSelector{
		m: map[string]limiters.Limiter{
			"17ea83af-53a5-4d70-ac30-113dee97c7b9": limiters.NewLimiter(2, 1),
			"17ea83af-53a5-4d70-ac30-113dee97c7b1": limiters.NewLimiter(2, 1),
			"17ea83af-53a5-4d70-ac30-113dee97c7b2": limiters.NewLimiter(3, 1),
			"17ea83af-53a5-4d70-ac30-113dee97c7b3": limiters.NewLimiter(4, 1),
		},
	}
	r.GET("/cvipauth", URLAuthAPIVIPRateLimitFunc(&l), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})

	return r
}

func (r *RateLimitSuite) TestLimitByWebsocketIPWithRemoteAddr() {
	testingIP := "140.118.118.118"
	limiters.ClearWebsocketIPBlackList(testingIP)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/ws/1", nil)
	ctx.Request.RemoteAddr = testingIP

	var success, fail int
	for i := 0; i < testIPRequestCount; i++ {
		w := apitest.PerformRequest(r.e, "GET", "/ws/1", ctx.Request)
		switch w.Code {
		case 200:
			success++
		default:
			fail++
		}
	}

	r.Require().False(
		fail != testIPRequestCount-10,
		"Remote addr limit fails: %v != 10 (success), or %v != 10 (fail)\n",
		success,
		fail,
	)

	success = 0
	fail = 0
	ctx.Request.RemoteAddr = "100.100.100.100"
	for i := 0; i < testIPRequestCount; i++ {
		w := apitest.PerformRequest(r.e, "GET", "/ws/1", ctx.Request)
		switch w.Code {
		case 200:
			success++
		default:
			fail++
		}
	}

	r.Require().False(
		success != testIPRequestCount,
		"Remote addr limit fails: %v != 10 (success), or %v != 10 (fail)\n",
		success,
		fail,
	)
}

func (r *RateLimitSuite) TestLimitByWebsocketConnectionWithRemoteAddr() {
	testingIP := "140.118.118.118"
	limiters.ClearWebsocketConnectionLimit(testingIP)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/ws/2", nil)
	ctx.Request.RemoteAddr = testingIP

	failtimeout := time.After(time.Second * time.Duration(5))
	respChan := make(chan *httptest.ResponseRecorder, testConnectRequestCount)
	var success, fail, unknown int
	for i := 0; i < testConnectRequestCount; i++ {
		go func() {
			w := apitest.PerformRequest(r.e, http.MethodGet, "/ws/2", ctx.Request)
			respChan <- w
		}()
	}

	for i := 0; i < testConnectRequestCount; i++ {
		select {
		case <-failtimeout:
			unknown++
		case w := <-respChan:
			switch w.Code {
			case 200:
				success++
			default:
				fail++
			}
		}
	}

	r.Require().False(
		fail != testConnectRequestCount-testMaxConnectionCount,
		"None auth websocket connection limit fails: "+
			"%v != %v (success), or %v != %v (fail)\n",
		success,
		testMaxConnectionCount,
		fail,
		testConnectRequestCount-testMaxConnectionCount,
	)

	// For privilegedIP
	success = 0
	fail = 0
	unknown = 0
	ctx.Request.RemoteAddr = "100.100.100.100"
	for i := 0; i < testConnectRequestCount; i++ {
		go func() {
			w := apitest.PerformRequest(r.e, http.MethodGet, "/ws/2", ctx.Request)
			respChan <- w
		}()
	}

	for i := 0; i < testConnectRequestCount; i++ {
		select {
		case <-failtimeout:
			unknown++
		case w := <-respChan:
			switch w.Code {
			case 200:
				success++
			default:
				fail++
			}
		}
	}

	r.Require().False(
		fail != 0,
		"None auth websocket connection limit fails: "+
			"%v != %v (success), or %v != %v (fail)\n",
		success,
		testMaxConnectionCount,
		fail,
		testConnectRequestCount-testMaxConnectionCount,
	)

	logging.NewLoggerTag("TEST").Notice("Configured Auth websocket")

	// If user login, the connection should be accepted. Because user key is
	// different from original one which is shared by all public connections.
	testingID := "17ea83af-53a5-4d70-ac30-113dee97c7b8"
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/ws/2", nil)
	ctx.Request.RemoteAddr = testingIP
	ctx.Request.Header.Set("user_id", testingID)

	authFailtimeout := time.After(time.Second * time.Duration(5))
	for i := 0; i < testConnectRequestCount; i++ {
		go func() {
			w := apitest.PerformRequest(r.e, http.MethodGet, "/ws/2", ctx.Request)
			respChan <- w
		}()
	}

	var authSuccess, authFail, authUnknown int
	for i := 0; i < testConnectRequestCount; i++ {
		select {
		case <-authFailtimeout:
			authUnknown++
		case w := <-respChan:
			switch w.Code {
			case 200:
				authSuccess++
			default:
				authFail++
			}
		}
	}

	r.Require().False(
		authFail != testConnectRequestCount-testMaxConnectionCount,
		"Auth websocket connection limit fails: "+
			"%v != %v (success), or %v != %v (fail)\n",
		authSuccess,
		testMaxConnectionCount,
		authFail,
		testConnectRequestCount-testMaxConnectionCount,
	)

	// For privileged IP.
	authSuccess = 0
	authFail = 0
	authUnknown = 0
	ctx.Request.RemoteAddr = "100.100.100.100"
	for i := 0; i < testConnectRequestCount; i++ {
		go func() {
			w := apitest.PerformRequest(r.e, http.MethodGet, "/ws/2", ctx.Request)
			respChan <- w
		}()
	}

	for i := 0; i < testConnectRequestCount; i++ {
		select {
		case <-authFailtimeout:
			authUnknown++
		case w := <-respChan:
			switch w.Code {
			case 200:
				authSuccess++
			default:
				authFail++
			}
		}
	}

	r.Require().False(
		authFail != 0,
		"Auth websocket connection limit fails: "+
			"%v != %v (success), or %v != %v (fail)\n",
		authSuccess,
		testMaxConnectionCount,
		authFail,
		testConnectRequestCount-testMaxConnectionCount,
	)
}

type RateLimitSuite struct {
	suite.Suite

	e                     *gin.Engine
	testPrivilegedIPRange string
	testPrivilegedIP      string
}

func (r *RateLimitSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	cache.Initialize(config.Cache)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)

	r.e = getTestEngine()

	r.testPrivilegedIPRange = "100.100.100.100/32"
	r.testPrivilegedIP = "100.100.100.100"
	utils.ResetPrivilegedIPRange(r.testPrivilegedIPRange)
}

func (r *RateLimitSuite) SetupTest() {
	atomic.StoreInt32(&delayCount, 0)
}

func (r *RateLimitSuite) TearDownSuite() {
	cache.Finalize()
	database.Finalize()
}

func (r *RateLimitSuite) TestURLIPLimiter() {
	testingIP := "140.116.118.112"
	req := httptest.NewRequest(http.MethodGet, "/cip", nil)
	req.RemoteAddr = testingIP

	w := apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

	time.Sleep(time.Second * 2)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	req.RemoteAddr = r.testPrivilegedIP
	for i := 0; i < 5; i++ {
		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusOK, w.Code)
	}
}

func (r *RateLimitSuite) TestRateLimitHeaderSet() {
	testURL := "/rate_header_test"
	testLimitCount := int64(10)
	testLimitPeriod := 10

	r.e.GET(testURL, URLIPRateLimitFunc(testLimitCount, testLimitPeriod),
		func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"success": true})
		})
	req := httptest.NewRequest(http.MethodGet, testURL, nil)
	req.RemoteAddr = "140.116.118.112"

	for i := int64(0); i <= testLimitCount+2; i++ {
		w := apitest.PerformRequest(r.e, "", "", req)
		if i >= testLimitCount {
			require.Equal(r.T(), http.StatusTooManyRequests, w.Code)
		} else {
			require.Equal(r.T(), http.StatusOK, w.Code)
		}

		headers := w.Header()
		remainingStr, ok := headers["X-Ratelimit-Remaining"]
		require.True(r.T(), ok)
		remaining, err := strconv.ParseInt(remainingStr[0], 10, 64)
		require.Nil(r.T(), err)
		require.Equal(r.T(), remaining, testLimitCount-i-1)
		limitStr, ok := headers["X-Ratelimit-Limit"]
		require.True(r.T(), ok)
		limit, err := strconv.ParseInt(limitStr[0], 10, 64)
		require.Nil(r.T(), err)
		require.Equal(r.T(), limit, testLimitCount)
		periodStr, ok := headers["X-Ratelimit-Period"]
		require.True(r.T(), ok)
		period, err := strconv.ParseInt(periodStr[0], 10, 64)
		require.Nil(r.T(), err)
		require.Equal(r.T(), period, int64(testLimitPeriod))
		resetStr, ok := headers["X-Ratelimit-Reset"]
		require.True(r.T(), ok)
		reset, err := strconv.ParseInt(resetStr[0], 10, 64)
		require.Nil(r.T(), err)
		// Check time may create the flaky test. Just check exist.
		require.True(r.T(), reset > 0)
	}
}

func (r *RateLimitSuite) TestURLAuthLimiter() {
	testingID := "17ea83af-53a5-4d70-ac30-113dee97c7b9"
	req := httptest.NewRequest(http.MethodGet, "/cauth", nil)
	req.Header.Set("user_id", testingID)

	w := apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

	time.Sleep(time.Second * 2)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	// mutliple users
	testingIDs := []string{
		"17ea83af-53a5-4d70-ac30-113dee97c7b1",
		"17ea83af-53a5-4d70-ac30-113dee97c7b2",
		"17ea83af-53a5-4d70-ac30-113dee97c7b3",
	}

	for i := 0; i < len(testingIDs); i++ {
		testingID = testingIDs[i]
		req := httptest.NewRequest(http.MethodGet, "/cauth", nil)
		req.Header.Set("user_id", testingID)

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusOK, w.Code)

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusOK, w.Code)

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

		time.Sleep(time.Second * 2)

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusOK, w.Code)
	}
}

func (r *RateLimitSuite) TestURLAuthVIPLimiter() {
	testingID := "17ea83af-53a5-4d70-ac30-113dee97c7b9"
	req := httptest.NewRequest(http.MethodGet, "/cvipauth", nil)
	req.Header.Set("user_id", testingID)

	w := apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

	time.Sleep(time.Second * 2)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	// mutliple users
	testingIDs := []string{
		"17ea83af-53a5-4d70-ac30-113dee97c7b1",
		"17ea83af-53a5-4d70-ac30-113dee97c7b2",
		"17ea83af-53a5-4d70-ac30-113dee97c7b3",
	}

	for i := 0; i < len(testingIDs); i++ {
		testingID = testingIDs[i]
		req := httptest.NewRequest(http.MethodGet, "/cvipauth", nil)
		req.Header.Set("user_id", testingID)

		for j := 0; j < i+2; j++ {
			w = apitest.PerformRequest(r.e, "", "", req)
			require.Equal(r.T(), http.StatusOK, w.Code)
		}

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

		time.Sleep(time.Second * 2)

		w = apitest.PerformRequest(r.e, "", "", req)
		require.Equal(r.T(), http.StatusOK, w.Code)
	}
}

func (r *RateLimitSuite) TestGlobalAuthLimiter() {
	testingID := "17ea83af-53a5-4d70-ac30-113dee97c7b8"
	req := httptest.NewRequest(http.MethodGet, "/auth1", nil)
	req.Header.Set("user_id", testingID)

	w := apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/auth2", nil)
	req.Header.Set("user_id", testingID)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusTooManyRequests, w.Code)

	time.Sleep(time.Second * 2)

	w = apitest.PerformRequest(r.e, "", "", req)
	require.Equal(r.T(), http.StatusOK, w.Code)

}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}
