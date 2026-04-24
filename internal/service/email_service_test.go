package service

import (
	"context"
	"testing"
)

func TestMockEmailProvider_SendWelcomeEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	ctx := context.Background()

	err := provider.SendWelcomeEmail(ctx, "user@example.com", "John Doe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}

	email := provider.SentEmails[0]
	if email.To != "user@example.com" {
		t.Errorf("expected recipient user@example.com, got %s", email.To)
	}
	if email.Template != "welcome" {
		t.Errorf("expected template welcome, got %s", email.Template)
	}
	if email.Subject != "Welcome to User Service" {
		t.Errorf("expected subject 'Welcome to User Service', got %s", email.Subject)
	}
}

func TestMockEmailProvider_SendPasswordResetEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	ctx := context.Background()

	resetLink := "https://example.com/reset?token=abc123"
	err := provider.SendPasswordResetEmail(ctx, "user@example.com", resetLink)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}

	email := provider.SentEmails[0]
	if email.Template != "password_reset" {
		t.Errorf("expected template password_reset, got %s", email.Template)
	}
	if email.Subject != "Password Reset Request" {
		t.Errorf("expected subject 'Password Reset Request', got %s", email.Subject)
	}
}

func TestMockEmailProvider_SendVerificationEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	ctx := context.Background()

	verifyLink := "https://example.com/verify?token=xyz789"
	err := provider.SendVerificationEmail(ctx, "user@example.com", verifyLink)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}

	email := provider.SentEmails[0]
	if email.Template != "verification" {
		t.Errorf("expected template verification, got %s", email.Template)
	}
}

func TestMockEmailProvider_FailOnEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	provider.FailOnEmail = "fail@example.com"
	ctx := context.Background()

	err := provider.SendWelcomeEmail(ctx, "fail@example.com", "Test User")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(provider.SentEmails) != 0 {
		t.Fatalf("expected 0 sent emails on failure, got %d", len(provider.SentEmails))
	}
}

func TestMockEmailProvider_GetSentEmails(t *testing.T) {
	provider := NewMockEmailProvider()
	ctx := context.Background()

	provider.SendWelcomeEmail(ctx, "user1@example.com", "User One")
	provider.SendPasswordResetEmail(ctx, "user2@example.com", "https://reset.link")
	provider.SendVerificationEmail(ctx, "user3@example.com", "https://verify.link")

	emails := provider.GetSentEmails()
	if len(emails) != 3 {
		t.Fatalf("expected 3 emails, got %d", len(emails))
	}
}

func TestMockEmailProvider_ClearSentEmails(t *testing.T) {
	provider := NewMockEmailProvider()
	ctx := context.Background()

	provider.SendWelcomeEmail(ctx, "user@example.com", "User")
	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 email before clear, got %d", len(provider.SentEmails))
	}

	provider.ClearSentEmails()
	if len(provider.SentEmails) != 0 {
		t.Fatalf("expected 0 emails after clear, got %d", len(provider.SentEmails))
	}
}

func TestEmailService_SendWelcomeEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	service := NewEmailService(provider)
	ctx := context.Background()

	err := service.SendWelcomeEmail(ctx, "user@example.com", "John Doe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}
}

func TestEmailService_SendWelcomeEmailFailure(t *testing.T) {
	provider := NewMockEmailProvider()
	provider.FailOnEmail = "fail@example.com"
	service := NewEmailService(provider)
	ctx := context.Background()

	err := service.SendWelcomeEmail(ctx, "fail@example.com", "Failed User")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEmailService_SendPasswordResetEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	service := NewEmailService(provider)
	ctx := context.Background()

	resetLink := "https://example.com/reset?token=abc123"
	err := service.SendPasswordResetEmail(ctx, "user@example.com", resetLink)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}
}

func TestEmailService_SendVerificationEmail(t *testing.T) {
	provider := NewMockEmailProvider()
	service := NewEmailService(provider)
	ctx := context.Background()

	verifyLink := "https://example.com/verify?token=xyz789"
	err := service.SendVerificationEmail(ctx, "user@example.com", verifyLink)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(provider.SentEmails))
	}
}
