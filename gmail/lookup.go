package gmail

import (
	"encoding/json"
	"fmt"
	"my-api/slack"
	"net/url"
	"os"
	"strings"
	"time"

	"context"

	"google.golang.org/api/gmail/v1"
)

type LastChecked struct {
	Timestamp time.Time `json:"timestamp"`
}

func slackSummary(threads []*gmail.Thread, permalink string) string {
	var sb strings.Builder
	sb.WriteString("*New FoodSpot requests*\n")
	for i, thread := range threads {
		url := fmt.Sprintf("%s/%s", permalink, thread.Id)
		sb.WriteString(fmt.Sprintf("<%s|View request %d>\n_%s_\n\n", url, i+1, thread.Snippet))
	}

	return sb.String()
}

func saveLastChecked(t time.Time) error {
	data, err := json.MarshalIndent(LastChecked{Timestamp: t}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal last checked timestamp: %v", err)
	}
	if err := os.WriteFile("gmail/last_checked.json", data, 0600); err != nil {
		return fmt.Errorf("failed to write last checked file: %s", err.Error())
	}

	return nil
}

// loadLastChecked loads the last checked timestamp from a file, returning a zero time if not found.
func loadLastChecked() (time.Time, error) {
	data, err := os.ReadFile("gmail/last_checked.json")
	if err != nil {
		if os.IsNotExist(err) {
			return time.Now().UTC(), nil
		}
		return time.Now().UTC(), fmt.Errorf("unexpected failure to read last checked file. Returning current timestamp: %s", err.Error())
	}

	var lastChecked LastChecked
	if err := json.Unmarshal(data, &lastChecked); err != nil {
		return time.Now().UTC(), fmt.Errorf("unexpected failure to parse last checked timestamp. Returning current timestamp: %s", err.Error())
	}

	return lastChecked.Timestamp, nil
}

func getLabelID(client *gmail.Service, labelName string) (string, error) {
	labels, err := client.Users.Labels.List("me").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get gmail labels: %w", err)
	}

	for _, label := range labels.Labels {
		if label.Name == labelName {
			return label.Id, nil
		}
	}

	return "", fmt.Errorf("gmail label %q not found", labelName)
}

func GetThreadsByLabel(client *gmail.Service, labelName string) ([]*gmail.Thread, error) {
	var threads []*gmail.Thread

	labelID, err := getLabelID(client, labelName)
	if err != nil {
		return threads, fmt.Errorf("returning empty thread list: %s", err.Error())
	}

	lastChecked, err := loadLastChecked()
	if err != nil {
		return threads, err
	}

	query := fmt.Sprintf("to:me after:%d", lastChecked.Unix())
	pageToken := ""

	for {
		req := client.Users.Threads.List("me").LabelIds(labelID).Q(query).MaxResults(500)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		response, err := req.Do()
		if err != nil {
			return threads, fmt.Errorf("failed to request gmail threads: %w", err)
		}

		threads = append(threads, response.Threads...)
		if pageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	if err := saveLastChecked(time.Now().UTC()); err != nil {
		return threads, err
	}

	return threads, nil
}

// TestListThreads lists and prints threads with label foodspot-requests
func GetNewThreadsByLabel(ctx context.Context, labelName string) error {
	service, err := getGmailService(ctx)
	if err != nil {
		return err
	}

	threads, err := GetThreadsByLabel(service, labelName)
	if err != nil {
		return fmt.Errorf("failed to list threads: %s", err.Error())
	}

	if len(threads) == 0 {
		return nil
	}

	permalink := fmt.Sprintf("https://mail.google.com/mail/u/%s/#label/%s",
		url.PathEscape(emailUser), url.PathEscape(labelName),
	)

	slackText := slackSummary(threads, permalink)
	payload := slack.NewMessage(slackText)
	return slack.Internal.Send(ctx, *payload)
}
