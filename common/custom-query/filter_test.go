package customquery

import (
	"encoding/json"
	"testing"

	"github.com/cobinhood/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/mochi/database"
	"github.com/cobinhood/mochi/database/exchangedb"
	"github.com/cobinhood/mochi/infra/app"
	models "github.com/cobinhood/mochi/models/exchange"
)

// FilterSuite tests filter
type FilterSuite struct {
	suite.Suite
	db              *gorm.DB
	filterStrs      []string
	expectedFormats []string
	expectedValues  [][]interface{}
	validColumnMap  map[string]struct{}
}

func (suite *FilterSuite) SetupSuite() {
	var config struct {
		Database database.Config
	}
	app.SetConfig(nil, &config)
	database.Initialize(config.Database, database.Default)
	database.Reset(database.GetDB(database.Default), &exchangedb.DBApp{}, true)
	suite.db = database.GetDB(database.Default)

	filterStrs := make([]string, 0)
	expectedFormats := make([]string, 0)
	expectedValues := [][]interface{}{}

	// all n/a users
	filterStrs = append(filterStrs, `{"and":[{"equal":{"column":"tier_one_status","value":"not_available"}},{"equal":{"column":"tier_two_status","value":"not_available"}},{"equal":{"column":"tier_three_status","value":"not_available"}}]}`)
	expectedFormats = append(expectedFormats, "(\"tier_one_status\" = ?) AND (\"tier_two_status\" = ?) AND (\"tier_three_status\" = ?)")
	expectedValues = append(expectedValues, []interface{}{"not_available", "not_available", "not_available"})
	// tier one queued ids
	filterStrs = append(filterStrs, `{"equal":{"column":"tier_one_status","value":"queued"}}`)
	expectedFormats = append(expectedFormats, "\"tier_one_status\" = ?")
	expectedValues = append(expectedValues, []interface{}{"queued"})
	// tier two queued ids
	filterStrs = append(filterStrs, `{"equal":{"column":"tier_two_status","value":"queued"}}`)
	expectedFormats = append(expectedFormats, "\"tier_two_status\" = ?")
	expectedValues = append(expectedValues, []interface{}{"queued"})
	// tier three rejected ids
	filterStrs = append(filterStrs, `{"equal":{"column":"tier_three_status","value":"rejected"}}`)
	expectedFormats = append(expectedFormats, "\"tier_three_status\" = ?")
	expectedValues = append(expectedValues, []interface{}{"rejected"})
	// all verified ids
	filterStrs = append(filterStrs, `{"and":[{"equal":{"column":"tier_one_status","value":"verified"}},{"equal":{"column":"tier_two_status","value":"verified"}},{"equal":{"column":"tier_three_status","value":"verified"}}]}`)
	expectedFormats = append(expectedFormats, "(\"tier_one_status\" = ?) AND (\"tier_two_status\" = ?) AND (\"tier_three_status\" = ?)")
	expectedValues = append(expectedValues, []interface{}{"verified", "verified", "verified"})
	// tier one queued or tier two queued
	filterStrs = append(filterStrs, `{"or":[{"equal":{"column":"tier_one_status","value":"queued"}},{"equal":{"column":"tier_two_status","value":"queued"}}]}`)
	expectedFormats = append(expectedFormats, "(\"tier_one_status\" = ?) OR (\"tier_two_status\" = ?)")
	expectedValues = append(expectedValues, []interface{}{"queued", "queued"})

	suite.filterStrs = filterStrs
	suite.expectedFormats = expectedFormats
	suite.expectedValues = expectedValues
	suite.validColumnMap = map[string]struct{}{
		"tier_one_status":   struct{}{},
		"tier_three_status": struct{}{},
	}
}

func (suite *FilterSuite) TearDownSuite() {
	database.DeleteAllData(&exchangedb.DBApp{})
	database.Finalize()
}

func (suite *FilterSuite) TestApply() {
	db := suite.db.Model(&models.User{})
	for i, filterStr := range suite.filterStrs {
		strBytes := []byte(filterStr)
		var f Filter
		err := json.Unmarshal(strBytes, &f)
		require.Nil(suite.T(), err)

		_, err = f.applyDB(db, suite.validColumnMap)
		switch i {
		case 0, 2, 4, 5:
			require.NotNil(suite.T(), err)
		default:
			require.Nil(suite.T(), err)

		}
	}
}

func (suite *FilterSuite) TestEvaluation() {

	for i, filterStr := range suite.filterStrs {
		strBytes := []byte(filterStr)
		var f Filter
		err := json.Unmarshal(strBytes, &f)
		require.Nil(suite.T(), err)
		format, values, err := evaluateFilterWithValidator(f, suite.validColumnMap)
		switch i {
		case 0, 2, 4, 5:
			require.NotNil(suite.T(), err)
		default:
			require.Nil(suite.T(), err)
			require.Equal(suite.T(), suite.expectedFormats[i], format)
			require.Equal(suite.T(), suite.expectedValues[i], values)
		}
	}
}

func (suite *FilterSuite) TestValidation() {
	for i, filterStr := range suite.filterStrs {
		strBytes := []byte(filterStr)
		var f Filter
		err := json.Unmarshal(strBytes, &f)
		require.Nil(suite.T(), err)
		_, _, err = evaluateFilterWithValidator(f, suite.validColumnMap)
		switch i {
		case 0, 2, 4, 5:
			require.NotNil(suite.T(), err)
		default:
			require.Nil(suite.T(), err)

		}
	}

}

func TestFilter(t *testing.T) {
	suite.Run(t, new(FilterSuite))
}
