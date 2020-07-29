// Copyright 2018-2019 Cobinhood Inc. All rights reserved.

package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LineTestSuite struct {
	suite.Suite
}

func (s *LineTestSuite) TestParseLinebot() {
	str := `{
		"events":
		[{
			"replyToken": "00000000000000000000000000000000",
			"type": "message",
			"timestamp": 1552372949883,
			"source": {
				"type": "user",
				"userId": "Udeadbeefdeadbeefdeadbeefdeadbeef"
			},
			"message": {
				"id": "100001",
				"type": "text",
				"text": "Hello, world"
			}
		}, {
			"replyToken": "ffffffffffffffffffffffffffffffff",
			"type": "message",
			"timestamp": 1552372949883,
			"source": {
				"type": "user",
				"userId": "Udeadbeefdeadbeefdeadbeefdeadbeef"
			},
			"message": {
				"id": "100002",
				"type": "sticker",
				"packageId": "1",
				"stickerId": "1"
			}
		}]
	}`

	events, err := ParseLinebot([]byte(str))
	s.Require().Nil(err)
	e := events[0]
	s.Require().Equal(e.Source.UserID, "Udeadbeefdeadbeefdeadbeefdeadbeef")
}

func TestLine(t *testing.T) {
	suite.Run(t, new(LineTestSuite))
}
