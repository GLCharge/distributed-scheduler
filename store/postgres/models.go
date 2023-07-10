package postgres

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
)

type jobDB struct {
	ID           uuid.UUID   `db:"id"`
	Type         string      `db:"type"`
	Status       string      `db:"status"`
	ExecuteAt    null.Time   `db:"execute_at"`
	CronSchedule null.String `db:"cron_schedule"`
	HTTPJob      []byte      `db:"http_job"`
	AMQPJob      []byte      `db:"amqp_job"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
	NextRun      null.Time   `db:"next_run"`
	LockedUntil  null.Time   `db:"locked_until"`
	LockedBy     null.String `db:"locked_by"`
}

func toJobDB(j *model.Job) (*jobDB, error) {

	dbJ := &jobDB{
		ID:           j.ID,
		Type:         string(j.Type),
		Status:       string(j.Status),
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		NextRun:      j.NextRun,
	}

	if j.HTTPJob != nil {
		httpJob, err := json.Marshal(j.HTTPJob)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal http job")
		}
		dbJ.HTTPJob = bytes.Trim(httpJob, "\x00")
	}

	if j.AMQPJob != nil {
		amqpJob, err := json.Marshal(j.AMQPJob)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal amqp job")
		}
		dbJ.AMQPJob = bytes.Trim(amqpJob, "\x00")
	}

	return dbJ, nil
}

func (j *jobDB) ToJob() (*model.Job, error) {
	job := &model.Job{
		ID:           j.ID,
		Type:         model.JobType(j.Type),
		Status:       model.JobStatus(j.Status),
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		NextRun:      j.NextRun,
	}

	if err := unmarshalNullableJSON(j.HTTPJob, &job.HTTPJob); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal http job")
	}

	if err := unmarshalNullableJSON(j.AMQPJob, &job.AMQPJob); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal amqp job")
	}

	return job, nil
}

func unmarshalNullableJSON(data []byte, v interface{}) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}

type executionDB struct {
	ID           int         `db:"id"`
	JobID        uuid.UUID   `db:"job_id"`
	Status       string      `db:"status"`
	StartTime    time.Time   `db:"start_time"`
	EndTime      time.Time   `db:"end_time"`
	ErrorMessage null.String `db:"error_message"`
	CreatedAt    time.Time   `db:"created_at"`
}

func (e *executionDB) ToModel() *model.JobExecution {

	return &model.JobExecution{
		ID:           e.ID,
		JobID:        e.JobID,
		Success:      e.Status == string(model.JobExecutionStatusSuccessful),
		StartTime:    e.StartTime,
		EndTime:      e.EndTime,
		ErrorMessage: e.ErrorMessage,
	}
}
