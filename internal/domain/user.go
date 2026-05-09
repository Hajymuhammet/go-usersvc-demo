package domain

import (
	"context"
	"time"
)

// User is the core entity of the service.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserInput holds data for creating a new user.
type CreateUserInput struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserInput holds optional fields for updating a user.
type UpdateUserInput struct {
	Name     *string `json:"name"     validate:"omitempty,min=2,max=100"`
	Email    *string `json:"email"    validate:"omitempty,email"`
	Password *string `json:"password" validate:"omitempty,min=6"`
}

// ListFilter holds pagination parameters.
type ListFilter struct {
	Page  int `json:"page"  validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// UserList is the paginated response for listing users.
type UserList struct {
	Data  []*User `json:"data"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
}

// UserRepository defines the persistence contract.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, filter ListFilter) (*UserList, error)
	Update(ctx context.Context, id int64, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id int64) error
}

// UserCache defines the caching contract.
type UserCache interface {
	Get(ctx context.Context, id int64) (*User, error)
	Set(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}

// EmailService defines the email sending contract.
type EmailService interface {
	SendWelcomeEmail(ctx context.Context, email string, name string) error
}
