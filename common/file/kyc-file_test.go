package file

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/apps/exchange/common/api/middlewares"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/common/aes"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/common/scope-auth"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	"github.com/cobinhood/mochi/infra/app"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/models/exchange/exchangetest"
	"github.com/cobinhood/mochi/types"
)

type KYCFileTestSuite struct {
	suite.Suite
	db          *gorm.DB
	router      *gin.Engine
	exUser      models.User
	exUserToken string
	exUserIP    string
}

func (s *KYCFileTestSuite) SetupSuite() {
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

	// Setup HTTP Router.
	// Read request body once
	s.router = gin.Default()
	s.router.Use(
		logger.NewLoggerMiddleware,
		middlewares.ContextLogger(),
		middleware.AppContextMiddleware(cobxtypes.Test),
		middleware.ResponseHandler,
		middlewares.PanicLogger,
		middleware.ScopeAuth(cobxtypes.APICobx),
	)

	// KYC Module
	v1 := s.router.Group("v1")
	{
		kycGrp := v1.Group("kyc")
		{
			kycGrp.POST(
				"ephemeral_picture",
				middleware.RequireAppContext(cobxtypes.APICobx,
					UploadKYCFileHandler(60*60, nil)),
			)
			kycGrp.GET(
				"pictures/:picture_id",
				middleware.RequireAppContext(cobxtypes.APICobx, DownloadKYCInfo),
			)
		}
	}
}

func (s *KYCFileTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
	scopeauth.Finalize()
}

func (s *KYCFileTestSuite) SetupTest() {
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

func (s *KYCFileTestSuite) TestUpload() {
	fileBytes, err := base64.StdEncoding.DecodeString(exchangetest.TestingUserDataProfilePic)
	s.Require().Nil(err)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("kyc_info", "file")
	s.Require().Nil(err)
	_, err = fw.Write(fileBytes)
	s.Require().Nil(err)
	w.Close()

	request := httptest.NewRequest(http.MethodPost, "/v1/kyc/ephemeral_picture", &b)
	request.Header.Set("Content-Type", w.FormDataContentType())
	request.Header.Set("Authorization", s.exUserToken)
	request.Header.Set("X-Forwarded-For", s.exUserIP)

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)
	var res testUploadResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	s.Require().Nil(err)

	// check cached file
	var kycData models.KYCData
	err = s.db.Model(&models.KYCData{}).
		Where("user_id = ?", s.exUser.ID).
		First(&kycData).Error
	s.Require().Nil(err)

	aesKey, err := kycData.GetAESKey(context.Background())
	s.Require().Nil(err)

	cacheKey := res.Result.CacheKey
	accessKey := s.exUser.ID.String()

	cachedB, err := cache.GetEphemeralFile(&accessKey, cacheKey)
	s.Require().Nil(err)

	decrypted, err := aes.CBCDecrypt(aesKey, cachedB)
	s.Require().Nil(err)

	s.Require().Equal(fileBytes, decrypted)
}

func (s *KYCFileTestSuite) TestDownload() {
	fileBytes, err := base64.StdEncoding.DecodeString(
		exchangetest.TestingUserDataProfilePic)
	s.Require().Nil(err)

	// prepare gcs file
	gcsFile, err := models.NewGCSFileWithAESKeyEnsured(context.Background(),
		types.GCSFileTypeKYCData)
	s.Require().Nil(err)
	err = s.db.Create(&gcsFile).Error
	s.Require().Nil(err)
	err = gcsFile.WriteGCSObject(context.Background(), fileBytes)
	s.Require().Nil(err)

	kycData := models.KYCData{
		UserID:                    s.exUser.ID,
		ProofOfIdentityBackFileID: &gcsFile.ID,
	}
	err = s.db.Create(&kycData).Error
	s.Require().Nil(err)

	request := httptest.NewRequest(http.MethodGet,
		"/v1/kyc/pictures/"+gcsFile.ID.String(), nil)
	request.Header.Set("Authorization", s.exUserToken)
	request.Header.Set("X-Forwarded-For", s.exUserIP)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	s.Require().Equal(http.StatusOK, recorder.Code)
	s.Require().Equal(fileBytes, recorder.Body.Bytes())
}

func TestKYCFile(t *testing.T) {
	suite.Run(t, new(KYCFileTestSuite))
}
