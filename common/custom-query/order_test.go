package customquery

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OrderSuite tests order
type OrderSuite struct {
	suite.Suite
	invalidKeywordOrder          Order
	invalidColumnOrder           Order
	validOrder                   Order
	expectedValidOrderEvaluation []string
	validColumnMap               map[string]struct{}
}

func (suite *OrderSuite) SetupSuite() {
	suite.invalidKeywordOrder = Order([]order{
		{
			Column:  "b",
			Keyword: "ascc",
		},
	})
	suite.invalidColumnOrder = Order([]order{
		{
			Column:  "d",
			Keyword: "asc",
		},
	})
	suite.validOrder = Order([]order{
		{
			Column:  "b",
			Keyword: "asc",
		},
		{
			Column:  "a",
			Keyword: "desc",
		},
	})
	suite.expectedValidOrderEvaluation = []string{"\"b\" ASC", "\"a\" DESC"}
	suite.validColumnMap = map[string]struct{}{
		"a": struct{}{},
		"b": struct{}{},
	}
}

func (suite *OrderSuite) TestEvaluation() {
	for i, o := range []order(suite.validOrder) {
		result, err := evaluateOrderWithValidator(o, suite.validColumnMap)
		require.Nil(suite.T(), err)
		require.Equal(suite.T(), suite.expectedValidOrderEvaluation[i], result)
	}

	for _, o := range []order(suite.invalidColumnOrder) {
		_, err := evaluateOrderWithValidator(o, suite.validColumnMap)
		require.NotNil(suite.T(), err)
	}

	for _, o := range []order(suite.invalidKeywordOrder) {
		_, err := evaluateOrderWithValidator(o, suite.validColumnMap)
		require.NotNil(suite.T(), err)
	}

}

func TestOrder(t *testing.T) {
	suite.Run(t, new(OrderSuite))
}
