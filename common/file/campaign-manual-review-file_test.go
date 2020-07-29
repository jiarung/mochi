package file

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cobxtypes "github.com/cobinhood/mochi/apps/exchange/cobx-types"
	"github.com/cobinhood/mochi/cache"
	apicontext "github.com/cobinhood/mochi/common/api/context"
	"github.com/cobinhood/mochi/common/logging"
	scopeauth "github.com/cobinhood/mochi/common/scope-auth"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/infra/app"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/models/exchange/exchangetest"
	"github.com/cobinhood/mochi/types"
)

type CampaignManualReviewTestSuite struct {
	suite.Suite
	db    *gorm.DB
	cache *cache.Redis
	users []models.User
}

func (s *CampaignManualReviewTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	s.db = database.GetDB(database.Default)
	s.cache = cache.GetRedis()
	scopeauth.Initialize(cobxtypes.APIAdmin)

	s.users = exchangetest.CreateTestingUser(s.db, 1)
	require.NotEmpty(s.T(), s.users)
}

func (s *CampaignManualReviewTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
	scopeauth.Finalize()
}

func (s *CampaignManualReviewTestSuite) TestDownload() {
	fileBytes, err := base64.StdEncoding.DecodeString(
		exchangetest.TestingUserDataProfilePic)
	s.Require().NoError(err)

	// Prepare GCS file.
	gcsFile, err := models.NewGCSFileWithAESKeyEnsured(context.Background(),
		types.GCSFileTypeSuspicionInvestigation)
	s.Require().NoError(err)
	s.Require().NoError(s.db.Create(&gcsFile).Error)
	s.Require().NoError(gcsFile.WriteGCSObject(context.Background(), fileBytes))

	submit := &models.ManualReviewSubmit{
		UserID: s.users[0].ID,
		Event:  "test",
		Data: models.ManualReviewSubmitDataMap{
			"file": models.ManualReviewSubmitData{
				IsFile: true,
				Text:   gcsFile.ID.String(),
			},
		},
	}
	s.Require().NoError(s.db.Create(submit).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest(http.MethodGet, "", nil)

	appCtx, err := apicontext.NewAppCtx(
		ctx,
		logging.NewLogger(),
		s.db,
		s.cache)
	s.Require().NoError(err)
	appCtx.UserID = &s.users[0].ID

	bytes, errStr := DownloadCampaignManualReviewFile(
		appCtx, "test", submit.ID, gcsFile.ID)
	s.Require().Empty(errStr)
	s.Require().Equal(fileBytes, bytes)

	// Test failed.
	_, errStr = DownloadCampaignManualReviewFile(
		appCtx, "test1", submit.ID, gcsFile.ID)
	s.Require().NotEmpty(errStr)

	_, errStr = DownloadCampaignManualReviewFile(
		appCtx, "test", uuid.NewV4(), gcsFile.ID)
	s.Require().NotEmpty(errStr)

	_, errStr = DownloadCampaignManualReviewFile(
		appCtx, "test", submit.ID, uuid.NewV4())
	s.Require().NotEmpty(errStr)
}

func TestCampaignManualReviewFile(t *testing.T) {
	suite.Run(t, new(CampaignManualReviewTestSuite))
}
