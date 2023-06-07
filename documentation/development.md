## 💻 Local Development Guide
This guide will help you get the Management API and Runner services set up and running on your local machine for development and testing purposes.

### Prerequisites
Ensure that you have Go 1.20 installed on your local machine. If you haven't, you can download it from the [official Go website](https://golang.org/dl/).

### Build and Run
The Management API and Runner services use make commands for building and running the application locally. Follow the steps below to start the services:

#### Common Steps

1. Run Postgres database migrations:
```bash
make db/migrate
```
Run the `db/migrate` command every time there are changes in the Postgres schema. This database is shared by both the Management API and Runner services.

### Management API

1. Build the Management API binary:
```bash
make build/api
```
2. Run the Management API binary:
```bash
make run/api
```

### Runner

1. Build the Runner binary:
```bash
make build/runner
```

2. Run the Runner binary:
```bash
make run/runner
```

The make `run/runner command starts the Runner service, which will begin to process jobs according to the schedule defined in the database.

### Run Tests
There is a single make command for running all tests:
```bash
make test
```

