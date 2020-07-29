package file

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/cobinhood-backend/apps/exchange/cobx-types"
	"github.com/cobinhood/cobinhood-backend/apps/exchange/common/api/middlewares"
	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/common/aes"
	"github.com/cobinhood/cobinhood-backend/common/api/apitest"
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

type SuspicionInvestigationFileTestSuite struct {
	suite.Suite
	db          *gorm.DB
	router      *gin.Engine
	user        models.User
	accessToken string
	ip          string
	aesKey      aes.Key
}

func (s *SuspicionInvestigationFileTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	s.db = database.GetDB(database.Default)
	scopeauth.Initialize(cobxtypes.APIAdmin)

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
		middleware.AppContextMiddleware(cobxtypes.Test),
		middleware.APIValidator(s, cobxtypes.APIAdmin),
		middleware.ResponseHandler,
		middlewares.PanicLogger,
		middleware.ScopeAuth(cobxtypes.APIAdmin),
	)

	v1 := s.router.Group("v1")
	{
		v1.POST(
			"crm/suspicion/file",
			middleware.RequireAppContext(
				cobxtypes.APIAdmin, SuspicionInvestigationFileHandler(
					60, s.aesKey,
				)),
		)
		v1.GET(
			"crm/suspicion/file/:file_id",
			middleware.RequireAppContext(
				cobxtypes.APIAdmin, DownloadSuspicionInvestigationFile),
		)
	}
}

func (s *SuspicionInvestigationFileTestSuite) TearDownSuite() {
	database.DeleteAllData(&exchangedb.DBApp{})
	database.Finalize()
	cache.Finalize()
	scopeauth.Finalize()
}

func (s *SuspicionInvestigationFileTestSuite) SetupTest() {
	database.Reset(s.db, &exchangedb.DBApp{}, true)
	s.ip = "127.0.0.1"
	s.Require().Nil(database.Transaction(s.db, func(tx *gorm.DB) error {
		s.user = exchangetest.CreateTestingUser(tx, 1)[0]
		s.accessToken = exchangetest.GenerateAccessTokenWithExistUser(tx,
			s.user.Email, []types.Role{types.RoleCSMember}, s.ip)
		return nil
	}))
}

func (s *SuspicionInvestigationFileTestSuite) TestUpload() {
	fileBytes, err := base64.StdEncoding.DecodeString(exchangetest.TestingUserDataProfilePic)
	s.Require().Nil(err)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile(suspicionInvestigationFileDelegate{}.FileName(), "file")
	s.Require().Nil(err)
	_, err = fw.Write(fileBytes)
	s.Require().Nil(err)
	w.Close()

	response, err := apitest.SendRequest(s.router, apitest.HTTPParameter{
		Method:     http.MethodPost,
		URL:        "/v1/crm/suspicion/file",
		StatusCode: http.StatusOK,
		Body:       &b,
		Headers: map[string]string{
			"Content-Type":    w.FormDataContentType(),
			"Authorization":   s.accessToken,
			"X-Forwarded-For": s.ip,
		},
	})
	s.Require().Nil(err)
	var result testUploadResponse
	s.Require().Nil(json.Unmarshal(response, &result))
	s.Require().True(result.Success)

	// Check cached file.
	cacheKey := result.Result.CacheKey
	accessKey := s.user.ID.String()

	cachedB, err := cache.GetEphemeralFile(&accessKey, cacheKey)
	s.Require().Nil(err)

	decrypted, err := aes.CBCDecrypt(s.aesKey, cachedB)
	s.Require().Nil(err)

	s.Require().Equal(fileBytes, decrypted)
}

func (s *SuspicionInvestigationFileTestSuite) TestDownload() {
	fileBytes, err := base64.StdEncoding.DecodeString(
		exchangetest.TestingUserDataProfilePic)
	s.Require().Nil(err)

	// Prepare GCS file.
	gcsFile, err := models.NewGCSFileWithAESKeyEnsured(context.Background(),
		types.GCSFileTypeSuspicionInvestigation)
	s.Require().Nil(err)
	s.Require().Nil(s.db.Create(&gcsFile).Error)
	s.Require().Nil(gcsFile.WriteGCSObject(context.Background(), fileBytes))

	response, err := apitest.SendRequest(s.router, apitest.HTTPParameter{
		Method:     http.MethodGet,
		URL:        "/v1/crm/suspicion/file/" + gcsFile.ID.String(),
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Authorization":   s.accessToken,
			"X-Forwarded-For": s.ip,
		},
	})
	s.Require().Nil(err)
	s.Require().Equal(fileBytes, response)
}

func TestSuspicionInvestigationFile(t *testing.T) {
	suite.Run(t, new(SuspicionInvestigationFileTestSuite))
}
