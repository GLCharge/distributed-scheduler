package executor

import (
	"context"
	"errors"
	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockExecutor for testing
type MockExecutor struct {
	CallCount    int
	ShouldFail   bool
	FailuresLeft int
}

func (me *MockExecutor) Execute(ctx context.Context, j *model.Job) error {
	me.CallCount++
	if me.ShouldFail && (me.FailuresLeft > 0) {
		me.FailuresLeft--
		return errors.New("execute error")
	}
	return nil
}

func TestRetryExecutor_Execute(t *testing.T) {
	t.Parallel()

	// Mocking a Job
	j := &model.Job{
		Type: model.JobTypeHTTP,
	}

	// Creating a mock executor that will fail twice before succeeding
	mockExec := &MockExecutor{
		ShouldFail:   true,
		FailuresLeft: 2,
	}

	// Creating the retry executor with the mock executor
	re := WithRetry(mockExec)

	// Execute
	err := re.Execute(context.Background(), j)

	// We expect that after 3 attempts, the function will succeed
	assert.Equal(t, 3, mockExec.CallCount)
	assert.Nil(t, err)

	// Now we test when the retries are exceeded
	// Creating a mock executor that will always fail
	mockExec = &MockExecutor{
		ShouldFail:   true,
		FailuresLeft: 10,
	}

	re = WithRetry(mockExec)

	err = re.Execute(context.Background(), j)

	// We expect that after 4 attempts, the function will give up and return an error
	assert.Equal(t, 4, mockExec.CallCount)
	assert.Error(t, err)
}
