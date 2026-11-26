package store

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")
var ErrEmailTaken = errors.New("email already registered")

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TaskStore interface {
	List(ctx context.Context, userID int) ([]Task, error)
	Create(ctx context.Context, userID int, t Task) (Task, error)
	Change(ctx context.Context, userID int, t Task) error
	Delete(ctx context.Context, userID int, id int) error
	Get(ctx context.Context, userID int, id int) (Task, error)
}

type User struct {
	ID           int
	Email        string
	PasswordHash string
}

type UserStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
