package scheduler

//
//import (
//	"context"
//	"errors"
//	"testing"
//	"time"
//
//	"github.com/GLCharge/distributed-scheduler/model"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"go.uber.org/zap"
//)
//
//type MockStorer struct {
//	mock.Mock
//}
//
//func (m *MockStorer) GetJobsToRun(ctx context.Context, t time.Time, instanceId string) ([]*model.Job, error) {
//	args := m.Called(ctx, t, instanceId)
//	return args.Get(0).([]*model.Job), args.Error(1)
//}
//
//func TestScheduler_Start_Stop(t *testing.T) {
//	logger, _ := zap.NewDevelopment()
//	sugar := logger.Sugar()
//
//	store := new(MockStorer)
//	scheduler := New(store, sugar, "test-instance")
//
//	// Start the scheduler in a separate goroutine
//	scheduler.Run()
//	time.Sleep(time.Second * 2) // Allow some time for the scheduler to start
//
//	// Stop the scheduler
//	scheduler.Stop()
//
//	// Assert scheduler stops within reasonable time
//	assert.Eventually(t, func() bool {
//		select {
//		case <-scheduler.ctx.Done():
//			return true
//		default:
//			return false
//		}
//	}, time.Second*5, time.Millisecond*100, "Scheduler did not stop in time")
//}
//
//func TestScheduler_runJobs(t *testing.T) {
//	logger, _ := zap.NewDevelopment()
//	sugar := logger.Sugar()
//
//	store := new(MockStorer)
//	scheduler := New(store, sugar, "test-instance")
//
//	jobs := []*model.Job{
//		{ID: "job1", Execute: func(context.Context) error { return nil }},
//		{ID: "job2", Execute: func(context.Context) error { return nil }},
//	}
//	store.On("GetJobsToRun", mock.Anything, mock.Anything, mock.Anything).Return(jobs, nil)
//
//	scheduler.runJobs()
//
//	// Assert all jobs are executed
//	assert.Eventually(t, func() bool { return scheduler.wg.N() == 0 }, time.Second*5, time.Millisecond*100, "Not all jobs completed in time")
//
//	// Test with an error
//	store = new(MockStorer)
//	scheduler = New(store, sugar, "test-instance")
//
//	store.On("GetJobsToRun", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test error"))
//
//	scheduler.runJobs()
//
//	// Assert no jobs are run
//	assert.Equal(t, 0, scheduler.wg.N())
//}
