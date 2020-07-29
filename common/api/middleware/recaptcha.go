package middleware

import (
	"github.com/gin-gonic/gin"

	apicontext "github.com/jiarung/mochi/common/api/context"
	apiutils "github.com/jiarung/mochi/common/api/utils"
	"github.com/jiarung/mochi/common/utils"
	"github.com/jiarung/mochi/infra/api/middleware/logger"
)

// RecaptchaRequired recaptcha validation required
func RecaptchaRequired(ctx *gin.Context) {
	var err error

	appCtx, err := apicontext.GetAppContext(ctx)
	if err != nil {
		logger := logger.Get(ctx)
		logger.Error("Error to get AppContext. Err: %v", err)
		ctx.Abort()
		return
	}
	logger := appCtx.Logger()

	captcha := apiutils.SelectCaptcha(ctx)
	platform := ctx.Request.Header.Get("platform")

	// Bypass on staging for automated testing with mobile.
	if utils.Environment() == utils.Staging &&
		(platform == "iOS" || platform == "Android") {
		return
	}

	switch captcha {
	case apiutils.RECAPTCHA:
		// for recaptcha verification
		gtoken := ctx.Request.Header.Get("g-recaptcha-token")
		err = apiutils.VerifyRecaptcha(platform, gtoken)
		if err != nil {
			logger.Warn("Recaptcha: verify captcha fail: %v", err)
			ctx.Abort()
			return
		}
	case apiutils.NECAPTCHA:
		// for ali cloud verification
		netoken := ctx.Request.Header.Get("ne-captcha-token")

		err = apiutils.VerifyNECaptcha(platform, netoken)
		if err != nil {
			logger.Warn("NE Captcha: verify captcha fail: %v", err)
			ctx.Abort()
			return
		}
	default:
		logger.Warn("Verification: verify fail.")
		ctx.Abort()
	}
}
