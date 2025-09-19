package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/LohanGuedes/modak-rate-limit-challenge/client/internal/gateway"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
)

// Gateway defines a notification metadata HTTP gateway.
type Gateway struct {
	addr string
}

// New creates a new HTTP gateway for a notification service.
func New(addr string) *Gateway {
	return &Gateway{addr}
}

// Send gets sends a notification to a given user id.
func (g *Gateway) Send(ctx context.Context, nt model.Notification) error {
	data, err := json.Marshal(nt)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, g.addr+"/notify/send", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	slog.Debug("Body data", "body", bodyData)

	if resp.StatusCode == http.StatusNotFound {
		return gateway.ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return gateway.ErrTooManyMessages
	}
	slog.Info("Sent notification with success", "message", nt.Message)
	return nil
}
