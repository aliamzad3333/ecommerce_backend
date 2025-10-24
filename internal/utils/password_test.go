package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "very-long-password-with-special-characters!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}
