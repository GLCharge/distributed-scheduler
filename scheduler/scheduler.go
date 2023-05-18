package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"go.uber.org/zap"

	"github.com/GLCharge/distributed-scheduler/store"
)

const maxConcurrentJobs = 100

type Scheduler struct {
	store  store.Storer
	ticker *time.Ticker
	log    *zap.SugaredLogger

	// Add an instance ID to identify the scheduler
	instanceId string

	// Add a context and cancel function to stop the scheduler
	ctx    context.Context
	cancel context.CancelFunc

	// add a wait group to wait for all jobs to finish
	wg sync.WaitGroup

	// Add a wait group to wait for the scheduler to stop
	stopWg sync.WaitGroup

	// Add a semaphore to limit the number of concurrent jobs
	jobSemaphore chan struct{}
}

func New(store store.Storer, log *zap.SugaredLogger, instanceId string) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		store:        store,
		instanceId:   instanceId,
		log:          log,
		ticker:       time.NewTicker(time.Second * 10),
		ctx:          ctx,
		cancel:       cancel,
		jobSemaphore: make(chan struct{}, maxConcurrentJobs),
	}

	s.stopWg.Add(1)

	return s
}

func (s *Scheduler) Run() {
	// Run the scheduler in a separate goroutine
	go func() {
		defer s.ticker.Stop()
		defer s.stopWg.Done() // Signal that the scheduler has stopped

		for {
			select {
			case <-s.ticker.C:
				s.runJobs()
			case <-s.ctx.Done():
				s.wg.Wait() // Wait for all jobs to finish
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	// Cancel the context to stop the scheduler
	s.cancel()

	// Wait for the scheduler to stop, with a timeout
	c := make(chan struct{})
	go func() {
		defer close(c)
		s.stopWg.Wait()
	}()

	select {
	case <-c:

		s.log.Infow("Scheduler stopped")
		// The scheduler stopped
	case <-time.After(time.Second * 10):
		// Timeout
		s.log.Warn("Timeout while stopping the scheduler")
	}
}

func (s *Scheduler) runJobs() {
	// Get the current time
	now := time.Now()

	// Get the jobs that should be run
	jobs, err := s.store.GetJobsToRun(s.ctx, now, s.instanceId)
	if err != nil {
		// Log the error and return
		s.log.Error("Failed to get jobs to run", err)
		return
	}

	s.log.Infow("Running jobs", "count", len(jobs))

	// Run each job
	for _, job := range jobs {
		s.jobSemaphore <- struct{}{} // Acquire a slot in the semaphore
		s.wg.Add(1)

		go func(job *model.Job) {
			defer s.wg.Done()
			defer func() { <-s.jobSemaphore }() // Release the semaphore slot

			s.log.Infow("Executing job", "jobID", job.ID)

			if err := job.Execute(s.ctx); err != nil {
				s.log.Errorw("Job execution failed", "jobID", job.ID, "error", err)
			}
		}(job)
	}
}
