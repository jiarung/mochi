package middleware

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/cache"
	apicontext "github.com/jiarung/mochi/common/api/context"
	"github.com/jiarung/mochi/common/config"
	"github.com/jiarung/mochi/common/config/misc"
	"github.com/jiarung/mochi/common/logging"
	"github.com/jiarung/mochi/database"
	"github.com/jiarung/mochi/database/exchangedb"
	"github.com/jiarung/mochi/infra/api/middleware/logger"
	"github.com/jiarung/mochi/infra/api/utils"
	"github.com/jiarung/mochi/infra/app"
	"github.com/jiarung/mochi/messaging"
)

// Test s for AppContext middleware
type appContextTestSuite struct {
	suite.Suite
	requestBody []byte
}

func (s *appContextTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	gin.SetMode(gin.TestMode)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	messaging.Initialize(messaging.Mock)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
}

func (s *appContextTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
	messaging.Finalize()
}

func (s *appContextTestSuite) SetupTest() {
	// Generate random octets for body
	rand.Seed(time.Now().UnixNano())
	s.requestBody = make([]byte, rand.Intn(1000)+500)
	rand.Read(s.requestBody)
}

func (s *appContextTestSuite) TestAppContextActivation() {
	// Create test context and attach middleware
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.RemoteAddr = "118.112.113.114"
	logger.NewLoggerMiddleware(ctx)
	ctx.Set(logging.LabelTag, "fake-request-id") // mock context-logger.go
	ctx.Set(config.RequestBody, s.requestBody)   // mock body-reader.go
	AppContextMiddleware(cobxtypes.Test)(ctx)

	// Check AppContext exists
	appCtxPtr, exists := ctx.Get(config.AppContext)
	require.True(s.T(), exists)
	require.NotNil(s.T(), appCtxPtr)

	// Verify AppContext members
	appCtx := *appCtxPtr.(*apicontext.AppContext)
	assert.NotNil(s.T(), appCtx.DB)
	assert.NotNil(s.T(), appCtx.Cache)
	assert.NotEmpty(s.T(), appCtx.RequestTag)
	assert.Nil(s.T(), appCtx.UserID)
	assert.Len(s.T(), appCtx.RequestBody(), len(s.requestBody))
	assert.Equal(s.T(), "118.112.113.114", appCtx.RequestIP)
}

func (s *appContextTestSuite) TestNoIP() {
	// Create test context and attach middleware
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	logger.NewLoggerMiddleware(ctx)
	ctx.Set(logging.LabelTag, "fake-request-id") // mock context-logger.go
	ctx.Set(config.RequestBody, s.requestBody)   // mock body-reader.go
	AppContextMiddleware(cobxtypes.Test)(ctx)

	// Check AppContext exists
	appCtxPtr, exists := ctx.Get(config.AppContext)
	require.True(s.T(), exists)
	require.NotNil(s.T(), appCtxPtr)

	// Verify AppContext members
	appCtx := *appCtxPtr.(*apicontext.AppContext)
	assert.NotEmpty(s.T(), appCtx.RequestIP)
}

func BenchmarkAppContextMiddleware(b *testing.B) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	logger.NewLoggerMiddleware(ctx)
	for i := 0; i < b.N; i++ {
		AppContextMiddleware(cobxtypes.Test)(ctx)
	}
}

func (s *appContextTestSuite) TestGetAppContext() {
	// Create test context and attach middleware
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	logger.NewLoggerMiddleware(ctx)
	ctx.Set(logging.LabelTag, "fake-request-id") // mock context-logger.go
	ctx.Set(config.RequestBody, s.requestBody)   // mock body-reader.go
	AppContextMiddleware(cobxtypes.Test)(ctx)

	// Test getting AppContext using utils package
	appCtx, err := apicontext.GetAppContext(ctx)
	require.NoError(s.T(), err)

	// Verify AppContext members
	assert.NotNil(s.T(), appCtx.DB)
	assert.NotNil(s.T(), appCtx.Cache)
	assert.NotEmpty(s.T(), appCtx.RequestTag)
	assert.Nil(s.T(), appCtx.UserID)
	assert.Len(s.T(), appCtx.RequestBody(), len(s.requestBody))
	assert.Equal(s.T(), appCtx.ServiceName, cobxtypes.Test)
}

func BenchmarkGetAppContext(b *testing.B) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	logger.NewLoggerMiddleware(ctx)
	AppContextMiddleware(cobxtypes.Test)(ctx)
	for i := 0; i < b.N; i++ {
		apicontext.GetAppContext(ctx)
	}
}

func (s *appContextTestSuite) TestSetAppContext() {
	// Create test context and attach middleware
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.Header.Set("X-Forwarded-For", "127.0.0.1")
	logger.NewLoggerMiddleware(ctx)
	AppContextMiddleware(cobxtypes.Test)(ctx)

	// Set UserID member
	userID := uuid.NewV4()
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		panic(err)
	}
	appCtx.UserID = &userID

	// Test getting AppContext using utils package
	newCtx, err := apicontext.GetAppContext(ctx)
	require.NoError(s.T(), err)

	// Make sure assigned member values equal
	userIDfromNewCtx, err := newCtx.GetUserID()
	require.Nil(s.T(), err)
	assert.Equal(s.T(), userID, userIDfromNewCtx)
}

func (s *appContextTestSuite) TestSetRequestIP() {
	var err error

	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	logger.NewLoggerMiddleware(ginCtx)

	appCtx, err := apicontext.NewAppCtx(
		ginCtx,
		logging.NewLogger(),
		database.GetDB(database.Default),
		cache.GetRedis(),
	)
	require.Nil(
		s.T(),
		err,
		"TestSetRequestIP: context.NewAppCtx() failed. err(%s)",
		err)

	// No IP
	utils.InvalidatePrivilegedIPCache()
	misc.SetPrivilegedIpRange("")

	testCases := map[string]bool{
		"172.17.0.3/32|172.17.0.2":                   false,
		"172.17.0.3/32|172.17.0.3":                   true,
		"172.17.0.3/32|172.17.0.4":                   false,
		"172.17.0.3/24|172.17.2.255":                 false,
		"172.17.0.3/24|172.17.3.1":                   false,
		"172.17.3.3/24|172.17.3.1":                   true,
		"172.17.3.3/24,172.17.10.3/24|172.17.2.255":  false,
		"172.17.3.3/24,172.17.10.3/24|172.17.3.1":    true,
		"172.17.3.3/24,172.17.10.3/24|172.17.3.255":  true,
		"172.17.3.3/24,172.17.10.3/24|172.17.4.1":    false,
		"172.17.3.3/24,172.17.10.3/24|172.17.9.255":  false,
		"172.17.3.3/24,172.17.10.3/24|172.17.10.1":   true,
		"172.17.3.3/24,172.17.10.3/24|172.17.10.255": true,
		"172.17.3.3/24,172.17.10.3/24|172.17.11.1":   false,
	}

	for testCase, isPrivilegedIP := range testCases {
		ipRange := strings.Split(testCase, "|")[0]
		ip := strings.Split(testCase, "|")[1]
		utils.ResetPrivilegedIPRange(ipRange)
		appCtx.Request().Header.Set("X-Forwarded-For", ip)
		s.Require().Nil(appCtx.SetRequestIP())
		if isPrivilegedIP {
			s.Require().True(appCtx.IsPrivilegedIP())
		} else {
			s.Require().False(appCtx.IsPrivilegedIP())
		}
	}
}

func TestAppContextMiddleware(test *testing.T) {
	suite.Run(test, new(appContextTestSuite))
}
