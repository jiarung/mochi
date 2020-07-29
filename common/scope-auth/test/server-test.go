package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/cache/helper"
	apicontext "github.com/cobinhood/cobinhood-backend/common/api/context"
	"github.com/cobinhood/cobinhood-backend/common/jwt"
	jwtFactory "github.com/cobinhood/cobinhood-backend/common/jwt"
	"github.com/cobinhood/cobinhood-backend/common/scope-auth"
	"github.com/cobinhood/cobinhood-backend/database"
	"github.com/cobinhood/cobinhood-backend/database/exchangedb"
	"github.com/cobinhood/cobinhood-backend/infra/app"
	"github.com/cobinhood/cobinhood-backend/models/exchange/exchangetest"
	"github.com/cobinhood/cobinhood-backend/types"
)

// ServerTestSuite tests if the server has registered the `middleware.ScopeAuth`.
type ServerTestSuite struct {
	suite.Suite
	ServiceName                cobxtypes.ServiceName
	MiddlewareRegisteredRouter *gin.Engine
	RegisterModule             func(engine *gin.Engine)
}

// SetupSuite initializes the database and cache.
func (s *ServerTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	cache.Initialize(config.Cache)
}

// TearDownSuite initializes the database and cache.
func (s *ServerTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
}

// SetupTest initializes the database and cache.
func (s *ServerTestSuite) SetupTest() {
	exchangetest.SetupSuiteX(&exchangedb.DBApp{})
	scopeauth.Initialize(s.ServiceName)
}

// TearDownTest finalizes the database and cache.
func (s *ServerTestSuite) TearDownTest() {
	exchangetest.TearDownSuiteX()
	scopeauth.Finalize()
}

// TestScopeMiddlewareExists tests if the middleware exists.
func (s *ServerTestSuite) TestScopeMiddlewareExists() {
	ip := "172.17.0.3"
	db := database.GetDB(database.Default)
	u, ac, da := exchangetest.CreateUserWithLogin(db)
	accessToken, err := jwtfactory.Build(jwtFactory.AccessTokenObj{
		UserID:                u,
		AccessTokenID:         ac,
		DeviceAuthorizationID: da,
		Platform:              types.DeviceIOS,
		LoginCount:            -1,
	}).Gen(s.ServiceName, 1800)
	s.Require().Nil(err)
	payload := helper.AccessTokenPayload{
		IP:    ip,
		Roles: types.GetAllRoles(),
	}
	s.Require().Nil(payload.Set(ac.String()))

	scopeMiddlewareChecker := func(ctx *gin.Context) {
		appCtx, err := apicontext.GetAppContext(ctx)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), appCtx.RequiredScopes)
		ctx.AbortWithStatus(799)
	}

	router := s.MiddlewareRegisteredRouter
	router.Use(scopeMiddlewareChecker)
	s.RegisterModule(router)

	// copied from nonce middleware
	nonceIgnoreMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodOptions: true,
	}

	for _, r := range router.Routes() {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(r.Method, r.Path, nil)
		req.Header.Set("Authorization", accessToken)
		if _, ok := nonceIgnoreMethods[r.Method]; !ok {
			req.Header.Set("nonce", fmt.Sprint(time.Now().UnixNano()))
		}
		req.RemoteAddr = ip
		s.MiddlewareRegisteredRouter.ServeHTTP(recorder, req)
		require.Equal(
			s.T(),
			799, recorder.Code,
			"Path: %v,%v != 799", r.Path, recorder.Code,
		)
	}

}

// TestScopesOfEveryEndpointsExist tests if the scopes of all endpoints are
// specified in `scope-map.go`.
func (s *ServerTestSuite) TestScopesOfEveryEndpointsExist() {
	for _, r := range s.MiddlewareRegisteredRouter.Routes() {
		s.T().Log(r.Method, r.Path)
		scopes, err := scopeauth.GetScopes(s.ServiceName, r.Method, r.Path)
		require.Nil(s.T(), err)
		require.NotZero(s.T(), len(scopes))
	}
}
