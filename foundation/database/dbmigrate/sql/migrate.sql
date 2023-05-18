-- Version: 1.01
-- Description: Create jobs table and job execution table
CREATE TYPE job_status_enum AS ENUM (
    'PENDING',
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
    status job_status_enum NOT NULL DEFAULT 'PENDING',
    execute_at TIMESTAMPTZ,
    cron_schedule VARCHAR(255),
    http_job JSONB,
    amqp_job JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_message VARCHAR(255),
    next_run TIMESTAMPTZ,
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(255)
);


CREATE TABLE job_executions (
    id SERIAL PRIMARY KEY,
    job_id uuid NOT NULL,
    status job_status_enum NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    error_message TEXT,
    result JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs (id)
);
