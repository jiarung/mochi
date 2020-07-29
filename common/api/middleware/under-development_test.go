package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/common/config/misc"
	"github.com/cobinhood/mochi/common/utils"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	"github.com/cobinhood/mochi/infra/app"
)

type UnderDevelopmentSuite struct {
	suite.Suite
	engine    *gin.Engine
	originEnv string
}

func (s *UnderDevelopmentSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	s.originEnv = misc.ServerEnvironment()
}

func (s *UnderDevelopmentSuite) TearDownSuite() {
	misc.SetServerEnvironment(s.originEnv)
	database.Finalize()
	cache.Finalize()
}

func (s *UnderDevelopmentSuite) SetupTest() {
	fakeResp := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "")
	}

	engine := gin.Default()
	engine.Use(logger.NewLoggerMiddleware)
	engine.Use(
		AppContextMiddleware(cobxtypes.Test),
		ResponseHandler,
	)

	v1 := engine.Group("v1")
	fakeGroup := v1.Group("fake")
	fakeGroup.GET("empty", UnderDevelopment, fakeResp)

	s.engine = engine
}

func (s *UnderDevelopmentSuite) setEnvAndTestHTTPCode(env string, code int) {
	// Set mode.
	// Const key is in utils.Environment().
	misc.SetServerEnvironment(env)

	req, err := http.NewRequest(http.MethodGet, "/v1/fake/empty", nil)
	s.Require().NoError(err)

	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)
	s.Require().Equal(code, w.Code)
}

func (s *UnderDevelopmentSuite) TestInLocalDevelopmentEnviron() {
	// Set to local dev mode.
	s.setEnvAndTestHTTPCode(utils.EnvLocalDevelopmentTag, http.StatusOK)
}

func (s *UnderDevelopmentSuite) TestInDevelopmentEnviron() {
	// Set to dev mode.
	s.setEnvAndTestHTTPCode(utils.EnvDevelopmentTag, http.StatusOK)
}

func (s *UnderDevelopmentSuite) TestInStagingEnviron() {
	// Set to staging mode.
	s.setEnvAndTestHTTPCode(utils.EnvStagingTag, http.StatusOK)
}

func (s *UnderDevelopmentSuite) TestInProductionEnviron() {
	// Set to prod mode.
	s.setEnvAndTestHTTPCode(utils.EnvProductionTag, http.StatusNotFound)
}

func TestUnderDevelopment(t *testing.T) {
	suite.Run(t, new(UnderDevelopmentSuite))
}
