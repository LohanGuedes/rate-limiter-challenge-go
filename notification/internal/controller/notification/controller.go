package notification

import (
	"context"
	"errors"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
)

var (
	ErrTooManyMessages        = errors.New("too many messages sent to given user")
	ErrUnknowNotificationType = errors.New("too many messages sent to given user")
)

type Controller struct {
	rl      rateLimiter
	configs config.Provider
}

type rateLimiter interface {
	IsAllowed(ctx context.Context, key string, limit, windowSize int) (bool, error)
}

func (c *Controller) Send(ctx context.Context, id model.UserID, notificationType model.NotificationType, message string) error {
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

	return nil
}
