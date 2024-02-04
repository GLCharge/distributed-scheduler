# 🎚 Configuration

The application's configuration can be managed through a set of environment variables or command line flags for the Management API. These variables enable you to customize various aspects of the application, such as the web server settings and the database connection parameters.

## 🎛 Management API Configuration

### 🌐 Web Server Parameters

These parameters control the behavior of the Management API's web server. The application uses timeouts to avoid hanging processes and facilitate resource management.

- `--web-read-timeout` / `$MANAGER_WEB_READ_TIMEOUT` (default: 5s)
- `--web-write-timeout` / `$MANAGER_WEB_WRITE_TIMEOUT` (default: 10s)
- `--web-idle-timeout` / `$MANAGER_WEB_IDLE_TIMEOUT` (default: 120s)
- `--web-shutdown-timeout` / `$MANAGER_WEB_SHUTDOWN_TIMEOUT` (default: 20s)
- `--web-api-host` / `$MANAGER_WEB_API_HOST` (default: 0.0.0.0:8000)

### 🗃 Database Connection Parameters

These parameters are used to connect to the database from the Management API.

- `--db-user` / `$MANAGER_DB_USER` (default: scheduler)
- `--db-password` / `$MANAGER_DB_PASSWORD` (default: xxxxxx)
- `--db-host` / `$MANAGER_DB_HOST` (default: localhost:5436)
- `--db-name` / `$MANAGER_DB_NAME` (default: scheduler)
- `--db-max-idle-conns` / `$MANAGER_DB_MAX_IDLE_CONNS` (default: 3)
- `--db-max-open-conns` / `$MANAGER_DB_MAX_OPEN_CONNS` (default: 2)
- `--db-disable-tls` / `$MANAGER_DB_DISABLE_TLS` (default: true)

### 📖 Open API Parameters

These parameters are used to configure the Open API settings for the Management API.

- `--open-api-scheme` / `$MANAGER_OPEN_API_SCHEME` (default: http)
- `--open-api-enable` / `$MANAGER_OPEN_API_ENABLE` (default: true)
- `--open-api-host` / `$MANAGER_OPEN_API_HOST` (default: localhost:8000)

### 🚩 Using Configuration Flags

You can pass these flags directly when starting the Management API. For example:

```bash
./manager --web-read-timeout=6s --db-user=myuser
```

### 🌱 Using Environment Variables

You can also use environment variables to configure the Management API. The environment variables are prefixed with `MANAGER_` and are uppercase. For example:

```bash
MANAGER_LOG_LEVEL=info MANAGER_WEB_READ_TIMEOUT=6s MANAGER_DB_USER=myuser ./manager 
```


*Note*: Please remember to replace the `xxxxxx` with your database password before starting the services.


## 🏃‍ Runner Configuration

The Runner service also supports configuration through environment variables or command line flags. These settings primarily relate to the database connection and the execution of the jobs.

### 🗃 Database Connection Parameters

These parameters are used to connect to the database from the Runner service.

- `--db-user` / `$RUNNER_DB_USER` (default: scheduler)
- `--db-password` / `$RUNNER_DB_PASSWORD` (default: xxxxxx)
- `--db-host` / `$RUNNER_DB_HOST` (default: localhost:5436)
- `--db-name` / `$RUNNER_DB_NAME` (default: scheduler)
- `--db-max-idle-conns` / `$RUNNER_DB_MAX_IDLE_CONNS` (default: 3)
- `--db-max-open-conns` / `$RUNNER_DB_MAX_OPEN_CONNS` (default: 2)
- `--db-disable-tls` / `$RUNNER_DB_DISABLE_TLS` (default: true)

### 🏃‍♂️ Runner Parameters

These parameters control the operation of the runner. They help manage the execution of jobs and the resources assigned to them.

- `--id` / `$RUNNER_ID` (default: instance1)
- `--interval` / `$RUNNER_INTERVAL` (default: 10s)
- `--max-concurrent-jobs` / `$RUNNER_MAX_CONCURRENT_JOBS` (default: 100)
- `--max-job-lock-time` / `$RUNNER_MAX_JOB_LOCK_TIME` (default: 1m)

### 🚩 Using Configuration Flags

You can pass these flags directly when starting the Runner. For example:

```bash
./runner --interval=15s --db-user=myuser
```

### 🌱 Using Environment Variables

You can also use environment variables to configure the Runner. The environment variables are prefixed with `RUNNER_` and are uppercase. For example:

```bash
RUNNER_LOG_LEVEL=info RUNNER_INTERVAL=15s RUNNER_DB_USER=myuser ./runner
```

*Note*: Please remember to replace the `xxxxxx` with your database password before starting the services.