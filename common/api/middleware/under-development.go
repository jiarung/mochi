package middleware

import (
	"github.com/gin-gonic/gin"

	apicontext "github.com/cobinhood/cobinhood-backend/common/api/context"
	apierrors "github.com/cobinhood/cobinhood-backend/common/api/errors"
	"github.com/cobinhood/cobinhood-backend/common/logging"
	"github.com/cobinhood/cobinhood-backend/common/utils"
)

// UnderDevelopment checks if environment is not on production environment.
func UnderDevelopment(ctx *gin.Context) {
	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		logging.NewLoggerTag("api:middleware:under-development").Error(
			"Error to get AppContext: %v\n", err)
		ctx.Abort()
		return
	}
	logger := appCtx.Logger()

	env := utils.Environment()
	if env == utils.Production {
		userID := ""
		if appCtx.UserID != nil {
			userID = appCtx.UserID.String()
		}
		logger.Notice(
			"env(%d) == prod userID(%s) %s %s",
			env,
			userID,
			ctx.Request.Method,
			ctx.Request.URL.Path)
		appCtx.SetError(apierrors.APIUnderDevelopment)
		return
	}
}
