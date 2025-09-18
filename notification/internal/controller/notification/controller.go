package notification

import (
	"context"
	"errors"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
)

var ErrTooManyMessages = errors.New("too many messages sent to given user")

type Controller struct {
	rl      rateLimiter
	configs config.Provider
}

type rateLimiter interface {
	IsAllowed(ctx context.Context, key string, limit, windowSize int) (bool, error)
}

func (c *Controller) Send(ctx context.Context, id model.UserID, notificationType model.NotificationType, message string) error {
	// if !c.rl.IsAllowed(ctx, key, limit, windowSize) {
	// 	return ErrTooManyMessages
	// }
	return nil
}
