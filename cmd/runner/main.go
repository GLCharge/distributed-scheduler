package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GLCharge/distributed-scheduler/executor"
	"github.com/GLCharge/distributed-scheduler/foundation/database"
	"github.com/GLCharge/distributed-scheduler/foundation/logger"
	"github.com/GLCharge/distributed-scheduler/runner"
	"github.com/GLCharge/distributed-scheduler/service/job"
	"github.com/GLCharge/distributed-scheduler/store/postgres"
	"github.com/ardanlabs/conf/v3"
	"go.uber.org/zap"
)

var build = "develop"

func main() {
	log, err := logger.New("RUNNER")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		DB struct {
			User         string `conf:"default:scheduler"`
			Password     string `conf:"default:scheduler,mask"`
			Host         string `conf:"default:localhost:5436"`
			Name         string `conf:"default:scheduler"`
			MaxIdleConns int    `conf:"default:3"`
			MaxOpenConns int    `conf:"default:2"`
			DisableTLS   bool   `conf:"default:true"`
		}
		ID                string        `conf:"default:instance1"`
		Interval          time.Duration `conf:"default:10s"`
		MaxConcurrentJobs int           `conf:"default:100"`
		MaxJobLockTime    time.Duration `conf:"default:1m"`
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "RUNNER"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// -------------------------------------------------------------------------
	// Database Support

	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// -------------------------------------------------------------------------
	// Start Runner Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Infow("startup", "status", "initializing runner")

	store := postgres.New(db, log)

	jobService := job.NewService(store, log)

	executorFactory := executor.NewFactory(&http.Client{Timeout: 30 * time.Second})

	runnner := runner.New(runner.Config{
		JobService:        jobService,
		Log:               log,
		ExecutorFactory:   executorFactory,
		InstanceId:        cfg.ID,
		Interval:          cfg.Interval,
		MaxConcurrentJobs: cfg.MaxConcurrentJobs,
		JobLockDuration:   cfg.MaxJobLockTime,
	})

	runnner.Start()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// stop the runner
		runnner.Stop(ctx)
	}

	return nil
}
