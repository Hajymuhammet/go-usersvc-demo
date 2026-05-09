package service

import (
	"context"
)

// EmailProvider defines the interface for email providers (e.g., SMTP, SendGrid).
type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

// MockEmailProvider is a mock implementation for testing.
type MockEmailProvider struct {
}

// Send is a no-op mock implementation.
func (m *MockEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	return nil
}

// NewMockEmailProvider creates a new mock email provider.
func NewMockEmailProvider() EmailProvider {
	return &MockEmailProvider{}
}

// EmailService is a concrete implementation of domain.EmailService.
type EmailService struct {
	provider EmailProvider
}

// NewEmailService creates a new EmailService with the given provider.
func NewEmailService(provider EmailProvider) *EmailService {
	return &EmailService{provider: provider}
}

// SendWelcomeEmail sends a welcome email to the new user.
func (es *EmailService) SendWelcomeEmail(ctx context.Context, email string, name string) error {
	subject := "Welcome to Our Service"
	body := "Hello " + name + ", welcome to our service!"
	return es.provider.Send(ctx, email, subject, body)
}

// SendPasswordResetEmail sends a password reset email.
func (es *EmailService) SendPasswordResetEmail(ctx context.Context, email string, resetLink string) error {
	subject := "Password Reset Request"
	body := "Click the link below to reset your password: " + resetLink
	return es.provider.Send(ctx, email, subject, body)
}

// SendVerificationEmail sends an email verification email.
func (es *EmailService) SendVerificationEmail(ctx context.Context, email string, verificationLink string) error {
	subject := "Verify Your Email"
	body := "Click the link below to verify your email: " + verificationLink
	return es.provider.Send(ctx, email, subject, body)
}
