package auth

import (
	"testing"
	"time"
)

func TestGenerateAccessToken(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	token, err := manager.GenerateAccessToken(123)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	token, err := manager.GenerateRefreshToken(456)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	userID := int64(789)
	token, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	validatedID, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if validatedID != userID {
		t.Errorf("expected userID %d, got %d", userID, validatedID)
	}
}

func TestValidateRefreshToken_Success(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	userID := int64(999)
	token, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	validatedID, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if validatedID != userID {
		t.Errorf("expected userID %d, got %d", userID, validatedID)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	_, err := manager.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	manager1 := NewManager("secret-1", 1*time.Hour, 7*24*time.Hour)
	manager2 := NewManager("secret-2", 1*time.Hour, 7*24*time.Hour)

	token, err := manager1.GenerateAccessToken(999)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = manager2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret, got nil")
	}
}

func TestTokenType_AccessToken(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	accessToken, err := manager.GenerateAccessToken(123)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// Should be able to validate as access token
	_, err = manager.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("expected valid access token, got error: %v", err)
	}
}

func TestTokenType_RefreshToken(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	refreshToken, err := manager.GenerateRefreshToken(456)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	// Should be able to validate as refresh token
	_, err = manager.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token, got error: %v", err)
	}
}

func TestMultipleTokens(t *testing.T) {
	manager := NewManager("test-secret-key", 1*time.Hour, 7*24*time.Hour)

	userID := int64(111)
	accessToken, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	refreshToken, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	validatedAccessID, err := manager.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	validatedRefreshID, err := manager.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if validatedAccessID != validatedRefreshID {
		t.Errorf("expected same userID in both tokens, got access=%d, refresh=%d", validatedAccessID, validatedRefreshID)
	}
}
