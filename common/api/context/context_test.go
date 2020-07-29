package context

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/cobinhood-backend/cache"
	"github.com/cobinhood/cobinhood-backend/cache/keys"
	"github.com/cobinhood/cobinhood-backend/common/logging"
	"github.com/cobinhood/cobinhood-backend/database"
	"github.com/cobinhood/cobinhood-backend/database/exchangedb"
	"github.com/cobinhood/cobinhood-backend/infra/app"
	models "github.com/cobinhood/cobinhood-backend/models/exchange"
	"github.com/cobinhood/cobinhood-backend/models/exchange/exchangetest"
	"github.com/cobinhood/cobinhood-backend/types"
)

type AppContextTestSuite struct {
	suite.Suite
	logger logging.Logger
	redis  *cache.Redis
	db     *gorm.DB
}

func (s *AppContextTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	gin.SetMode(gin.TestMode)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	s.logger = logging.NewLoggerTag("appcontext-test")
	s.redis = cache.GetRedis()
	s.db = database.GetDB(database.Default)
}

func (s *AppContextTestSuite) TearDownSuite() {
	database.Finalize()
	cache.Finalize()
}

func (s *AppContextTestSuite) TestGetUserEmail() {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.RemoteAddr = "118.112.113.114"
	appCtx, _ := NewAppCtx(ctx, s.logger, s.db, s.redis)

	profilePic := ""
	accountType := types.Corporation
	timezone := 3
	local := types.LocaleZhTW
	prizeGer := models.User{
		ID:                uuid.NewV4(),
		Email:             "prize@cobinhood.com",
		Password:          "testpassword",
		ProfilePicture:    &profilePic,
		AccountType:       &accountType,
		Corporation:       &exchangetest.TestingUserDataCorp,
		KYCLevel:          3,
		Timezone:          &timezone,
		Locale:            &local,
		EmailNotification: false,
		IsFreezed:         false,
		TwoFactorAuthConfirmationBase: models.TwoFactorAuthConfirmationBase{
			TwoFactorAuthMethod: types.TwoFactorAuthNone,
		},
	}
	s.Require().Nil(s.db.Create(&prizeGer).Error)
	appCtx.UserID = &prizeGer.ID

	email := appCtx.GetUserEmail()
	s.Require().Equal("prize@cobinhood.com", email)
}

func (s *AppContextTestSuite) TestGetUserNationality() {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.RemoteAddr = "118.112.113.114"
	appCtx, err := NewAppCtx(ctx, s.logger, s.db, s.redis)

	user := exchangetest.CreateTestingUser(appCtx.DB, 1)[0]
	userID := user.ID

	nationality := appCtx.GetUserNationality()
	s.Require().Equal("", nationality)

	appCtx.UserID = &userID

	kycData, err := models.NewKYCDataWithAESKeyEnsured(ctx, userID)
	s.Require().NoError(err)

	utc, err := time.LoadLocation("UTC")
	s.Require().NoError(err)

	utc, err = time.LoadLocation("UTC")
	s.Require().NoError(err)

	birthday := time.Date(2000, 1, 4, 0, 0, 0, 0, utc)
	basicInfo := models.BasicInformation{
		FirstName:                "popo",
		LastName:                 "didi",
		NationalityCountry:       "usa",
		NationalityStateProvince: "tpe",
		ResidenceCountry:         "usa",
		ResidenceStateProvince:   "ny",
		Occupation:               "student",
		Birthday:                 types.JSONDate(birthday),
	}
	err = kycData.WriteFromDataBlock(ctx, &basicInfo)
	s.Require().NoError(err)
	proofOfIdentity := models.ProofOfIdentity{
		IdentityType:   types.Passport,
		IdentityNumber: "1234567",
	}
	err = kycData.WriteFromDataBlock(ctx, &proofOfIdentity)
	s.Require().NoError(err)
	phoneNumber := models.PhoneNumber{
		CountryCode: "886",
		PhoneNumber: "999999999",
	}
	err = kycData.WriteFromDataBlock(ctx, &phoneNumber)
	s.Require().NoError(err)
	fatca := models.FATCA{
		FATCATaxPayerKind:   types.FATCATaxPayerKindUSCitizen,
		FATCATaxPayerIDKind: types.FATCATaxPayerIDKindSSN,
		FATCATaxPayerID:     "xxxx",
	}
	err = kycData.WriteFromDataBlock(ctx, &fatca)
	s.Require().NoError(err)

	err = appCtx.DB.Create(kycData).Error
	s.Require().NoError(err)

	nationality = appCtx.GetUserNationality()
	s.Require().Equal("usa", nationality)

	nationality, err = appCtx.Cache.GetString(keys.GetUserNationalityCacheKey(*appCtx.UserID))
	s.Require().NoError(err)

	s.Require().Equal("usa", nationality)
}

func (s *AppContextTestSuite) TestSubRequestCtx() {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.RemoteAddr = "118.112.113.114"
	appCtx, err := NewAppCtx(ctx, s.logger, s.db, s.redis)
	s.Require().NoError(err)

	userID := uuid.NewV4()
	appCtx.UserID = &userID

	subAppCtx, err := appCtx.CreateSubRequestCtx()
	s.Require().NoError(err)

	s.Require().False(subAppCtx.IsAborted())

	s.Require().Equal(ctx.Request.RemoteAddr, subAppCtx.RequestIP)

	subAppCtx.SetRequestBody([]byte(`{"a":"123"}`))
	subAppCtx.SetParams(map[string]string{"a": "456"})

	req := &struct {
		A string `json:"a" binding:"required"`
	}{}

	s.Require().NoError(subAppCtx.BindJSON(req))
	s.Require().Equal("123", req.A)
	s.Require().Equal("456", subAppCtx.Param("a"))
	s.Require().Equal(userID, *subAppCtx.UserID)

	// Original appCtx has empty body, so it should failed.
	s.Require().Error(appCtx.BindJSON(req))

	// Test get correct appCtx.
	testAppCtx, err := GetAppContext(subAppCtx.ctx)
	s.Require().NoError(err)
	s.Require().Equal(subAppCtx, testAppCtx)
}

func (s *AppContextTestSuite) TestCheckEmployee() {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.RemoteAddr = "118.112.113.114"
	appCtx, _ := NewAppCtx(ctx, s.logger, s.db, s.redis)

	profilePic := ""
	accountType := types.Corporation
	timezone := 3
	local := types.LocaleZhTW
	prizeGer := models.User{
		ID:                uuid.NewV4(),
		Email:             "prize@cobinhood.com",
		Password:          "testpassword",
		ProfilePicture:    &profilePic,
		AccountType:       &accountType,
		Corporation:       &exchangetest.TestingUserDataCorp,
		KYCLevel:          3,
		Timezone:          &timezone,
		Locale:            &local,
		EmailNotification: false,
		IsFreezed:         false,
		TwoFactorAuthConfirmationBase: models.TwoFactorAuthConfirmationBase{
			TwoFactorAuthMethod: types.TwoFactorAuthNone,
		},
	}
	s.Require().Nil(s.db.Create(&prizeGer).Error)
	appCtx.UserID = &prizeGer.ID

	employee := appCtx.CheckEmployeeIdentity()
	s.Require().Equal(employee, true)
}

func TestAppContext(test *testing.T) {
	suite.Run(test, new(AppContextTestSuite))
}
