package test

import (
	"encoding/json"
	"testing"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestErrorResponseFormat(t *testing.T) {
	// Test the new error response format
	errorResp := models.NewErrorResponse("Test error message")

	// Verify the structure
	assert.Equal(t, "error", errorResp.Status)
	assert.Len(t, errorResp.Messages, 1)
	assert.Equal(t, "Test error message", errorResp.Messages[0])

	// Test with multiple messages
	messages := []string{"Error 1", "Error 2", "Error 3"}
	errorRespMulti := models.NewErrorResponseWithMessages(messages)

	assert.Equal(t, "error", errorRespMulti.Status)
	assert.Len(t, errorRespMulti.Messages, 3)
	assert.Equal(t, messages, errorRespMulti.Messages)

	// Test JSON marshaling
	jsonData, err := json.Marshal(errorResp)
	assert.NoError(t, err)

	var unmarshaled models.ErrorResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, errorResp.Status, unmarshaled.Status)
	assert.Equal(t, errorResp.Messages, unmarshaled.Messages)
}

func TestErrorResponseFromError(t *testing.T) {
	// Test creating error response from an error
	testError := assert.AnError
	errorResp := models.NewErrorResponseFromError(testError)

	assert.Equal(t, "error", errorResp.Status)
	assert.Len(t, errorResp.Messages, 1)
	assert.Equal(t, testError.Error(), errorResp.Messages[0])
}
