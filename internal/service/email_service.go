package service

import (
	"context"
	"fmt"
	"log"
)

// EmailProvider defines the interface for email sending.
type EmailProvider interface {
	SendWelcomeEmail(ctx context.Context, email, name string) error
	SendPasswordResetEmail(ctx context.Context, email, resetLink string) error
	SendVerificationEmail(ctx context.Context, email, verificationLink string) error
}

// EmailService handles email operations.
type EmailService struct {
	provider EmailProvider
	logger   *log.Logger
}

// NewEmailService creates a new EmailService.
func NewEmailService(provider EmailProvider) *EmailService {
	return &EmailService{
		provider: provider,
		logger:   log.New(log.Writer(), "[EMAIL] ", log.LstdFlags),
	}
}

// SendWelcomeEmail sends a welcome email to a new user.
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	s.logger.Printf("Sending welcome email to %s", email)
	if err := s.provider.SendWelcomeEmail(ctx, email, name); err != nil {
		s.logger.Printf("Failed to send welcome email to %s: %v", email, err)
		return fmt.Errorf("send welcome email: %w", err)
	}
	s.logger.Printf("Welcome email sent successfully to %s", email)
	return nil
}

// SendPasswordResetEmail sends a password reset email.
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, resetLink string) error {
	s.logger.Printf("Sending password reset email to %s", email)
	if err := s.provider.SendPasswordResetEmail(ctx, email, resetLink); err != nil {
		s.logger.Printf("Failed to send password reset email to %s: %v", email, err)
		return fmt.Errorf("send password reset email: %w", err)
	}
	s.logger.Printf("Password reset email sent successfully to %s", email)
	return nil
}

// SendVerificationEmail sends a verification email.
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, verificationLink string) error {
	s.logger.Printf("Sending verification email to %s", email)
	if err := s.provider.SendVerificationEmail(ctx, email, verificationLink); err != nil {
		s.logger.Printf("Failed to send verification email to %s: %v", email, err)
		return fmt.Errorf("send verification email: %w", err)
	}
	s.logger.Printf("Verification email sent successfully to %s", email)
	return nil
}

// MockEmailProvider is a mock implementation for testing.
type MockEmailProvider struct {
	SentEmails  []EmailRecord
	FailOnEmail string
}

// EmailRecord tracks sent emails.
type EmailRecord struct {
	To       string
	Subject  string
	Body     string
	Template string
}

// NewMockEmailProvider creates a new MockEmailProvider.
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		SentEmails: []EmailRecord{},
	}
}

// SendWelcomeEmail sends a mock welcome email.
func (m *MockEmailProvider) SendWelcomeEmail(ctx context.Context, email, name string) error {
	if m.FailOnEmail == email {
		return fmt.Errorf("mock error: failed to send to %s", email)
	}
	m.SentEmails = append(m.SentEmails, EmailRecord{
		To:       email,
		Subject:  "Welcome to User Service",
		Body:     fmt.Sprintf("Hello %s, welcome to our service!", name),
		Template: "welcome",
	})
	return nil
}

// SendPasswordResetEmail sends a mock password reset email.
func (m *MockEmailProvider) SendPasswordResetEmail(ctx context.Context, email, resetLink string) error {
	if m.FailOnEmail == email {
		return fmt.Errorf("mock error: failed to send to %s", email)
	}
	m.SentEmails = append(m.SentEmails, EmailRecord{
		To:       email,
		Subject:  "Password Reset Request",
		Body:     fmt.Sprintf("Click here to reset your password: %s", resetLink),
		Template: "password_reset",
	})
	return nil
}

// SendVerificationEmail sends a mock verification email.
func (m *MockEmailProvider) SendVerificationEmail(ctx context.Context, email, verificationLink string) error {
	if m.FailOnEmail == email {
		return fmt.Errorf("mock error: failed to send to %s", email)
	}
	m.SentEmails = append(m.SentEmails, EmailRecord{
		To:       email,
		Subject:  "Verify Your Email",
		Body:     fmt.Sprintf("Click here to verify your email: %s", verificationLink),
		Template: "verification",
	})
	return nil
}

// GetSentEmails returns all sent emails (for testing).
func (m *MockEmailProvider) GetSentEmails() []EmailRecord {
	return m.SentEmails
}

// ClearSentEmails clears the sent emails log (for testing).
func (m *MockEmailProvider) ClearSentEmails() {
	m.SentEmails = []EmailRecord{}
}
