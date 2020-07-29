package file

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/common/aes"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	"github.com/cobinhood/mochi/common/api/middleware"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/api/middleware/logger"
	"github.com/cobinhood/mochi/infra/app"
	"github.com/cobinhood/mochi/models/exchange/exchangetest"
)

type testUploadResponse struct {
	Success bool             `json:"success"`
	Result  testUploadResult `json:"result"`
}

type testUploadResult struct {
	CacheKey string `json:"cache_key"`
}

type EphemeralFileTestSuite struct {
	suite.Suite
	router    *gin.Engine
	aesKey    aes.Key
	accessKey *string
}

// Upload delegate
func (s *EphemeralFileTestSuite) SizeLimit() int64 {
	return MB
}
func (s *EphemeralFileTestSuite) FileName() string {
	return "file"
}
func (s *EphemeralFileTestSuite) ValidContentTypes() []string {
	return []string{PNGType}
}
func (s *EphemeralFileTestSuite) ExpireSec() int {
	return 10
}
func (s *EphemeralFileTestSuite) ShouldEncrypt() bool {
	return true
}
func (s *EphemeralFileTestSuite) GetEncryptionAESKey(appCtx *apicontext.AppContext) (aes.Key, error) {
	return s.aesKey, nil
}
func (s *EphemeralFileTestSuite) ShouldSetAccessKey() bool {
	return true
}
func (s *EphemeralFileTestSuite) GetAccessKey(appCtx *apicontext.AppContext) (
	*string, error) {
	return s.accessKey, nil
}

// Download delgate
func (s *EphemeralFileTestSuite) GetPathParameter() string {
	return "cache_key"
}
func (s *EphemeralFileTestSuite) ShouldDecrypt() bool {
	return true
}
func (s *EphemeralFileTestSuite) GetDecryptionAESKey(appCtx *apicontext.AppContext) (
	aes.Key, error) {
	return s.aesKey, nil
}

func (s *EphemeralFileTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)

	aesKey, err := aes.GenerateAESKey(aes.AES256KeySize)
	require.Nil(s.T(), err)
	s.aesKey = aesKey
	accessKey := "access key"
	s.accessKey = &accessKey

	router := gin.New()
	router.Use(logger.NewLoggerMiddleware)
	router.Use(middleware.ResponseHandler)
	router.Use(middleware.AppContextMiddleware(cobxtypes.Test))
	router.POST(
		"upload",
		middleware.RequireAppContext(cobxtypes.Test, UploadEphemeralFileHandler(s)),
	)
	router.GET(
		"download/:cache_key",
		middleware.RequireAppContext(cobxtypes.Test, DownloadEphemeralFileHandler(s)),
	)

	s.router = router
}

func (s *EphemeralFileTestSuite) TearDownSuite() {
	database.DeleteAllData(&exchangedb.DBApp{})
	database.Finalize()
	cache.Finalize()
}

func (s *EphemeralFileTestSuite) TestUploadDownloadEphemeralFile() {
	var fileBytes []byte
	var buf bytes.Buffer
	var writer *multipart.Writer
	var fileWriter io.Writer
	var err error

	// should fail with wrong type
	fileBytes = []byte("demo file")
	buf = bytes.Buffer{}
	writer = multipart.NewWriter(&buf)
	fileWriter, err = writer.CreateFormFile("file", "file")
	s.Require().Nil(err)
	_, err = fileWriter.Write(fileBytes)
	s.Require().Nil(err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	s.Require().Equal(http.StatusBadRequest, recorder.Code)

	// should succeed with popo's photo.
	fileBytes, err = base64.StdEncoding.DecodeString(exchangetest.TestingUserDataProfilePic)
	buf = bytes.Buffer{}
	writer = multipart.NewWriter(&buf)
	fileWriter, err = writer.CreateFormFile("file", "file")
	s.Require().Nil(err)
	_, err = fileWriter.Write(fileBytes)
	s.Require().Nil(err)
	writer.Close()

	req = httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	require.Equal(s.T(), http.StatusOK, recorder.Code)
	var res testUploadResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	s.Require().Nil(err)

	cacheKey := res.Result.CacheKey
	fileFromRedis, err := cache.GetEphemeralFile(s.accessKey, cacheKey)
	s.Require().Nil(err)

	decrypted, err := aes.CBCDecrypt(s.aesKey, fileFromRedis)
	s.Require().Nil(err)
	s.Require().Equal(fileBytes, decrypted)

	req = httptest.NewRequest(http.MethodGet, "/download/"+cacheKey, nil)
	recorder = httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	require.Equal(s.T(), http.StatusOK, recorder.Code)
	require.Equal(s.T(), fileBytes, recorder.Body.Bytes())
}

func TestEphemeralFile(t *testing.T) {
	suite.Run(t, new(EphemeralFileTestSuite))
}
