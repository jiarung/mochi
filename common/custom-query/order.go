package customquery

import (
	"fmt"
	"strings"

	"github.com/jiarung/gorm"
)

// Order is a slice of order that will be evaluated to `ORDER BY` arguments in
// a SQL query.
type Order []order

func (o *Order) applyDB(db *gorm.DB, validColumnMap map[string]struct{}) (
	*gorm.DB, error) {
	result := db
	var err error
	for _, order := range []order(*o) {
		result, err = order.applyDB(result, validColumnMap)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

type order struct {
	// Column is the column to order
	Column string `json:"column" binding:"required"`
	// Keyword is the way rows will be ordered. (asc|desc)
	Keyword string `json:"keyword" binding:"required"`
}

var validOrderKeywordMap = map[string]struct{}{
	"asc":  struct{}{},
	"desc": struct{}{},
}

func (o *order) applyDB(db *gorm.DB, validColumnMap map[string]struct{}) (
	*gorm.DB, error) {
	arg, err := evaluateOrderWithValidator(*o, validColumnMap)
	if err != nil {
		return nil, err
	}
	return db.Order(arg), nil
}

func evaluateOrderWithValidator(order order, validColumnMap map[string]struct{}) (
	string, error) {
	err := order.validate(validColumnMap)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\"%s\" %s", order.Column, strings.ToUpper(order.Keyword)), nil
}

func (o *order) validate(validColumnMap map[string]struct{}) error {
	if _, isValidColumn := validColumnMap[o.Column]; !isValidColumn {
		return fmt.Errorf("invalid column. %s", o.Column)
	}
	// validate order keyword
	if _, ok := validOrderKeywordMap[string(o.Keyword)]; !ok {
		return fmt.Errorf("invalid order keyword (%s)", o.Keyword)
	}

	return nil
}
