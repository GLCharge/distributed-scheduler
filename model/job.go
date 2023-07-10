package model

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type JobType string

// JobType is the type of job. Currently, only HTTP and AMQP jobs are supported.
const (
	JobTypeHTTP JobType = "HTTP"
	JobTypeAMQP JobType = "AMQP"
)

func (jt JobType) Valid() bool {
	switch jt {
	case JobTypeHTTP, JobTypeAMQP:
		return true
	default:
		return false
	}
}

type JobStatus string

const (
	JobStatusRunning JobStatus = "RUNNING"
	JobStatusStopped JobStatus = "STOPPED"
)

func (js JobStatus) Valid() bool {
	switch js {
	case JobStatusStopped, JobStatusRunning:
		return true
	default:
		return false
	}
}

type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeBearer AuthType = "bearer"
)

func (at AuthType) Valid() bool {
	switch at {
	case AuthTypeNone, AuthTypeBasic, AuthTypeBearer:
		return true
	default:
		return false
	}
}

type BodyEncoding string

const (
	BodyEncodingBase64 BodyEncoding = "base64"
)

func (be *BodyEncoding) Valid() bool {
	if be == nil {
		return true
	}
	switch *be {
	case BodyEncodingBase64:
		return true
	default:
		return false
	}
}

// swagger:model Job
type Job struct {
	ID     uuid.UUID `json:"id"`
	Type   JobType   `json:"type"`
	Status JobStatus `json:"status"`

	ExecuteAt    null.Time   `json:"execute_at" swaggertype:"string"`    // for one-off jobs
	CronSchedule null.String `json:"cron_schedule" swaggertype:"string"` // for recurring jobs

	HTTPJob *HTTPJob `json:"http_job,omitempty"`

	AMQPJob *AMQPJob `json:"amqp_job,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// when the job is scheduled to run next (can be null if the job is not scheduled to run again)
	NextRun null.Time `json:"next_run"`
}

// swagger:model JobUpdate
type JobUpdate struct {
	Type *JobType `json:"type,omitempty"`
	HTTP *HTTPJob `json:"http,omitempty"`
	AMQP *AMQPJob `json:"amqp,omitempty"`

	CronSchedule *string    `json:"cron_schedule,omitempty"`
	ExecuteAt    *time.Time `json:"execute_at,omitempty"`
}

func (j *Job) ApplyUpdate(update JobUpdate) {

	if update.Type != nil {
		j.Type = *update.Type
	}

	if update.HTTP != nil {
		j.HTTPJob = update.HTTP
		j.AMQPJob = nil
	}

	if update.AMQP != nil {
		j.AMQPJob = update.AMQP
		j.HTTPJob = nil
	}

	if update.CronSchedule != nil {
		j.CronSchedule = null.StringFromPtr(update.CronSchedule)
	}

	if update.ExecuteAt != nil {
		j.ExecuteAt = null.TimeFromPtr(update.ExecuteAt)
	}

	j.UpdatedAt = time.Now()

	j.SetInitialRunTime()
}

type HTTPJob struct {
	URL                string            `json:"url"`                       // e.g., "https://example.com"
	Method             string            `json:"method"`                    // e.g., "GET", "POST", "PUT", "PATCH", "DELETE"
	Headers            map[string]string `json:"headers"`                   // e.g., {"Content-Type": "application/json"}
	Body               null.String       `json:"body" swaggertype:"string"` // e.g., "{\"hello\": \"world\"}"
	ValidResponseCodes []int             `json:"valid_response_codes"`      // e.g., [200, 201, 202]
	Auth               Auth              `json:"auth"`                      // e.g., {"type": "basic", "username": "foo", "password": "bar"}
}

type AMQPJob struct {
	Connection   string                 `json:"connection"`    // e.g., "amqp://guest:guest@localhost:5672/"
	Exchange     string                 `json:"exchange"`      // e.g., "my_exchange"
	RoutingKey   string                 `json:"routing_key"`   // e.g., "my_routing_key"
	Headers      map[string]interface{} `json:"headers"`       // e.g., {"x-delay": 10000}
	Body         string                 `json:"body"`          // e.g., "Hello, world!"
	BodyEncoding *BodyEncoding          `json:"body_encoding"` // e.g., null, "base64"
	ContentType  string                 `json:"content_type"`  // e.g., "text/plain"
}

type Auth struct {
	Type        AuthType    `json:"type"`                                        // e.g., "none", "basic", "bearer"
	Username    null.String `json:"username,omitempty" swaggertype:"string"`     // for "basic"
	Password    null.String `json:"password,omitempty" swaggertype:"string"`     // for "basic"
	BearerToken null.String `json:"bearer_token,omitempty" swaggertype:"string"` // for "bearer"
}

// Validate validates a Job struct.
func (j *Job) Validate() error {
	if j.ID == uuid.Nil {
		return ErrInvalidJobID
	}

	if !j.Type.Valid() {
		return ErrInvalidJobType
	}

	if !j.Status.Valid() {
		return ErrInvalidJobStatus
	}

	if j.Type == JobTypeHTTP {
		if err := j.HTTPJob.Validate(); err != nil {
			return err
		}

		if j.AMQPJob != nil {
			return ErrInvalidJobFields
		}
	}

	if j.Type == JobTypeAMQP {
		if err := j.AMQPJob.Validate(); err != nil {
			return err
		}

		if j.HTTPJob != nil {
			return ErrInvalidJobFields
		}
	}

	// only one of execute_at or cron_schedule can be defined
	if j.ExecuteAt.Valid == j.CronSchedule.Valid {
		return ErrInvalidJobSchedule
	}

	if j.CronSchedule.Valid {
		if _, err := cron.ParseStandard(j.CronSchedule.String); err != nil {
			return ErrInvalidCronSchedule
		}
		cron.NewChain()
	}

	if j.ExecuteAt.Valid {
		if j.ExecuteAt.Time.Before(time.Now()) {
			return ErrInvalidExecuteAt
		}
	}

	return nil
}

// Validate validates an HTTPJob struct.
func (httpJob *HTTPJob) Validate() error {
	if httpJob == nil {
		return ErrHTTPJobNotDefined
	}

	if httpJob.URL == "" {
		return ErrEmptyHTTPJobURL
	}

	if httpJob.Method == "" {
		return ErrEmptyHTTPJobMethod
	}

	if err := httpJob.Auth.Validate(); err != nil {
		return err
	}

	return nil
}

func (j *Job) SetNextRunTime() {
	// if the job is a recurring job, set NextRun to the next time the job should run
	if j.CronSchedule.Valid {
		schedule, err := cron.ParseStandard(j.CronSchedule.String)
		if err != nil {
			return
		}

		j.NextRun = null.TimeFrom(schedule.Next(time.Now()))
	}

	// if the job is a one-off job, set NextRun to null
	if j.ExecuteAt.Valid {
		j.NextRun = null.Time{}
	}

	j.UpdatedAt = time.Now()
}

func (j *Job) SetInitialRunTime() {
	if j.CronSchedule.Valid {
		schedule, err := cron.ParseStandard(j.CronSchedule.String)
		if err != nil {
			return
		}

		j.NextRun = null.TimeFrom(schedule.Next(time.Now()))
	}

	if j.ExecuteAt.Valid {
		j.NextRun = null.TimeFrom(j.ExecuteAt.Time)
	}
}

// Validate validates an AMQPJob struct.
func (amqpJob *AMQPJob) Validate() error {
	if amqpJob == nil {
		return ErrAMQPJobNotDefined
	}

	if amqpJob.Exchange == "" {
		return ErrEmptyExchange
	}

	if amqpJob.RoutingKey == "" {
		return ErrEmptyRoutingKey
	}

	if !amqpJob.BodyEncoding.Valid() {
		return ErrInvalidBodyEncoding
	}

	return nil
}

func (auth *Auth) Validate() error {
	if auth == nil {
		return ErrAuthMethodNotDefined
	}

	if !auth.Type.Valid() {
		return ErrInvalidAuthType
	}

	if auth.Type == AuthTypeBasic {
		if !auth.Username.Valid || auth.Username.String == "" {
			return ErrEmptyUsername
		}

		if !auth.Password.Valid || auth.Password.String == "" {
			return ErrEmptyPassword
		}
	}

	if auth.Type == AuthTypeBearer && (!auth.BearerToken.Valid || auth.BearerToken.String == "") {
		return ErrEmptyBearerToken
	}

	return nil
}

type JobCreate struct {

	// Job type
	Type JobType `json:"type"`

	// ExecuteAt and CronSchedule are mutually exclusive.
	ExecuteAt    null.Time   `json:"execute_at" swaggertype:"string"`    // for one-off jobs
	CronSchedule null.String `json:"cron_schedule" swaggertype:"string"` // for recurring jobs

	// HTTPJob and AMQPJob are mutually exclusive.
	HTTPJob *HTTPJob `json:"http_job,omitempty"`
	AMQPJob *AMQPJob `json:"amqp_job,omitempty"`
}

func (j *JobCreate) ToJob() *Job {
	job := &Job{
		ID:           uuid.New(),
		Type:         j.Type,
		Status:       JobStatusRunning,
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		HTTPJob:      j.HTTPJob,
		AMQPJob:      j.AMQPJob,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	job.SetInitialRunTime()

	return job
}

func (j *Job) Execute(ctx context.Context, executor Executor) error {
	return executor.Execute(ctx, j)
}

type Executor interface {
	Execute(ctx context.Context, job *Job) error
}
