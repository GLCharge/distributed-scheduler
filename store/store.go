package store

import (
	"context"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
)

type Storer interface {
	// CRUD operations for jobs
	CreateJob(ctx context.Context, job *model.Job) error
	GetJob(ctx context.Context, id string) (*model.Job, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, limit, offset uint64) ([]*model.Job, error)
	UpdateJob(ctx context.Context, job *model.Job) error

	// Get jobs to run
	GetJobsToRun(ctx context.Context, t time.Time, instanceID string) ([]*model.Job, error)
	FinishJob(ctx context.Context, jobID uuid.UUID, nextRun null.Time) error
	CreateJobExecution(ctx context.Context, jobID uuid.UUID, startTime, stopTime time.Time, status model.JobExecutionStatus, errorMessage null.String) error

	GetJobExecutions(ctx context.Context, jobID uuid.UUID, failedOnly bool, limit, offset uint64) ([]*model.JobExecution, error)
}
