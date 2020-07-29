package middleware

import (
	"github.com/gin-gonic/gin"
	ginTracer "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"

	"github.com/cobinhood/cobinhood-backend/common"
)

// DatadogMiddleware provide datadog logging.
func DatadogMiddleware(serviceName string) gin.HandlerFunc {
	if !common.TracerEnabled() {
		return func(c *gin.Context) {
		}
	}

	return ginTracer.Middleware(serviceName)
}
