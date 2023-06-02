# 🏗️ Architecture Overview

The Golang Scheduler is designed to enable other services to schedule jobs that will execute at specific points in the future.
The system consists of several crucial components:

1. **Management API**: This is the interface through which users can interact with the scheduler. 
The API provides capabilities for creating, updating, retrieving, and deleting jobs, along with retrieving all executions of a specific job.

2. **Postgres Database**: This is where all the jobs are stored. 
Each job record includes details such as its creation time, when it is due to run next, and its lock status.

3. **Scheduler**: The Scheduler queries the Postgres database for all jobs due to run (those where the `next_run` field is set to a time before "now"). 
Post execution, the Scheduler updates the job record in the database and creates a new execution record.

4. **Executor**: This component is in charge of executing the jobs fetched by the Scheduler. The Executor supports two types of jobs:

   - **HTTP Jobs**: For these jobs, users provide an endpoint to call along with the HTTP method, body, and authentication details.
   - **AMQP Jobs**: For these jobs, users provide all the details necessary to publish a message to an AMQP exchange.

Jobs can be scheduled to run as either One-off or Recurring jobs:

- **One-off Jobs**: Users set a specific timestamp in the future when the job should run.
- **Recurring Jobs**: Users set a cron schedule to specify when the job should run repeatedly.

The system also includes a built-in retry mechanism to bolster its reliability in case of temporary failures or network issues.

To prevent a job from executing multiple times simultaneously, the system leverages Postgres' locking mechanism. 
When the Scheduler fetches a job to run from the database, it sets the `locked_until` field to a future timestamp. 
This prevents other Scheduler instances from attempting to execute the job until the `locked_until` time has elapsed. 
Once a job finishes executing, the Scheduler sets `locked_until` back to null and updates the `next_run` field to schedule the next execution.

This architecture allows for the deployment of multiple Scheduler instances without the risk of a job being executed multiple times. 
This scalability and reliability make the system capable of handling a large volume of scheduled jobs.



## 💻 Local Development Guide
This guide will help you get the Golang Scheduler set up and running on your local machine for development and testing purposes.

### Prerequisites
Ensure that you have Go 1.20 installed on your local machine. If you haven't, you can download it from the official Go website.

### Build and Run
The Golang Scheduler uses make commands for building and running the application locally. Follow the steps below to start the scheduler:

1. Build the application: 
```bash
make build
```
2. Run Postgres database migrations: 
```bash
make db-migrate
```
Run the `db-migrate` command every time there are changes in the Postgres schema.
3. Start the application: 
```bash
make run
```

The `make run` command starts a Postgres database and runs the scheduler binary.


## 🚀 Deployment Guide
The Golang Scheduler is a single binary application that utilizes environment variables and command-line flags for configuration. 
It connects to a Postgres database that needs to be hosted somewhere accessible to the application.

### Set Up Environment Variables and Command-line Flags
The Golang Scheduler relies on environment variables and command-line flags for its configuration. You need to set these up in your deployment environment.

To display all configuration options for the application, run the make command:

```bash
make get/flags
```
This will display all the environment variables and command-line flags that can be used to configure the application.

Based on the output, set up the necessary environment variables and command-line flags in your deployment environment.

### Build and Run

Build the Docker image using the provided Dockerfile. Replace <image-name> with the desired name for the Docker image:

```bash
docker build -t <image-name> .
```

Image building step can be part of CI/CD pipeline. After building the image, you can push it to a Docker registry.

If you are using a Docker registry, you can pull the image from the registry and run it on your deployment environment.

```bash
docker run -d <image-name>
```

If you are not using a Docker registry, you can copy the image to your deployment environment and run it.

```bash
docker save <image-name> | ssh user@host "docker load"
ssh user@host "docker run -d <image-name>"
```

If you are not using Docker, you can copy the binary to your deployment environment and run it.

```bash
scp scheduler user@host:~/scheduler
ssh user@host "chmod +x scheduler"
ssh user@host "./scheduler"
```

If you are using AWS fargate, you can use the provided Dockerfile to build the image and push it to ECR.
    
    ```bash
    aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com
    docker build -t <image-name> .
    docker tag <image-name>:latest <account-id>.dkr.ecr.<region>.amazonaws.com/<image-name>:latest
    docker push <account-id>.dkr.ecr.<region>.amazonaws.com/<image-name>:latest
    ```
Then you can create a task definition and run it on fargate.


# Configuration

The application's configuration can be managed through a set of environment variables or command line flags. These variables allow you to customize various aspects of the application, such as the web server settings, database connection parameters, and scheduler parameters.

## Web Server Parameters

These parameters control the web server's behavior. The application uses timeouts to avoid hanging processes and to facilitate resource management.

- `web-read-timeout`, `SCHEDULER_WEB_READ_TIMEOUT` (default: 5s)
- `web-write-timeout`, `SCHEDULER_WEB_WRITE_TIMEOUT` (default: 10s)
- `web-idle-timeout`, `SCHEDULER_WEB_IDLE_TIMEOUT` (default: 120s)
- `web-shutdown-timeout`, `SCHEDULER_WEB_SHUTDOWN_TIMEOUT` (default: 20s)
- `web-api-host`, `SCHEDULER_WEB_API_HOST` (default: 0.0.0.0:8000)

## Database Connection Parameters

These parameters are used to connect to the database.

- `db-user`, `SCHEDULER_DB_USER` (default: scheduler)
- `db-password`, `SCHEDULER_DB_PASSWORD` (default: xxxxxx)
- `db-host`, `SCHEDULER_DB_HOST` (default: localhost:5436)
- `db-name`, `SCHEDULER_DB_NAME` (default: scheduler)
- `db-max-idle-conns`, `SCHEDULER_DB_MAX_IDLE_CONNS` (default: 3)
- `db-max-open-conns`, `SCHEDULER_DB_MAX_OPEN_CONNS` (default: 2)
- `db-disable-tls`, `SCHEDULER_DB_DISABLE_TLS` (default: true)

## Runner Parameters

These parameters control the operation of the runner. They help manage the execution of jobs and the resources assigned to them.

- `scheduler-id`, `SCHEDULER_SCHEDULER_ID` (default: instance1)
- `scheduler-interval`, `SCHEDULER_SCHEDULER_INTERVAL` (default: 10s)
- `scheduler-max-concurrent-jobs`, `SCHEDULER_SCHEDULER_MAX_CONCURRENT_JOBS` (default: 100)
- `scheduler-max-job-lock-time`, `SCHEDULER_SCHEDULER_MAX_JOB_LOCK_TIME` (default: 1m)

## Using Configuration Flags

You can pass these flags directly when starting the application. For example:

```bash
./scheduler --web-read-timeout=6s --db-user=myuser
```

## Using Environment Variables

You can also set these variables as environment variables. For example:

```bash
export SCHEDULER_WEB_READ_TIMEOUT=6s
export SCHEDULER_DB_USER=myuser
./scheduler
```
