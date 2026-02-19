package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceValidationError(t *testing.T) {
	valError := ValidationError{
		FieldName:    "name",
		ViolatedRule: "Field cannot be empty",
	}
	resValError := ResourceValidationError{
		Type:             "builtin.template",
		Title:            "create_file_from_template",
		ValidationErrors: []ValidationError{valError},
	}
	assert.Equal(
		t,
		"Invalid resource [builtin.template / 'create_file_from_template'] : [{\"field_name\":\"name\",\"violated_rule\":\"Field cannot be empty\"}]",
		resValError.Error(),
	)
}
