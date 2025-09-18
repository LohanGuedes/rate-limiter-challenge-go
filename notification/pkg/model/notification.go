package model

import (
	"context"
	"fmt"
	"slices"

	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/validator"
	"github.com/google/uuid"
)

// RecordType defines a record type. Together with RecordID identifies unique records across all types.
type NotificationType string

// Existing record types.
const (
	NotificationTypeNews      = NotificationType("news-notification")
	NotificationTypeStatus    = NotificationType("status-notification")
	NotificationTypeMarketing = NotificationType("marketing-notification")
)

// UserID defines a user id.
type UserID uuid.UUID

// Notification defines an individual rating created by a user for some record.
type Notification struct {
	NotificationType NotificationType `json:"notificationType"`
	UserID           UserID           `json:"userId"`
	Message          string           `json:"message"`
}

// NOTE: add checks as needed, this is just an example of how I usually create
// small microservices validation, for more complex cases, we can and should
// use a more tested solution
func (n Notification) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	validNotificationTypes := []NotificationType{NotificationTypeNews, NotificationTypeStatus, NotificationTypeMarketing}

	// Field: "message"
	eval.CheckField(validator.NotBlank(n.Message), "message", "this field cannot be blank")
	eval.CheckField(
		validator.MinChars(n.Message, 10) &&
			validator.MaxChars(n.Message, 255),
		"message",
		"this field must have a length > 10 and be < 255")

	// Field: NotificationType
	eval.CheckField(
		slices.Contains(validNotificationTypes, n.NotificationType),
		"notificationType",
		fmt.Sprintf("Must be a valid notificationType: %v", validNotificationTypes),
	)

	return eval
}
