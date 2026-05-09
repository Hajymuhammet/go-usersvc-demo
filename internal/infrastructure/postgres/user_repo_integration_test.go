package postgres

import (
	"context"
	"testing"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/testhelpers"
)

func TestPostgresUserRepo_Integration(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := testhelpers.StartPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	// Run migrations
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if err := pgContainer.RunMigrations(ctx, createTableSQL); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repository
	repo := NewUserRepo(pgContainer.GetPool())

	t.Run("CreateUser_Success", func(t *testing.T) {
		user := &domain.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "hashed_password",
		}

		created, err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if created.ID == 0 {
			t.Error("Expected non-zero user ID")
		}
		if created.Name != user.Name {
			t.Errorf("Expected name %s, got %s", user.Name, created.Name)
		}
		if created.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, created.Email)
		}
	})

	t.Run("GetByID_Success", func(t *testing.T) {
		user := &domain.User{
			Name:     "Get By ID User",
			Email:    "getbyid@example.com",
			Password: "hashed_password",
		}

		created, _ := repo.Create(ctx, user)

		retrieved, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
		}
		if retrieved.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
		}
	})

	t.Run("GetByEmail_Success", func(t *testing.T) {
		user := &domain.User{
			Name:     "Get By Email User",
			Email:    "getbyemail@example.com",
			Password: "hashed_password",
		}

		created, _ := repo.Create(ctx, user)

		retrieved, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
		}
	})

	t.Run("GetByEmail_NotFound", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Fatal("Expected error for non-existent user")
		}
	})

	t.Run("Update_Success", func(t *testing.T) {
		user := &domain.User{
			Name:     "Update User",
			Email:    "update@example.com",
			Password: "hashed_password",
		}

		created, _ := repo.Create(ctx, user)
		newName := "Updated Name"

		updated, err := repo.Update(ctx, created.ID, domain.UpdateUserInput{
			Name: &newName,
		})

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, updated.Name)
		}
	})

	t.Run("Delete_Success", func(t *testing.T) {
		user := &domain.User{
			Name:     "Delete User",
			Email:    "delete@example.com",
			Password: "hashed_password",
		}

		created, _ := repo.Create(ctx, user)

		err := repo.Delete(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		_, err = repo.GetByID(ctx, created.ID)
		if err == nil {
			t.Fatal("Expected error after deletion")
		}
	})

	t.Run("List_Success", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 3; i++ {
			user := &domain.User{
				Name:     "List User " + string(rune(i)),
				Email:    "list" + string(rune(i)) + "@example.com",
				Password: "hashed_password",
			}
			repo.Create(ctx, user)
		}

		list, err := repo.List(ctx, domain.ListFilter{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(list.Data) == 0 {
			t.Error("Expected non-empty user list")
		}
	})
}

func TestPostgresUserRepo_DuplicateEmail(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := testhelpers.StartPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	pgContainer.RunMigrations(ctx, createTableSQL)

	repo := NewUserRepo(pgContainer.GetPool())

	user1 := &domain.User{
		Name:     "User 1",
		Email:    "duplicate@example.com",
		Password: "pass1",
	}

	_, err = repo.Create(ctx, user1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	user2 := &domain.User{
		Name:     "User 2",
		Email:    "duplicate@example.com", // Same email
		Password: "pass2",
	}

	_, err = repo.Create(ctx, user2)
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}
}
