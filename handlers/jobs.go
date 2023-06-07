package handlers

import (
	"github.com/google/uuid"
	"net/http"
	"strconv"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/gin-gonic/gin"
)

func JobsRoutesV1(router *gin.Engine, jobsHandler *Jobs) {
	jobsRouter := router.Group("/v1/jobs")
	{
		jobsRouter.POST("", jobsHandler.CreateJob())
		jobsRouter.GET("/:id", jobsHandler.GetJob())
		jobsRouter.PUT("/:id", jobsHandler.UpdateJob())
		jobsRouter.DELETE("/:id", jobsHandler.DeleteJob())
		jobsRouter.GET("", jobsHandler.ListJobs())
		jobsRouter.GET("/:id/executions", jobsHandler.GetJobExecutions())
	}
}

func NewJobsHandler(service *job.Service) *Jobs {
	return &Jobs{
		service: service,
	}
}

type Jobs struct {
	service *job.Service
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateJob godoc
// @Summary Create a job
// @Description Create a job with the given job create request
// @Tags jobs
// @Accept json
// @Produce json
// @Param job body model.JobCreate true "Job Create"
// @Success 201 {object} model.Job
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs [post]
func (j *Jobs) CreateJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		create := &model.JobCreate{}
		if err := ctx.BindJSON(create); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		job, err := j.service.CreateJob(ctx.Request.Context(), create)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, ErrorResponse{Error: jobErr.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, job)

	}
}

// UpdateJob godoc
// @Summary Update a job
// @Description Update a job with the given job update request
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param job body model.JobUpdate true "Job Update"
// @Success 200 {object} model.Job
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs/{id} [put]
func (j *Jobs) UpdateJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		update := model.JobUpdate{}
		if err := ctx.BindJSON(&update); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		job, err := j.service.UpdateJob(ctx.Request.Context(), id, update)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, ErrorResponse{Error: jobErr.Error()})
			return
		}

		ctx.JSON(http.StatusOK, job)

	}
}

// GetJob godoc
// @Summary Get a job
// @Description Get a job with the given job ID
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} model.Job
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs/{id} [get]
func (j *Jobs) GetJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		job, err := j.service.GetJob(ctx.Request.Context(), id)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, ErrorResponse{Error: jobErr.Error()})
			return
		}

		ctx.JSON(http.StatusOK, job)

	}
}

// DeleteJob godoc
// @Summary Delete a job
// @Description Delete a job with the given job ID
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs/{id} [delete]
func (j *Jobs) DeleteJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		if err := j.service.DeleteJob(ctx.Request.Context(), id); err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, ErrorResponse{Error: jobErr.Error()})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

// ListJobs godoc
// @Summary List jobs
// @Description List jobs with the given limit and offset
// @Tags jobs
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} []model.Job
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs [get]
func (j *Jobs) ListJobs() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		limit, offset := LimitAndOffset(ctx)

		jobs, err := j.service.ListJobs(ctx.Request.Context(), limit, offset)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, map[string]interface {
		}{
			"jobs": jobs,
		})

	}
}

// GetJobExecutions godoc
// @Summary Get job executions
// @Description Get job executions with the given job ID, failed only flag, limit and offset
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param failedOnly query bool false "Failed Only"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} []model.JobExecution
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /jobs/{id}/executions [get]
func (j *Jobs) GetJobExecutions() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jobID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		failedOnly, _ := strconv.ParseBool(ctx.Query("failedOnly"))

		limit, offset := LimitAndOffset(ctx)

		executions, err := j.service.GetJobExecutions(ctx.Request.Context(), jobID, failedOnly, limit, offset)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, ErrorResponse{Error: jobErr.Error()})
			return
		}

		ctx.JSON(http.StatusOK, map[string]interface {
		}{
			"executions": executions,
		})
	}
}

func LimitAndOffset(ctx *gin.Context) (uint64, uint64) {
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")

	// convert limit and offset to uint
	var limit, offset uint64
	var err error
	limit, err = strconv.ParseUint(limitStr, 10, 32)
	if err != nil || limit == 0 {
		limit = 10
	}

	offset, err = strconv.ParseUint(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
