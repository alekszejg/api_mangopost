package hooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"my-api/utils"
	"my-api/webhooks/handlers"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type eventHandler map[string]func(*gin.Context, json.RawMessage) error

var (
	eventHandlers map[string]eventHandler
	wcSecret      string
)

func InitEventHandling() error {
	wcSecret = os.Getenv("WC_SECRET")
	if wcSecret == "" {
		slog.Error("failed to initialize wcSecret .env variable")
		return fmt.Errorf("failed to initialize wcSecret .env variable")
	}

	eventHandlers = map[string]eventHandler{
		"wc": {
			"order_created": handlers.HandleNewOrder,
			"user_created":  handlers.HandleNewUser,
		},
		"timelines": {
			"new_message":          handlers.NewTimelinesMessage, // actual event - "message:received:new"
			"account_connected":    handlers.AccountConnected,    // actual event - "whatsapp:account:connected"
			"account_disconnected": handlers.AccountDisconnected, // actual event - "whatsapp:account:disconnected"
		},
	}
	return nil
}

func logReceiver(source, from, event string) *slog.Logger {
	return slog.With("source", source, "from", from, "event", event)
}

func Receiver(ctx *gin.Context) {
	from := ctx.Query("from")
	event := ctx.Query("event")
	logger := logReceiver("hooks.Receiver()", from, event)

	logger.Debug("Webhook received")

	handler, ok := eventHandlers[from][event]
	if !ok {
		logger.Warn("Invalid query params")
		ctx.JSON(400, gin.H{"error": "Invalid query parameters values received"})
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "failed to read request body"})
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if from == "wc" && ctx.Request.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if strings.HasPrefix(string(body), "webhook_id=") {
			logger.Info("WooCommerce handshake received", slog.String("body", string(body)))
			ctx.JSON(200, gin.H{"message": "Webhook handshake accepted"})
			return
		}
	}

	var rawData json.RawMessage
	if err := ctx.ShouldBindJSON(&rawData); err != nil {
		logger.Warn("Unexpected request body received")
		ctx.JSON(400, gin.H{"error": "Unexpected request body"})
		return
	}

	if from == "wc" {
		signature := ctx.GetHeader("X-WC-Webhook-Signature")
		if err := utils.ValidateSignature(signature, wcSecret, rawData); err != nil {
			logger.Warn("Invalid webhook signature")
			ctx.JSON(401, gin.H{"error": "Invalid webhook signature"})
			return
		}
	}

	if err := handler(ctx, rawData); err != nil {
		logger.Error("Failed to process webhook data", slog.Any("error", err))

		var apiErr *utils.APIError
		if errors.As(err, &apiErr) && apiErr.Status != 500 {
			ctx.JSON(apiErr.Status, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(500, gin.H{"error": "Internal Error"})
		return
	}

	logger.Info("Successfully handled webhook data")
	ctx.JSON(200, gin.H{"message": "Successfully handled webhook data"})
}
