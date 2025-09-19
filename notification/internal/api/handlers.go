package api

import (
	"errors"
	"net/http"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/jsonvalidator"
)

func (api *Application) handleSendNotification(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonvalidator.DecodeValidJson[model.Notification](r)
	if err != nil {
		jsonvalidator.EncodeJson(w, r, http.StatusBadRequest, problems)
		return
	}

	err = api.ctrl.Send(r.Context(), data.UserID, data.NotificationType, data.Message)
	if err != nil {
		if errors.Is(err, notification.ErrTooManyMessages) {
			jsonvalidator.EncodeJson(w, r, http.StatusTooManyRequests,
				map[string]any{"message": "too many messages of that type sent"})
			return
		}
		if errors.Is(err, notification.ErrUnknowNotificationType) {
			api.Logger.Error("unknown message sent", "body", data)
			jsonvalidator.EncodeJson(w, r, http.StatusInternalServerError,
				map[string]any{"message": "this notification type handler was not found"})
			return
		}
	}

	jsonvalidator.EncodeJson(w, r, http.StatusCreated,
		map[string]any{"message": "Message Sent"})
}
