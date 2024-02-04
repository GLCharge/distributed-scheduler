-- Version: 1.01
-- Description: Create jobs table and job execution table
CREATE TYPE job_status_enum AS ENUM (
    'RUNNING',
    'STOPPED'
);

CREATE TYPE job_execution_status_enum AS ENUM (
    'SUCCESSFUL',
    'FAILED'
);

CREATE TYPE job_type_enum AS ENUM (
    'HTTP',
    'AMQP'
);

CREATE TABLE jobs (
    id uuid PRIMARY KEY,
    type job_type_enum NOT NULL,
    status job_status_enum NOT NULL DEFAULT 'RUNNING',

    execute_at TIMESTAMPTZ,
    cron_schedule VARCHAR(255),

    http_job JSONB,
    amqp_job JSONB,

    next_run TIMESTAMPTZ,
    locked_until TIMESTAMPTZ,
    locked_by VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure that only one of http_job or amqp_job is set
ALTER TABLE jobs ADD CONSTRAINT
    check_job_type CHECK (
        (type = 'HTTP' AND http_job IS NOT NULL AND amqp_job IS NULL) OR
        (type = 'AMQP' AND http_job IS NULL AND amqp_job IS NOT NULL)
    );

-- Ensure that only one of execute_at or cron_schedule is set
ALTER TABLE jobs ADD CONSTRAINT
    check_job_schedule CHECK (
        (execute_at IS NOT NULL AND cron_schedule IS NULL) OR
        (execute_at IS NULL AND cron_schedule IS NOT NULL)
    );

CREATE INDEX next_run_index ON jobs (next_run);

CREATE INDEX locked_until_index ON jobs (locked_until);

CREATE TABLE job_executions (
    id SERIAL PRIMARY KEY,
    job_id uuid NOT NULL,
    status job_execution_status_enum NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE
);

CREATE INDEX job_id_index ON job_executions (job_id);

CREATE INDEX job_executions_start_time_index ON job_executions (start_time);

-- Version: 1.02
-- Description: Add tags column to jobs table

ALTER TABLE jobs ADD tags TEXT[];