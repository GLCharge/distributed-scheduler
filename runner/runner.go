package runner

import (
	"context"
	"sync"
	"time"

	"github.com/GLCharge/distributed-scheduler/executor"
	"github.com/GLCharge/distributed-scheduler/model"
	"go.uber.org/zap"
)

type Runner struct {
	jobService JobService

	executorFactory executor.Factory
	ticker          *time.Ticker
	log             *zap.SugaredLogger

	// Add an instance ID to identify the runner
	instanceId string

	// Add a context and cancel function to stop the runner
	ctx    context.Context
	cancel context.CancelFunc

	// add a wait group to wait for all jobs to finish
	wg sync.WaitGroup

	// Add a wait group to wait for the runner to stop
	stopWg sync.WaitGroup

	// Add a semaphore to limit the number of concurrent jobs
	jobSemaphore chan struct{}

	// Add a sync.Once to ensure the runner only starts once
	startOnce sync.Once

	// limit the number of concurrent jobs
	maxConcurrentJobs int

	// job lock duration
	jobLockDuration time.Duration
}

type JobService interface {
	GetJobsToRun(ctx context.Context, at time.Time, lockedUntil time.Time, instanceID string, limit uint) ([]*model.Job, error)
	FinishJobExecution(ctx context.Context, job *model.Job, startTime, stopTime time.Time, err error) error
}

type Config struct {
	JobService      JobService
	ExecutorFactory executor.Factory
	Log             *zap.SugaredLogger
	InstanceId      string

	Interval          time.Duration
	MaxConcurrentJobs int
	JobLockDuration   time.Duration
}

func New(cfg Config) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Runner{
		jobService:        cfg.JobService,
		instanceId:        cfg.InstanceId,
		log:               cfg.Log,
		ticker:            time.NewTicker(cfg.Interval),
		ctx:               ctx,
		executorFactory:   cfg.ExecutorFactory,
		cancel:            cancel,
		jobSemaphore:      make(chan struct{}, cfg.MaxConcurrentJobs),
		maxConcurrentJobs: cfg.MaxConcurrentJobs,
		jobLockDuration:   cfg.JobLockDuration,
	}

	s.stopWg.Add(1)

	return s
}

// Start is a method to start the runner.
// It is safe to call this method multiple times. Only the first
// call will start the runner. Subsequent calls will be ignored.
func (s *Runner) Start() {
	// Use a sync.Once to ensure the runner only starts once
	s.startOnce.Do(func() {
		s.start()
	})
}

// start is a private method to start the runner
// in a separate goroutine.
// It will run until the runner is stopped.
func (s *Runner) start() {
	// Run the runner in a separate goroutine
	go func() {
		defer s.stopWg.Done() // Signal that the runner has stopped
		defer s.ticker.Stop() // Stop the ticker

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

// Stop is a method to stop the runner, with a context
// to allow for a timeout. if the context has no deadline,
// default to a 10-second timeout.
func (s *Runner) Stop(ctx context.Context) {
	// check if context has a deadline, and if not, create one
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
	}

	// Cancel the runner context to stop the runner
	s.cancel()

	// Wait for the runner to stop, with a timeout
	c := make(chan struct{})
	go func() {
		defer close(c)
		s.stopWg.Wait()
	}()

	select {
	case <-c:
		// The runner stopped
		s.log.Infow("Runner stopped")

	case <-ctx.Done():
		// Timeout
		s.log.Warn("Timeout while stopping the runner")
	}
}

func (s *Runner) runJobs() {
	// Get the current time
	now := time.Now()

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
	defer cancel()

	// Get the jobs that should be run
	jobs, err := s.jobService.GetJobsToRun(ctx, now, now.Add(s.jobLockDuration), s.instanceId, uint(s.maxConcurrentJobs))
	if err != nil {
		// Log the error and return
		s.log.Errorw("Failed to get jobs to run", err)
		return
	}

	s.log.Infow("Running jobs", "count", len(jobs))

	// Run each job
	for _, j := range jobs {
		s.executeJob(j)
	}
}

func (s *Runner) executeJob(job *model.Job) {

	s.jobSemaphore <- struct{}{} // Acquire a slot in the semaphore
	s.wg.Add(1)                  // Increment the wait group counter

	go func() {
		defer s.wg.Done()                   // Decrement the wait group counter
		defer func() { <-s.jobSemaphore }() // Release the semaphore slot

		s.log.Infow("Executing job", "jobID", job.ID)

		// Create a new executor for the job with retry enabled
		jobExecutor, err := s.executorFactory.NewExecutor(job, executor.WithRetry)
		if err != nil {
			s.log.Errorw("Failed to create job executor", "jobID", job.ID, "error", err)
			return
		}

		startTime := time.Now()

		// Execute the job
		err = job.Execute(s.ctx, jobExecutor)

		stopTime := time.Now()

		// Report the job as finished
		err = s.jobService.FinishJobExecution(s.ctx, job, startTime, stopTime, err)
		if err != nil {
			s.log.Errorw("Failed to report job as finished", "jobID", job.ID, "error", err)
		}

		s.log.Infow("Job finished", "jobID", job.ID)
	}()
}
