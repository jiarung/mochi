package kubernetes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cobinhood/cobinhood-backend/common/config/misc"
	"github.com/cobinhood/cobinhood-backend/common/global"
	"github.com/cobinhood/cobinhood-backend/common/logging"
)

// Response response the JSON response to the k8s probes.
type Response struct {
	Result bool `json:"result"`
}

// LivenessProbe response to the k8s liveness probe.
func LivenessProbe(ctx *gin.Context) {
	// Liveness Probe JSON Response.
	response := Response{
		Result: !global.IsShuttingDown,
	}

	// Respond According to Current Shutdown Status.
	if global.IsShuttingDown {
		ctx.JSON(http.StatusServiceUnavailable, response)
	} else {
		ctx.JSON(http.StatusOK, response)
	}
}

// ReadinessProbe response to the k8s readiness probe.
func ReadinessProbe(ctx *gin.Context) {
	// Readiness Probe JSON Response.
	response := Response{
		Result: global.IsReady,
	}

	// Respond According to Current Shutdown Status.
	if global.IsReady {
		ctx.JSON(http.StatusOK, response)
	} else {
		ctx.JSON(http.StatusServiceUnavailable, response)
	}
}

// RegisterShutdownHandler register default shutdown handler.
func RegisterShutdownHandler(
	ctx context.Context,
	logger logging.Logger,
	server *http.Server) {
	// Start Goroutine to Catch Shutdown Signals.
	newLogger := logger.CloneLogger()
	go func() {
		logger := newLogger

		<-ctx.Done()
		// Do Graceful Shutdown.
		logger.Warn("Initiating graceful shutdown...")
		global.IsShuttingDown = true
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown: %s", err.Error())
		}
	}()
}

// StartHealthCheckServer starts kubernetes health check HTTP server.
func StartHealthCheckServer(ctx context.Context, logger logging.Logger) {
	router := gin.New()

	// Configure HTTP Router Settings.
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = false
	router.HandleMethodNotAllowed = false
	router.ForwardedByClientIP = true
	router.AppEngine = false
	router.UseRawPath = false
	router.UnescapePathValues = true

	// Setup HTTP Server.
	server := &http.Server{
		Addr:    misc.HealthCheckServerListenAddress(),
		Handler: router,
	}

	// Register Probes.
	router.GET("/alive", LivenessProbe)
	router.GET("/ready", ReadinessProbe)

	// Register shutdown handler.
	RegisterShutdownHandler(ctx, logger, server)

	newLogger := logger.CloneLogger()
	go func() {
		logger := newLogger
		// Start Running HTTP Server.
		if err := server.ListenAndServe(); err != nil {
			logger.Info("HealthCheckServer: %s", err.Error())
		}

		<-ctx.Done()
	}()
}
