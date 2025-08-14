package utils

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Err    error
	Status int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%v", e.Err)
}

func UnmarshalOrErr(rawData json.RawMessage, target any) error {
	if err := json.Unmarshal(rawData, target); err != nil {
		return &APIError{
			Err:    fmt.Errorf("failed to unmarshal json: %w", err),
			Status: 400,
		}
	}

	return nil
}
