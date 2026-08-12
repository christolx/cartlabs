package httpapi

import (
	"fmt"
	"net/http"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/google/uuid"
)

type itemsResponse[T any] struct {
	Items []T `json:"items"`
}

func contractUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("map %s as UUID: %w", field, err)
	}
	return id, nil
}

func contractHealth(status contract.HealthStatus, dependencies map[string]string) contract.Health {
	result := contract.Health{Status: status}
	if dependencies != nil {
		result.Dependencies = &dependencies
	}
	return result
}

func contractProblem(status int, detail string) contract.Problem {
	return contract.Problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: &detail,
	}
}
