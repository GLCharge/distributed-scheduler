package job

import (
	"context"
	"fmt"
	"github.com/GLCharge/distributed-scheduler/foundation/database/dbtest"
	"github.com/GLCharge/distributed-scheduler/foundation/docker"
	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/store/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gopkg.in/guregu/null.v4"
	"runtime/debug"
	"testing"
	"time"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_Job(t *testing.T) {
	t.Run("crud", crud)
	t.Run("job_execution", jobExecution)
}

func crud(t *testing.T) {
	// Init
	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	jobService := NewService(postgres.New(test.DB, test.Log), test.Log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create job 1
	// -------------------------------------------------------------------------

	job, err := jobService.CreateJob(ctx, &model.JobCreate{
		Type:         model.JobTypeHTTP,
		CronSchedule: null.StringFrom("@every 1m"),
		HTTPJob:      &model.HTTPJob{URL: "https://google.com", Method: "GET", Auth: model.Auth{Type: model.AuthTypeNone}},
	})

	if err != nil {
		t.Fatalf("Should be able to create a job: %s", err)
	}

	if job.ID == uuid.Nil {
		t.Fatalf("Should get back an ID: %s", job.ID)
	}

	if job.NextRun.IsZero() {
		t.Fatalf("Should get back a next run time: %v", job.NextRun)
	}

	// Get job 1
	// -------------------------------------------------------------------------
	job1, err1 := jobService.GetJob(ctx, job.ID)
	if err1 != nil {
		t.Fatalf("Should be able to get a job: %s", err1)
	}

	// compare jobs
	if !cmp.Equal(job, job1) {
		t.Fatalf("Should get back the same job: %s", cmp.Diff(job, job1))
	}

	// Create job 2
	// -------------------------------------------------------------------------

	job2, err := jobService.CreateJob(ctx, &model.JobCreate{
		Type:         model.JobTypeHTTP,
		CronSchedule: null.StringFrom("@every 1m"),
		HTTPJob:      &model.HTTPJob{URL: "https://google.com", Method: "GET", Auth: model.Auth{Type: model.AuthTypeNone}},
	})

	if err != nil {
		t.Fatalf("Should be able to create a job: %s", err)
	}

	if job2.ID == uuid.Nil {
		t.Fatalf("Should get back an ID: %s", job2.ID)
	}

	// update job
	// -------------------------------------------------------------------------
	job, err = jobService.UpdateJob(ctx, job.ID, model.JobUpdate{
		CronSchedule: lo.ToPtr("@every 2m"),
	})

	if err != nil {
		t.Fatalf("Should be able to update a job: %s", err)
	}

	if job.CronSchedule.String != "@every 2m" {
		t.Fatalf("Should get back an updated cron schedule: %s", job.CronSchedule.String)
	}

	// Get jobs
	// -------------------------------------------------------------------------

	jobs, err := jobService.ListJobs(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Should be able to list jobs: %s", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("Should get back 2 jobs: %d", len(jobs))
	}

	// Get jobs with limit
	// -------------------------------------------------------------------------

	jobs, err = jobService.ListJobs(ctx, 1, 0)
	if err != nil {
		t.Fatalf("Should be able to list jobs: %s", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Should get back 1 job: %d", len(jobs))
	}

	// Delete job
	// -------------------------------------------------------------------------
	err = jobService.DeleteJob(ctx, job.ID)

	if err != nil {
		t.Fatalf("Should be able to delete a job: %s", err)
	}

	// Get job
	// -------------------------------------------------------------------------
	_, err = jobService.GetJob(ctx, job.ID)

	if err == nil {
		t.Fatalf("Should not be able to get a deleted job: %s", err)
	}
}

func jobExecution(t *testing.T) {
	// Init
	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	jobService := NewService(postgres.New(test.DB, test.Log), test.Log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Create job
	// -------------------------------------------------------------------------

	job, err := jobService.CreateJob(ctx, &model.JobCreate{
		Type:      model.JobTypeHTTP,
		ExecuteAt: null.TimeFrom(now.Add(1 * time.Second)),
		HTTPJob:   &model.HTTPJob{URL: "https://www.ardanlabs.com", Method: "GET", Auth: model.Auth{Type: model.AuthTypeNone}},
	})

	if err != nil {
		t.Fatalf("Should be able to create a job: %s", err)
	}

	if job.ID == uuid.Nil {
		t.Fatalf("Should get back an ID: %s", job.ID)
	}

	// Get jobs to run
	// -------------------------------------------------------------------------

	jobs, err := jobService.GetJobsToRun(ctx, now.Add(2*time.Second), now.Add(5*time.Second), "instance1", 10)
	if err != nil {
		t.Fatalf("Should be able to get jobs to run: %s", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Should get back 1 job: %d", len(jobs))
	}

	if jobs[0].ID != job.ID {
		t.Fatalf("Should get back the correct job: %s", jobs[0].ID)
	}

	// Get jobs to run
	// -------------------------------------------------------------------------

	jobs, err = jobService.GetJobsToRun(ctx, now.Add(4*time.Second), now.Add(6*time.Second), "instance1", 10)
	if err != nil {
		t.Fatalf("Should be able to get jobs to run: %s", err)
	}

	// job is locked so should not get back any jobs
	if len(jobs) != 0 {
		t.Fatalf("Should get back 0 jobs: %d", len(jobs))
	}

	// Get jobs to run
	// -------------------------------------------------------------------------

	jobs, err = jobService.GetJobsToRun(ctx, now.Add(6*time.Second), now.Add(8*time.Second), "instance2", 10)
	if err != nil {
		t.Fatalf("Should be able to get jobs to run: %s", err)
	}

	// too much time has passed, so lock should be expired, and we should get back the job
	if len(jobs) != 1 {
		t.Fatalf("Should get back 1 job: %d", len(jobs))
	}

	if jobs[0].ID != job.ID {
		t.Fatalf("Should get back the correct job: %s", jobs[0].ID)
	}

	// complete job
	// -------------------------------------------------------------------------

	err = jobService.FinishJobExecution(ctx, jobs[0], now.Add(6*time.Second), now.Add(7*time.Second), nil)
	if err != nil {
		t.Fatalf("Should be able to finish job execution: %s", err)
	}

	jobs, err = jobService.GetJobsToRun(ctx, now.Add(10*time.Second), now.Add(12*time.Second), "instance2", 10)
	if err != nil {
		t.Fatalf("Should be able to get jobs to run: %s", err)
	}

	// job is complete so should not get back any jobs to run
	if len(jobs) != 0 {
		t.Fatalf("Should get back 0 jobs: %d", len(jobs))
	}

	// get job execution
	// -------------------------------------------------------------------------

	jobExecutions, err := jobService.GetJobExecutions(ctx, job.ID, false, 10, 0)
	if err != nil {
		t.Fatalf("Should be able to get job executions: %s", err)
	}

	if len(jobExecutions) != 1 {
		t.Fatalf("Should get back 1 job execution: %d", len(jobExecutions))
	}

	if jobExecutions[0].JobID != job.ID {
		t.Fatalf("Should get back the correct job execution: %s", jobExecutions[0].JobID)
	}

	jobExecutions, err = jobService.GetJobExecutions(ctx, job.ID, true, 10, 0)
	if err != nil {
		t.Fatalf("Should be able to get job executions: %s", err)
	}

	if len(jobExecutions) != 0 {
		t.Fatalf("Should get back 0 failed job executions: %d", len(jobExecutions))
	}
}
