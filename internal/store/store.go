package store

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TaskStore interface {
	List(context.Context) ([]Task, error)
	Create(context.Context, Task) (Task, error)
	Change(context.Context, Task) error
	Delete(context.Context, int) error
	Get(context.Context, int) (Task, error)
}
