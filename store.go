package main

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

type TaskStore interface {
	List(context.Context) ([]Task, error)
	Create(context.Context, Task) (Task, error)
	Change(context.Context, Task) error
	Delete(context.Context, int) error
	Get(context.Context, int) (Task, error)
}
