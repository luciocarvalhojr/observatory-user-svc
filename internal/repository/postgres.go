// Package repository implements the PostgreSQL-backed user store for user-svc.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/domain"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// Repository defines the data access interface for users.
type Repository interface {
	Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}

type pgxRepository struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL-backed Repository.
func New(ctx context.Context, connStr string) (Repository, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("repository: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("repository: ping: %w", err)
	}
	return &pgxRepository{pool: pool}, nil
}

// Close releases the connection pool.
func Close(r Repository) {
	if repo, ok := r.(*pgxRepository); ok {
		repo.pool.Close()
	}
}

func (r *pgxRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *pgxRepository) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	const q = `
		INSERT INTO users (email, name)
		VALUES ($1, $2)
		RETURNING id, email, name, created_at, updated_at`

	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, req.Email, req.Name).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository: create user: %w", err)
	}
	return u, nil
}

func (r *pgxRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT id, email, name, created_at, updated_at
		FROM users WHERE id = $1`

	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get user: %w", err)
	}
	return u, nil
}

func (r *pgxRepository) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	const q = `
		UPDATE users SET name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, name, created_at, updated_at`

	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, req.Name, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: update user: %w", err)
	}
	return u, nil
}

func (r *pgxRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repository: delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
