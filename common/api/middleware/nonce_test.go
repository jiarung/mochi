package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/jiarung/mochi/apps/exchange/cobx-types"
	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/common/api/apitest"
	apicontext "github.com/jiarung/mochi/common/api/context"
	"github.com/jiarung/mochi/database"
	"github.com/jiarung/mochi/database/exchangedb"
	"github.com/jiarung/mochi/infra/api/middleware/logger"
	"github.com/jiarung/mochi/infra/app"
)

// this is helper middleware since access token API is not ready
func setUserIDfromRequestHeader(ctx *gin.Context) {
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		panic(err)
	}
	userID, _ := uuid.FromString(ctx.Request.Header.Get("user_id"))
	appCtx.UserID = &userID
	ctx.Next()
}

type TestNonceSuite struct {
	suite.Suite

	router *gin.Engine
}

func (s *TestNonceSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	gin.SetMode(gin.TestMode)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)

	s.router = obtainMockRouter()
}

func (s *TestNonceSuite) TearDownTest() {
	testuserIDs := []string{
		"17ea83af-53a5-4d70-ac30-113dee97c7b9",
		"2ba89329-2547-47a9-b0fb-6700762610c5",
		"359c9915-c7f1-4bb5-8993-dc7f61021a7b",
	}

	cacheClient := cache.GetRedis()

	for _, v := range testuserIDs {
		cacheClient.Delete("", prepareNonceCacheKey(v+":/nonce/test")) // delete old key
	}
}

func (s *TestNonceSuite) TearDownSuite() {
	cache.Finalize()
	database.Finalize()
}

func (s *TestNonceSuite) TestPrepareKey() {

	testSuite := []struct {
		key      string
		testWith string
		expected bool
	}{
		{
			"17ea83af-53a5-4d70-ac30-113dee97c7b9",
			"api:middleware:nonce:17ea83af-53a5-4d70-ac30-113dee97c7b9",
			true,
		},
		{
			"2ba89329-2547-47a9-b0fb-6700762610c5",
			"api:middleware:nonce:2ba89329-2547-47a9-b0fb-6700762610c5",
			true,
		},
	}

	for k, v := range testSuite {
		s.Require().True(
			prepareNonceCacheKey(v.key) == v.testWith,
			"Index: %v. %v != %v (Expected). key: %v != %v\n",
			k,
			prepareNonceCacheKey(v.key) == v.testWith,
			v.expected,
			prepareNonceCacheKey(v.key),
			v.testWith,
		)
	}
}

func (s *TestNonceSuite) TestPrepareKeyForEachPath() {
	redis := cache.GetRedis()
	redis.Delete(fmt.Sprintf(
		"api:middleware:nonce:17ea83af-53a5-4d70-ac30-113dee97c7b9:/nonce/test"))
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodPost, "/nonce/test", nil)
	req.Header.Set("user_id", "17ea83af-53a5-4d70-ac30-113dee97c7b9")
	req.Header.Set(nonceHeader, fmt.Sprint(time.Now().UnixNano()))
	ctx.Request = req
	logger.NewLoggerMiddleware(ctx)
	AppContextMiddleware(cobxtypes.Test)(ctx)
	setUserIDfromRequestHeader(ctx)
	NonceMiddleware(ctx)

	_, err := redis.Get(
		"api:middleware:nonce:17ea83af-53a5-4d70-ac30-113dee97c7b9:/nonce/test")
	s.Require().Nil(err)
}

func BenchmarkPrepareKey(b *testing.B) {
	userID := "17ea83af-53a5-4d70-ac30-113dee97c7b9"
	for i := 0; i < b.N; i++ {
		prepareNonceCacheKey(userID)
	}
}

func obtainMockRouter() *gin.Engine {
	r := gin.Default()
	r.Use(logger.NewLoggerMiddleware)
	nonce := r.Group("/nonce")
	nonce.Use(
		ResponseHandler,
		AppContextMiddleware(cobxtypes.Test),
		setUserIDfromRequestHeader,
		NonceMiddlewareFunc())
	nonce.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	nonce.PUT("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	nonce.PATCH("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	nonce.POST("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"success": true})
	})
	nonce.DELETE("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusNoContent, gin.H{"success": true})
	})
	nonce.OPTIONS("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	nonce.HEAD("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})
	return r
}

