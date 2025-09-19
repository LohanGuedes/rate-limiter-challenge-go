package config

import (
	"context"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/validator"
)

// RLConfig defines a rate-limiter config.
// WindowSize must be in seconds.
type RLConfig struct {
	Limit      int `json:"limit"`
	WindowSize int `json:"window_size"`
}

// Provider defines a rate-limiter config provider
type Provider interface {
	GetConfig(model.NotificationType) (RLConfig, bool)
}

// Valid check each field from a given config returning a validator.Evaluator.
// See: jsonvalidator.Validator where it must and usually is used.
func (c RLConfig) Valid(_ context.Context) validator.Evaluator {
	var eval validator.Evaluator

	// Field: Limit
	eval.CheckField(c.Limit > 0, "limit", "this field cannot be blank nor 0")

	// Field: WindowSize
	eval.CheckField(c.WindowSize > 0, "window_size", "this field cannot be blank nor 0")

	return eval
}
