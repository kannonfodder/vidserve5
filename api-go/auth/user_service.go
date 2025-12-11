package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with hashed password
func CreateUser(db *pgxpool.Pool, username, password string) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New().String()

	query := `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username
	`

	user := &User{Id: id, Username: username}
	err = db.QueryRow(context.Background(), query, id, username, string(hashedPassword)).Scan(
		&user.Id, &user.Username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// AuthenticateUser validates username/password and returns user if valid
func AuthenticateUser(db *pgxpool.Pool, username, password string) (*User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`

	var user User
	var passwordHash string

	err := db.QueryRow(context.Background(), query, username).Scan(
		&user.Id, &user.Username, &passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *pgxpool.Pool, userID string) (*User, error) {
	query := `SELECT id, username FROM users WHERE id = $1`

	var user User
	err := db.QueryRow(context.Background(), query, userID).Scan(&user.Id, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(db *pgxpool.Pool, username string) (*User, error) {
	query := `SELECT id, username FROM users WHERE username = $1`

	var user User
	err := db.QueryRow(context.Background(), query, username).Scan(&user.Id, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}
