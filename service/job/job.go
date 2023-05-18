package job

import (
	"context"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/store"
	"go.uber.org/zap"
)

type Service struct {
	store store.Storer
	log   *zap.SugaredLogger
}

func NewService(store store.Storer, log *zap.SugaredLogger) *Service {
	return &Service{
		store: store,
		log:   log,
	}
}

func (s *Service) CreateJob(ctx context.Context, jobCreate *model.JobCreate) error {
	// Implement the creation of a job using the store

	job := jobCreate.ToJob()

	if err := job.Validate(); err != nil {
		return err
	}

	s.log.Infow("Creating job", "job", job)

	err := s.store.CreateJob(ctx, job)
	if err != nil {
		return err
	}

	return nil

}

func (s *Service) GetJob(ctx context.Context, id string) (*model.Job, error) {
	// Implement getting a specific job using the store

	return s.store.GetJob(ctx, id)
}

func (s *Service) UpdateJob(ctx context.Context, job *model.Job) error {
	// Implement updating a specific job using the store

	return nil
}

func (s *Service) DeleteJob(ctx context.Context, id string) error {

	return s.store.DeleteJob(ctx, id)
}

func (s *Service) ListJobs(ctx context.Context, limit, offset uint64) ([]*model.Job, error) {

	return s.store.ListJobs(ctx, limit, offset)
}
