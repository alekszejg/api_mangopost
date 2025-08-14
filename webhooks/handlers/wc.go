package handlers

import (
	"encoding/json"
	"fmt"
	"my-api/slack"
	"my-api/utils"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleNewUser(ctx *gin.Context, rawData json.RawMessage) error {
	var user NewUser
	if err := utils.UnmarshalOrErr(rawData, &user); err != nil {
		return err
	}

	payload := slack.NewMessage(
		fmt.Sprintf("*New User - %s*\nName: %s %s\nEmail: %s",
			user.Username, user.FirstName, user.LastName, user.Email),
	)

	return slack.Internal.Send(ctx.Request.Context(), *payload)
}

func HandleNewOrder(ctx *gin.Context, rawData json.RawMessage) error {
	var order NewOrder
	if err := utils.UnmarshalOrErr(rawData, &order); err != nil {
		return err
	}

	orderID := strconv.Itoa(order.ID)
	orderURL := os.Getenv("WC_ORDER_URL")
	slackURLFormat := fmt.Sprintf("<%s&id=%s|View in Wordpress>", orderURL, orderID)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*New Order #\u200B%s*\n\n", orderID))
	order.slackFormatPayment(&sb)
	if err := order.slackFormatDeliveryDate(&sb); err != nil {
		return err
	}
	order.slackFormatCustomer(&sb)
	order.slackFormatVendor(&sb)

	text := sb.String()
	payload := slack.NewMessage(text).Attach(
		[]slack.Attachment{
			{
				Color: "#96588a",
				Text:  slackURLFormat,
			},
		})

	return slack.OrderHistory.Send(ctx.Request.Context(), *payload)
}

func (o *NewOrder) slackFormatPayment(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("*Payment*\nTotal: %s€ (incl. %s€ tax)\nPaid with: %s\n\n",
		o.Total, o.TotalTax, o.PayMethod))
}

func (o *NewOrder) slackFormatDeliveryDate(sb *strings.Builder) error {
	date := ""
	timeslot := ""

	var value string
	for _, meta := range o.MetaData {
		if date != "" && timeslot != "" {
			break
		}
		if meta.Key == "dokan_delivery_time_date" {
			if err := json.Unmarshal(meta.Value, &value); err == nil {
				date = value
			} else {
				return &utils.APIError{
					Err:    fmt.Errorf("failed to unmarshal value for dokan_delivery_time_date: %w", err),
					Status: 500,
				}
			}
		} else if meta.Key == "dokan_delivery_time_slot" {
			if err := json.Unmarshal(meta.Value, &value); err == nil {
				timeslot = value
			} else {
				return &utils.APIError{
					Err:    fmt.Errorf("failed to unmarshal value for dokan_delivery_time_slot: %w", err),
					Status: 500,
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("*Delivery by*\nDate: %s\nTimeslot: %s\n\n",
		date, timeslot))

	return nil
}

func (o *NewOrder) slackFormatCustomer(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("*Customer*\nName: %s %s", o.Billing.FirstName, o.Billing.LastName))
	if o.Billing.Phone != "" {
		sb.WriteString(fmt.Sprintf("Phone: <tel:+%s|+%s>\n", o.Billing.Phone, o.Billing.Phone))
	} else {
		sb.WriteString("Phone:\n")
	}

	sb.WriteString(fmt.Sprintf("\nEmail: %s\nAddress: %s %s\nCompany: %s\n\n",
		o.Billing.Email, o.Billing.Address1, o.Billing.PostCode, o.Billing.Company))
}

func (o *NewOrder) slackFormatVendor(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("*Vendor*\nName: %s\nAddress: %s %s",
		o.Vendor.Name, o.Vendor.Address.Street, o.Vendor.Address.PostCode))
}
