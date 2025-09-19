package notification

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/google/uuid"
)

var (
	ErrUnknowNotificationType = errors.New("unknown notification type")
	ErrTooManyMessages        = errors.New("too many messages sent to given user")
)

type Controller struct {
	rl      rateLimiter
	configs config.Provider
}

func NewController(rateLimiter rateLimiter, configs config.Provider) *Controller {
	return &Controller{
		rateLimiter,
		configs,
	}
}

type rateLimiter interface {
	IsAllowed(ctx context.Context, key string, limit, windowSize int) (bool, error)
}

func (c *Controller) Send(ctx context.Context, id uuid.UUID, notificationType model.NotificationType, message string) error {
	cfg, ok := c.configs.GetConfig(notificationType)
	if !ok {
		return ErrUnknowNotificationType
	}

	valid, err := c.rl.IsAllowed(
		ctx,
		notificationType.GenKey(id.String()),
		cfg.Limit,
		cfg.WindowSize,
	)
	if err != nil {
		return err
	}
	if !valid {
		return ErrTooManyMessages
	}

	slog.Info("Message Sent!", "user-id", id, "notification-type", notificationType, "message", message)
	return nil
}
