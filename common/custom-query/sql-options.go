package customquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	apicontext "github.com/jiarung/mochi/common/api/context"
	"github.com/jiarung/gorm"
)

const maxLimit = 100

// NoLimit stands for no query limit.
const NoLimit = 0

// SQLOptions is the struct that allows user to create custom query criteria.
type SQLOptions struct {
	Filter *Filter `json:"filter;omitempty"`
	Order  *Order  `json:"order;omitempty"`
	Limit  *int    `json:"limit;omitempty"`
	Page   *int    `json:"page;omitempty"`
}

// NewSQLOptionsFromAppCtxBody returns a new instance of SQLOptions,
// validating the body of AppContext and bind to the model
func NewSQLOptionsFromAppCtxBody(appCtx *apicontext.AppContext) (
	*SQLOptions, error) {
	var opt SQLOptions
	err := appCtx.BindJSON(&opt)
	if err != nil {
		return nil, err
	}

	return &opt, nil
}

// NewSQLOptionsFromAppCtxQuery returns a new instance of SQLOptions,
// validating the body of AppContext and bind to the model
func NewSQLOptionsFromAppCtxQuery(appCtx *apicontext.AppContext) (
	*SQLOptions, error) {
	opt := SQLOptions{}

	if filterStr, exists := appCtx.GetQuery("filter"); exists {
		filterB := []byte(filterStr)
		var filter Filter
		err := json.Unmarshal(filterB, &filter)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter. err: %v", err)
		}
		opt.Filter = &filter
	}

	if orderStr, exists := appCtx.GetQuery("order"); exists {
		orderB := []byte(orderStr)
		var order Order
		err := json.Unmarshal(orderB, &order)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal order. err: %v", err)
		}
		opt.Order = &order
	}

	if limitStr, exists := appCtx.GetQuery("limit"); exists {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse limit. err: %v", err)
		}
		opt.Limit = &limit
	}

	if pageStr, exists := appCtx.GetQuery("page"); exists {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse page. err: %v", err)
		}
		opt.Page = &page
	}

	return &opt, nil
}

// Apply applies db with query params, assuming params is valid, returning
// result, limit, page, totalPage and error. `nil` db and `-1` value of
// limit/page/totalPage are all equivalent with `err != nil`.
// The query will be unlimited if opt.Limit == nil and allowNoLimitArg == [bool]
// Note that the db.Value should not be nil. The validation allows only the gorm
// defined columns of db.Value (model), and fails when db.Value == nil.
func (opt *SQLOptions) Apply(db *gorm.DB, defaultLimit, defaultPage, maxRecords int,
	allowNoLimitArg ...bool) (
	result *gorm.DB, limit, page, totalPage, totalCount int, err error) {

	allowNoLimit := false
	if len(allowNoLimitArg) > 0 {
		allowNoLimit = allowNoLimitArg[0]
	}

	result = db

	result, err = opt.applyFilterIfNeeded(result)
	if err != nil {
		return nil, -1, -1, -1, -1, err
	}

	result, err = opt.applyOrderIfNeeded(result)
	if err != nil {
		return nil, -1, -1, -1, -1, err
	}

	if defaultLimit < 1 {
		err = errors.New("default limit should be greater than 1")
		return nil, -1, -1, -1, -1, err
	}

	if defaultPage < 1 {
		err = errors.New("default page should be greater than 1")
		return nil, -1, -1, -1, -1, err
	}

	var count int
	if maxRecords != 0 {
		count = maxRecords
	} else {
		countResult := result.Count(&count)
		if countResult.Error != nil {
			err = fmt.Errorf("failed to count result. err: %v", countResult.Error)
			return nil, -1, -1, -1, 1, err
		}
	}

	totalCount = count
	limitPtr := opt.limit(defaultLimit, allowNoLimit)
	if limitPtr != nil {
		limit = *limitPtr
		page = opt.page(defaultPage)
		offset := opt.offset(limit, page)
		result = result.Limit(limit).Offset(offset)
		totalPage = count / limit
		if count%limit != 0 {
			totalPage++
		}
	}

	return
}

func (opt *SQLOptions) applyFilterIfNeeded(db *gorm.DB) (
	*gorm.DB, error) {
	if opt.Filter != nil {
		validColumnMap, err := opt.validColumnMap(db.Value)
		if err != nil {
			return nil, err
		}
		result, err := opt.Filter.applyDB(db, validColumnMap)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return db, nil
}

func (opt *SQLOptions) applyOrderIfNeeded(db *gorm.DB) (
	*gorm.DB, error) {
	if opt.Order != nil {
		validColumnMap, err := opt.validColumnMap(db.Value)
		if err != nil {
			return nil, err
		}
		result, err := opt.Order.applyDB(db, validColumnMap)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return db, nil
}

// limit returns the limit pointer and nil for no limit.
func (opt *SQLOptions) limit(defaultLimit int, allowNoLimit bool) *int {
	if opt.Limit == nil {
		l := int(math.Min(float64(defaultLimit), float64(maxLimit)))
		return &l
	}
	if allowNoLimit && *opt.Limit == NoLimit {
		return nil
	}
	if *opt.Limit <= 0 {
		l := int(math.Min(float64(defaultLimit), float64(maxLimit)))
		return &l
	}
	l := int(math.Min(float64(*opt.Limit), float64(maxLimit)))
	return &l
}

func (opt *SQLOptions) page(defaultPage int) int {
	if opt.Page != nil {
		return *opt.Page
	}
	return defaultPage
}

func (opt *SQLOptions) offset(limit, page int) int {
	return (page - 1) * limit
}

func (opt *SQLOptions) validColumnMap(model interface{}) (
	map[string]struct{}, error) {
	m := map[string]struct{}{}
	err := opt.validColumnMapInternal(&m, model)
	if err != nil {
		return nil, err
	}
	delete(m, "sig")
	return m, nil
}

func (opt *SQLOptions) validColumnMapInternal(columnMap *map[string]struct{}, model interface{}) (err error) {
	if columnMap == nil || *columnMap == nil {
		return errors.New("nil destination map")
	}

	if model == nil {
		return errors.New("nil model")
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelValue := reflect.ValueOf(model)
		if modelValue.IsNil() {
			return fmt.Errorf("nil model value. %v", modelType)
		}
		return opt.validColumnMapInternal(columnMap, modelValue.Elem().Interface())
	}

	if modelType.Kind() != reflect.Struct {
		return fmt.Errorf("invalid kind other than struct. %s", modelType.Kind())
	}

	for i := 0; i < modelType.NumField(); i++ {
		gormTag := modelType.Field(i).Tag.Get("gorm")
		if len(gormTag) > 0 {
			strs := strings.Split(gormTag, ";")
			for _, str := range strs {
				if splited := strings.Split(str, ":"); len(splited) == 2 &&
					splited[0] == "column" {
					m := *columnMap
					m[splited[1]] = struct{}{}
					break
				}
			}
			continue
		}

		modelValue := reflect.ValueOf(model)
		valueField := modelValue.Field(i)
		if valueField.Kind() == reflect.Struct {
			if !valueField.CanInterface() {
				return fmt.Errorf("can't interface %v", valueField)
			}
			opt.validColumnMapInternal(columnMap, valueField.Interface())
		}
	}

	return nil
}
