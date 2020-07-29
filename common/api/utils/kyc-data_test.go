package utils

import (
	"context"
	"testing"
	"time"

	"github.com/jiarung/gorm"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/cache/keys"
	"github.com/jiarung/mochi/database"
	"github.com/jiarung/mochi/infra/app"
	"github.com/jiarung/mochi/models/exchange"
	"github.com/jiarung/mochi/models/exchange/exchangetest"
	"github.com/jiarung/mochi/types"
)

type KYCDataTestSuite struct {
	suite.Suite
	ctx   context.Context
	db    *gorm.DB
	redis *cache.Redis
}

func (s *KYCDataTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	cache.Initialize(config.Cache)

	s.ctx = context.Background()
	s.db = database.GetDB(database.Default)
	s.redis = cache.GetRedis()
}

func (s *KYCDataTestSuite) TearDownSuite() {
	cache.Finalize()
	database.Finalize()
}

func (s *KYCDataTestSuite) populateKYCData(
	kycData *exchange.KYCData, blocks ...exchange.KYCDataBlock) {

	for _, block := range blocks {
		err := kycData.WriteFromDataBlock(s.ctx, block)
		require.NoError(s.T(), err)
	}
}

func (s *KYCDataTestSuite) TestGetAndCacheKYCData() {
	user := exchangetest.CreateTestingUser(s.db, 1)[0]
	birthday := time.Date(1992, 2, 29, 0, 0, 0, 0, time.UTC)

	kycData, err := exchange.NewKYCDataWithAESKeyEnsured(s.ctx, user.ID)
	require.NoError(s.T(), err)

	s.populateKYCData(kycData,
		&exchange.BasicInformation{
			FirstName:          "信義",
			LastName:           "路",
			NationalityCountry: "TW",
			ResidenceCountry:   "TW",
			Occupation:         "農場",
			Birthday:           types.JSONDate(birthday)},
		&exchange.ProofOfIdentity{
			IdentityType:   types.Identification,
			IdentityNumber: "Y164445430"},
		&exchange.PhoneNumber{
			CountryCode: "886",
			PhoneNumber: "924264101"},
		&exchange.FATCA{
			FATCATaxPayerKind: types.FATCATaxPayerKindNone})
	result := s.db.Create(kycData)
	require.NoError(s.T(), result.Error)

	now1 := time.Date(2018, 2, 28, 9, 48, 7, 0, time.UTC)

	nationalityCountry1, log, err :=
		GetKYCDataNationalityCountry(s.ctx, s.db, s.redis, user.ID, now1)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "TW", nationalityCountry1)

	residenceCountry1, log, err :=
		GetKYCDataResidenceCountry(s.ctx, s.db, s.redis, user.ID, now1)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "TW", residenceCountry1)

	age1, ageValid1, log, err :=
		GetKYCDataAge(s.ctx, s.db, s.redis, user.ID, now1)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.True(s.T(), ageValid1)
	require.Equal(s.T(), uint8(25), age1)

	s.populateKYCData(kycData,
		&exchange.BasicInformation{
			FirstName:          "莊敬",
			LastName:           "路",
			NationalityCountry: "ID",
			ResidenceCountry:   "TH",
			Occupation:         "演員",
			Birthday:           types.JSONDate(birthday)})
	result = s.db.Save(kycData)
	require.NoError(s.T(), result.Error)

	now2 := time.Date(2018, 3, 1, 5, 4, 3, 0, time.UTC)

	nationalityCountry2, log, err :=
		GetKYCDataNationalityCountry(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "TW", nationalityCountry2)

	residenceCountry2, log, err :=
		GetKYCDataResidenceCountry(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "TW", residenceCountry2)

	age2, ageValid2, log, err :=
		GetKYCDataAge(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.True(s.T(), ageValid2)
	require.Equal(s.T(), uint8(25), age2)

	log, err = CacheKYCData(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)

	nationalityCountry3, log, err :=
		GetKYCDataNationalityCountry(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "ID", nationalityCountry3)

	residenceCountry3, log, err :=
		GetKYCDataResidenceCountry(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Equal(s.T(), "TH", residenceCountry3)

	age3, ageValid3, log, err :=
		GetKYCDataAge(s.ctx, s.db, s.redis, user.ID, now2)
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.True(s.T(), ageValid3)
	require.Equal(s.T(), uint8(26), age3)
}

func (s *KYCDataTestSuite) TestEmptyKYCData() {
	var err error
	var log string

	user := exchangetest.CreateTestingUser(s.db, 1)[0]
	birthday := time.Date(1959, 8, 7, 8, 7, 8, 7, time.UTC)

	_, err = s.redis.GetString(keys.GetUserNationalityCacheKey(user.ID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))
	_, err = s.redis.GetString(keys.GetUserResidenceCacheKey(user.ID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))
	_, err = s.redis.GetString(keys.GetUserAgeKey(user.ID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))

	log, err = CacheKYCData(s.ctx, s.db, s.redis, user.ID, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)

	nationalityCountry2, err :=
		s.redis.GetString(keys.GetUserNationalityCacheKey(user.ID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), nationalityCountry2)
	residenceCountry2, err :=
		s.redis.GetString(keys.GetUserResidenceCacheKey(user.ID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), residenceCountry2)
	age2, err :=
		s.redis.GetString(keys.GetUserAgeKey(user.ID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), age2)

	kycData, err := exchange.NewKYCDataWithAESKeyEnsured(s.ctx, user.ID)
	require.NoError(s.T(), err)

	s.populateKYCData(kycData,
		&exchange.BasicInformation{
			FirstName:          "基隆",
			LastName:           "路",
			NationalityCountry: "MT",
			ResidenceCountry:   "FR",
			Occupation:         "蓋房子",
			Birthday:           types.JSONDate(birthday)},
		&exchange.ProofOfIdentity{
			IdentityType:   types.Passport,
			IdentityNumber: "11876196141847223936"},
		&exchange.PhoneNumber{
			CountryCode: "4175",
			PhoneNumber: "20245704729385"},
		&exchange.FATCA{
			FATCATaxPayerKind: types.FATCATaxPayerKindNone})
	result := s.db.Create(kycData)
	require.NoError(s.T(), result.Error)

	basicInformation := exchange.BasicInformation{}
	err = kycData.ReadToDataBlock(s.ctx, &basicInformation)
	require.NoError(s.T(), err)
	err = basicInformation.ValidateFormat()
	require.NoError(s.T(), err)

	nationalityCountry3, log, err :=
		GetKYCDataNationalityCountry(s.ctx, s.db, s.redis, user.ID, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Empty(s.T(), nationalityCountry3)

	residenceCountry3, log, err :=
		GetKYCDataResidenceCountry(s.ctx, s.db, s.redis, user.ID, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.Empty(s.T(), residenceCountry3)

	_, ageValid3, log, err :=
		GetKYCDataAge(s.ctx, s.db, s.redis, user.ID, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)
	require.False(s.T(), ageValid3)

	noSuchID := uuid.NewV4()

	_, err = s.redis.GetString(keys.GetUserNationalityCacheKey(noSuchID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))
	_, err = s.redis.GetString(keys.GetUserResidenceCacheKey(noSuchID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))
	_, err = s.redis.GetString(keys.GetUserAgeKey(noSuchID))
	require.Equal(s.T(), cache.ErrNilKey, cache.ParseCacheErrorCode(err))

	log, err = CacheKYCData(s.ctx, s.db, s.redis, noSuchID, time.Now())
	require.NoError(s.T(), err)
	require.Empty(s.T(), log)

	nationalityCountry5, err :=
		s.redis.GetString(keys.GetUserNationalityCacheKey(noSuchID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), nationalityCountry5)
	residenceCountry5, err :=
		s.redis.GetString(keys.GetUserResidenceCacheKey(noSuchID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), residenceCountry5)
	age5, err :=
		s.redis.GetString(keys.GetUserAgeKey(noSuchID))
	require.NoError(s.T(), err)
	require.Empty(s.T(), age5)
}

func TestKYCData(t *testing.T) {
	suite.Run(t, &KYCDataTestSuite{})
}
