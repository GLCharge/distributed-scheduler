package handlers

import (
	"net/http"
	"os"

	"github.com/GLCharge/distributed-scheduler/handlers/jobs"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/GLCharge/distributed-scheduler/store/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	DB       *sqlx.DB
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler {

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// ==================
	// Health Check

	router.GET("/health", healthCheck(cfg))

	// ==================
	// Jobs

	jobStore := postgres.New(cfg.DB, cfg.Log)

	jobService := job.NewService(jobStore, cfg.Log)

	jobsHandler := jobs.NewJobsHandler(jobService)

	jobsRouter := router.Group("/v1/jobs")
	{
		jobsRouter.POST("", jobsHandler.CreateJob())
		jobsRouter.GET("/:id", jobsHandler.GetJob())
		jobsRouter.DELETE("/:id", jobsHandler.DeleteJob())
		jobsRouter.GET("", jobsHandler.ListJobs())
	}

	return router
}

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

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}
