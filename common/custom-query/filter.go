package customquery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cobinhood/gorm"
)

// Filter is a nested map that will be evaluated to `Where` arguments in a SQL query.
// Each layer is a single-key map, with operator as the key and parameters as
// the value. There are two kinds of filters (operators), comparison filters and
// logic filters. Comparison filter is the atomic element of a filter and logic
// filter is to combine multiple filters into one single filter
type Filter map[string]interface{}

func (f *Filter) applyDB(db *gorm.DB, validColumnMap map[string]struct{}) (result *gorm.DB, err error) {
	format, values, _err := evaluateFilterWithValidator(*f, validColumnMap)
	if _err != nil {
		result = nil
		err = _err
		return
	}
	result = db.Where(format, values...)
	err = nil
	return
}

const (
	equalOperatorStr              = "equal"
	notEqualOperatorStr           = "not_equal"
	greaterThanOperatorStr        = "greater_than"
	greaterThanOrEqualOperatorStr = "greater_than_or_equal"
	smallerThanOperatorStr        = "smaller_than"
	smallerThanOrEqualOperatorStr = "smaller_than_or_equal"
	inOperatorStr                 = "in"
	likeOperatorStr               = "like"
	ilikeOperatorStr              = "ilike"
	betweenOperatorStr            = "between"
	andOperatorStr                = "and"
	orOperatorStr                 = "or"
)

var comparisonOperators = []string{
	equalOperatorStr,
	notEqualOperatorStr,
	greaterThanOperatorStr,
	greaterThanOrEqualOperatorStr,
	smallerThanOperatorStr,
	smallerThanOrEqualOperatorStr,
	betweenOperatorStr,
	inOperatorStr,
	likeOperatorStr,
	ilikeOperatorStr,
}
var logicOperators = []string{andOperatorStr, orOperatorStr}

var operatorMap = map[string]string{
	equalOperatorStr:              "=",
	notEqualOperatorStr:           "!=",
	greaterThanOperatorStr:        ">",
	greaterThanOrEqualOperatorStr: ">=",
	smallerThanOperatorStr:        "<",
	smallerThanOrEqualOperatorStr: "<=",
	inOperatorStr:                 "IN",
	likeOperatorStr:               "LIKE",
	ilikeOperatorStr:              "ILIKE",
	betweenOperatorStr:            "BETWEEN",
	andOperatorStr:                "AND",
	orOperatorStr:                 "OR",
}

func evaluateFilterWithValidator(filter map[string]interface{},
	validColumnMap map[string]struct{}) (
	format string, values []interface{}, err error) {
	for _, comparisonOp := range comparisonOperators {
		payload, ok := filter[comparisonOp]
		if !ok {
			continue
		}
		payloadMap, valid := payload.(map[string]interface{})
		if !valid {
			format = ""
			values = nil
			err = fmt.Errorf("can't assert to %s map (%s)", comparisonOp, payload)
			return
		}
		column, columnExists := payloadMap["column"]
		columnStr := fmt.Sprintf("%v", column)
		value, valueExists := payloadMap["value"]
		if !columnExists || !valueExists {
			format = ""
			values = nil
			err = fmt.Errorf("invalid %s map (%s)", comparisonOp, payload)
			return
		}

		if _, isColumnValid := validColumnMap[columnStr]; !isColumnValid {
			format = ""
			values = nil
			err = fmt.Errorf("filter comparison with invalid column. %s", columnStr)
			return
		}

		if comparisonOp == inOperatorStr {
			valueArr, isValidValue := value.([]interface{})
			if !isValidValue {
				format = ""
				values = nil
				err = fmt.Errorf("invalid value for in. %s", value)
				return
			}
			format = fmt.Sprintf("\"%s\" %s (?)", columnStr, operatorMap[comparisonOp])
			values = []interface{}{valueArr}
			err = nil
			return
		}

		if comparisonOp == betweenOperatorStr {
			valueArr, isValidValue := value.([]interface{})
			if !isValidValue || len(valueArr) != 2 {
				format = ""
				values = nil
				err = fmt.Errorf("invalid value for between. %s", value)
				return
			}
			format = fmt.Sprintf("\"%s\" %s ? AND ?", columnStr, operatorMap[comparisonOp])
			values = valueArr
			err = nil
			return
		}

		format = fmt.Sprintf("\"%s\" %s ?", columnStr, operatorMap[comparisonOp])
		values = []interface{}{value}
		err = nil
		return
	}

	for _, combinedOp := range logicOperators {
		payload, ok := filter[combinedOp]
		if !ok {
			continue
		}
		payloadArr, valid := payload.([]interface{})
		if !valid || len(payloadArr) < 2 {
			format = ""
			values = nil
			err = fmt.Errorf("can't assert to %s array (%s)", combinedOp, payload)
			return
		}

		formats := make([]string, 0)
		values = make([]interface{}, 0)

		for _, payload := range payloadArr {
			subFilter, valid := payload.(map[string]interface{})
			if !valid {
				format = ""
				values = nil
				err = fmt.Errorf("invalid sub filter (%s)", payload)
				return
			}
			subFormat,
				subValues,
				_err := evaluateFilterWithValidator(subFilter, validColumnMap)
			if _err != nil {
				format = ""
				values = nil
				err = _err
				return
			}
			formats = append(formats, fmt.Sprintf("(%s)", subFormat))
			values = append(values, subValues...)
		}

		format = strings.Join(
			formats,
			fmt.Sprintf(" %s ", operatorMap[combinedOp]),
		)
		err = nil
		return
	}

	format = ""
	values = []interface{}{}
	err = errors.New("invalid filter")
	return
}
