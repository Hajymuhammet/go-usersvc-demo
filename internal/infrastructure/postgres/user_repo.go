package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-usersvc-demo/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, created_at, updated_at`

	now := time.Now().UTC()
	result := &domain.User{}

	err := r.db.QueryRow(context.Background(), query,
		user.Name, user.Email, user.Password, now, now,
	).Scan(&result.ID, &result.Name, &result.Email, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("postgres: create user: %w", err)
	}

	return result, nil
}

// GetByID fetches a single user by primary key.
func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("postgres: get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail fetches a user by their email address.
func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("postgres: get user by email: %w", err)
	}

	return user, nil
}

// List returns a paginated list of users.
func (r *UserRepo) List(filter domain.ListFilter) (*domain.UserList, error) {
	offset := (filter.Page - 1) * filter.Limit

	// Count total
	var total int64
	if err := r.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, fmt.Errorf("postgres: count users: %w", err)
	}

	rows, err := r.db.Query(context.Background(),
		`SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC LIMIT $1 OFFSET $2`,
		filter.Limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan user: %w", err)
		}
		users = append(users, u)
	}

	return &domain.UserList{
		Data:  users,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}

// Update modifies an existing user's fields.
func (r *UserRepo) Update(id int64, input domain.UpdateUserInput) (*domain.User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *input.Email)
		argIdx++
	}
	if input.Password != nil {
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, *input.Password)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetByID(id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now().UTC())
	argIdx++
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d RETURNING id, name, email, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, args...).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("postgres: update user: %w", err)
	}

	return user, nil
}

// Delete removes a user by ID.
func (r *UserRepo) Delete(id int64) error {
	result, err := r.db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
