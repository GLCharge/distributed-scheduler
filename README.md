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
Once a job finishes executing, the Scheduler sets `locked_until` back to null and updates the `next_run field to schedule the next execution.

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

2. Start the application: 
```bash
make run
```

The `make run` command starts a Postgres database and runs the scheduler binary.