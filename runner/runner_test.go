package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	s := createRunnerWithMockExecutor(time.Second, 1, nil, nil, nil, nil)
	if s == nil {
		t.Fatal("Expected scheduler to be created, but got nil")
	}
	if s.jobService == nil {
		t.Error("Expected jobService to be initialized, but got nil")
	}
	if s.executorFactory == nil {
		t.Error("Expected executorFactory to be initialized, but got nil")
	}
	if s.log == nil {
		t.Error("Expected log to be initialized, but got nil")
	}
	if s.ticker == nil {
		t.Error("Expected ticker to be initialized, but got nil")
	}
	if s.ctx == nil {
		t.Error("Expected ctx to be initialized, but got nil")
	}
	if s.instanceId == "" {
		t.Error("Expected instanceId to be initialized, but got empty")
	}
	if s.jobSemaphore == nil {
		t.Error("Expected jobSemaphore to be initialized, but got nil")
	}
}

func TestStart(t *testing.T) {

	s := createRunnerWithMockExecutor(time.Second, 1, nil, nil, nil, nil)

	// Ensure startOnce is locked before calling Start
	if s.startOnce != (sync.Once{}) {
		t.Error("Expected startOnce to be locked before calling Start")
	}

	s.Start()
	s.Start() // Second start should be ignored

	// Sleep for a moment to ensure the scheduler's goroutine starts
	time.Sleep(time.Millisecond * 100)

	// Ensure startOnce is unlocked after calling Start
	if s.startOnce == (sync.Once{}) {
		t.Error("Expected startOnce to be unlocked after calling Start")
	}
}

func TestStop(t *testing.T) {

	s := createRunnerWithMockExecutor(time.Millisecond, 1, nil, nil, nil, nil)

	s.Start()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	s.Stop(ctx)

	select {
	case <-s.ctx.Done():
		// Context should be cancelled after calling Stop
	default:
		t.Error("Expected ctx to be cancelled after calling Stop")
	}
}

func TestRunJobs(t *testing.T) {

	// Test the happy path where GetJobsToRun and FinishJobExecution succeed
	s := createRunnerWithMockExecutor(time.Millisecond*50, 1, nil, nil, nil, nil)
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that all the jobs have been processed
	assertJobsProcessed(t, s.jobService.(*mockJobService))

	// Test the sad path where GetJobsToRun returns an error
	s = createRunnerWithMockExecutor(time.Millisecond*50, 1, errors.New("get jobs error"), nil, nil, nil)
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that no jobs have been processed
	if len(s.jobService.(*mockJobService).Jobs) != 3 {
		t.Errorf("Expected no jobs to have been processed, but got %d", len(s.jobService.(*mockJobService).Jobs))
	}

	// Test the sad path where FinishJobExecution returns an error
	s = createRunnerWithMockExecutor(time.Millisecond*50, 1, nil, errors.New("finish job error"), nil, nil)
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that all jobs are still present
	if len(s.jobService.(*mockJobService).Jobs) != 3 {
		t.Errorf("Expected all jobs to be still present, but got %d", len(s.jobService.(*mockJobService).Jobs))
	}
}

func TestExecuteJob(t *testing.T) {

	// Test the happy path where NewExecutor and Execute succeed
	s := createRunnerWithMockExecutor(time.Millisecond*50, 1, nil, nil, nil, nil)
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that all the jobs have been processed
	assertJobsProcessed(t, s.jobService.(*mockJobService))

	// Test the sad path where NewExecutor returns an error
	s = createRunnerWithMockExecutor(time.Millisecond*50, 1, nil, nil, errors.New("new executor error"), nil)
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that no jobs have been processed
	if len(s.jobService.(*mockJobService).Jobs) != 3 {
		t.Errorf("Expected no jobs to have been processed, but got %d", len(s.jobService.(*mockJobService).Jobs))
	}

	// Test the sad path where Execute returns an error
	s = createRunnerWithMockExecutor(time.Millisecond*50, 1, nil, nil, nil, errors.New("execute error"))
	s.Start()

	// Sleep for a moment to allow the scheduler to run jobs
	time.Sleep(time.Millisecond * 200)

	// Stop the scheduler
	s.Stop(context.Background())

	// Check that no jobs have been processed
	if len(s.jobService.(*mockJobService).Jobs) != 0 {
		t.Errorf("Expected all jobs to have been processed, but got %d", len(s.jobService.(*mockJobService).Jobs))
	}
}
