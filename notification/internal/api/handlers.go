package api

import (
	"net/http"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/jsonvalidator"
)

func (api *Application) handleSendNotification(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonvalidator.DecodeValidJson[model.Notification](r)
	if err != nil {
		jsonvalidator.EncodeJson(w, r, http.StatusBadRequest, problems)
		return
	}
}
