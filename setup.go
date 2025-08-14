package main

import (
	"context"
	"fmt"
	"log/slog"
	"my-api/gmail"
	"my-api/jobs"
	"my-api/slack"
	hooks "my-api/webhooks"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func setupLogger(mode string) *slog.Logger {
	var logger *slog.Logger
	if mode == "debug" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else if mode == "release" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return logger
}

func setupRouter(mode string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(ctx *gin.Context) {
		start := time.Now()
		slog.Info("Request",
			slog.String("id", uuid.New().String()),
			slog.String("route", fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)),
			slog.String("ip", ctx.ClientIP()))
		ctx.Next()
		slog.Info("HTTP Response",
			slog.Int("status", ctx.Writer.Status()),
			slog.Any("error", ctx.Err()),
			slog.Duration("latency", time.Since(start)))
	})

	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	return router
}

func initAtStartup() error {
	inits := []struct {
		name string
		fn   func() error
	}{
		{name: "gmail.InitConfig()", fn: gmail.InitConfig},
		{name: "slack.InitChannels()", fn: slack.InitChannels},
		{name: "hooks.InitEventHandling()", fn: hooks.InitEventHandling},
	}

	for _, init := range inits {
		if err := init.fn(); err != nil {
			slog.Error(fmt.Sprintf("initialization failed: %s", err.Error()))
			return fmt.Errorf("initialization failed. Source: %s Error: %w", init.name, err)
		}
	}

	return nil
}

func startScheduledJobs(ctx context.Context) {
	jm := jobs.NewJobManager(ctx)
	jm.AppendJob((jobs.FoodSpotThreadsJob{}))
	jm.ScheduleCronjobs()

	go func() {
		<-ctx.Done()
		slog.Info("Received shutdown signal, stopping all scheduled jobs")
		jm.Stop()
	}()
}
