package model

import (
	"context"
	"fmt"
	"io"
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

type Job struct {
	ID     uuid.UUID `json:"id"`
	Type   JobType   `json:"type"`
	Status JobStatus `json:"status"`

	ExecuteAt    null.Time   `json:"execute_at"`    // for one-off jobs
	CronSchedule null.String `json:"cron_schedule"` // for recurring jobs

	HTTPJob *HTTPJob `json:"http_job,omitempty"`

	AMQPJob *AMQPJob `json:"amqp_job,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// when the job is scheduled to run next (can be null if the job is not scheduled to run again)
	NextRun null.Time `json:"next_run"`
}

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

	j.SetNextRunTime()
}

type HTTPJob struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    null.String       `json:"body"`
	Auth    AuthMethod        `json:"auth"`
}

type AMQPJob struct {
	Connection   string         `json:"connection"`    // e.g., "amqp://guest:guest@localhost:5672/"
	Exchange     string         `json:"exchange"`      // e.g., "my_exchange"
	ExchangeType string         `json:"exchange_type"` // e.g., "direct"
	RoutingKey   string         `json:"routing_key"`   // e.g., "my_routing_key"
	Headers      map[string]any `json:"headers"`
	Body         string         `json:"body"`
	ContentType  string         `json:"content_type"`
	AutoDelete   bool           `json:"auto_delete"`
	Internal     bool           `json:"internal"`
	Durable      bool           `json:"durable"`
	NoWait       bool           `json:"no_wait"`
}

type AuthMethod struct {
	Type        AuthType    `json:"type"`                   // e.g., "none", "basic", "bearer"
	Username    null.String `json:"username,omitempty"`     // for "basic"
	Password    null.String `json:"password,omitempty"`     // for "basic"
	BearerToken null.String `json:"bearer_token,omitempty"` // for "bearer"
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
		if j.ExecuteAt.Time.Before(time.Now()) {
			j.NextRun = null.Time{}
		} else {
			j.NextRun = j.ExecuteAt
		}
	}

	j.UpdatedAt = time.Now()
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

	job.SetNextRunTime()

	return job
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
	// Create the HTTP request
	req, err := j.createHTTPRequest(ctx)
	if err != nil {
		return err
	}

	// Send the request and get the response
	resp, err := j.sendHTTPRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	return nil
}

// createHTTPRequest creates an HTTP request for the job.
func (j *Job) createHTTPRequest(ctx context.Context) (*http.Request, error) {
	// Create the request body
	body := j.createHTTPRequestBody()

	// Create the request URL
	urlStr := j.createHTTPRequestURL()

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, j.HTTPJob.Method, urlStr, body)
	if err != nil {
		return nil, err
	}

	// Set the headers
	j.setHTTPRequestHeaders(req)

	// Set the auth
	j.setHTTPRequestAuth(req)

	return req, nil
}

// createHTTPRequestBody creates the request body for the job.
func (j *Job) createHTTPRequestBody() io.Reader {
	if !j.HTTPJob.Body.Valid || j.HTTPJob.Body.String == "" {
		return nil
	}

	return strings.NewReader(j.HTTPJob.Body.String)
}

// createHTTPRequestURL creates the request URL for the job.
func (j *Job) createHTTPRequestURL() string {
	if strings.HasPrefix(j.HTTPJob.URL, "http://") || strings.HasPrefix(j.HTTPJob.URL, "https://") {
		return j.HTTPJob.URL
	}

	return "https://" + j.HTTPJob.URL
}

// setHTTPRequestHeaders sets the headers for the HTTP request.
func (j *Job) setHTTPRequestHeaders(req *http.Request) {
	for key, value := range j.HTTPJob.Headers {
		req.Header.Set(key, value)
	}
}

// setHTTPRequestAuth sets the auth for the HTTP request.
func (j *Job) setHTTPRequestAuth(req *http.Request) {
	switch j.HTTPJob.Auth.Type {
	case AuthTypeBasic:
		req.SetBasicAuth(j.HTTPJob.Auth.Username.String, j.HTTPJob.Auth.Password.String)
	case AuthTypeBearer:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", j.HTTPJob.Auth.BearerToken.String))
	}
}

// sendHTTPRequest sends an HTTP request and returns the response.
func (j *Job) sendHTTPRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (j *Job) executeAMQP(ctx context.Context) error {
	// Create a new AMQP connection
	conn, err := amqp.Dial(j.AMQPJob.Connection)
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
		j.AMQPJob.Exchange,     // name
		j.AMQPJob.ExchangeType, // type
		j.AMQPJob.Durable,      // durable
		j.AMQPJob.AutoDelete,   // auto-deleted
		j.AMQPJob.Internal,     // internal
		j.AMQPJob.NoWait,       // no-wait
		nil,                    // arguments
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
			Headers:     j.AMQPJob.Headers,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
