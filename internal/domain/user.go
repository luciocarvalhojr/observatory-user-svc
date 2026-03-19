// Package domain defines the core data structures for user-svc.
package domain

import "time"

// User represents a platform user.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest is the body for POST /users.
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"  binding:"required"`
}

// UpdateUserRequest is the body for PUT /users/:id.
type UpdateUserRequest struct {
	Name string `json:"name" binding:"required"`
}

// UserCreatedEvent is published to NATS when a user is created.
type UserCreatedEvent struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserDeletedEvent is published to NATS when a user is deleted.
type UserDeletedEvent struct {
	ID string `json:"id"`
}
