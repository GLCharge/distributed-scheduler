package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/GLCharge/otelzap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GLCharge/distributed-scheduler/foundation/database"
	"github.com/GLCharge/distributed-scheduler/foundation/logger"
	"github.com/GLCharge/distributed-scheduler/handlers"
	"github.com/ardanlabs/conf/v3"
	"go.uber.org/zap"
)

var build = "develop"

func main() {
	logLevel := os.Getenv("MANAGER_LOG_LEVEL")
	log, err := logger.New(logLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Error("startup error", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}
}

func run(log *otelzap.Logger) error {

	// -------------------------------------------------------------------------
	// Configuration
	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:8000"`
		}
		DB struct {
			User         string `conf:"default:scheduler"`
			Password     string `conf:"default:scheduler,mask"`
			Host         string `conf:"default:localhost:5436"`
			Name         string `conf:"default:scheduler"`
			MaxIdleConns int    `conf:"default:3"`
			MaxOpenConns int    `conf:"default:2"`
			DisableTLS   bool   `conf:"default:true"`
		}
		OpenAPI struct {
			Scheme string `conf:"default:http"`
			Enable bool   `conf:"default:true"`
			Host   string `conf:"default:localhost:8000"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "MANAGER"
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

	log.Info("starting service", zap.String("version", build))
	defer log.Info("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info("startup", zap.String("config", out))

	// -------------------------------------------------------------------------
	// Database Support

	log.Info("startup", zap.String("status", "initializing database support"), zap.String("host", cfg.DB.Host))

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
		log.Info("shutdown", zap.String("status", "stopping database support"), zap.String("host", cfg.DB.Host))
		db.Close()
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info("startup", zap.String("status", "initializing Management API support"))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Log: log,
		DB:  db,
		OpenApi: handlers.OpenApiConfig{
			Enabled: cfg.OpenAPI.Enable,
			Scheme:  cfg.OpenAPI.Scheme,
			Host:    cfg.OpenAPI.Host,
		},
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Logger),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("startup", zap.String("status", "api router started"), zap.String("host", api.Addr))
		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Info("shutdown", zap.String("status", "shutdown started"), zap.Any("signal", sig))
		defer log.Info("shutdown", zap.String("status", "shutdown complete"), zap.Any("signal", sig))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
