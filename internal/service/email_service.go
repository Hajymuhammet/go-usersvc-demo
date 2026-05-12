package service

import (
	"context"
)

type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

type MockEmailProvider struct {
}

func (m *MockEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	return nil
}

func NewMockEmailProvider() EmailProvider {
	return &MockEmailProvider{}
}

type EmailService struct {
	provider EmailProvider
}

func NewEmailService(provider EmailProvider) *EmailService {
	return &EmailService{provider: provider}
}

func (es *EmailService) SendWelcomeEmail(ctx context.Context, email string, name string) error {
	subject := "Welcome to Our Service"
	body := "Hello " + name + ", welcome to our service!"
	return es.provider.Send(ctx, email, subject, body)
}

func (es *EmailService) SendPasswordResetEmail(ctx context.Context, email string, resetLink string) error {
	subject := "Password Reset Request"
	body := "Click the link below to reset your password: " + resetLink
	return es.provider.Send(ctx, email, subject, body)
}

func (es *EmailService) SendVerificationEmail(ctx context.Context, email string, verificationLink string) error {
	subject := "Verify Your Email"
	body := "Click the link below to verify your email: " + verificationLink
	return es.provider.Send(ctx, email, subject, body)
}
