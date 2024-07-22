package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestErrorResponse(t *testing.T) {
	// Create an instance of ErrorResponse
	errResp := &raiden.ErrorResponse{
		StatusCode: 400,
		Code:       "BadRequest",
		Details:    "Invalid input",
		Hint:       "Check the input parameters",
		Message:    "Bad Request",
	}

	// Test ErrorResponse fields
	assert.Equal(t, 400, errResp.StatusCode)
	assert.Equal(t, "BadRequest", errResp.Code)
	assert.Equal(t, "Invalid input", errResp.Details)
	assert.Equal(t, "Check the input parameters", errResp.Hint)
	assert.Equal(t, "Bad Request", errResp.Message)

	// Test the Error method
	assert.Equal(t, "Bad Request", errResp.Error())
}

func TestErrorResponse_EmptyFields(t *testing.T) {
	// Create an instance of ErrorResponse with minimal fields
	errResp := &raiden.ErrorResponse{
		Message: "Error occurred",
	}

	// Test ErrorResponse fields
	assert.Equal(t, 0, errResp.StatusCode)
	assert.Equal(t, "", errResp.Code)
	assert.Nil(t, errResp.Details)
	assert.Equal(t, "", errResp.Hint)
	assert.Equal(t, "Error occurred", errResp.Message)

	// Test the Error method
	assert.Equal(t, "Error occurred", errResp.Error())
}
