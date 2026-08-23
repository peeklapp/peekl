package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndVerifyPassword(t *testing.T) {
	var defaultMemory uint32 = 64 * 1024
	var defaultIterations uint32 = 3
	var defaultParallelism uint8 = 2
	var defaultSaltLength uint32 = 16
	var defaultKeyLength uint32 = 32

	params := DefaultParams()
	assert.Equal(t, defaultMemory, params.Memory)
	assert.Equal(t, defaultIterations, params.Iterations)
	assert.Equal(t, defaultParallelism, params.Parallelism)
	assert.Equal(t, defaultSaltLength, params.SaltLength)
	assert.Equal(t, defaultKeyLength, params.KeyLength)

	password := "testing"

	hashedPassword, err := HashPassword(password, params)
	if err != nil {
		t.Error("Shouldn't have returned an error when calling 'HashPassword'")
	}

	hashedParams, hashedSalt, _, err := decodeHash(hashedPassword)
	if err != nil {
		t.Error("Shouldn't have returned an error when calling 'decodeHash'")
	}
	assert.Equal(t, defaultMemory, hashedParams.Memory)
	assert.Equal(t, defaultIterations, hashedParams.Iterations)
	assert.Equal(t, defaultParallelism, hashedParams.Parallelism)
	assert.Equal(t, defaultSaltLength, uint32(len(hashedSalt)))
	assert.Equal(t, defaultKeyLength, hashedParams.KeyLength)

	valid, err := VerifyPassword(
		password,
		hashedPassword,
	)
	if err != nil {
		t.Error("shouldn't have returned an error")
	}
	assert.True(t, valid)

	invalid, err := VerifyPassword(
		"not-correct-password",
		hashedPassword,
	)
	if err != nil {
		t.Error("Shouldn't have return an error")
	}
	assert.False(t, invalid)
}
