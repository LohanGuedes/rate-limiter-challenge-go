package main

import (
	"context"
	"log/slog"
	"time"

	httpnotification "github.com/LohanGuedes/modak-rate-limit-challenge/client/internal/gateway/notification/http"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/google/uuid"
)

func main() {
	id := uuid.New()
	gateway := httpnotification.New("http://localhost:8080")

	slog.Info("Start sending notifications", "USERID", id)

	err := gateway.Send(context.Background(), model.Notification{NotificationType: model.NotificationTypeStatus, UserID: id, Message: "Modak just got a new challenger!!!"})
	if err != nil {
		slog.Error("Failed to send", "error", err)
	}
	err = gateway.Send(context.Background(), model.Notification{NotificationType: model.NotificationTypeStatus, UserID: id, Message: "Modak just got a new challenger!!!"})
	if err != nil {
		slog.Error("Failed to send", "error", err)
	}

	// Expect to be hold
	err = gateway.Send(context.Background(), model.Notification{NotificationType: model.NotificationTypeStatus, UserID: id, Message: "OOOoooooooooh no!"})
	if err != nil {
		slog.Error("Failed to send", "error", err)
	}
	time.Sleep(10 * time.Second)
	// Should be sent with sucess
	err = gateway.Send(context.Background(), model.Notification{NotificationType: model.NotificationTypeStatus, UserID: id, Message: "We're in!!!"})
	if err != nil {
		slog.Error("Failed to send", "error", err)
	}
}
