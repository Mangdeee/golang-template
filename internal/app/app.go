package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/felixlambertv/go-cleanplate/internal/di"
	"github.com/getsentry/sentry-go"

	"github.com/felixlambertv/go-cleanplate/config"
	v1 "github.com/felixlambertv/go-cleanplate/internal/controller/http/v1"
	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/pkg/httpserver"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run(cfg *config.Config) {
	l := logger.NewLogger(cfg.Log.Level)
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Monitoring.Sentry,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.PG.GetDbConnectionUrl()))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres: %w", err))
	}
	err = db.AutoMigrate(
		&model.User{},
	)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - migrate: %w", err))
	}

	di := di.NewDependencyInjection(db, l, cfg)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, l, db, cfg, di)
	httpServer := httpserver.NewServer(handler, httpserver.Port(cfg.HTTP.Port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	for {
		fmt.Println("worker receive message")
		err := di.QueueService.ReceiveMessage()
		if err != nil {
			fmt.Printf("failed to process or get message: %v\n", err)
			continue
		}
	}

	select {
	case s := <-interrupt:
		l.Info("app run: " + s.String())
	case err := <-httpServer.Notify():
		l.Error(fmt.Errorf("%w", err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("%w", err))
	}
}
