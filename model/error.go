package model

import "errors"

var (
	ErrInvalidJobType       = errors.New("job type must be either HTTP or AMQP")
	ErrInvalidJobStatus     = errors.New("job status must be either PENDING, SCHEDULED, SUCCESSFUL, or FAILED")
	ErrInvalidJobFields     = errors.New("job cannot have both HTTP and AMQP fields defined")
	ErrInvalidJobSchedule   = errors.New("job must have only one of execute_at and cron_schedule defined")
	ErrInvalidCronSchedule  = errors.New("invalid cron schedule")
	ErrInvalidExecuteAt     = errors.New("execute_at must be in the future")
	ErrEmptyHTTPJobURL      = errors.New("HTTP job URL cannot be empty")
	ErrHTTPJobNotDefined    = errors.New("HTTP job must be defined")
	ErrEmptyHTTPJobMethod   = errors.New("HTTP job method cannot be empty")
	ErrAMQPJobNotDefined    = errors.New("AMQP job must be defined")
	ErrEmptyExchange        = errors.New("exchange must be defined for AMQP jobs")
	ErrEmptyRoutingKey      = errors.New("routing key must be defined for AMQP jobs")
	ErrInvalidAuthType      = errors.New("auth type must be either none, basic, or bearer")
	ErrEmptyUsername        = errors.New("username must be defined for basic auth")
	ErrEmptyPassword        = errors.New("password must be defined for basic auth")
	ErrEmptyBearerToken     = errors.New("bearer token must be defined for bearer auth")
	ErrAuthMethodNotDefined = errors.New("auth method must be defined")
)

type CustomError struct {
	Err  error
	Code int
}

func (e *CustomError) Error() string {
	return e.Err.Error()
}

func ToCustomJobError(err error) *CustomError {
	switch err {
	case ErrInvalidJobType, ErrInvalidJobStatus, ErrInvalidJobFields, ErrInvalidJobSchedule, ErrInvalidCronSchedule, ErrInvalidExecuteAt,
		ErrEmptyHTTPJobURL, ErrHTTPJobNotDefined, ErrEmptyHTTPJobMethod, ErrAMQPJobNotDefined, ErrEmptyExchange, ErrEmptyRoutingKey,
		ErrInvalidAuthType, ErrEmptyUsername, ErrEmptyPassword, ErrEmptyBearerToken, ErrAuthMethodNotDefined:
		return &CustomError{err, 400}

	default:
		return &CustomError{err, 500}
	}
}
