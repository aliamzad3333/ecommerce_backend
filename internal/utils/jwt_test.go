package utils

import (
	"testing"
	"time"

	"ecommerce-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key",
		Expiration: 24 * time.Hour,
	}

	jwtManager := NewJWTManager(cfg)

	tests := []struct {
		name    string
		userID  string
		email   string
		role    string
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  "507f1f77bcf86cd799439011",
			email:   "test@example.com",
			role:    "user",
			wantErr: false,
		},
		{
			name:    "admin token generation",
			userID:  "507f1f77bcf86cd799439012",
			email:   "admin@example.com",
			role:    "admin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tt.userID, tt.email, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key",
		Expiration: 24 * time.Hour,
	}

	jwtManager := NewJWTManager(cfg)

	// Generate a valid token
	userID := "507f1f77bcf86cd799439011"
	email := "test@example.com"
	role := "user"

	token, err := jwtManager.GenerateToken(userID, email, role)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		wantErr        bool
		expectedUserID string
		expectedEmail  string
		expectedRole   string
	}{
		{
			name:           "valid token",
			token:          token,
			wantErr:        false,
			expectedUserID: userID,
			expectedEmail:  email,
			expectedRole:   role,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.expectedUserID, claims.UserID)
				assert.Equal(t, tt.expectedEmail, claims.Email)
				assert.Equal(t, tt.expectedRole, claims.Role)
			}
		})
	}
}
