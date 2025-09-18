package config

import (
	"context"
	"log/slog"
	"os"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/jsonvalidator"
	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/validator"
)

type RLConfigProvider struct {
	limits rlConfigMap
}

type rlConfigMap map[model.NotificationType]RLConfig

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

func NewRLConfigLoader(limits rlConfigMap) *RLConfigProvider {
	return &RLConfigProvider{limits}
}

func (rlc *RLConfigProvider) GetConfig(t model.NotificationType) (RLConfig, bool) {
	cfg, ok := rlc.limits[t]
	return cfg, ok
}

func LoadFromJson(path string) (map[model.NotificationType]RLConfig, error) {
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
