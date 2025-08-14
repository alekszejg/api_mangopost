package handlers

import (
	"encoding/json"
	"fmt"
	"my-api/slack"
	"my-api/utils"

	"github.com/gin-gonic/gin"
)

type newMessage struct {
	Chat struct {
		FullName string `json:"full_name"`
		ChatURL  string `json:"chat_url"`
		Phone    string `json:"phone"`
	} `json:"chat"`

	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

func NewTimelinesMessage(ctx *gin.Context, rawData json.RawMessage) error {
	var data newMessage
	if err := utils.UnmarshalOrErr(rawData, &data); err != nil {
		return err
	}

	phone := data.Chat.Phone
	if phone != "" {
		phone = fmt.Sprintf("<tel:+%s|+%s>", data.Chat.Phone, data.Chat.Phone)
	}

	slackText := fmt.Sprintf("*New message from %s*\n_%q_", data.Chat.FullName, data.Message.Text)
	payload := slack.NewMessage(slackText).Attach([]slack.Attachment{
		{
			Color:  "#0088cc",
			Text:   fmt.Sprintf("Phone: %s\nLink: <%s|Open TimelinesAI>", phone, data.Chat.ChatURL),
			Footer: "TimelinesAI via GAS",
		},
	})

	return slack.Internal.Send(ctx.Request.Context(), *payload)
}

func AccountConnected(ctx *gin.Context, _ json.RawMessage) error {
	chatUrl := "https://app.timelines.ai/whatsapp"
	slackText := fmt.Sprintf("*WA account is connected again!*\n<%s|Manage in TimelinesAI>", chatUrl)
	payload := slack.NewMessage(slackText)
	return slack.Internal.Send(ctx.Request.Context(), *payload)
}

func AccountDisconnected(ctx *gin.Context, _ json.RawMessage) error {
	chatUrl := "https://app.timelines.ai/whatsapp"
	slackText := fmt.Sprintf("*WA account was disconnected!*\n<%s|Manage in TimelinesAI>", chatUrl)
	payload := slack.NewMessage(slackText)
	return slack.Internal.Send(ctx.Request.Context(), *payload)
}
