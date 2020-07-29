package customquery

import (
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/app"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/cobinhood/mochi/models/exchange/exchangetest"
)

type SQLOptionsTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *SQLOptionsTestSuite) SetupSuite() {
	var config struct {
		Database database.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	database.DeleteAllData(&exchangedb.DBApp{})
	suite.db = database.GetDB(database.Default)
}

func (suite *SQLOptionsTestSuite) TearDownSuite() {
	database.DeleteAllData(&exchangedb.DBApp{})
	database.Finalize()
}

func (suite *SQLOptionsTestSuite) SetupTest() {
	database.DeleteAllData(&exchangedb.DBApp{})
}

func (suite *SQLOptionsTestSuite) TestFilterSQLInjection() {
	hacker := exchangetest.CreateTestingUser(suite.db, 1)[0]
	err := suite.db.Model(hacker).Update("email", "hacker@cobinhood.com").Error
	suite.Require().Nil(err)

	qqUser := exchangetest.CreateTestingUser(suite.db, 1)[0]
	err = suite.db.Model(qqUser).Update("email", "qq@cobinhood.com").Error
	suite.Require().Nil(err)

	filter := map[string]interface{}{
		"like": map[string]interface{}{
			"column": "email\"=email) and password=password limit (select ascii('A'));" +
				"update \"user\" set email = 'deadQQ@cobinhood.com' where email =" +
				" 'qq@cobinhood.com';--",
			"value": "%Postg%",
		},
	}
	f := Filter(filter)
	limit := 1000
	opt := SQLOptions{
		Filter: &f,
		Order:  nil,
		Limit:  &limit,
		Page:   nil,
	}

	db := suite.db.Model(&models.User{})
	_, _, _, _, _, err = opt.Apply(db.Debug(), 50, 1, 0)
	suite.Require().NotNil(err)

}

func (suite *SQLOptionsTestSuite) TestOrderSQLInjection() {
	hacker := exchangetest.CreateTestingUser(suite.db, 1)[0]
	err := suite.db.Model(hacker).Update("email", "hacker@cobinhood.com").Error
	suite.Require().Nil(err)

	qqUser := exchangetest.CreateTestingUser(suite.db, 1)[0]
	err = suite.db.Model(qqUser).Update("email", "qq@cobinhood.com").Error
	suite.Require().Nil(err)

	o := order{
		Column: "email\"=email) and password=password limit (select ascii('A'));" +
			"update \"user\" set email = 'deadQQ@cobinhood.com' where email =" +
			" 'qq@cobinhood.com';--",
		Keyword: "asc",
	}
	order := Order([]order{o})
	limit := 1000
	opt := SQLOptions{
		Filter: nil,
		Order:  &order,
		Limit:  &limit,
		Page:   nil,
	}

	db := suite.db.Model(&models.User{})
	_, _, _, _, _, err = opt.Apply(db.Debug(), 50, 1, 0)
	suite.Require().NotNil(err)

}

func (suite *SQLOptionsTestSuite) TestValidColumnMap() {
	opt := SQLOptions{}

	columnMap, err := opt.validColumnMap(&models.Balance{})
	suite.Require().Nil(err)

	validColumns := []string{
		"created_at",
		"id",
		"user_id",
		"currency_id",
		"type",
		"total",
		"on_order",
		"locked",
	}

	for _, c := range validColumns {
		_, ok := columnMap[c]
		suite.Require().True(ok)
	}

	invalidColumns := []string{
		"sig", // intentionally remove.
		"idd",
		"UPDATE balance SET total = 100",
		"",
	}

	for _, c := range invalidColumns {
		_, ok := columnMap[c]
		suite.Require().False(ok)
	}
}

func (suite *SQLOptionsTestSuite) TestValidColumnMapInternal() {
	opt := SQLOptions{}
	// test m is nil
	var m *map[string]struct{}
	err := opt.validColumnMapInternal(m, &models.User{})
	suite.Require().NotNil(err)

	// test func(nil)
	err = opt.validColumnMapInternal(nil, &models.User{})
	suite.Require().NotNil(err)

	emptyMap := map[string]struct{}{}

	// test nil interface
	var a interface{}
	err = opt.validColumnMapInternal(&emptyMap, a)
	suite.Require().NotNil(err)

	// test nil pointer
	var bPtr *models.Balance
	err = opt.validColumnMapInternal(&emptyMap, bPtr)
	suite.Require().NotNil(err)

	// test nil
	err = opt.validColumnMapInternal(&emptyMap, nil)
	suite.Require().NotNil(err)

	// test success
	var b models.Balance
	err = opt.validColumnMapInternal(&emptyMap, b)
	suite.Require().Nil(err)

	err = opt.validColumnMapInternal(&emptyMap, &b)
	suite.Require().Nil(err)
}

func (suite *SQLOptionsTestSuite) TestLimit() {
	type testCase struct {
		optLimit     *int
		defaultLimit int
		allowNoLimit bool
		expected     int // 0 for no limit
	}

	overMaxLimit := maxLimit + 100
	twenty := 20
	zero := 0
	minusOne := -1
	cases := []testCase{
		testCase{nil, 1, true, 1},
		testCase{nil, overMaxLimit, true, maxLimit},
		testCase{&twenty, 1, true, 20},
		testCase{&overMaxLimit, 1, true, maxLimit},
		testCase{&minusOne, 1, true, 1},
		testCase{&zero, 1, false, 1},
		testCase{&zero, 1, true, 0},
	}

	for _, c := range cases {
		opt := SQLOptions{}
		opt.Limit = c.optLimit
		if c.expected == 0 {
			suite.Require().Nil(opt.limit(c.defaultLimit, c.allowNoLimit))
		} else {
			l := opt.limit(c.defaultLimit, c.allowNoLimit)
			suite.Require().Equal(c.expected, *l)
		}
	}
}

func TestSQLOptions(t *testing.T) {
	suite.Run(t, new(SQLOptionsTestSuite))
}
