package entity

import (
	assert "github.com/stretchr/testify/assert"
	"testing"
)

func TestCepValidation(t *testing.T) {
	// Test valid Cep
	validCep := NewCep("12345678", "Test Street", "", "Test District", "Test City", "TS")
	assert.NoError(t, validCep.Validate(), "Valid Cep should pass validation")

	// Test invalid Cep (missing Cep field)
	invalidCep := NewCep("", "Test Street", "", "Test District", "Test City", "TS")
	assert.Error(t, invalidCep.Validate(), "Cep with missing Cep field should fail validation")

}
