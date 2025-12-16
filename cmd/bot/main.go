package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"

	"git.ykonkov.com/ykonkov/survey-bot/internal/config"
	"git.ykonkov.com/ykonkov/survey-bot/internal/http"
	"git.ykonkov.com/ykonkov/survey-bot/internal/listener"
	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	"git.ykonkov.com/ykonkov/survey-bot/internal/repository/db"
	"git.ykonkov.com/ykonkov/survey-bot/internal/repository/telegram"
	"git.ykonkov.com/ykonkov/survey-bot/internal/resultsprocessor"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("failed to init config: %w", err)
	}

	flushSentryF, err := logger.InitSentry(config.SentryDSN, config.SentryTimeout)
	if err != nil {
		log.Fatal("failed to init sentry: ", err)
	}
	defer flushSentryF()

	logger := logger.New(config.Env, config.Level, config.ReleaseVersion, os.Stdout)
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	sqlDB, err := db.ConnectWithTimeout(time.Minute, config.DB)
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}

	repo := db.New(sqlDB)
	telegramClient := telegram.NewClient()
	processor := resultsprocessor.New()
	svc := service.New(telegramClient, repo, processor, logger)

	var g run.Group
	{
		logger := logger.WithPrefix("task-name", "message-listener")
		listener, err := listener.New(logger, config.Token, config.AdminUserIDs, config.PollInterval, svc)
		if err != nil {
			log.Fatal("failed to create listener: ", err)
		}

		g.Add(func() error {
			logger.Infof(ctx, "started")
			listener.Start()
			return nil
		}, func(err error) {
			logger.Infof(ctx, "stopped")
			listener.Stop()
		})
	}
	{
		logger := logger.WithPrefix("task-name", "metrics-server")
		metricsServer := http.NewMetricsServer(config.MetricsPort, logger)
		g.Add(func() error {
			logger.Infof(ctx, "started")
			metricsServer.Start()
			return nil
		}, func(err error) {
			metricsServer.Stop()
			logger.Infof(ctx, "stopped")
		})
	}
	{
		logger := logger.WithPrefix("task-name", "http-server")
		httpServer := http.NewAPIServer(
			config.APIPort,
			config.Token,
			config.AllowedOrigins,
			config.AdminUserIDs,
			svc,
			logger,
		)
		g.Add(func() error {
			logger.Infof(ctx, "started")
			return httpServer.Start()
		}, func(err error) {
			if err := httpServer.Stop(); err != nil {
				logger.Errorf(ctx, "failed to stop http server: %v", err)
			} else {
				logger.Infof(ctx, "stopped")
			}
		})
	}
	{
		logger := logger.WithPrefix("task-name", "sig-listener")
		c := make(chan os.Signal, 1)

		g.Add(func() error {
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			logger.Infof(ctx, "started")
			s := <-c
			logger.Warnf(ctx, "program terminated with signal: %v", s)
			return fmt.Errorf("interrupted with sig %q", s)
		}, func(err error) {
			logger.Infof(ctx, "stopped")
			close(c)
			cancelCtx()
		})
	}

	if err := g.Run(); err != nil {
		logger.Errorf(ctx, "stopped with error: %v", err)
	}
}
