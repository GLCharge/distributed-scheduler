package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/GLCharge/distributed-scheduler/handlers/jobs"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/GLCharge/distributed-scheduler/store/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown           chan os.Signal
	Log                *zap.SugaredLogger
	DB                 *sqlx.DB
	MaxJobLockDuration time.Duration
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler {

	// Create a new Gin router
	router := gin.New()

	// Use Gin's built-in logger and recovery middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// ==================
	// Health Check

	// Define a route for the health check endpoint
	router.GET("/health", healthCheck(cfg))

	// ==================
	// Jobs

	// Create a new PostgresSQL job store
	jobStore := postgres.New(cfg.DB, cfg.Log, cfg.MaxJobLockDuration)

	// Create a new job service with the job store and logger
	jobService := job.NewService(jobStore, cfg.Log)

	// Create a new jobs handler with the job service
	jobsHandler := jobs.NewJobsHandler(jobService)

	// Define a group of routes for the jobs endpoint
	jobsRouter := router.Group("/v1/jobs")
	{
		jobsRouter.POST("", jobsHandler.CreateJob())
		jobsRouter.GET("/:id", jobsHandler.GetJob())
		jobsRouter.PUT("/:id", jobsHandler.UpdateJob())
		jobsRouter.DELETE("/:id", jobsHandler.DeleteJob())
		jobsRouter.GET("", jobsHandler.ListJobs())
		jobsRouter.GET("/:id/executions", jobsHandler.GetJobExecutions())
	}

	// Return the router as a http.Handler
	return router
}

// healthCheck returns a Gin handler function for the health check endpoint
func healthCheck(cfg APIMuxConfig) gin.HandlerFunc {

	return func(c *gin.Context) {

		// Check the database connection
		if err := cfg.DB.Ping(); err != nil {
			cfg.Log.Errorw("database ping failed", "error", err)
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
