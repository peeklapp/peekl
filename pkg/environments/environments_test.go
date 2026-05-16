package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentNameIsValid(t *testing.T) {
	type testCase struct {
		Environment string
		Valid       bool
	}

	testCases := []testCase{
		{Environment: "testing-invalid/environment", Valid: false},
		{Environment: "im-a-valid_environment-123", Valid: true},
		{Environment: "ME-TOO_IM_A_VALID_ENVIRONMENT-123", Valid: true},
	}

	for _, v := range testCases {
		res := EnvironmentNameIsValid(v.Environment)
		assert.Equalf(t, v.Valid, res, "Environment name: %s", v.Environment)
	}
}
