package limiters

import (
	"errors"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/mochi/cache"
	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/app"
)

type simpleLimiterSelectorTestSuite struct {
	suite.Suite
}

func (s *simpleLimiterSelectorTestSuite) SetupSuite() {
	var config struct {
		Database database.Config

		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)

	cache.Initialize(config.Cache)
}

func (s *simpleLimiterSelectorTestSuite) TearDownSuite() {
	cache.GetRedis().FlushAll()
	cache.Finalize()

	database.Finalize()
}

func (s *simpleLimiterSelectorTestSuite) TestSimpleLimiterSelector() {
	selector := NewSimpleLimiterSelector(2, 1)
	u1 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b1")
	s.Require().NotNil(u1)
	u2 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b2")
	s.Require().NotNil(u2)
	u3 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b3")
	s.Require().NotNil(u3)

	limiter := NewLimiter(2, 1)
	// just test the selector will return exactly same limiter everytime
	testCases := []struct {
		u       *uuid.UUID
		limiter Limiter
	}{
		{&u1, limiter},
		{&u2, limiter},
		{&u3, limiter},
	}

	for _, tc := range testCases {
		l := selector.SelectLimiter(database.GetDB(database.Default), tc.u)
		s.Require().Equal(tc.limiter, l)
	}
}

func (s *simpleLimiterSelectorTestSuite) TestVIPLimiterSelector() {
	u1 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b1")
	s.Require().NotNil(u1)
	u2 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b2")
	s.Require().NotNil(u2)
	u3 := uuid.FromStringOrNil("17ea83af-53a5-4d70-ac30-113dee97c7b3")
	s.Require().NotNil(u3)

	l1 := NewLimiter(20, 2)
	l2 := NewLimiter(30, 3)

	mockFunc := func(db *gorm.DB, userID uuid.UUID) (Limiter, error) {
		if userID.String() == u1.String() {
			return l1, nil
		} else if userID.String() == u2.String() {
			return l2, nil
		}
		return nil, errors.New("fail")
	}

	defaultLimit := int64(2)
	defaultSeconds := int(1)
	selector := NewVIPLimiterSelector(defaultLimit, defaultSeconds, mockFunc)

	testCases := []struct {
		u       *uuid.UUID
		limiter Limiter
	}{
		{&u1, NewLimiter(20, 2)},
		{&u2, NewLimiter(30, 3)},
		{&u3, NewLimiter(defaultLimit, defaultSeconds)},
	}

	for _, tc := range testCases {
		l := selector.SelectLimiter(database.GetDB(database.Default), tc.u)
		s.Require().Equal(tc.limiter, l)
	}
}

func TestLimiterSelector(t *testing.T) {
	suite.Run(t, &simpleLimiterSelectorTestSuite{})
}
