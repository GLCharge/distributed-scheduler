package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/GLCharge/distributed-scheduler/model"
	"github.com/GLCharge/distributed-scheduler/store"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const maxJobLockDuration = 5 * time.Minute

type pgStore struct {
	db  *sqlx.DB
	log *zap.SugaredLogger
}

// New creates a new PostgreSQL store.
func New(db *sqlx.DB, log *zap.SugaredLogger) store.Storer {
	return &pgStore{
		db:  db,
		log: log,
	}
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
		 error_message,
		 next_run,
		 locked_at,
		 locked_by
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
		 :error_message,
		 :next_run,
		 :locked_at,
		 :locked_by
	 )
 `

	_, err = s.db.NamedExecContext(ctx, query, dbJob)
	if err != nil {
		return fmt.Errorf("failed to insert job into database: %w", err)
	}

	return nil
}

func (s *pgStore) GetJob(ctx context.Context, id string) (*model.Job, error) {
	// create a JobDB struct to hold the result of the query
	var dbJob jobDB

	// execute the query to get the job by ID
	query := `
        SELECT * FROM jobs WHERE id = $1
    `
	err := s.db.GetContext(ctx, &dbJob, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found: %w", err)
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

func (s *pgStore) DeleteJob(ctx context.Context, id string) error {
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
        SELECT * FROM jobs LIMIT $1 OFFSET $2
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

func (s *pgStore) GetJobsToRun(ctx context.Context, t time.Time, instanceID string) ([]*model.Job, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get jobs that should be run at time t and are not currently locked
	rows, err := tx.QueryContext(ctx, `
	   SELECT *
	   FROM jobs
	   WHERE next_run <= $1 AND (locked_at IS NULL OR locked_at < $2)
	   LIMIT 10
	   FOR UPDATE SKIP LOCKED
	`, t, t.Add(-maxJobLockDuration))
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
	       SET locked_at = $1, locked_by = $2
	       WHERE id = $3
	   `, t, instanceID, job.ID); err != nil {
			return nil, fmt.Errorf("failed to lock job: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return jobs, nil
}

func (s *pgStore) UnlockJob(ctx context.Context, jobID uuid.UUID) error {

	// unlock job in database
	query := `
		UPDATE jobs SET locked_at = NULL, locked_by = NULL WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to unlock job in database: %w", err)
	}

	return nil

}
