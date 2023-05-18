package store

import (
	"context"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
)

type Storer interface {
	// CRUD operations for jobs
	CreateJob(ctx context.Context, job *model.Job) error
	GetJob(ctx context.Context, id string) (*model.Job, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, limit, offset uint64) ([]*model.Job, error)

	// Get jobs to run
	GetJobsToRun(ctx context.Context, t time.Time, instanceID string) ([]*model.Job, error)
}
