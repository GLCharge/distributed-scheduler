package model

import (
	"errors"
	"testing"
)

func TestToCustomJobError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{"ErrInvalidJobType", ErrInvalidJobType, 400},
		{"ErrInvalidJobID", ErrInvalidJobID, 400},
		{"ErrEmptyPassword", ErrEmptyPassword, 400},
		{"ErrEmptyBearerToken", ErrEmptyBearerToken, 400},
		{"ErrAuthMethodNotDefined", ErrAuthMethodNotDefined, 400},
		{"Other error", errors.New("other error"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customError := ToCustomJobError(tt.err)
			if customError.Code != tt.expectedStatus {
				t.Errorf("Expected status %v but got %v", tt.expectedStatus, customError.Code)
			}
			if customError.Error() != tt.err.Error() {
				t.Errorf("Expected error message %v but got %v", tt.err.Error(), customError.Error())
			}
		})
	}
}
