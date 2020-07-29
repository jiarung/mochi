package customquery

import (
	"fmt"
	"strings"

	"github.com/jiarung/gorm"

	apicontext "github.com/jiarung/mochi/common/api/context"
	apierrors "github.com/jiarung/mochi/common/api/errors"
)

// PageFindParams defines page find utility params.
type PageFindParams struct {
	AppCtx       *apicontext.AppContext
	SQL          *gorm.DB
	Object       interface{}
	DftOrder     string
	AllowNoLimit bool
	MaxRecords   int
}

// PageFind makes an easy way for use customquery.
func PageFind(params *PageFindParams) (*Result, error) {
	appCtx := params.AppCtx
	sql := params.SQL
	object := params.Object
	dftOrder := params.DftOrder
	allowNoLimit := params.AllowNoLimit
	maxRecords := params.MaxRecords

	opt, err := NewSQLOptionsFromAppCtxQuery(appCtx)
	if err != nil {
		appCtx.SetError(apierrors.InvalidQueryParameter)
		return nil, fmt.Errorf("invalid params. err(%s)", err)
	}

	if opt.Order == nil {
		if len(dftOrder) > 0 {
			if !strings.Contains(strings.ToUpper(dftOrder), "DESC") &&
				!strings.Contains(strings.ToUpper(dftOrder), "ASC") {
				return nil, fmt.Errorf("invalid order(%v)", dftOrder)
			}
			sql = sql.Order(dftOrder)
		} else {
			sql = sql.Order("created_at DESC")
		}
	}

	query, limit, page, totalPage, totalCount, err := opt.Apply(
		sql, 50, 1, maxRecords, allowNoLimit)
	if err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to apply sql option. err(%s)", err)
	}

	if err := query.Find(object).Error; err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to find: %v", err)
	}

	return &Result{
		Limit:      limit,
		Page:       page,
		TotalPage:  totalPage,
		TotalCount: totalCount,
		Data:       nil,
	}, nil
}

// PageScan makes an easy way for use customquery.
func PageScan(params *PageFindParams) (*Result, error) {
	appCtx := params.AppCtx
	sql := params.SQL
	object := params.Object
	dftOrder := params.DftOrder
	allowNoLimit := params.AllowNoLimit
	maxRecords := params.MaxRecords

	opt, err := NewSQLOptionsFromAppCtxQuery(appCtx)
	if err != nil {
		appCtx.SetError(apierrors.InvalidQueryParameter)
		return nil, fmt.Errorf("invalid params. err(%s)", err)
	}

	if opt.Order == nil {
		if len(dftOrder) > 0 {
			if !strings.Contains(strings.ToUpper(dftOrder), "DESC") &&
				!strings.Contains(strings.ToUpper(dftOrder), "ASC") {
				return nil, fmt.Errorf("invalid order(%v)", dftOrder)
			}
			sql = sql.Order(dftOrder)
		} else {
			sql = sql.Order("created_at DESC")
		}
	}

	query, limit, page, totalPage, totalCount, err := opt.Apply(
		sql, 50, 1, maxRecords, allowNoLimit)
	if err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to apply sql option. err(%s)", err)
	}

	if err := query.Scan(object).Error; err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to scan: %v", err)
	}

	return &Result{
		Limit:      limit,
		Page:       page,
		TotalPage:  totalPage,
		TotalCount: totalCount,
		Data:       nil,
	}, nil
}

// PageFindFromGinBody makes an easy way for use customquery.
func PageFindFromGinBody(params *PageFindParams) (*Result, error) {
	appCtx := params.AppCtx
	sql := params.SQL
	object := params.Object
	dftOrder := params.DftOrder
	allowNoLimit := params.AllowNoLimit
	maxRecords := params.MaxRecords

	opt, err := NewSQLOptionsFromAppCtxBody(appCtx)
	if err != nil {
		appCtx.SetError(apierrors.InvalidQueryParameter)
		return nil, fmt.Errorf("invalid params. err(%s)", err)
	}

	if opt.Order == nil {
		if len(dftOrder) > 0 {
			if !strings.Contains(strings.ToUpper(dftOrder), "DESC") &&
				!strings.Contains(strings.ToUpper(dftOrder), "ASC") {
				return nil, fmt.Errorf("invalid order(%v)", dftOrder)
			}
			sql = sql.Order(dftOrder)
		} else {
			sql = sql.Order("created_at DESC")
		}
	}

	query, limit, page, totalPage, totalCount, err := opt.Apply(
		sql, 50, 1, maxRecords, allowNoLimit)
	if err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to apply sql option. err(%s)", err)
	}

	if err := query.Find(object).Error; err != nil {
		appCtx.SetError(apierrors.DBError)
		return nil, fmt.Errorf("failed to Find object: %v", err)
	}

	return &Result{
		Limit:      limit,
		Page:       page,
		TotalPage:  totalPage,
		TotalCount: totalCount,
		Data:       nil,
	}, nil
}
