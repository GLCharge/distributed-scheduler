package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GLCharge/otelzap"
	"gopkg.in/guregu/null.v4"
	"time"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/store"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type pgStore struct {
	db  *sqlx.DB
	log *otelzap.Logger
}

// New creates a new PostgresSQL store.
func New(db *sqlx.DB, log *otelzap.Logger) store.Storer {
	return &pgStore{
		db:  db,
		log: log,
	}
}

func (s *pgStore) UpdateJob(ctx context.Context, job *model.Job) error {

	dbJob, err := toJobDB(job)
	if err != nil {
		return fmt.Errorf("failed to convert job to database job: %w", err)
	}

	query := `
		UPDATE
			jobs
		SET
			 type = :type,
			 execute_at = :execute_at,
			 cron_schedule = :cron_schedule,
			 http_job = :http_job,
			 amqp_job = :amqp_job,
			 updated_at = :updated_at,
			 next_run = :next_run
		WHERE id = :id
		`

	_, err = s.db.NamedExecContext(ctx, query, dbJob)
	if err != nil {
		return fmt.Errorf("failed to update job in database: %w", err)
	}

	return nil
}

func (s *pgStore) GetJobExecutions(ctx context.Context, jobID uuid.UUID, failedOnly bool, limit, offset uint64) ([]*model.JobExecution, error) {

	extraFilter := ""
	if failedOnly {
		extraFilter = " AND status = 'FAILED'"
	}

	query := `
		SELECT
			*
		FROM
			job_executions
		WHERE
			job_id = $1` + extraFilter +
		` ORDER BY start_time DESC
		LIMIT $2 OFFSET $3`

	var dbExecutions []*executionDB
	err := s.db.SelectContext(ctx, &dbExecutions, query, jobID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get job executions from database: %w", err)
	}

	// convert the JobExecutionDB struct to a JobExecution struct
	var executions []*model.JobExecution
	for _, dbExecution := range dbExecutions {
		executions = append(executions, dbExecution.ToModel())
	}

	return executions, nil

}

func (s *pgStore) CreateJob(ctx context.Context, job *model.Job) error {

	dbJob, err := toJobDB(job)
	if err != nil {
		return fmt.Errorf("failed to convert job to db job: %w", err)
	}

	// insert job struct into database
	query := `
	 INSERT INTO jobs (
		 id,
		 type,
		 status,
		 execute_at,
		 cron_schedule,
		 http_job,
		 amqp_job,
		 created_at,
		 updated_at,
		 next_run
	 ) VALUES (
		 :id,
		 :type,
		 :status,
		 :execute_at,
		 :cron_schedule,
		 :http_job,
		 :amqp_job,
		 :created_at,
		 :updated_at,
		 :next_run
	 )
 `

	_, err = s.db.NamedExecContext(ctx, query, dbJob)
	if err != nil {
		return fmt.Errorf("failed to insert job into database: %w", err)
	}

	return nil
}

func (s *pgStore) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	// create a JobDB struct to hold the result of the query
	var dbJob jobDB

	// execute the query to get the job by ID
	query := `
        SELECT * FROM jobs WHERE id = $1
    `
	err := s.db.GetContext(ctx, &dbJob, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job from database: %w", err)
	}

	// convert the JobDB struct to a Job struct
	job, err := dbJob.ToJob()
	if err != nil {
		return nil, fmt.Errorf("failed to convert db job to job: %w", err)
	}

	return job, nil
}

func (s *pgStore) DeleteJob(ctx context.Context, id uuid.UUID) error {
	// delete job from database
	query := `
        DELETE FROM jobs WHERE id = $1
    `
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job from database: %w", err)
	}

	return nil
}

func (s *pgStore) ListJobs(ctx context.Context, limit, offset uint64) ([]*model.Job, error) {
	// get all jobs from database
	query := `
        SELECT * FROM jobs ORDER BY id DESC LIMIT $1 OFFSET $2 
    `
	var dbJobs []*jobDB
	err := s.db.SelectContext(ctx, &dbJobs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs from database: %w", err)
	}

	// convert JobDB structs to Job structs
	var jobs []*model.Job
	for _, dbJob := range dbJobs {
		job, err := dbJob.ToJob()
		if err != nil {
			return nil, fmt.Errorf("failed to convert db job to job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *pgStore) GetJobsToRun(ctx context.Context, at time.Time, lockedUntil time.Time, instanceID string, limit uint) ([]*model.Job, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer rollback(tx, s.log)

	// Get jobs that should be run at time at and are not currently locked
	rows, err := tx.QueryContext(ctx, `
	   SELECT *
	   FROM jobs
	   WHERE next_run <= $1 AND (locked_until IS NULL OR locked_until <= $2) AND status = 'RUNNING'
	   LIMIT $3
	   FOR UPDATE SKIP LOCKED
	`, at, at, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var dbjobs []*jobDB
	err = sqlx.StructScan(rows, &dbjobs)
	if err != nil {
		return nil, fmt.Errorf("failed to scan job: %w", err)
	}

	var jobs []*model.Job
	for _, dbJob := range dbjobs {

		job, err := dbJob.ToJob()
		if err != nil {
			return nil, fmt.Errorf("failed to convert db job to job: %w", err)
		}
		jobs = append(jobs, job)

		// Mark the job as locked by this instance
		if _, err := tx.ExecContext(ctx, `
	       UPDATE jobs
	       SET locked_until = $1, locked_by = $2
	       WHERE id = $3
	   `, lockedUntil, instanceID, job.ID); err != nil {
			return nil, fmt.Errorf("failed to lock job: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return jobs, nil
}

func (s *pgStore) FinishJob(ctx context.Context, jobID uuid.UUID, nextRun null.Time) error {

	// finish job in database
	query := `
		UPDATE jobs SET 
		        next_run = $1, 
		        locked_until = null, locked_by = null, updated_at = now() 
		WHERE id = $2
	`
	_, err := s.db.ExecContext(ctx, query, nextRun, jobID)
	if err != nil {
		return fmt.Errorf("failed to finish job in database: %w", err)
	}

	return nil
}
func (s *pgStore) CreateJobExecution(ctx context.Context, jobID uuid.UUID, startTime, stopTime time.Time, status model.JobExecutionStatus, errorMessage null.String) error {

	// create job execution in database
	query := `
		INSERT INTO job_executions (job_id, start_time, end_time, status, error_message, created_at) 
		VALUES ($1, $2, $3, $4, $5, now())
	`
	_, err := s.db.ExecContext(ctx, query, jobID, startTime, stopTime, status, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to create job execution in database: %w", err)
	}

	return nil
}
