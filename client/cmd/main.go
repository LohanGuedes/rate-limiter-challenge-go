package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/google/uuid"
)

// When sending, send the type and use a custom json decoder for that type on
// the notification API in order to define the correct type and spawn the
// appropiated worker
func send(ctx context.Context, nt model.NotificationType, userID uuid.UUID, message string) error {
	switch nt {
	case model.NotificationTypeNews:
		// TODO: Call gateway
		return errors.New("news error")
	case model.NotificationTypeStatus:
		// TODO: Call gateway
		return errors.New("status errro")
	// TODO: Call gateway
	case model.NotificationTypeMarketing:
		return errors.New("marketing error")
	}
	return nil
}

func main() {
	id := uuid.New()
	slog.Info("Start sending notifications", "USERID", id)
	send(context.Background(), model.NotificationTypeNews, id, "Modak just got a new challenger!!!")
}
