package executor

import (
	"fmt"
	"github.com/GLCharge/distributed-scheduler/model"
)

type Factory struct {
	client HttpClient
}

func NewFactory(client HttpClient) *Factory {
	return &Factory{
		client: client,
	}
}

// Option is a function that modifies an executor before it is returned (e.g. WithRetry)
type Option func(executor model.Executor) model.Executor

func (f *Factory) NewExecutor(job *model.Job, options ...Option) (model.Executor, error) {

	var executor model.Executor
	switch job.Type {
	case model.JobTypeHTTP:
		executor = &hTTPExecutor{Client: f.client}
	case model.JobTypeAMQP:
		executor = &aMQPExecutor{}
	default:
		return nil, fmt.Errorf("unknown job type: %v", job.Type)
	}

	for _, option := range options {
		executor = option(executor)
	}

	return executor, nil
}
