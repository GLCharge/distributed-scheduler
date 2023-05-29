package postgres

import (
	"testing"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestJobDB_ToJob_AMQPJobNull(t *testing.T) {
	jobDB := &jobDB{
		ID:           uuid.MustParse("a787fa30-2cbe-40de-9a51-f7c9fc43a747"),
		Type:         "http",
		Status:       "scheduled",
		ExecuteAt:    null.TimeFrom(time.Now()),
		CronSchedule: null.StringFrom("0 0 * * *"),
		HTTPJob:      []byte(`{"url": "localhost:3000", "auth": {"type": "none", "password": null, "username": null, "bearer_token": null}, "body": "", "method": "POST", "headers": null}`),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	job, err := jobDB.ToJob()
	require.NoError(t, err)

	assert.Equal(t, job.ID, jobDB.ID)
	assert.Equal(t, job.Type, model.JobType(jobDB.Type))
	assert.Equal(t, job.Status, model.JobStatus(jobDB.Status))
	assert.Equal(t, job.ExecuteAt, jobDB.ExecuteAt)
	assert.Equal(t, job.CronSchedule, jobDB.CronSchedule)
	assert.Nil(t, job.AMQPJob)
	assert.NotNil(t, job.HTTPJob)
	assert.Equal(t, job.CreatedAt, jobDB.CreatedAt)
	assert.Equal(t, job.UpdatedAt, jobDB.UpdatedAt)
	assert.Equal(t, job.NextRun.Valid, false)
}
