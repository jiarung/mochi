package middleware

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/common/api/apitest"
	apierrors "github.com/cobinhood/cobinhood-backend/common/api/errors"
	apiutils "github.com/cobinhood/cobinhood-backend/common/api/utils"
	"github.com/cobinhood/cobinhood-backend/common/logging"
	"github.com/cobinhood/cobinhood-backend/database"
	"github.com/cobinhood/cobinhood-backend/database/exchangedb"
	"github.com/cobinhood/cobinhood-backend/infra/api/middleware/logger"
	"github.com/cobinhood/cobinhood-backend/infra/app"
)

type responseHandlerSuite struct {
	suite.Suite
	e *gin.Engine
}

func (s *responseHandlerSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	cache.Initialize(config.Cache)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)

	e := gin.New()
	e.Use(logger.NewLoggerMiddleware)
	e.Use(AppContextMiddleware(cobxtypes.Test))
	e.Use(ResponseHandler)

	hackTag := func(tag string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(logging.LabelTag, tag)
		}
	}

	e.GET("/api/v1/error",
		hackTag("error_tag"),
		func(c *gin.Context) {
			apiutils.SetError(c, apierrors.InvalidPayLoad)
			c.Abort()
		})
	e.GET("/api/v1/unexpect",
		hackTag("unexpect_tag"),
		func(c *gin.Context) {
			c.Abort()
		})
	s.e = e
}

func (s *responseHandlerSuite) TearDownSuite() {
	s.Require().Nil(cache.GetRedis().FlushAll())
	cache.Finalize()
	database.Finalize()
}

func (s *responseHandlerSuite) TestErrorHandler() {
	w := apitest.PerformRequest(s.e, http.MethodGet, "/api/v1/error", nil)
	s.Require().Equal(apierrors.HTTPStatus(apierrors.InvalidPayLoad), w.Code)

	response := apiutils.FailureWithTagObj{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().Nil(err)
	require.False(s.T(), response.Success)
	// check field exists. context.middleware can't be imported and tested due
	// to circular import
	s.Require().Equal("error_tag", response.Tag)

	w = apitest.PerformRequest(s.e, http.MethodGet, "/api/v1/unexpect", nil)
	require.Equal(
		s.T(), apierrors.HTTPStatus(apierrors.UnexpectedError), w.Code)

	err = json.NewDecoder(w.Body).Decode(&response)
	s.Require().Nil(err)
	require.False(s.T(), response.Success)
	s.Require().Equal("unexpect_tag", response.Tag)
}

func TestResponseHandler(t *testing.T) {
	suite.Run(t, new(responseHandlerSuite))
}
