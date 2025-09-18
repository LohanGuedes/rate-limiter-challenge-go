package jsonvalidator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LohanGuedes/modak-rate-limit-challenge/pkg/validator"
)

func EncodeJson[T any](w http.ResponseWriter, r *http.Request, statusCode int, data T) error {
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func DecodeJson[T any](r *http.Request) (T, error) {
	var data T

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("decode json: %w", err)
	}

	return data, nil
}

func DecodeValidJson[T validator.Validator](r *http.Request) (T, map[string]string, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := data.Valid(r.Context()); len(problems) > 0 {
		return data, problems, fmt.Errorf("invalid %T: %d problems", data, len(problems))
	}
	return data, nil, nil
}

func DecodeValidJsonFromBytes[T validator.Validator](ctx context.Context, body []byte) (T, map[string]string, error) {
	var data T
	if err := json.Unmarshal(body, &data); err != nil {
		return data, nil, fmt.Errorf("unmarshal json: %w", err)
	}
	if problems := data.Valid(ctx); len(problems) > 0 {
		return data, problems, fmt.Errorf("invalid %T: %d problems", data, len(problems))
	}
	return data, nil, nil
}

func PrefixEvaluator(e validator.Evaluator, prefix string) validator.Evaluator {
	out := validator.Evaluator{}
	for k, v := range e {
		out[fmt.Sprintf("%s.%s", prefix, k)] = v
	}
	return out
}
