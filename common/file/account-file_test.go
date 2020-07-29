package file

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/apps/exchange/common/api/middlewares"
	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/common/aes"
	"github.com/cobinhood/cobinhood-backend/common/api/middleware"
	"github.com/cobinhood/cobinhood-backend/common/scope-auth"
	"github.com/cobinhood/cobinhood-backend/database"
	"github.com/cobinhood/cobinhood-backend/database/exchangedb"
	"github.com/cobinhood/cobinhood-backend/infra/api/middleware/logger"
	"github.com/cobinhood/cobinhood-backend/infra/app"
	models "github.com/cobinhood/cobinhood-backend/models/exchange"
	"github.com/cobinhood/cobinhood-backend/models/exchange/exchangetest"
	"github.com/cobinhood/cobinhood-backend/types"
)

type AccountFileTestSuite struct {
	suite.Suite
	db          *gorm.DB
	router      *gin.Engine
	exUser      models.User
	exUserToken string
	exUserIP    string
	aesKey      aes.Key
}

func (s *AccountFileTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	scopeauth.Initialize(cobxtypes.APICobx)

	s.db = database.GetDB(database.Default)

	aesKey, err := hex.DecodeString(
		"4e096440cccae9a842d13bfea427631e3b63e964f4e6b090a398fba687ec21a2")
	s.Require().Nil(err)
	s.aesKey = aesKey

	// Setup HTTP Router.
	// Read request body once
	s.router = gin.Default()
	s.router.Use(
		logger.NewLoggerMiddleware,
		middlewares.ContextLogger(),
		middleware.APIValidator(s, cobxtypes.APICobx),
		middleware.AppContextMiddleware(cobxtypes.Test),
		middleware.ResponseHandler,
		middlewares.PanicLogger,
		middleware.ScopeAuth(cobxtypes.APICobx),
	)

	// KYC Module
	v1 := s.router.Group("v1")
	{
		kycGrp := v1.Group("account")
		{
			kycGrp.POST(
				"ephemeral_picture",
				middleware.RequireAppContext(cobxtypes.APICobx,
					UploadAccountFileHandler(60, s.aesKey)),
			)
		}
	}
}

func (s *AccountFileTestSuite) TearDownSuite() {
	database.DeleteAllData(&exchangedb.DBApp{})
	database.Finalize()
	cache.Finalize()
	scopeauth.Finalize()
}

func (s *AccountFileTestSuite) SetupTest() {
	database.DeleteAllData(&exchangedb.DBApp{})
	tx := s.db.Begin()
	// Create test user
	user := exchangetest.CreateTestingUser(tx, 1)
	s.exUser = user[0]
	s.exUserIP = "127.0.0.1"
	roles := []types.Role{types.RoleExchangeDefaultUser}
	accessToken := exchangetest.GenerateAccessTokenWithExistUser(tx,
		s.exUser.Email, roles, s.exUserIP)
	s.exUserToken = accessToken

	tx.Commit()
}

func (s *AccountFileTestSuite) TestUpload() {
	fileBytes, err := base64.StdEncoding.DecodeString(exchangetest.TestingUserDataProfilePic)
	s.Require().Nil(err)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("account_info", "file")
	s.Require().Nil(err)
	_, err = fw.Write(fileBytes)
	s.Require().Nil(err)
	w.Close()

	request := httptest.NewRequest(http.MethodPost, "/v1/account/ephemeral_picture", &b)
	request.Header.Set("Content-Type", w.FormDataContentType())
	request.Header.Set("Authorization", s.exUserToken)
	request.Header.Set("X-Forwarded-For", s.exUserIP)

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)
	var res testUploadResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	s.Require().Nil(err)

	cacheKey := res.Result.CacheKey
	accessKey := s.exUser.ID.String()

	cachedB, err := cache.GetEphemeralFile(&accessKey, cacheKey)
	s.Require().Nil(err)

	decrypted, err := aes.CBCDecrypt(s.aesKey, cachedB)
	s.Require().Nil(err)

	s.Require().Equal(fileBytes, decrypted)
}

func TestAccountFile(t *testing.T) {
	suite.Run(t, new(AccountFileTestSuite))
}
