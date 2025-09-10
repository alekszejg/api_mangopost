package main

import (
	"context"
	"log/slog"
	"my-api/gmail"
	hooks "my-api/webhooks"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load("./.env.mangopost"); err != nil {
		slog.Error("failed to load .env.mangopost variables:", slog.Any("error", err))
		os.Exit(1)
	}

	mode := os.Getenv("GIN_MODE")
	slog.SetDefault(setupLogger(mode))

	if err := initAtStartup(); err != nil {
		slog.Error("Failed to start application", slog.Any("error", err))
		os.Exit(1)
	}

	go startScheduledJobs(ctx)

	router := setupRouter(mode)

	if mode == "debug" {
		router.SetTrustedProxies(nil) // No proxies on localhost
	} else {
		proxies := os.Getenv("TRUSTED_PROXIES")
		if proxies == "" {
			slog.Error("Failed to start application", slog.Any("error", "failed to load TRUSTED_PROXIES .env"))
			os.Exit(1)
		}

		proxyList := strings.Split(proxies, ",")
		if err := router.SetTrustedProxies(proxyList); err != nil {
			slog.Error("Failed to set trusted proxies", slog.Any("error", err))
			os.Exit(1)
		}
	}

	router.GET("/auth", gmail.OAuthHandler)           // FOR MANUAL OAUTH SETUP
	router.GET("/auth/callback", gmail.OAuthCallback) // FOR MANUAL OAUTH SETUP
	router.POST("/api/events", hooks.Receiver)

	slog.Info("Starting server", slog.String("port", "8080"))
	if err := router.Run(":8080"); err != nil {
		slog.Error("Failed to start server", slog.Any("error", err))
		os.Exit(1)
	}

	<-ctx.Done()
	slog.Info("Shutdown complete")
}
