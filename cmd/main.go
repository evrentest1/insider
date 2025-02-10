package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/evrentest1/insider/docs"
	"github.com/evrentest1/insider/internal/api/handlers"
	"github.com/evrentest1/insider/internal/app/domain/task"
	messagecache "github.com/evrentest1/insider/internal/business/domain/message/stores/cache"
	messagerepository "github.com/evrentest1/insider/internal/business/domain/message/stores/db"
	"github.com/evrentest1/insider/internal/config"
	"github.com/evrentest1/insider/internal/foundation/cache"
	"github.com/evrentest1/insider/internal/foundation/httpkit"
	"github.com/evrentest1/insider/internal/foundation/persistence"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const shutdownTimeout = 20 * time.Second

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logrus.Logger) error {
	conf, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	postgresURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.PostgresUser,
		conf.PostgresPassword,
		conf.PostgresHost,
		conf.PostgresPort,
		conf.PostgresDB,
	)
	db, err := persistence.NewDatabase(ctx, postgresURL)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	if err := db.Migrate(postgresURL); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	redisAddress := fmt.Sprintf("%s:%s", conf.RedisHost, conf.RedisPort)
	redis, err := cache.New(ctx, redisAddress)
	if err != nil {
		return fmt.Errorf("init cache: %w", err)
	}

	messageRepository := messagerepository.New(db.DB)
	_ = messagecache.NewStore(redis.Client)

	httpRequester := httpkit.NewRequester(log)

	messageSenderService := task.NewService(redisAddress, log, messageRepository, httpRequester, conf.WebHookSite)
	if err := messageSenderService.Start(ctx); err != nil {
		return fmt.Errorf("start message sender service: %w", err)
	}

	log.Infof("Initializing Echo server")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Gzip())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			err := next(c)

			log.WithFields(logrus.Fields{
				"method":    req.Method,
				"path":      req.URL.Path,
				"status":    res.Status,
				"latency":   time.Since(start),
				"client_ip": c.RealIP(),
			}).Info("Handled request")

			return err
		}
	})

	e.Use(middleware.Recover())

	handler := handlers.New(log, e, db, redis, messageRepository, messageSenderService)
	handler.RegisterRoutes()

	serverErrors := make(chan error, 1)

	go func() {
		address := fmt.Sprint(":", conf.Port)

		log.Infof("Server startup on: %s", address)

		serverErrors <- e.Start(address)
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infof("Shutdown started: %s", sig)
		defer log.Infof("Shutdown completed: %s", sig)

		ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		messageSenderService.Stop(ctx)

		if err := db.DB.Close(); err != nil {
			log.Errorf("could not close database connection: %v", err)
		}

		if err := redis.Client.Close(); err != nil {
			log.Errorf("could not close redis connection: %v", err)
		}

		if err := e.Shutdown(ctx); err != nil {
			e.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
