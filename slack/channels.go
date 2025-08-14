package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"my-api/utils"
	"net/http"
	"os"
	"time"
)

type Channel struct{ Name, URL string }

var (
	client                               *http.Client
	Internal, OrderHistory, ScriptErrors Channel
)

func InitChannels() error {
	client = &http.Client{Timeout: 10 * time.Second}

	Internal = Channel{Name: "internal-notifications", URL: os.Getenv("SLACK_INTERNAL_NOTIFICATIONS")}
	OrderHistory = Channel{Name: "order-history", URL: os.Getenv("SLACK_ORDER_HISTORY")}
	ScriptErrors = Channel{Name: "script-errors", URL: os.Getenv("SLACK_SCRIPT_ERRORS")}

	if Internal.URL == "" || OrderHistory.URL == "" || ScriptErrors.URL == "" {
		return fmt.Errorf("failed to initialize Slack channels. Invalid .env variables")
	}

	return nil
}

func NewMessage(text string) *Payload {
	return &Payload{Text: text}
}

func (c Channel) Send(ctx context.Context, payload Payload) error {
	if payload.Text == "" {
		return &utils.APIError{
			Err:    fmt.Errorf("empty slack payload text"),
			Status: 400,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &utils.APIError{
			Err:    fmt.Errorf("failed to marshal json: %w", err),
			Status: 500,
		}
	}

	for attempt := 0; attempt <= 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewBuffer(body))
		if err != nil {
			return &utils.APIError{
				Err:    fmt.Errorf("failed to create Slack POST request: %w", err),
				Status: 500,
			}
		}

		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			if attempt < 2 {
				time.Sleep(1 * time.Second)
				continue
			} else {
				return &utils.APIError{
					Err:    fmt.Errorf("failed to send Slack request after %d attempts: %w", attempt+1, err),
					Status: 504,
				}
			}
		}

		defer res.Body.Close()

		if res.StatusCode == 200 {
			break
		}

		if attempt < 2 {
			time.Sleep(1 * time.Second)
			continue
		} else {
			return &utils.APIError{
				Err:    fmt.Errorf("slack returned non-200 status: %d", res.StatusCode),
				Status: res.StatusCode,
			}
		}
	}

	slog.Debug("Successfully sent message to Slack")
	return nil
}
