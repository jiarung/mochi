package apimochi

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

    "github.com/jiarung/mochi/common/scope-auth/test"
)

type ServerTestSuite struct {
	test.ServerTestSuite
}

func (at *ServerTestSuite) SetupSuite() {
	logger := logging.NewLoggerTag("api-mochi-server-test")
	cfg := Config{}
    //app.SetConfig(nil, &cfg)

	router := gin.New()
	registerMiddleware(logger, cfg, router)
    //at.ServiceName = cobxtypes.APICobx
	at.ServiceName = "api-mochi"
	at.MiddlewareRegisteredRouter = router
	at.RegisterModule = func(engine *gin.Engine) {
		registerModules(cfg, engine)
	}
}
func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
