package database

import (
	"testing"

//	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InitializeTablesTestSuite struct {
	suite.Suite
}

func (suite *InitializeTablesTestSuite) SetupSuite() {
	defaultDB.Initialize(Default)
	defaultDB.Reset(defaultDB.GetDB(Default), DBApp{})
}

func (suite *InitializeTablesTestSuite) TearDownSuite() {
	defaultDB.Finalize()
}
/*
func (suite *InitializeTablesTestSuite) TestEmpty() {
	db := GetDB(Default)
	require.NotNil(suite.T(), db, "GetDB() must not return nil")

	var recordList []interface{}
	err := InitializeTables(recordList, db)
	ok := assert.Nil(suite.T(), err, "InitializeTables() failed. err(%s)", err)
	if !ok {
		suite.logRecordList(recordList)
		return
	}
}*/
/*

func (suite *InitializeTablesTestSuite) TestNotDataModel() {
	db := database.GetDB(database.Default)
	require.NotNil(suite.T(), db, "GetDB() must not return nil")

	recordList := []interface{}{
		struct {
			Name string
		}{
			Name: "123",
		}}
	err := database.InitializeTables(recordList, db)
	ok := assert.NotNil(
		suite.T(),
		err,
		"InitializeTables() with a type which is not a data model must fail.")
	if !ok {
		suite.logRecordList(recordList)
		return
	}
}

func (suite *InitializeTablesTestSuite) TestNotPointer() {
	db := database.GetDB(database.Default)
	require.NotNil(suite.T(), db, "GetDB() must not return nil")

	recordList := []interface{}{
		models.Currency{
			ID:      "BTC",
			Name:    "Bitcoin",
			MinUnit: decimal.NewFromFloat(123.45),
		}}
	err := database.InitializeTables(recordList, db)
	ok := assert.NotNil(
		suite.T(),
		err,
		"InitializeTables() with non-pointer data model must fail.")
	if !ok {
		suite.logRecordList(recordList)
		return
	}
}

func (suite *InitializeTablesTestSuite) TestOneRecord() {
	database.DeleteAllData(&DBApp{})

	db := database.GetDB(database.Default)
	require.NotNil(suite.T(), db, "GetDB() must not return nil")

	btc := models.Currency{
		ID:      "BTC",
		Name:    "Bitcoin",
		MinUnit: decimal.NewFromFloat(123.45),
	}
	recordList := []interface{}{&btc}

	err := database.InitializeTables(recordList, db)
	ok := assert.Nil(suite.T(), err, "InitializeTables() failed. err(%s)", err)
	if !ok {
		suite.logRecordList(recordList)
		return
	}
	if result := db.Exec(
		"delete from currency where id = (?)", btc.ID); result.Error != nil {
		suite.T().Logf(
			"db.Exec() failed. btc.ID(%s) result.Error(%s)",
			btc.ID,
			result.Error)
	}
}

func (suite *InitializeTablesTestSuite) TestMultiRecord() {
	database.DeleteAllData(&DBApp{})

	db := database.GetDB(database.Default)
	require.NotNil(suite.T(), db, "GetDB() must not return nil")

	btc := models.Currency{
		ID:      "BTC",
		Name:    "Bitcoin",
		MinUnit: decimal.NewFromFloat(123.45),
	}
	eth := models.Currency{
		ID:      "ETH",
		Name:    "Ether",
		MinUnit: decimal.NewFromFloat(678.9),
	}
	tradingPair := models.TradingPair{
		ID:              "BTC-ETH",
		BaseCurrencyID:  "BTC",
		QuoteCurrencyID: "ETH",
		BaseMaxSize:     decimal.NewFromFloat(1234.567),
		BaseMinSize:     decimal.NewFromFloat(89012.3),
	}
	recordList := []interface{}{&btc, &eth, &tradingPair}

	err := database.InitializeTables(recordList, db)
	ok := assert.Nil(suite.T(), err, "InitializeTables() failed. err(%s)", err)
	if !ok {
		suite.logRecordList(recordList)
		return
	}
	if result := db.Exec(
		"delete from trading_pair where id = (?)",
		tradingPair.ID); result.Error != nil {
		suite.T().Logf(
			"db.Exec() failed. tradingPair.ID(%s) result.Error(%s)",
			tradingPair.ID,
			result.Error)
	}
	if result := db.Exec(
		"delete from currency where id = (?)", btc.ID); result.Error != nil {
		suite.T().Logf(
			"db.Exec() failed. btc.ID(%s) result.Error(%s)",
			btc.ID,
			result.Error)
	}
	if result := db.Exec(
		"delete from currency where id = (?)", eth.ID); result.Error != nil {
		suite.T().Logf(
			"db.Exec() failed. eth.ID(%s) result.Error(%s)",
			eth.ID,
			result.Error)
	}
}

*/

func (suite *InitializeTablesTestSuite) logRecordList(
	recordList []interface{}) {
	for idx, record := range recordList {
		suite.T().Logf("idx(%d), %+v", idx, record)
	}
}

func TestInitializeTables(t *testing.T) {
	suite.Run(t, &InitializeTablesTestSuite{})
}
