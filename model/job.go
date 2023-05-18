package model

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
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
	JobStatusPending    JobStatus = "PENDING"
	JobStatusSuccessful JobStatus = "SUCCESSFUL"
	JobStatusFailed     JobStatus = "FAILED"
)

func (js JobStatus) Valid() bool {
	switch js {
	case JobStatusPending, JobStatusSuccessful, JobStatusFailed:
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

type Job struct {
	ID           uuid.UUID   `json:"id"`
	Type         JobType     `json:"type"`
	Status       JobStatus   `json:"status"`
	ExecuteAt    null.Time   `json:"execute_at"`    // for one-off jobs
	CronSchedule null.String `json:"cron_schedule"` // for recurring jobs

	HTTPJob *HTTPJob `json:"http_job,omitempty"`

	AMQPJob *AMQPJob `json:"amqp_job,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// if last job run failed, the error message will be here
	ErrorMessage null.String `json:"error_message"`
	// when the job is scheduled to run next (can be null if the job is not scheduled to run again)
	NextRun null.Time `json:"next_run"`
}

type HTTPJob struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Auth    AuthMethod        `json:"auth"`
}

type AMQPJob struct {
	Exchange    string            `json:"exchange"`
	RoutingKey  string            `json:"routing_key"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	ContentType string            `json:"content_type"`
	Auth        AuthMethod        `json:"auth"`
}

type AuthMethod struct {
	Type        AuthType    `json:"type"`                   // e.g., "none", "basic", "bearer"
	Username    null.String `json:"username,omitempty"`     // for "basic"
	Password    null.String `json:"password,omitempty"`     // for "basic"
	BearerToken null.String `json:"bearer_token,omitempty"` // for "bearer"
}

// ValidateJob validates a Job struct.
func (job *Job) Validate() error {
	if !job.Type.Valid() {
		return ErrInvalidJobType
	}

	if !job.Status.Valid() {
		return ErrInvalidJobStatus
	}

	if job.Type == JobTypeHTTP {
		if err := job.HTTPJob.Validate(); err != nil {
			return err
		}

		if job.AMQPJob != nil {
			return ErrInvalidJobFields
		}
	}

	if job.Type == JobTypeAMQP {
		if err := job.AMQPJob.Validate(); err != nil {
			return err
		}

		if job.HTTPJob != nil {
			return ErrInvalidJobFields
		}
	}

	// only one of execute_at or cron_schedule can be defined
	if job.ExecuteAt.Valid == job.CronSchedule.Valid {
		return ErrInvalidJobSchedule
	}

	if job.CronSchedule.Valid {
		if _, err := cron.ParseStandard(job.CronSchedule.String); err != nil {
			return ErrInvalidCronSchedule
		}
	}

	if job.ExecuteAt.Valid {
		if job.ExecuteAt.Time.Before(time.Now()) {
			return ErrInvalidExecuteAt
		}
	}

	return nil
}

// ValidateHTTPJob validates an HTTPJob struct.
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

func (job *Job) SetNextRunTime() {
	if job.CronSchedule.Valid {
		schedule, err := cron.ParseStandard(job.CronSchedule.String)
		if err != nil {
			return
		}

		job.NextRun = null.TimeFrom(schedule.Next(time.Now()))
	}

	if job.ExecuteAt.Valid {
		job.NextRun = job.ExecuteAt
	}
}

// ValidateAMQPJob validates an AMQPJob struct.
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

	if err := amqpJob.Auth.Validate(); err != nil {
		return err
	}

	return nil
}

func (auth *AuthMethod) Validate() error {
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
	Type         JobType     `json:"type"`
	ExecuteAt    null.Time   `json:"execute_at"`    // for one-off jobs
	CronSchedule null.String `json:"cron_schedule"` // for recurring jobs

	HTTPJob *HTTPJob `json:"http_job,omitempty"`

	AMQPJob *AMQPJob `json:"amqp_job,omitempty"`
}

func (j *JobCreate) ToJob() *Job {
	return &Job{
		ID:           uuid.New(),
		Type:         j.Type,
		Status:       JobStatusPending,
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		HTTPJob:      j.HTTPJob,
		AMQPJob:      j.AMQPJob,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Execute runs the job.
func (j *Job) Execute(ctx context.Context) error {
	switch j.Type {
	case JobTypeHTTP:
		return j.executeHTTP(ctx)
	case JobTypeAMQP:
		return j.executeAMQP(ctx)
	default:
		return fmt.Errorf("unknown job type: %v", j.Type)
	}
}

// executeHTTP executes an HTTP job.
func (j *Job) executeHTTP(ctx context.Context) error {
	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, j.HTTPJob.Method, j.HTTPJob.URL, strings.NewReader(j.HTTPJob.Body))
	if err != nil {
		return err
	}

	// Set the headers
	for key, value := range j.HTTPJob.Headers {
		req.Header.Set(key, value)
	}

	// Set the auth
	switch j.HTTPJob.Auth.Type {
	case AuthTypeBasic:
		req.SetBasicAuth(j.HTTPJob.Auth.Username.String, j.HTTPJob.Auth.Password.String)
	case AuthTypeBearer:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", j.HTTPJob.Auth.BearerToken.String))
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Close the response body
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	return nil
}

func (j *Job) executeAMQP(ctx context.Context) error {
	// Create a new AMQP connection
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP: %w", err)
	}
	defer conn.Close()

	// Create a new AMQP channel
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Declare an exchange
	err = ch.ExchangeDeclare(
		j.AMQPJob.Exchange, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %w", err)
	}

	// Publish a message to the exchange
	err = ch.PublishWithContext(
		ctx,
		j.AMQPJob.Exchange,   // exchange
		j.AMQPJob.RoutingKey, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: j.AMQPJob.ContentType,
			Body:        []byte(j.AMQPJob.Body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}
