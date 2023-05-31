package jobs

import (
	"github.com/google/uuid"
	"net/http"
	"strconv"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/gin-gonic/gin"
)

func NewJobsHandler(service *job.Service) *Jobs {
	return &Jobs{
		service: service,
	}
}

type Jobs struct {
	service *job.Service
}

func (j *Jobs) CreateJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		create := &model.JobCreate{}
		if err := ctx.BindJSON(create); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		job, err := j.service.CreateJob(ctx.Request.Context(), create)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, gin.H{
				"error": jobErr.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, job)

	}
}

func (j *Jobs) UpdateJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		update := model.JobUpdate{}
		if err := ctx.BindJSON(&update); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		job, err := j.service.UpdateJob(ctx.Request.Context(), id, update)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, gin.H{
				"error": jobErr.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, job)

	}
}

func (j *Jobs) GetJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		job, err := j.service.GetJob(ctx.Request.Context(), id)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, gin.H{
				"error": jobErr.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, job)

	}
}

func (j *Jobs) DeleteJob() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		if err := j.service.DeleteJob(ctx.Request.Context(), id); err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, gin.H{
				"error": jobErr.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (j *Jobs) ListJobs() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		limit, offset := LimitAndOffset(ctx)

		jobs, err := j.service.ListJobs(ctx.Request.Context(), limit, offset)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, map[string]interface {
		}{
			"jobs": jobs,
		})

	}
}

func (j *Jobs) GetJobExecutions() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		jobID, err := uuid.Parse(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		failedOnlyStr := ctx.Query("failedOnly")
		failedOnly, err := strconv.ParseBool(failedOnlyStr)
		if err != nil {
			failedOnly = false
		}

		limit, offset := LimitAndOffset(ctx)

		executions, err := j.service.GetJobExecutions(ctx.Request.Context(), jobID, failedOnly, limit, offset)
		if err != nil {
			jobErr := model.ToCustomJobError(err)

			ctx.JSON(jobErr.Code, gin.H{
				"error": jobErr.Error(),
			})
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
