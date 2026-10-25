package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"fmt"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1,$2,$3)
	          RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE id=$1`
	u := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE email=$1`
	u := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByEmail: %w", err)
	}
	return u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE username=$1`
	u := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByUsername: %w", err)
	}
	return u, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	return exists, err
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, username).Scan(&exists)
	return exists, err
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET username=$1, email=$2, password=$3, updated_at=NOW() WHERE id=$4`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.ID)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}
