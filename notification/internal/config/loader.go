package config

import (
	"context"
	"log/slog"
	"os"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/jsonvalidator"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/validator"
)

// RLConfigProvider defines a configuration provider
type RLConfigProvider struct {
	limits rlConfigMap
}

type rlConfigMap map[model.NotificationType]RLConfig

// Valid checks for each valid RLConfig within the rlConfigMap
func (m rlConfigMap) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	for key, cfg := range m {
		prefixed := jsonvalidator.PrefixEvaluator(cfg.Valid(ctx), string(key))
		for field, msg := range prefixed {
			eval.AddFieldError(field, msg)
		}
	}

	return eval
}

// New creates a RLConfigProvider and returns it.
func NewRLConfigProvider(limits rlConfigMap) *RLConfigProvider {
	return &RLConfigProvider{limits}
}

// GetConfig Gets a config from the map
func (rlc *RLConfigProvider) GetConfig(t model.NotificationType) (RLConfig, bool) {
	cfg, ok := rlc.limits[t]
	return cfg, ok
}

// LoadFromJsonFile read configs from a json file and returns a
// map[model.NotificationType]RLConfig that must be used with a provider.
func LoadFromJsonFile(path string) (map[model.NotificationType]RLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	configMap, problems, err := jsonvalidator.DecodeValidJsonFromBytes[rlConfigMap](context.Background(), data)
	if err != nil {
		slog.Error("failed to unmarshall configurations from file", "filepath", path, "problems", problems, "original-error", err)
		return nil, err
	}
	return configMap, nil
}
