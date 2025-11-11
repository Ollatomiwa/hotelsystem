package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ollatomiwa/hotelsystem/user-service/internal/models"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db :db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, phone, role) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, user.Id, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Phone, user.Role,)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code.Name() == "unique_violation" {
					return fmt.Errorf("user with this email already exists")
				}
			}
			return fmt.Errorf("failed to create user: %w", err)
		}
		return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, userId string)(*models.User, error) {
	query := `
		SELECT id, email, firstName, lastName, role 
		FROM Users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.Id,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil 
}

func (r *UserRepository) GetUserByEmailAuth(ctx context.Context, email string)(*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, role 
		FROM Users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil 
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, phone = $3 WHERE email = $4`

	_, err := r.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.Phone, user.Email,)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil 
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE email =$2`

	_, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserById(ctx context.Context, id string) (*models.User, error) {
    query := `SELECT id, email, first_name, last_name, phone, role FROM users WHERE id = $1`
    
    var user models.User
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.Id,
        &user.Email,
        &user.FirstName,
        &user.LastName,
        &user.Phone,
        &user.Role,
    )
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

