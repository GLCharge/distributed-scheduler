# 🏗️ Architecture Overview 🌇

The Golang Distributed Scheduler 🕑 is a powerful system designed to enable other services to schedule jobs that will execute at specified points in the future 📆. 
It's composed of two main, separately deployable components, the Management API 🖥️ and the Runner service 🏃‍♀️.

## 🌇 Management API
The Management API is the user interface for interacting with the scheduling system 🎛️. 
Deployable as a separate binary, it provides an intuitive and straightforward means to create, update, retrieve, and delete jobs 📝. 
In addition, it allows users to fetch all executions of a specific job 👀.

## 🏃‍♂️Runner Service
The Runner service, also deployable as a distinct binary, handles the execution of jobs 🎬. 
It queries the Postgres database for all jobs due to run (those where the `next_run` field is set to a time before "now" ⏰) and updates the job records post-execution. 
It also creates new execution records.

### Components of the Runner Service
1. **Postgres Database** 🗃️: This is where all the job records are stored. Each job record consists of details such as its creation time, when it is due to run next, and its lock status 🔒.

2. **Executor** ⚙️: The Executor component is responsible for executing the jobs fetched by the Runner service. It supports two types of jobs:

   - **HTTP Jobs** 🌐: Users provide an endpoint to call, along with the HTTP method, body, and authentication details for these jobs.
   - **AMQP Jobs** 🐇: Users provide all the details necessary to publish a message to an AMQP exchange for these jobs.

## 📚 Job Types
Jobs can be scheduled as either One-off or Recurring jobs:

- **One-off Jobs** ⏲️: Users set a specific timestamp in the future when the job should run.
- **Recurring Jobs** 🔄: Users set a cron schedule to specify when the job should run repeatedly.

The system also includes a built-in retry mechanism to bolster its reliability in case of temporary failures or network issues⚡.

##  🔐 Job Execution and Locking Mechanism
To prevent a job from executing multiple times simultaneously, the system leverages Postgres' locking mechanism. When the Runner service fetches a job to run from the database, it sets the `locked_until` field to a future timestamp⏱️. 
This action bars other Runner service instances from attempting to execute the job until the `locked_until` time has elapsed. 
Once a job finishes executing, the Runner service sets `locked_until` back to null and updates the `next_run` field to schedule the next execution 🗓️.

This distributed architecture allows for the deployment of multiple instances of both the Management API and Runner services without the risk of a job being executed multiple times 🔄. 
The robust scalability and reliability make this system capable of handling a large volume of scheduled jobs. 🏋️‍♂️
