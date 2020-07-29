package limiters

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jiarung/mochi/cache"
	"github.com/jiarung/mochi/infra/app"
)

// Test suite for System API Module
type WebsocketTestSuite struct {
	suite.Suite
}

func (suite *WebsocketTestSuite) SetupSuite() {
	var config struct {
		Cache cache.Config
	}
	app.SetConfig(nil, &config)
	cache.Initialize(config.Cache)
}

func (suite *WebsocketTestSuite) TearDownSuite() {
	cache.Finalize()
}

// Setup System API Module test suite
func (suite *WebsocketTestSuite) SetupTest() {
	ip := "10.10.10.1"
	ClearWebsocketAPIIP10RPS(ip)
	ClearWebsocketIPBlackList(ip)
	ClearWebsocketConnectionLimit(ip)
}

func (suite *WebsocketTestSuite) TestReachWebsocketAPIIP10RPS() {
	ip := "10.10.10.1"

	for i := 0; i < 10; i++ {
		assert.False(suite.T(), ReachWebsocketAPIIP10RPS(ip), "index: %v. Limit under rate 10/s\n", i)
	}

	assert.True(suite.T(), ReachWebsocketAPIIP10RPS(ip), "Not limit above rate 10/s")
}

func (suite *WebsocketTestSuite) TestReachWesocketIPBlackList() {
	ip := "10.10.10.1"

	assert.False(suite.T(), ReachWesocketIPBlackList(ip), "IP is in jail.\n")
	PutWebsocketIPBlackList(ip)
	assert.True(suite.T(), ReachWesocketIPBlackList(ip), "IP is not in jail.\n")
}

func (suite *WebsocketTestSuite) TestReachWebsocket10Connection() {
	ip := "10.10.10.1"
	limit := 10
	for i := 0; i < 10; i++ {
		assert.False(suite.T(), ReachWebsocketConnectionlimit(ip, strconv.Itoa(i), limit), "index: "+
			"%v. Limit under %s connections\n", i, limit)
	}

	assert.True(suite.T(), ReachWebsocketConnectionlimit(ip, "11", limit), "Not limit above %s "+
		"connections", limit)
}

// Test System API Module
func TestWebsocket(test *testing.T) {
	suite.Run(test, &WebsocketTestSuite{})
}