func (s *TestNonceSuite) TestNonceMiddlewareNaiveRequest() {
	// Naive request
	naiveTestSuite := []struct {
		Method string
		Status int
	}{
		{http.MethodGet, http.StatusOK}, {http.MethodHead, http.StatusOK}, {http.MethodOptions, http.StatusOK},
		{http.MethodPut, http.StatusConflict}, {http.MethodPost, http.StatusConflict},
		{http.MethodPatch, http.StatusConflict}, {http.MethodDelete, http.StatusConflict},
	}

	for k, v := range naiveTestSuite {
		req := httptest.NewRequest(v.Method, "/nonce/test", nil)
		w := apitest.PerformRequest(s.router, v.Method, "/nonce/test", req)
		s.Require().False(
			w.Code != v.Status,
			"Index :%v. Wrong expected response. StatusCode: %v != %v(Expected). Body: %v\n",
			k, w.Code, v.Status, w.Body.String(),
		)
	}
}

func (s *TestNonceSuite) TestSerialRequests() {
	serialTestSuite := []struct {
		Method string
	}{
		{http.MethodGet}, {http.MethodHead}, {http.MethodPut}, {http.MethodPatch},
		{http.MethodPost}, {http.MethodDelete}, {http.MethodOptions},
		{http.MethodGet}, {http.MethodHead}, {http.MethodPut}, {http.MethodPatch},
		{http.MethodPost}, {http.MethodDelete}, {http.MethodOptions},
	}

	for k, v := range serialTestSuite {
		req, err := http.NewRequest(v.Method, "/nonce/test", nil)
		s.Require().Nil(
			err,
			"Index: %v. Fail to create request. Err: %v\n",
			k,
			err,
		)
		req.Header.Set(nonceHeader, fmt.Sprint(time.Now().UnixNano()))
		req.Header.Set("user_id", "17ea83af-53a5-4d70-ac30-113dee97c7b9")
		w := apitest.PerformRequest(s.router, v.Method, "/nonce/test", req)
		s.Require().False(
			w.Code >= http.StatusMultipleChoices,
			"Index: %v. Fail during serial rquest: %v\n",
			k,
			v,
		)
	}
}

func (s *TestNonceSuite) TestNonceConcurrentRequest() {
	token := fmt.Sprint(time.Now().UnixNano())

	concurrentNum := 10
	respChan := make(chan *httptest.ResponseRecorder, concurrentNum)

	for i := 0; i < concurrentNum; i++ {
		go func() {
			req, err := http.NewRequest(http.MethodPost, "/nonce/test", nil)
			if err != nil {
				respChan <- &httptest.ResponseRecorder{
					Code: http.StatusInternalServerError,
				}
				return
			}
			req.Header.Set("user_id", "17ea83af-53a5-4d70-ac30-113dee97c7b9")
			req.Header.Set(nonceHeader, token)
			w := apitest.PerformRequest(s.router, http.MethodPost, "/nonce/test", req)
			respChan <- w

		}()
	}

	failtimeout := time.After(time.Minute * time.Duration(1))
	count200, count409, unknow := 0, 0, 0
	for i := 0; i < concurrentNum; i++ {
		select {
		case <-failtimeout:
			unknow++
		case v := <-respChan:
			switch v.Code {
			case http.StatusCreated:
				count200++
			case http.StatusConflict:
				count409++
			default:
				unknow++
			}
		}
	}

	s.Require().False(
		count200 != 1 || count409 != 9 || unknow != 0,
		"Nonce is not robust: %v, %v, %v\n",
		count200, count409, unknow,
	)

}

func BenchmarkNonceMiddleware(b *testing.B) {
	r := obtainMockRouter()
	token := fmt.Sprint(time.Now().UnixNano())
	req, err := http.NewRequest(http.MethodPost, "/nonce/test", nil)
	if err != nil {
		b.Fatal("Can't create testing request")
		return
	}
	req.Header.Set("user_id", "17ea83af-53a5-4d70-ac30-113dee97c7b9")
	req.Header.Set(nonceHeader, token)

	for i := 0; i < b.N; i++ {
		apitest.PerformRequest(r, http.MethodPost, "/nonce/test", req)
	}
}

func TestNonce(t *testing.T) {
	suite.Run(t, new(TestNonceSuite))
}
