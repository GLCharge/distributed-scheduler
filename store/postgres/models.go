package postgres

import (
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
	ErrorMessage null.String `db:"error_message"`
	NextRun      null.Time   `db:"next_run"`
	LockedAt     null.Time   `db:"locked_at"`
	LockedBy     null.String `db:"locked_by"`
}

func toJobDB(j *model.Job) (*jobDB, error) {
	httpJob, err := json.Marshal(j.HTTPJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal http job")
	}

	amqpJob, err := json.Marshal(j.AMQPJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal amqp job")
	}

	return &jobDB{
		ID:           j.ID,
		Type:         string(j.Type),
		Status:       string(j.Status),
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		HTTPJob:      httpJob,
		AMQPJob:      amqpJob,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		ErrorMessage: j.ErrorMessage,
		NextRun:      j.NextRun,
	}, nil
}

func (j *jobDB) ToJob() (*model.Job, error) {
	var httpJob model.HTTPJob
	if err := json.Unmarshal(j.HTTPJob, &httpJob); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal http job")
	}

	var amqpJob model.AMQPJob
	if err := json.Unmarshal(j.AMQPJob, &amqpJob); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal amqp job")
	}

	return &model.Job{
		ID:           j.ID,
		Type:         model.JobType(j.Type),
		Status:       model.JobStatus(j.Status),
		ExecuteAt:    j.ExecuteAt,
		CronSchedule: j.CronSchedule,
		HTTPJob:      &httpJob,
		AMQPJob:      &amqpJob,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		ErrorMessage: j.ErrorMessage,
		NextRun:      j.NextRun,
	}, nil
}
