package customquery_test

import (
	"net/http"

	apicontext "github.com/cobinhood/mochi/common/api/context"
	apiutils "github.com/cobinhood/mochi/common/api/utils"
	"github.com/cobinhood/mochi/common/custom-query"
	"github.com/cobinhood/mochi/database"
	models "github.com/cobinhood/mochi/models/exchange"
	"github.com/gin-gonic/gin"
)

func Example() {
	GetUsersHandler := func(ctx *gin.Context) {
		appCtx, err := apicontext.NewAppCtx(ctx, nil, nil, nil)
		if err != nil {
			// handle err
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		opt, err := customquery.NewSQLOptionsFromAppCtxQuery(appCtx)
		if err != nil {
			// handle err
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		db := database.GetDB(database.Default)
		result, limit, page, totalPage, totalCount, err := opt.Apply(
			db.Model(&models.User{}), 50, 1, 0)
		if err != nil {
			// handler evaluation error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if result.Error != nil {
			// handle query total page error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var users []models.User
		err = db.Find(&users).Error
		if err != nil {
			// handle find error
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(200, apiutils.Success(customquery.Result{
			Limit:      limit,
			Page:       page,
			TotalPage:  totalPage,
			TotalCount: totalCount,
			Data:       users,
		}))
	}

	gin.Default().GET("users", GetUsersHandler)
}
