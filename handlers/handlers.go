package handlers

import (
	"context"
	"github.com/GLCharge/otelzap"
	ginzap "github.com/gin-contrib/zap"
	timeout "github.com/vearne/gin-timeout"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/GLCharge/distributed-scheduler/foundation/database"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/GLCharge/distributed-scheduler/store/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Log     *otelzap.Logger
	DB      *sqlx.DB
	OpenApi OpenApiConfig
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler {

	// Create a new Gin router
	router := gin.New()

	// Use Gin's built-in logger and recovery middleware
	router.Use(
		ginzap.RecoveryWithZap(cfg.Log, true),
		ginzap.Ginzap(cfg.Log, time.RFC3339, true),
		timeout.Timeout(timeout.WithErrorHttpCode(http.StatusServiceUnavailable)),
	)

	// ==================
	// Health Check

	// Define a route for the health check endpoint
	router.GET("/health", healthCheck(cfg))

	// ==================
	// OpenAPI (will only mount if enabled)
	OpenApiRoute(cfg.OpenApi, router)

	// ==================
	// Jobs

	// Create a new PostgresSQL job store
	jobStore := postgres.New(cfg.DB, cfg.Log)

	// Create a new job service with the job store and logger
	jobService := job.NewService(jobStore, cfg.Log)

	// Create a new jobs handler with the job service
	jobsHandler := NewJobsHandler(jobService)

	// Define a group of routes for the jobs endpoint
	JobsRoutesV1(router, jobsHandler)

	// Return the router as a http.Handler
	return router
}

// Runner API
func RunnerAPI(cfg APIMuxConfig) http.Handler {
	// Create a new Gin router
	router := gin.New()

	// Use Gin's built-in logger and recovery middleware
	router.Use(
		ginzap.RecoveryWithZap(cfg.Log, true),
		ginzap.Ginzap(cfg.Log, time.RFC3339, true),
		timeout.Timeout(timeout.WithErrorHttpCode(http.StatusServiceUnavailable)),
	)

	// ==================
	// Health Check

	// Define a route for the health check endpoint
	router.GET("/health", healthCheck(cfg))

	return router
}

// healthCheck returns a Gin handler function for the health check endpoint
func healthCheck(cfg APIMuxConfig) gin.HandlerFunc {

	return func(c *gin.Context) {

		// Check the database connection
		if err := database.StatusCheck(context.Background(), cfg.DB); err != nil {
			cfg.Log.Error("database status check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "database ping failed",
			})
			return
		}

		// Return a JSON response with a status of "OK"
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}
