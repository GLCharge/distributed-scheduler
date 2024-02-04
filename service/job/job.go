package job

import (
	"context"
	"github.com/GLCharge/otelzap"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/store"
)

// Service is a struct that contains a store and a logger.
type Service struct {
	store store.Storer
	log   *otelzap.Logger
}

// NewService creates a new job service with the given store and logger.
func NewService(store store.Storer, log *otelzap.Logger) *Service {
	return &Service{
		store: store,
		log:   log,
	}
}

// CreateJob creates a new job using the given job create request and returns the created job.
// If the job create request is invalid, an error is returned.
func (s *Service) CreateJob(ctx context.Context, jobCreate *model.JobCreate) (*model.Job, error) {

	// Convert the job create request to a job
	job := jobCreate.ToJob()

	// Validate the job
	if err := job.Validate(); err != nil {
		return nil, err
	}

	// Create the job using the store
	err := s.store.CreateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetJob returns the job with the given ID.
func (s *Service) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	return s.store.GetJob(ctx, id)
}

// UpdateJob updates the given job.
func (s *Service) UpdateJob(ctx context.Context, jobID uuid.UUID, jobUpdate model.JobUpdate) (*model.Job, error) {
	// get the job from the store
	job, err := s.store.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// update the job
	job.ApplyUpdate(jobUpdate)

	// validate the job
	if err := job.Validate(); err != nil {
		return nil, err
	}

	// update the job in the store
	err = s.store.UpdateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// DeleteJob deletes the job with the given ID.
func (s *Service) DeleteJob(ctx context.Context, id uuid.UUID) error {
	// Implement deleting a specific job using the store

	return s.store.DeleteJob(ctx, id)
}

// ListJobs returns a list of jobs with the given limit and offset.
func (s *Service) ListJobs(ctx context.Context, limit, offset uint64, tags []string) ([]model.Job, error) {
	// Implement listing jobs using the store

	return s.store.ListJobs(ctx, limit, offset, tags)
}

// GetJobsToRun returns a list of jobs that should be run at the given time.
func (s *Service) GetJobsToRun(ctx context.Context, at time.Time, lockedUntil time.Time, instanceID string, limit uint) ([]*model.Job, error) {

	return s.store.GetJobsToRun(ctx, at, lockedUntil, instanceID, limit)
}

func (s *Service) FinishJobExecution(ctx context.Context, job *model.Job, startTime, stopTime time.Time, err error) error {

	// Update the job execution
	job.SetNextRunTime()

	// finish the job in the store (update the next run time and clear lock)
	err2 := s.store.FinishJob(ctx, job.ID, job.NextRun)
	if err2 != nil {
		return err
	}

	jobExecutionStatus := model.JobExecutionStatusSuccessful
	errorMessage := null.String{}
	if err != nil {
		jobExecutionStatus = model.JobExecutionStatusFailed
		errorMessage = null.StringFrom(err.Error())
	}

	// Create the job execution
	err2 = s.store.CreateJobExecution(ctx, job.ID, startTime, stopTime, jobExecutionStatus, errorMessage)
	if err2 != nil {
		return err2
	}

	return nil
}

func (s *Service) GetJobExecutions(ctx context.Context, id uuid.UUID, failedOnly bool, limit uint64, offset uint64) ([]*model.JobExecution, error) {

	return s.store.GetJobExecutions(ctx, id, failedOnly, limit, offset)

}
